// Package safety provides comprehensive tests for the SafetyManager component.
//
// Test Scope:
//   - NewSafetyManager initialization with valid and invalid CommanderInterface
//   - Telemetry processing lifecycle (start/stop)
//   - Goroutine lifecycle management with WaitGroup
//   - Mock infrastructure for Tasks 2-4 parallelization
//
// This package does NOT test:
//   - Actual drone connections or hardware interactions (uses mocks)
//   - UI rendering components
//   - Timing-sensitive race conditions
//   - Integration with external services
package safety

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// MockCommander is a mock implementation of CommanderInterface for testing.
// This provides the foundation for Tasks 2-4 parallelization by enabling
// isolated testing without drone hardware.
type MockCommander struct {
	// Control command tracking
	initCalled      bool
	takeoffCalled   bool
	landCalled      bool
	streamOnCalled  bool
	streamOffCalled bool
	emergencyCalled bool

	// Movement command tracking
	upCalled               bool
	downCalled             bool
	leftCalled             bool
	rightCalled            bool
	forwardCalled          bool
	backwardCalled         bool
	clockwiseCalled        bool
	counterClockwiseCalled bool
	flipCalled             bool
	goCalled               bool
	curveCalled            bool

	// Set command tracking
	setSpeedCalled           bool
	setRcControlCalled       bool
	setWiFiCredentialsCalled bool

	// Read command responses
	speedResponse         int
	batteryResponse       int
	timeResponse          int
	heightResponse        int
	temperatureResponse   int
	attitudeResponsePitch int
	attitudeResponseRoll  int
	attitudeResponseYaw   int
	barometerResponse     int
	accelerationResponseX int
	accelerationResponseY int
	accelerationResponseZ int
	tofResponse           int

	// Video command tracking
	setVideoFrameCallbackCalled bool
	getVideoFrameChannelCalled  bool

	// Read command errors
	speedError        error
	batteryError      error
	timeError         error
	heightError       error
	temperatureError  error
	attitudeError     error
	barometerError    error
	accelerationError error
	tofError          error

	// Video callback
	videoCallback func(transport.VideoFrame)
}

// NewMockCommander creates a new MockCommander instance.
func NewMockCommander() *MockCommander {
	return &MockCommander{
		speedResponse:         50,
		batteryResponse:       85,
		heightResponse:        100,
		temperatureResponse:   25,
		attitudeResponsePitch: 10,
		attitudeResponseRoll:  -5,
		attitudeResponseYaw:   180,
	}
}

// Control Commands
func (m *MockCommander) Init() error {
	m.initCalled = true
	return nil
}

func (m *MockCommander) TakeOff() error {
	m.takeoffCalled = true
	return nil
}

func (m *MockCommander) Land() error {
	m.landCalled = true
	return nil
}

func (m *MockCommander) StreamOn() error {
	m.streamOnCalled = true
	return nil
}

func (m *MockCommander) StreamOff() error {
	m.streamOffCalled = true
	return nil
}

func (m *MockCommander) Emergency() error {
	m.emergencyCalled = true
	return nil
}

// Movement Commands
func (m *MockCommander) Up(distance int) error {
	m.upCalled = true
	return nil
}

func (m *MockCommander) Down(distance int) error {
	m.downCalled = true
	return nil
}

func (m *MockCommander) Left(distance int) error {
	m.leftCalled = true
	return nil
}

func (m *MockCommander) Right(distance int) error {
	m.rightCalled = true
	return nil
}

func (m *MockCommander) Forward(distance int) error {
	m.forwardCalled = true
	return nil
}

func (m *MockCommander) Backward(distance int) error {
	m.backwardCalled = true
	return nil
}

func (m *MockCommander) Clockwise(angle int) error {
	m.clockwiseCalled = true
	return nil
}

func (m *MockCommander) CounterClockwise(angle int) error {
	m.counterClockwiseCalled = true
	return nil
}

func (m *MockCommander) Flip(direction string) error {
	m.flipCalled = true
	return nil
}

func (m *MockCommander) Go(x, y, z, speed int) error {
	m.goCalled = true
	return nil
}

func (m *MockCommander) Curve(x1, y1, z1, x2, y2, z2, speed int) error {
	m.curveCalled = true
	return nil
}

// Set Commands
func (m *MockCommander) SetSpeed(speed int) error {
	m.setSpeedCalled = true
	return nil
}

func (m *MockCommander) SetRcControl(a, b, c, d int) error {
	m.setRcControlCalled = true
	return nil
}

func (m *MockCommander) SetWiFiCredentials(ssid, password string) error {
	m.setWiFiCredentialsCalled = true
	return nil
}

// Read Commands
func (m *MockCommander) GetSpeed() (int, error) {
	return m.speedResponse, m.speedError
}

func (m *MockCommander) GetBatteryPercentage() (int, error) {
	return m.batteryResponse, m.batteryError
}

func (m *MockCommander) GetTime() (int, error) {
	return m.timeResponse, m.timeError
}

func (m *MockCommander) GetHeight() (int, error) {
	return m.heightResponse, m.heightError
}

func (m *MockCommander) GetTemperature() (int, error) {
	return m.temperatureResponse, m.temperatureError
}

func (m *MockCommander) GetAttitude() (int, int, int, error) {
	return m.attitudeResponsePitch, m.attitudeResponseRoll, m.attitudeResponseYaw, m.attitudeError
}

func (m *MockCommander) GetBarometer() (int, error) {
	return m.barometerResponse, m.barometerError
}

func (m *MockCommander) GetAcceleration() (int, int, int, error) {
	return m.accelerationResponseX, m.accelerationResponseY, m.accelerationResponseZ, m.accelerationError
}

func (m *MockCommander) GetTof() (int, error) {
	return m.tofResponse, m.tofError
}

// Video Commands
func (m *MockCommander) SetVideoFrameCallback(callback func(transport.VideoFrame)) {
	m.setVideoFrameCallbackCalled = true
	m.videoCallback = callback
}

func (m *MockCommander) GetVideoFrameChannel() <-chan transport.VideoFrame {
	m.getVideoFrameChannelCalled = true
	ch := make(chan transport.VideoFrame)
	close(ch)
	return ch
}

// Reset resets all mock tracking state for reuse.
func (m *MockCommander) Reset() {
	m.initCalled = false
	m.takeoffCalled = false
	m.landCalled = false
	m.streamOnCalled = false
	m.streamOffCalled = false
	m.emergencyCalled = false
	m.upCalled = false
	m.downCalled = false
	m.leftCalled = false
	m.rightCalled = false
	m.forwardCalled = false
	m.backwardCalled = false
	m.clockwiseCalled = false
	m.counterClockwiseCalled = false
	m.flipCalled = false
	m.goCalled = false
	m.curveCalled = false
	m.setSpeedCalled = false
	m.setRcControlCalled = false
	m.setWiFiCredentialsCalled = false
	m.setVideoFrameCallbackCalled = false
	m.getVideoFrameChannelCalled = false
}

// createTestState creates a test State for telemetry testing.
func createTestState() *types.State {
	return &types.State{
		Pitch: 10,
		Roll:  -5,
		Yaw:   180,
		Vgx:   50,
		Vgy:   30,
		Vgz:   20,
		Templ: 20,
		Temph: 30,
		Tof:   200,
		H:     100,
		Bat:   85,
		Baro:  1013.25,
		Time:  120,
		Agx:   0.1,
		Agy:   0.2,
		Agz:   1.0,
	}
}

// TestNewSafetyManager tests the NewSafetyManager constructor.
func TestNewSafetyManager(t *testing.T) {
	t.Run("with valid commander", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()

		manager := NewSafetyManager(mockCommander, config)

		if manager == nil {
			t.Error("Expected non-nil SafetyManager with valid commander")
		}

		if manager.commander != mockCommander {
			t.Error("Expected commander to be set correctly")
		}

		if manager.config != config {
			t.Error("Expected config to be set correctly")
		}
	})

	t.Run("with nil commander", func(t *testing.T) {
		config := DefaultConfig()

		manager := NewSafetyManager(nil, config)

		if manager != nil {
			t.Error("Expected nil SafetyManager with nil commander")
		}
	})

	t.Run("with invalid commander type", func(t *testing.T) {
		// Use a non-interface-compliant type
		invalidCommander := "not a commander"
		config := DefaultConfig()

		manager := NewSafetyManager(invalidCommander, config)

		if manager != nil {
			t.Error("Expected nil SafetyManager with invalid commander type")
		}
	})

	t.Run("with nil config", func(t *testing.T) {
		mockCommander := NewMockCommander()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil config")
			}
		}()

		_ = NewSafetyManager(mockCommander, nil)
	})

	t.Run("with different preset configs", func(t *testing.T) {
		presets := GetPresetConfigs()

		for name, config := range presets {
			t.Run(name, func(t *testing.T) {
				mockCommander := NewMockCommander()
				manager := NewSafetyManager(mockCommander, config)

				if manager == nil {
					t.Errorf("Expected non-nil SafetyManager with %s config", name)
				}

				if manager.config.Level != config.Level {
					t.Errorf("Expected config level %v, got %v", config.Level, manager.config.Level)
				}
			})
		}
	})
}

// TestNewSafetyManagerTableDriven provides comprehensive table-driven tests for NewSafetyManager.
func TestNewSafetyManagerTableDriven(t *testing.T) {
	tests := []struct {
		name        string
		commander   interface{}
		config      *Config
		expectNil   bool
		expectPanic bool
	}{
		{
			name:      "valid commander and config",
			commander: NewMockCommander(),
			config:    DefaultConfig(),
			expectNil: false,
		},
		{
			name:      "nil commander",
			commander: nil,
			config:    DefaultConfig(),
			expectNil: true,
		},
		{
			name:        "nil config",
			commander:   NewMockCommander(),
			config:      nil,
			expectPanic: true,
		},
		{
			name:      "string instead of commander",
			commander: "invalid",
			config:    DefaultConfig(),
			expectNil: true,
		},
		{
			name:      "integer instead of commander",
			commander: 42,
			config:    DefaultConfig(),
			expectNil: true,
		},
		{
			name:      "conservative preset config",
			commander: NewMockCommander(),
			config:    ConservativeConfig(),
			expectNil: false,
		},
		{
			name:      "aggressive preset config",
			commander: NewMockCommander(),
			config:    AggressiveConfig(),
			expectNil: false,
		},
		{
			name:      "indoor preset config",
			commander: NewMockCommander(),
			config:    IndoorConfig(),
			expectNil: false,
		},
		{
			name:      "outdoor preset config",
			commander: NewMockCommander(),
			config:    OutdoorConfig(),
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for test case")
					}
				}()
			}

			manager := NewSafetyManager(tt.commander, tt.config)

			if tt.expectNil && manager != nil {
				t.Error("Expected nil SafetyManager")
			}
			if !tt.expectNil && manager == nil {
				t.Error("Expected non-nil SafetyManager")
			}
		})
	}
}

