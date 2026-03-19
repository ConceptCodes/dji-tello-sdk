// Example: Xbox controller gamepad support
// This example demonstrates:
// - Xbox controller detection and connection
// - Custom button mappings for flight control
// - Real-time RC control via analog sticks
// - Button-based actions (takeoff, land, flips, etc.)

package main

import (
	"fmt"
	"log"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
	// Initialize the drone
	drone, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize drone: %v", err)
	}

	// Enter SDK mode
	fmt.Println("Entering SDK mode...")
	if err := drone.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Load Xbox preset configuration
	fmt.Println("Loading Xbox controller configuration...")
	config, err := gamepad.LoadConfigFromFile("configs/gamepad-xbox.json")
	if err != nil {
		fmt.Println("Xbox config not found, using Xbox preset...")
		config = gamepad.XboxConfig()
	}

	// Create gamepad handler with Xbox mappings
	handler, err := gamepad.NewHandler(gamepad.HandlerOptions{
		Config: config,
		OnCommand: func(cmd gamepad.Command) {
			// Handle commands based on type
			handleCommand(drone, cmd)
		},
		OnError: func(err error) {
			log.Printf("Gamepad error: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("Failed to create gamepad handler: %v", err)
	}
	defer handler.Stop()

	// Start gamepad processing
	fmt.Println("Starting gamepad handler...")
	if err := handler.Start(); err != nil {
		log.Fatalf("Failed to start gamepad handler: %v", err)
	}

	// Display Xbox controller layout
	printXboxLayout()

	// Keep the program running
	fmt.Println("\nGamepad control active. Press Ctrl+C to exit.")
	select {}
}

// handleCommand processes gamepad commands and executes drone actions
func handleCommand(drone tello.TelloCommander, cmd gamepad.Command) {
	switch cmd.Type {
	case gamepad.CommandRC:
		// Handle RC control values
		if rcValues, ok := cmd.Data.(gamepad.RCValues); ok {
			err := drone.SetRcControl(rcValues.A, rcValues.B, rcValues.C, rcValues.D)
			if err != nil {
				log.Printf("RC control failed: %v", err)
			}
		}

	case gamepad.CommandAction:
		// Handle discrete actions
		if action, ok := cmd.Data.(gamepad.DroneAction); ok {
			handleDroneAction(drone, action)
		}
	}
}

// handleDroneAction executes drone actions based on button presses
func handleDroneAction(drone tello.TelloCommander, action gamepad.DroneAction) {
	switch action {
	case gamepad.ActionTakeoff:
		fmt.Println("ACTION: Takeoff")
		if err := drone.TakeOff(); err != nil {
			log.Printf("Takeoff failed: %v", err)
		}

	case gamepad.ActionLand:
		fmt.Println("ACTION: Land")
		if err := drone.Land(); err != nil {
			log.Printf("Landing failed: %v", err)
		}

	case gamepad.ActionEmergency:
		fmt.Println("ACTION: Emergency Stop")
		if err := drone.Emergency(); err != nil {
			log.Printf("Emergency stop failed: %v", err)
		}

	case gamepad.ActionFlipForward:
		fmt.Println("ACTION: Flip Forward")
		if err := drone.Flip(tello.FlipForward); err != nil {
			log.Printf("Flip forward failed: %v", err)
		}

	case gamepad.ActionFlipBackward:
		fmt.Println("ACTION: Flip Backward")
		if err := drone.Flip(tello.FlipBackward); err != nil {
			log.Printf("Flip backward failed: %v", err)
		}

	case gamepad.ActionFlipLeft:
		fmt.Println("ACTION: Flip Left")
		if err := drone.Flip(tello.FlipLeft); err != nil {
			log.Printf("Flip left failed: %v", err)
		}

	case gamepad.ActionFlipRight:
		fmt.Println("ACTION: Flip Right")
		if err := drone.Flip(tello.FlipRight); err != nil {
			log.Printf("Flip right failed: %v", err)
		}

	case gamepad.ActionStreamOn:
		fmt.Println("ACTION: Start Video Stream")
		if err := drone.StreamOn(); err != nil {
			log.Printf("Stream on failed: %v", err)
		}

	case gamepad.ActionStreamOff:
		fmt.Println("ACTION: Stop Video Stream")
		if err := drone.StreamOff(); err != nil {
			log.Printf("Stream off failed: %v", err)
		}

	default:
		fmt.Printf("ACTION: Unknown (%v)\n", action)
	}
}

// printXboxLayout displays the Xbox controller button layout
func printXboxLayout() {
	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              XBOX CONTROLLER MAPPING                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                          ║")
	fmt.Println("║  [Y] Flip Fwd    [B] Land   [X] Emergency   [A] Takeoff ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  [LB] Flip Left  [RB] Flip Right                        ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  [LT] Ascend      [RT] Descend                           ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  [SELECT] Video Off  [START] Video On                     ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  LEFT STICK: Throttle (Y) | Yaw (X)                     ║")
	fmt.Println("║  RIGHT STICK: Pitch (Y) | Roll (X)                      ║")
	fmt.Println("║                                                          ║")
	fmt.Println("║  DPAD: Available for custom mappings                      ║")
	fmt.Println("║                                                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println("\nSTICK DEADZONE:", 0.15)
	fmt.Println("RC LIMITS: H:50 V:50 Yaw:50")
	fmt.Println("\nFLYING SAFETY:")
	fmt.Println("  - Start in a clear, open area")
	fmt.Println("  - Ensure battery > 20% before takeoff")
	fmt.Println("  - Keep line of sight with the drone")
	fmt.Println("  - Be prepared to use Emergency (X) if needed")
}
