package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	cmdPort   = 8889  // drone <-> laptop
	statePort = 8890  // drone -> laptop
	videoPort = 11111 // drone -> laptop

	maxPacket = 2048
)

type Conn struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// underlying sockets ---------------------------------------------------
	cmd   *CommandConn       // active client (Send)
	state *UDPListenerConfig // passive listener
	video *UDPListenerConfig // passive listener

	// user‑facing channels --------------------------------------------------
	stateCh chan []byte // raw state packets (parse elsewhere)
	videoCh chan []byte // raw H.264 payloads
	errCh   chan error  // async errors (e.g. socket closed)
	cmdQ    chan []byte // internal queue, 1‑Hz drain
}

func NewConn(timeout time.Duration, ctx context.Context) (*Conn, error) {
	ctx, cancel := context.WithCancel(ctx)

	cmd, err := NewCommandConn(cmdPort, timeout, ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("command conn: %w", err)
	}
	state, err := NewUDPListener(statePort, timeout, ctx)
	if err != nil {
		cancel()
		cmd.Close()
		return nil, fmt.Errorf("state listener: %w", err)
	}
	video, err := NewUDPListener(videoPort, timeout, ctx)
	if err != nil {
		cancel()
		cmd.Close()
		state.Close()
		return nil, fmt.Errorf("video listener: %w", err)
	}

	c := &Conn{
		ctx:     ctx,
		cancel:  cancel,
		cmd:     cmd,
		state:   state,
		video:   video,
		stateCh: make(chan []byte, 32),
		videoCh: make(chan []byte, 32),
		errCh:   make(chan error, 4),
		cmdQ:    make(chan []byte, 32),
	}
	c.start()
	return c, nil
}

func (c *Conn) SendCommand(cmd []byte) error {
	select {
	case c.cmdQ <- cmd:
		return nil
	case <-c.ctx.Done():
		return errors.New("conn closed")
	}
}

func (c *Conn) State() <-chan []byte { return c.stateCh }
func (c *Conn) Video() <-chan []byte { return c.videoCh }
func (c *Conn) Errors() <-chan error { return c.errCh }

func (c *Conn) Close() error {
	c.cancel()
	c.wg.Wait()
	_ = c.cmd.Close()
	_ = c.state.Close()
	_ = c.video.Close()
	close(c.stateCh)
	close(c.videoCh)
	close(c.errCh)
	return nil
}

func (c *Conn) start() {
	// 1‑Hz rate‑limited command sender
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case data := <-c.cmdQ:
				<-tick.C // rate‑limit
				if err := c.cmd.Send(data); err != nil {
					c.errCh <- fmt.Errorf("cmd send: %w", err)
				}
			}
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		buf := make([]byte, maxPacket)
		for {
			n, _, err := c.state.Receive(buf)
			if err != nil {
				select {
				case <-c.ctx.Done():
					return
				default:
					c.errCh <- fmt.Errorf("state recv: %w", err)
					continue
				}
			}
			pkt := make([]byte, n)
			copy(pkt, buf[:n])
			c.stateCh <- pkt
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		buf := make([]byte, maxPacket)
		for {
			n, _, err := c.video.Receive(buf)
			if err != nil {
				select {
				case <-c.ctx.Done():
					return
				default:
					c.errCh <- fmt.Errorf("video recv: %w", err)
					continue
				}
			}
			pkt := make([]byte, n)
			copy(pkt, buf[:n])
			c.videoCh <- pkt
		}
	}()
}
