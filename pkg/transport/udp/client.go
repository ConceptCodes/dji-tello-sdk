package udp

import (
	"fmt"
	"net"
	"time"
)

type UDPClient struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewUDPClient(host string) (*UDPClient, error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address '%s': %w", host, err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP address '%s': %w", host, err)
	}
	return &UDPClient{
		conn: conn,
		addr: addr,
	}, nil
}

func (c *UDPClient) Send(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("UDP client connection is not initialized")
	}
	_, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	return nil
}

func (c *UDPClient) Receive(bufferSize int, timeout time.Duration) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("UDP client connection is not initialized")
	}
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}
	buf := make([]byte, bufferSize)
	n, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response data: %w", err)
	}
	return string(buf[:n]), nil
}

func (c *UDPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
