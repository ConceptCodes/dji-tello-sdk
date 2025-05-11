package telloctl

import (
	"os"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	DefaultTelloHost = "192.168.10.1"
)

var drone tello.TelloCommander

func main() {
	sdk := tello.NewTello(DefaultTelloHost)
	var err error
	drone, err = sdk.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
	}

	if err := drone.Init(); err != nil {
		utils.Logger.Errorf("Error sending initial 'command' to Tello: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "telloctl",
		Short: "CLI for controlling the DJI Tello drone",
	}

	rootCmd.AddCommand(
		newTakeoffCmd(),
		newLandCmd(),
		newUpCmd(),
		newDownCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
