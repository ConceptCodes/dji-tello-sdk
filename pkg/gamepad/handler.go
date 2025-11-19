package gamepad

import (
	"fmt"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/veandco/go-sdl2/sdl"
)

// Handler manages gamepad input and converts it to drone commands
type Handler struct {
	config    *Config
	state     *GamepadState
	mapper    Mapper
	isRunning bool
	mu        sync.RWMutex

	// SDL2 specific
	gamepad  *sdl.GameController
	joystick *sdl.Joystick

	// Callbacks
	onCommand func(Command)
	onError   func(error)

	// Timing
	updateInterval time.Duration
	lastUpdate     time.Time
}

// HandlerOptions contains options for creating a new Handler
type HandlerOptions struct {
	Config    *Config
	Mapper    Mapper
	OnCommand func(Command)
	OnError   func(error)
}

// NewHandler creates a new gamepad handler
func NewHandler(opts HandlerOptions) (*Handler, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create mapper if not provided
	mapper := opts.Mapper
	if mapper == nil {
		mapper = NewDefaultMapper(opts.Config)
	}

	// Initialize SDL2
	// Note: SDL2 functions should generally be called from the main thread.
	if err := sdl.Init(sdl.INIT_GAMECONTROLLER); err != nil {
		return nil, fmt.Errorf("failed to initialize SDL2: %w", err)
	}

	// Load game controller mappings from database
	if err := loadControllerMappings(); err != nil {
		utils.Logger.Warnf("Failed to load controller mappings: %v", err)
	}

	// Calculate update interval
	updateInterval := time.Second / time.Duration(opts.Config.Controller.UpdateRate)

	handler := &Handler{
		config:         opts.Config,
		state:          NewGamepadState(),
		mapper:         mapper,
		updateInterval: updateInterval,
		onCommand:      opts.OnCommand,
		onError:        opts.OnError,
	}

	utils.Logger.Info("Gamepad handler created with SDL2 support")
	return handler, nil
}

// Start initializes the gamepad connection.
// Note: This does NOT start a background loop. You must call ProcessEvents() periodically from the main thread.
func (h *Handler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.isRunning {
		return fmt.Errorf("handler is already running")
	}

	// Try to open the first available gamepad
	if err := h.openGamepad(); err != nil {
		return fmt.Errorf("failed to open gamepad: %w", err)
	}

	h.isRunning = true

	utils.Logger.Infof("Starting gamepad handler with %d Hz update rate", h.config.Controller.UpdateRate)
	utils.Logger.Infof("Connected to gamepad: %s", h.gamepad.Name())

	return nil
}

// Stop stops the gamepad handler and cleans up resources
func (h *Handler) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isRunning {
		return fmt.Errorf("handler is not running")
	}

	h.isRunning = false

	// Close gamepad and cleanup SDL2
	if h.gamepad != nil {
		h.gamepad.Close()
		h.gamepad = nil
	}
	if h.joystick != nil {
		h.joystick.Close()
		h.joystick = nil
	}

	utils.Logger.Info("Gamepad handler stopped")
	return nil
}

// IsRunning returns whether the handler is currently running
func (h *Handler) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isRunning
}

// GetState returns the current gamepad state
func (h *Handler) GetState() *GamepadState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	stateCopy := NewGamepadState()
	for btn, state := range h.state.Buttons {
		stateCopy.Buttons[btn] = &ButtonState{
			Pressed:     state.Pressed,
			PressTime:   state.PressTime,
			LastRelease: state.LastRelease,
			TapCount:    state.TapCount,
			LastTapTime: state.LastTapTime,
		}
	}
	for axis, state := range h.state.Axes {
		stateCopy.Axes[axis] = &AxisState{
			Value:     state.Value,
			LastValue: state.LastValue,
		}
	}
	stateCopy.LastUpdate = h.state.LastUpdate

	return stateCopy
}

// ProcessEvents processes gamepad input events.
// This MUST be called from the main thread.
func (h *Handler) ProcessEvents() {
	if h.gamepad == nil {
		return
	}

	// Update SDL2 events
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.ControllerAxisEvent:
			h.handleAxisEvent(e)
		case *sdl.ControllerButtonEvent:
			h.handleButtonEvent(e)
		case *sdl.ControllerDeviceEvent:
			h.handleDeviceEvent(e)
		}
	}

	// Process continuous axes (like joysticks)
	h.processAxes()

	// Generate commands from current state
	commands, err := h.mapper.MapState(h.state)
	if err != nil {
		h.onError(fmt.Errorf("failed to map state to commands: %w", err))
	} else {
		for _, cmd := range commands {
			h.triggerCommand(cmd)
		}
	}

	h.state.LastUpdate = time.Now()
	h.lastUpdate = time.Now()
}

