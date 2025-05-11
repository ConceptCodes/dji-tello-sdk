package udp

import (
	"fmt"
	"net"
	"strings"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type UDPServer struct {
	Addr    string
	Conn    *net.UDPConn
	OnError func(err error)
	OnData  func(data []byte, addr *net.UDPAddr)
}

type UDPServerOption func(*UDPServer)

func WithOnData(onData func(data []byte, addr *net.UDPAddr)) UDPServerOption {
	return func(s *UDPServer) {
		if onData != nil {
			s.OnData = onData
		}
	}
}

func WithOnError(onError func(err error)) UDPServerOption {
	return func(s *UDPServer) {
		if onError != nil {
			s.OnError = onError
		}
	}
}

func NewUDPServer(listenAddr string, opts ...UDPServerOption) (*UDPServer, error) {
	_, _, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid listen address '%s': %w. Expected format 'host:port' or ':port'", listenAddr, err)
	}

	srv := &UDPServer{
		Addr: listenAddr,
		OnData: func(data []byte, rAddr *net.UDPAddr) {
			utils.Logger.Infof("Default OnData: Received %d bytes from %s on %s: %s", len(data), rAddr.String(), listenAddr, string(data))
		},
		OnError: func(err error) {
			utils.Logger.Errorf("Default OnError: UDP Server Error on %s: %v", listenAddr, err)
		},
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *UDPServer) Start() error {
	if s.Conn != nil {
		return fmt.Errorf("UDP server on %s is already started or connection is not nil", s.Addr)
	}
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", s.Addr, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server on %s: %w", s.Addr, err)
	}
	s.Conn = conn

	utils.Logger.Infof("UDP server listening on %s", s.Addr)

	buffer := make([]byte, 2048)
	for {
		n, remoteAddr, readErr := s.Conn.ReadFromUDP(buffer)
		if readErr != nil {
			if s.Conn == nil || strings.Contains(readErr.Error(), "use of closed network connection") {
				return nil
			}

			if s.OnError != nil {
				s.OnError(fmt.Errorf("UDP server read error on %s: %w", s.Addr, readErr))
			}
			return fmt.Errorf("unrecoverable UDP server read error on %s: %w", s.Addr, readErr)
		}

		if n > 0 {
			dataCopy := make([]byte, n)
			copy(dataCopy, buffer[:n])

			if s.OnData != nil {
				go s.OnData(dataCopy, remoteAddr)
			}
		}
	}
}

func (s *UDPServer) Stop() {
	if s.Conn != nil {
		connToClose := s.Conn
		s.Conn = nil
		err := connToClose.Close()
		if err != nil && s.OnError != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				s.OnError(fmt.Errorf("error stopping UDP server on %s: %w", s.Addr, err))
			}
		}
		utils.Logger.Infof("UDP server on %s stopped.", s.Addr)
	} else {
		utils.Logger.Warnf("UDP server on %s was already stopped or not started.", s.Addr)
	}
}
