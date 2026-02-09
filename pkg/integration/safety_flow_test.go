package integration

import (
	"errors"
	"testing"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// MockCommander is a mock implementation of safety.CommanderInterface for testing
type MockCommander struct {
	initCalled      bool
	takeoffCalled   bool
	landCalled      bool
	emergencyCalled bool
	upCalled        bool
	forwardCalled   bool
	flipCalled      bool

	// Error injection
	initError      error
	takeoffError   error
	landError      error
	emergencyError error
	upError        error
	forwardError   error
	flipError      error

	// State responses
	batteryResponse int
	heightResponse  int
}

// NewMockCommander creates a new mock commander with default values
func NewMockCommander() *MockCommander {
	return &MockCommander{
		batteryResponse: 85,
		heightResponse:  100,
	}
}

// Control Commands
func (m *MockCommander) Init() error {
	m.initCalled = true
	return m.initError
}

func (m *MockCommander) TakeOff() error {
	m.takeoffCalled = true
	return m.takeoffError
}

func (m *MockCommander) Land() error {
	m.landCalled = true
	return m.landError
}

func (m *MockCommander) StreamOn() error  { return nil }
func (m *MockCommander) StreamOff() error { return nil }

func (m *MockCommander) Emergency() error {
	m.emergencyCalled = true
	return m.emergencyError
}

// Movement Commands
func (m *MockCommander) Up(distance int) error {
	m.upCalled = true
	return m.upError
}

func (m *MockCommander) Down(distance int) error  { return nil }
func (m *MockCommander) Left(distance int) error  { return nil }
func (m *MockCommander) Right(distance int) error { return nil }

func (m *MockCommander) Forward(distance int) error {
	m.forwardCalled = true
	return m.forwardError
}

func (m *MockCommander) Backward(distance int) error      { return nil }
func (m *MockCommander) Clockwise(angle int) error        { return nil }
func (m *MockCommander) CounterClockwise(angle int) error { return nil }

func (m *MockCommander) Flip(direction string) error {
	m.flipCalled = true
	return m.flipError
}

func (m *MockCommander) Go(x, y, z, speed int) error                   { return nil }
func (m *MockCommander) Curve(x1, y1, z1, x2, y2, z2, speed int) error { return nil }

// Set Commands
func (m *MockCommander) SetSpeed(speed int) error                       { return nil }
func (m *MockCommander) SetRcControl(a, b, c, d int) error              { return nil }
func (m *MockCommander) SetWiFiCredentials(ssid, password string) error { return nil }

// Read Commands
func (m *MockCommander) GetSpeed() (int, error)                  { return 50, nil }
func (m *MockCommander) GetBatteryPercentage() (int, error)      { return m.batteryResponse, nil }
func (m *MockCommander) GetTime() (int, error)                   { return 120, nil }
func (m *MockCommander) GetHeight() (int, error)                 { return m.heightResponse, nil }
func (m *MockCommander) GetTemperature() (int, error)            { return 25, nil }
func (m *MockCommander) GetAttitude() (int, int, int, error)     { return 0, 0, 0, nil }
func (m *MockCommander) GetBarometer() (int, error)              { return 1013, nil }
func (m *MockCommander) GetAcceleration() (int, int, int, error) { return 0, 0, 1000, nil }
func (m *MockCommander) GetTof() (int, error)                    { return 200, nil }

// Video Commands
func (m *MockCommander) SetVideoFrameCallback(callback func(transport.VideoFrame)) {}
func (m *MockCommander) GetVideoFrameChannel() <-chan transport.VideoFrame {
	ch := make(chan transport.VideoFrame)
	close(ch)
	return ch
}

// Reset clears all tracking state
func (m *MockCommander) Reset() {
	m.initCalled = false
	m.takeoffCalled = false
	m.landCalled = false
	m.emergencyCalled = false
	m.upCalled = false
	m.forwardCalled = false
	m.flipCalled = false
	m.initError = nil
	m.takeoffError = nil
	m.landError = nil
	m.emergencyError = nil
	m.upError = nil
	m.forwardError = nil
	m.flipError = nil
}

// TestSafetyFlow_Integration tests basic safety flow with command wrapper
func TestSafetyFlow_Integration(t *testing.T) {
	t.Run("safety manager wraps commander and validates commands", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		if manager == nil {
			t.Fatal("Expected non-nil SafetyManager")
		}

		// Test that safety manager is properly initialized
		status := manager.GetSafetyStatus()
		if status == nil {
			t.Fatal("Expected non-nil safety status")
		}

		if !status.SafetyEnabled {
			t.Error("Expected safety to be enabled by default")
		}

		// Test command passes through to wrapped commander
		err := manager.Init()
		if err != nil {
			t.Errorf("Expected Init to succeed, got error: %v", err)
		}
		if !mockCmdr.initCalled {
			t.Error("Expected wrapped commander Init to be called")
		}
	})

	t.Run("takeoff command flows through safety manager", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		err := manager.TakeOff()
		if err != nil {
			t.Errorf("Expected TakeOff to succeed, got error: %v", err)
		}
		if !mockCmdr.takeoffCalled {
			t.Error("Expected wrapped commander TakeOff to be called")
		}
	})

	t.Run("movement commands are validated by safety manager", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Set a safe height state
		state := &types.State{
			H:   100, // 100cm height
			Bat: 85,
		}
		manager.UpdateState(state)

		// Test forward movement within limits
		err := manager.Forward(50)
		if err != nil {
			t.Errorf("Expected Forward to succeed within limits, got error: %v", err)
		}
		if !mockCmdr.forwardCalled {
			t.Error("Expected wrapped commander Forward to be called")
		}
	})

	t.Run("safety manager blocks unsafe commands", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		// Disable flips for safety
		config.Behavioral.EnableFlips = false
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Set height for flip test
		state := &types.State{
			H:   150,
			Bat: 85,
		}
		manager.UpdateState(state)

		// Try flip when disabled - should be blocked
		err := manager.Flip("left")
		if err == nil {
			t.Error("Expected Flip to be blocked when disabled")
		}
		if mockCmdr.flipCalled {
			t.Error("Expected wrapped commander Flip NOT to be called when blocked")
		}
	})

	t.Run("safety manager allows safe commands after state update", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		config.Behavioral.EnableFlips = true
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Set sufficient height for flip
		state := &types.State{
			H:   150, // Above MinFlipHeight of 100cm
			Bat: 85,
		}
		manager.UpdateState(state)

		// Flip should be allowed now
		err := manager.Flip("left")
		if err != nil {
			t.Errorf("Expected Flip to succeed with sufficient height, got error: %v", err)
		}
		if !mockCmdr.flipCalled {
			t.Error("Expected wrapped commander Flip to be called")
		}
	})

	t.Run("emergency command always allowed", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Even with safety enabled, emergency should work
		err := manager.Emergency()
		if err != nil {
			t.Errorf("Expected Emergency to always succeed, got error: %v", err)
		}
		if !mockCmdr.emergencyCalled {
			t.Error("Expected wrapped commander Emergency to be called")
		}
	})

	t.Run("land command always allowed for safety", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Land should always be allowed
		err := manager.Land()
		if err != nil {
			t.Errorf("Expected Land to always succeed, got error: %v", err)
		}
		if !mockCmdr.landCalled {
			t.Error("Expected wrapped commander Land to be called")
		}
	})
}

