package transport

import (
	"fmt"
	"sync"
	"time"
	"strings"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
)

const host = "192.168.10.1:8889"
const localCommandAddr = "0.0.0.0:8889"
const commandRetries = 3
const commandSendDelay = 200 * time.Millisecond

// UDPClientInterface defines the interface for UDP client operations
type UDPClientInterface interface {
	Send(data []byte) error
	Receive(bufferSize int, timeout time.Duration) (string, error)
	Close() error
}

type CommandConnection struct {
	client UDPClientInterface
	mutex  sync.Mutex
	useBound bool
}

func NewCommandConnection() (*CommandConnection, error) {
	client, err := udp.NewUDPClientWithLocalAddr(host, localCommandAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP client for commands: %w", err)
	}
	return &CommandConnection{
		client: client,
		useBound: true,
	}, nil
}

// reconnect tears down and recreates the UDP client bound to the Tello command port.
func (c *CommandConnection) reconnect() error {
	return c.reconnectWithLocal(c.useBound)
}

func (c *CommandConnection) reconnectWithLocal(bindToCommandPort bool) error {
	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}

	localAddr := ""
	if bindToCommandPort {
		localAddr = localCommandAddr
	}

	client, err := udp.NewUDPClientWithLocalAddr(host, localAddr)
	if err != nil {
		return fmt.Errorf("failed to recreate UDP client for commands: %w", err)
	}
	c.client = client
	c.useBound = bindToCommandPort
	return nil
}

// NewCommandConnectionWithClient creates a CommandConnection with a custom UDP client (for testing)
func NewCommandConnectionWithClient(client UDPClientInterface) *CommandConnection {
	return &CommandConnection{
		client: client,
	}
}

func (c *CommandConnection) SendCommand(command string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client == nil {
		return "", fmt.Errorf("UDP client is not initialized")
	}

	var lastErr error
	data := []byte(command + "\r\n")
	fallbackTried := false

	for attempt := 1; attempt <= commandRetries; attempt++ {
		// Small delay before first attempt to give the socket a moment to stabilize after bind.
		if attempt == 1 {
			time.Sleep(commandSendDelay)
		}

		if err := c.client.Send(data); err != nil {
			lastErr = fmt.Errorf("failed to send command '%s': %w", command, err)
			// Recreate the socket on transient network errors before retrying.
			if isBrokenPipe(err) && !fallbackTried {
				fallbackTried = true
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", fmt.Errorf("%v; additionally failed to rebuild command socket (ephemeral port): %w", lastErr, recErr)
				}
			} else {
				if recErr := c.reconnect(); recErr != nil {
					return "", fmt.Errorf("%v; additionally failed to reconnect command socket: %w", lastErr, recErr)
				}
			}
			continue
		}

		response, err := c.client.Receive(2048, 7*time.Second)
		if err != nil {
			lastErr = fmt.Errorf("failed to receive response for command '%s': %w", command, err)
			if isBrokenPipe(err) && !fallbackTried {
				fallbackTried = true
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", fmt.Errorf("%v; additionally failed to rebuild command socket (ephemeral port): %w", lastErr, recErr)
				}
			} else {
				if recErr := c.reconnect(); recErr != nil {
					return "", fmt.Errorf("%v; additionally failed to reconnect command socket: %w", lastErr, recErr)
				}
			}
			continue
		}

		return response, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("failed to send command '%s': unknown error", command)
	}
	return "", lastErr
}

func isBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "broken pipe")
}

func (c *CommandConnection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
