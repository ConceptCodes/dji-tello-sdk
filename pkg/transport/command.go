package transport

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/errors"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// UDPClientInterface defines the interface for UDP client operations
type UDPClientInterface interface {
	Send(data []byte) error
	Receive(bufferSize int, timeout time.Duration) (string, error)
	Close() error
}

// CommandConnection manages UDP communication with the Tello drone
type CommandConnection struct {
	config      config.TransportConfig
	client      UDPClientInterface
	mutex       sync.Mutex
	useBound    bool
	inFallback  bool // Track if we're in ephemeral port fallback mode
	fallbackTry int  // Count of fallback attempts
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

	// Check network connectivity before attempting to create socket
	if err := validateNetworkConnectivity(cfg.DroneHost); err != nil {
		return nil, errors.WrapSDKError(err, errors.ErrConnectionFailed, "CommandConnection",
			"network connectivity check failed")
	}

	client, err := udp.NewUDPClientWithLocalAddr(cfg.DroneHost, cfg.LocalCommandAddr)
	if err != nil {
		return nil, errors.ConnectionError("CommandConnection", "create UDP client", err)
	}

	return &CommandConnection{
		config:      cfg,
		client:      client,
		useBound:    true,
		inFallback:  false,
		fallbackTry: 0,
	}, nil
}

// reconnect tears down and recreates the UDP client bound to the Tello command port.
func (c *CommandConnection) reconnect() error {
	return c.reconnectWithLocal(c.useBound)
}

func (c *CommandConnection) reconnectWithLocal(bindToCommandPort bool) error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			utils.Logger.Warnf("Failed to close existing UDP client during reconnect: %v", err)
		}
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

	// Only update useBound if we're not in temporary fallback mode
	// or if we're explicitly switching back to bound mode
	if !c.inFallback || bindToCommandPort {
		c.useBound = bindToCommandPort
	}

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

			// Determine error type and handle appropriately
			if isBindingError(err) && !fallbackTried {
				// Port binding error - try ephemeral port as fallback
				fallbackTried = true
				c.inFallback = true
				c.fallbackTry++
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to rebuild command socket (ephemeral port)")
				}
				utils.Logger.Warnf("Port binding error, switching to ephemeral port (attempt %d)", c.fallbackTry)
			} else if isNetworkError(err) {
				// Network error - try to reconnect with current mode
				if recErr := c.reconnect(); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to reconnect command socket")
				}
				utils.Logger.Warnf("Network error detected: %v", err)
			} else {
				// Other error - try to reconnect
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

			// Determine error type and handle appropriately
			if isBindingError(err) && !fallbackTried {
				// Port binding error - try ephemeral port as fallback
				fallbackTried = true
				c.inFallback = true
				c.fallbackTry++
				if recErr := c.reconnectWithLocal(false); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to rebuild command socket (ephemeral port)")
				}
				utils.Logger.Warnf("Port binding error, switching to ephemeral port (attempt %d)", c.fallbackTry)
			} else if isNetworkError(err) {
				// Network error - try to reconnect with current mode
				if recErr := c.reconnect(); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to reconnect command socket")
				}
				utils.Logger.Warnf("Network error detected: %v", err)
			} else {
				// Other error - try to reconnect
				if recErr := c.reconnect(); recErr != nil {
					return "", errors.WrapSDKError(recErr, errors.ErrConnectionFailed, "CommandConnection",
						"failed to reconnect command socket")
				}
			}
			continue
		}

		// If we're in fallback mode and got a successful response,
		// try to switch back to bound port for next command
		if c.inFallback && c.fallbackTry > 0 {
			c.tryRevertFromFallback()
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
	errStr := err.Error()
	return strings.Contains(errStr, "broken pipe")
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()

	// Network-related errors
	networkErrors := []string{
		"broken pipe",
		"connection refused",
		"network is unreachable",
		"no route to host",
		"host is down",
		"connection timed out",
		"i/o timeout",
	}

	for _, netErr := range networkErrors {
		if strings.Contains(strings.ToLower(errStr), netErr) {
			return true
		}
	}

	return false
}

func isBindingError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()

	// Port binding errors
	bindingErrors := []string{
		"address already in use",
		"permission denied",
		"bind:",
	}

	for _, bindErr := range bindingErrors {
		if strings.Contains(strings.ToLower(errStr), bindErr) {
			return true
		}
	}

	return false
}

// tryRevertFromFallback attempts to switch back from ephemeral port to bound port
func (c *CommandConnection) tryRevertFromFallback() {
	if !c.inFallback || c.fallbackTry == 0 {
		return
	}

	// Only try to revert after a few successful commands in fallback mode
	// This prevents flapping between modes
	if c.fallbackTry < 3 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Try to reconnect with bound port
	if err := c.reconnectWithLocal(true); err != nil {
		utils.Logger.Debugf("Failed to revert from fallback mode: %v", err)
		// Stay in fallback mode if we can't bind to port
		return
	}

	// Successfully reverted to bound port
	c.inFallback = false
	c.fallbackTry = 0
	utils.Logger.Info("Successfully reverted from ephemeral port to bound port")
}

// validateNetworkConnectivity checks if the drone host is reachable
func validateNetworkConnectivity(droneHost string) error {
	// Parse the drone host address
	addr, err := net.ResolveUDPAddr("udp", droneHost)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve drone host '%s'", droneHost)
	}

	// Try to create a temporary connection to check network reachability
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		// Check if it's a network unreachable error
		if strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
			strings.Contains(strings.ToLower(err.Error()), "no route to host") {
			return errors.NewSDKError(errors.ErrConnectionFailed, "CommandConnection",
				fmt.Sprintf("cannot reach drone at %s. Please ensure you are connected to Tello WiFi", droneHost))
		}
		return errors.Wrapf(err, "failed to connect to drone at %s", droneHost)
	}
	conn.Close()

	return nil
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
