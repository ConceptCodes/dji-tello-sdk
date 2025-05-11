package tello

import (
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

const (
	DefaultStateListenAddr = "0.0.0.0:8890"
	DefaultVideoListenAddr = "0.0.0.0:11111"
)

type TelloSDK struct {
	Host string
}

func NewTello(host string) *TelloSDK {
	return &TelloSDK{
		Host: host,
	}
}

func (t *TelloSDK) Initialize() (TelloCommander, error) {
	commandQueue := NewCommandQueue()
	commandConnection, err := transport.NewCommandConnection(t.Host, 8889)
	if err != nil {
		return nil, err
	}

	stateListener, err := transport.NewStateListener(DefaultStateListenAddr)
	if err != nil {
		return nil, err
	}

	go stateListener.Start()

	videoStreamListener, err := transport.NewVideoStreamListener(DefaultVideoListenAddr)
	if err != nil {
		return nil, err
	}

	go videoStreamListener.Start()

	commander := NewTelloCommander(commandConnection, commandQueue, stateListener, videoStreamListener)

	if err := commander.Init(); err != nil {
		utils.Logger.Errorf("Error sending initial 'command' to Tello: %v", err)
	}

	return commander, nil
}
