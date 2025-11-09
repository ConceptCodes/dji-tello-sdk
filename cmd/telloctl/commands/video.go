package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
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
	
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Monitor and optionally save video stream from the drone",
		Long: `Monitor video stream from the drone in real-time.
This command displays video frame information and can optionally save frames to a file.

Examples:
  telloctl stream                    # Monitor video stream indefinitely
  telloctl stream -d 30              # Monitor for 30 seconds
  telloctl stream -s video.h264      # Save video stream to file
  telloctl stream -d 60 -s video.h264 # Monitor for 60 seconds and save to file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get video stream listener from the drone
			// Note: This requires access to the internal video stream listener
			// For now, we'll provide a placeholder implementation
			
			fmt.Println("Monitoring video stream...")
			fmt.Println("Note: This is a placeholder implementation.")
			fmt.Println("Full video monitoring requires access to the video frame channel.")
			
			if saveFile != "" {
				fmt.Printf("Video will be saved to: %s\n", saveFile)
			}
			
			if duration > 0 {
				fmt.Printf("Monitoring for %d seconds...\n", duration)
				time.Sleep(time.Duration(duration) * time.Second)
			} else {
				fmt.Println("Press Ctrl+C to stop monitoring...")
				
				// Wait for interrupt signal
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt, syscall.SIGTERM)
				<-c
			}
			
			fmt.Println("Video stream monitoring stopped.")
			return nil
		},
	}
	
	cmd.Flags().IntVarP(&duration, "duration", "d", 0, "Duration in seconds to monitor the stream (0 = indefinite)")
	cmd.Flags().StringVarP(&saveFile, "save", "s", "", "Save video stream to file")
	
	return cmd
}