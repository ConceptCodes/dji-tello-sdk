package transport

import (
	"fmt"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
)

const host = "192.168.10.1:8889"

type CommandConnection struct {
	client *udp.UDPClient
}

func NewCommandConnection() (*CommandConnection, error) {
	client, err := udp.NewUDPClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP client for commands: %w", err)
	}
	return &CommandConnection{
		client: client,
	}, nil
}

func (c *CommandConnection) SendCommand(command string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("UDP client is not initialized")
	}

	data := []byte(command + "\r\n") 
	if err := c.client.Send(data); err != nil {
		return "", fmt.Errorf("failed to send command '%s': %w", command, err)
	}

	response, err := c.client.Receive(2048, 7*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to receive response for command '%s': %w", command, err)
	}

	return response, nil
}

func (c *CommandConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
