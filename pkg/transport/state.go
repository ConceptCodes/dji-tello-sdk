package transport

import (
	"context"
	"fmt"
	"net"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type StateListener struct {
	server    *udp.UDPServer
	stateChan chan *types.State
}

func NewStateListener(listenAddr string) (*StateListener, error) {
	sl := &StateListener{
		stateChan: make(chan *types.State, 100), // Buffer for 100 states
	}

	server, err := udp.NewUDPServer(
		listenAddr,
		udp.WithOnData(func(data []byte, addr *net.UDPAddr) {
			sl.onStateData(data, addr)
		}),
		udp.WithOnError(onStateError),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server for state listener on %s: %w", listenAddr, err)
	}

	sl.server = server
	return sl, nil
}

// Start begins listening for state data.
// This is a blocking call and should typically be run in a goroutine.
func (sl *StateListener) Start() error {
	if sl.server == nil {
		return fmt.Errorf("state listener server is not initialized")
	}
	utils.Logger.Infof("Starting Tello state listener on %s", sl.server.Addr)
	return sl.server.Start(context.Background())
}

func (sl *StateListener) Stop() {
	if sl.server != nil {
		utils.Logger.Infof("Stopping Tello state listener on %s", sl.server.Addr)
		sl.server.Stop()
	} else {
		utils.Logger.Warnf("Attempted to stop a nil state listener server")
	}

	// Close state channel
	if sl.stateChan != nil {
		close(sl.stateChan)
		sl.stateChan = nil
	}
}

// GetStateChannel returns a read-only channel for receiving telemetry states
func (sl *StateListener) GetStateChannel() <-chan *types.State {
	return sl.stateChan
}

func (sl *StateListener) onStateData(data []byte, addr *net.UDPAddr) {
	state, err := utils.ParseState(string(data))
	if err != nil {
		utils.Logger.Warnf("Error parsing state data from %s: %v. Data: %s", addr.String(), err, string(data))
		return
	}

	// Send state to channel (non-blocking)
	select {
	case sl.stateChan <- state:
		utils.Logger.Debugf("State sent to channel from %s: %+v", addr.String(), state)
	default:
		utils.Logger.Warnf("State channel full, dropping state from %s", addr.String())
	}
}

func onStateError(err error) {
	utils.Logger.Errorf("State listener UDP server error: %v", err)
}
