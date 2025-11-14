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

func StreamOnCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "streamon",
		Short: "Start video streaming from the drone",
		Long: `Start video streaming from the drone.
This command enables the video stream and begins receiving H.264 video data.
The video data can be accessed through the SDK's video frame channel.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting video stream...")

			if err := drone.StreamOn(); err != nil {
				return fmt.Errorf("failed to start video stream: %w", err)
			}

			fmt.Println("Video stream started successfully!")
			fmt.Println("Video frames are now being received on the video channel.")
			fmt.Println("Use 'telloctl stream' command to monitor the video stream.")

			return nil
		},
	}
}

func StreamOffCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "streamoff",
		Short: "Stop video streaming from the drone",
		Long: `Stop video streaming from the drone.
This command disables the video stream and stops receiving video data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Stopping video stream...")

			if err := drone.StreamOff(); err != nil {
				return fmt.Errorf("failed to stop video stream: %w", err)
			}

			fmt.Println("Video stream stopped successfully!")

			return nil
		},
	}
}

func StreamCmd(drone tello.TelloCommander) *cobra.Command {
	var duration int
	var saveFile string
	var format string

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Monitor and optionally save video stream from the drone",
		Long: `Monitor video stream from the drone in real-time.
This command displays video frame information and can optionally save frames to a file.

Examples:
  telloctl stream                    # Monitor video stream indefinitely
  telloctl stream -d 30              # Monitor for 30 seconds
  telloctl stream -s video.h264      # Save video stream to H.264 file
  telloctl stream -s video.mp4 -f mp4 # Save video stream to MP4 file
  telloctl stream -d 60 -s video.mp4 -f mp4 # Monitor for 60 seconds and save to MP4 file`,
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

			// Determine video format
			videoFormat := transport.FormatH264
			if format == "mp4" {
				videoFormat = transport.FormatMP4
			}

			// Setup video recorder if saving
			var recorder *transport.VideoRecorder
			var err error

			if saveFile != "" {
				// Use existing frame channel from drone instead of creating duplicate listener
				recorder, err = transport.NewVideoRecorderFromChannel(frameChan, saveFile)
				if err != nil {
					return fmt.Errorf("failed to create video recorder: %w", err)
				}

				if err := recorder.StartRecording(); err != nil {
					return fmt.Errorf("failed to start recording: %w", err)
				}
				defer recorder.Close()
			}

			fmt.Printf("Monitoring video stream (%s format)...\n", videoFormat)
			if saveFile != "" {
				fmt.Printf("Video will be saved to: %s\n", saveFile)
			}

			// Setup interrupt handling
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

			// Setup timeout if duration is specified
			timeout := time.After(time.Duration(duration) * time.Second)

			frameCount := 0
			startTime := time.Now()

			// Monitor frames
			for {
				select {
				case frame, ok := <-frameChan:
					if !ok {
						fmt.Println("Video frame channel closed")
						return nil
					}

					frameCount++

					// Log frame info every 30 frames
					if frameCount%30 == 0 {
						elapsed := time.Since(startTime)
						fps := float64(frameCount) / elapsed.Seconds()
						fmt.Printf("Received frame %d: %d bytes, keyframe: %v, %.2f FPS\n",
							frame.SeqNum, frame.Size, frame.IsKeyFrame, fps)
					}

				case <-interrupt:
					fmt.Println("\nReceived interrupt signal, stopping...")
					return nil

				case <-timeout:
					if duration > 0 {
						fmt.Printf("Duration of %d seconds reached, stopping...\n", duration)
						return nil
					}
				}
			}
		},
	}

	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds to monitor the stream (0 = indefinite)")
	cmd.Flags().StringVarP(&saveFile, "save", "s", "", "Save video stream to file")
	cmd.Flags().StringVarP(&format, "format", "f", "h264", "Video format (h264 or mp4)")

	return cmd
}
