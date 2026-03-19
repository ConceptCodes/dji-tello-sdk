// Example: Flip maneuvers demonstration
// This example demonstrates:
// - All flip directions
// - Flip timing and safety
// - Altitude requirements for flips
// - Flip sequence patterns

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

	// Wait for takeoff to complete and stabilize
	fmt.Println("Waiting for stable altitude...")
	time.Sleep(3 * time.Second)

	// Set up interrupt handler
	go func() {
		<-sigChan
		fmt.Println("\n\nStopping flip demo...")
		if err := drone.Land(); err != nil {
			log.Printf("Landing failed: %v", err)
		}
		os.Exit(0)
	}()

	// Print flip demo header
	printFlipDemoHeader()

	// Verify safe altitude for flips
	height, err := drone.GetHeight()
	if err != nil {
		log.Printf("Failed to check height: %v", err)
	}
	if height < 50 {
		fmt.Printf("\nWARNING: Current height is %d cm. ", height)
		fmt.Println("Minimum safe height for flips is 50 cm.")
		fmt.Println("Ascending to safe altitude...")
		drone.Up(60)
		time.Sleep(1 * time.Second)
	}

	// Demo 1: Basic flips in each direction
	fmt.Println("\n--- Demo 1: Basic Flips ---")
	basicFlipDemo(drone)

	// Demo 2: Flip sequence pattern
	fmt.Println("\n--- Demo 2: Flip Sequence ---")
	flipSequenceDemo(drone)

	// Demo 3: Altitude-aware flips
	fmt.Println("\n--- Demo 3: Altitude-Aware Flips ---")
	altitudeFlipDemo(drone)

	// Land
	fmt.Println("\nLanding...")
	if err := drone.Land(); err != nil {
		log.Fatalf("Landing failed: %v", err)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("\nFlip demo completed!")
}

// printFlipDemoHeader displays flip demo information
func printFlipDemoHeader() {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║              FLIP MANEUVER DEMO                         ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                            ║")
	fmt.Println("║  FLIP SAFETY REQUIREMENTS:                                 ║")
	fmt.Println("║  - Minimum altitude: 50 cm                                 ║")
	fmt.Println("║  - Clear, open space (10m diameter)                      ║")
	fmt.Println("║  - Battery > 20% recommended                              ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  FLIP DIRECTIONS:                                          ║")
	fmt.Println("║  Forward, Backward, Left, Right                         ║")
	fmt.Println("║                                                            ║")
	fmt.Println("║  NOTE: Flips consume significant battery                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}

// basicFlipDemo performs basic flips in each direction
func basicFlipDemo(drone tello.TelloCommander) {
	flips := []struct {
		direction string
		flip     tello.FlipDirection
	}{
		{"Forward", tello.FlipForward},
		{"Backward", tello.FlipBackward},
		{"Left", tello.FlipLeft},
		{"Right", tello.FlipRight},
	}

	for _, fl := range flips {
		fmt.Printf("\nFlip %s...", fl.direction)
		if err := drone.Flip(fl.flip); err != nil {
			log.Printf("Flip failed: %v", err)
			fmt.Println("    Stabilizing...")
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("    Success!")
		}

		// Wait between flips
		time.Sleep(1 * time.Second)
	}
}

// flipSequenceDemo performs a choreographed flip sequence
func flipSequenceDemo(drone tello.TelloCommander) {
	sequences := []struct {
		name     string
		pattern  []tello.FlipDirection
	}{
		{
			name: "The Box",
			pattern: []tello.FlipDirection{
				tello.FlipForward,
				tello.FlipRight,
				tello.FlipBackward,
				tello.FlipLeft,
			},
		},
		{
			name: "The Cross",
			pattern: []tello.FlipDirection{
				tello.FlipLeft,
				tello.FlipForward,
				tello.FlipRight,
				tello.FlipBackward,
			},
		},
		{
			name: "The Spin",
			pattern: []tello.FlipDirection{
				tello.FlipForward,
				tello.FlipForward,
				tello.FlipBackward,
				tello.FlipBackward,
			},
		},
	}

	for _, seq := range sequences {
		fmt.Printf("\nSequence: %s\n", seq.name)
		fmt.Println("Executing pattern:", seq.pattern)

		for i, flip := range seq.pattern {
			fmt.Printf("  Step %d/%d: ", i+1, len(seq.pattern))
			if err := drone.Flip(flip); err != nil {
				log.Printf("Failed: %v", err)
			} else {
				fmt.Println("OK")
			}
			time.Sleep(500 * time.Millisecond)
		}

		// Pause between sequences
		fmt.Println("  Pausing...")
		time.Sleep(1 * time.Second)
	}
}

// altitudeFlipDemo demonstrates flips at different altitudes
func altitudeFlipDemo(drone tello.TelloCommander) {
	altitudes := []struct {
		name string
		h    int
	}{
		{"Low (50cm)", 50},
		{"Medium (100cm)", 100},
		{"High (150cm)", 150},
	}

	for _, alt := range altitudes {
		fmt.Printf("\nAltitude: %s", alt.name)

		// Adjust altitude if needed
		height, _ := drone.GetHeight()
		if height < alt.h {
			drone.Up(alt.h - height)
			time.Sleep(1 * time.Second)
		} else if height > alt.h {
			drone.Down(height - alt.h)
			time.Sleep(1 * time.Second)
		}

		// Perform forward flip at this altitude
		fmt.Println("  Performing forward flip...")
		if err := drone.Flip(tello.FlipForward); err != nil {
			log.Printf("    Failed: %v", err)
		} else {
			fmt.Println("    Success!")
		}

		time.Sleep(1 * time.Second)
	}
}
