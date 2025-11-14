package safety

import (
	"encoding/json"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// SafetyLevel represents the safety configuration level
type SafetyLevel string

const (
	SafetyLevelConservative SafetyLevel = "conservative"
	SafetyLevelNormal       SafetyLevel = "normal"
	SafetyLevelAggressive   SafetyLevel = "aggressive"
	SafetyLevelIndoor       SafetyLevel = "indoor"
	SafetyLevelOutdoor      SafetyLevel = "outdoor"
)

// SafetyEvent represents a safety-related event
type SafetyEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"` // "info", "warning", "critical", "emergency"
	Type      string                 `json:"type"`  // "altitude", "battery", "sensor", "behavioral"
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SafetyStatus represents the current safety status
type SafetyStatus struct {
	IsSafe        bool          `json:"is_safe"`
	ActiveEvents  []SafetyEvent `json:"active_events"`
	LastEvent     *SafetyEvent  `json:"last_event,omitempty"`
	ConfigLevel   SafetyLevel   `json:"config_level"`
	SafetyEnabled bool          `json:"safety_enabled"`
	EmergencyMode bool          `json:"emergency_mode"`
	CurrentState  *types.State  `json:"current_state,omitempty"`
}

// AltitudeLimits defines altitude safety limits
type AltitudeLimits struct {
	MinHeight     int `json:"min_height"`     // cm - minimum safe altitude
	MaxHeight     int `json:"max_height"`     // cm - maximum safe altitude
	TakeoffHeight int `json:"takeoff_height"` // cm - recommended takeoff altitude
}

// VelocityLimits defines velocity safety limits
type VelocityLimits struct {
	MaxHorizontal int `json:"max_horizontal"` // cm/s - maximum horizontal velocity
	MaxVertical   int `json:"max_vertical"`   // cm/s - maximum vertical velocity
	MaxYaw        int `json:"max_yaw"`        // degrees/s - maximum yaw rate
}

// BatterySafety defines battery-related safety settings
type BatterySafety struct {
	WarningThreshold   int    `json:"warning_threshold"`   // percentage - warning level
	CriticalThreshold  int    `json:"critical_threshold"`  // percentage - critical level
	EmergencyThreshold int    `json:"emergency_threshold"` // percentage - emergency level
	EnableAutoLand     bool   `json:"enable_auto_land"`    // auto-land on low battery
	LowBatteryAction   string `json:"low_battery_action"`  // "land", "hover", "emergency"
}

// SensorSafety defines sensor-related safety settings
type SensorSafety struct {
	MinTOFDistance      int     `json:"min_tof_distance"`      // cm - minimum TOF distance
	MaxTiltAngle        int     `json:"max_tilt_angle"`        // degrees - maximum tilt angle
	MaxAcceleration     float64 `json:"max_acceleration"`      // G-force - maximum acceleration
	BaroPressureDelta   float64 `json:"baro_pressure_delta"`   // mbar - pressure change threshold
	SensorFailureAction string  `json:"sensor_failure_action"` // "land", "hover", "emergency"
}

// EmergencyProcedures defines emergency behavior settings
type EmergencyProcedures struct {
	ConnectionTimeout   int    `json:"connection_timeout"`    // milliseconds - connection timeout
	SensorFailureAction string `json:"sensor_failure_action"` // "land", "hover", "emergency"
	EnableAutoLand      bool   `json:"enable_auto_land"`      // auto-land on emergencies
	LowBatteryAction    string `json:"low_battery_action"`    // "land", "hover", "emergency"
}

// BehavioralLimits defines behavioral safety settings
type BehavioralLimits struct {
	EnableFlips    bool `json:"enable_flips"`     // allow flip maneuvers
	MinFlipHeight  int  `json:"min_flip_height"`  // cm - minimum altitude for flips
	MaxFlightTime  int  `json:"max_flight_time"`  // seconds - maximum continuous flight time
	MaxCommandRate int  `json:"max_command_rate"` // commands/second - rate limiting
}

// Config is the main safety configuration structure
type Config struct {
	Version    string              `json:"version"`
	Level      SafetyLevel         `json:"level"`
	Altitude   AltitudeLimits      `json:"altitude"`
	Velocity   VelocityLimits      `json:"velocity"`
	Battery    BatterySafety       `json:"battery"`
	Sensors    SensorSafety        `json:"sensors"`
	Emergency  EmergencyProcedures `json:"emergency"`
	Behavioral BehavioralLimits    `json:"behavioral"`
}

// CommandValidationResult represents the result of command validation
type CommandValidationResult struct {
	Allowed bool         `json:"allowed"`
	Reason  string       `json:"reason,omitempty"`
	Event   *SafetyEvent `json:"event,omitempty"`
}

// SafetyAction represents actions the safety manager can take
type SafetyAction string

const (
	SafetyActionLand      SafetyAction = "land"
	SafetyActionHover     SafetyAction = "hover"
	SafetyActionEmergency SafetyAction = "emergency"
	SafetyActionNone      SafetyAction = "none"
)

// SafetyEventType represents types of safety events
type SafetyEventType string

const (
	SafetyEventAltitude   SafetyEventType = "altitude"
	SafetyEventBattery    SafetyEventType = "battery"
	SafetyEventSensor     SafetyEventType = "sensor"
	SafetyEventBehavioral SafetyEventType = "behavioral"
	SafetyEventConnection SafetyEventType = "connection"
	SafetyEventEmergency  SafetyEventType = "emergency"
)

// SafetyEventLevel represents severity levels of safety events
type SafetyEventLevel string

const (
	SafetyEventLevelInfo      SafetyEventLevel = "info"
	SafetyEventLevelWarning   SafetyEventLevel = "warning"
	SafetyEventLevelCritical  SafetyEventLevel = "critical"
	SafetyEventLevelEmergency SafetyEventLevel = "emergency"
)

// MarshalJSON implements custom JSON marshaling for Config
func (c *Config) MarshalJSON() ([]byte, error) {
	type Alias Config
	return json.Marshal((*Alias)(c))
}

// UnmarshalJSON implements custom JSON unmarshaling for Config
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(data, &aux.Alias)
}

// IsValidSafetyLevel checks if a safety level is valid
func IsValidSafetyLevel(level string) bool {
	validLevels := []SafetyLevel{
		SafetyLevelConservative,
		SafetyLevelNormal,
		SafetyLevelAggressive,
		SafetyLevelIndoor,
		SafetyLevelOutdoor,
	}

	for _, validLevel := range validLevels {
		if string(validLevel) == level {
			return true
		}
	}
	return false
}

// IsValidSafetyAction checks if a safety action is valid
func IsValidSafetyAction(action string) bool {
	validActions := []SafetyAction{
		SafetyActionLand,
		SafetyActionHover,
		SafetyActionEmergency,
		SafetyActionNone,
	}

	for _, validAction := range validActions {
		if string(validAction) == action {
			return true
		}
	}
	return false
}

// NewSafetyEvent creates a new safety event
func NewSafetyEvent(eventType SafetyEventType, level SafetyEventLevel, message string, data map[string]interface{}) *SafetyEvent {
	return &SafetyEvent{
		Timestamp: time.Now(),
		Level:     string(level),
		Type:      string(eventType),
		Message:   message,
		Data:      data,
	}
}

// NewSafetyStatus creates a new safety status
func NewSafetyStatus() *SafetyStatus {
	return &SafetyStatus{
		IsSafe:        true,
		ActiveEvents:  make([]SafetyEvent, 0),
		SafetyEnabled: true,
		EmergencyMode: false,
	}
}
