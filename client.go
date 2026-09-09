// Package rcon provides a Go client for the Minecraft RCON protocol.
package rcon

import (
	"fmt"
)

// ClientOption defines a function that can be used to configure a Client.
type ClientOption func(*Client)

// WithOptions returns an ClientOption that sets the character set for the client.
func WithOptions(cs CharSet) ClientOption {
	return func(c *Client) {
		c.charSet = cs
	}
}

// Client acts as the entrypoint to send RCON messages. Client hold
// all the configuration necessary to connect to Minecraft server.
type Client struct {
	addr     string
	password string
	charSet  CharSet
}

// NewClient creates and returns a configured RCON client.
func NewClient(addr, password string, opts ...ClientOption) *Client {
	c := &Client{
		addr:     addr,
		password: password,
		charSet:  CharSetASCII,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
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
	conn, err := Dial(c.addr, c.password, c.charSet)
	if err != nil {
		return "", fmt.Errorf("failed to establish connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	return conn.SendCommand(command)
}
