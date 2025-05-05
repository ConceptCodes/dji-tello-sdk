// internal/transport/state.go
package transport

import (
	"context"
	"fmt"
	"sync"

	"github.com/conceptcodes/dji-tello-sdk-go/internal/parser"
	"github.com/conceptcodes/dji-tello-sdk-go/tello"
)

type StateStream struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	in    <-chan []byte
	out   chan tello.TelloState
	errCh chan error
}

func NewStateStream(parent context.Context, conn *Conn) *StateStream {
	ctx, cancel := context.WithCancel(parent)

	s := &StateStream{
		ctx:    ctx,
		cancel: cancel,
		in:     conn.State(),
		out:    make(chan tello.TelloState, 32),
		errCh:  make(chan error, 4),
	}

	s.wg.Add(1)
	go s.loop()
	return s
}

func (s *StateStream) Out() <-chan tello.TelloState { return s.out }

func (s *StateStream) Errors() <-chan error { return s.errCh }

func (s *StateStream) Close() {
	s.cancel()
	s.wg.Wait()
	close(s.out)
	close(s.errCh)
}

func (s *StateStream) loop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return

		case pkt := <-s.in:
			state, err := parser.ParseState(string(pkt))
			if err != nil {
				s.errCh <- fmt.Errorf("parse state: %w", err)
				continue
			}
			s.out <- *state
		}
	}
}
