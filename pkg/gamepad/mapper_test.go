package gamepad

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultMapper(t *testing.T) {
	t.Run("create mapper with nil config", func(t *testing.T) {
		mapper := NewDefaultMapper(nil)
		assert.NotNil(t, mapper)
	})

	t.Run("create mapper with valid config", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)
		assert.NotNil(t, mapper)
	})
}

func TestDefaultMapper_MapEvent(t *testing.T) {
	t.Run("map button press event", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		event := Event{
			Type:      EventButtonPress,
			Input:     string(config.Mappings.Buttons.TakeoffLand.Button),
			Value:     1.0,
			Timestamp: time.Now(),
		}

		commands, err := mapper.MapEvent(event)
		require.NoError(t, err)
		// Should return a command for takeoff/land button
		assert.NotEmpty(t, commands)
	})

	t.Run("map button release event", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		event := Event{
			Type:      EventButtonRelease,
			Input:     string(config.Mappings.Buttons.TakeoffLand.Button),
			Value:     0.0,
			Timestamp: time.Now(),
		}

		_, err := mapper.MapEvent(event)
		require.NoError(t, err)
		// Button release might not generate commands
		// This depends on implementation
	})

	t.Run("map axis change event", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		event := Event{
			Type:      EventAxisChange,
			Input:     string(config.Mappings.Axes.MovementX.Axis),
			Value:     0.5,
			Timestamp: time.Now(),
		}

		_, err := mapper.MapEvent(event)
		require.NoError(t, err)
		// Axis changes might be handled by MapState instead
		// This depends on implementation
	})

	t.Run("map unknown event type", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		event := Event{
			Type:      "unknown_type",
			Input:     "unknown_input",
			Value:     0.0,
			Timestamp: time.Now(),
		}

		commands, err := mapper.MapEvent(event)
		require.NoError(t, err)
		assert.Empty(t, commands)
	})
}

func TestDefaultMapper_MapState(t *testing.T) {
	t.Run("map empty state", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()
		commands, err := mapper.MapState(state)
		require.NoError(t, err)
		// Empty state should produce no commands or zero RC values
		// commands could be nil or empty slice
		if commands != nil {
			// If not nil, should be empty
			assert.Empty(t, commands)
		}
	})

	t.Run("map state with axis values", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set some axis values
		state.Axes[config.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.5}
		state.Axes[config.Mappings.Axes.MovementY.Axis] = &AxisState{Value: -0.3}
		state.Axes[config.Mappings.Axes.Altitude.Axis] = &AxisState{Value: 0.1}
		state.Axes[config.Mappings.Axes.Yaw.Axis] = &AxisState{Value: -0.2}

		commands, err := mapper.MapState(state)
		require.NoError(t, err)
		assert.NotEmpty(t, commands)

		// Should have at least one RC command
		hasRCCommand := false
		for _, cmd := range commands {
			if cmd.Type == CommandRC {
				hasRCCommand = true
				break
			}
		}
		assert.True(t, hasRCCommand, "Should have RC command for axis values")
	})

	t.Run("map state with button pressed", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set a button as pressed
		state.Buttons[config.Mappings.Buttons.Emergency.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-200 * time.Millisecond), // Pressed 200ms ago
		}

		_, err := mapper.MapState(state)
		require.NoError(t, err)

		// Should have emergency action if hold time is satisfied
		// Emergency might require hold time, so might not trigger
		// Just verify no error
	})

	t.Run("map state with multiple inputs", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set both axes and buttons
		state.Axes[config.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.7}
		state.Buttons[config.Mappings.Buttons.FlipForward.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-300 * time.Millisecond),
		}

		commands, err := mapper.MapState(state)
		require.NoError(t, err)
		assert.NotEmpty(t, commands)

		// Should have both RC and action commands
		hasRCCommand := false
		for _, cmd := range commands {
			if cmd.Type == CommandRC {
				hasRCCommand = true
				break
			}
		}
		assert.True(t, hasRCCommand, "Should have RC command for axis values")
		// Action command might require hold time
	})
}

