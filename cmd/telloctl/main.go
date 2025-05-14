package telloctl

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/conceptcodes/dji-tello-sdk-go/cmd/telloctl/commands"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

const (
	DefaultTelloHost = "192.168.10.1"
)

func newGetCmd(drone tello.TelloCommander) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve various telemetry data from the drone",
	}

	getCmd.AddCommand(
		commands.NewGetBatteryCmd(drone),
		commands.NewGetHeightCmd(drone),
		commands.NewGetSpeedCmd(drone),
		commands.NewGetTimeCmd(drone),
		commands.NewGetTemperatureCmd(drone),
		commands.NewGetAttitudeCmd(drone),
		commands.NewGetBarometerCmd(drone),
		commands.NewGetAccelerationCmd(drone),
		commands.NewGetTofCmd(drone),
	)

	return getCmd
}

func newSetCmd(drone tello.TelloCommander) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set various parameters for the drone",
	}

	setCmd.AddCommand(
		commands.NewSetSpeedCmd(drone),
		commands.NewSetWifiCredentialsCmd(drone),
	)

	return setCmd
}

func main() {
	sdk := tello.NewTelloSDK(DefaultTelloHost)
	drone, err := sdk.Initialize()
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
		commands.NewTakeOffCmd(drone),
		commands.NewLandCmd(drone),
		commands.NewEmergencyCmd(drone),
		commands.NewUpCmd(drone),
		commands.NewDownCmd(drone),
		commands.NewFlipCmd(drone),
		commands.NewBackwardCmd(drone),
		commands.NewForwardCmd(drone),
		commands.NewLeftCmd(drone),
		commands.NewRightCmd(drone),
		commands.NewClockwiseCmd(drone),
		commands.NewCounterClockwiseCmd(drone),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
