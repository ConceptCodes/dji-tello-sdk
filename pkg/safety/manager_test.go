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
	"sync"
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
	return nil
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

	t.Run("concurrent start and stop is safe", func(t *testing.T) {
		mockCommander := NewMockCommander()
		config := DefaultConfig()
		manager := NewSafetyManager(mockCommander, config)

		stateChan := make(chan *types.State)

		// Concurrently start and stop
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			manager.StartTelemetryProcessing(stateChan)
		}()

		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			manager.StopTelemetryProcessing()
		}()

		wg.Wait()
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
