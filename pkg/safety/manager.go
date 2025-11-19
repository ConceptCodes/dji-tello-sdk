package safety

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// CommanderInterface defines the interface for drone commands (same as tello.TelloCommander)
type CommanderInterface interface {
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
	Flip(direction interface{}) error
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
	SetVideoFrameCallback(callback interface{})
	GetVideoFrameChannel() <-chan transport.VideoFrame
}

// SafetyManager wraps CommanderInterface to provide safety validation and monitoring
type SafetyManager struct {
	commander     CommanderInterface
	config        *Config
	status        *SafetyStatus
	mutex         sync.RWMutex
	eventCallback func(*SafetyEvent)

	// State tracking
	lastStateUpdate time.Time
	flightStartTime time.Time
	commandCount    int
	lastCommandTime time.Time

	// Telemetry processing
	stateChan       <-chan *types.State
	telemetryWg     sync.WaitGroup
	telemetryCtx    context.Context
	telemetryCancel context.CancelFunc

	// Emergency state
	emergencyMode bool
	safetyEnabled bool
}

// NewSafetyManager creates a new safety manager
func NewSafetyManager(commander interface{}, config *Config) *SafetyManager {
	// Type assert to ensure commander implements the required interface
	cmd, ok := commander.(CommanderInterface)
	if !ok {
		// If it doesn't implement our interface, we can't create a safety manager
		// This should never happen if used correctly
		return nil
	}

	sm := &SafetyManager{
		commander:     cmd,
		config:        config,
		status:        NewSafetyStatus(),
		safetyEnabled: true,
		emergencyMode: false,
	}

	sm.status.ConfigLevel = config.Level

	return sm
}

// StartTelemetryProcessing starts continuous telemetry processing from the state channel
func (sm *SafetyManager) StartTelemetryProcessing(stateChan <-chan *types.State) {
	sm.stateChan = stateChan
	sm.telemetryCtx, sm.telemetryCancel = context.WithCancel(context.Background())

	sm.telemetryWg.Add(1)
	go func() {
		defer sm.telemetryWg.Done()
		sm.processTelemetry()
	}()
}

// processTelemetry continuously processes telemetry data from the state channel
func (sm *SafetyManager) processTelemetry() {
	for {
		select {
		case <-sm.telemetryCtx.Done():
			return
		case state, ok := <-sm.stateChan:
			if !ok {
				utils.Logger.Info("Telemetry channel closed")
				return
			}
			sm.UpdateState(state)
		}
	}
}

// UpdateState updates the safety manager with current drone state
func (sm *SafetyManager) UpdateState(state *types.State) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.status.CurrentState = state
	sm.lastStateUpdate = time.Now()

	// Perform safety checks
	sm.checkAltitudeSafety(state)
	sm.checkBatterySafety(state)
	sm.checkSensorSafety(state)
	sm.checkBehavioralSafety(state)

	// Update overall safety status
	sm.updateSafetyStatus()
}

// StopTelemetryProcessing stops the telemetry processing goroutine
func (sm *SafetyManager) StopTelemetryProcessing() {
	if sm.telemetryCancel != nil {
		sm.telemetryCancel()
	}
	sm.telemetryWg.Wait()
}

// SetEventCallback sets a callback for safety events
func (sm *SafetyManager) SetEventCallback(callback func(*SafetyEvent)) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.eventCallback = callback
}

// GetSafetyStatus returns current safety status
func (sm *SafetyManager) GetSafetyStatus() *SafetyStatus {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to avoid concurrent access issues
	statusCopy := *sm.status
	if sm.status.CurrentState != nil {
		stateCopy := *sm.status.CurrentState
		statusCopy.CurrentState = &stateCopy
	}

	return &statusCopy
}

// GetSafetyEvents returns recent safety events
func (sm *SafetyManager) GetSafetyEvents() []SafetyEvent {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	events := make([]SafetyEvent, len(sm.status.ActiveEvents))
	copy(events, sm.status.ActiveEvents)

	return events
}

// SetSafetyEnabled enables or disables safety checks
func (sm *SafetyManager) SetSafetyEnabled(enabled bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.safetyEnabled = enabled
	sm.status.SafetyEnabled = enabled

	if enabled {
		utils.Logger.Info("Safety manager enabled")
	} else {
		utils.Logger.Warn("Safety manager disabled - use with caution!")
	}
}

