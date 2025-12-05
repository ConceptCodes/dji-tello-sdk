// Example: How web interface can use the refactored pkg/gamepad
// This file demonstrates the intended usage pattern for web handlers

package examples

import (
	"log"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

// WebGamepadHandler demonstrates how web interfaces can use pkg/gamepad
type WebGamepadHandler struct {
	drone  tello.TelloCommander
	mapper gamepad.Mapper
}

// NewWebGamepadHandler creates a new web gamepad handler
func NewWebGamepadHandler(drone tello.TelloCommander) *WebGamepadHandler {
	// Load gamepad configuration (could be from user preferences, database, etc.)
	config, err := gamepad.LoadDefaultConfigFromFile()
	if err != nil {
		// Fallback to default config
		config = gamepad.DefaultConfig()
	}

	// Create mapper for converting events to commands
	mapper := gamepad.NewDefaultMapper(config)

	return &WebGamepadHandler{
		drone:  drone,
		mapper: mapper,
	}
}

// HandleWebInput processes web input (e.g., from WebSocket, keyboard, or virtual gamepad)
// and converts it to drone commands using the gamepad mapper
func (w *WebGamepadHandler) HandleWebInput(webInput WebInputEvent) error {
	// Convert web input to normalized gamepad event
	gamepadEvent := w.convertWebToGamepadEvent(webInput)

	// Map the gamepad event to drone commands
	commands, err := w.mapper.MapEvent(gamepadEvent)
	if err != nil {
		return err
	}

	// Execute commands
	for _, cmd := range commands {
		if err := w.executeCommand(cmd); err != nil {
			log.Printf("Failed to execute command: %v", err)
		}
	}

	return nil
}

// WebInputEvent represents input from web interface
type WebInputEvent struct {
	Type  string  // "keydown", "keyup", "joystick_move"
	Key   string  // "w", "a", "s", "d", "space", etc.
	Axis  string  // "left_x", "left_y", "right_x", "right_y"
	Value float64 // -1.0 to 1.0 for axes, 0.0/1.0 for buttons
}

// convertWebToGamepadEvent converts web input to gamepad events
func (w *WebGamepadHandler) convertWebToGamepadEvent(webInput WebInputEvent) gamepad.Event {
	var eventType gamepad.EventType
	var input string
	var value float64

	switch webInput.Type {
	case "keydown":
		eventType = gamepad.EventButtonPress
		value = 1.0
	case "keyup":
		eventType = gamepad.EventButtonRelease
		value = 0.0
	case "joystick_move":
		eventType = gamepad.EventAxisChange
		value = webInput.Value
	}

	// Map web controls to gamepad controls
	switch webInput.Key {
	case "w":
		input = string(gamepad.AxisLeftStickY)
		value = -1.0 // Forward is negative Y
	case "s":
		input = string(gamepad.AxisLeftStickY)
		value = 1.0 // Backward is positive Y
	case "a":
		input = string(gamepad.AxisLeftStickX)
		value = -1.0 // Left is negative X
	case "d":
		input = string(gamepad.AxisLeftStickX)
		value = 1.0 // Right is positive X
	case "space":
		input = string(gamepad.ButtonA) // Takeoff/Land
	case "e":
		input = string(gamepad.ButtonB) // Emergency
	case "q":
		input = string(gamepad.ButtonX) // Flip Forward
	case "r":
		input = string(gamepad.ButtonY) // Flip Backward
	default:
		if webInput.Axis != "" {
			input = webInput.Axis
		}
	}

	return gamepad.Event{
		Type:      eventType,
		Input:     input,
		Value:     value,
		Timestamp: time.Now(),
	}
}

// executeCommand executes a drone command
func (w *WebGamepadHandler) executeCommand(cmd gamepad.Command) error {
	switch cmd.Type {
	case gamepad.CommandRC:
		if rcValues, ok := cmd.Data.(gamepad.RCValues); ok {
			return w.drone.SetRcControl(rcValues.A, rcValues.B, rcValues.C, rcValues.D)
		}
	case gamepad.CommandAction:
		if action, ok := cmd.Data.(gamepad.DroneAction); ok {
			switch action {
			case gamepad.ActionTakeoff:
				return w.drone.TakeOff()
			case gamepad.ActionLand:
				return w.drone.Land()
			case gamepad.ActionEmergency:
				return w.drone.Emergency()
			case gamepad.ActionFlipForward:
				return w.drone.Flip(tello.FlipForward)
			case gamepad.ActionFlipBackward:
				return w.drone.Flip(tello.FlipBackward)
			case gamepad.ActionFlipLeft:
				return w.drone.Flip(tello.FlipLeft)
			case gamepad.ActionFlipRight:
				return w.drone.Flip(tello.FlipRight)
			case gamepad.ActionStreamOn:
				return w.drone.StreamOn()
			}
		}
	}
	return nil
}

// Example WebSocket handler usage:
/*
func handleWebSocket(ws *websocket.Conn, drone tello.TelloCommander) {
	handler := NewWebGamepadHandler(drone)

	for {
		var webInput WebInputEvent
		if err := ws.ReadJSON(&webInput); err != nil {
			break
		}

		if err := handler.HandleWebInput(webInput); err != nil {
			log.Printf("Error handling web input: %v", err)
		}
	}
}
*/
