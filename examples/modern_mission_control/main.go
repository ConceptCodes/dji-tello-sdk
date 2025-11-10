package main

import (
	"log"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/web"
)

func main() {
	// Initialize Tello commander
	commander, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize Tello commander: %v", err)
	}

	// Create enhanced video display with modern web UI
	display := web.NewEnhancedVideoDisplay(commander, nil)

	// Start the enhanced display
	if err := display.StartEnhanced(); err != nil {
		log.Fatalf("Failed to start enhanced display: %v", err)
	}

	log.Println("Modern Mission Control UI started on http://localhost:8080")

	// Keep the application running
	select {}
}