// SetEmergencyMode sets emergency mode
func (sm *SafetyManager) SetEmergencyMode(emergency bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.emergencyMode = emergency
	sm.status.EmergencyMode = emergency

	if emergency {
		event := NewSafetyEvent(SafetyEventEmergency, SafetyEventLevelEmergency,
			"Emergency mode activated", map[string]any{"manual": true})
		sm.addEvent(event)
		utils.Logger.Error("Emergency mode activated!")
	}
}

// SetSafetyConfig updates the safety configuration
func (sm *SafetyManager) SetSafetyConfig(config *Config) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.config = config
	sm.status.ConfigLevel = config.Level

	utils.Logger.Infof("Safety configuration updated to: %s", config.Level)
}

// TelloCommander interface implementation

func (sm *SafetyManager) Init() error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Init()
	}

	// Validate initialization
	result := sm.validateCommand("init", map[string]any{})
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Init()
}

func (sm *SafetyManager) TakeOff() error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.TakeOff()
	}

	// Validate takeoff
	result := sm.validateCommand("takeoff", map[string]any{})
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	// Record flight start time
	sm.mutex.Lock()
	sm.flightStartTime = time.Now()
	sm.mutex.Unlock()

	return sm.commander.TakeOff()
}

func (sm *SafetyManager) Land() error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Land()
	}

	// Always allow landing for safety
	return sm.commander.Land()
}

func (sm *SafetyManager) StreamOn() error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.StreamOn()
	}

	// Validate stream on
	result := sm.validateCommand("streamon", map[string]any{})
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.StreamOn()
}

func (sm *SafetyManager) StreamOff() error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.StreamOff()
	}

	// Always allow stream off for safety
	return sm.commander.StreamOff()
}

func (sm *SafetyManager) Emergency() error {
	// Emergency is always allowed
	return sm.commander.Emergency()
}

func (sm *SafetyManager) Up(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Up(distance)
	}

	result := sm.validateMovementCommand("up", distance, 0, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Up(distance)
}

func (sm *SafetyManager) Down(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Down(distance)
	}

	result := sm.validateMovementCommand("down", -distance, 0, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Down(distance)
}

func (sm *SafetyManager) Left(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Left(distance)
	}

	result := sm.validateMovementCommand("left", 0, -distance, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Left(distance)
}

func (sm *SafetyManager) Right(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Right(distance)
	}

	result := sm.validateMovementCommand("right", 0, distance, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Right(distance)
}

func (sm *SafetyManager) Forward(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Forward(distance)
	}

	result := sm.validateMovementCommand("forward", distance, 0, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Forward(distance)
}

func (sm *SafetyManager) Backward(distance int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Backward(distance)
	}

	result := sm.validateMovementCommand("backward", -distance, 0, 0)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Backward(distance)
}

func (sm *SafetyManager) Clockwise(angle int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Clockwise(angle)
	}

	result := sm.validateRotationCommand("clockwise", angle)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Clockwise(angle)
}

func (sm *SafetyManager) CounterClockwise(angle int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.CounterClockwise(angle)
	}

	result := sm.validateRotationCommand("counterclockwise", angle)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.CounterClockwise(angle)
}

func (sm *SafetyManager) Flip(direction interface{}) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Flip(direction)
	}

	result := sm.validateFlipCommand(fmt.Sprintf("%v", direction))
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Flip(direction)
}

func (sm *SafetyManager) Go(x, y, z, speed int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Go(x, y, z, speed)
	}

	result := sm.validateGoCommand(x, y, z, speed)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Go(x, y, z, speed)
}

func (sm *SafetyManager) Curve(x1, y1, z1, x2, y2, z2, speed int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.Curve(x1, y1, z1, x2, y2, z2, speed)
	}

	result := sm.validateCurveCommand(x1, y1, z1, x2, y2, z2, speed)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.Curve(x1, y1, z1, x2, y2, z2, speed)
}

func (sm *SafetyManager) SetSpeed(speed int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.SetSpeed(speed)
	}

	result := sm.validateSpeedCommand(speed)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.SetSpeed(speed)
}

func (sm *SafetyManager) SetRcControl(a, b, c, d int) error {
	if !sm.safetyEnabled || sm.emergencyMode {
		return sm.commander.SetRcControl(a, b, c, d)
	}

	result := sm.validateRCCommand(a, b, c, d)
	if !result.Allowed {
		return fmt.Errorf("safety check failed: %s", result.Reason)
	}

	return sm.commander.SetRcControl(a, b, c, d)
}