// triggerCommand calls the command callback
func (h *Handler) triggerCommand(command Command) {
	if h.onCommand != nil {
		utils.Logger.Debugf("Triggering command: %v", command)
		h.onCommand(command)
	}
}

// ListGamepads returns a list of available gamepads
func ListGamepads() []string {
	var gamepads []string

	numJoysticks := sdl.NumJoysticks()
	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			name := sdl.GameControllerNameForIndex(i)
			gamepads = append(gamepads, fmt.Sprintf("%d: %s", i, name))
		}
	}

	return gamepads
}

// GetGamepadInfo returns information about available gamepads
func GetGamepadInfo() []map[string]string {
	var gamepads []map[string]string

	numJoysticks := sdl.NumJoysticks()
	for i := 0; i < numJoysticks; i++ {
		info := map[string]string{
			"id":   fmt.Sprintf("%d", i),
			"name": sdl.JoystickNameForIndex(i),
		}

		if sdl.IsGameController(i) {
			info["type"] = "GameController"
			info["mapping"] = sdl.GameControllerMappingForDeviceIndex(i)
		} else {
			info["type"] = "Joystick"
		}

		gamepads = append(gamepads, info)
	}

	return gamepads
}

// openGamepad opens the first available gamepad
func (h *Handler) openGamepad() error {
	numJoysticks := sdl.NumJoysticks()
	if numJoysticks == 0 {
		return fmt.Errorf("no gamepads connected")
	}

	// Try to open the first available game controller
	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			gamepad := sdl.GameControllerOpen(i)
			if gamepad == nil {
				utils.Logger.Warnf("Failed to open gamepad %d", i)
				continue
			}

			h.gamepad = gamepad
			h.joystick = gamepad.Joystick()

			utils.Logger.Infof("Connected to gamepad: %s", gamepad.Name())
			return nil
		}
	}

	return fmt.Errorf("no compatible gamepads found")
}

// loadControllerMappings loads game controller database mappings
func loadControllerMappings() error {
	// SDL2 automatically loads system game controller database
	// Additional mappings can be loaded here if needed
	return nil
}

// handleAxisEvent processes controller axis movement
func (h *Handler) handleAxisEvent(event *sdl.ControllerAxisEvent) {
	axisName := h.getAxisName(sdl.GameControllerAxis(event.Axis))
	if axisName == "" {
		return
	}

	// Normalize axis value to [-1.0, 1.0]
	value := float32(event.Value) / 32767.0
	if value < -1.0 {
		value = -1.0
	} else if value > 1.0 {
		value = 1.0
	}

	h.updateAxis(axisName, value)

	// Generate event for mapper
	gamepadEvent := Event{
		Type:      EventAxisChange,
		Input:     axisName,
		Value:     float64(value),
		Timestamp: time.Now(),
	}

	commands, err := h.mapper.MapEvent(gamepadEvent)
	if err != nil {
		h.onError(fmt.Errorf("failed to map axis event: %w", err))
	} else {
		for _, cmd := range commands {
			h.triggerCommand(cmd)
		}
	}
}

// handleButtonEvent processes controller button presses
func (h *Handler) handleButtonEvent(event *sdl.ControllerButtonEvent) {
	buttonName := h.getButtonName(sdl.GameControllerButton(event.Button))
	if buttonName == "" {
		return
	}

	pressed := event.State == sdl.PRESSED
	h.updateButton(buttonName, pressed)

	// Generate event for mapper
	var eventType EventType
	var value float64
	if pressed {
		eventType = EventButtonPress
		value = 1.0
	} else {
		eventType = EventButtonRelease
		value = 0.0
	}

	gamepadEvent := Event{
		Type:      eventType,
		Input:     buttonName,
		Value:     value,
		Timestamp: time.Now(),
	}

	commands, err := h.mapper.MapEvent(gamepadEvent)
	if err != nil {
		h.onError(fmt.Errorf("failed to map button event: %w", err))
	} else {
		for _, cmd := range commands {
			h.triggerCommand(cmd)
		}
	}
}

