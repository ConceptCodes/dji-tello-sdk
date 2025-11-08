package main

import (
	"fmt"
	"log"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
)

func main() {
	// Initialize the Tello commander
	commander, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize Tello commander: %v", err)
	}

	// Enter SDK mode
	fmt.Println("Entering SDK mode...")
	if err := commander.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Wait a moment for initialization
	time.Sleep(2 * time.Second)

	// Test read commands
	fmt.Println("\n=== Testing Read Commands ===")

	// Get battery percentage
	battery, err := commander.GetBatteryPercentage()
	if err != nil {
		fmt.Printf("Error getting battery: %v\n", err)
	} else {
		fmt.Printf("Battery: %d%%\n", battery)
	}

	// Get current speed
	speed, err := commander.GetSpeed()
	if err != nil {
		fmt.Printf("Error getting speed: %v\n", err)
	} else {
		fmt.Printf("Speed: %d cm/s\n", speed)
	}

	// Get height
	height, err := commander.GetHeight()
	if err != nil {
		fmt.Printf("Error getting height: %v\n", err)
	} else {
		fmt.Printf("Height: %d cm\n", height)
	}

	// Get flight time
	flightTime, err := commander.GetTime()
	if err != nil {
		fmt.Printf("Error getting flight time: %v\n", err)
	} else {
		fmt.Printf("Flight time: %d seconds\n", flightTime)
	}

	// Get temperature
	temp, err := commander.GetTemperature()
	if err != nil {
		fmt.Printf("Error getting temperature: %v\n", err)
	} else {
		fmt.Printf("Temperature: %d°C\n", temp)
	}

	// Get attitude (pitch, roll, yaw)
	pitch, roll, yaw, err := commander.GetAttitude()
	if err != nil {
		fmt.Printf("Error getting attitude: %v\n", err)
	} else {
		fmt.Printf("Attitude - Pitch: %d°, Roll: %d°, Yaw: %d°\n", pitch, roll, yaw)
	}

	// Get barometer reading
	baro, err := commander.GetBarometer()
	if err != nil {
		fmt.Printf("Error getting barometer: %v\n", err)
	} else {
		fmt.Printf("Barometer: %d m\n", baro)
	}

	// Get acceleration
	agx, agy, agz, err := commander.GetAcceleration()
	if err != nil {
		fmt.Printf("Error getting acceleration: %v\n", err)
	} else {
		fmt.Printf("Acceleration - X: %d, Y: %d, Z: %d (0.001g units)\n", agx, agy, agz)
	}

	// Get TOF distance
	tof, err := commander.GetTof()
	if err != nil {
		fmt.Printf("Error getting TOF distance: %v\n", err)
	} else {
		fmt.Printf("TOF distance: %d cm\n", tof)
	}

	fmt.Println("\n=== Read Commands Test Complete ===")
}