package telloctl

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newTakeoffCmd() *cobra.Command {
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

func newLandCmd() *cobra.Command {
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

func newUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [distance]",
		Short: "Make the drone go up by a specific distance (cm)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			distance := 50
			var parseErr error
			if len(args) > 0 {
				distance, parseErr = strconv.Atoi(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid distance argument: '%s'. Please provide a number. %w", args[0], parseErr)
				}
			}
			err := drone.Up(distance)
			if err != nil {
				return fmt.Errorf("up command failed: %w", err)
			}
			cmd.Printf("Up command sent for %d cm.\n", distance)
			return nil
		},
	}
}

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down [distance]",
		Short: "Make the drone go down by a specific distance (cm)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			distance := 50
			var parseErr error
			if len(args) > 0 {
				distance, parseErr = strconv.Atoi(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid distance argument: '%s'. Please provide a number. %w", args[0], parseErr)
				}
			}
			err := drone.Down(distance)
			if err != nil {
				return fmt.Errorf("down command failed: %w", err)
			}
			cmd.Printf("Down command sent for %d cm.\n", distance)
			return nil
		},
	}
}
