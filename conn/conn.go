package conn

import (
	"errors"
	"fmt"
	"github.com/jltobler/go-rcon/packet"
	"net"
)

// RCON represents a remote connection to a Minecraft
// server. The RCON functionality exposed allows reading
// and writing to RCON connection.
type RCON struct {
	conn net.Conn
}

// New returns RCON connection to Minecraft server.
// RCON login is performed as part of connection creation.
func New(address string, port uint16, password string) (*RCON, error) {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return nil, err
	}

	rcon := &RCON{c}

	req := packet.New(packet.Login, password)
	if err := rcon.WritePacket(req); err != nil {
		return nil, err
	}

	resp, err := rcon.ReadPackets()
	if err != nil {
		return nil, err
	}

	// Check response packet ID for failed login. Packet with
	// the same request ID represents successful authentication.
	// Packet with ID of -1 represents failed authentication.
	if resp[0].ID != req.ID || resp[0].ID == -1 {
		return nil, errors.New("auth failed")
	}

	return rcon, nil
}

// WritePacket sends request packet to RCON connection.
// Minecraft's server cannot handle queued request packets,
// so it is important to make sure a request is processed
// before making an additional request.
func (r *RCON) WritePacket(p *packet.Packet) error {
	data, err := packet.Marshal(p)
	if err != nil {
		return err
	}

	if _, err := r.conn.Write(data); err != nil {
		return err
	}

	return nil
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
func (r *RCON) ReadPackets() ([]*packet.Packet, error) {
	packets := make([]*packet.Packet, 0)
	for {
		buf := make([]byte, 1)
		data := make([]byte, 0)

		// The minimum length of a RCON packet is 14 bytes and is terminated
		// with two null bytes at the end. Bytes are read one at a time from
		// the connection until a complete packet has been read.
		for len(data) < 14 || data[len(data)-1] != 0 || data[len(data)-2] != 0 {
			_, err := r.conn.Read(buf)
			if err != nil {
				return nil, err
			}

			data = append(data, buf[0])
		}

		// Send termination packet if it has not been sent.
		if len(packets) == 0 {
			tp := packet.New(packet.Termination, "MESSAGE-END")
			tb, err := packet.Marshal(tp)
			if err != nil {
				return nil, err
			}

			if _, err := r.conn.Write(tb); err != nil {
				return nil, err
			}
		}

		p := &packet.Packet{}
		if err := packet.Unmarshal(data, p); err != nil {
			return nil, err
		}

		if p.Payload == packet.TerminalResponse {
			break
		}

		packets = append(packets, p)
	}

	return packets, nil
}

// Close terminates RCON connection to Minecraft server.
func (r *RCON) Close() error {
	return r.conn.Close()
}