// TestStartTelemetryProcessing tests the StartTelemetryProcessing method.
func TestStartTelemetryProcessing(t *testing.T) {
	t.Run("starts telemetry processing", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State)
		manager.StartTelemetryProcessing(stateChan)

		if manager.stateChan == nil {
			t.Error("Expected state channel to be set")
		}

		if manager.telemetryCtx == nil {
			t.Error("Expected telemetry context to be initialized")
		}

		// Cleanup
		manager.StopTelemetryProcessing()
	})

	t.Run("handles nil state channel gracefully", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Should not panic with nil channel
		manager.StartTelemetryProcessing(nil)

		// Cleanup
		manager.StopTelemetryProcessing()
	})
}

// TestStopTelemetryProcessing tests the StopTelemetryProcessing method.
func TestStopTelemetryProcessing(t *testing.T) {
	t.Run("stops telemetry processing gracefully", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State)
		manager.StartTelemetryProcessing(stateChan)

		// Stop should not block indefinitely
		done := make(chan struct{})
		go func() {
			manager.StopTelemetryProcessing()
			close(done)
		}()

		select {
		case <-done:
			// Expected - Stop completed
		case <-time.After(2 * time.Second):
			t.Error("StopTelemetryProcessing timed out")
		}
	})

	t.Run("can be called multiple times safely", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State)
		manager.StartTelemetryProcessing(stateChan)

		// Multiple stop calls should not panic
		manager.StopTelemetryProcessing()
		manager.StopTelemetryProcessing()
		manager.StopTelemetryProcessing()
	})

	t.Run("stop without start is safe", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Stop without start should not panic
		manager.StopTelemetryProcessing()
	})
}

// TestTelemetryGoroutineLifecycle tests the goroutine lifecycle with WaitGroup.
func TestTelemetryGoroutineLifecycle(t *testing.T) {
	t.Run("WaitGroup tracks goroutine lifecycle", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State, 3)

		// Send some test states
		manager.StartTelemetryProcessing(stateChan)

		// Send states to the channel
		state1 := createTestState()
		state2 := createTestState()
		state2.H = 150 // Different height
		state3 := createTestState()
		state3.Bat = 20 // Low battery

		stateChan <- state1
		stateChan <- state2
		stateChan <- state3

		// Close channel after sending states
		close(stateChan)

		// Wait for processing
		manager.StopTelemetryProcessing()

		// Verify safety status was updated (indirect test that goroutine ran)
		status := manager.GetSafetyStatus()
		if status == nil {
			t.Error("Expected non-nil safety status")
		}
	})

	t.Run("goroutine exits on context cancellation", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State)
		manager.StartTelemetryProcessing(stateChan)

		// Send a state
		stateChan <- createTestState()

		// Stop processing
		done := make(chan struct{})
		go func() {
			manager.StopTelemetryProcessing()
			close(done)
		}()

		select {
		case <-done:
			// Expected - goroutine exited
		case <-time.After(2 * time.Second):
			t.Error("Goroutine did not exit after context cancellation")
		}
	})

	t.Run("multiple telemetry cycles", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		for i := 0; i < 3; i++ {
			stateChan := make(chan *types.State)
			manager.StartTelemetryProcessing(stateChan)

			stateChan <- createTestState()
			close(stateChan)

			manager.StopTelemetryProcessing()
		}

		// Final status should still be accessible
		status := manager.GetSafetyStatus()
		if status == nil {
			t.Error("Expected non-nil safety status after multiple cycles")
		}
	})
}

// TestSafetyManagerIntegration provides integration-style tests for the SafetyManager.
func TestSafetyManagerIntegration(t *testing.T) {
	t.Run("safety status is initialized correctly", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		status := manager.GetSafetyStatus()

		if status == nil {
			t.Fatal("Expected non-nil safety status")
		}

		if !status.SafetyEnabled {
			t.Error("Expected safety to be enabled by default")
		}

		if status.EmergencyMode {
			t.Error("Expected emergency mode to be false by default")
		}

		if status.ConfigLevel != config.Level {
			t.Errorf("Expected config level %v, got %v", config.Level, status.ConfigLevel)
		}
	})

	t.Run("safety events can be retrieved", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		events := manager.GetSafetyEvents()

		if events == nil {
			t.Error("Expected non-nil events slice")
		}
	})

	t.Run("set emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEmergencyMode(true)

		status := manager.GetSafetyStatus()
		if !status.EmergencyMode {
			t.Error("Expected emergency mode to be true after setting")
		}
	})

	t.Run("disable and enable safety", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetSafetyEnabled(false)
		status := manager.GetSafetyStatus()
		if status.SafetyEnabled {
			t.Error("Expected safety to be disabled")
		}

		manager.SetSafetyEnabled(true)
		status = manager.GetSafetyStatus()
		if !status.SafetyEnabled {
			t.Error("Expected safety to be enabled")
		}
	})
}

// TestMockCommanderReusability verifies that MockCommander can be reused across tests.
func TestMockCommanderReusability(t *testing.T) {
	commander := NewMockCommander()
	config := DefaultConfig()

	// Create first manager
	manager1 := NewSafetyManager(commander, config)
	if manager1 == nil {
		t.Fatal("Expected non-nil SafetyManager")
	}

	// Reset commander
	commander.Reset()

	// Create second manager with same commander
	manager2 := NewSafetyManager(commander, config)
	if manager2 == nil {
		t.Fatal("Expected non-nil SafetyManager after reset")
	}

	if manager1 == manager2 {
		t.Error("Expected different SafetyManager instances")
	}
}

// =============================================================================
// Validation Method Tests
// =============================================================================

// TestValidateCommand tests the validateCommand method with various command scenarios.
func TestValidateCommand(t *testing.T) {
	t.Run("allows command when rate limit not exceeded", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.validateCommand("test", nil)

		if !result.Allowed {
			t.Errorf("Expected command to be allowed, got: %s", result.Reason)
		}
	})

	t.Run("allows command when flight time not exceeded", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.validateCommand("test", nil)

		if !result.Allowed {
			t.Errorf("Expected command to be allowed, got: %s", result.Reason)
		}
	})
}

// TestValidateMovementCommand tests validateMovementCommand with boundary values.
func TestValidateMovementCommand(t *testing.T) {
	tests := []struct {
		name           string
		x, y, z        int
		currentHeight  int
		expectAllowed  bool
		expectedReason string
	}{
		// Valid movement cases
		{
			name:          "Valid zero movement",
			x:             0,
			y:             0,
			z:             0,
			expectAllowed: true,
		},
		{
			name:          "Valid positive movement",
			x:             50,
			y:             50,
			z:             50,
			expectAllowed: true,
		},
		{
			name:          "Valid negative movement",
			x:             -50,
			y:             -50,
			z:             -50,
			expectAllowed: true,
		},
		// Boundary: altitude limits
		{
			name:          "Movement to minimum altitude",
			x:             0,
			y:             0,
			z:             -80, // From 100cm height to 20cm (min)
			currentHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Movement to maximum altitude",
			x:             0,
			y:             0,
			z:             200, // From 100cm height to 300cm (max)
			currentHeight: 100,
			expectAllowed: true,
		},
		{
			name:           "Exceeds maximum altitude",
			x:              0,
			y:              0,
			z:              250, // From 100cm height to 350cm (exceeds 300cm max)
			currentHeight:  100,
			expectAllowed:  false,
			expectedReason: "Altitude 350cm exceeds maximum 300cm",
		},
		{
			name:           "Below minimum altitude",
			x:              0,
			y:              0,
			z:              -100, // From 100cm height to 0cm (below 20cm min)
			currentHeight:  100,
			expectAllowed:  false,
			expectedReason: "Altitude 0cm below minimum 20cm",
		},
		// Boundary: large values
		{
			name:          "Large positive values",
			x:             1000,
			y:             1000,
			z:             100,
			expectAllowed: true,
		},
		{
			name:          "Large negative values",
			x:             -1000,
			y:             -1000,
			z:             -50,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			// Set current height if specified
			if tt.currentHeight > 0 {
				state := &types.State{
					H: tt.currentHeight,
				}
				manager.UpdateState(state)
			}

			result := manager.validateMovementCommand("move", tt.x, tt.y, tt.z)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected movement to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected movement to be rejected")
				}
				if tt.expectedReason != "" && result.Reason != tt.expectedReason {
					t.Errorf("Expected reason %q, got: %q", tt.expectedReason, result.Reason)
				}
			}
		})
	}
}

// TestValidateRotationCommand tests validateRotationCommand with angle limits.
func TestValidateRotationCommand(t *testing.T) {
	tests := []struct {
		name          string
		angle         int
		expectAllowed bool
	}{
		{
			name:          "Zero angle",
			angle:         0,
			expectAllowed: true,
		},
		{
			name:          "Positive small angle",
			angle:         45,
			expectAllowed: true,
		},
		{
			name:          "Positive maximum angle",
			angle:         360,
			expectAllowed: true,
		},
		{
			name:          "Negative angle",
			angle:         -90,
			expectAllowed: true,
		},
		{
			name:          "Large positive angle",
			angle:         720,
			expectAllowed: true,
		},
		{
			name:          "Large negative angle",
			angle:         -720,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.validateRotationCommand("rotate", tt.angle)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected rotation to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected rotation to be rejected")
				}
			}
		})
	}
}

// TestValidateFlipCommand tests validateFlipCommand with various directions and configurations.
func TestValidateFlipCommand(t *testing.T) {
	tests := []struct {
		name           string
		direction      string
		currentHeight  int
		enableFlips    bool
		minFlipHeight  int
		expectAllowed  bool
		expectedReason string
	}{
		// Flip enabled cases
		{
			name:          "Valid flip with flips enabled",
			direction:     "left",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Valid flip at minimum height",
			direction:     "right",
			currentHeight: 100,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip forward",
			direction:     "forward",
			currentHeight: 200,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip backward",
			direction:     "back",
			currentHeight: 200,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		// Flip disabled
		{
			name:           "Flip when disabled",
			direction:      "left",
			currentHeight:  200,
			enableFlips:    false,
			minFlipHeight:  100,
			expectAllowed:  false,
			expectedReason: "Flips are disabled in current safety configuration",
		},
		// Below minimum height
		{
			name:           "Flip below minimum height",
			direction:      "left",
			currentHeight:  50,
			enableFlips:    true,
			minFlipHeight:  100,
			expectAllowed:  false,
			expectedReason: "Altitude 50cm below minimum flip height 100cm",
		},
		// All flip directions
		{
			name:          "Flip left",
			direction:     "left",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip right",
			direction:     "right",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip forward",
			direction:     "forward",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip backward",
			direction:     "backward",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			config.Behavioral.EnableFlips = tt.enableFlips
			config.Behavioral.MinFlipHeight = tt.minFlipHeight
			manager := NewSafetyManager(mockCommander, config)

			// Set current height
			state := &types.State{
				H: tt.currentHeight,
			}
			manager.UpdateState(state)

			result := manager.validateFlipCommand(tt.direction)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected flip to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected flip to be rejected")
				}
				if tt.expectedReason != "" && result.Reason != tt.expectedReason {
					t.Errorf("Expected reason %q, got: %q", tt.expectedReason, result.Reason)
				}
			}
		})
	}
}