// handleDeviceEvent processes controller connection/disconnection
func (h *Handler) handleDeviceEvent(event *sdl.ControllerDeviceEvent) {
	switch event.Type {
	case sdl.CONTROLLERDEVICEADDED:
		utils.Logger.Infof("Gamepad connected: %d", event.Which)
		if h.gamepad == nil {
			h.openGamepad()
		}
	case sdl.CONTROLLERDEVICEREMOVED:
		utils.Logger.Infof("Gamepad disconnected: %d", event.Which)
		if h.gamepad != nil && h.gamepad.Joystick().InstanceID() == event.Which {
			h.gamepad.Close()
			h.gamepad = nil
			h.joystick = nil
		}
	}
}

// processAxes processes continuous axis values
func (h *Handler) processAxes() {
	if h.gamepad == nil {
		return
	}

	// Process all known axes
	axes := []sdl.GameControllerAxis{
		sdl.CONTROLLER_AXIS_LEFTX,
		sdl.CONTROLLER_AXIS_LEFTY,
		sdl.CONTROLLER_AXIS_RIGHTX,
		sdl.CONTROLLER_AXIS_RIGHTY,
		sdl.CONTROLLER_AXIS_TRIGGERLEFT,
		sdl.CONTROLLER_AXIS_TRIGGERRIGHT,
	}

	for _, axis := range axes {
		axisName := h.getAxisName(axis)
		if axisName == "" {
			continue
		}

		value := h.gamepad.Axis(axis)
		normalizedValue := float32(value) / 32767.0
		if normalizedValue < -1.0 {
			normalizedValue = -1.0
		} else if normalizedValue > 1.0 {
			normalizedValue = 1.0
		}

		h.updateAxis(axisName, normalizedValue)
	}
}

// getAxisName converts SDL axis enum to configuration axis name
func (h *Handler) getAxisName(axis sdl.GameControllerAxis) string {
	switch axis {
	case sdl.CONTROLLER_AXIS_LEFTX:
		return "left_x"
	case sdl.CONTROLLER_AXIS_LEFTY:
		return "left_y"
	case sdl.CONTROLLER_AXIS_RIGHTX:
		return "right_x"
	case sdl.CONTROLLER_AXIS_RIGHTY:
		return "right_y"
	case sdl.CONTROLLER_AXIS_TRIGGERLEFT:
		return "left_trigger"
	case sdl.CONTROLLER_AXIS_TRIGGERRIGHT:
		return "right_trigger"
	default:
		return ""
	}
}

// getButtonName converts SDL button enum to configuration button name
func (h *Handler) getButtonName(button sdl.GameControllerButton) string {
	switch button {
	case sdl.CONTROLLER_BUTTON_A:
		return "a"
	case sdl.CONTROLLER_BUTTON_B:
		return "b"
	case sdl.CONTROLLER_BUTTON_X:
		return "x"
	case sdl.CONTROLLER_BUTTON_Y:
		return "y"
	case sdl.CONTROLLER_BUTTON_BACK:
		return "back"
	case sdl.CONTROLLER_BUTTON_GUIDE:
		return "guide"
	case sdl.CONTROLLER_BUTTON_START:
		return "start"
	case sdl.CONTROLLER_BUTTON_LEFTSTICK:
		return "left_stick"
	case sdl.CONTROLLER_BUTTON_RIGHTSTICK:
		return "right_stick"
	case sdl.CONTROLLER_BUTTON_LEFTSHOULDER:
		return "left_shoulder"
	case sdl.CONTROLLER_BUTTON_RIGHTSHOULDER:
		return "right_shoulder"
	case sdl.CONTROLLER_BUTTON_DPAD_UP:
		return "dpad_up"
	case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
		return "dpad_down"
	case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
		return "dpad_left"
	case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
		return "dpad_right"
	default:
		return ""
	}
}

// updateAxis updates axis state in the gamepad state
func (h *Handler) updateAxis(name string, value float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	axisType := AxisType(name)
	if _, exists := h.state.Axes[axisType]; !exists {
		h.state.Axes[axisType] = &AxisState{}
	}

	h.state.Axes[axisType].LastValue = h.state.Axes[axisType].Value
	h.state.Axes[axisType].Value = float64(value)
}

// updateButton updates button state in the gamepad state
func (h *Handler) updateButton(name string, pressed bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	buttonType := ButtonType(name)

	if _, exists := h.state.Buttons[buttonType]; !exists {
		h.state.Buttons[buttonType] = &ButtonState{}
	}

	button := h.state.Buttons[buttonType]

	if pressed && !button.Pressed {
		// Button just pressed
		button.Pressed = true
		button.PressTime = now
		button.TapCount++
		button.LastTapTime = now
	} else if !pressed && button.Pressed {
		// Button just released
		button.Pressed = false
		button.LastRelease = now
	}
}