// TestSafetyFlow_ErrorHandling tests error propagation through safety system
func TestSafetyFlow_ErrorHandling(t *testing.T) {
	t.Run("commander errors propagate through safety manager", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		mockCmdr.initError = errors.New("init failed")
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		err := manager.Init()
		if err == nil {
			t.Error("Expected error from wrapped commander to propagate")
		}
		if err.Error() != "init failed" {
			t.Errorf("Expected 'init failed' error, got: %v", err)
		}
	})

	t.Run("safety validation errors take precedence", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		config.Behavioral.EnableFlips = false
		manager := safety.NewSafetyManager(mockCmdr, config)

		state := &types.State{
			H:   150,
			Bat: 85,
		}
		manager.UpdateState(state)

		// Try flip when disabled - should get safety error, not call commander
		err := manager.Flip("left")
		if err == nil {
			t.Error("Expected safety validation error")
		}
		if err != nil && err.Error() == "" {
			t.Error("Expected error message from safety validation")
		}
		// Commander should not be called when safety blocks
		if mockCmdr.flipCalled {
			t.Error("Expected commander not to be called when safety blocks")
		}
	})

	t.Run("altitude limit violations are caught", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Set height near max
		state := &types.State{
			H:   250, // 250cm current height
			Bat: 85,
		}
		manager.UpdateState(state)

		// Try to go to z=350 - would exceed 300cm max
		// Go command uses z parameter for altitude validation
		err := manager.Go(0, 0, 350, 50)
		if err == nil {
			t.Error("Expected altitude limit error")
		}
		if mockCmdr.forwardCalled {
			t.Error("Expected commander not to be called when altitude limit exceeded")
		}
	})

	t.Run("disabled safety allows all commands", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		mockCmdr.upError = errors.New("up command failed")
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Disable safety
		manager.SetSafetyEnabled(false)

		// Even unsafe commands should pass through (but may fail at commander level)
		err := manager.Up(500) // Would normally be blocked
		if err == nil {
			t.Error("Expected error from wrapped commander")
		}
		if !mockCmdr.upCalled {
			t.Error("Expected commander to be called when safety disabled")
		}
	})

	t.Run("emergency mode bypasses safety checks", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		mockCmdr.forwardError = errors.New("forward failed")
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Enable emergency mode
		manager.SetEmergencyMode(true)

		// Commands should pass through in emergency mode
		err := manager.Forward(100)
		if err == nil {
			t.Error("Expected error from wrapped commander")
		}
		if !mockCmdr.forwardCalled {
			t.Error("Expected commander to be called in emergency mode")
		}
	})

	t.Run("safety events recorded for violations", func(t *testing.T) {
		mockCmdr := NewMockCommander()
		config := safety.DefaultConfig()
		manager := safety.NewSafetyManager(mockCmdr, config)

		// Set low battery state
		state := &types.State{
			H:   100,
			Bat: 10, // Below emergency threshold
		}
		manager.UpdateState(state)

		// Check that safety events were recorded
		events := manager.GetSafetyEvents()
		if len(events) == 0 {
			t.Error("Expected safety events to be recorded for low battery")
		}

		// Verify at least one battery event exists
		foundBatteryEvent := false
		for _, event := range events {
			if event.Type == "battery" {
				foundBatteryEvent = true
				break
			}
		}
		if !foundBatteryEvent {
			t.Error("Expected battery safety event to be recorded")
		}
	})
}
