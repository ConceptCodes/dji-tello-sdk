package tello

import (
	"fmt"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

const (
	DefaultStateListenAddr = "0.0.0.0:8890"
	DefaultVideoListenAddr = "0.0.0.0:11111"
)

func Initialize() (TelloCommander, error) {
	commandQueue := NewCommandQueue()
	commandConnection, err := transport.NewCommandConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize command connection: %w", err)
	}

	stateListener, err := transport.NewStateListener(DefaultStateListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize state listener: %w", err)
	}

	// go stateListener.Start()

	videoStreamListener, err := transport.NewVideoStreamListener(DefaultVideoListenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize video stream listener: %w", err)
	}

	// go videoStreamListener.Start()

	commander := NewTelloCommander(commandConnection, commandQueue, stateListener, videoStreamListener)

	if err := commander.Init(); err != nil {
		return nil, fmt.Errorf("failed to send the initial 'command' to Tello: %w", err)
	}

	return commander, nil
}
