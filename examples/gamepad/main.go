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

	// Start the handler
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

// printGamepadState prints the current gamepad state in a readable format
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
