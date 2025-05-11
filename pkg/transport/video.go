package transport

import (
	"fmt"
	"net"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type VideoStreamListener struct {
	server *udp.UDPServer
}

func NewVideoStreamListener(listenAddr string) (*VideoStreamListener, error) {
	server, err := udp.NewUDPServer(
		listenAddr,
		udp.WithOnData(onVideoStreamData),
		udp.WithOnError(onVideoStreamError),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server for video stream listener on %s: %w", listenAddr, err)
	}

	return &VideoStreamListener{
		server: server,
	}, nil
}

func (vsl *VideoStreamListener) Start() error {
	if vsl.server == nil {
		return fmt.Errorf("video stream listener server is not initialized")
	}
	utils.Logger.Infof("Starting Tello video stream listener on %s", vsl.server.Addr)
	return vsl.server.Start()
}

func (vsl *VideoStreamListener) Stop() {
	if vsl.server != nil {
		utils.Logger.Infof("Stopping Tello video stream listener on %s", vsl.server.Addr)
		vsl.server.Stop()
	} else {
		utils.Logger.Warnf("Attempted to stop a nil video stream listener server")
	}
}

func onVideoStreamData(data []byte, addr *net.UDPAddr) {
	utils.Logger.Debugf("Received %d bytes of video data from %s. (H.264 frame data - not parsed)", len(data), addr.String())

	// TODO: Implement H.264 frame processing (e.g., send to a decoder or save to file).
	// Example: videoFrameChannel <- dataCopy (if you have a channel for frames)
}

func onVideoStreamError(err error) {
	utils.Logger.Errorf("Video stream listener UDP server error: %v", err)
	// TODO: Consider if any specific errors need more handling (e.g., re-establishing connection if possible).
}
