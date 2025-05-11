package telloctl

import (
	"os"
	"github.com/spf13/cobra"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

const (
	DefaultTelloHost = "192.168.10.1"
)

var drone tello.TelloCommander

func main() {
	sdk := tello.NewTelloSDK(DefaultTelloHost)
	var err error
	drone, err = sdk.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
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