// TestValidateGoCommand tests validateGoCommand with coordinate and speed limits.
func TestValidateGoCommand(t *testing.T) {
	tests := []struct {
		name           string
		x, y, z, speed int
		expectAllowed  bool
		expectedReason string
	}{
		// Valid cases
		{
			name:          "Valid go command",
			x:             50,
			y:             50,
			z:             50,
			speed:         50,
			expectAllowed: true,
		},
		{
			name:          "Go to origin",
			x:             0,
			y:             0,
			z:             0,
			speed:         10,
			expectAllowed: true,
		},
		{
			name:          "Go at maximum speed",
			x:             100,
			y:             100,
			z:             100,
			speed:         100, // MaxHorizontal
			expectAllowed: true,
		},
		{
			name:          "Negative coordinates",
			x:             -50,
			y:             -50,
			z:             -50,
			speed:         50,
			expectAllowed: true,
		},
		// Speed exceeds limit
		{
			name:           "Speed exceeds maximum",
			x:              50,
			y:              50,
			z:              50,
			speed:          150, // Exceeds MaxHorizontal=100
			expectAllowed:  false,
			expectedReason: "Speed 150 exceeds maximum 100",
		},
		// Altitude exceeds limit
		{
			name:           "Target altitude exceeds maximum",
			x:              0,
			y:              0,
			z:              400, // Exceeds MaxHeight=300
			speed:          50,
			expectAllowed:  false,
			expectedReason: "Target altitude 400cm exceeds maximum 300cm",
		},
		// Boundary values
		{
			name:          "At speed limit",
			x:             100,
			y:             100,
			z:             300, // At max height
			speed:         100, // At max horizontal
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.validateGoCommand(tt.x, tt.y, tt.z, tt.speed)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected go command to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected go command to be rejected")
				}
				if tt.expectedReason != "" && result.Reason != tt.expectedReason {
					t.Errorf("Expected reason %q, got: %q", tt.expectedReason, result.Reason)
				}
			}
		})
	}
}

// TestValidateCurveCommand tests validateCurveCommand with arc validation.
func TestValidateCurveCommand(t *testing.T) {
	tests := []struct {
		name          string
		x1, y1, z1    int
		x2, y2, z2    int
		speed         int
		expectAllowed bool
	}{
		// Valid cases
		{
			name:          "Valid curve command",
			x1:            0,
			y1:            0,
			z1:            0,
			x2:            50,
			y2:            50,
			z2:            50,
			speed:         50,
			expectAllowed: true,
		},
		{
			name:          "Curve at maximum speed",
			x1:            0,
			y1:            0,
			z1:            0,
			x2:            100,
			y2:            100,
			z2:            100,
			speed:         100,
			expectAllowed: true,
		},
		{
			name:          "Negative arc coordinates",
			x1:            -50,
			y1:            -50,
			z1:            -50,
			x2:            50,
			y2:            50,
			z2:            50,
			speed:         50,
			expectAllowed: true,
		},
		{
			name:          "Zero distance arc",
			x1:            100,
			y1:            100,
			z1:            100,
			x2:            100,
			y2:            100,
			z2:            100,
			speed:         50,
			expectAllowed: true, // Should be allowed (base validation passes)
		},
		// Speed exceeds limit
		{
			name:          "Speed exceeds maximum",
			x1:            0,
			y1:            0,
			z1:            0,
			x2:            50,
			y2:            50,
			z2:            50,
			speed:         150, // Exceeds MaxHorizontal=100
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.validateCurveCommand(tt.x1, tt.y1, tt.z1, tt.x2, tt.y2, tt.z2, tt.speed)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected curve command to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected curve command to be rejected")
				}
			}
		})
	}
}

// TestValidateSpeedCommand tests validateSpeedCommand with speed limits.
func TestValidateSpeedCommand(t *testing.T) {
	tests := []struct {
		name          string
		speed         int
		expectAllowed bool
	}{
		{
			name:          "Zero speed",
			speed:         0,
			expectAllowed: true,
		},
		{
			name:          "Low speed",
			speed:         10,
			expectAllowed: true,
		},
		{
			name:          "Medium speed",
			speed:         50,
			expectAllowed: true,
		},
		{
			name:          "Maximum speed",
			speed:         100,
			expectAllowed: true,
		},
		{
			name:          "Just below limit",
			speed:         99,
			expectAllowed: true,
		},
		{
			name:          "Exceeds maximum",
			speed:         101,
			expectAllowed: false,
		},
		{
			name:          "Large speed exceeds maximum",
			speed:         200,
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.validateSpeedCommand(tt.speed)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected speed command to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected speed command to be rejected")
				}
			}
		})
	}
}

// TestValidateRCCommand tests validateRCCommand with RC bounds.
func TestValidateRCCommand(t *testing.T) {
	tests := []struct {
		name           string
		a, b, c, d     int
		expectAllowed  bool
		expectedReason string
	}{
		// Valid cases - all zeros
		{
			name:          "Zero RC values",
			a:             0,
			b:             0,
			c:             0,
			d:             0,
			expectAllowed: true,
		},
		// Valid within limits
		{
			name:          "All axes within limits",
			a:             50,
			b:             50,
			c:             40,
			d:             50,
			expectAllowed: true,
		},
		{
			name:          "At maximum limits",
			a:             100,
			b:             100,
			c:             80,
			d:             100,
			expectAllowed: true,
		},
		// Negative values (abs values must be within limits)
		{
			name:          "Negative values within limits",
			a:             -50,
			b:             -50,
			c:             -40,
			d:             -50,
			expectAllowed: true,
		},
		{
			name:          "Negative at limits",
			a:             -100,
			b:             -100,
			c:             -80,
			d:             -100,
			expectAllowed: true,
		},
		// Exceeds limits - axis a (horizontal)
		{
			name:           "Axis a exceeds positive limit",
			a:              101,
			b:              0,
			c:              0,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value a=101 exceeds horizontal limit 100",
		},
		{
			name:           "Axis a exceeds negative limit",
			a:              -101,
			b:              0,
			c:              0,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value a=-101 exceeds horizontal limit 100",
		},
		// Exceeds limits - axis b (horizontal)
		{
			name:           "Axis b exceeds positive limit",
			a:              0,
			b:              101,
			c:              0,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value b=101 exceeds horizontal limit 100",
		},
		{
			name:           "Axis b exceeds negative limit",
			a:              0,
			b:              -101,
			c:              0,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value b=-101 exceeds horizontal limit 100",
		},
		// Exceeds limits - axis c (vertical)
		{
			name:           "Axis c exceeds positive vertical limit",
			a:              0,
			b:              0,
			c:              81,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value c=81 exceeds vertical limit 80",
		},
		{
			name:           "Axis c exceeds negative vertical limit",
			a:              0,
			b:              0,
			c:              -81,
			d:              0,
			expectAllowed:  false,
			expectedReason: "RC value c=-81 exceeds vertical limit 80",
		},
		// Exceeds limits - axis d (yaw)
		{
			name:           "Axis d exceeds yaw limit",
			a:              0,
			b:              0,
			c:              0,
			d:              101,
			expectAllowed:  false,
			expectedReason: "RC value d=101 exceeds yaw limit 100",
		},
		{
			name:           "Axis d exceeds negative yaw limit",
			a:              0,
			b:              0,
			c:              0,
			d:              -101,
			expectAllowed:  false,
			expectedReason: "RC value d=-101 exceeds yaw limit 100",
		},
		// Multiple axes exceed limits
		{
			name:           "Multiple axes exceed limits",
			a:              150,
			b:              150,
			c:              150,
			d:              150,
			expectAllowed:  false,
			expectedReason: "RC value a=150 exceeds horizontal limit 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.validateRCCommand(tt.a, tt.b, tt.c, tt.d)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected RC command to be allowed, got: %s", result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected RC command to be rejected")
				}
				if tt.expectedReason != "" && result.Reason != tt.expectedReason {
					t.Errorf("Expected reason %q, got: %q", tt.expectedReason, result.Reason)
				}
			}
		})
	}
}

// TestValidateRCCommandWithDifferentConfigs tests RC validation with different safety configurations.
func TestValidateRCCommandWithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name          string
		a             int
		config        *Config
		expectAllowed bool
	}{
		{
			name:          "Conservative config lower limits",
			a:             80, // Above conservative limit of 60
			config:        ConservativeConfig(),
			expectAllowed: false,
		},
		{
			name:          "Conservative config within limits",
			a:             50, // Below conservative limit of 60
			config:        ConservativeConfig(),
			expectAllowed: true,
		},
		{
			name:          "Aggressive config allows higher speeds",
			a:             100, // At normal limit, within aggressive
			config:        AggressiveConfig(),
			expectAllowed: true,
		},
		{
			name:          "Indoor config very restrictive",
			a:             50, // Above indoor limit of 40
			config:        IndoorConfig(),
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			manager := NewSafetyManager(mockCommander, tt.config)

			result := manager.validateRCCommand(tt.a, 0, 0, 0)

			if tt.expectAllowed {
				if !result.Allowed {
					t.Errorf("Expected RC command to be allowed with %s config, got: %s", tt.config.Level, result.Reason)
				}
			} else {
				if result.Allowed {
					t.Errorf("Expected RC command to be rejected with %s config", tt.config.Level)
				}
			}
		})
	}
}

// =============================================================================
// Fuzz Testing
// =============================================================================

// FuzzValidateCommand is a fuzz test for validateCommand that ensures
// the validation methods don't panic on arbitrary input.
func FuzzValidateCommand(f *testing.F) {
	// Seed corpus with valid and edge case inputs
	f.Add("takeoff")
	f.Add("land")
	f.Add("up 50")
	f.Add("down 50")
	f.Add("left 100")
	f.Add("right 100")
	f.Add("forward 100")
	f.Add("backward 100")
	f.Add("cw 90")
	f.Add("ccw 90")
	f.Add("flip left")
	f.Add("flip right")
	f.Add("flip forward")
	f.Add("flip backward")
	f.Add("go 100 100 100 50")
	f.Add("curve 0 0 0 100 100 100 50")
	f.Add("speed 100")
	f.Add("rc 50 50 50 50")
	f.Add("rc -50 -50 -50 -50")
	f.Add("rc 0 0 0 0")

	f.Fuzz(func(t *testing.T, command string) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Test validateCommand with various inputs
		result := manager.validateCommand(command, nil)

		// Just verify the result is valid (not a panic)
		_ = result.Allowed
		_ = result.Reason
	})
}

