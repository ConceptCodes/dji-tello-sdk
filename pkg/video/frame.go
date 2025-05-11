package video

import (
	"time"

	"gocv.io/x/gocv"
)

type Format int

const (
	FormatUnknown Format = iota
	FormatYUV420P
	FormatBGR // convenient for OpenCV
	FormatRGB
)

type Frame struct {
	Img      gocv.Mat
	PTS      time.Duration
	Width    int
	Height   int
	PixelFmt Format
}
