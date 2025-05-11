package tello

import (
	"context"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

type ConfigurableSafetyManager struct {
	batteryThreshold  int
	keepAliveInterval time.Duration
	maxAltitude       int
	minAltitude       int
	telemetryTimeout  time.Duration
}

const (
	defaultBatteryThreshold  = 20
	defaultKeepAliveInterval = 5 * time.Second
	defaultMaxAltitude       = 100
	defaultMinAltitude       = 0
	defaultTelemetryTimeout  = 10 * time.Second
)

func InitializeSDK(ctx context.Context, commChannel transport.CommandConn, stateStream *transport.StateStream, config *ConfigurableSafetyManager) *TelloCommanderConfig {
	commandQueue := NewCommandQueue(ctx)
	commander := NewTelloCommander(ctx, commChannel, commandQueue)

	commandCh := make(chan string)
	commander.StartQueue()
	commander.StartCommandListener(commandCh)

	defaultSafetyManager := ConfigurableSafetyManager{
		batteryThreshold:  defaultBatteryThreshold,
		keepAliveInterval: defaultKeepAliveInterval,
		maxAltitude:       defaultMaxAltitude,
		minAltitude:       defaultMinAltitude,
		telemetryTimeout:  defaultTelemetryTimeout,
	}

	if config != nil {
		if config.batteryThreshold != 0 {
			defaultSafetyManager.batteryThreshold = config.batteryThreshold
		}
		if config.keepAliveInterval != 0 {
			defaultSafetyManager.keepAliveInterval = config.keepAliveInterval
		}
		if config.maxAltitude != 0 {
			defaultSafetyManager.maxAltitude = config.maxAltitude
		}
		if config.minAltitude != 0 {
			defaultSafetyManager.minAltitude = config.minAltitude
		}
		if config.telemetryTimeout != 0 {
			defaultSafetyManager.telemetryTimeout = config.telemetryTimeout
		}
	}

	safetyManager := NewSafetyManager(
		ctx,
		stateStream.Out(),
		commandQueue,
		defaultSafetyManager.batteryThreshold,
		defaultSafetyManager.keepAliveInterval,
		defaultSafetyManager.maxAltitude,
		defaultSafetyManager.minAltitude,
		defaultSafetyManager.telemetryTimeout,
	)

	safetyManager.Start()

	return commander
}
