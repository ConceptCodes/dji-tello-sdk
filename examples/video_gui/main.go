package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

func main() {
	// Initialize drone
	drone, err := tello.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize drone: %v", err)
	}

	// Enter SDK mode
	if err := drone.Init(); err != nil {
		log.Fatalf("Failed to enter SDK mode: %v", err)
	}

	// Start video stream
	if err := drone.StreamOn(); err != nil {
		log.Fatalf("Failed to start video stream: %v", err)
	}
	defer drone.StreamOff()

	// Get video frame channel
	frameChan := drone.GetVideoFrameChannel()
	if frameChan == nil {
		log.Fatalf("Failed to get video frame channel")
	}

	// Create video display (web interface)
	display := transport.NewVideoDisplay(transport.DisplayTypeWeb)
	display.SetVideoChannel(frameChan)
	display.SetWebPort(8080)

	// Start display
	if err := display.Start(); err != nil {
		log.Fatalf("Failed to start video display: %v", err)
	}
	defer display.Close()

	log.Println("🎥 Video GUI started!")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Stopping video GUI...")
}
