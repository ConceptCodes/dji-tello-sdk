package gamepad

import (
	"fmt"
	"time"
)

// DefaultMapper implements the Mapper interface with configurable gamepad-to-drone mapping
type DefaultMapper struct {
	config *Config
}

// NewDefaultMapper creates a new mapper with the given configuration
func NewDefaultMapper(config *Config) *DefaultMapper {
	return &DefaultMapper{
		config: config,
	}
}

// MapEvent converts a gamepad event to one or more drone commands
func (m *DefaultMapper) MapEvent(event Event) ([]Command, error) {
	var commands []Command

	switch event.Type {
	case EventButtonPress:
		if event.Value == 1.0 { // Button pressed
			if action := m.mapButtonToAction(ButtonType(event.Input)); action != "" {
				commands = append(commands, Command{
					Type: CommandAction,
					Data: action,
				})
			}
		}
	case EventAxisChange:
		// Axis changes are handled via MapState for continuous updates
		// Individual axis events could be handled here if needed
	}

	return commands, nil
}

// MapState converts the current gamepad state to drone commands
func (m *DefaultMapper) MapState(state *GamepadState) ([]Command, error) {
	var commands []Command

	// Process axes for RC control
	rcValues := m.processAxes(state)
	if !rcValues.IsZero() {
		// Apply safety limits
		rcValues = rcValues.Clamp(m.config.Safety.RCLimits)

		commands = append(commands, Command{
			Type: CommandRC,
			Data: rcValues,
		})
	}

	// Process buttons for actions
	actions := m.processButtons(state)
	for _, action := range actions {
		commands = append(commands, Command{
			Type: CommandAction,
			Data: action,
		})
	}

	return commands, nil
}

// UpdateConfig updates the mapper configuration
func (m *DefaultMapper) UpdateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	m.config = config
	return nil
}

// processAxes converts axis states to RC values
func (m *DefaultMapper) processAxes(state *GamepadState) RCValues {
	rc := NewRCValues()

	// Process Movement X (left/right)
	if axis, exists := state.Axes[m.config.Mappings.Axes.MovementX.Axis]; exists {
		value := m.applyAxisProcessing(axis.Value, m.config.Mappings.Axes.MovementX)
		rc.A = int(value * float64(m.config.Safety.RCLimits.Horizontal))
	}

	// Process Movement Y (forward/backward)
	if axis, exists := state.Axes[m.config.Mappings.Axes.MovementY.Axis]; exists {
		value := m.applyAxisProcessing(axis.Value, m.config.Mappings.Axes.MovementY)
		rc.B = int(value * float64(m.config.Safety.RCLimits.Horizontal))
	}

	// Process Altitude (up/down)
	if axis, exists := state.Axes[m.config.Mappings.Axes.Altitude.Axis]; exists {
		value := m.applyAxisProcessing(axis.Value, m.config.Mappings.Axes.Altitude)
		rc.C = int(value * float64(m.config.Safety.RCLimits.Vertical))
	}

	// Process Yaw (rotation)
	if axis, exists := state.Axes[m.config.Mappings.Axes.Yaw.Axis]; exists {
		value := m.applyAxisProcessing(axis.Value, m.config.Mappings.Axes.Yaw)
		rc.D = int(value * float64(m.config.Safety.RCLimits.Yaw))
	}

	return rc
}

// applyAxisProcessing applies deadzone, sensitivity, and inversion to axis values
func (m *DefaultMapper) applyAxisProcessing(value float64, mapping AxisMapping) float64 {
	// Apply deadzone
	deadzone := m.config.Controller.Deadzone
	if mapping.Deadzone != nil {
		deadzone = *mapping.Deadzone
	}

	if abs(value) < deadzone {
		return 0.0
	}

	// Apply sensitivity
	sensitivity := m.config.Controller.Sensitivity
	if mapping.Sensitivity != nil {
		sensitivity = *mapping.Sensitivity
	}
	value *= sensitivity

	// Apply inversion
	if mapping.Invert {
		value = -value
	}

	// Clamp to [-1.0, 1.0]
	if value > 1.0 {
		return 1.0
	} else if value < -1.0 {
		return -1.0
	}

	return value
}

