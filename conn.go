package rcon

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Conn represents a remote connection to a Minecraft
// server. The RCON functionality exposed allows reading
// and writing to RCON connection.
type Conn struct {
	net.Conn
}

// NewConn returns RCON connection to Minecraft server.
// RCON login is performed as part of connection creation.
func NewConn(address string, port uint16, password string) (*Conn, error) {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	conn := &Conn{c}

	if err := conn.Authenticate(password); err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	return conn, nil
}

// Authenticate performs RCON login for connection.
// Error is returned if authentication is unsuccessful.
func (c *Conn) Authenticate(password string) error {
	req := NewPacket(LoginPacket, password)
	if err := c.WritePacket(req); err != nil {
		return fmt.Errorf("failed writing packet: %v", err)
	}

	resp, err := c.ReadPackets()
	if err != nil {
		return fmt.Errorf("failed reading packets: %v", err)
	}

	// Check response packet ID for failed login. Packet with
	// the same request ID represents successful authentication.
	// Packet with ID of -1 represents failed authentication.
	if len(resp) != 1 || resp[0].ID != req.ID {
		return errors.New("invalid password/response")
	}

	return nil
}

// SendCommand sends RCON command to server and returns response.
//
// Commands sent through the connection should be done sequentially.
// If concurrent command transmission is desired a new separate
// authenticated connection should be set up.
func (c *Conn) SendCommand(command string) (string, error) {
	req := NewPacket(CommandPacket, command)
	if err := c.WritePacket(req); err != nil {
		return "", fmt.Errorf("failed writing packet: %v", err)
	}

	resp, err := c.ReadPackets()
	if err != nil {
		return "", fmt.Errorf("failed reading packets: %v", err)
	}

	// Responses can be fragmented across multiple packets. Payloads from
	// each packet are combined to form the complete command response string.
	sb := strings.Builder{}
	for _, p := range resp {
		sb.WriteString(p.Payload)
	}

	return sb.String(), nil
}

// ReadPackets returns slice of response packets following a request.
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
func (c *Conn) ReadPackets() ([]*Packet, error) {
	packets := make([]*Packet, 0)
	for {
		buf := make([]byte, 1)
		data := make([]byte, 0)

		// The minimum length of a RCON packet is 14 bytes and is terminated
		// with two null bytes at the end. Bytes are read one at a time from
		// the connection until a complete packet has been read.
		for len(data) < 14 || data[len(data)-1] != 0 || data[len(data)-2] != 0 {
			_, err := c.Read(buf)
			if err != nil {
				return nil, err
			}

			data = append(data, buf[0])
		}

		// Send termination packet if it has not been sent.
		if len(packets) == 0 {
			tp := NewPacket(TerminationPacket, "MESSAGE-END")
			tb, _ := Marshal(tp)

			if _, err := c.Write(tb); err != nil {
				return nil, fmt.Errorf("failed writing termination packet: %v", err)
			}
		}

		p := &Packet{}
		if err := Unmarshal(data, p); err != nil {
			return nil, err
		}

		if p.Payload == TerminalResponse {
			break
		}

		packets = append(packets, p)
	}

	return packets, nil
}

// WritePacket sends request packet to RCON connection.
//
// Minecraft's server cannot handle queued request packets,
// so it is important to make sure a request is processed
// before making an additional request.
func (c *Conn) WritePacket(p *Packet) error {
	data, err := Marshal(p)
	if err != nil {
		return err
	}

	if _, err := c.Write(data); err != nil {
		return err
	}

	return nil
}