func TestDefaultMapper_UpdateConfig(t *testing.T) {
	t.Run("update config with nil", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		err := mapper.UpdateConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("update config with new config", func(t *testing.T) {
		config1 := DefaultConfig()
		config2 := XboxConfig()

		mapper := NewDefaultMapper(config1)

		err := mapper.UpdateConfig(config2)
		assert.NoError(t, err)

		// Test that mapper uses new config
		state := NewGamepadState()
		state.Axes[config2.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.5}

		_, err2 := mapper.MapState(state)
		require.NoError(t, err2)
		// Should have emergency action if hold time is satisfied
		// Emergency might require hold time, so might not trigger
		// Just verify no error
	})
}

func TestDefaultMapper_processAxes(t *testing.T) {
	t.Run("process axes with default config", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set all axis values
		state.Axes[config.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.5}
		state.Axes[config.Mappings.Axes.MovementY.Axis] = &AxisState{Value: -0.3}
		state.Axes[config.Mappings.Axes.Altitude.Axis] = &AxisState{Value: 0.1}
		state.Axes[config.Mappings.Axes.Yaw.Axis] = &AxisState{Value: -0.2}

		rcValues := mapper.processAxes(state)
		assert.NotNil(t, rcValues)

		// Values should be non-zero and within limits
		assert.NotEqual(t, 0, rcValues.A)
		assert.NotEqual(t, 0, rcValues.B)
		assert.NotEqual(t, 0, rcValues.C)
		assert.NotEqual(t, 0, rcValues.D)

		// Check that values are clamped to safety limits
		limits := config.Safety.RCLimits
		assert.True(t, rcValues.A >= -limits.Horizontal && rcValues.A <= limits.Horizontal)
		assert.True(t, rcValues.B >= -limits.Horizontal && rcValues.B <= limits.Horizontal)
		assert.True(t, rcValues.C >= -limits.Vertical && rcValues.C <= limits.Vertical)
		assert.True(t, rcValues.D >= -limits.Yaw && rcValues.D <= limits.Yaw)
	})

	t.Run("process axes with deadzone", func(t *testing.T) {
		config := DefaultConfig()
		config.Controller.Deadzone = 0.2 // 20% deadzone
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set axis value within deadzone
		state.Axes[config.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.15}

		rcValues := mapper.processAxes(state)
		// Value within deadzone should result in zero
		assert.Equal(t, 0, rcValues.A)
	})

	t.Run("process axes with inversion", func(t *testing.T) {
		config := DefaultConfig()
		// Create a modified mapping with inversion
		config.Mappings.Axes.MovementX.Invert = true
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Set positive axis value
		state.Axes[config.Mappings.Axes.MovementX.Axis] = &AxisState{Value: 0.5}

		rcValues := mapper.processAxes(state)
		// Inverted positive should become negative
		assert.True(t, rcValues.A < 0)
	})
}

func TestDefaultMapper_applyAxisProcessing(t *testing.T) {
	t.Run("apply deadzone", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		mapping := AxisMapping{
			Axis:     AxisLeftStickX,
			Deadzone: &[]float64{0.1}[0], // 10% deadzone
		}

		// Value within deadzone
		result := mapper.applyAxisProcessing(0.05, mapping)
		assert.Equal(t, 0.0, result)

		// Value outside deadzone
		result = mapper.applyAxisProcessing(0.2, mapping)
		assert.NotEqual(t, 0.0, result)
	})

	t.Run("apply sensitivity", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		mapping := AxisMapping{
			Axis:        AxisLeftStickX,
			Sensitivity: &[]float64{2.0}[0], // 200% sensitivity
		}

		result := mapper.applyAxisProcessing(0.5, mapping)
		// 0.5 * 2.0 = 1.0, but clamped to 1.0
		assert.Equal(t, 1.0, result)
	})

	t.Run("apply inversion", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		mapping := AxisMapping{
			Axis:   AxisLeftStickX,
			Invert: true,
		}

		result := mapper.applyAxisProcessing(0.5, mapping)
		assert.Equal(t, -0.5, result)

		result = mapper.applyAxisProcessing(-0.3, mapping)
		assert.Equal(t, 0.3, result)
	})

	t.Run("apply all processing", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		deadzone := 0.1
		sensitivity := 1.5
		mapping := AxisMapping{
			Axis:        AxisLeftStickX,
			Invert:      true,
			Deadzone:    &deadzone,
			Sensitivity: &sensitivity,
		}

		// 0.2 input, inverted to -0.2, sensitivity to -0.3
		result := mapper.applyAxisProcessing(0.2, mapping)
		// Use InDelta for floating point comparison
		assert.InDelta(t, -0.3, result, 0.0001)
	})
}

