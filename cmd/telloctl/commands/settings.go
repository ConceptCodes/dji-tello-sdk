package commands

import (
	"fmt"
	"strconv"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/spf13/cobra"
)

func SetSpeedCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "speed [speed]",
		Short: "Set the speed of the drone (cm/s)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			speed, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid speed argument: '%s'. Please provide a number. %w", args[0], err)
			}
			err = drone.SetSpeed(speed)
			if err != nil {
				return fmt.Errorf("set speed command failed: %w", err)
			}
			cmd.Printf("Set Speed command sent with speed %d cm/s.\n", speed)
			return nil
		},
	}
}

func SetWifiCredentialsCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "wifi [SSID] [PASSWORD]",
		Short: "Set the Wi-Fi credentials of the drone",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ssid := args[0]
			password := args[1]
			err := drone.SetWiFiCredentials(ssid, password)
			if err != nil {
				return fmt.Errorf("set wifi command failed: %w", err)
			}
			cmd.Printf("Set Wi-Fi credentials command sent with SSID '%s' and PASSWORD '%s'.\n", ssid, password)
			return nil
		},
	}
}
