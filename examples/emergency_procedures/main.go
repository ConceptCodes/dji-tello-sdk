// Example: Emergency procedures and recovery
// This example demonstrates:
// - Emergency stop command
// - Connection recovery
// - Error handling patterns
// - Safe shutdown procedures

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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

	// Take off
	fmt.Println("Taking off...")
	if err := drone.TakeOff(); err != nil {
		log.Fatalf("Takeoff failed: %v", err)
	}

	// Set up interrupt handler
	go func() {
		<-sigChan
		fmt.Println("\n\nEmergency stop triggered...")
		emergencyStop(drone)
		os.Exit(0)
	}()

	fmt.Println("\nEmergency Procedures Demo")
	fmt.Println("═══════════════════════════════════════════════════")

	// Demo 1: Emergency stop on demand
	demoEmergencyStop(drone)

	// Demo 2: Safe recovery from error
	demoErrorRecovery(drone)

	// Demo 3: Graceful shutdown
	demoGracefulShutdown(drone)
}

// demoEmergencyStop demonstrates emergency stop procedure
func demoEmergencyStop(drone tello.TelloCommander) {
	fmt.Println("\n--- Demo 1: Emergency Stop ---")
	fmt.Println("Type 'e' and press Enter to trigger emergency stop")

	// Simple input channel for emergency trigger
	inputChan := make(chan string, 1)
	go func() {
		var input string
		fmt.Scanln(&input)
		inputChan <- input
	}()

	select {
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout - no emergency stop triggered")
		return
	case input := <-inputChan:
		if input == "e" || input == "E" {
			fmt.Println("EMERGENCY STOP INITIATED!")
			emergencyStop(drone)
		}
	}
}

// emergencyStop performs emergency stop procedures
func emergencyStop(drone tello.TelloCommander) {
	// Send emergency stop command
	if err := drone.Emergency(); err != nil {
		log.Printf("Emergency stop command failed: %v", err)
	} else {
		fmt.Println("Emergency stop command sent successfully")
	}

	// Wait for drone to stabilize
	time.Sleep(2 * time.Second)

	// Verify drone has stopped
	battery, _ := drone.GetBatteryPercentage()
	height, _ := drone.GetHeight()
	fmt.Printf("Drone status - Battery: %d%%, Height: %d cm\n", battery, height)

	if height <= 10 {
		fmt.Println("Drone has successfully stopped")
	} else {
		fmt.Println("Warning: Drone may still be airborne")
	}
}

// demoErrorRecovery demonstrates error handling and recovery
func demoErrorRecovery(drone tello.TelloCommander) {
	fmt.Println("\n--- Demo 2: Error Recovery ---")
	fmt.Println("Attempting movement with error handling...")

	// Attempt movements with error handling
	movements := []struct {
		name     string
		distance int
	}{
		{"Forward 100cm", 100},
		{"Backward 50cm", 50},
		{"Left 50cm", 50},
		{"Right 50cm", 50},
	}

	for _, mv := range movements {
		fmt.Printf("  %s... ", mv.name)

		// Try-catch pattern for error handling
		if err := drone.Forward(mv.distance); err != nil {
			fmt.Printf("FAILED: %v\n", err)
			handleMovementError(err, drone)
		} else {
			fmt.Println("OK")
		}
		time.Sleep(1 * time.Second)
	}
}

// handleMovementError processes movement errors and attempts recovery
func handleMovementError(err error, drone tello.TelloCommander) {
	// Check error type and take appropriate action
	fmt.Printf("    Error type: %T\n", err)

	// In a real application, you might:
	// 1. Log the error with timestamp
	// 2. Check battery status
	// 3. Verify connection
	// 4. Attempt safe landing if critical

	battery, _ := drone.GetBatteryPercentage()
	if battery < 20 {
		fmt.Println("    Action: Low battery detected - landing")
		drone.Land()
		return
	}

	fmt.Println("    Action: Continuing (error logged)")
}

// demoGracefulShutdown demonstrates proper shutdown procedure
func demoGracefulShutdown(drone tello.TelloCommander) {
	fmt.Println("\n--- Demo 3: Graceful Shutdown ---")
	fmt.Println("Performing graceful shutdown sequence...")

	// Step 1: Stop any active video streaming
	fmt.Println("  1. Stopping video streaming...")
	if err := drone.StreamOff(); err != nil {
		log.Printf("    Failed to stop video: %v", err)
	}

	// Step 2: Ensure drone is at safe height
	fmt.Println("  2. Checking altitude...")
	height, _ := drone.GetHeight()
	if height > 30 {
		fmt.Printf("    Current height: %d cm - descending...\n", height)
		for height > 30 {
			drone.Down(20)
			time.Sleep(1 * time.Second)
			height, _ = drone.GetHeight()
		}
	} else {
		fmt.Printf("    Current height: %d cm - safe to land\n", height)
	}

	// Step 3: Land the drone
	fmt.Println("  3. Landing...")
	if err := drone.Land(); err != nil {
		log.Printf("    Landing failed: %v", err)
		// Fallback: emergency stop
		fmt.Println("    Fallback: Using emergency stop")
		drone.Emergency()
	} else {
		fmt.Println("    Landed successfully")
	}

	// Step 4: Wait for landing to complete
	fmt.Println("  4. Waiting for landing confirmation...")
	time.Sleep(2 * time.Second)

	// Step 5: Final status check
	fmt.Println("  5. Final status check...")
	battery, _ := drone.GetBatteryPercentage()
	height, _ = drone.GetHeight()
	fmt.Printf("    Battery: %d%%, Height: %d cm\n", battery, height)

	if height == 0 {
		fmt.Println("    Status: Drone safely landed")
	} else {
		fmt.Println("    Warning: Drone may still be airborne")
	}

	fmt.Println("Graceful shutdown complete")
}
