package transport

import (
	"fmt"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
)

type CommandConnection struct {
	client *udp.UDPClient
}

func NewCommandConnection(addr string, port int) (*CommandConnection, error) {
	client, err := udp.NewUDPClient(addr, port)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP client for commands: %w", err)
	}
	return &CommandConnection{
		client: client,
	}, nil
}

func (c *CommandConnection) SendCommand(command string) ([]byte, error) {
	data := []byte(command)
	if err := c.client.Send(data); err != nil {
		return nil, fmt.Errorf("failed to send command '%s': %w", command, err)
	}

	response, err := c.client.Receive(1024, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to receive response for command '%s': %w", command, err)
	}

	return response, nil
}

func (c *CommandConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
