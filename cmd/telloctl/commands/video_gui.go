package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
)

// VideoGUICmd creates the video GUI command
func VideoGUICmd(drone tello.TelloCommander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video-gui",
		Short: "Start native GUI window for drone video feed",
		Long: `Start a native graphical interface to display real-time video feed from the drone.
This command opens a desktop GUI window showing the live video stream from the drone.

Examples:
  telloctl video-gui                    # Start native GUI window`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure SDL runs on main thread
			var runErr error
			sdl.Main(func() {
				runErr = runVideoGUI(drone)
			})
			return runErr
		},
	}

	return cmd
}

func runVideoGUI(drone tello.TelloCommander) error {
	// Start video stream
	if err := drone.StreamOn(); err != nil {
		return fmt.Errorf("failed to start video stream: %w", err)
	}
	defer drone.StreamOff()

	// Get video frame channel
	frameChan := drone.GetVideoFrameChannel()
	if frameChan == nil {
		return fmt.Errorf("failed to get video frame channel")
	}

	// Create native video display
	display := transport.NewVideoDisplay(transport.DisplayTypeNative)
	display.SetVideoChannel(frameChan)

	// Start display
	if err := display.Start(); err != nil {
		return fmt.Errorf("failed to start video display: %w", err)
	}
	defer display.Stop()

	fmt.Printf("🎥 Video GUI started (native mode)\n")
	fmt.Println("🖥️ Native GUI window opened")
	fmt.Println("Press Ctrl+C to stop")

	// Setup interrupt handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Monitor display
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS for UI updates
	defer ticker.Stop()

	for {
		select {
		case <-interrupt:
			fmt.Println("\nReceived interrupt signal, stopping video GUI...")
			return nil

		case <-ticker.C:
			// Process SDL events on main thread
			display.ProcessEvents()

			// Check if display is still running
			if !display.IsRunning() {
				fmt.Println("Video display stopped")
				return nil
			}
		}
	}
}
