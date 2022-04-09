package rcon

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultPort uint16 = 25575
)

// Conn represents a remote RCON connection to a Minecraft server.
//
// The RCON connection allows server administrators to remotely
// execute commands on Minecraft servers.
type Conn struct {
	conn     net.Conn
	mutex    sync.Mutex
	packets  chan *Packet
	isClosed bool
}

// Dial connects and authenticates to the specified URL.
//
// The underlying transport layer connection is created along
// with the configured RCON connection.
func Dial(addr, password string) (*Conn, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "rcon" {
		return nil, fmt.Errorf("unsupported scheme '%s'", u.Scheme)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// we assume that error is due to missing port
		host = u.Host
		port = strconv.Itoa(int(DefaultPort))
	}

	c, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}

	return NewConn(c, password)
}

// NewConn wraps transport layer connection with RCON configuration.
//
// RCON authentication is performed as part of connection configuration.
// Failed authentication closes the transport layer connection.
func NewConn(c net.Conn, password string) (*Conn, error) {
	conn := &Conn{
		conn:     c,
		mutex:    sync.Mutex{},
		packets:  make(chan *Packet),
		isClosed: false,
	}

	conn.start()

	if err := conn.authenticate(password); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return conn, nil
}

// SendCommand sends RCON command to server and returns response.
//
// Commands sent are processed sequentially and the next command
// cannot execute until the previous completes. All connection errors
// result in the connection being closed.
func (c *Conn) SendCommand(command string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isClosed {
		return "", errors.New("connection closed")
	}

	req := NewPacket(CommandPacket, command)
	if err := c.writePacket(req); err != nil {
		_ = c.Close()
		return "", fmt.Errorf("failed writing packet: %w", err)
	}

	resp, err := c.readPackets()
	if err != nil {
		_ = c.Close()
		return "", fmt.Errorf("failed reading packets: %w", err)
	}

	// Responses can be fragmented across multiple packets. Payloads from
	// each packet are combined to form the complete command response string.
	sb := strings.Builder{}
	for _, p := range resp {
		sb.WriteString(p.Payload)
	}

	return sb.String(), nil
}

// IsClosed returns whether the RCON connection is closed.
//
// It is possible that an RCON connection becomes closed due to the
// server hanging up or other connection errors.
func (c *Conn) IsClosed() bool {
	return c.isClosed
}

// Close closes the connection.
// Any blocked command executions will be unblocked and return errors.
func (c *Conn) Close() error {
	c.isClosed = true
	return c.conn.Close()
}

// start begins reading response packets asynchronously from connection.
//
// Read packets are queued in channel for later response processing.
// Errors reading packets result in the connection being closed.
func (c *Conn) start() {
	go func() {
		for {
			packet, err := c.readPacket()
			if err != nil {
				_ = c.Close()
				close(c.packets)
				return
			}

			c.packets <- packet
		}
	}()
}

// authenticate performs RCON login for connection.
//
// Error is returned if authentication is unsuccessful or
// there are issues reading or writing to the connection.
func (c *Conn) authenticate(password string) error {
	req := NewPacket(LoginPacket, password)
	if err := c.writePacket(req); err != nil {
		return fmt.Errorf("failed writing packet: %w", err)
	}

	resp, err := c.readPackets()
	if err != nil {
		return fmt.Errorf("failed reading packets: %w", err)
	}

	// Check response packet ID for failed login. Packet with
	// the same request ID represents successful authentication.
	// Packet with ID of -1 represents failed authentication.
	if len(resp) != 1 || resp[0].ID != req.ID {
		return errors.New("invalid password/response")
	}

	return nil
}

// readPackets returns slice of response packets following a request.
//
// Since responses can be fragmented across multiple packets, all
// requests are accompanied by a single no-op "termination" packet
// used to indicate that all response packets have been received.
//
// Since Minecraft's server does not support queued request packets
// (which is very annoying), the "termination" packet cannot be sent
// until the original request packet has been processed. Upon receiving
// the first full response packet a "termination" packet is sent allowing
// for the reader to know when all responses have been received for the
// initial request.
//
// The returned response slice contains all packets up to the
// "termination" packet. This group of packets represents the response
// to a single request packet. The ID of each packet will match the
// corresponding request packet ID.
func (c *Conn) readPackets() ([]*Packet, error) {
	packets := make([]*Packet, 0)
	for p := range c.packets {
		// Send termination packet if it has not been sent.
		if len(packets) == 0 {
			tp := NewPacket(TerminationPacket, "MESSAGE-END")
			tb, _ := Marshal(tp)

			if _, err := c.conn.Write(tb); err != nil {
				return nil, fmt.Errorf("failed writing termination packet: %w", err)
			}
		}

		if p.Payload == TerminalResponse {
			break
		}

		packets = append(packets, p)
	}

	return packets, nil
}

// readPacket reads from connection and creates next packet.
func (c *Conn) readPacket() (*Packet, error) {
	buf := make([]byte, 1)
	data := make([]byte, 0)

	// The minimum length of a RCON packet is 14 bytes and is terminated
	// with two null bytes at the end. Bytes are read one at a time from
	// the connection until a complete packet has been read.
	for len(data) < 14 || data[len(data)-1] != 0 || data[len(data)-2] != 0 {
		_, err := c.conn.Read(buf)
		if err != nil {
			return nil, err
		}

		data = append(data, buf[0])
	}

	p := &Packet{}
	if err := Unmarshal(data, p); err != nil {
		return nil, err
	}

	return p, nil
}

// writePacket sends request packet to RCON connection.
//
// Minecraft's server cannot handle queued request packets,
// so it is important to make sure a request is processed
// before making an additional request.
func (c *Conn) writePacket(p *Packet) error {
	data, err := Marshal(p)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}
