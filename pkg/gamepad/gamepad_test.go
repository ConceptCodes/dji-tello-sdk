package gamepad

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0", config.Version)
	assert.Equal(t, 60, config.Controller.UpdateRate)
	assert.True(t, config.Controller.AutoDetect)
	assert.NotNil(t, config.Mappings)
}

func TestXboxConfig(t *testing.T) {
	config := XboxConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0-xbox", config.Version)
	assert.Equal(t, 60, config.Controller.UpdateRate)
}

func TestPlayStationConfig(t *testing.T) {
	config := PlayStationConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "1.0.0-ps", config.Version)
	assert.Equal(t, 60, config.Controller.UpdateRate)
}

func TestGamepadState(t *testing.T) {
	state := NewGamepadState()

	assert.NotNil(t, state)
	assert.NotNil(t, state.Axes)
	assert.NotNil(t, state.Buttons)
	assert.False(t, state.LastUpdate.IsZero())
}

func TestAxisState(t *testing.T) {
	axis := &AxisState{}

	assert.Equal(t, 0.0, axis.Value)
	assert.Equal(t, 0.0, axis.LastValue)
}

func TestButtonState(t *testing.T) {
	button := &ButtonState{}

	assert.False(t, button.Pressed)
	assert.True(t, button.PressTime.IsZero())
	assert.True(t, button.LastRelease.IsZero())
	assert.Equal(t, 0, button.TapCount)
	assert.True(t, button.LastTapTime.IsZero())
}

func TestButtonStatePress(t *testing.T) {
	button := &ButtonState{}
	now := time.Now()

	// Test press
	button.Pressed = true
	button.PressTime = now
	button.TapCount = 1
	button.LastTapTime = now

	assert.True(t, button.Pressed)
	assert.Equal(t, now, button.PressTime)
	assert.Equal(t, 1, button.TapCount)
	assert.Equal(t, now, button.LastTapTime)

	// Test release
	later := now.Add(100 * time.Millisecond)
	button.Pressed = false
	button.LastRelease = later

	assert.False(t, button.Pressed)
	assert.Equal(t, later, button.LastRelease)
}

// func TestConfigValidation(t *testing.T) {
// 	// Test valid config
// 	config := DefaultConfig()
//
// 	loader, err := NewConfigLoader()
// 	require.NoError(t, err)
//
// 	// Skip validation test if schema file is not found
// 	if err := loader.ValidateConfig(config); err != nil {
// 		t.Skipf("Skipping validation test: %v", err)
// 	}
// }

func TestPresetConfigs(t *testing.T) {
	presets := GetPresetConfigs()

	assert.Contains(t, presets, "default")
	assert.Contains(t, presets, "xbox")
	assert.Contains(t, presets, "playstation")

	names := GetConfigNames()
	assert.Contains(t, names, "default")
	assert.Contains(t, names, "xbox")
	assert.Contains(t, names, "playstation")
}

func TestRCValues(t *testing.T) {
	rc := RCValues{
		A: 10,
		B: -20,
		C: 30,
		D: -40,
	}

	assert.Equal(t, 10, rc.A)
	assert.Equal(t, -20, rc.B)
	assert.Equal(t, 30, rc.C)
	assert.Equal(t, -40, rc.D)
}

func TestDroneAction(t *testing.T) {
	action := ActionTakeoff
	assert.Equal(t, "takeoff", string(action))

	action = ActionLand
	assert.Equal(t, "land", string(action))
}

func TestAxisAndButtonTypes(t *testing.T) {
	// Test axis types
	assert.Equal(t, "left_stick_x", string(AxisLeftStickX))
	assert.Equal(t, "left_stick_y", string(AxisLeftStickY))
	assert.Equal(t, "right_stick_x", string(AxisRightStickX))
	assert.Equal(t, "right_stick_y", string(AxisRightStickY))

	// Test button types
	assert.Equal(t, "button_a", string(ButtonA))
	assert.Equal(t, "button_b", string(ButtonB))
	assert.Equal(t, "button_x", string(ButtonX))
	assert.Equal(t, "button_y", string(ButtonY))
}
