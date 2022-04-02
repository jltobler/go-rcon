package rcon

import "fmt"

// Client acts as the entrypoint to send RCON messages. Client hold
// all the configuration necessary to connect to Minecraft server.
type Client struct {
	address  string
	port     uint16
	password string
}

// NewClient creates and returns a configured RCON client.
func NewClient(address string, port uint16, password string) *Client {
	return &Client{
		address:  address,
		port:     port,
		password: password,
	}
}

// Send establishes a new authenticated connection to the Minecraft
// server and transmits requested command. Not all commands generate
// a response from the server. Any response from the server is returned
// to requester. If connection failure occurs an error is returned.
//
// This function can be used concurrency-safe since each command sent to
// the Minecraft server creates a new connection. Upon completion of
// request the established connection is closed.
func (c *Client) Send(command string) (string, error) {
	conn, err := NewConn(c.address, c.port, c.password)
	if err != nil {
		return "", fmt.Errorf("failed to establish connection: %v", err)
	}
	defer conn.Close()

	return conn.SendCommand(command)
}
