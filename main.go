package main

import (
	"os"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

func main() {
	drone, err := tello.Initialize()
	if err != nil {
		utils.Logger.Errorf("Error initializing Tello SDK: %v", err)
		os.Exit(1)
	}
	defer drone.Shutdown()

	utils.Logger.Info("SDK mode initialized successfully")

	percentage, err := drone.GetBatteryPercentage()
	if err != nil {
		utils.Logger.Errorf("Error getting battery percentage: %v", err)
		os.Exit(1)
	}

	utils.Logger.Infof("Battery percentage: %d%%", percentage)

	// ⚠️ SAFETY WARNING: Uncommenting the code below will cause immediate drone TAKEOFF
	// when running "go run main.go" - only uncomment for testing with a connected drone
	// in a SAFE OPEN AREA, keeping drone at least 5 meters away from people and obstacles
	//
	// if err := drone.TakeOff(); err != nil {
	// 	utils.Logger.Errorf("Error with takeoff: %v", err)
	// 	os.Exit(1)
	// }
	//
	// time.Sleep(3 * time.Second)
	//
	// if err := drone.Land(); err != nil {
	// 	utils.Logger.Errorf("Error with landing the drone: %v", err)
	// 	os.Exit(1)
	// }
}
