package gamepad

import (
	"encoding/json"
	"time"
)

// AxisType represents the type of axis on a gamepad
type AxisType string

const (
	AxisLeftStickX   AxisType = "left_stick_x"
	AxisLeftStickY   AxisType = "left_stick_y"
	AxisRightStickX  AxisType = "right_stick_x"
	AxisRightStickY  AxisType = "right_stick_y"
	AxisLeftTrigger  AxisType = "left_trigger"
	AxisRightTrigger AxisType = "right_trigger"
	AxisDPadX        AxisType = "dpad_x"
	AxisDPadY        AxisType = "dpad_y"
)

// ButtonType represents the type of button on a gamepad
type ButtonType string

const (
	ButtonA          ButtonType = "button_a"
	ButtonB          ButtonType = "button_b"
	ButtonX          ButtonType = "button_x"
	ButtonY          ButtonType = "button_y"
	LeftBumper       ButtonType = "left_bumper"
	RightBumper      ButtonType = "right_bumper"
	LeftTrigger      ButtonType = "left_trigger"
	RightTrigger     ButtonType = "right_trigger"
	ButtonSelect     ButtonType = "button_select"
	ButtonStart      ButtonType = "button_start"
	LeftStickButton  ButtonType = "left_stick_button"
	RightStickButton ButtonType = "right_stick_button"
	DPadUp           ButtonType = "dpad_up"
	DPadDown         ButtonType = "dpad_down"
	DPadLeft         ButtonType = "dpad_left"
	DPadRight        ButtonType = "dpad_right"
)

// AxisMapping defines how an axis is mapped to drone control
type AxisMapping struct {
	Axis        AxisType `json:"axis"`
	Invert      bool     `json:"invert"`
	Deadzone    *float64 `json:"deadzone,omitempty"`
	Sensitivity *float64 `json:"sensitivity,omitempty"`
}

// ButtonMapping defines how a button is mapped to drone actions
type ButtonMapping struct {
	Button    ButtonType `json:"button"`
	HoldTime  int        `json:"hold_time,omitempty"`  // milliseconds
	DoubleTap bool       `json:"double_tap,omitempty"` // require double tap
}

// AxesMapping contains all axis mappings
type AxesMapping struct {
	MovementX AxisMapping `json:"movement_x"`
	MovementY AxisMapping `json:"movement_y"`
	Altitude  AxisMapping `json:"altitude"`
	Yaw       AxisMapping `json:"yaw"`
}

// ButtonsMapping contains all button mappings
type ButtonsMapping struct {
	TakeoffLand  ButtonMapping  `json:"takeoff_land"`
	Emergency    ButtonMapping  `json:"emergency"`
	FlipForward  ButtonMapping  `json:"flip_forward"`
	FlipBackward ButtonMapping  `json:"flip_backward"`
	FlipLeft     *ButtonMapping `json:"flip_left,omitempty"`
	FlipRight    *ButtonMapping `json:"flip_right,omitempty"`
	StreamToggle *ButtonMapping `json:"stream_toggle,omitempty"`
}

// Mappings contains all controller mappings
type Mappings struct {
	Axes    AxesMapping    `json:"axes"`
	Buttons ButtonsMapping `json:"buttons"`
}

// RCLimits defines safety limits for RC control values
type RCLimits struct {
	Horizontal int `json:"horizontal"`
	Vertical   int `json:"vertical"`
	Yaw        int `json:"yaw"`
}

// EmergencyActions defines safety behaviors
type EmergencyActions struct {
	ConnectionTimeout   int  `json:"connection_timeout"`    // milliseconds
	LowBatteryThreshold int  `json:"low_battery_threshold"` // percentage
	EnableAutoLand      bool `json:"enable_auto_land"`
}

// Safety contains all safety-related configuration
type Safety struct {
	RCLimits         RCLimits         `json:"rc_limits"`
	EmergencyActions EmergencyActions `json:"emergency_actions"`
}

// ControllerConfig contains controller-specific settings
type ControllerConfig struct {
	Deadzone     float64 `json:"deadzone"`
	Sensitivity  float64 `json:"sensitivity"`
	UpdateRate   int     `json:"update_rate"` // Hz
	AutoDetect   bool    `json:"auto_detect"`
	ControllerID *string `json:"controller_id,omitempty"`
}

// Config is the main gamepad configuration structure
type Config struct {
	Version    string           `json:"version"`
	Controller ControllerConfig `json:"controller"`
	Safety     Safety           `json:"safety"`
	Mappings   Mappings         `json:"mappings"`
}

// ButtonState represents the current state of a button
type ButtonState struct {
	Pressed     bool
	PressTime   time.Time
	LastRelease time.Time
	TapCount    int
	LastTapTime time.Time
}

// AxisState represents the current state of an axis
type AxisState struct {
	Value     float64
	LastValue float64
}

// GamepadState represents the complete current state of the gamepad
type GamepadState struct {
	Buttons    map[ButtonType]*ButtonState
	Axes       map[AxisType]*AxisState
	LastUpdate time.Time
}