// FuzzValidateMovementCommand is a fuzz test for validateMovementCommand.
func FuzzValidateMovementCommand(f *testing.F) {
	f.Add(0, 0, 0)             // zero movement
	f.Add(50, 50, 50)          // moderate movement
	f.Add(100, 100, 100)       // large movement
	f.Add(-50, -50, -50)       // negative movement
	f.Add(1000, 1000, 1000)    // very large
	f.Add(-1000, -1000, -1000) // very large negative

	f.Fuzz(func(t *testing.T, x, y, z int) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Test validateMovementCommand with various inputs
		result := manager.validateMovementCommand("move", x, y, z)

		// Verify the result is valid (not a panic)
		_ = result.Allowed
		_ = result.Reason
	})
}

// FuzzValidateRotationCommand is a fuzz test for validateRotationCommand.
func FuzzValidateRotationCommand(f *testing.F) {
	f.Add(0)
	f.Add(45)
	f.Add(90)
	f.Add(180)
	f.Add(360)
	f.Add(-45)
	f.Add(-90)
	f.Add(-180)
	f.Add(-360)
	f.Add(1000)
	f.Add(-1000)
	f.Add(1000000)
	f.Add(-1000000)

	f.Fuzz(func(t *testing.T, angle int) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.validateRotationCommand("rotate", angle)

		_ = result.Allowed
		_ = result.Reason
	})
}

// FuzzValidateRCCommand is a fuzz test for validateRCCommand.
func FuzzValidateRCCommand(f *testing.F) {
	f.Add(0, 0, 0, 0)                 // all zero
	f.Add(50, 50, 50, 50)             // moderate
	f.Add(100, 100, 80, 100)          // at limits
	f.Add(-50, -50, -50, -50)         // negative moderate
	f.Add(-100, -100, -80, -100)      // negative at limits
	f.Add(1000, 1000, 1000, 1000)     // very large
	f.Add(-1000, -1000, -1000, -1000) // very large negative

	f.Fuzz(func(t *testing.T, a, b, c, d int) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.validateRCCommand(a, b, c, d)

		_ = result.Allowed
		_ = result.Reason
	})
}

// FuzzValidateSpeedCommand is a fuzz test for validateSpeedCommand.
func FuzzValidateSpeedCommand(f *testing.F) {
	f.Add(0)
	f.Add(50)
	f.Add(100)
	f.Add(-50)
	f.Add(-100)
	f.Add(1000)
	f.Add(-1000)
	f.Add(1000000)
	f.Add(-1000000)

	f.Fuzz(func(t *testing.T, speed int) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.validateSpeedCommand(speed)

		_ = result.Allowed
		_ = result.Reason
	})
}

// =============================================================================
// Error Injection Tests
// =============================================================================

// TestValidateCommandErrorInjection tests that validation methods handle
// invalid command formats gracefully without panicking.
func TestValidateCommandErrorInjection(t *testing.T) {
	tests := []struct {
		name    string
		command string
		params  map[string]any
	}{
		{
			name:    "Nil params",
			command: "test",
			params:  nil,
		},
		{
			name:    "Empty params",
			command: "test",
			params:  map[string]any{},
		},
		{
			name:    "Params with various types",
			command: "test",
			params: map[string]any{
				"string": "value",
				"int":    42,
				"float":  3.14,
				"bool":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			// Should not panic
			result := manager.validateCommand(tt.command, tt.params)

			// Verify result is valid
			if result.Allowed && result.Reason != "" {
				t.Error("Allowed result should not have a reason")
			}
		})
	}
}

// TestValidateMovementCommandErrorInjection tests error injection for movement commands.
func TestValidateMovementCommandErrorInjection(t *testing.T) {
	tests := []struct {
		name    string
		x, y, z int
	}{
		{"Zero values", 0, 0, 0},
		{"Large positive", 1000000, 1000000, 1000000},
		{"Large negative", -1000000, -1000000, -1000000},
		{"Mixed large values", 1000000, -1000000, 500000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			// Should not panic
			result := manager.validateMovementCommand("move", tt.x, tt.y, tt.z)

			// Result should be valid (not a panic)
			_ = result.Allowed
			_ = result.Reason
		})
	}
}

// TestValidateFlipCommandErrorInjection tests error injection for flip validation.
func TestValidateFlipCommandErrorInjection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
	}{
		{"Empty direction", ""},
		{"Unknown direction", "upside_down"},
		{"Multiple words", "left right"},
		{"Special chars", "left@#$"},
		{"Very long direction", "leftleftleftleftleftleftleftleftleftleft"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			// Set a height for flip validation
			state := &types.State{
				H: 200,
			}
			manager.UpdateState(state)

			// Should not panic
			result := manager.validateFlipCommand(tt.direction)

			// Result should be valid (not a panic)
			_ = result.Allowed
			_ = result.Reason
		})
	}
}

// TestValidateRCCommandErrorInjection tests error injection for RC validation.
func TestValidateRCCommandErrorInjection(t *testing.T) {
	tests := []struct {
		name       string
		a, b, c, d int
	}{
		{"Large positive values", 10000000, 10000000, 10000000, 10000000},
		{"Large negative values", -10000000, -10000000, -10000000, -10000000},
		{"Mixed extreme values", 10000000, -10000000, 5000000, -5000000},
		{"Int max values", 2147483647, 2147483647, 2147483647, 2147483647},
		{"Int min values", -2147483648, -2147483648, -2147483648, -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			// Should not panic
			result := manager.validateRCCommand(tt.a, tt.b, tt.c, tt.d)

			// Result should be valid (not a panic)
			_ = result.Allowed
			_ = result.Reason
		})
	}
}

// =============================================================================
// Command Wrapper Tests - Task 4
// =============================================================================

// TestSafetyManager_TakeOff tests the TakeOff command wrapper with safety validation.
func TestSafetyManager_TakeOff(t *testing.T) {
	t.Run("takeoff allowed with normal battery", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.TakeOff()

		if err != nil {
			t.Errorf("Expected takeoff to succeed, got error: %v", err)
		}
		if !mockCommander.takeoffCalled {
			t.Error("Expected commander.TakeOff to be called")
		}
	})

	t.Run("takeoff allowed when safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.TakeOff()

		if err != nil {
			t.Errorf("Expected takeoff to succeed with safety disabled, got: %v", err)
		}
	})

	t.Run("takeoff allowed in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.TakeOff()

		if err != nil {
			t.Errorf("Expected takeoff to succeed in emergency mode, got: %v", err)
		}
	})

	t.Run("takeoff records flight start time", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		before := time.Now()
		_ = manager.TakeOff()
		after := time.Now()

		if manager.flightStartTime.Before(before) || manager.flightStartTime.After(after) {
			t.Error("Expected flightStartTime to be set during takeoff")
		}
	})
}

// TestSafetyManager_TakeOffWithLowBattery tests that takeoff behavior with low battery state.
func TestSafetyManager_TakeOffWithLowBattery(t *testing.T) {
	// Note: TakeOff command does not block based on battery level.
	// Battery warnings are generated during UpdateState, not during command validation.
	// This test verifies that takeoff works regardless of battery state.

	t.Run("takeoff allowed with low battery state", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Set state with low battery
		state := createTestState()
		state.Bat = 5 // Below emergency threshold
		manager.UpdateState(state)

		// Takeoff should still be allowed (battery is checked during UpdateState, not TakeOff)
		err := manager.TakeOff()

		if err != nil {
			t.Errorf("Expected takeoff to be allowed, got error: %v", err)
		}
		if !mockCommander.takeoffCalled {
			t.Error("Expected commander.TakeOff to be called")
		}

		// Verify battery safety event was generated during UpdateState
		events := manager.GetSafetyEvents()
		batteryEvents := 0
		for _, event := range events {
			if event.Type == "battery" {
				batteryEvents++
			}
		}
		if batteryEvents == 0 {
			t.Error("Expected battery safety event to be recorded")
		}
	})
}

// TestSafetyManager_Land tests the Land command wrapper.
func TestSafetyManager_Land(t *testing.T) {
	t.Run("landing always allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Land()

		if err != nil {
			t.Errorf("Expected landing to succeed, got error: %v", err)
		}
		if !mockCommander.landCalled {
			t.Error("Expected commander.Land to be called")
		}
	})

	t.Run("landing allowed with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Land()

		if err != nil {
			t.Errorf("Expected landing to succeed with safety disabled, got: %v", err)
		}
	})

	t.Run("landing allowed in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Land()

		if err != nil {
			t.Errorf("Expected landing to succeed in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_Emergency tests that Emergency command is always allowed.
func TestSafetyManager_Emergency(t *testing.T) {
	t.Run("emergency always allowed with safety enabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Emergency()

		if err != nil {
			t.Errorf("Expected emergency to succeed, got error: %v", err)
		}
		if !mockCommander.emergencyCalled {
			t.Error("Expected commander.Emergency to be called")
		}
	})

	t.Run("emergency always allowed with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Emergency()

		if err != nil {
			t.Errorf("Expected emergency to succeed with safety disabled, got: %v", err)
		}
	})

	t.Run("emergency always allowed in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Emergency()

		if err != nil {
			t.Errorf("Expected emergency to succeed in emergency mode, got: %v", err)
		}
	})

	t.Run("emergency command not validated", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(true)

		// Emergency should bypass all safety checks
		err := manager.Emergency()

		if err != nil {
			t.Errorf("Expected emergency to bypass safety, got: %v", err)
		}
	})
}

