package gamepad

import (
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// DefaultConfig returns the default gamepad configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		Controller: ControllerConfig{
			Deadzone:     0.1,
			Sensitivity:  1.0,
			UpdateRate:   60,
			AutoDetect:   true,
			ControllerID: nil,
		},
		Safety: Safety{
			RCLimits: RCLimits{
				Horizontal: 80,
				Vertical:   60,
				Yaw:        100,
			},
			EmergencyActions: EmergencyActions{
				ConnectionTimeout:   3000,
				LowBatteryThreshold: 20,
				EnableAutoLand:      true,
			},
		},
		Mappings: Mappings{
			Axes: AxesMapping{
				MovementX: AxisMapping{
					Axis:        AxisLeftStickX,
					Invert:      false,
					Deadzone:    nil, // Use controller default
					Sensitivity: nil, // Use controller default
				},
				MovementY: AxisMapping{
					Axis:        AxisLeftStickY,
					Invert:      false,
					Deadzone:    nil,
					Sensitivity: nil,
				},
				Altitude: AxisMapping{
					Axis:        AxisRightStickY,
					Invert:      true, // Invert for intuitive up/down
					Deadzone:    nil,
					Sensitivity: nil,
				},
				Yaw: AxisMapping{
					Axis:        AxisRightStickX,
					Invert:      false,
					Deadzone:    nil,
					Sensitivity: nil,
				},
			},
			Buttons: ButtonsMapping{
				TakeoffLand: ButtonMapping{
					Button:    ButtonA,
					HoldTime:  500, // Hold 500ms for takeoff/land
					DoubleTap: false,
				},
				Emergency: ButtonMapping{
					Button:    ButtonB,
					HoldTime:  1000, // Hold 1s for emergency stop
					DoubleTap: false,
				},
				FlipForward: ButtonMapping{
					Button:    ButtonX,
					HoldTime:  0, // Instant action
					DoubleTap: false,
				},
				FlipBackward: ButtonMapping{
					Button:    ButtonY,
					HoldTime:  0,
					DoubleTap: false,
				},
				FlipLeft: &ButtonMapping{
					Button:    LeftBumper,
					HoldTime:  0,
					DoubleTap: false,
				},
				FlipRight: &ButtonMapping{
					Button:    RightBumper,
					HoldTime:  0,
					DoubleTap: false,
				},
				StreamToggle: &ButtonMapping{
					Button:    ButtonSelect,
					HoldTime:  0,
					DoubleTap: false,
				},
			},
		},
	}
}

// XboxConfig returns Xbox controller specific configuration
func XboxConfig() *Config {
	config := DefaultConfig()
	config.Version = "1.0.0-xbox"

	// Xbox-specific mappings
	config.Mappings.Buttons.TakeoffLand.Button = ButtonA
	config.Mappings.Buttons.Emergency.Button = ButtonB
	config.Mappings.Buttons.FlipForward.Button = ButtonX
	config.Mappings.Buttons.FlipBackward.Button = ButtonY
	config.Mappings.Buttons.FlipLeft.Button = LeftBumper
	config.Mappings.Buttons.FlipRight.Button = RightBumper
	config.Mappings.Buttons.StreamToggle.Button = ButtonSelect

	utils.Logger.Info("Created Xbox controller configuration")
	return config
}

// PlayStationConfig returns PlayStation controller specific configuration
func PlayStationConfig() *Config {
	config := DefaultConfig()
	config.Version = "1.0.0-ps"

	// PlayStation-specific mappings (different button layout)
	config.Mappings.Buttons.TakeoffLand.Button = ButtonA       // Cross button (mapped to A)
	config.Mappings.Buttons.Emergency.Button = ButtonB         // Circle button (mapped to B)
	config.Mappings.Buttons.FlipForward.Button = ButtonY       // Triangle button (mapped to Y)
	config.Mappings.Buttons.FlipBackward.Button = ButtonX      // Square button (mapped to X)
	config.Mappings.Buttons.FlipLeft.Button = LeftBumper       // L1
	config.Mappings.Buttons.FlipRight.Button = RightBumper     // R1
	config.Mappings.Buttons.StreamToggle.Button = ButtonSelect // Select button

	utils.Logger.Info("Created PlayStation controller configuration")
	return config
}

// GetPresetConfigs returns all available preset configurations
func GetPresetConfigs() map[string]*Config {
	return map[string]*Config{
		"default":     DefaultConfig(),
		"xbox":        XboxConfig(),
		"playstation": PlayStationConfig(),
	}
}

// GetConfigNames returns the names of all available preset configurations
func GetConfigNames() []string {
	return []string{"default", "xbox", "playstation"}
}
