package udp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type UDPServer struct {
	Addr           string
	Conn           *net.UDPConn
	OnError        func(err error)
	OnData         func(data []byte, addr *net.UDPAddr)
	ctx            context.Context
	mu             sync.RWMutex
	handlerLimit   chan struct{} // semaphore for bounded concurrency
	droppedPackets int64         // counter for dropped packets due to full handler limit
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

const defaultMaxHandlers = 100

// WithMaxHandlers sets the maximum number of concurrent OnData handlers.
// When the limit is reached, new packets are dropped and logged.
// Default is 100 handlers if not configured.
func WithMaxHandlers(n int) UDPServerOption {
	return func(s *UDPServer) {
		if n > 0 {
			s.handlerLimit = make(chan struct{}, n)
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
		// Initialize default handler limit (100 concurrent handlers)
		handlerLimit: make(chan struct{}, defaultMaxHandlers),
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

// Start begins listening for UDP packets on the configured address.
// The provided context is used for cancellation propagation to OnData callback goroutines.
// This is a blocking call that runs until the server is stopped or an unrecoverable error occurs.
func (s *UDPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.Conn != nil {
		s.mu.Unlock()
		return fmt.Errorf("UDP server on %s is already started or connection is not nil", s.Addr)
	}

	// Store context for propagation to handlers
	s.ctx = ctx

	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to resolve UDP address %s: %w", s.Addr, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to start UDP server on %s: %w", s.Addr, err)
	}
	s.Conn = conn
	s.mu.Unlock()

	utils.Logger.Infof("UDP server listening on %s", s.Addr)

	buffer := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s.mu.RLock()
		conn := s.Conn
		s.mu.RUnlock()

		if conn == nil {
			return nil
		}

		n, remoteAddr, readErr := conn.ReadFromUDP(buffer)
		if readErr != nil {
			if strings.Contains(readErr.Error(), "use of closed network connection") {
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
				// Try to acquire semaphore non-blocking
				select {
				case s.handlerLimit <- struct{}{}:
					// Acquired - spawn handler with automatic release
					go func(data []byte, addr *net.UDPAddr, ctx context.Context) {
						defer func() { <-s.handlerLimit }()

						select {
						case <-ctx.Done():
							return
						default:
							s.OnData(data, addr)
						}
					}(dataCopy, remoteAddr, ctx)
				default:
					// Semaphore full - drop packet and log
					atomic.AddInt64(&s.droppedPackets, 1)
					if s.OnError != nil {
						s.OnError(fmt.Errorf("UDP packet dropped on %s: handler limit reached (%d concurrent handlers)", s.Addr, cap(s.handlerLimit)))
					}
				}
			}
		}
	}
}

func (s *UDPServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

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
