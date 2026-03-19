// Example: Safety manager usage
// This example demonstrates:
// - Wrapping drone with safety manager
// - Configurable safety limits
// - Auto-land on low battery
// - Connection timeout handling

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize the drone
	baseDrone, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize drone: %v", err)
	}

	// Enter SDK mode
	fmt.Println("Entering SDK mode...")
	if err := baseDrone.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Create safety manager with default configuration
	fmt.Println("Initializing safety manager...")
	safetyConfig := safety.DefaultConfig()
	safetyManager, err := safety.NewManager(baseDrone, safetyConfig)
	if err != nil {
		log.Fatalf("Failed to create safety manager: %v", err)
	}

	// Print safety limits
	printSafetyLimits(safetyConfig)

	// Set up event callback
	safetyManager.SetEventCallback(func(event *safety.SafetyEvent) {
		printSafetyEvent(event)
	})

	// Take off with safety manager
	fmt.Println("\nTaking off with safety protection...")
	if err := safetyManager.TakeOff(); err != nil {
		log.Fatalf("Takeoff failed: %v", err)
	}

	// Set up interrupt handler
	go func() {
		<-sigChan
		fmt.Println("\n\nStopping flight...")
		if err := safetyManager.Land(); err != nil {
			log.Printf("Landing failed: %v", err)
		}
		os.Exit(0)
	}()

	// Fly with safety manager
	fmt.Println("\nFlying with safety manager active...")
	fmt.Println("Safety features:")
	fmt.Println("  - Auto-land when battery < 20%")
	fmt.Println("  - RC limits enforced (50cm max)")
	fmt.Println("  - Connection timeout protection")
	fmt.Println("\nPress Ctrl+C to land and exit\n")

	// Perform some flights to demonstrate safety features
	flightDemo(safetyManager)

	// Land
	fmt.Println("\nLanding with safety manager...")
	if err := safetyManager.Land(); err != nil {
		log.Printf("Landing failed: %v", err)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("Flight completed safely")
}

// flightDemo performs a simple flight pattern
func flightDemo(manager safety.Manager) {
	// Move forward
	fmt.Println("Moving forward 50cm...")
	if err := manager.Forward(50); err != nil {
		log.Printf("Forward movement blocked by safety: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Try to move beyond limit (will be blocked)
	fmt.Println("\nAttempting to move 200cm (beyond limit)...")
	if err := manager.Forward(200); err != nil {
		fmt.Printf("Movement blocked: %v\n", err)
	}
	time.Sleep(1 * time.Second)

	// Move right
	fmt.Println("\nMoving right 50cm...")
	if err := manager.Right(50); err != nil {
		log.Printf("Right movement blocked by safety: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Hover
	fmt.Println("\nHovering for 3 seconds...")
	time.Sleep(3 * time.Second)
}

// printSafetyLimits displays current safety configuration
func printSafetyLimits(config *safety.Config) {
	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  SAFETY MANAGER CONFIGURATION               ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")

	fmt.Println("║ ALTITUDE LIMITS                                            ║")
	fmt.Printf("║  Min Height:     %3d cm                                       ║\n", config.Altitude.MinHeight)
	fmt.Printf("║  Max Height:     %3d cm                                       ║\n", config.Altitude.MaxHeight)
	fmt.Printf("║  Takeoff Height: %3d cm                                       ║\n", config.Altitude.TakeoffHeight)

	fmt.Println("║                                                            ║")
	fmt.Println("║ VELOCITY LIMITS                                            ║")
	fmt.Printf("║  Max Horizontal:  %3d cm/s                                   ║\n", config.Velocity.MaxHorizontal)
	fmt.Printf("║  Max Vertical:    %3d cm/s                                   ║\n", config.Velocity.MaxVertical)
	fmt.Printf("║  Max Yaw:        %3d°/s                                     ║\n", config.Velocity.MaxYaw)

	fmt.Println("║                                                            ║")
	fmt.Println("║ BATTERY SAFETY                                            ║")
	fmt.Printf("║  Warning:        %3d%%                                      ║\n", config.Battery.WarningThreshold)
	fmt.Printf("║  Critical:       %3d%%                                      ║\n", config.Battery.CriticalThreshold)
	fmt.Printf("║  Emergency:      %3d%%                                      ║\n", config.Battery.EmergencyThreshold)
	fmt.Printf("║  Auto-Land:      %-5v                                     ║\n", config.Battery.EnableAutoLand)

	fmt.Println("║                                                            ║")
	fmt.Println("║ CONNECTION SAFETY                                          ║")
	fmt.Printf("║  Timeout:        %5d ms                                   ║\n", config.Emergency.ConnectionTimeout)

	fmt.Println("║                                                            ║")
	fmt.Println("║ BEHAVIORAL LIMITS                                         ║")
	fmt.Printf("║  Enable Flips:   %-5v                                     ║\n", config.Behavioral.EnableFlips)
	fmt.Printf("║  Min Flip Ht:    %3d cm                                    ║\n", config.Behavioral.MinFlipHeight)
	fmt.Printf("║  Max Flight Tm:  %3d sec                                    ║\n", config.Behavioral.MaxFlightTime)

	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

// printSafetyEvent displays safety events
func printSafetyEvent(event *safety.SafetyEvent) {
	fmt.Printf("\n[SAFETY EVENT] %s: %s\n", event.Type, event.Message)
	if event.Data != nil {
		fmt.Printf("  Details: %+v\n", event.Data)
	}
}