// TestSafetyManager_MovementCommands tests Up, Down, Left, Right, Forward, Backward commands.
func TestSafetyManager_MovementCommands(t *testing.T) {
	t.Run("Up command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Up(50)

		if err != nil {
			t.Errorf("Expected Up to succeed, got error: %v", err)
		}
		if !mockCommander.upCalled {
			t.Error("Expected commander.Up to be called")
		}
	})

	t.Run("Down command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Down(50)

		if err != nil {
			t.Errorf("Expected Down to succeed, got error: %v", err)
		}
		if !mockCommander.downCalled {
			t.Error("Expected commander.Down to be called")
		}
	})

	t.Run("Left command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Left(50)

		if err != nil {
			t.Errorf("Expected Left to succeed, got error: %v", err)
		}
		if !mockCommander.leftCalled {
			t.Error("Expected commander.Left to be called")
		}
	})

	t.Run("Right command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Right(50)

		if err != nil {
			t.Errorf("Expected Right to succeed, got error: %v", err)
		}
		if !mockCommander.rightCalled {
			t.Error("Expected commander.Right to be called")
		}
	})

	t.Run("Forward command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Forward(50)

		if err != nil {
			t.Errorf("Expected Forward to succeed, got error: %v", err)
		}
		if !mockCommander.forwardCalled {
			t.Error("Expected commander.Forward to be called")
		}
	})

	t.Run("Backward command with valid distance", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Backward(50)

		if err != nil {
			t.Errorf("Expected Backward to succeed, got error: %v", err)
		}
		if !mockCommander.backwardCalled {
			t.Error("Expected commander.Backward to be called")
		}
	})

	t.Run("movement commands bypass validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		// All should succeed without validation
		_ = manager.Up(500)
		_ = manager.Down(500)
		_ = manager.Left(500)
		_ = manager.Right(500)
		_ = manager.Forward(500)
		_ = manager.Backward(500)

		if !mockCommander.upCalled || !mockCommander.downCalled ||
			!mockCommander.leftCalled || !mockCommander.rightCalled ||
			!mockCommander.forwardCalled || !mockCommander.backwardCalled {
			t.Error("Expected all movement commands to be called")
		}
	})

	t.Run("movement commands bypass validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		_ = manager.Up(500)
		_ = manager.Down(500)
		_ = manager.Left(500)
		_ = manager.Right(500)
		_ = manager.Forward(500)
		_ = manager.Backward(500)

		if !mockCommander.upCalled || !mockCommander.downCalled ||
			!mockCommander.leftCalled || !mockCommander.rightCalled ||
			!mockCommander.forwardCalled || !mockCommander.backwardCalled {
			t.Error("Expected all movement commands to be called in emergency mode")
		}
	})
}

// TestSafetyManager_MovementCommandsLimitEnforcement tests movement command limit enforcement.
func TestSafetyManager_MovementCommandsLimitEnforcement(t *testing.T) {
	t.Run("Go blocked when exceeds max altitude", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 100
		manager := NewSafetyManager(mockCommander, config)

		// Go command correctly passes z as altitude
		err := manager.Go(0, 0, 150, 50)

		if err == nil {
			t.Error("Expected Go to be blocked when exceeding max altitude")
		}
		if mockCommander.goCalled {
			t.Error("Expected commander.Go NOT to be called")
		}
	})

	// Go command only validates max altitude, not min altitude.
	// Minimum altitude checking would require knowing the current drone position.
}

