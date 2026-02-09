package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/conceptcodes/dji-tello-sdk-go/cmd/telloctl/commands"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

func newGetCmd(drone tello.TelloCommander) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve various telemetry data from the drone",
	}

	getCmd.AddCommand(
		commands.GetBatteryCmd(drone),
		commands.GetHeightCmd(drone),
		commands.GetSpeedCmd(drone),
		commands.GetTimeCmd(drone),
		commands.GetTemperatureCmd(drone),
		commands.GetAttitudeCmd(drone),
		commands.GetBarometerCmd(drone),
		commands.GetAccelerationCmd(drone),
		commands.GetTofCmd(drone),
	)

	return getCmd
}

func newSetCmd(drone tello.TelloCommander) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set various parameters for the drone",
	}

	setCmd.AddCommand(
		commands.SetSpeedCmd(drone),
		commands.SetWifiCredentialsCmd(drone),
	)

	return setCmd
}

func isWebCommand(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return arg == "web"
	}
	return false
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "telloctl",
		Short: "CLI for controlling the DJI Tello drone",
	}

	// Create drone commander with default configuration
	drone, err := tello.InitializeWithOptions()
	if err != nil {
		if isWebCommand(os.Args[1:]) {
			utils.Logger.Warnf("Web interface starting without drone connection: %v", err)
			drone = nil
		} else {
			utils.Logger.Errorf("Error initializing Tello commander: %v", err)
			os.Exit(1)
		}
	}

	// Add shutdown handling
	if drone != nil {
		defer func() {
			if err := drone.Shutdown(); err != nil {
				utils.Logger.Errorf("Error shutting down drone: %v", err)
			}
		}()
	}

	// Add all commands
	rootCmd.AddCommand(
		newGetCmd(drone),
		newSetCmd(drone),
		commands.TakeOffCmd(drone),
		commands.LandCmd(drone),
		commands.EmergencyCmd(drone),
		commands.UpCmd(drone),
		commands.DownCmd(drone),
		commands.FlipCmd(drone),
		commands.BackwardCmd(drone),
		commands.ForwardCmd(drone),
		commands.LeftCmd(drone),
		commands.RightCmd(drone),
		commands.ClockwiseCmd(drone),
		commands.CounterClockwiseCmd(drone),
		commands.StreamOnCmd(drone),
		commands.StreamOffCmd(drone),
		commands.StreamCmd(drone),
		commands.VideoGUICmd(drone),
		commands.WebCmd(drone),
		commands.GamepadCmd(drone),
		commands.MLCmd(),
		commands.SafetyCmd,
		commands.TuiCmd(drone),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
