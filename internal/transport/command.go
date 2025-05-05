package transport

import (
	"context"
	"fmt"
	"net"
	"time"
)

const droneIP = "192.168.10.1"

type CommandConn struct {
	conn       net.PacketConn
	timeout    time.Duration
	ctx        context.Context
	remoteAddr *net.UDPAddr // 192.168.10.1:<port>
}

func NewCommandConn(port int, timeout time.Duration, ctx context.Context) (*CommandConn, error) {
	ra, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", droneIP, port))
	if err != nil {
		return nil, fmt.Errorf("resolve command addr: %w", err)
	}
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("open command socket: %w", err)
	}

	return &CommandConn{
		conn:       conn,
		timeout:    timeout,
		ctx:        ctx,
		remoteAddr: ra,
	}, nil
}

func (c *CommandConn) Send(b []byte) error {
	_, err := c.conn.WriteTo(b, c.remoteAddr)
	if err != nil {
		return fmt.Errorf("command send: %w", err)
	}
	return nil
}

func (c *CommandConn) Receive(buf []byte) (int, net.Addr, error) {
	deadline := time.Now().Add(c.timeout)
	if d, ok := c.ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return 0, nil, err
	}

	n, addr, err := c.conn.ReadFrom(buf)
	_ = c.conn.SetReadDeadline(time.Time{})

	if err != nil {
		select {
		case <-c.ctx.Done():
			return 0, nil, c.ctx.Err()
		default:
			return 0, nil, fmt.Errorf("command recv: %w", err)
		}
	}
	return n, addr, nil
}

func (c *CommandConn) Close() error        { return c.conn.Close() }
func (c *CommandConn) LocalAddr() net.Addr { return c.conn.LocalAddr() }