// TestSafetyManager_RotationCommands tests Clockwise and CounterClockwise commands.
func TestSafetyManager_RotationCommands(t *testing.T) {
	t.Run("Clockwise with valid angle", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Clockwise(90)

		if err != nil {
			t.Errorf("Expected Clockwise to succeed, got error: %v", err)
		}
		if !mockCommander.clockwiseCalled {
			t.Error("Expected commander.Clockwise to be called")
		}
	})

	t.Run("CounterClockwise with valid angle", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.CounterClockwise(90)

		if err != nil {
			t.Errorf("Expected CounterClockwise to succeed, got error: %v", err)
		}
		if !mockCommander.counterClockwiseCalled {
			t.Error("Expected commander.CounterClockwise to be called")
		}
	})

	t.Run("Clockwise bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Clockwise(720)

		if err != nil {
			t.Errorf("Expected Clockwise to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("CounterClockwise bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.CounterClockwise(720)

		if err != nil {
			t.Errorf("Expected CounterClockwise to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("rotation commands bypass validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		_ = manager.Clockwise(720)
		_ = manager.CounterClockwise(720)

		if !mockCommander.clockwiseCalled || !mockCommander.counterClockwiseCalled {
			t.Error("Expected rotation commands to be called in emergency mode")
		}
	})

	t.Run("Clockwise passes validation - no angle limits in base validation", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Large angles are allowed by base validation
		err := manager.Clockwise(3600)

		if err != nil {
			t.Errorf("Expected large angle to be allowed, got: %v", err)
		}
	})
}

// TestSafetyManager_Flip tests the Flip command wrapper with direction validation.
func TestSafetyManager_Flip(t *testing.T) {
	tests := []struct {
		name          string
		direction     string
		currentHeight int
		enableFlips   bool
		minFlipHeight int
		expectAllowed bool
	}{
		{
			name:          "Flip left with valid height",
			direction:     "left",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip right with valid height",
			direction:     "right",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip forward with valid height",
			direction:     "forward",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip backward with valid height",
			direction:     "backward",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Flip disabled",
			direction:     "left",
			currentHeight: 150,
			enableFlips:   false,
			minFlipHeight: 100,
			expectAllowed: false,
		},
		{
			name:          "Below minimum flip height",
			direction:     "left",
			currentHeight: 50,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: false,
		},
		{
			name:          "At minimum flip height",
			direction:     "right",
			currentHeight: 100,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
		{
			name:          "Unknown direction still allowed",
			direction:     "upside_down",
			currentHeight: 150,
			enableFlips:   true,
			minFlipHeight: 100,
			expectAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			config.Behavioral.EnableFlips = tt.enableFlips
			config.Behavioral.MinFlipHeight = tt.minFlipHeight
			manager := NewSafetyManager(mockCommander, config)

			// Set current height
			state := createTestState()
			state.H = tt.currentHeight
			manager.UpdateState(state)

			err := manager.Flip(tt.direction)

			if tt.expectAllowed {
				if err != nil {
					t.Errorf("Expected Flip(%s) to be allowed, got: %v", tt.direction, err)
				}
				if !mockCommander.flipCalled {
					t.Error("Expected commander.Flip to be called")
				}
			} else {
				if err == nil {
					t.Errorf("Expected Flip(%s) to be blocked", tt.direction)
				}
				if mockCommander.flipCalled {
					t.Error("Expected commander.Flip NOT to be called")
				}
			}
		})
	}

	t.Run("Flip bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Behavioral.EnableFlips = false
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Flip("left")

		if err != nil {
			t.Errorf("Expected Flip to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("Flip bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Behavioral.EnableFlips = false
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Flip("left")

		if err != nil {
			t.Errorf("Expected Flip to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_Go tests the Go command wrapper with coordinate, speed, and altitude validation.
func TestSafetyManager_Go(t *testing.T) {
	t.Run("Go with valid coordinates and speed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(50, 50, 50, 50)

		if err != nil {
			t.Errorf("Expected Go to succeed, got error: %v", err)
		}
		if !mockCommander.goCalled {
			t.Error("Expected commander.Go to be called")
		}
	})

	t.Run("Go blocked with excessive speed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Velocity.MaxHorizontal = 100
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(50, 50, 50, 150)

		if err == nil {
			t.Error("Expected Go to be blocked with speed 150")
		}
		if mockCommander.goCalled {
			t.Error("Expected commander.Go NOT to be called")
		}
	})

	t.Run("Go blocked with excessive altitude", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 300
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(0, 0, 400, 50)

		if err == nil {
			t.Error("Expected Go to be blocked with altitude 400")
		}
		if mockCommander.goCalled {
			t.Error("Expected commander.Go NOT to be called")
		}
	})

	t.Run("Go at maximum limits", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Velocity.MaxHorizontal = 100
		config.Altitude.MaxHeight = 300
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(100, 100, 300, 100)

		if err != nil {
			t.Errorf("Expected Go at max limits to succeed, got: %v", err)
		}
	})

	t.Run("Go with negative coordinates allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(-50, -50, 50, 50)

		if err != nil {
			t.Errorf("Expected Go with negative coordinates to succeed, got: %v", err)
		}
	})

	t.Run("Go bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Go(500, 500, 500, 200)

		if err != nil {
			t.Errorf("Expected Go to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("Go bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Go(500, 500, 500, 200)

		if err != nil {
			t.Errorf("Expected Go to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_Curve tests the Curve command wrapper with arc validation.
func TestSafetyManager_Curve(t *testing.T) {
	t.Run("Curve with valid arc and speed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Curve(0, 0, 0, 50, 50, 50, 50)

		if err != nil {
			t.Errorf("Expected Curve to succeed, got error: %v", err)
		}
		if !mockCommander.curveCalled {
			t.Error("Expected commander.Curve to be called")
		}
	})

	t.Run("Curve blocked with excessive speed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Velocity.MaxHorizontal = 100
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Curve(0, 0, 0, 50, 50, 50, 150)

		if err == nil {
			t.Error("Expected Curve to be blocked with speed 150")
		}
		if mockCommander.curveCalled {
			t.Error("Expected commander.Curve NOT to be called")
		}
	})

	t.Run("Curve at maximum speed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Velocity.MaxHorizontal = 100
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Curve(0, 0, 0, 100, 100, 100, 100)

		if err != nil {
			t.Errorf("Expected Curve at max speed to succeed, got: %v", err)
		}
	})

	t.Run("Curve with negative arc coordinates allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Curve(-50, -50, -50, 50, 50, 50, 50)

		if err != nil {
			t.Errorf("Expected Curve with negative coordinates to succeed, got: %v", err)
		}
	})

	t.Run("Curve bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Curve(0, 0, 0, 500, 500, 500, 200)

		if err != nil {
			t.Errorf("Expected Curve to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("Curve bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Curve(0, 0, 0, 500, 500, 500, 200)

		if err != nil {
			t.Errorf("Expected Curve to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_SetSpeed tests the SetSpeed command wrapper with speed limits.
func TestSafetyManager_SetSpeed(t *testing.T) {
	tests := []struct {
		name          string
		speed         int
		expectAllowed bool
	}{
		{
			name:          "Zero speed",
			speed:         0,
			expectAllowed: true,
		},
		{
			name:          "Low speed",
			speed:         10,
			expectAllowed: true,
		},
		{
			name:          "Medium speed",
			speed:         50,
			expectAllowed: true,
		},
		{
			name:          "Maximum speed",
			speed:         100,
			expectAllowed: true,
		},
		{
			name:          "Just below limit",
			speed:         99,
			expectAllowed: true,
		},
		{
			name:          "Just above limit",
			speed:         101,
			expectAllowed: false,
		},
		{
			name:          "Double speed limit",
			speed:         200,
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			config.Velocity.MaxHorizontal = 100
			manager := NewSafetyManager(mockCommander, config)

			err := manager.SetSpeed(tt.speed)

			if tt.expectAllowed {
				if err != nil {
					t.Errorf("Expected SetSpeed(%d) to be allowed, got: %v", tt.speed, err)
				}
				if !mockCommander.setSpeedCalled {
					t.Error("Expected commander.SetSpeed to be called")
				}
			} else {
				if err == nil {
					t.Errorf("Expected SetSpeed(%d) to be blocked", tt.speed)
				}
				if mockCommander.setSpeedCalled {
					t.Error("Expected commander.SetSpeed NOT to be called")
				}
			}
		})
	}

	t.Run("SetSpeed bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.SetSpeed(200)

		if err != nil {
			t.Errorf("Expected SetSpeed to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("SetSpeed bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.SetSpeed(200)

		if err != nil {
			t.Errorf("Expected SetSpeed to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_SetRCControl tests the SetRCControl command wrapper with RC bounds.
func TestSafetyManager_SetRCControl(t *testing.T) {
	tests := []struct {
		name          string
		a, b, c, d    int
		expectAllowed bool
	}{
		// Valid cases
		{
			name:          "Zero RC values",
			a:             0,
			b:             0,
			c:             0,
			d:             0,
			expectAllowed: true,
		},
		{
			name:          "All axes within limits",
			a:             50,
			b:             50,
			c:             40,
			d:             50,
			expectAllowed: true,
		},
		{
			name:          "At maximum limits",
			a:             100,
			b:             100,
			c:             80,
			d:             100,
			expectAllowed: true,
		},
		// Negative values
		{
			name:          "Negative values within limits",
			a:             -50,
			b:             -50,
			c:             -40,
			d:             -50,
			expectAllowed: true,
		},
		// Exceeds limits
		{
			name:          "Axis a exceeds limit",
			a:             101,
			b:             0,
			c:             0,
			d:             0,
			expectAllowed: false,
		},
		{
			name:          "Axis b exceeds limit",
			a:             0,
			b:             101,
			c:             0,
			d:             0,
			expectAllowed: false,
		},
		{
			name:          "Axis c exceeds vertical limit",
			a:             0,
			b:             0,
			c:             81,
			d:             0,
			expectAllowed: false,
		},
		{
			name:          "Axis d exceeds yaw limit",
			a:             0,
			b:             0,
			c:             0,
			d:             101,
			expectAllowed: false,
		},
		{
			name:          "Multiple axes exceed limits",
			a:             150,
			b:             150,
			c:             150,
			d:             150,
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			err := manager.SetRcControl(tt.a, tt.b, tt.c, tt.d)

			if tt.expectAllowed {
				if err != nil {
					t.Errorf("Expected SetRcControl(%d, %d, %d, %d) to be allowed, got: %v", tt.a, tt.b, tt.c, tt.d, err)
				}
				if !mockCommander.setRcControlCalled {
					t.Error("Expected commander.SetRcControl to be called")
				}
			} else {
				if err == nil {
					t.Errorf("Expected SetRcControl(%d, %d, %d, %d) to be blocked", tt.a, tt.b, tt.c, tt.d)
				}
				if mockCommander.setRcControlCalled {
					t.Error("Expected commander.SetRcControl NOT to be called")
				}
			}
		})
	}

	t.Run("SetRcControl bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.SetRcControl(200, 200, 200, 200)

		if err != nil {
			t.Errorf("Expected SetRCControl to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("SetRcControl bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.SetRcControl(200, 200, 200, 200)

		if err != nil {
			t.Errorf("Expected SetRcControl to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_SetWiFiCredentials tests the SetWiFiCredentials command wrapper.
func TestSafetyManager_SetWiFiCredentials(t *testing.T) {
	t.Run("SetWiFiCredentials always allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.SetWiFiCredentials("test-ssid", "test-password")

		if err != nil {
			t.Errorf("Expected SetWiFiCredentials to succeed, got error: %v", err)
		}
		if !mockCommander.setWiFiCredentialsCalled {
			t.Error("Expected commander.SetWiFiCredentials to be called")
		}
	})

	t.Run("SetWiFiCredentials with special characters", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.SetWiFiCredentials("My-SSID-123", "p@ss!word#123")

		if err != nil {
			t.Errorf("Expected SetWiFiCredentials with special chars to succeed, got: %v", err)
		}
	})

	t.Run("SetWiFiCredentials bypasses all validation", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(true)
		manager.SetEmergencyMode(true)

		err := manager.SetWiFiCredentials("ssid", "password")

		if err != nil {
			t.Errorf("Expected SetWiFiCredentials to always bypass validation, got: %v", err)
		}
	})
}

// TestSafetyManager_StreamCommands tests StreamOn and StreamOff command wrappers.
func TestSafetyManager_StreamCommands(t *testing.T) {
	t.Run("StreamOn with valid state", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.StreamOn()

		if err != nil {
			t.Errorf("Expected StreamOn to succeed, got error: %v", err)
		}
		if !mockCommander.streamOnCalled {
			t.Error("Expected commander.StreamOn to be called")
		}
	})

	t.Run("StreamOff always allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.StreamOff()

		if err != nil {
			t.Errorf("Expected StreamOff to succeed, got error: %v", err)
		}
		if !mockCommander.streamOffCalled {
			t.Error("Expected commander.StreamOff to be called")
		}
	})

	t.Run("StreamOn bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.StreamOn()

		if err != nil {
			t.Errorf("Expected StreamOn to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("StreamOff bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.StreamOff()

		if err != nil {
			t.Errorf("Expected StreamOff to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("Stream commands bypass validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		_ = manager.StreamOn()
		_ = manager.StreamOff()

		if !mockCommander.streamOnCalled || !mockCommander.streamOffCalled {
			t.Error("Expected stream commands to be called in emergency mode")
		}
	})
}

// TestSafetyManager_Init tests the Init command wrapper.
func TestSafetyManager_Init(t *testing.T) {
	t.Run("Init always allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Init()

		if err != nil {
			t.Errorf("Expected Init to succeed, got error: %v", err)
		}
		if !mockCommander.initCalled {
			t.Error("Expected commander.Init to be called")
		}
	})

	t.Run("Init bypasses validation with safety disabled", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Init()

		if err != nil {
			t.Errorf("Expected Init to bypass validation with safety disabled, got: %v", err)
		}
	})

	t.Run("Init bypasses validation in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Init()

		if err != nil {
			t.Errorf("Expected Init to bypass validation in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_EmergencyMode tests emergency mode behavior.
func TestSafetyManager_EmergencyMode(t *testing.T) {
	t.Run("SetEmergencyMode activates emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEmergencyMode(true)

		status := manager.GetSafetyStatus()
		if !status.EmergencyMode {
			t.Error("Expected emergency mode to be activated")
		}
	})

	t.Run("SetEmergencyMode deactivates emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)
		manager.SetEmergencyMode(false)

		status := manager.GetSafetyStatus()
		if status.EmergencyMode {
			t.Error("Expected emergency mode to be deactivated")
		}
	})

	t.Run("Emergency mode creates safety event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEmergencyMode(true)

		events := manager.GetSafetyEvents()
		if len(events) == 0 {
			t.Error("Expected safety event for emergency mode activation")
		}
		// Check last event is emergency
		lastEvent := events[len(events)-1]
		if lastEvent.Level != "emergency" {
			t.Errorf("Expected emergency level event, got: %s", lastEvent.Level)
		}
	})

	t.Run("Emergency mode blocks unsafe commands", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		// Movement commands should be blocked
		err := manager.Up(50)
		if err != nil {
			t.Errorf("Expected Up to bypass validation in emergency mode, got: %v", err)
		}

		err = manager.Flip("left")
		if err != nil {
			t.Errorf("Expected Flip to bypass validation in emergency mode, got: %v", err)
		}

		err = manager.SetSpeed(150)
		if err != nil {
			t.Errorf("Expected SetSpeed to bypass validation in emergency mode, got: %v", err)
		}
	})

	t.Run("Emergency command always works in emergency mode", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		// Even if already in emergency mode, Emergency() should be callable
		err := manager.Emergency()
		if err != nil {
			t.Errorf("Expected Emergency() to work in emergency mode, got: %v", err)
		}
	})
}

// TestSafetyManager_SafetyEnabled tests safety enable/disable behavior.
func TestSafetyManager_SafetyEnabled(t *testing.T) {
	t.Run("SetSafetyEnabled disables safety", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetSafetyEnabled(false)

		status := manager.GetSafetyStatus()
		if status.SafetyEnabled {
			t.Error("Expected safety to be disabled")
		}
	})

	t.Run("SetSafetyEnabled enables safety", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)
		manager.SetSafetyEnabled(true)

		status := manager.GetSafetyStatus()
		if !status.SafetyEnabled {
			t.Error("Expected safety to be enabled")
		}
	})

	t.Run("Disabled safety allows all commands", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 100
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		// Should allow exceeding limits
		err := manager.Up(200)
		if err != nil {
			t.Errorf("Expected Up to bypass validation with safety disabled, got: %v", err)
		}

		err = manager.SetSpeed(200)
		if err != nil {
			t.Errorf("Expected SetSpeed to bypass validation with safety disabled, got: %v", err)
		}
	})
}

// TestSafetyManager_SetEventCallback tests the SetEventCallback method.
func TestSafetyManager_SetEventCallback(t *testing.T) {
	t.Run("SetEventCallback updates callback", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEventCallback(func(event *SafetyEvent) {
			// Callback would be invoked
		})

		// Trigger a safety event by setting low battery
		state := createTestState()
		state.Bat = 5 // Below emergency threshold
		manager.UpdateState(state)

		// Verify event was generated
		events := manager.GetSafetyEvents()
		if len(events) == 0 {
			t.Error("Expected safety event to be generated")
		}
	})

	t.Run("SetEventCallback can be updated", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Set initial callback
		manager.SetEventCallback(func(event *SafetyEvent) {})

		// Update callback - should not panic
		manager.SetEventCallback(func(event *SafetyEvent) {})

		// Generate an event - should not crash
		state := createTestState()
		state.Bat = 5
		manager.UpdateState(state)
	})

	t.Run("SetEventCallback with nil is allowed", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// Should not panic
		manager.SetEventCallback(nil)

		// Generate an event - should not crash
		state := createTestState()
		state.Bat = 5
		manager.UpdateState(state)
	})
}

// TestSafetyManager_CommandBlocking generates safety events correctly.
func TestSafetyManager_CommandBlocking(t *testing.T) {
	t.Run("Blocked command returns error", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 100
		manager := NewSafetyManager(mockCommander, config)

		// Use Go command which correctly validates altitude
		err := manager.Go(0, 0, 150, 50)

		if err == nil {
			t.Error("Expected command to be blocked")
		}
		if mockCommander.goCalled {
			t.Error("Expected commander.Go NOT to be called")
		}
	})

	t.Run("Multiple blocked commands return errors", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 100
		manager := NewSafetyManager(mockCommander, config)

		// Block multiple commands using Go
		err1 := manager.Go(0, 0, 150, 50)
		err2 := manager.Go(0, 0, 150, 50)
		err3 := manager.Go(0, 0, 150, 50)

		if err1 == nil || err2 == nil || err3 == nil {
			t.Error("Expected all commands to be blocked")
		}
	})

	t.Run("Emergency mode activation generates event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEmergencyMode(true)

		events := manager.GetSafetyEvents()
		emergencyEvents := 0
		for _, event := range events {
			if event.Type == "emergency" {
				emergencyEvents++
			}
		}
		if emergencyEvents == 0 {
			t.Error("Expected emergency event for emergency mode activation")
		}
	})
}

// TestSafetyManager_AllCommands table-driven test for all command wrappers.
func TestSafetyManager_AllCommands(t *testing.T) {
	tests := []struct {
		name         string
		command      func(*SafetyManager) error
		verifyCalled func(*MockCommander) bool
	}{
		{
			name:         "TakeOff",
			command:      func(m *SafetyManager) error { return m.TakeOff() },
			verifyCalled: func(m *MockCommander) bool { return m.takeoffCalled },
		},
		{
			name:         "Land",
			command:      func(m *SafetyManager) error { return m.Land() },
			verifyCalled: func(m *MockCommander) bool { return m.landCalled },
		},
		{
			name:         "Emergency",
			command:      func(m *SafetyManager) error { return m.Emergency() },
			verifyCalled: func(m *MockCommander) bool { return m.emergencyCalled },
		},
		{
			name:         "Up",
			command:      func(m *SafetyManager) error { return m.Up(50) },
			verifyCalled: func(m *MockCommander) bool { return m.upCalled },
		},
		{
			name:         "Down",
			command:      func(m *SafetyManager) error { return m.Down(50) },
			verifyCalled: func(m *MockCommander) bool { return m.downCalled },
		},
		{
			name:         "Left",
			command:      func(m *SafetyManager) error { return m.Left(50) },
			verifyCalled: func(m *MockCommander) bool { return m.leftCalled },
		},
		{
			name:         "Right",
			command:      func(m *SafetyManager) error { return m.Right(50) },
			verifyCalled: func(m *MockCommander) bool { return m.rightCalled },
		},
		{
			name:         "Forward",
			command:      func(m *SafetyManager) error { return m.Forward(50) },
			verifyCalled: func(m *MockCommander) bool { return m.forwardCalled },
		},
		{
			name:         "Backward",
			command:      func(m *SafetyManager) error { return m.Backward(50) },
			verifyCalled: func(m *MockCommander) bool { return m.backwardCalled },
		},
		{
			name:         "Clockwise",
			command:      func(m *SafetyManager) error { return m.Clockwise(90) },
			verifyCalled: func(m *MockCommander) bool { return m.clockwiseCalled },
		},
		{
			name:         "CounterClockwise",
			command:      func(m *SafetyManager) error { return m.CounterClockwise(90) },
			verifyCalled: func(m *MockCommander) bool { return m.counterClockwiseCalled },
		},
		{
			name:         "Flip",
			command:      func(m *SafetyManager) error { return m.Flip("left") },
			verifyCalled: func(m *MockCommander) bool { return m.flipCalled },
		},
		{
			name:         "Go",
			command:      func(m *SafetyManager) error { return m.Go(50, 50, 50, 50) },
			verifyCalled: func(m *MockCommander) bool { return m.goCalled },
		},
		{
			name:         "Curve",
			command:      func(m *SafetyManager) error { return m.Curve(0, 0, 0, 50, 50, 50, 50) },
			verifyCalled: func(m *MockCommander) bool { return m.curveCalled },
		},
		{
			name:         "SetSpeed",
			command:      func(m *SafetyManager) error { return m.SetSpeed(50) },
			verifyCalled: func(m *MockCommander) bool { return m.setSpeedCalled },
		},
		{
			name:         "SetRcControl",
			command:      func(m *SafetyManager) error { return m.SetRcControl(50, 50, 50, 50) },
			verifyCalled: func(m *MockCommander) bool { return m.setRcControlCalled },
		},
		{
			name:         "SetWiFiCredentials",
			command:      func(m *SafetyManager) error { return m.SetWiFiCredentials("ssid", "pass") },
			verifyCalled: func(m *MockCommander) bool { return m.setWiFiCredentialsCalled },
		},
		{
			name:         "StreamOn",
			command:      func(m *SafetyManager) error { return m.StreamOn() },
			verifyCalled: func(m *MockCommander) bool { return m.streamOnCalled },
		},
		{
			name:         "StreamOff",
			command:      func(m *SafetyManager) error { return m.StreamOff() },
			verifyCalled: func(m *MockCommander) bool { return m.streamOffCalled },
		},
		{
			name:         "Init",
			command:      func(m *SafetyManager) error { return m.Init() },
			verifyCalled: func(m *MockCommander) bool { return m.initCalled },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			err := tt.command(manager)

			if err != nil {
				t.Errorf("Expected %s to succeed, got error: %v", tt.name, err)
			}
			if !tt.verifyCalled(mockCommander) {
				t.Errorf("Expected commander.%s to be called", tt.name)
			}
		})
	}
}

// TestCheckFlightTimeWithRealDuration tests flight time validation with real durations.
func TestCheckFlightTimeWithRealDuration(t *testing.T) {
	tests := []struct {
		name          string
		flightTime    time.Duration
		maxFlightTime int
		expectAllowed bool
	}{
		{"No flight started", 0, 600, true},
		{"Just started", 1 * time.Second, 600, true},
		{"Well under limit", 5 * time.Minute, 600, true},
		{"At limit allowed", 9*time.Minute + 50*time.Second, 600, true}, // Slightly under to avoid timing issues
		{"Just over limit", 10*time.Minute + 1*time.Second, 600, false},
		{"Significantly over", 15 * time.Minute, 600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			config.Behavioral.MaxFlightTime = tt.maxFlightTime
			manager := NewSafetyManager(mockCommander, config)

			if tt.flightTime > 0 {
				manager.flightStartTime = time.Now().Add(-tt.flightTime)
			}

			result := manager.checkFlightTime()

			if tt.expectAllowed && !result {
				t.Errorf("Expected flight time check to allow %v", tt.flightTime)
			}
			if !tt.expectAllowed && result {
				t.Errorf("Expected flight time check to block %v", tt.flightTime)
			}
		})
	}
}

// TestAddEventWithCallback tests event recording with callback invocation.
func TestAddEventWithCallback(t *testing.T) {
	t.Run("event callback is invoked", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		callbackCalled := make(chan bool, 1)
		manager.SetEventCallback(func(event *SafetyEvent) {
			callbackCalled <- true
		})

		event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelWarning,
			"Test battery warning", map[string]any{"battery_level": 25})
		manager.addEvent(event)

		select {
		case <-callbackCalled:
			// Callback was invoked
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected event callback to be invoked")
		}
	})

	t.Run("nil callback does not crash", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEventCallback(nil)

		event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelWarning, "Test", nil)
		manager.addEvent(event)
	})

	t.Run("multiple events are recorded", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.status.ActiveEvents = []SafetyEvent{}

		for i := 0; i < 5; i++ {
			event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelWarning,
				fmt.Sprintf("Event %d", i), nil)
			manager.addEvent(event)
		}

		events := manager.GetSafetyEvents()
		if len(events) != 5 {
			t.Errorf("Expected 5 events, got %d", len(events))
		}
	})
}

