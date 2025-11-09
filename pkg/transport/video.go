package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport/udp"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type VideoFrame struct {
	Data       []byte
	Timestamp  time.Time
	Size       int
	SeqNum     int
	NALUnits   []NALUnit
	IsKeyFrame bool
}

type VideoStreamListener struct {
	server    *udp.UDPServer
	FrameChan chan VideoFrame
	seqNum    int
}

func NewVideoStreamListener(listenAddr string) (*VideoStreamListener, error) {
	vsl := &VideoStreamListener{
		FrameChan: make(chan VideoFrame, 100), // Buffer for 100 frames
		seqNum:    0,
	}

	server, err := udp.NewUDPServer(
		listenAddr,
		udp.WithOnData(vsl.onVideoStreamData),
		udp.WithOnError(onVideoStreamError),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP server for video stream listener on %s: %w", listenAddr, err)
	}

	vsl.server = server
	return vsl, nil
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

	// Close frame channel
	if vsl.FrameChan != nil {
		close(vsl.FrameChan)
		vsl.FrameChan = nil
	}
}

// GetFrameChannel returns a read-only channel for receiving video frames
func (vsl *VideoStreamListener) GetFrameChannel() <-chan VideoFrame {
	return vsl.FrameChan
}

func (vsl *VideoStreamListener) onVideoStreamData(data []byte, addr *net.UDPAddr) {
	utils.Logger.Debugf("Received %d bytes of video data from %s", len(data), addr.String())

	// Parse H.264 data
	parser := NewH264Parser()
	nalUnits, err := parser.ParseFrame(data)
	if err != nil {
		utils.Logger.Errorf("Failed to parse H.264 frame: %v", err)
		// Still send the frame even if parsing fails
		nalUnits = []NALUnit{}
	}

	// Create video frame with metadata
	frame := VideoFrame{
		Data:       make([]byte, len(data)),
		Timestamp:  time.Now(),
		Size:       len(data),
		SeqNum:     vsl.seqNum,
		NALUnits:   nalUnits,
		IsKeyFrame: parser.HasKeyFrame(nalUnits),
	}

	// Copy data to avoid race conditions
	copy(frame.Data, data)
	vsl.seqNum++

	// Send frame to channel (non-blocking to prevent UDP listener blocking)
	select {
	case vsl.FrameChan <- frame:
		utils.Logger.Debugf("Frame %d sent to channel (%d bytes, %d NAL units, keyframe: %v)",
			frame.SeqNum, frame.Size, len(frame.NALUnits), frame.IsKeyFrame)
	default:
		utils.Logger.Warnf("Frame channel full, dropping frame %d", frame.SeqNum)
	}
}

// ToEnhancedFrame converts VideoFrame to EnhancedVideoFrame for ML processing
func (vf *VideoFrame) ToEnhancedFrame() *ml.EnhancedVideoFrame {
	enhanced := ml.NewEnhancedVideoFrame(vf.Data, vf.Timestamp, vf.SeqNum)
	enhanced.IsKeyFrame = vf.IsKeyFrame
	return enhanced
}

func onVideoStreamError(err error) {
	utils.Logger.Errorf("Video stream listener UDP server error: %v", err)
	// TODO: Consider if any specific errors need more handling (e.g., re-establishing connection if possible).
}
