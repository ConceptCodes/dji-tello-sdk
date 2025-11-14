package transport

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"net"
	"sync"
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

const (
	defaultVideoFrameWidth  = 960
	defaultVideoFrameHeight = 720
	colorBlockSize          = 8
)

var (
	parserPool = sync.Pool{
		New: func() interface{} {
			return NewH264Parser()
		},
	}

	bufferPool = sync.Pool{
		New: func() interface{} {
			return image.NewRGBA(image.Rect(0, 0, defaultVideoFrameWidth, defaultVideoFrameHeight))
		},
	}
)

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

// Close gracefully shuts down the video stream listener and cleans up resources
func (vsl *VideoStreamListener) Close() error {
	vsl.Stop()

	// Clear pools to release memory (optional but good practice)
	parserPool = sync.Pool{}
	bufferPool = sync.Pool{}

	utils.Logger.Info("Video stream listener closed successfully")
	return nil
}

// GetFrameChannel returns a read-only channel for receiving video frames
func (vsl *VideoStreamListener) GetFrameChannel() <-chan VideoFrame {
	return vsl.FrameChan
}

func (vsl *VideoStreamListener) onVideoStreamData(data []byte, addr *net.UDPAddr) {
	utils.Logger.Debugf("Received %d bytes of video data from %s", len(data), addr.String())

	// Get parser from pool
	parser := parserPool.Get().(*H264Parser)
	defer parserPool.Put(parser)

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

	img := vf.toRGBImage()
	if img != nil {
		enhanced.Image = img
		enhanced.Width = defaultVideoFrameWidth
		enhanced.Height = defaultVideoFrameHeight
		enhanced.Channels = 3
	}

	return enhanced
}

func onVideoStreamError(err error) {
	utils.Logger.Errorf("Video stream listener UDP server error: %v", err)
	// TODO: Consider if any specific errors need more handling (e.g., re-establishing connection if possible).
}

// toRGBImage converts the raw frame bytes into an RGB image so downstream
// processors (e.g. YOLO) always have pixel data available even before
// a full H.264 decoder is integrated.
func (vf *VideoFrame) toRGBImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, defaultVideoFrameWidth, defaultVideoFrameHeight))

	baseColor := color.RGBA{R: 30, G: 30, B: 30, A: 255}
	if vf.IsKeyFrame {
		baseColor = color.RGBA{R: 30, G: 90, B: 160, A: 255}
	}

	draw.Draw(img, img.Bounds(), &image.Uniform{C: baseColor}, image.Point{}, draw.Src)

	data := vf.Data
	if len(data) == 0 {
		return img
	}

	dataLen := len(data)
	dataIdx := 0

	for y := 0; y < defaultVideoFrameHeight; y += colorBlockSize {
		for x := 0; x < defaultVideoFrameWidth; x += colorBlockSize {
			r := data[dataIdx%dataLen]
			g := data[(dataIdx+1)%dataLen]
			b := data[(dataIdx+2)%dataLen]
			col := color.RGBA{R: r, G: g, B: b, A: 255}

			for by := 0; by < colorBlockSize && (y+by) < defaultVideoFrameHeight; by++ {
				for bx := 0; bx < colorBlockSize && (x+bx) < defaultVideoFrameWidth; bx++ {
					img.Set(x+bx, y+by, col)
				}
			}

			dataIdx += 3
		}
	}

	return img
}
