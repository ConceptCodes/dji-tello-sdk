package transport

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/errors"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
)

// UDPClientInterface defines the interface for UDP client operations
type UDPClientInterface interface {
	Send(data []byte) error
	Receive(bufferSize int, timeout time.Duration) (string, error)
	Close() error
}

// CommandConnection manages UDP communication with the Tello drone
type CommandConnection struct {
	config   config.TransportConfig
	client   UDPClientInterface
	mutex    sync.Mutex
	useBound bool
}

// NewCommandConnection creates a new command connection with default configuration
func NewCommandConnection() (*CommandConnection, error) {
	return NewCommandConnectionWithConfig(config.DefaultTransportConfig())
}

// NewCommandConnectionWithConfig creates a new command connection with custom configuration
func NewCommandConnectionWithConfig(cfg config.TransportConfig) (*CommandConnection, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, errors.WrapSDKError(err, errors.ErrConfigValidation, "CommandConnection",
			"invalid transport configuration")
	}

	client, err := udp.NewUDPClientWithLocalAddr(cfg.DroneHost, cfg.LocalCommandAddr)
	if err != nil {
		return nil, errors.ConnectionError("CommandConnection", "create UDP client", err)
	}

	return &CommandConnection{
		config:   cfg,
		client:   client,
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
		localAddr = c.config.LocalCommandAddr
	}

	client, err := udp.NewUDPClientWithLocalAddr(c.config.DroneHost, localAddr)
	if err != nil {
		return errors.ConnectionError("CommandConnection", "recreate UDP client", err)
	}
	c.client = client
	c.useBound = bindToCommandPort
	return nil
}

// NewCommandConnectionWithClient creates a CommandConnection with a custom UDP client (for testing)
func NewCommandConnectionWithClient(client UDPClientInterface) *CommandConnection {
	return &CommandConnection{
		config: config.DefaultTransportConfig(),
		client: client,
	}
}

// SendCommand sends a command to the drone and returns the response
func (c *CommandConnection) SendCommand(command string) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.client == nil {
		return "", errors.NewSDKError(errors.ErrConnectionFailed, "CommandConnection",
			"UDP client is not initialized")
	}

	var lastErr error
	data := []byte(command + "\r\n")
	fallbackTried := false

	for attempt := 1; attempt <= c.config.CommandRetries; attempt++ {
		// Small delay before first attempt to give the socket a moment to stabilize after bind.
		if attempt == 1 {
			time.Sleep(c.config.CommandSendDelay)
		}

		if err := c.client.Send(data); err != nil {
			lastErr = errors.Wrapf(err, "failed to send command '%s'", command)

			// Recreate the socket on transient network errors before retrying.
			if isBrokenPipe(err) && !fallbackTried {
				fallbackTried = true
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to rebuild command socket (ephemeral port)")
				}
			} else {
				if recErr := c.reconnect(); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to reconnect command socket")
				}
			}
			continue
		}

		response, err := c.client.Receive(2048, c.config.CommandTimeout)
		if err != nil {
			lastErr = errors.Wrapf(err, "failed to receive response for command '%s'", command)
			if isBrokenPipe(err) && !fallbackTried {
				fallbackTried = true
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to rebuild command socket (ephemeral port)")
				}
			} else {
				if recErr := c.reconnect(); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to reconnect command socket")
				}
			}
			continue
		}

		return response, nil
	}

	if lastErr == nil {
		lastErr = errors.NewSDKError(errors.ErrCommandFailed, "CommandConnection",
			fmt.Sprintf("failed to send command '%s': unknown error", command))
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

// GetConfig returns the current transport configuration
func (c *CommandConnection) GetConfig() config.TransportConfig {
	return c.config
}

// UpdateConfig updates the transport configuration
func (c *CommandConnection) UpdateConfig(cfg config.TransportConfig) error {
	if err := cfg.Validate(); err != nil {
		return errors.WrapSDKError(err, errors.ErrConfigValidation, "CommandConnection",
			"invalid transport configuration")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.config = cfg

	// Reconnect with new configuration if client exists
	if c.client != nil {
		return c.reconnect()
	}

	return nil
}
