// Example: Basic flight control with DJI Tello
// This example demonstrates basic drone operations:
// - Initialization and SDK mode entry
// - Takeoff and landing
// - Simple flight patterns
// - Telemetry monitoring

package main

import (
	"fmt"
	"log"
	"time"

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
	fmt.Println("SDK mode initialized successfully")

	// Check battery before takeoff
	battery, err := drone.GetBatteryPercentage()
	if err != nil {
		log.Printf("Warning: Could not check battery: %v", err)
	} else {
		fmt.Printf("Battery level: %d%%\n", battery)
		if battery < 20 {
			log.Fatal("Battery too low for flight")
		}
	}

	// Take off
	fmt.Println("Taking off...")
	if err := drone.TakeOff(); err != nil {
		log.Fatalf("Takeoff failed: %v", err)
	}
	fmt.Println("Drone is airborne")

	// Wait for takeoff to complete
	time.Sleep(2 * time.Second)

	// Fly a square pattern
	fmt.Println("\nFlying a square pattern...")
	flySquare(drone)

	// Fly in a circle
	fmt.Println("\nFlying a circle...")
	flyCircle(drone)

	// Hover for a moment
	fmt.Println("\nHovering...")
	time.Sleep(2 * time.Second)

	// Land
	fmt.Println("\nLanding...")
	if err := drone.Land(); err != nil {
		log.Fatalf("Landing failed: %v", err)
	}
	fmt.Println("Landed successfully")

	// Final battery check
	battery, _ = drone.GetBatteryPercentage()
	fmt.Printf("\nFinal battery level: %d%%\n", battery)
}

// flySquare flies a square pattern
func flySquare(drone tello.TelloCommander) {
	sideLength := 100 // cm

	// Forward
	fmt.Printf("Moving forward %d cm...\n", sideLength)
	if err := drone.Forward(sideLength); err != nil {
		log.Printf("Forward movement failed: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Right
	fmt.Printf("Moving right %d cm...\n", sideLength)
	if err := drone.Right(sideLength); err != nil {
		log.Printf("Right movement failed: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Backward
	fmt.Printf("Moving backward %d cm...\n", sideLength)
	if err := drone.Backward(sideLength); err != nil {
		log.Printf("Backward movement failed: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Left
	fmt.Printf("Moving left %d cm...\n", sideLength)
	if err := drone.Left(sideLength); err != nil {
		log.Printf("Left movement failed: %v", err)
	}
	time.Sleep(1 * time.Second)
}

// flyCircle flies in a circle using rotation and forward movement
func flyCircle(drone tello.TelloCommander) {
	segments := 12
	segmentDistance := 20 // cm
	rotationAngle := 360 / segments

	for i := 0; i < segments; i++ {
		fmt.Printf("Segment %d/%d: Moving %d cm and rotating %d°...\n", i+1, segments, segmentDistance, rotationAngle)
		if err := drone.Forward(segmentDistance); err != nil {
			log.Printf("Forward movement failed: %v", err)
		}
		time.Sleep(500 * time.Millisecond)

		if err := drone.Clockwise(rotationAngle); err != nil {
			log.Printf("Rotation failed: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
