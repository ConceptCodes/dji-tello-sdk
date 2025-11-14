# Gamepad Control Example

This example demonstrates how to use a gamepad to control the DJI Tello drone.

## Overview

The gamepad example shows:
- How to detect and list available gamepads
- Loading gamepad configurations
- Converting gamepad inputs to drone RC values
- Handling gamepad events and errors
- Real-time gamepad state monitoring

## Code

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	utils.Logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🎮 DJI Tello Gamepad Example")
	fmt.Println("=============================")

	// List available gamepads
	gamepads := gamepad.ListGamepads()
	if len(gamepads) == 0 {
		fmt.Println("❌ No gamepads found. Please connect a gamepad and try again.")
		return
	}

	fmt.Println("📋 Available gamepads:")
	for i, gp := range gamepads {
		fmt.Printf("  %d. %s\n", i+1, gp)
	}

	// Show detailed gamepad info
	fmt.Println("\n🔍 Gamepad Details:")
	info := gamepad.GetGamepadInfo()
	for _, gp := range info {
		fmt.Printf("  ID: %s\n", gp["id"])
		fmt.Printf("  Name: %s\n", gp["name"])
		fmt.Printf("  Type: %s\n", gp["type"])
		if mapping, exists := gp["mapping"]; exists {
			fmt.Printf("  Mapping: %s\n", mapping)
		}
		fmt.Println()
	}

	// Load gamepad configuration
	config, err := gamepad.LoadConfigFromFile("configs/gamepad-default.json")
	if err != nil {
		fmt.Printf("❌ Failed to load gamepad config: %v\n", err)
		fmt.Println("💡 Using default configuration...")
		config = gamepad.DefaultConfig()
	}

	// Create gamepad handler
	handler, err := gamepad.NewHandler(gamepad.HandlerOptions{
		Config: config,
		OnRCValues: func(rcValues gamepad.RCValues) {
			fmt.Printf("🚁 RC: A=%d B=%d C=%d D=%d\n",
				rcValues.A, rcValues.B, rcValues.C, rcValues.D)
		},
		OnDroneAction: func(action gamepad.DroneAction) {
			fmt.Printf("🎯 Action: %s\n", action)
		},
		OnError: func(err error) {
			fmt.Printf("❌ Error: %v\n", err)
		},
	})
	if err != nil {
		log.Fatalf("Failed to create gamepad handler: %v", err)
	}
	defer handler.Stop()

	// Start handler
	if err := handler.Start(); err != nil {
		log.Fatalf("Failed to start gamepad handler: %v", err)
	}

	fmt.Println("🎮 Gamepad handler started!")
	fmt.Println("📝 Try moving the sticks and pressing buttons...")
	fmt.Println("⏹️  Press Ctrl+C to stop")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop - print gamepad state periodically
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutting down...")
			return
		case <-ticker.C:
			state := handler.GetState()
			printGamepadState(state)
		}
	}
}

// printGamepadState prints current gamepad state in a readable format
func printGamepadState(state *gamepad.GamepadState) {
	fmt.Printf("\r🎮 [%.1fs] ", time.Since(state.LastUpdate).Seconds())

	// Print axes
	axes := []gamepad.AxisType{"left_stick_x", "left_stick_y", "right_stick_x", "right_stick_y", "left_trigger", "right_trigger"}
	for _, axisType := range axes {
		if axis, exists := state.Axes[axisType]; exists {
			fmt.Printf("%s:%.2f ", string(axisType), axis.Value)
		}
	}

	// Print pressed buttons
	var pressedButtons []string
	for buttonType, button := range state.Buttons {
		if button.Pressed {
			pressedButtons = append(pressedButtons, string(buttonType))
		}
	}

	if len(pressedButtons) > 0 {
		fmt.Printf("🔘 %v", pressedButtons)
	}

	fmt.Print("   ") // Clear any remaining characters
}
```

## How to Run

```bash
go run -c 'package main; import "fmt"; func main() { fmt.Println("Use the code above in a main.go file") }'
```

Or save the code to a file and run:

```bash
# Save as gamepad_example.go
go run gamepad_example.go
```

## Features Demonstrated

1. **Gamepad Detection**: Automatically detects connected gamepads
2. **Configuration Loading**: Loads gamepad mappings from JSON config files
3. **Real-time Input**: Continuously monitors gamepad state
4. **RC Value Conversion**: Converts gamepad inputs to drone RC values
5. **Action Mapping**: Maps button presses to drone actions
6. **Error Handling**: Graceful error handling and logging
7. **Graceful Shutdown**: Proper cleanup on Ctrl+C

## Expected Output

```
🎮 DJI Tello Gamepad Example
=============================
📋 Available gamepads:
  1. Xbox Controller

🔍 Gamepad Details:
  ID: /dev/input/js0
  Name: Xbox Controller
  Type: xbox

🎮 Gamepad handler started!
📝 Try moving the sticks and pressing buttons...
⏹️  Press Ctrl+C to stop

🎮 [0.1s] left_stick_x:0.00 left_stick_y:0.00 right_stick_x:0.12 right_stick_y:-0.05
🚁 RC: A=50 B=50 C=60 D=45
🎯 Action: takeoff
```

## Configuration

The example uses `configs/gamepad-default.json` for gamepad mappings. You can customize this file to:
- Map different button layouts
- Adjust sensitivity curves
- Configure dead zones
- Set custom action mappings

## Requirements

- A connected gamepad (Xbox, PlayStation, or generic USB gamepad)
- Proper permissions to access input devices (may require sudo on Linux)
- Gamepad configuration file in `configs/` directory