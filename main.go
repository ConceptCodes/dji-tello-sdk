package main

import (
	"os"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

func main() {
	drone, err := tello.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
	}

	utils.Logger.Info("SDK mode initialized successfully")

	if err := drone.TakeOff(); err != nil {
		utils.Logger.Errorf("Error with takeoff: %v", err)
		os.Exit(1)
	}

	time.Sleep(3 * time.Second)

	if err := drone.Land(); err != nil {
		utils.Logger.Errorf("Error with landing the drone: %v", err)
		os.Exit(1)
	}
}
