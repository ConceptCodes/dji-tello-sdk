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
)

func VideoGUICmd(drone tello.TelloCommander) *cobra.Command {
	var displayType string
	var webPort int

	cmd := &cobra.Command{
		Use:   "video-gui",
		Short: "Display real-time video feed from drone in GUI",
		Long: `Display real-time video feed from the drone in a graphical interface.
This command opens a video display window showing the live video stream from the drone.

Display Types:
  terminal - ASCII art display in terminal
  web      - Web browser interface (default)

Examples:
  telloctl video-gui                    # Start web GUI (default)
  telloctl video-gui -t terminal         # Start terminal GUI
  telloctl video-gui -t web -p 8080    # Start web GUI on port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Determine display type
			var videoDisplayType transport.VideoDisplayType
			switch displayType {
			case "terminal":
				videoDisplayType = transport.DisplayTypeTerminal
			case "web":
				videoDisplayType = transport.DisplayTypeWeb
			default:
				videoDisplayType = transport.DisplayTypeWeb
			}

			// Create video display
			display := transport.NewVideoDisplay(videoDisplayType)
			display.SetVideoChannel(frameChan)

			if videoDisplayType == transport.DisplayTypeWeb {
				display.SetWebPort(webPort)
			}

			// Start display
			if err := display.Start(); err != nil {
				return fmt.Errorf("failed to start video display: %w", err)
			}
			defer display.Close()

			fmt.Printf("🎥 Video GUI started (%s mode)\n", videoDisplayType)
			if videoDisplayType == transport.DisplayTypeWeb {
				fmt.Printf("🌐 Open http://localhost:%d in your browser\n", webPort)
			}
			fmt.Println("Press Ctrl+C to stop")

			// Setup interrupt handling
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

			// Monitor display
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-interrupt:
					fmt.Println("\nReceived interrupt signal, stopping video GUI...")
					return nil

				case <-ticker.C:
					// Update statistics
					stats := display.GetStats()
					if isRunning, ok := stats["is_running"].(bool); ok && isRunning {
						if frameCount, ok := stats["frame_count"].(int); ok {
							if duration, ok := stats["duration"].(time.Duration); ok {
								fps := float64(frameCount) / duration.Seconds()
								fmt.Printf("📊 Stats: %d frames | %.1f FPS | %.1fs elapsed\n",
									frameCount, fps, duration.Seconds())
							}
						}
					} else {
						fmt.Println("Video display stopped")
						return nil
					}
				}
			}
		},
	}

	cmd.Flags().StringVarP(&displayType, "type", "t", "web", "Display type (terminal or web)")
	cmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Web server port (for web display)")

	return cmd
}