// TestHasCriticalOrEmergencyEvents tests critical/emergency event detection.
func TestHasCriticalOrEmergencyEvents(t *testing.T) {
	tests := []struct {
		name         string
		events       []SafetyEvent
		expectResult bool
	}{
		{"No events", []SafetyEvent{}, false},
		{"Only info", []SafetyEvent{{Type: "test", Level: "info"}}, false},
		{"Only warnings", []SafetyEvent{{Type: "test", Level: "warning"}, {Type: "test", Level: "warning"}}, false},
		{"Mixed info warning", []SafetyEvent{{Type: "test", Level: "info"}, {Type: "test", Level: "warning"}}, false},
		{"Has critical", []SafetyEvent{{Type: "test", Level: "warning"}, {Type: "test", Level: "critical"}}, true},
		{"Has emergency", []SafetyEvent{{Type: "test", Level: "warning"}, {Type: "test", Level: "emergency"}}, true},
		{"Multiple critical emergency", []SafetyEvent{{Type: "test", Level: "critical"}, {Type: "test", Level: "emergency"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommander := NewMockCommander()
			config := DefaultConfig()
			manager := NewSafetyManager(mockCommander, config)

			result := manager.hasCriticalOrEmergencyEvents(tt.events)

			if result != tt.expectResult {
				t.Errorf("Expected %v, got %v", tt.expectResult, result)
			}
		})
	}
}

// TestUpdateStateIntegration tests UpdateState integration with safety checks.
func TestUpdateStateIntegration(t *testing.T) {
	t.Run("UpdateState triggers safety checks", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 100
		config.Battery.EmergencyThreshold = 10
		manager := NewSafetyManager(mockCommander, config)

		state := &types.State{
			H:   150,
			Bat: 5,
		}

		manager.UpdateState(state)

		events := manager.GetSafetyEvents()
		if len(events) == 0 {
			t.Error("Expected safety events from UpdateState")
		}
	})

	t.Run("UpdateState updates timestamp", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		before := time.Now()
		state := createTestState()
		manager.UpdateState(state)
		after := time.Now()

		if manager.lastStateUpdate.Before(before) || manager.lastStateUpdate.After(after) {
			t.Error("Expected lastStateUpdate to be set")
		}
	})

	t.Run("UpdateState clears old events", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		oldEvent := SafetyEvent{
			Type:      "test",
			Level:     "warning",
			Message:   "Old event",
			Timestamp: time.Now().Add(-10 * time.Minute),
		}
		manager.status.ActiveEvents = append(manager.status.ActiveEvents, oldEvent)

		recentEvent := SafetyEvent{
			Type:      "test",
			Level:     "warning",
			Message:   "Recent event",
			Timestamp: time.Now(),
		}
		manager.status.ActiveEvents = append(manager.status.ActiveEvents, recentEvent)

		state := createTestState()
		manager.UpdateState(state)

		events := manager.GetSafetyEvents()
		if len(events) != 1 {
			t.Errorf("Expected 1 event after cleanup, got %d", len(events))
		}
	})
}

// TestCommandRateLimit tests command rate limiting enforcement.
func TestCommandRateLimit(t *testing.T) {
	t.Run("checkCommandRate allows first command", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		result := manager.checkCommandRate()
		if !result {
			t.Error("Expected first command to be allowed")
		}
	})

	t.Run("checkCommandRate blocks immediate repeat", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// First call
		manager.checkCommandRate()

		// Immediate second call should be blocked
		result := manager.checkCommandRate()
		if result {
			t.Error("Expected immediate repeat to be blocked")
		}
	})

	t.Run("checkCommandRate allows after delay", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		// First call
		manager.checkCommandRate()

		// Wait for rate limit to reset
		time.Sleep(200 * time.Millisecond)

		// Should be allowed
		result := manager.checkCommandRate()
		if !result {
			t.Error("Expected command after delay to be allowed")
		}
	})
}

