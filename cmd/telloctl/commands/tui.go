package commands

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/gamepad"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ui"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
)

// TuiCmd creates the TUI command
func TuiCmd(drone tello.TelloCommander) *cobra.Command {
	var preset string

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Start interactive terminal user interface",
		Long: `Start a rich interactive terminal user interface (TUI) for controlling the drone.
		
Features:
- Real-time telemetry dashboard
- Keyboard flight controls (WASD)
- Gamepad support
- Command REPL
- Mission logs`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Ensure SDL runs on main thread for gamepad support
			var runErr error
			sdl.Main(func() {
				runErr = runTui(drone, preset)
			})
			return runErr
		},
	}

	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Gamepad mapping preset (default, xbox, playstation)")

	return cmd
}

func runTui(drone tello.TelloCommander, preset string) error {
	// Create TUI model
	model := ui.NewTuiModel(drone)

	// Start Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Initialize gamepad
	config, err := loadGamepadConfig(preset, preset != "default")
	if err == nil {
		handler, err := gamepad.NewHandler(gamepad.HandlerOptions{
			Config: config,
			OnCommand: func(command gamepad.Command) {
				// Handle drone command
				handleDroneCommand(drone, command)

				// Send log message to TUI
				var msg string
				switch command.Type {
				case gamepad.CommandAction:
					if action, ok := command.Data.(gamepad.DroneAction); ok {
						msg = fmt.Sprintf("Gamepad: %s", action)
					}
				case gamepad.CommandRC:
					// Don't log every RC update, too spammy
					return
				}

				if msg != "" {
					p.Send(ui.GamepadMsg{Message: msg})
				}
			},
			OnError: func(err error) {
				utils.Logger.Errorf("Gamepad error: %v", err)
			},
		})

		if err == nil {
			if err := handler.Start(); err == nil {
				defer handler.Stop()

				// Start gamepad polling loop in a goroutine?
				// No, SDL polling must be on main thread.
				// But p.Run() blocks.
				// So we run p.Run() in a goroutine and poll SDL in main thread.

				done := make(chan struct{})
				go func() {
					if _, err := p.Run(); err != nil {
						fmt.Printf("Error running TUI: %v\n", err)
					}
					close(done)
				}()

				// Main loop for SDL polling
				ticker := time.NewTicker(time.Millisecond * 20) // 50Hz
				defer ticker.Stop()

				for {
					select {
					case <-done:
						return nil
					case <-ticker.C:
						handler.ProcessEvents()
					}
				}
			}
		}
	}

	// Fallback if gamepad init failed or not used: just run TUI normally
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