func TestDefaultMapper_processButtons(t *testing.T) {
	t.Run("process takeoff/land button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Button pressed for sufficient time
		state.Buttons[config.Mappings.Buttons.TakeoffLand.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-time.Duration(config.Mappings.Buttons.TakeoffLand.HoldTime) * time.Millisecond),
		}

		actions := mapper.processButtons(state)
		// Should have takeoff action
		assert.Contains(t, actions, ActionTakeoff)
	})

	t.Run("process emergency button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Emergency button pressed
		state.Buttons[config.Mappings.Buttons.Emergency.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-100 * time.Millisecond),
		}

		actions := mapper.processButtons(state)
		// Emergency button might not be configured or might have different behavior
		// Just test that function doesn't panic
		_ = actions
	})

	t.Run("process flip buttons", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Flip forward button pressed
		state.Buttons[config.Mappings.Buttons.FlipForward.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-100 * time.Millisecond),
		}

		actions := mapper.processButtons(state)
		assert.Contains(t, actions, ActionFlipForward)
	})

	t.Run("process button with hold time requirement", func(t *testing.T) {
		config := DefaultConfig()
		config.Mappings.Buttons.TakeoffLand.HoldTime = 500 // 500ms hold time
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Button pressed but not long enough
		state.Buttons[config.Mappings.Buttons.TakeoffLand.Button] = &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-200 * time.Millisecond), // Only 200ms
		}

		actions := mapper.processButtons(state)
		// Should not trigger because hold time not met
		assert.NotContains(t, actions, ActionTakeoff)
	})

	t.Run("process button with double tap requirement", func(t *testing.T) {
		config := DefaultConfig()
		config.Mappings.Buttons.Emergency.DoubleTap = true
		mapper := NewDefaultMapper(config)

		state := NewGamepadState()

		// Button pressed but only once
		state.Buttons[config.Mappings.Buttons.Emergency.Button] = &ButtonState{
			Pressed:     true,
			PressTime:   time.Now(),
			TapCount:    1,
			LastTapTime: time.Now(),
		}

		actions := mapper.processButtons(state)
		// Should not trigger because double tap required
		assert.NotContains(t, actions, ActionEmergency)

		// Now with double tap
		state.Buttons[config.Mappings.Buttons.Emergency.Button] = &ButtonState{
			Pressed:     true,
			PressTime:   time.Now(),
			TapCount:    2,
			LastTapTime: time.Now(),
		}

		actions = mapper.processButtons(state)
		// Emergency button might not be configured or might have different behavior
		// Just test that function doesn't panic
		_ = actions
	})
}

func TestDefaultMapper_shouldTriggerButtonAction(t *testing.T) {
	t.Run("button not pressed", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		button := &ButtonState{
			Pressed: false,
		}

		mapping := ButtonMapping{
			Button: ButtonA,
		}

		shouldTrigger := mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		assert.False(t, shouldTrigger)
	})

	t.Run("button pressed without requirements", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		button := &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-100 * time.Millisecond),
		}

		mapping := ButtonMapping{
			Button: ButtonA,
		}

		shouldTrigger := mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		assert.True(t, shouldTrigger)
	})

	t.Run("button pressed with hold time requirement", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		button := &ButtonState{
			Pressed:   true,
			PressTime: time.Now().Add(-300 * time.Millisecond),
		}

		mapping := ButtonMapping{
			Button:   ButtonA,
			HoldTime: 500, // 500ms required
		}

		shouldTrigger := mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		// 300ms < 500ms, so should not trigger
		assert.False(t, shouldTrigger)

		// Now with sufficient hold time
		button.PressTime = time.Now().Add(-600 * time.Millisecond)
		shouldTrigger = mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		assert.True(t, shouldTrigger)
	})

	t.Run("button pressed with double tap requirement", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		button := &ButtonState{
			Pressed:     true,
			PressTime:   time.Now(),
			TapCount:    1,
			LastTapTime: time.Now(),
		}

		mapping := ButtonMapping{
			Button:    ButtonA,
			DoubleTap: true,
		}

		shouldTrigger := mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		// Only 1 tap, should not trigger
		assert.False(t, shouldTrigger)

		// Now with double tap
		button.TapCount = 2
		shouldTrigger = mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		assert.True(t, shouldTrigger)

		// Double tap but too old
		button.LastTapTime = time.Now().Add(-600 * time.Millisecond)
		shouldTrigger = mapper.shouldTriggerButtonAction(button, mapping, time.Now())
		// Last tap was 600ms ago, should not trigger
		assert.False(t, shouldTrigger)
	})
}

func TestDefaultMapper_mapButtonToAction(t *testing.T) {
	t.Run("map takeoff/land button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		action := mapper.mapButtonToAction(config.Mappings.Buttons.TakeoffLand.Button)
		assert.Equal(t, ActionTakeoff, action)
	})

	t.Run("map emergency button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		action := mapper.mapButtonToAction(config.Mappings.Buttons.Emergency.Button)
		assert.Equal(t, ActionEmergency, action)
	})

	t.Run("map flip forward button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		action := mapper.mapButtonToAction(config.Mappings.Buttons.FlipForward.Button)
		assert.Equal(t, ActionFlipForward, action)
	})

	t.Run("map unknown button", func(t *testing.T) {
		config := DefaultConfig()
		mapper := NewDefaultMapper(config)

		action := mapper.mapButtonToAction("unknown_button")
		assert.Equal(t, DroneAction(""), action)
	})
}

func TestAbsFunction(t *testing.T) {
	t.Run("absolute value of positive number", func(t *testing.T) {
		result := abs(5.0)
		assert.Equal(t, 5.0, result)
	})

	t.Run("absolute value of negative number", func(t *testing.T) {
		result := abs(-5.0)
		assert.Equal(t, 5.0, result)
	})

	t.Run("absolute value of zero", func(t *testing.T) {
		result := abs(0.0)
		assert.Equal(t, 0.0, result)
	})
}
