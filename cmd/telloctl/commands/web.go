package commands

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/web"
	"github.com/spf13/cobra"
)

// WebCmd creates the web interface command
func WebCmd(drone tello.TelloCommander) *cobra.Command {
	var webPort int

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Start comprehensive web interface for drone control and monitoring",
		Long: `Start a comprehensive web interface that provides drone control, real-time video streaming,
telemetry monitoring, and ML integration in a single browser-based dashboard.

This command creates a full-featured web application that includes:
- Live video stream from drone camera
- Real-time telemetry display (battery, altitude, attitude)
- Web-based drone controls (takeoff, land, movement)
- ML detection overlays (if ML is enabled)
- Flight statistics and system status

Examples:
  telloctl web                         # Start web interface on default port 8080
  telloctl web -p 9000                  # Start web interface on port 9000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			webServer := web.NewWebServer(drone, nil)

			frameChan := drone.GetVideoFrameChannel()
			if frameChan == nil {
				return fmt.Errorf("failed to get video frame channel")
			}

			// Create web video display with enhanced features
			display := transport.NewVideoDisplay(transport.DisplayTypeWeb)
			display.SetVideoChannel(frameChan)
			display.SetWebPort(webPort)
			display.SetCustomWebHandlers(
				webServer.HandleIndex,
				func(mux *http.ServeMux) {
					webServer.SetupRoutesWithoutIndex(mux)
				},
			)

			// Start display
			if err := display.Start(); err != nil {
				return fmt.Errorf("failed to start web display: %w", err)
			}
			defer display.Close()

			// Attempt to connect automatically but continue if it fails
			if coordinator := webServer.ConnectionCoordinator(); coordinator != nil {
				defer func() {
					if err := coordinator.Disconnect(); err != nil {
						fmt.Printf("Error stopping video stream: %v\n", err)
					}
				}()

				if err := coordinator.Connect(); err != nil {
					fmt.Printf("⚠️ Unable to auto-connect to drone: %v\n", err)
					fmt.Println("   Open the web UI and use the Connect Drone button when ready.")
				} else {
					fmt.Println("✅ Drone connected and streaming")
				}
			}

			fmt.Printf("🌐 Web interface started\n")
			fmt.Printf("🔗 Open http://localhost:%d in your browser\n", webPort)
			fmt.Printf("📹 Video stream: %s\n", "enabled")
			fmt.Printf("🎮 Controls: %s\n", "enabled")
			fmt.Printf("📊 Telemetry: %s\n", "enabled")
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
					fmt.Println("\nReceived interrupt signal, stopping web interface...")
					return nil

				case <-ticker.C:
					// Update statistics
					stats := display.GetStats()
					if isRunning, ok := stats["is_running"].(bool); ok && isRunning {
						if frameCount, ok := stats["frame_count"].(int); ok {
							if duration, ok := stats["duration"].(time.Duration); ok {
								fps := float64(frameCount) / duration.Seconds()
								fmt.Printf("📈 Status: %d frames | %.1f FPS | running\n",
									frameCount, fps)
							}
						}
					} else {
						fmt.Println("Web interface stopped")
						return nil
					}
				}
			}
		},
	}

	cmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Web server port")

	return cmd
}
