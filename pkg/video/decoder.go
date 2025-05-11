package video

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Decoder struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	r io.Reader

	outCh chan Frame
	errCh chan error
}

func NewDecoder(ctx context.Context, r io.Reader) *Decoder {
	ctx, cancel := context.WithCancel(ctx)
	return &Decoder{
		ctx:    ctx,
		cancel: cancel,
		r:      r,
		outCh:  make(chan Frame, 16),
		errCh:  make(chan error, 4),
	}
}

// Start spins up FFmpeg + GoCV capture
func (d *Decoder) Start() {
	d.wg.Add(1)
	go d.ffmpegLoop()
}

func (d *Decoder) ffmpegLoop() {
	defer d.wg.Done()

	//-----------------------------------------------------------------
	// 1. Launch FFmpeg as a child process that reads H.264 from stdin
	//    and writes raw BGR24 frames to stdout.
	//-----------------------------------------------------------------
	cmd := exec.CommandContext(d.ctx,
		"ffmpeg", "-fflags", "nobuffer",
		"-i", "pipe:0",
		"-f", "rawvideo",
		"-pix_fmt", "bgr24",
		"pipe:1",
	)
	cmd.Stderr = io.Discard // silence FFmpeg logs
	stdin, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		d.errCh <- err
		return
	}

	//-----------------------------------------------------------------
	// 2. Tee your H.264 byte stream into FFmpeg’s stdin
	//-----------------------------------------------------------------
	go io.Copy(stdin, d.r) // closes when ctx cancels

	//-----------------------------------------------------------------
	// 3. Open a GoCV VideoCapture on FFmpeg’s stdout
	//-----------------------------------------------------------------
	cap, err := gocv.VideoCaptureFile("pipe:")
	if err != nil {
		d.errCh <- err
		return
	}
	defer cap.Close()

	mat := gocv.NewMat()
	defer mat.Close()

	for cap.Read(&mat) {
		if !mat.Empty() {
			frame := Frame{
				Img:      mat,
				Width:    mat.Cols(),
				Height:   mat.Rows(),
				PixelFmt: FormatBGR,
				PTS:      time.Now().UTC().Sub(time.Unix(0, 0)), // rough PTS
			}
			d.outCh <- frame
		}
		select {
		case <-d.ctx.Done():
			return
		default:
		}
	}
}
