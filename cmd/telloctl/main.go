package telloctl

import (
	"os"

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

func main() {
	drone, err := tello.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "telloctl",
		Short: "CLI for controlling the DJI Tello drone",
	}

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
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