func (sm *SafetyManager) SetWiFiCredentials(ssid, password string) error {
	return sm.commander.SetWiFiCredentials(ssid, password)
}

// Read commands (no safety validation needed)
func (sm *SafetyManager) GetSpeed() (int, error) {
	return sm.commander.GetSpeed()
}

func (sm *SafetyManager) GetBatteryPercentage() (int, error) {
	return sm.commander.GetBatteryPercentage()
}

func (sm *SafetyManager) GetTime() (int, error) {
	return sm.commander.GetTime()
}

func (sm *SafetyManager) GetHeight() (int, error) {
	return sm.commander.GetHeight()
}

func (sm *SafetyManager) GetTemperature() (int, error) {
	return sm.commander.GetTemperature()
}

func (sm *SafetyManager) GetAttitude() (int, int, int, error) {
	return sm.commander.GetAttitude()
}

func (sm *SafetyManager) GetBarometer() (int, error) {
	return sm.commander.GetBarometer()
}

func (sm *SafetyManager) GetAcceleration() (int, int, int, error) {
	return sm.commander.GetAcceleration()
}

func (sm *SafetyManager) GetTof() (int, error) {
	return sm.commander.GetTof()
}

// Video commands
func (sm *SafetyManager) SetVideoFrameCallback(callback interface{}) {
	sm.commander.SetVideoFrameCallback(callback)
}

func (sm *SafetyManager) GetVideoFrameChannel() <-chan transport.VideoFrame {
	return sm.commander.GetVideoFrameChannel()
}

// Private validation methods