// RCValues represents the current RC control values to send to the drone
type RCValues struct {
	A int // left/right (-100 to 100)
	B int // forward/backward (-100 to 100)
	C int // up/down (-100 to 100)
	D int // yaw (-100 to 100)
}

// DroneAction represents a discrete drone action triggered by button presses
type DroneAction string

const (
	ActionTakeoff      DroneAction = "takeoff"
	ActionLand         DroneAction = "land"
	ActionEmergency    DroneAction = "emergency"
	ActionFlipForward  DroneAction = "flip_forward"
	ActionFlipBackward DroneAction = "flip_backward"
	ActionFlipLeft     DroneAction = "flip_left"
	ActionFlipRight    DroneAction = "flip_right"
	ActionStreamOn     DroneAction = "stream_on"
	ActionStreamOff    DroneAction = "stream_off"
)

// CommandType represents the type of command
type CommandType string

const (
	CommandRC     CommandType = "rc"     // RC control values
	CommandAction CommandType = "action" // Discrete drone action
)

// Command represents a high-level drone command
type Command struct {
	Type CommandType `json:"type"`
	Data interface{} `json:"data"` // RCValues for rc, DroneAction for action
}

// Mapper defines the interface for converting gamepad events to drone commands
type Mapper interface {
	// MapEvent converts a gamepad event to one or more drone commands
	MapEvent(event Event) ([]Command, error)

	// MapState converts the current gamepad state to drone commands
	MapState(state *GamepadState) ([]Command, error)

	// UpdateConfig updates the mapper configuration
	UpdateConfig(config *Config) error
}

// Event represents a normalized gamepad input event
type Event struct {
	Type      EventType `json:"type"`
	Input     string    `json:"input"` // ButtonType or AxisType as string
	Value     float64   `json:"value"` // Normalized value: 0.0/1.0 for buttons, -1.0 to 1.0 for axes
	Timestamp time.Time `json:"timestamp"`
}

// EventType represents the type of gamepad event
type EventType string

const (
	EventButtonPress   EventType = "button_press"
	EventButtonRelease EventType = "button_release"
	EventAxisChange    EventType = "axis_change"
)

// InputEvent represents a gamepad input event (legacy, kept for compatibility)
type InputEvent struct {
	Type      string      // "button_press", "button_release", "axis_change"
	Input     interface{} // ButtonType or AxisType
	Value     interface{} // bool for buttons, float64 for axes
	Timestamp time.Time
}

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

// NewGamepadState creates a new gamepad state with all buttons and axes initialized
func NewGamepadState() *GamepadState {
	buttons := make(map[ButtonType]*ButtonState)
	axes := make(map[AxisType]*AxisState)

	// Initialize all buttons
	buttonTypes := []ButtonType{
		ButtonA, ButtonB, ButtonX, ButtonY,
		LeftBumper, RightBumper, LeftTrigger, RightTrigger,
		ButtonSelect, ButtonStart, LeftStickButton, RightStickButton,
		DPadUp, DPadDown, DPadLeft, DPadRight,
	}

	for _, btn := range buttonTypes {
		buttons[btn] = &ButtonState{
			Pressed:     false,
			PressTime:   time.Time{},
			LastRelease: time.Time{},
			TapCount:    0,
			LastTapTime: time.Time{},
		}
	}

	// Initialize all axes
	axisTypes := []AxisType{
		AxisLeftStickX, AxisLeftStickY,
		AxisRightStickX, AxisRightStickY,
		AxisLeftTrigger, AxisRightTrigger,
		AxisDPadX, AxisDPadY,
	}

	for _, axis := range axisTypes {
		axes[axis] = &AxisState{
			Value:     0.0,
			LastValue: 0.0,
		}
	}

	return &GamepadState{
		Buttons:    buttons,
		Axes:       axes,
		LastUpdate: time.Now(),
	}
}

// NewRCValues creates new RC values with all zeros
func NewRCValues() RCValues {
	return RCValues{A: 0, B: 0, C: 0, D: 0}
}

// IsZero returns true if all RC values are zero
func (rc RCValues) IsZero() bool {
	return rc.A == 0 && rc.B == 0 && rc.C == 0 && rc.D == 0
}

// Clamp ensures RC values are within valid range
func (rc RCValues) Clamp(limits RCLimits) RCValues {
	if rc.A > limits.Horizontal {
		rc.A = limits.Horizontal
	} else if rc.A < -limits.Horizontal {
		rc.A = -limits.Horizontal
	}

	if rc.B > limits.Horizontal {
		rc.B = limits.Horizontal
	} else if rc.B < -limits.Horizontal {
		rc.B = -limits.Horizontal
	}

	if rc.C > limits.Vertical {
		rc.C = limits.Vertical
	} else if rc.C < -limits.Vertical {
		rc.C = -limits.Vertical
	}

	if rc.D > limits.Yaw {
		rc.D = limits.Yaw
	} else if rc.D < -limits.Yaw {
		rc.D = -limits.Yaw
	}

	return rc
}
