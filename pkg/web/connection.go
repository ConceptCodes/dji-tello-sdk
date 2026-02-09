package web

import (
	"fmt"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// ConnectionStatus captures the current connection information exposed via the API.
type ConnectionStatus struct {
	Connected   bool   `json:"connected"`
	Streaming   bool   `json:"streaming"`
	LastError   string `json:"last_error,omitempty"`
	LastAttempt string `json:"last_attempt,omitempty"`
}

// ConnectionCoordinator serializes connection attempts and keeps state for the UI.
type ConnectionCoordinator struct {
	commander tello.TelloCommander

	mu          sync.RWMutex
	connected   bool
	streaming   bool
	lastError   string
	lastAttempt time.Time
}

// NewConnectionCoordinator creates a new coordinator for the provided commander.
func NewConnectionCoordinator(commander tello.TelloCommander) *ConnectionCoordinator {
	return &ConnectionCoordinator{
		commander: commander,
	}
}

// Connect attempts to enter SDK mode and enable the video stream.
func (cc *ConnectionCoordinator) Connect() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.lastAttempt = time.Now()

	if cc.connected {
		return nil
	}

	if cc.commander == nil {
		cc.lastError = "commander is not available"
		return fmt.Errorf("%s", cc.lastError)
	}

	utils.Logger.Info("Attempting to connect to Tello via web interface...")

	if err := cc.commander.Init(); err != nil {
		cc.lastError = err.Error()
		return err
	}

	if err := cc.commander.StreamOn(); err != nil {
		cc.lastError = err.Error()
		return err
	}

	cc.connected = true
	cc.streaming = true
	cc.lastError = ""
	utils.Logger.Info("Tello connection established")
	return nil
}

// Disconnect stops the video stream if it was running.
func (cc *ConnectionCoordinator) Disconnect() error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.commander == nil {
		return fmt.Errorf("commander is not available")
	}

	var firstErr error
	if cc.streaming {
		if err := cc.commander.StreamOff(); err != nil {
			firstErr = err
		} else {
			cc.streaming = false
		}
	}

	cc.connected = false
	return firstErr
}

// Status returns a snapshot of the current connection information.
func (cc *ConnectionCoordinator) Status() ConnectionStatus {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	status := ConnectionStatus{
		Connected: cc.connected,
		Streaming: cc.streaming,
	}

	if cc.lastError != "" {
		status.LastError = cc.lastError
	}

	if !cc.lastAttempt.IsZero() {
		status.LastAttempt = cc.lastAttempt.Format(time.RFC3339)
	}

	return status
}

// IsConnected reports whether the drone is currently connected.
func (cc *ConnectionCoordinator) IsConnected() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.connected
}
