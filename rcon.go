package rcon

import (
	"git.tobler.codes/minecraft/go-rcon/conn"
	"git.tobler.codes/minecraft/go-rcon/packet"
	"strings"
)

// Client acts as the entrypoint to send RCON messages. Client hold
// all the configuration necessary to connect to Minecraft server.
type Client struct {
	address  string
	port     uint16
	password string
}

// New creates and returns a configured RCON client.
func New(address string, port uint16, password string) *Client {
	return &Client{
		address:  address,
		port:     port,
		password: password,
	}
}

// Send establishes a new authenticated connection to the Minecraft
// server and transmits requested command. Not all commands generate
// a response from the server. Any response from the server is return
// to requester. If connection failure occurs an error is returned.
//
// Each command sent to the Minecraft server creates a new connection.
// Upon completion of request the established connection is closed.
func (c *Client) Send(command string) (string, error) {
	rcon, err := conn.New(c.address, c.port, c.password)
	if err != nil {
		return "", err
	}
	defer rcon.Close()

	req := packet.New(packet.Command, command)
	if err := rcon.WritePacket(req); err != nil {
		return "", err
	}

	resp, err := rcon.ReadPackets()
	if err != nil {
		return "", err
	}

	// Responses can be fragmented across multiple packets. Payloads from
	// each packet are combined to form the complete command response string.
	sb := strings.Builder{}
	for _, p := range resp {
		sb.WriteString(p.Payload)
	}

	return sb.String(), nil
}
