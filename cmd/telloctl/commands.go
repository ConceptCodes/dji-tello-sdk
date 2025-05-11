package telloctl

import (
	"strconv"

	"github.com/spf13/cobra"
)

var height = 50 // Default height in cm
var err error

func newTakeoffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "takeoff",
		Short: "Make the drone take off",
		Run: func(cmd *cobra.Command, args []string) {
			sdk.TakeOff()
		},
	}
}

func newLandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "land",
		Short: "Make the drone land",
		Run: func(cmd *cobra.Command, args []string) {
			sdk.Land()
		},
	}
}

func newUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Make the drone go up",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				height, err = strconv.Atoi(args[0])
				if err != nil {
					cmd.Println("Invalid height argument. Using default height of 50 cm.")
				}
			}
			sdk.GoUp(height)
		},
	}
}

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Make the drone go down",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				height, err = strconv.Atoi(args[0])
				if err != nil {
					cmd.Println("Invalid height argument. Using default height of 50 cm.")
				}
			}
			sdk.GoDown(height)
		},
	}
}
