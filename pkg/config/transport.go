package config

import (
	"fmt"
	"net"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/errors"
)

// TransportConfig holds configuration for drone communication
type TransportConfig struct {
	// DroneHost is the IP:port of the Tello drone (default: "192.168.10.1:8889")
	DroneHost string `json:"drone_host" yaml:"drone_host"`

	// LocalCommandAddr is the local address to bind for command communication (default: "0.0.0.0:8889")
	LocalCommandAddr string `json:"local_command_addr" yaml:"local_command_addr"`

	// LocalStateAddr is the local address to bind for state updates (default: "0.0.0.0:8890")
	LocalStateAddr string `json:"local_state_addr" yaml:"local_state_addr"`

	// LocalVideoAddr is the local address to bind for video stream (default: "0.0.0.0:11111")
	LocalVideoAddr string `json:"local_video_addr" yaml:"local_video_addr"`

	// CommandRetries is the number of retry attempts for failed commands (default: 3)
	CommandRetries int `json:"command_retries" yaml:"command_retries"`

	// CommandTimeout is the timeout for command responses (default: 7s)
	CommandTimeout time.Duration `json:"command_timeout" yaml:"command_timeout"`

	// CommandSendDelay is the delay before sending first command attempt (default: 200ms)
	CommandSendDelay time.Duration `json:"command_send_delay" yaml:"command_send_delay"`

	// StateUpdateInterval is the interval for state updates (default: 100ms)
	StateUpdateInterval time.Duration `json:"state_update_interval" yaml:"state_update_interval"`

	// VideoBufferSize is the size of video frame buffer (default: 30)
	VideoBufferSize int `json:"video_buffer_size" yaml:"video_buffer_size"`

	// EnableCommandLogging enables logging of all commands sent/received
	EnableCommandLogging bool `json:"enable_command_logging" yaml:"enable_command_logging"`

	// EnableStateLogging enables logging of state updates
	EnableStateLogging bool `json:"enable_state_logging" yaml:"enable_state_logging"`

	// EnableVideoLogging enables logging of video stream statistics
	EnableVideoLogging bool `json:"enable_video_logging" yaml:"enable_video_logging"`
}

// DefaultTransportConfig returns the default transport configuration
func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		DroneHost:            "192.168.10.1:8889",
		LocalCommandAddr:     "0.0.0.0:8889",
		LocalStateAddr:       "0.0.0.0:8890",
		LocalVideoAddr:       "0.0.0.0:11111",
		CommandRetries:       3,
		CommandTimeout:       7 * time.Second,
		CommandSendDelay:     200 * time.Millisecond,
		StateUpdateInterval:  100 * time.Millisecond,
		VideoBufferSize:      30,
		EnableCommandLogging: false,
		EnableStateLogging:   false,
		EnableVideoLogging:   false,
	}
}

