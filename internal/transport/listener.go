package transport

import (
	"context"
	"fmt"
	"net"
	"time"
)

type UDPListenerConfig struct {
	conn    net.PacketConn
	timeout time.Duration
	ctx     context.Context
}

func NewUDPListener(port int, timeout time.Duration, ctx context.Context) (*UDPListenerConfig, error) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listen on :%d: %w", port, err)
	}
	return &UDPListenerConfig{conn: conn, timeout: timeout, ctx: ctx}, nil
}

func (l *UDPListenerConfig) Receive(buf []byte) (int, net.Addr, error) {
	deadline := time.Now().Add(l.timeout)
	if err := l.conn.SetReadDeadline(deadline); err != nil {
		return 0, nil, err
	}

	n, addr, err := l.conn.ReadFrom(buf)
	if err != nil {
		select {
		case <-l.ctx.Done():
			return 0, nil, l.ctx.Err()
		default:
			return 0, nil, err
		}
	}
	return n, addr, nil
}

func (l *UDPListenerConfig) Close() error        { return l.conn.Close() }
func (l *UDPListenerConfig) LocalAddr() net.Addr { return l.conn.LocalAddr() }
