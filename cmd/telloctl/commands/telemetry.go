package commands

import (
	"fmt"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/spf13/cobra"
)

func GetBatteryCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "battery",
		Short: "Get the current battery percentage of the drone",
		RunE: func(cmd *cobra.Command, args []string) error {
			battery, err := drone.GetBatteryPercentage()
			if err != nil {
				return fmt.Errorf("get battery command failed: %w", err)
			}
			cmd.Printf("Battery percentage: %d%%\n", battery)
			return nil
		},
	}
}

func GetHeightCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "height",
		Short: "Get the current height of the drone (cm)",
		RunE: func(cmd *cobra.Command, args []string) error {
			height, err := drone.GetHeight()
			if err != nil {
				return fmt.Errorf("get height command failed: %w", err)
			}
			cmd.Printf("Height: %d cm\n", height)
			return nil
		},
	}
}

func GetSpeedCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "speed",
		Short: "Get the current speed of the drone (cm/s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			speed, err := drone.GetSpeed()
			if err != nil {
				return fmt.Errorf("get speed command failed: %w", err)
			}
			cmd.Printf("Speed: %d cm/s\n", speed)
			return nil
		},
	}
}

func GetTimeCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "time",
		Short: "Get the current flight time of the drone (seconds)",
		RunE: func(cmd *cobra.Command, args []string) error {
			time, err := drone.GetTime()
			if err != nil {
				return fmt.Errorf("get time command failed: %w", err)
			}
			cmd.Printf("Flight time: %d seconds\n", time)
			return nil
		},
	}
}

func GetTemperatureCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "temperature",
		Short: "Get the current temperature of the drone (°C)",
		RunE: func(cmd *cobra.Command, args []string) error {
			temp, err := drone.GetTemperature()
			if err != nil {
				return fmt.Errorf("get temperature command failed: %w", err)
			}
			cmd.Printf("Temperature: %d °C\n", temp)
			return nil
		},
	}
}

func GetAttitudeCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "attitude",
		Short: "Get the current attitude of the drone (pitch, roll, yaw)",
		RunE: func(cmd *cobra.Command, args []string) error {
			pitch, roll, yaw, err := drone.GetAttitude()
			if err != nil {
				return fmt.Errorf("get attitude command failed: %w", err)
			}
			cmd.Printf("Attitude - Pitch: %d, Roll: %d, Yaw: %d\n", pitch, roll, yaw)
			return nil
		},
	}
}

func GetBarometerCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "barometer",
		Short: "Get the current barometer reading of the drone (m)",
		RunE: func(cmd *cobra.Command, args []string) error {
			barometer, err := drone.GetBarometer()
			if err != nil {
				return fmt.Errorf("get barometer command failed: %w", err)
			}
			cmd.Printf("Barometer: %d m\n", barometer)
			return nil
		},
	}
}

func GetAccelerationCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "acceleration",
		Short: "Get the current acceleration of the drone (x, y, z)",
		RunE: func(cmd *cobra.Command, args []string) error {
			x, y, z, err := drone.GetAcceleration()
			if err != nil {
				return fmt.Errorf("get acceleration command failed: %w", err)
			}
			cmd.Printf("Acceleration - X: %d, Y: %d, Z: %d\n", x, y, z)
			return nil
		},
	}
}

func GetTofCmd(drone tello.TelloCommander) *cobra.Command {
	return &cobra.Command{
		Use:   "tof",
		Short: "Get the distance value from time of flight of the drone (cm)",
		RunE: func(cmd *cobra.Command, args []string) error {
			tof, err := drone.GetTof()
			if err != nil {
				return fmt.Errorf("get tof command failed: %w", err)
			}
			cmd.Printf("Time of Flight distance: %d cm\n", tof)
			return nil
		},
	}
}
