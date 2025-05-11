package transport

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Typical writers:
//   - os.Stdout                         -> pipe to `ffplay -i -`
//   - *os.File                          -> save to .h264 / .mp4 via later mux
//   - bytes.Buffer / net.Pipe() writer  -> hand off to gmf / gocv decoder
type VideoStream struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	in   <-chan []byte // from Conn.Video()
	out  io.Writer
	errC chan error
}

func NewVideoStream(parent context.Context, conn *Conn, w io.Writer) *VideoStream {
	ctx, cancel := context.WithCancel(parent)

	vs := &VideoStream{
		ctx:    ctx,
		cancel: cancel,
		in:     conn.Video(),
		out:    w,
		errC:   make(chan error, 4),
	}

	vs.wg.Add(1)
	go vs.forward()
	return vs
}

func (vs *VideoStream) Errors() <-chan error { return vs.errC }

func (vs *VideoStream) Close() {
	vs.cancel()
	vs.wg.Wait()
	close(vs.errC)
}

func (vs *VideoStream) forward() {
	defer vs.wg.Done()

	for {
		select {
		case <-vs.ctx.Done():
			return

		case pkt := <-vs.in:
			if len(pkt) == 0 {
				continue
			}
			if _, err := vs.out.Write(pkt); err != nil {
				vs.errC <- fmt.Errorf("video write: %w", err)
				// Here we bail out because downstream likely closed.
				return
			}
		}
	}
}
