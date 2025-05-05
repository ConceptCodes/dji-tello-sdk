package transport

import (
	"context"
	"fmt"
	"net"
	"time"
)

const remoteAddr = "192.168.10.1"
const port = 8889

type UDPConnection struct {
	conn       net.PacketConn
	timeout    time.Duration
	ctx        context.Context
	remoteAddr *net.UDPAddr
}

type Connection interface {
	Send(data []byte) error
	Receive(buf []byte) (int, net.Addr, error)
	Close() error
	LocalAddr() net.Addr
}

func New(timeout time.Duration, ctx context.Context) (*UDPConnection, error) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(remoteAddr),
		Port: port,
	}
	if addr == nil {
		return nil, fmt.Errorf("failed to parse remote address: %s", remoteAddr)
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote address: %w", err)
	}

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %w", err)
	}

	return &UDPConnection{
		conn:       conn,
		timeout:    timeout,
		ctx:        ctx,
		remoteAddr: remoteAddr,
	}, nil
}

func (c *UDPConnection) Send(data []byte) error {
	_, err := c.conn.WriteTo(data, c.remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	return nil
}

func (c *UDPConnection) Receive(buf []byte) (int, net.Addr, error) {
	var deadline time.Time
	connDeadline := time.Now().Add(c.timeout)
	ctxDeadline, ok := c.ctx.Deadline()

	if ok && ctxDeadline.Before(connDeadline) {
		deadline = ctxDeadline
	} else {
		deadline = connDeadline
	}

	err := c.conn.SetReadDeadline(deadline)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	n, addr, err := c.conn.ReadFrom(buf)
	if err != nil {
		select {
		case <-c.ctx.Done():
			c.conn.SetReadDeadline(time.Time{})
			return 0, nil, fmt.Errorf("receive cancelled by context: %w", c.ctx.Err())
		default:
			return 0, nil, fmt.Errorf("failed to receive data: %w", err)
		}
	}

	c.conn.SetReadDeadline(time.Time{})

	return n, addr, nil
}

func (c *UDPConnection) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

func (c *UDPConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
