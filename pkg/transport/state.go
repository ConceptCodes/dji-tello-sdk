package transport

import (
	"fmt"
	"net"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type StateListener struct {
	server *udp.UDPServer
}

func NewStateListener(listenAddr string) (*StateListener, error) {

	server, err := udp.NewUDPServer(
		listenAddr,
		udp.WithOnData(onStateData),
		udp.WithOnError(onStateError),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server for state listener on %s: %w", listenAddr, err)
	}

	return &StateListener{
		server: server,
	}, nil
}

// Start begins listening for state data.
// This is a blocking call and should typically be run in a goroutine.
func (sl *StateListener) Start() error {
	if sl.server == nil {
		return fmt.Errorf("state listener server is not initialized")
	}
	utils.Logger.Infof("Starting Tello state listener on %s", sl.server.Addr)
	return sl.server.Start()
}

func (sl *StateListener) Stop() {
	if sl.server != nil {
		utils.Logger.Infof("Stopping Tello state listener on %s", sl.server.Addr)
		sl.server.Stop()
	} else {
		utils.Logger.Warnf("Attempted to stop a nil state listener server")
	}
}

func onStateData(data []byte, addr *net.UDPAddr) {
	state, err := utils.ParseState(string(data))
	if err != nil {
		utils.Logger.Warnf("Error parsing state data from %s: %v. Data: %s", addr.String(), err, string(data))
		return
	}
	utils.Logger.Debugf("Received Tello State from %s: %+v", addr.String(), state)
}

func onStateError(err error) {
	utils.Logger.Errorf("State listener UDP server error: %v", err)
}