// processButtons converts button states to drone actions
func (m *DefaultMapper) processButtons(state *GamepadState) []DroneAction {
	var actions []DroneAction
	now := time.Now()

	// Check Takeoff/Land button
	if button, exists := state.Buttons[m.config.Mappings.Buttons.TakeoffLand.Button]; exists {
		if m.shouldTriggerButtonAction(button, m.config.Mappings.Buttons.TakeoffLand, now) {
			// Determine takeoff vs land based on current state or toggle logic
			// For now, we'll use a simple toggle based on press duration
			if button.Pressed && time.Since(button.PressTime) >= time.Duration(m.config.Mappings.Buttons.TakeoffLand.HoldTime)*time.Millisecond {
				actions = append(actions, ActionTakeoff)
			} else if !button.Pressed && time.Since(button.PressTime) < time.Duration(m.config.Mappings.Buttons.TakeoffLand.HoldTime)*time.Millisecond {
				actions = append(actions, ActionLand)
			}
		}
	}

	// Check Emergency button
	if button, exists := state.Buttons[m.config.Mappings.Buttons.Emergency.Button]; exists {
		if m.shouldTriggerButtonAction(button, m.config.Mappings.Buttons.Emergency, now) {
			actions = append(actions, ActionEmergency)
		}
	}

	// Check Flip Forward button
	if button, exists := state.Buttons[m.config.Mappings.Buttons.FlipForward.Button]; exists {
		if m.shouldTriggerButtonAction(button, m.config.Mappings.Buttons.FlipForward, now) {
			actions = append(actions, ActionFlipForward)
		}
	}

	// Check Flip Backward button
	if button, exists := state.Buttons[m.config.Mappings.Buttons.FlipBackward.Button]; exists {
		if m.shouldTriggerButtonAction(button, m.config.Mappings.Buttons.FlipBackward, now) {
			actions = append(actions, ActionFlipBackward)
		}
	}

	// Check optional Flip Left button
	if m.config.Mappings.Buttons.FlipLeft != nil {
		if button, exists := state.Buttons[m.config.Mappings.Buttons.FlipLeft.Button]; exists {
			if m.shouldTriggerButtonAction(button, *m.config.Mappings.Buttons.FlipLeft, now) {
				actions = append(actions, ActionFlipLeft)
			}
		}
	}

	// Check optional Flip Right button
	if m.config.Mappings.Buttons.FlipRight != nil {
		if button, exists := state.Buttons[m.config.Mappings.Buttons.FlipRight.Button]; exists {
			if m.shouldTriggerButtonAction(button, *m.config.Mappings.Buttons.FlipRight, now) {
				actions = append(actions, ActionFlipRight)
			}
		}
	}

	// Check optional Stream Toggle button
	if m.config.Mappings.Buttons.StreamToggle != nil {
		if button, exists := state.Buttons[m.config.Mappings.Buttons.StreamToggle.Button]; exists {
			if m.shouldTriggerButtonAction(button, *m.config.Mappings.Buttons.StreamToggle, now) {
				// This would need state tracking for proper toggle behavior
				// For now, just trigger stream on
				actions = append(actions, ActionStreamOn)
			}
		}
	}

	return actions
}

// shouldTriggerButtonAction determines if a button action should be triggered
func (m *DefaultMapper) shouldTriggerButtonAction(button *ButtonState, mapping ButtonMapping, _ time.Time) bool {
	if !button.Pressed {
		return false
	}

	// Check hold time requirement
	if mapping.HoldTime > 0 {
		if time.Since(button.PressTime) < time.Duration(mapping.HoldTime)*time.Millisecond {
			return false
		}
	}

	// Check double tap requirement
	if mapping.DoubleTap {
		return button.TapCount >= 2 && time.Since(button.LastTapTime) < 500*time.Millisecond
	}

	return true
}

// mapButtonToAction maps a button type to a drone action (simplified version)
func (m *DefaultMapper) mapButtonToAction(button ButtonType) DroneAction {
	switch button {
	case m.config.Mappings.Buttons.TakeoffLand.Button:
		return ActionTakeoff // Simplified - real logic would need state tracking
	case m.config.Mappings.Buttons.Emergency.Button:
		return ActionEmergency
	case m.config.Mappings.Buttons.FlipForward.Button:
		return ActionFlipForward
	case m.config.Mappings.Buttons.FlipBackward.Button:
		return ActionFlipBackward
	default:
		return ""
	}
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
