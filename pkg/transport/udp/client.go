package udp

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type UDPClient struct {
	conn    *net.UDPConn
	address *net.UDPAddr
}

func NewUDPClient(serverAddr string, serverPort int) (*UDPClient, error) {
	addrStr := net.JoinHostPort(serverAddr, strconv.Itoa(serverPort))
	udpAddr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address '%s': %w", addrStr, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP address '%s': %w", addrStr, err)
	}

	return &UDPClient{
		conn:    conn,
		address: udpAddr,
	}, nil
}

func (c *UDPClient) Send(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("UDP client connection is not initialized")
	}
	_, err := c.conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *UDPClient) Receive(bufferSize int, timeout time.Duration) ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("UDP client connection is not initialized")
	}
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	buffer := make([]byte, bufferSize)
	n, _, err := c.conn.ReadFromUDP(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("timeout receiving UDP data: %w", err)
		}
		return nil, fmt.Errorf("failed to read UDP data: %w", err)
	}

	_ = c.conn.SetReadDeadline(time.Time{})

	return buffer[:n], nil
}

func (c *UDPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
