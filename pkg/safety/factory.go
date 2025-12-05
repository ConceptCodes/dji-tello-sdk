package safety

import (
	"fmt"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// Manager is the interface that safety-wrapped commanders implement.
// This is essentially the same as tello.TelloCommander but defined in safety package
// to avoid circular dependencies.
type Manager interface {
	// Control Commands
	Init() error
	TakeOff() error
	Land() error
	StreamOn() error
	StreamOff() error
	Emergency() error
	Up(distance int) error
	Down(distance int) error
	Left(distance int) error
	Right(distance int) error
	Forward(distance int) error
	Backward(distance int) error
	Clockwise(angle int) error
	CounterClockwise(angle int) error
	Flip(direction string) error
	Go(x, y, z, speed int) error
	Curve(x1, y1, z1, x2, y2, z2, speed int) error

	// Set Commands
	SetSpeed(speed int) error
	SetRcControl(a, b, c, d int) error
	SetWiFiCredentials(ssid, password string) error

	// Read Commands
	GetSpeed() (int, error)
	GetBatteryPercentage() (int, error)
	GetTime() (int, error)
	GetHeight() (int, error)
	GetTemperature() (int, error)
	GetAttitude() (int, int, int, error)
	GetBarometer() (int, error)
	GetAcceleration() (int, int, int, error)
	GetTof() (int, error)

	// Video Commands
	SetVideoFrameCallback(callback func(transport.VideoFrame))
	GetVideoFrameChannel() <-chan transport.VideoFrame

	// Safety-specific methods
	GetSafetyStatus() *SafetyStatus
	GetSafetyEvents() []SafetyEvent
	SetSafetyEnabled(enabled bool)
	SetEmergencyMode(emergency bool)
	SetSafetyConfig(config *Config)
	SetEventCallback(callback func(*SafetyEvent))
	StartTelemetryProcessing(stateChan <-chan *types.State)
	StopTelemetryProcessing()
}

// NewManager creates a safety manager wrapper around a base tello commander
func NewManager(base tello.TelloCommander, config *Config) (Manager, error) {
	if config == nil {
		loadedConfig, source, err := LoadAutoConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load safety config: %w", err)
		}
		config = loadedConfig
		if source != "" {
			utils.Logger.Infof("Using safety configuration: %s", source)
		} else {
			utils.Logger.Info("Using embedded default safety configuration")
		}
	}

	// Create safety manager
	safetyManager := NewSafetyManager(base, config)
	return safetyManager, nil
}

// NewManagerWithPreset creates a safety manager with a preset configuration
func NewManagerWithPreset(base tello.TelloCommander, preset string) (Manager, error) {
	config, err := LoadPresetConfig(preset)
	if err != nil {
		return nil, fmt.Errorf("failed to load safety preset '%s': %w", preset, err)
	}

	return NewManager(base, config)
}

// NewManagerFromConfig creates a safety manager from a config file
func NewManagerFromConfig(base tello.TelloCommander, configPath string) (Manager, error) {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load safety config from '%s': %w", configPath, err)
	}

	return NewManager(base, config)
}

// NewManagerFromProfile creates a safety manager using a profile name
// This is a convenience function that maps to presets
func NewManagerFromProfile(base tello.TelloCommander, profile string) (Manager, error) {
	return NewManagerWithPreset(base, profile)
}
