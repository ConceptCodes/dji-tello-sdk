package commands

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/spf13/cobra"
)

// GamepadCmd creates the gamepad control command
func GamepadCmd(drone tello.TelloCommander) *cobra.Command {
	var preset string
	var listGamepads bool

	cmd := &cobra.Command{
		Use:   "gamepad",
		Short: "Control the drone using a gamepad",
		Long: `Control the DJI Tello drone using a gamepad controller.
Supports Xbox, PlayStation, and generic USB controllers with customizable button mappings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listGamepads {
				return listAvailableGamepads()
			}

			// Load configuration
			config, err := loadGamepadConfig(preset, cmd.Flags().Changed("preset"))
			if err != nil {
				return fmt.Errorf("failed to load gamepad configuration: %w", err)
			}

			// Initialize drone (redundant now but kept for safety)
			if err := drone.Init(); err != nil {
				return fmt.Errorf("SDK mode handshake failed: %w", err)
			}

			// Create gamepad handler
			handler, err := gamepad.NewHandler(gamepad.HandlerOptions{
				Config: config,
				OnCommand: func(command gamepad.Command) {
					handleDroneCommand(drone, command)
				},
				OnError: func(err error) {
					utils.Logger.Errorf("Gamepad error: %v", err)
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create gamepad handler: %w", err)
			}

			// Start gamepad handler
			if err := handler.Start(); err != nil {
				return fmt.Errorf("failed to start gamepad handler: %w", err)
			}
			defer handler.Stop()

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			cmd.Println("Gamepad control started. Press Ctrl+C to stop.")
			cmd.Println("Use left stick for movement, right stick for altitude/yaw.")
			cmd.Println("Press A for takeoff/land, B for emergency stop.")

			// Wait for shutdown signal
			<-sigChan
			cmd.Println("\nShutting down gamepad control...")

			// Land the drone before exiting
			if err := drone.Land(); err != nil {
				utils.Logger.Errorf("Failed to land drone: %v", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Use preset configuration (default, xbox, playstation)")
	cmd.Flags().BoolVarP(&listGamepads, "list", "l", false, "List available gamepads and exit")

	return cmd
}

// loadGamepadConfig loads the gamepad configuration using auto-discovery when no preset override is supplied.
func loadGamepadConfig(preset string, presetExplicit bool) (*gamepad.Config, error) {
	if !presetExplicit {
		config, source, err := gamepad.LoadAutoConfig()
		if err == nil {
			utils.Logger.Infof("Using gamepad configuration file: %s", source)
			return config, nil
		}
		if err != nil && !errors.Is(err, gamepad.ErrConfigNotFound) {
			return nil, fmt.Errorf("failed to load discovered config: %w", err)
		}
		if errors.Is(err, gamepad.ErrConfigNotFound) {
			utils.Logger.Info("No config.json found locally or globally; falling back to preset")
		}
	} else {
		utils.Logger.Infof("Preset '%s' explicitly selected; skipping auto config lookup", preset)
	}

	// Load preset fallback
	presets := gamepad.GetPresetConfigs()
	config, exists := presets[preset]
	if !exists {
		return nil, fmt.Errorf("unknown preset: %s. Available presets: %v", preset, gamepad.GetConfigNames())
	}

	utils.Logger.Infof("Using preset configuration: %s", preset)
	return config, nil
}

// handleDroneCommand handles drone commands from the gamepad
func handleDroneCommand(drone tello.TelloCommander, command gamepad.Command) {
	switch command.Type {
	case gamepad.CommandRC:
		if rcValues, ok := command.Data.(gamepad.RCValues); ok {
			if err := drone.SetRcControl(rcValues.A, rcValues.B, rcValues.C, rcValues.D); err != nil {
				utils.Logger.Errorf("Failed to set RC control: %v", err)
			}
		} else {
			utils.Logger.Errorf("Invalid RC command data type")
		}

	case gamepad.CommandAction:
		if action, ok := command.Data.(gamepad.DroneAction); ok {
			handleDroneAction(drone, action)
		} else {
			utils.Logger.Errorf("Invalid action command data type")
		}

	default:
		utils.Logger.Warnf("Unknown command type: %s", command.Type)
	}
}

// handleDroneAction handles drone actions triggered by gamepad buttons
func handleDroneAction(drone tello.TelloCommander, action gamepad.DroneAction) {
	switch action {
	case gamepad.ActionTakeoff:
		utils.Logger.Info("Takeoff triggered")
		if err := drone.TakeOff(); err != nil {
			utils.Logger.Errorf("Takeoff failed: %v", err)
		}

	case gamepad.ActionLand:
		utils.Logger.Info("Land triggered")
		if err := drone.Land(); err != nil {
			utils.Logger.Errorf("Land failed: %v", err)
		}

	case gamepad.ActionEmergency:
		utils.Logger.Warn("Emergency stop triggered")
		if err := drone.Emergency(); err != nil {
			utils.Logger.Errorf("Emergency stop failed: %v", err)
		}

	case gamepad.ActionFlipForward:
		utils.Logger.Info("Flip forward triggered")
		if err := drone.Flip(tello.FlipForward); err != nil {
			utils.Logger.Errorf("Flip forward failed: %v", err)
		}

	case gamepad.ActionFlipBackward:
		utils.Logger.Info("Flip backward triggered")
		if err := drone.Flip(tello.FlipBackward); err != nil {
			utils.Logger.Errorf("Flip backward failed: %v", err)
		}

	case gamepad.ActionFlipLeft:
		utils.Logger.Info("Flip left triggered")
		if err := drone.Flip(tello.FlipLeft); err != nil {
			utils.Logger.Errorf("Flip left failed: %v", err)
		}

	case gamepad.ActionFlipRight:
		utils.Logger.Info("Flip right triggered")
		if err := drone.Flip(tello.FlipRight); err != nil {
			utils.Logger.Errorf("Flip right failed: %v", err)
		}

	case gamepad.ActionStreamOn:
		utils.Logger.Info("Toggle video stream triggered")
		// This would need state tracking to toggle properly
		// For now, just turn on the stream
		if err := drone.StreamOn(); err != nil {
			utils.Logger.Errorf("Stream on failed: %v", err)
		}

	default:
		utils.Logger.Warnf("Unknown drone action: %s", action)
	}
}

// listAvailableGamepads lists all connected gamepads
func listAvailableGamepads() error {
	gamepads := gamepad.ListGamepads()

	if len(gamepads) == 0 {
		fmt.Println("No gamepads found.")
		fmt.Println("Make sure your controller is connected and try again.")
		return nil
	}

	fmt.Printf("Found %d gamepad(s):\n", len(gamepads))
	for i, info := range gamepad.GetGamepadInfo() {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, info["name"], info["id"])
	}

	fmt.Println("\nAvailable presets:")
	for _, name := range gamepad.GetConfigNames() {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