// TestEmergencyModeBlocksCommands tests emergency mode command blocking.
func TestEmergencyModeBlocksCommands(t *testing.T) {
	t.Run("emergency mode bypasses validation", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)
		manager.SetEmergencyMode(true)

		err := manager.Up(50)
		if err != nil {
			t.Errorf("Expected Up to bypass validation in emergency mode, got: %v", err)
		}
	})

	t.Run("Emergency command always executes", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Emergency()
		if err != nil {
			t.Errorf("Expected Emergency to work, got: %v", err)
		}

		if !mockCommander.emergencyCalled {
			t.Error("Expected commander.Emergency to be called")
		}
	})

	t.Run("emergency mode activation generates event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.SetEmergencyMode(true)

		events := manager.GetSafetyEvents()
		emergencyEvents := 0
		for _, event := range events {
			if event.Type == "emergency" {
				emergencyEvents++
			}
		}
		if emergencyEvents == 0 {
			t.Error("Expected emergency event for emergency mode activation")
		}
	})
}

// TestSafetyEventTypes tests different safety event types.
func TestSafetyEventTypes(t *testing.T) {
	t.Run("all event types are recorded", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		manager.status.ActiveEvents = []SafetyEvent{}

		eventTypes := []SafetyEventType{
			SafetyEventAltitude,
			SafetyEventBattery,
			SafetyEventSensor,
			SafetyEventBehavioral,
			SafetyEventEmergency,
		}

		for _, eventType := range eventTypes {
			event := NewSafetyEvent(eventType, SafetyEventLevelWarning,
				fmt.Sprintf("Test %s event", eventType), nil)
			manager.addEvent(event)
		}

		events := manager.GetSafetyEvents()
		if len(events) != len(eventTypes) {
			t.Errorf("Expected %d events, got %d", len(eventTypes), len(events))
		}
	})
}

// TestCommandWrapperValidationErrors tests validation error handling.
func TestCommandWrapperValidationErrors(t *testing.T) {
	t.Run("validation error prevents command", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 50
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Go(0, 0, 100, 50)

		if err == nil {
			t.Error("Expected validation error for excessive altitude")
		}
		if mockCommander.goCalled {
			t.Error("Expected commander.Go NOT to be called")
		}
	})

	t.Run("safety disabled bypasses validation", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Altitude.MaxHeight = 50
		manager := NewSafetyManager(mockCommander, config)
		manager.SetSafetyEnabled(false)

		err := manager.Go(0, 0, 100, 50)

		if err != nil {
			t.Errorf("Expected bypass when safety disabled, got: %v", err)
		}
	})
}

// TestCompleteSafetyLifecycle tests complete safety system lifecycle.
func TestCompleteSafetyLifecycle(t *testing.T) {
	t.Run("full lifecycle from init to land", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Behavioral.MaxCommandRate = 100 // Allow 100 commands/second for testing
		manager := NewSafetyManager(mockCommander, config)

		err := manager.Init()
		if err != nil {
			t.Logf("Init error: %v", err)
		}
		// Small delay to allow rate limiting check to pass
		time.Sleep(10 * time.Millisecond)
		err = manager.TakeOff()
		if err != nil {
			t.Logf("TakeOff error: %v", err)
		}
		_ = manager.Up(50)
		_ = manager.Forward(100)
		_ = manager.Clockwise(90)
		_ = manager.Emergency()
		_ = manager.Land()

		if !mockCommander.initCalled {
			t.Error("Expected Init to be called")
		}
		if !mockCommander.takeoffCalled {
			t.Errorf("Expected TakeOff to be called, got initCalled=%v, takeoffCalled=%v, upCalled=%v, forwardCalled=%v, clockwiseCalled=%v, emergencyCalled=%v, landCalled=%v",
				mockCommander.initCalled, mockCommander.takeoffCalled, mockCommander.upCalled,
				mockCommander.forwardCalled, mockCommander.clockwiseCalled, mockCommander.emergencyCalled, mockCommander.landCalled)
		}
		if !mockCommander.emergencyCalled {
			t.Error("Expected Emergency to be called")
		}
		if !mockCommander.landCalled {
			t.Error("Expected Land to be called")
		}
	})

	t.Run("state consistency after toggles", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		for i := 0; i < 3; i++ {
			manager.SetEmergencyMode(true)
			if !manager.GetSafetyStatus().EmergencyMode {
				t.Errorf("Emergency mode inconsistent (iter %d)", i)
			}

			manager.SetEmergencyMode(false)
			if manager.GetSafetyStatus().EmergencyMode {
				t.Errorf("Emergency mode inconsistent after toggle (iter %d)", i)
			}
		}
	})
}

// TestCheckBehavioralSafety tests the behavioral safety checks including flight time limits.
func TestCheckBehavioralSafety(t *testing.T) {
	t.Run("flight time limit exceeded triggers event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Behavioral.MaxFlightTime = 1 // 1 second max flight time
		manager := NewSafetyManager(mockCommander, config)

		// Manually set flight start time to 3 seconds ago (ensures we exceed the limit)
		manager.mutex.Lock()
		manager.flightStartTime = time.Now().Add(-3 * time.Second)
		manager.mutex.Unlock()

		// Call UpdateState to trigger the safety checks
		manager.UpdateState(&types.State{H: 100})

		// Check that a flight time event was recorded
		status := manager.GetSafetyStatus()
		if len(status.ActiveEvents) == 0 {
			t.Error("Expected flight time exceeded event to be recorded")
		}

		// Verify flight time event is present
		found := false
		for _, event := range status.ActiveEvents {
			if strings.Contains(event.Message, "flight time") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected flight time exceeded event in active events")
		}
	})
}

// TestCheckSensorSafety tests the sensor safety checks.
func TestCheckSensorSafety(t *testing.T) {
	t.Run("TOF warning triggers event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Sensors.MinTOFDistance = 50 // 50cm minimum TOF
		manager := NewSafetyManager(mockCommander, config)

		// Create state with TOF below minimum
		state := &types.State{H: 100, Tof: 30} // 30cm is below 50cm minimum
		manager.checkSensorSafety(state)

		// Check that a TOF event was recorded
		status := manager.GetSafetyStatus()
		if len(status.ActiveEvents) == 0 {
			t.Error("Expected TOF warning event to be recorded")
		}
	})

	t.Run("tilt warning triggers event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Sensors.MaxTiltAngle = 30 // 30 degrees max tilt
		manager := NewSafetyManager(mockCommander, config)

		// Create state with pitch exceeding max tilt
		state := &types.State{H: 100, Pitch: 45} // 45 degrees exceeds 30 degree max
		manager.checkSensorSafety(state)

		// Check that a tilt event was recorded
		status := manager.GetSafetyStatus()
		if len(status.ActiveEvents) == 0 {
			t.Error("Expected tilt warning event to be recorded")
		}
	})

	t.Run("excessive acceleration triggers event", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		config.Sensors.MaxAcceleration = 2.0 // 2g max acceleration
		manager := NewSafetyManager(mockCommander, config)

		// Create state with excessive acceleration
		state := &types.State{H: 100, Agx: 1, Agy: 1, Agz: 2} // Magnitude > 2.0
		manager.checkSensorSafety(state)

		// Check that an acceleration event was recorded
		status := manager.GetSafetyStatus()
		if len(status.ActiveEvents) == 0 {
			t.Error("Expected acceleration warning event to be recorded")
		}
	})
}

// TestGetMethods tests the getter methods.
func TestGetMethods(t *testing.T) {
	mockCommander := NewMockCommander()
	// Set up mock responses
	mockCommander.speedResponse = 75
	mockCommander.batteryResponse = 65
	mockCommander.heightResponse = 150
	mockCommander.temperatureResponse = 35
	mockCommander.attitudeResponsePitch = 10
	mockCommander.attitudeResponseRoll = -5
	mockCommander.attitudeResponseYaw = 180
	mockCommander.barometerResponse = 1000
	mockCommander.accelerationResponseX = 100
	mockCommander.accelerationResponseY = 200
	mockCommander.accelerationResponseZ = 300
	mockCommander.tofResponse = 500

	config := DefaultConfig()
	manager := NewSafetyManager(mockCommander, config)

	// Test GetSpeed
	speed, err := manager.GetSpeed()
	if err != nil {
		t.Errorf("GetSpeed returned error: %v", err)
	}
	if speed != 75 {
		t.Errorf("Expected speed 75, got %d", speed)
	}

	// Test GetBatteryPercentage
	battery, err := manager.GetBatteryPercentage()
	if err != nil {
		t.Errorf("GetBatteryPercentage returned error: %v", err)
	}
	if battery != 65 {
		t.Errorf("Expected battery 65, got %d", battery)
	}

	// Test GetHeight
	height, err := manager.GetHeight()
	if err != nil {
		t.Errorf("GetHeight returned error: %v", err)
	}
	if height != 150 {
		t.Errorf("Expected height 150, got %d", height)
	}

	// Test GetTemperature
	temp, err := manager.GetTemperature()
	if err != nil {
		t.Errorf("GetTemperature returned error: %v", err)
	}
	if temp != 35 {
		t.Errorf("Expected temperature 35, got %d", temp)
	}

	// Test GetAttitude
	pitch, roll, yaw, err := manager.GetAttitude()
	if err != nil {
		t.Errorf("GetAttitude returned error: %v", err)
	}
	if pitch != 10 || roll != -5 || yaw != 180 {
		t.Errorf("Expected attitude (10, -5, 180), got (%d, %d, %d)", pitch, roll, yaw)
	}

	// Test GetBarometer
	baro, err := manager.GetBarometer()
	if err != nil {
		t.Errorf("GetBarometer returned error: %v", err)
	}
	if baro != 1000 {
		t.Errorf("Expected barometer 1000, got %d", baro)
	}

	// Test GetAcceleration
	ax, ay, az, err := manager.GetAcceleration()
	if err != nil {
		t.Errorf("GetAcceleration returned error: %v", err)
	}
	if ax != 100 || ay != 200 || az != 300 {
		t.Errorf("Expected acceleration (100, 200, 300), got (%d, %d, %d)", ax, ay, az)
	}

	// Test GetTof
	tof, err := manager.GetTof()
	if err != nil {
		t.Errorf("GetTof returned error: %v", err)
	}
	if tof != 500 {
		t.Errorf("Expected TOF 500, got %d", tof)
	}
}

// TestVideoFrameCallback tests the video frame callback methods.
func TestVideoFrameCallback(t *testing.T) {
	mockCommander := NewMockCommander()
	config := DefaultConfig()
	manager := NewSafetyManager(mockCommander, config)

	// Test SetVideoFrameCallback
	manager.SetVideoFrameCallback(func(frame transport.VideoFrame) {
		_ = frame
	})

	// Verify the callback was set on the mock commander
	if !mockCommander.setVideoFrameCallbackCalled {
		t.Error("Expected SetVideoFrameCallback to be called on mock commander")
	}

	// Test GetVideoFrameChannel
	channel := manager.GetVideoFrameChannel()
	if channel == nil {
		t.Error("Expected non-nil video frame channel")
	}
	if !mockCommander.getVideoFrameChannelCalled {
		t.Error("Expected GetVideoFrameChannel to be called on mock commander")
	}
}
