package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/conceptcodes/dji-tello-sdk-go/cmd/telloctl/commands"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
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
	rootCmd := &cobra.Command{
		Use:   "telloctl",
		Short: "CLI for controlling the DJI Tello drone",
	}

	// Create drone commander (without initialization)
	commandClient, err := transport.NewCommandConnection()
	if err != nil {
		utils.Logger.Errorf("Error creating command connection: %v", err)
		os.Exit(1)
	}

	commandQueue := tello.NewPriorityCommandQueue()
	stateListener, err := transport.NewStateListener(":8890")
	if err != nil {
		utils.Logger.Errorf("Error creating state listener: %v", err)
		os.Exit(1)
	}

	videoStreamListener, err := transport.NewVideoStreamListener(":11111")
	if err != nil {
		utils.Logger.Errorf("Error creating video stream listener: %v", err)
		os.Exit(1)
	}

	drone := tello.NewTelloCommander(
		commandClient,
		commandQueue,
		stateListener,
		videoStreamListener,
	)
	if err != nil {
		utils.Logger.Errorf("Error creating Tello commander: %v", err)
		os.Exit(1)
	}

	// Add shutdown handling
	defer func() {
		if err := drone.Shutdown(); err != nil {
			utils.Logger.Errorf("Error shutting down drone: %v", err)
		}
	}()

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
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
