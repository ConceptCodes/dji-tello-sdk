package telloctl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	cmdPort   = 8889
	statePort = 8890
	timeout   = 5 * time.Second
)

var sdk *tello.TelloCommanderConfig

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sdk, err := initializeSDK(ctx)
	if err != nil {
		utils.Logger.Errorf("Failed to initialize SDK: %v", err)
		os.Exit(1)
	}
	defer sdk.Stop()

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
		utils.Logger.Errorf("Error: %v", err)
		os.Exit(1)
	}
}

func initializeSDK(ctx context.Context) (*tello.TelloCommanderConfig, error) {
	stateConn, err := transport.NewConn(ctx, statePort)
	if err != nil {
		utils.Logger.Errorf("failed to create state connection: %v", err)
		return nil, fmt.Errorf("failed to create state connection: %w", err)
	}

	commandChannel, err := transport.NewCommandConn(ctx, cmdPort, timeout)
	if err != nil {
		utils.Logger.Errorf("failed to create command connection: %v", err)
		return nil, fmt.Errorf("failed to create command connection: %w", err)
	}

	stateStream := transport.NewStateStream(ctx, stateConn)
	return tello.InitializeSDK(ctx, *commandChannel, stateStream, nil), nil
}