// Validate validates the transport configuration
func (c *TransportConfig) Validate() error {
	// Validate DroneHost
	if c.DroneHost == "" {
		return errors.ValidationError("TransportConfig", "drone_host", "cannot be empty")
	}
	if _, _, err := net.SplitHostPort(c.DroneHost); err != nil {
		return errors.ValidationError("TransportConfig", "drone_host",
			fmt.Sprintf("invalid format: %v", err))
	}

	// Validate LocalCommandAddr
	if c.LocalCommandAddr != "" {
		if _, _, err := net.SplitHostPort(c.LocalCommandAddr); err != nil {
			return errors.ValidationError("TransportConfig", "local_command_addr",
				fmt.Sprintf("invalid format: %v", err))
		}
	}

	// Validate LocalStateAddr
	if c.LocalStateAddr != "" {
		if _, _, err := net.SplitHostPort(c.LocalStateAddr); err != nil {
			return errors.ValidationError("TransportConfig", "local_state_addr",
				fmt.Sprintf("invalid format: %v", err))
		}
	}

	// Validate LocalVideoAddr
	if c.LocalVideoAddr != "" {
		if _, _, err := net.SplitHostPort(c.LocalVideoAddr); err != nil {
			return errors.ValidationError("TransportConfig", "local_video_addr",
				fmt.Sprintf("invalid format: %v", err))
		}
	}

	// Validate numeric ranges
	if c.CommandRetries < 0 {
		return errors.ValidationError("TransportConfig", "command_retries",
			"must be >= 0")
	}
	if c.CommandRetries > 10 {
		return errors.ValidationError("TransportConfig", "command_retries",
			"must be <= 10")
	}

	if c.CommandTimeout < 1*time.Second {
		return errors.ValidationError("TransportConfig", "command_timeout",
			"must be >= 1s")
	}
	if c.CommandTimeout > 30*time.Second {
		return errors.ValidationError("TransportConfig", "command_timeout",
			"must be <= 30s")
	}

	if c.CommandSendDelay < 0 {
		return errors.ValidationError("TransportConfig", "command_send_delay",
			"must be >= 0")
	}
	if c.CommandSendDelay > 5*time.Second {
		return errors.ValidationError("TransportConfig", "command_send_delay",
			"must be <= 5s")
	}

	if c.StateUpdateInterval < 10*time.Millisecond {
		return errors.ValidationError("TransportConfig", "state_update_interval",
			"must be >= 10ms")
	}
	if c.StateUpdateInterval > 1*time.Second {
		return errors.ValidationError("TransportConfig", "state_update_interval",
			"must be <= 1s")
	}

	if c.VideoBufferSize < 1 {
		return errors.ValidationError("TransportConfig", "video_buffer_size",
			"must be >= 1")
	}
	if c.VideoBufferSize > 1000 {
		return errors.ValidationError("TransportConfig", "video_buffer_size",
			"must be <= 1000")
	}

	return nil
}

// WithDroneHost sets the drone host and returns the config for chaining
func (c TransportConfig) WithDroneHost(host string) TransportConfig {
	c.DroneHost = host
	return c
}

// WithLocalCommandAddr sets the local command address and returns the config for chaining
func (c TransportConfig) WithLocalCommandAddr(addr string) TransportConfig {
	c.LocalCommandAddr = addr
	return c
}

// WithLocalStateAddr sets the local state address and returns the config for chaining
func (c TransportConfig) WithLocalStateAddr(addr string) TransportConfig {
	c.LocalStateAddr = addr
	return c
}

// WithLocalVideoAddr sets the local video address and returns the config for chaining
func (c TransportConfig) WithLocalVideoAddr(addr string) TransportConfig {
	c.LocalVideoAddr = addr
	return c
}

// WithCommandRetries sets the command retries and returns the config for chaining
func (c TransportConfig) WithCommandRetries(retries int) TransportConfig {
	c.CommandRetries = retries
	return c
}

// WithCommandTimeout sets the command timeout and returns the config for chaining
func (c TransportConfig) WithCommandTimeout(timeout time.Duration) TransportConfig {
	c.CommandTimeout = timeout
	return c
}

// WithCommandSendDelay sets the command send delay and returns the config for chaining
func (c TransportConfig) WithCommandSendDelay(delay time.Duration) TransportConfig {
	c.CommandSendDelay = delay
	return c
}

// WithStateUpdateInterval sets the state update interval and returns the config for chaining
func (c TransportConfig) WithStateUpdateInterval(interval time.Duration) TransportConfig {
	c.StateUpdateInterval = interval
	return c
}

// WithVideoBufferSize sets the video buffer size and returns the config for chaining
func (c TransportConfig) WithVideoBufferSize(size int) TransportConfig {
	c.VideoBufferSize = size
	return c
}

// WithCommandLogging enables/disables command logging and returns the config for chaining
func (c TransportConfig) WithCommandLogging(enabled bool) TransportConfig {
	c.EnableCommandLogging = enabled
	return c
}

// WithStateLogging enables/disables state logging and returns the config for chaining
func (c TransportConfig) WithStateLogging(enabled bool) TransportConfig {
	c.EnableStateLogging = enabled
	return c
}

// WithVideoLogging enables/disables video logging and returns the config for chaining
func (c TransportConfig) WithVideoLogging(enabled bool) TransportConfig {
	c.EnableVideoLogging = enabled
	return c
}