func (sm *SafetyManager) validateCommand(command string, params map[string]any) CommandValidationResult {
	// Check command rate limiting
	if !sm.checkCommandRate() {
		return CommandValidationResult{
			Allowed: false,
			Reason:  "Command rate limit exceeded",
		}
	}

	// Check flight time limits
	if !sm.checkFlightTime() {
		return CommandValidationResult{
			Allowed: false,
			Reason:  "Maximum flight time exceeded",
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateMovementCommand(command string, x, y, z int) CommandValidationResult {
	baseResult := sm.validateCommand(command, map[string]any{
		"x": x, "y": y, "z": z,
	})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check altitude limits
	if sm.status.CurrentState != nil {
		newHeight := sm.status.CurrentState.H + z
		if newHeight > sm.config.Altitude.MaxHeight {
			return CommandValidationResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Altitude %dcm exceeds maximum %dcm", newHeight, sm.config.Altitude.MaxHeight),
			}
		}

		if newHeight < sm.config.Altitude.MinHeight {
			return CommandValidationResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Altitude %dcm below minimum %dcm", newHeight, sm.config.Altitude.MinHeight),
			}
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateRotationCommand(command string, angle int) CommandValidationResult {
	baseResult := sm.validateCommand(command, map[string]any{"angle": angle})
	if !baseResult.Allowed {
		return baseResult
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateFlipCommand(direction string) CommandValidationResult {
	baseResult := sm.validateCommand("flip", map[string]any{"direction": direction})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check if flips are enabled
	if !sm.config.Behavioral.EnableFlips {
		return CommandValidationResult{
			Allowed: false,
			Reason:  "Flips are disabled in current safety configuration",
		}
	}

	// Check minimum flip height
	if sm.status.CurrentState != nil && sm.status.CurrentState.H < sm.config.Behavioral.MinFlipHeight {
		return CommandValidationResult{
			Allowed: false,
			Reason: fmt.Sprintf("Altitude %dcm below minimum flip height %dcm",
				sm.status.CurrentState.H, sm.config.Behavioral.MinFlipHeight),
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateGoCommand(x, y, z, speed int) CommandValidationResult {
	baseResult := sm.validateCommand("go", map[string]any{
		"x": x, "y": y, "z": z, "speed": speed,
	})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check speed limits
	if speed > sm.config.Velocity.MaxHorizontal {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Speed %d exceeds maximum %d", speed, sm.config.Velocity.MaxHorizontal),
		}
	}

	// Check altitude limits
	if z > sm.config.Altitude.MaxHeight {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Target altitude %dcm exceeds maximum %dcm", z, sm.config.Altitude.MaxHeight),
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateCurveCommand(x1, y1, z1, x2, y2, z2, speed int) CommandValidationResult {
	baseResult := sm.validateCommand("curve", map[string]any{
		"x1": x1, "y1": y1, "z1": z1,
		"x2": x2, "y2": y2, "z2": z2,
		"speed": speed,
	})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check speed limits
	if speed > sm.config.Velocity.MaxHorizontal {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Speed %d exceeds maximum %d", speed, sm.config.Velocity.MaxHorizontal),
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateSpeedCommand(speed int) CommandValidationResult {
	baseResult := sm.validateCommand("speed", map[string]any{"speed": speed})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check speed limits
	if speed > sm.config.Velocity.MaxHorizontal {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Speed %d exceeds maximum %d", speed, sm.config.Velocity.MaxHorizontal),
		}
	}

	return CommandValidationResult{Allowed: true}
}

func (sm *SafetyManager) validateRCCommand(a, b, c, d int) CommandValidationResult {
	baseResult := sm.validateCommand("rc", map[string]any{
		"a": a, "b": b, "c": c, "d": d,
	})
	if !baseResult.Allowed {
		return baseResult
	}

	// Check RC limits
	if abs(a) > sm.config.Velocity.MaxHorizontal {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("RC value a=%d exceeds horizontal limit %d", a, sm.config.Velocity.MaxHorizontal),
		}
	}

	if abs(b) > sm.config.Velocity.MaxHorizontal {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("RC value b=%d exceeds horizontal limit %d", b, sm.config.Velocity.MaxHorizontal),
		}
	}

	if abs(c) > sm.config.Velocity.MaxVertical {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("RC value c=%d exceeds vertical limit %d", c, sm.config.Velocity.MaxVertical),
		}
	}

	if abs(d) > sm.config.Velocity.MaxYaw {
		return CommandValidationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("RC value d=%d exceeds yaw limit %d", d, sm.config.Velocity.MaxYaw),
		}
	}

	return CommandValidationResult{Allowed: true}
}

// Safety monitoring methods

func (sm *SafetyManager) checkAltitudeSafety(state *types.State) {
	// Check maximum altitude
	if state.H > sm.config.Altitude.MaxHeight {
		event := NewSafetyEvent(SafetyEventAltitude, SafetyEventLevelWarning,
			"Maximum altitude exceeded", map[string]any{
				"current_height": state.H,
				"max_height":     sm.config.Altitude.MaxHeight,
			})
		sm.addEvent(event)
	}

	// Check minimum altitude
	if state.H < sm.config.Altitude.MinHeight {
		event := NewSafetyEvent(SafetyEventAltitude, SafetyEventLevelWarning,
			"Minimum altitude violated", map[string]any{
				"current_height": state.H,
				"min_height":     sm.config.Altitude.MinHeight,
			})
		sm.addEvent(event)
	}
}

func (sm *SafetyManager) checkBatterySafety(state *types.State) {
	battery := state.Bat

	// Check emergency threshold
	if battery <= sm.config.Battery.EmergencyThreshold {
		event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelEmergency,
			"Emergency battery level", map[string]any{
				"battery_level": battery,
				"threshold":     sm.config.Battery.EmergencyThreshold,
			})
		sm.addEvent(event)

		// Trigger emergency landing
		if sm.config.Battery.EnableAutoLand {
			utils.Logger.Error("Emergency battery level - triggering auto-land")
			go sm.commander.Land()
		}
		return
	}

	// Check critical threshold
	if battery <= sm.config.Battery.CriticalThreshold {
		event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelCritical,
			"Critical battery level", map[string]any{
				"battery_level": battery,
				"threshold":     sm.config.Battery.CriticalThreshold,
			})
		sm.addEvent(event)
		return
	}

	// Check warning threshold
	if battery <= sm.config.Battery.WarningThreshold {
		event := NewSafetyEvent(SafetyEventBattery, SafetyEventLevelWarning,
			"Low battery level", map[string]any{
				"battery_level": battery,
				"threshold":     sm.config.Battery.WarningThreshold,
			})
		sm.addEvent(event)
	}
}

func (sm *SafetyManager) checkSensorSafety(state *types.State) {
	// Check TOF distance
	if state.Tof > 0 && state.Tof < sm.config.Sensors.MinTOFDistance {
		event := NewSafetyEvent(SafetyEventSensor, SafetyEventLevelWarning,
			"Obstacle detected - TOF distance too small", map[string]any{
				"tof_distance": state.Tof,
				"min_distance": sm.config.Sensors.MinTOFDistance,
			})
		sm.addEvent(event)
	}

	// Check tilt angles
	pitch, roll, _ := sm.getAttitudeAngles(state)
	maxTilt := max(abs(pitch), abs(roll))

	if maxTilt > sm.config.Sensors.MaxTiltAngle {
		event := NewSafetyEvent(SafetyEventSensor, SafetyEventLevelWarning,
			"Excessive tilt angle", map[string]any{
				"pitch":    pitch,
				"roll":     roll,
				"max_tilt": sm.config.Sensors.MaxTiltAngle,
			})
		sm.addEvent(event)
	}

	// Check acceleration
	accelMagnitude := math.Sqrt(state.Agx*state.Agx + state.Agy*state.Agy + state.Agz*state.Agz)
	if accelMagnitude > sm.config.Sensors.MaxAcceleration {
		event := NewSafetyEvent(SafetyEventSensor, SafetyEventLevelWarning,
			"Excessive acceleration", map[string]any{
				"acceleration": accelMagnitude,
				"max_accel":    sm.config.Sensors.MaxAcceleration,
			})
		sm.addEvent(event)
	}
}

func (sm *SafetyManager) checkBehavioralSafety(state *types.State) {
	// Check flight time
	if !sm.flightStartTime.IsZero() {
		flightTime := time.Since(sm.flightStartTime).Seconds()
		if flightTime > float64(sm.config.Behavioral.MaxFlightTime) {
			event := NewSafetyEvent(SafetyEventBehavioral, SafetyEventLevelWarning,
				"Maximum flight time exceeded", map[string]any{
					"flight_time": flightTime,
					"max_time":    sm.config.Behavioral.MaxFlightTime,
				})
			sm.addEvent(event)
		}
	}
}

func (sm *SafetyManager) checkCommandRate() bool {
	now := time.Now()
	if !sm.lastCommandTime.IsZero() {
		duration := now.Sub(sm.lastCommandTime)
		if duration < time.Second/time.Duration(sm.config.Behavioral.MaxCommandRate) {
			return false
		}
	}
	sm.lastCommandTime = now
	return true
}

func (sm *SafetyManager) checkFlightTime() bool {
	if sm.flightStartTime.IsZero() {
		return true
	}

	flightTime := time.Since(sm.flightStartTime).Seconds()
	return flightTime <= float64(sm.config.Behavioral.MaxFlightTime)
}

func (sm *SafetyManager) updateSafetyStatus() {
	// Clear old events (older than 5 minutes)
	now := time.Now()
	cutoff := now.Add(-5 * time.Minute)

	activeEvents := make([]SafetyEvent, 0)
	for _, event := range sm.status.ActiveEvents {
		if event.Timestamp.After(cutoff) {
			activeEvents = append(activeEvents, event)
		}
	}
	sm.status.ActiveEvents = activeEvents

	// Determine overall safety status
	sm.status.IsSafe = len(activeEvents) == 0 ||
		!sm.hasCriticalOrEmergencyEvents(activeEvents)

	if len(activeEvents) > 0 {
		sm.status.LastEvent = &activeEvents[len(activeEvents)-1]
	}
}

func (sm *SafetyManager) addEvent(event *SafetyEvent) {
	sm.status.ActiveEvents = append(sm.status.ActiveEvents, *event)
	sm.status.LastEvent = event

	// Log the event
	switch event.Level {
	case "emergency":
		utils.Logger.Errorf("Safety Emergency: %s", event.Message)
	case "critical":
		utils.Logger.Errorf("Safety Critical: %s", event.Message)
	case "warning":
		utils.Logger.Warnf("Safety Warning: %s", event.Message)
	default:
		utils.Logger.Infof("Safety Info: %s", event.Message)
	}

	// Call event callback if set
	if sm.eventCallback != nil {
		go sm.eventCallback(event)
	}
}

func (sm *SafetyManager) hasCriticalOrEmergencyEvents(events []SafetyEvent) bool {
	for _, event := range events {
		if event.Level == "critical" || event.Level == "emergency" {
			return true
		}
	}
	return false
}

func (sm *SafetyManager) getAttitudeAngles(state *types.State) (pitch, roll, yaw int) {
	return state.Pitch, state.Roll, state.Yaw
}

// Utility functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
