package rcon

import "fmt"

// Client acts as the entrypoint to send RCON messages. Client hold
// all the configuration necessary to connect to Minecraft server.
type Client struct {
	addr     string
	password string
}

// NewClient creates and returns a configured RCON client.
func NewClient(addr, password string) *Client {
	return &Client{
		addr:     addr,
		password: password,
	}
}

// Send establishes a new authenticated connection to the Minecraft
// server and transmits requested command. Not all commands generate
// a response from the server. Any response from the server is returned
// to requester. If connection failure occurs an error is returned.
//
// This function is concurrency-safe since each command sent to the
// Minecraft server creates a new connection. Upon completion of the
// request the established connection is closed.
func (c *Client) Send(command string) (string, error) {
	conn, err := Dial(c.addr, c.password)
	if err != nil {
		return "", fmt.Errorf("failed to establish connection: %w", err)
	}
	defer conn.Close()

	return conn.SendCommand(command)
}
