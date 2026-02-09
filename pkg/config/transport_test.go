package config

import (
	"testing"
	"time"
)

func TestDefaultTransportConfig(t *testing.T) {
	config := DefaultTransportConfig()

	if config.DroneHost != "192.168.10.1:8889" {
		t.Errorf("Default DroneHost = %v, want '192.168.10.1:8889'", config.DroneHost)
	}

	if config.LocalCommandAddr != "0.0.0.0:8889" {
		t.Errorf("Default LocalCommandAddr = %v, want '0.0.0.0:8889'", config.LocalCommandAddr)
	}

	if config.LocalStateAddr != "0.0.0.0:8890" {
		t.Errorf("Default LocalStateAddr = %v, want '0.0.0.0:8890'", config.LocalStateAddr)
	}

	if config.LocalVideoAddr != "0.0.0.0:11111" {
		t.Errorf("Default LocalVideoAddr = %v, want '0.0.0.0:11111'", config.LocalVideoAddr)
	}

	if config.CommandRetries != 3 {
		t.Errorf("Default CommandRetries = %d, want 3", config.CommandRetries)
	}

	if config.CommandTimeout != 7*time.Second {
		t.Errorf("Default CommandTimeout = %v, want 7s", config.CommandTimeout)
	}

	if config.CommandSendDelay != 200*time.Millisecond {
		t.Errorf("Default CommandSendDelay = %v, want 200ms", config.CommandSendDelay)
	}

	if config.StateUpdateInterval != 100*time.Millisecond {
		t.Errorf("Default StateUpdateInterval = %v, want 100ms", config.StateUpdateInterval)
	}

	if config.VideoBufferSize != 30 {
		t.Errorf("Default VideoBufferSize = %d, want 30", config.VideoBufferSize)
	}

	if config.EnableCommandLogging {
		t.Error("Default EnableCommandLogging = true, want false")
	}

	if config.EnableStateLogging {
		t.Error("Default EnableStateLogging = true, want false")
	}

	if config.EnableVideoLogging {
		t.Error("Default EnableVideoLogging = true, want false")
	}
}

func TestTransportConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TransportConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultTransportConfig(),
			wantErr: false,
		},
		{
			name:    "empty drone host",
			config:  DefaultTransportConfig().WithDroneHost(""),
			wantErr: true,
		},
		{
			name:    "invalid drone host format",
			config:  DefaultTransportConfig().WithDroneHost("invalid"),
			wantErr: true,
		},
		{
			name:    "invalid local command addr",
			config:  DefaultTransportConfig().WithLocalCommandAddr("invalid"),
			wantErr: true,
		},
		{
			name:    "invalid local state addr",
			config:  DefaultTransportConfig().WithLocalStateAddr("invalid"),
			wantErr: true,
		},
		{
			name:    "invalid local video addr",
			config:  DefaultTransportConfig().WithLocalVideoAddr("invalid"),
			wantErr: true,
		},
		{
			name:    "negative command retries",
			config:  DefaultTransportConfig().WithCommandRetries(-1),
			wantErr: true,
		},
		{
			name:    "too many command retries",
			config:  DefaultTransportConfig().WithCommandRetries(11),
			wantErr: true,
		},
		{
			name:    "command timeout too short",
			config:  DefaultTransportConfig().WithCommandTimeout(500 * time.Millisecond),
			wantErr: true,
		},
		{
			name:    "command timeout too long",
			config:  DefaultTransportConfig().WithCommandTimeout(31 * time.Second),
			wantErr: true,
		},
		{
			name:    "negative command send delay",
			config:  DefaultTransportConfig().WithCommandSendDelay(-1 * time.Millisecond),
			wantErr: true,
		},
		{
			name:    "command send delay too long",
			config:  DefaultTransportConfig().WithCommandSendDelay(6 * time.Second),
			wantErr: true,
		},
		{
			name:    "state update interval too short",
			config:  DefaultTransportConfig().WithStateUpdateInterval(5 * time.Millisecond),
			wantErr: true,
		},
		{
			name:    "state update interval too long",
			config:  DefaultTransportConfig().WithStateUpdateInterval(2 * time.Second),
			wantErr: true,
		},
		{
			name:    "video buffer size too small",
			config:  DefaultTransportConfig().WithVideoBufferSize(0),
			wantErr: true,
		},
		{
			name:    "video buffer size too large",
			config:  DefaultTransportConfig().WithVideoBufferSize(1001),
			wantErr: true,
		},
		{
			name: "valid custom config",
			config: TransportConfig{
				DroneHost:            "192.168.10.1:8889",
				LocalCommandAddr:     "0.0.0.0:8889",
				LocalStateAddr:       "0.0.0.0:8890",
				LocalVideoAddr:       "0.0.0.0:11111",
				CommandRetries:       5,
				CommandTimeout:       10 * time.Second,
				CommandSendDelay:     100 * time.Millisecond,
				StateUpdateInterval:  50 * time.Millisecond,
				VideoBufferSize:      50,
				EnableCommandLogging: true,
				EnableStateLogging:   true,
				EnableVideoLogging:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TransportConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransportConfig_WithMethods(t *testing.T) {
	config := DefaultTransportConfig()

	// Test WithDroneHost
	config = config.WithDroneHost("10.0.0.1:9999")
	if config.DroneHost != "10.0.0.1:9999" {
		t.Errorf("WithDroneHost() = %v, want '10.0.0.1:9999'", config.DroneHost)
	}

	// Test WithLocalCommandAddr
	config = config.WithLocalCommandAddr("127.0.0.1:8889")
	if config.LocalCommandAddr != "127.0.0.1:8889" {
		t.Errorf("WithLocalCommandAddr() = %v, want '127.0.0.1:8889'", config.LocalCommandAddr)
	}

	// Test WithLocalStateAddr
	config = config.WithLocalStateAddr("127.0.0.1:8890")
	if config.LocalStateAddr != "127.0.0.1:8890" {
		t.Errorf("WithLocalStateAddr() = %v, want '127.0.0.1:8890'", config.LocalStateAddr)
	}

	// Test WithLocalVideoAddr
	config = config.WithLocalVideoAddr("127.0.0.1:11111")
	if config.LocalVideoAddr != "127.0.0.1:11111" {
		t.Errorf("WithLocalVideoAddr() = %v, want '127.0.0.1:11111'", config.LocalVideoAddr)
	}

	// Test WithCommandRetries
	config = config.WithCommandRetries(5)
	if config.CommandRetries != 5 {
		t.Errorf("WithCommandRetries() = %d, want 5", config.CommandRetries)
	}

	// Test WithCommandTimeout
	config = config.WithCommandTimeout(5 * time.Second)
	if config.CommandTimeout != 5*time.Second {
		t.Errorf("WithCommandTimeout() = %v, want 5s", config.CommandTimeout)
	}

	// Test WithCommandSendDelay
	config = config.WithCommandSendDelay(100 * time.Millisecond)
	if config.CommandSendDelay != 100*time.Millisecond {
		t.Errorf("WithCommandSendDelay() = %v, want 100ms", config.CommandSendDelay)
	}

	// Test WithStateUpdateInterval
	config = config.WithStateUpdateInterval(200 * time.Millisecond)
	if config.StateUpdateInterval != 200*time.Millisecond {
		t.Errorf("WithStateUpdateInterval() = %v, want 200ms", config.StateUpdateInterval)
	}

	// Test WithVideoBufferSize
	config = config.WithVideoBufferSize(50)
	if config.VideoBufferSize != 50 {
		t.Errorf("WithVideoBufferSize() = %d, want 50", config.VideoBufferSize)
	}

	// Test WithCommandLogging
	config = config.WithCommandLogging(true)
	if !config.EnableCommandLogging {
		t.Error("WithCommandLogging(true) = false, want true")
	}

	// Test WithStateLogging
	config = config.WithStateLogging(true)
	if !config.EnableStateLogging {
		t.Error("WithStateLogging(true) = false, want true")
	}

	// Test WithVideoLogging
	config = config.WithVideoLogging(true)
	if !config.EnableVideoLogging {
		t.Error("WithVideoLogging(true) = false, want true")
	}

	// Validate the final config
	if err := config.Validate(); err != nil {
		t.Errorf("Final config validation failed: %v", err)
	}
}

func TestTransportConfig_Chaining(t *testing.T) {
	// Test method chaining
	config := DefaultTransportConfig().
		WithDroneHost("10.0.0.1:9999").
		WithLocalCommandAddr("127.0.0.1:8889").
		WithCommandRetries(5).
		WithCommandTimeout(5 * time.Second).
		WithCommandLogging(true)

	if config.DroneHost != "10.0.0.1:9999" {
		t.Errorf("Chained WithDroneHost() = %v, want '10.0.0.1:9999'", config.DroneHost)
	}

	if config.LocalCommandAddr != "127.0.0.1:8889" {
		t.Errorf("Chained WithLocalCommandAddr() = %v, want '127.0.0.1:8889'", config.LocalCommandAddr)
	}

	if config.CommandRetries != 5 {
		t.Errorf("Chained WithCommandRetries() = %d, want 5", config.CommandRetries)
	}

	if config.CommandTimeout != 5*time.Second {
		t.Errorf("Chained WithCommandTimeout() = %v, want 5s", config.CommandTimeout)
	}

	if !config.EnableCommandLogging {
		t.Error("Chained WithCommandLogging(true) = false, want true")
	}

	// Validate the chained config
	if err := config.Validate(); err != nil {
		t.Errorf("Chained config validation failed: %v", err)
	}
}
