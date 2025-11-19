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

// NewUDPClient creates a UDP client that dials the provided host using an
// ephemeral local port.
func NewUDPClient(host string) (*UDPClient, error) {
	return NewUDPClientWithLocalAddr(host, "")
}

// NewUDPClientWithLocalAddr allows binding to a specific local address:port,
// which is required by some devices (like the Tello SDK using 0.0.0.0:8889 for commands).
func NewUDPClientWithLocalAddr(host, localAddr string) (*UDPClient, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address '%s': %w", host, err)
	}

	var local *net.UDPAddr
	if localAddr != "" {
		local, err = net.ResolveUDPAddr("udp", localAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve local UDP address '%s': %w", localAddr, err)
		}
	}

	conn, err := net.DialUDP("udp", local, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP address '%s': %w", host, err)
	}
	return &UDPClient{
		conn: conn,
		addr: remoteAddr,
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
