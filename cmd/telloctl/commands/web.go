package commands

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/models"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/pipeline"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/web"
	"github.com/spf13/cobra"
)

// WebCmd creates the web interface command
func WebCmd(drone tello.TelloCommander) *cobra.Command {
	var webPort int
	var enableML bool
	var configDir string

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
  telloctl web -p 9000                  # Start web interface on port 9000
  telloctl web --ml                     # Start with ML enabled`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var mlPipeline *pipeline.ConcurrentMLPipeline
			var mlResultChan <-chan ml.MLResult

			if enableML {
				fmt.Println("🤖 Initializing ML pipeline...")
				configManager := config.NewConfigManager(configDir)
				modelManager, err := models.NewModelManager("models")
				if err != nil {
					return fmt.Errorf("failed to create model manager: %w", err)
				}

				// Load ML configuration
				mlConfig, err := configManager.LoadMLConfig("ml-pipeline-config.json")
				if err != nil {
					// Try to create default config if not found
					if err := configManager.CreateDefaultConfigs(); err != nil {
						return fmt.Errorf("failed to create default ML configs: %w", err)
					}
					mlConfig, err = configManager.LoadMLConfig("ml-pipeline-config.json")
					if err != nil {
						return fmt.Errorf("failed to load ML configuration: %w", err)
					}
				}

				// Create and start pipeline
				mlPipeline = pipeline.NewConcurrentMLPipeline(&mlConfig.Pipeline, mlConfig.Processors, modelManager)
				if err := mlPipeline.Start(); err != nil {
					return fmt.Errorf("failed to start ML pipeline: %w", err)
				}
				defer mlPipeline.Stop()

				mlResultChan = mlPipeline.GetResults()
				fmt.Println("✅ ML pipeline started")
			}

			frameChan := drone.GetVideoFrameChannel()
			if frameChan == nil {
				return fmt.Errorf("failed to get video frame channel")
			}

			// Create channels for fan-out
			displayChan := make(chan transport.VideoFrame, 100)
			recorderChan := make(chan transport.VideoFrame, 100)
			mlChan := make(chan transport.VideoFrame, 100)

			// Start fan-out goroutine
			go func() {
				for frame := range frameChan {
					// Non-blocking sends to avoid stalling if one consumer is slow
					select {
					case displayChan <- frame:
					default:
					}

					select {
					case recorderChan <- frame:
					default:
					}

					if enableML {
						select {
						case mlChan <- frame:
						default:
						}
					}
				}
				close(displayChan)
				close(recorderChan)
				close(mlChan)
			}()

			// Initialize recorder
			// We use a timestamped filename for recording
			recordPath := fmt.Sprintf("tello_recording_%s.mp4", time.Now().Format("20060102_150405"))
			recorder, err := transport.NewVideoRecorderWithFormatAndChannel(recorderChan, recordPath, transport.FormatMP4)
			if err != nil {
				fmt.Printf("⚠️ Failed to initialize video recorder: %v\n", err)
			}

			webServer := web.NewWebServer(drone, recorder, mlPipeline, mlResultChan)

			// Create web video display with enhanced features
			display := transport.NewVideoDisplay(transport.DisplayTypeWeb)
			display.SetVideoChannel(displayChan)
			display.SetWebPort(webPort)
			display.SetMLResultChannel(mlResultChan)

			// Set ML config for overlay if enabled
			if enableML {
				configManager := config.NewConfigManager(configDir)
				if mlConfig, err := configManager.LoadMLConfig("ml-pipeline-config.json"); err == nil {
					display.SetMLConfig(mlConfig)
				}
			}

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

			// Start recorder (it will wait for StartRecording call to actually write to file,
			// but needs to be running to consume frames from channel)
			if recorder != nil {
				// We don't start recording immediately, but we need to ensure the recorder is ready
				// The recorder's StartRecording method starts both processing and writing.
				// Since we want to control writing via web UI, but we need to consume frames from recorderChan
				// to prevent blocking (though we use non-blocking send), we rely on the fact that
				// non-blocking send drops frames if channel is full.
				// So when recording is stopped, frames are dropped. When started, they are processed.
				// This is the desired behavior.
			}

			// If ML is enabled, feed frames to the pipeline
			if enableML && mlPipeline != nil {
				go func() {
					for frame := range mlChan {
						mlPipeline.ProcessFrame(frame.ToEnhancedFrame())
					}
				}()
			}

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
			if enableML {
				fmt.Printf("🧠 ML Pipeline: %s\n", "enabled")
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
					fmt.Println("\nReceived interrupt signal, stopping web interface...")
					return nil

				case <-ticker.C:
					// Update statistics
					stats := display.GetStats()
					if isRunning, ok := stats["is_running"].(bool); ok && isRunning {
						if frameCount, ok := stats["frame_count"].(int); ok {
							if duration, ok := stats["duration"].(time.Duration); ok {
								fps := float64(frameCount) / duration.Seconds()
								status := fmt.Sprintf("📈 Status: %d frames | %.1f FPS | running", frameCount, fps)
								if enableML {
									mlStats := mlPipeline.GetMetrics()
									status += fmt.Sprintf(" | ML: %.1f FPS", mlStats.FPS)
								}
								fmt.Println(status)
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
	cmd.Flags().BoolVar(&enableML, "ml", false, "Enable Machine Learning pipeline")
	cmd.Flags().StringVar(&configDir, "config-dir", "configs", "Configuration directory")

	return cmd
}
