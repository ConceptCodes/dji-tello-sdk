// Example: Real-time telemetry monitoring
// This example demonstrates:
// - Continuous telemetry monitoring
// - Battery, altitude, and attitude tracking
// - Formatted console output
// - Graceful shutdown handling

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

type TelemetryData struct {
	Battery     int
	Height      int
	Pitch       int
	Roll        int
	Yaw         int
	Speed       int
	Temperature int
	Barometer   int
	AccX        int
	AccY        int
	AccZ        int
	ToF         int
}

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

	// Take off for realistic telemetry
	fmt.Println("Taking off...")
	if err := drone.TakeOff(); err != nil {
		log.Fatalf("Takeoff failed: %v", err)
	}

	// Wait for takeoff to complete
	time.Sleep(2 * time.Second)

	// Start telemetry monitoring
	fmt.Println("\n" + string(rune(033)) + "[2J") // Clear screen
	fmt.Println("Starting telemetry monitoring...")
	fmt.Println("Press Ctrl+C to stop\n")

	// Create ticker for regular telemetry updates
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Start goroutine for graceful shutdown
	go func() {
		<-sigChan
		fmt.Println("\n\nStopping telemetry monitoring...")
		if err := drone.Land(); err != nil {
			log.Printf("Landing failed: %v", err)
		}
		os.Exit(0)
	}()

	// Monitor telemetry continuously
	for {
		select {
		case <-ticker.C:
			// Fetch all telemetry data
			data, err := fetchTelemetry(drone)
			if err != nil {
				log.Printf("Failed to fetch telemetry: %v", err)
				continue
			}

			// Display telemetry
			displayTelemetry(data)
		}
	}
}

// fetchTelemetry collects all available telemetry data
func fetchTelemetry(drone tello.TelloCommander) (TelemetryData, error) {
	var data TelemetryData
	var err error

	data.Battery, err = drone.GetBatteryPercentage()
	if err != nil {
		log.Printf("Battery error: %v", err)
	}

	data.Height, err = drone.GetHeight()
	if err != nil {
		log.Printf("Height error: %v", err)
	}

	data.Pitch, data.Roll, data.Yaw, err = drone.GetAttitude()
	if err != nil {
		log.Printf("Attitude error: %v", err)
	}

	data.Speed, err = drone.GetSpeed()
	if err != nil {
		log.Printf("Speed error: %v", err)
	}

	data.Temperature, err = drone.GetTemperature()
	if err != nil {
		log.Printf("Temperature error: %v", err)
	}

	data.Barometer, err = drone.GetBarometer()
	if err != nil {
		log.Printf("Barometer error: %v", err)
	}

	data.AccX, data.AccY, data.AccZ, err = drone.GetAcceleration()
	if err != nil {
		log.Printf("Acceleration error: %v", err)
	}

	data.ToF, err = drone.GetTof()
	if err != nil {
		log.Printf("ToF error: %v", err)
	}

	return data, nil
}

// displayTelemetry formats and displays telemetry data
func displayTelemetry(data TelemetryData) {
	fmt.Println(string(rune(033)) + "[H") // Move cursor to top

	// Header
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                 DJI TELLO TELEMETRY                          ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	// Battery and Power
	fmt.Println("║ POWER                                                       ║")
	fmt.Printf("║  Battery:        ", " ")
	printBar(data.Battery, 100, 20, 10, 80)
	fmt.Printf("%3d%%", data.Battery)
	for i := 18 - data.Battery/3; i > 0; i-- {
		fmt.Print(" ")
	}
	fmt.Println(" ║")

	fmt.Printf("║  Temperature:    %-4d°C                                          ║\n", data.Temperature)

	// Flight Status
	fmt.Println("║ FLIGHT STATUS                                               ║")
	fmt.Printf("║  Altitude:       %4d cm                                      ║\n", data.Height)
	fmt.Printf("║  Speed:          %4d cm/s                                    ║\n", data.Speed)
	fmt.Printf("║  Time of Flight: %4d cm                                      ║\n", data.ToF)

	// Attitude
	fmt.Println("║ ATTITUDE (deg)                                              ║")
	fmt.Printf("║  Pitch:    ", " ")
	printAttitudeBar(data.Pitch, -45, 45, 15)
	fmt.Printf(" %3d", data.Pitch)
	fmt.Println("                                   ║")

	fmt.Printf("║  Roll:     ", " ")
	printAttitudeBar(data.Roll, -45, 45, 15)
	fmt.Printf(" %3d", data.Roll)
	fmt.Println("                                   ║")

	fmt.Printf("║  Yaw:      ", " ")
	printAttitudeBar(data.Yaw, 0, 360, 15)
	fmt.Printf(" %3d", data.Yaw)
	fmt.Println("                                   ║")

	// Sensors
	fmt.Println("║ SENSORS                                                     ║")
	fmt.Printf("║  Barometer:    %4d                                          ║\n", data.Barometer)
	fmt.Println("║  Acceleration:")
	fmt.Printf("║    X: %4d    Y: %4d    Z: %4d                            ║\n", data.AccX, data.AccY, data.AccZ)

	// Footer
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("\nLast updated: %s\n", time.Now().Format("15:04:05.000"))
}

// printBar prints a horizontal progress bar
func printBar(value, max, width, lowThreshold, highThreshold int) {
	filled := (value * width) / max
	if filled > width {
		filled = width
	}

	fmt.Print("[")
	for i := 0; i < width; i++ {
		if i < filled {
			if value < lowThreshold {
				fmt.Print("█") // Low - solid
			} else if value > highThreshold {
				fmt.Print("█") // High - solid
			} else {
				fmt.Print("█") // Normal - solid
			}
		} else {
			fmt.Print("░")
		}
	}
	fmt.Print("]")
}

// printAttitudeBar prints a centered attitude indicator
func printAttitudeBar(value, min, max, width int) {
	normalized := (value - min) * width / (max - min)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > width {
		normalized = width
	}

	mid := width / 2
	fmt.Print("[")
	for i := 0; i < width; i++ {
		if i == mid {
			fmt.Print("|")
		} else if i < normalized {
			fmt.Print(">")
		} else {
			fmt.Print("-")
		}
	}
	fmt.Print("]")
}
