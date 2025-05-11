package tello

import (
	"context"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/conceptcodes/dji-tello-sdk-go/shared"
)

type SafetyManager struct {
	ctx                    context.Context
	cancel                 context.CancelFunc
	stateCh                <-chan shared.TelloState
	commandQueue           *CommandQueue
	batteryLow             int
	keepAlive              time.Duration
	maxAltitude            int // Maximum safe altitude
	minAltitude            int // Minimum safe altitude
	telemetryTimeout       time.Duration
	lastTelemetryTimestamp time.Time
}

func NewSafetyManager(
	ctx context.Context,
	stateCh <-chan shared.TelloState,
	commandQueue *CommandQueue,
	batteryLow int,
	keepAlive time.Duration,
	maxAltitude int,
	minAltitude int,
	telemetryTimeout time.Duration) *SafetyManager {

	ctx, cancel := context.WithCancel(ctx)
	return &SafetyManager{
		ctx:                    ctx,
		cancel:                 cancel,
		stateCh:                stateCh,
		commandQueue:           commandQueue,
		batteryLow:             batteryLow,
		keepAlive:              keepAlive,
		maxAltitude:            maxAltitude,
		minAltitude:            minAltitude,
		telemetryTimeout:       telemetryTimeout,
		lastTelemetryTimestamp: time.Now(),
	}
}

func (sm *SafetyManager) Start() {
	go sm.keepAliveLoop()
	go sm.monitorState()
	go sm.monitorTelemetryTimeout()
}

func (sm *SafetyManager) Stop() {
	sm.cancel()
}

func (sm *SafetyManager) monitorState() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case state := <-sm.stateCh:
			sm.lastTelemetryTimestamp = time.Now()

			// Monitor battery level
			if state.Bat <= sm.batteryLow {
				utils.Logger.Warn("Battery low! Initiating auto-landing...")
				sm.commandQueue.Enqueue("land")
				return
			}

			// Monitor altitude
			if state.H > sm.maxAltitude {
				utils.Logger.Warnf("Altitude too high (%d cm)! Initiating auto-landing...", state.H)
				sm.commandQueue.Enqueue("land")
				return
			}
			if state.H < sm.minAltitude {
				utils.Logger.Warnf("Altitude too low (%d cm)! Initiating auto-landing...", state.H)
				sm.commandQueue.Enqueue("land")
				return
			}
		}
	}
}

func (sm *SafetyManager) keepAliveLoop() {
	ticker := time.NewTicker(sm.keepAlive)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			// Send keep-alive command
			sm.commandQueue.Enqueue("command")
		}
	}
}

func (sm *SafetyManager) monitorTelemetryTimeout() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			if time.Since(sm.lastTelemetryTimestamp) > sm.telemetryTimeout {
				utils.Logger.Warn("Telemetry timeout! Initiating auto-landing...")
				sm.commandQueue.Enqueue("land")
				return
			}
		}
	}
}
