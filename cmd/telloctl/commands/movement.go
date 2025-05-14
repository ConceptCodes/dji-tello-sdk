package commands

import (
	"fmt"
	"strconv"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/spf13/cobra"
)

func NewUpCmd(drone tello.TelloCommander) *cobra.Command {
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

func NewDownCmd(drone tello.TelloCommander) *cobra.Command {
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

func NewLeftCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "left [distance]",
		Short: "Make the drone go left by a specific distance (cm)",
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
			err := drone.Left(distance)
			if err != nil {
				return fmt.Errorf("left command failed: %w", err)
			}
			cmd.Printf("Left command sent for %d cm.\n", distance)
			return nil
		},
	}
}

func NewRightCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "right [distance]",
		Short: "Make the drone go right by a specific distance (cm)",
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
			err := drone.Right(distance)
			if err != nil {
				return fmt.Errorf("right command failed: %w", err)
			}
			cmd.Printf("Right command sent for %d cm.\n", distance)
			return nil
		},
	}
}

func NewForwardCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "forward [distance]",
		Short: "Make the drone go forward by a specific distance (cm)",
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
			err := drone.Forward(distance)
			if err != nil {
				return fmt.Errorf("forward command failed: %w", err)
			}
			cmd.Printf("Forward command sent for %d cm.\n", distance)
			return nil
		},
	}
}

func NewBackwardCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "backward [distance]",
		Short: "Make the drone go backward by a specific distance (cm)",
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
			err := drone.Backward(distance)
			if err != nil {
				return fmt.Errorf("backward command failed: %w", err)
			}
			cmd.Printf("Backward command sent for %d cm.\n", distance)
			return nil
		},
	}
}

func NewClockwiseCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "clockwise [angle]",
		Short: "Rotate the drone clockwise by a specific angle (degrees)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			angle := 90
			var parseErr error
			if len(args) > 0 {
				angle, parseErr = strconv.Atoi(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid angle argument: '%s'. Please provide a number. %w", args[0], parseErr)
				}
			}
			err := drone.Clockwise(angle)
			if err != nil {
				return fmt.Errorf("clockwise command failed: %w", err)
			}
			cmd.Printf("Clockwise command sent for %d degrees.\n", angle)
			return nil
		},
	}
}

func NewCounterClockwiseCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "counterclockwise [angle]",
		Short: "Rotate the drone counterclockwise by a specific angle (degrees)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			angle := 90
			var parseErr error
			if len(args) > 0 {
				angle, parseErr = strconv.Atoi(args[0])
				if parseErr != nil {
					return fmt.Errorf("invalid angle argument: '%s'. Please provide a number. %w", args[0], parseErr)
				}
			}
			err := drone.CounterClockwise(angle)
			if err != nil {
				return fmt.Errorf("counterclockwise command failed: %w", err)
			}
			cmd.Printf("Counterclockwise command sent for %d degrees.\n", angle)
			return nil
		},
	}
}

func NewFlipCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "flip [direction]",
		Short: "Make the drone flip in a specific direction (l/r/f/b)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			direction := args[0]
			if direction != "l" && direction != "r" && direction != "f" && direction != "b" {
				return fmt.Errorf("invalid direction argument: '%s'. Please use one of l, r, f, b", direction)
			}
			err := drone.Flip(tello.FlipDirection(direction))
			if err != nil {
				return fmt.Errorf("flip command failed: %w", err)
			}
			cmd.Printf("Flip command sent for direction '%s'.\n", direction)
			return nil
		},
	}
}
