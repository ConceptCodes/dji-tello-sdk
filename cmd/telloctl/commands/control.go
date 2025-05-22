package commands

import (
	"fmt"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/spf13/cobra"
)

func TakeOffCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "takeoff",
		Short: "Make the drone take off",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := drone.TakeOff()
			if err != nil {
				return fmt.Errorf("takeoff failed: %w", err)
			}
			cmd.Println("Takeoff command sent.")
			return nil
		},
	}
}

func LandCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "land",
		Short: "Make the drone land",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := drone.Land()
			if err != nil {
				return fmt.Errorf("land failed: %w", err)
			}
			cmd.Println("Land command sent.")
			return nil
		},
	}
}

func EmergencyCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "emergency",
		Short: "Stop all motors immediately",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := drone.Emergency()
			if err != nil {
				return fmt.Errorf("emergency command failed: %w", err)
			}
			cmd.Println("Emergency command sent.")
			return nil
		},
	}
}
