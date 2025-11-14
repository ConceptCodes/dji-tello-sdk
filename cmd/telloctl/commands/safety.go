package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/spf13/cobra"
)

// SafetyCmd represents the safety command
var SafetyCmd = &cobra.Command{
	Use:   "safety",
	Short: "Manage safety configurations and settings",
	Long: `Manage safety configurations, presets, and monitoring settings for the drone.
This command allows you to view, validate, and manage safety configurations.`,
}

// safetyListCmd lists available safety presets and configurations
var safetyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available safety presets and configurations",
	Long:  `List all available safety presets and configuration files.`,
	Run: func(cmd *cobra.Command, args []string) {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PRESET\tDESCRIPTION\tLEVEL")
		fmt.Fprintln(w, "-----\t-----------\t-----")

		presets := safety.GetPresetConfigs()
		for name, config := range presets {
			fmt.Fprintf(w, "%s\t%s safety\t%s\n", name, config.Level, config.Level)
		}

		w.Flush()

		fmt.Println("\nConfiguration Files:")
		configFiles := []string{
			"safety-default.json",
			"safety-conservative.json",
			"safety-aggressive.json",
			"safety-indoor.json",
			"safety-outdoor.json",
		}

		for _, file := range configFiles {
			if safety.ConfigExists(file) {
				fmt.Printf("  ✓ %s\n", file)
			} else {
				fmt.Printf("  ✗ %s (not found)\n", file)
			}
		}
	},
}

// safetyValidateCmd validates a safety configuration file
var safetyValidateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate a safety configuration file",
	Long:  `Validate a safety configuration file against the JSON schema.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]

		err := safety.ValidateConfigFile(configFile)
		if err != nil {
			utils.Logger.Errorf("Configuration validation failed: %v", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Configuration file '%s' is valid\n", configFile)
	},
}

// safetyShowCmd shows details of a safety preset or configuration
var safetyShowCmd = &cobra.Command{
	Use:   "show [preset|config-file]",
	Short: "Show details of a safety preset or configuration",
	Long:  `Display detailed information about a safety preset or configuration file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Try to load as preset first
		config, err := safety.LoadPresetConfig(name)
		if err != nil {
			// Try to load as config file
			config, err = safety.LoadConfigFromFile(name)
			if err != nil {
				utils.Logger.Errorf("Failed to load safety preset or config '%s': %v", name, err)
				os.Exit(1)
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintf(w, "Name:\t%s\n", name)
		fmt.Fprintf(w, "Version:\t%s\n", config.Version)
		fmt.Fprintf(w, "Level:\t%s\n", config.Level)
		fmt.Fprintf(w, "Safety Level:\t%s\n", config.Level)
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "ALTITUDE LIMITS")
		fmt.Fprintf(w, "  Min Height:\t%d cm\n", config.Altitude.MinHeight)
		fmt.Fprintf(w, "  Max Height:\t%d cm\n", config.Altitude.MaxHeight)
		fmt.Fprintf(w, "  Takeoff Height:\t%d cm\n", config.Altitude.TakeoffHeight)
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "VELOCITY LIMITS")
		fmt.Fprintf(w, "  Max Horizontal:\t%d cm/s\n", config.Velocity.MaxHorizontal)
		fmt.Fprintf(w, "  Max Vertical:\t%d cm/s\n", config.Velocity.MaxVertical)
		fmt.Fprintf(w, "  Max Yaw:\t%d deg/s\n", config.Velocity.MaxYaw)
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "BATTERY THRESHOLDS")
		fmt.Fprintf(w, "  Warning:\t%d%%\n", config.Battery.WarningThreshold)
		fmt.Fprintf(w, "  Critical:\t%d%%\n", config.Battery.CriticalThreshold)
		fmt.Fprintf(w, "  Emergency:\t%d%%\n", config.Battery.EmergencyThreshold)
		fmt.Fprintf(w, "  Auto Land:\t%t\n", config.Battery.EnableAutoLand)
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "SENSOR LIMITS")
		fmt.Fprintf(w, "  Min TOF Distance:\t%d cm\n", config.Sensors.MinTOFDistance)
		fmt.Fprintf(w, "  Max Tilt Angle:\t%d deg\n", config.Sensors.MaxTiltAngle)
		fmt.Fprintf(w, "  Max Acceleration:\t%.1f g\n", config.Sensors.MaxAcceleration)
		fmt.Fprintln(w, "")

		fmt.Fprintln(w, "BEHAVIORAL LIMITS")
		fmt.Fprintf(w, "  Enable Flips:\t%t\n", config.Behavioral.EnableFlips)
		fmt.Fprintf(w, "  Min Flip Height:\t%d cm\n", config.Behavioral.MinFlipHeight)
		fmt.Fprintf(w, "  Max Flight Time:\t%d seconds\n", config.Behavioral.MaxFlightTime)
		fmt.Fprintf(w, "  Max Command Rate:\t%d cmd/s\n", config.Behavioral.MaxCommandRate)

		w.Flush()
	},
}

func init() {
	SafetyCmd.AddCommand(safetyListCmd)
	SafetyCmd.AddCommand(safetyValidateCmd)
	SafetyCmd.AddCommand(safetyShowCmd)
}
