package main

import (
	"os"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

const (
	DefaultTelloHost = "192.168.10.1"
)

func main() {
	sdk := tello.NewTelloSDK(DefaultTelloHost)

	drone, err := sdk.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
	}

	drone.TakeOff()
	time.Sleep(5 * time.Second)
	drone.Land()
}
