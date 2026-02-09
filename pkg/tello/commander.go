package tello

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/errors"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

type FlipDirection string

const (
	FlipLeft     FlipDirection = "l"
	FlipRight    FlipDirection = "r"
	FlipForward  FlipDirection = "f"
	FlipBackward FlipDirection = "b"
)

type telloCommander struct {
	commandClient       CommandConnection
	commandQueue        *PriorityCommandQueue
	stateListener       *transport.StateListener
	videoStreamListener *transport.VideoStreamListener
	videoFrameCallback  VideoFrameCallback
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
}

// CommandConnection interface for command sending
type CommandConnection interface {
	SendCommand(command string) (string, error)
	Close() error
}

// VideoFrameCallback is called when a new video frame is received
type VideoFrameCallback func(frame transport.VideoFrame)

type TelloCommander interface {
	// Control Commands
	Init() error                                   // Enter SDK mode
	TakeOff() error                                // Automatic takeoff
	Land() error                                   // Automatic landing
	StreamOn() error                               // Start video stream
	StreamOff() error                              // Stop video stream
	Emergency() error                              // Stop all motors immediately
	Up(distance int) error                         // Fly up a certain distance (cm)
	Down(distance int) error                       // Fly down a certain distance (cm)
	Left(distance int) error                       // Fly left a certain distance (cm)
	Right(distance int) error                      // Fly right a certain distance (cm)
	Forward(distance int) error                    // Fly forward a certain distance (cm)
	Backward(distance int) error                   // Fly backward a certain distance (cm)
	Clockwise(angle int) error                     // Rotate clockwise a certain angle (degrees)
	CounterClockwise(angle int) error              // Rotate counter-clockwise a certain angle (degrees)
	Flip(direction FlipDirection) error            // Flip in a certain direction (l/r/f/b)
	Go(x, y, z, speed int) error                   // Fly to a certain position (x, y, z) with a certain speed (cm/s)
	Curve(x1, y1, z1, x2, y2, z2, speed int) error // Fly in a curve to a certain position (x1, y1, z1) and (x2, y2, z2) with a certain speed (cm/s)

	// Set Commands
	SetSpeed(speed int) error                       // Set the speed of the drone (cm/s)
	SetRcControl(a, b, c, d int) error              // Set the RC control values (a: left/right, b: forward/backward, c: up/down, d: yaw)
	SetWiFiCredentials(ssid, password string) error // Set the WiFi credentials for the drone

	// Read Commands
	GetSpeed() (int, error)                  // Get the current speed of the drone (cm/s)
	GetBatteryPercentage() (int, error)      // Get the current battery percentage of the drone
	GetTime() (int, error)                   // Get the current flight time of the drone (seconds)
	GetHeight() (int, error)                 // Get the current height of the drone (cm)
	GetTemperature() (int, error)            // Get the current temperature of the drone (degrees Celsius)
	GetAttitude() (int, int, int, error)     // Get the current attitude of the drone (pitch, roll, yaw)
	GetBarometer() (int, error)              // Get the current barometer of the drone (m)
	GetAcceleration() (int, int, int, error) // Get the current acceleration of the drone (x, y, z)
	GetTof() (int, error)                    // Get the distance value from time of flight of the drone (cm)

	// Video Commands
	SetVideoFrameCallback(callback VideoFrameCallback) // Set callback for video frames
	GetVideoFrameChannel() <-chan transport.VideoFrame // Get read-only channel for video frames

	// Lifecycle Commands
	Shutdown() error // Gracefully shutdown all components and clean up resources
}

func NewTelloCommander(
	commandClient CommandConnection,
	commandQueue *PriorityCommandQueue,
	stateListener *transport.StateListener,
	videoStreamListener *transport.VideoStreamListener,
) TelloCommander {
	ctx, cancel := context.WithCancel(context.Background())
	tc := &telloCommander{
		commandClient:       commandClient,
		commandQueue:        commandQueue,
		stateListener:       stateListener,
		videoStreamListener: videoStreamListener,
		ctx:                 ctx,
		cancel:              cancel,
	}
	tc.wg.Add(1)
	go tc.processCommandQueue()
	return tc
}

func (t *telloCommander) processCommandQueue() {
	defer t.wg.Done()

	for {
		select {
		case <-t.ctx.Done():
			utils.Logger.Info("Command queue processor shutting down...")
			return
		default:
			req, ok := t.commandQueue.Dequeue()
			if !ok {
				continue
			}

			respStr, err := t.sendCommand(req.Command)

			// If there's a response channel, send the result back
			if req.ResponseChan != nil {
				req.ResponseChan <- CommandResponse{
					Response: respStr,
					Error:    err,
				}
				close(req.ResponseChan)
			} else if err != nil {
				// Only log errors for fire-and-forget commands
				utils.Logger.Errorf("Failed to process command '%s': %v", req.Command, err)
			}
		}
	}
}

func (t *telloCommander) sendCommand(cmd string) (string, error) {
	utils.Logger.Debugf("Sending command: %s", cmd)

	response, err := t.commandClient.SendCommand(cmd)
	if err != nil {
		return "", errors.CommandError(cmd, err)
	}

	respStr := string(response)
	if respStr != "ok" && respStr != "OK" {
		// For read commands, the response is the value, so we don't treat it as an error unless it says "error"
		if respStr == "error" || respStr == "ERROR" {
			return "", errors.CommandError(cmd, fmt.Errorf("command returned error response: %s", respStr))
		}
		// For control commands, we expect "ok", but for read commands we expect a value.
		// Since we don't know the command type here easily without parsing, we return the response string.
		// The caller (or the response channel receiver) can decide if it's valid.
	}

	utils.Logger.Infof("Command '%s' sent successfully. Response: %s", cmd, respStr)
	return respStr, nil
}

func (t *telloCommander) sendReadCommand(cmd string) (string, error) {
	utils.Logger.Debugf("Enqueuing read command: %s", cmd)

	// Enqueue with high priority and get response channel
	respChan := t.commandQueue.EnqueueRead(cmd)

	// Create a timeout context to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(t.ctx, 10*time.Second)
	defer cancel()

	// Wait for response
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return "", resp.Error
		}
		return resp.Response, nil
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.NewSDKError(errors.ErrTimeout, "TelloCommander",
				fmt.Sprintf("timeout waiting for response to command '%s' (10s exceeded)", cmd))
		}
		return "", errors.NewSDKError(errors.ErrTimeout, "TelloCommander",
			fmt.Sprintf("context cancelled while waiting for command '%s'", cmd))
	}
}

func (t *telloCommander) Init() error {
	utils.Logger.Debugf("Initializing SDK mode")
	// Init is a control command but we want to wait for it to ensure it succeeds
	// However, sendCommand is now internal.
	// We can use EnqueueRead for synchronous execution even if it's not strictly a "read"
	// Or we can just fire and forget, but Init usually requires confirmation.
	// Let's use sendReadCommand for Init to ensure we wait for "ok"
	resp, err := t.sendReadCommand("command")
	if err != nil {
		return err
	}
	if resp != "ok" && resp != "OK" {
		return errors.NewSDKError(errors.ErrCommandFailed, "TelloCommander",
			fmt.Sprintf("unexpected response to init: %s", resp))
	}
	return nil
}

func (t *telloCommander) TakeOff() error {
	utils.Logger.Debugf("Enqueuing Take off Command")
	t.commandQueue.EnqueueControl("takeoff")
	return nil
}

func (t *telloCommander) Land() error {
	utils.Logger.Debugf("Enqueuing Land Command")
	t.commandQueue.EnqueueControl("land")
	return nil
}

func (t *telloCommander) StreamOn() error {
	utils.Logger.Debugf("Starting video stream")
	t.commandQueue.EnqueueControl("streamon")
	return nil
}

func (t *telloCommander) StreamOff() error {
	utils.Logger.Debugf("Stopping video stream")
	t.commandQueue.EnqueueControl("streamoff")
	return nil
}

func (t *telloCommander) Emergency() error {
	utils.Logger.Debugf("Emergency landing")
	t.commandQueue.EnqueueControl("emergency")
	return nil
}

func (t *telloCommander) Up(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying up %d cm", distance)
	cmd := fmt.Sprintf("up %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Down(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying down %d cm", distance)
	cmd := fmt.Sprintf("down %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Left(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying left %d cm", distance)
	cmd := fmt.Sprintf("left %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Right(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying right %d cm", distance)
	cmd := fmt.Sprintf("right %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Forward(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying forward %d cm", distance)
	cmd := fmt.Sprintf("forward %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Backward(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying backward %d cm", distance)
	cmd := fmt.Sprintf("back %d", distance)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Clockwise(angle int) error {
	if err := utils.ValidateNumberInRange(angle, 1, 3600); err != nil {
		return err
	}

	utils.Logger.Debugf("Rotating clockwise %d degrees", angle)
	cmd := fmt.Sprintf("cw %d", angle)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) CounterClockwise(angle int) error {
	if err := utils.ValidateNumberInRange(angle, 1, 3600); err != nil {
		return err
	}

	utils.Logger.Debugf("Rotating counter-clockwise %d degrees", angle)
	cmd := fmt.Sprintf("ccw %d", angle)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Flip(direction FlipDirection) error {
	utils.Logger.Debugf("Flipping %s", direction)
	cmd := fmt.Sprintf("flip %s", direction)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Go(x, y, z, speed int) error {
	if err := utils.ValidateNumberInRange(x, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(y, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(z, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(speed, 10, 100); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying to (%d, %d, %d) with speed %d", x, y, z, speed)
	cmd := fmt.Sprintf("go %d %d %d %d", x, y, z, speed)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) Curve(x1, y1, z1, x2, y2, z2, speed int) error {
	if err := utils.ValidateNumberInRange(x1, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(y1, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(z1, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(x2, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(y2, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(z2, 20, 500); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(speed, 10, 60); err != nil {
		return err
	}

	// Ensure x, y, z are not all between -20 and 20 at the same time
	if (x1 >= -20 && x1 <= 20) && (y1 >= -20 && y1 <= 20) && (z1 >= -20 && z1 <= 20) {
		return errors.InvalidArgumentError("TelloCommander", "curve parameters",
			"x1, y1, z1 cannot all be between -20 and 20 at the same time")
	}
	if (x2 >= -20 && x2 <= 20) && (y2 >= -20 && y2 <= 20) && (z2 >= -20 && z2 <= 20) {
		return errors.InvalidArgumentError("TelloCommander", "curve parameters",
			"x2, y2, z2 cannot all be between -20 and 20 at the same time")
	}

	if err := utils.ValidateArcRadius(x1, x2, y1, y2, z1, z2, 0.5, 10); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying in a curve to (%d, %d, %d) and (%d, %d, %d) with speed %d", x1, y1, z1, x2, y2, z2, speed)
	cmd := fmt.Sprintf("curve %d %d %d %d %d %d %d", x1, y1, z1, x2, y2, z2, speed)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) SetSpeed(speed int) error {
	if err := utils.ValidateNumberInRange(speed, 10, 100); err != nil {
		return err
	}

	utils.Logger.Debugf("Setting speed to %d cm/s", speed)
	cmd := fmt.Sprintf("speed %d", speed)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) SetRcControl(a, b, c, d int) error {
	if err := utils.ValidateNumberInRange(a, -100, 100); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(b, -100, 100); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(c, -100, 100); err != nil {
		return err
	}
	if err := utils.ValidateNumberInRange(d, -100, 100); err != nil {
		return err
	}

	utils.Logger.Debugf("Setting RC control to (%d, %d, %d, %d)", a, b, c, d)
	cmd := fmt.Sprintf("rc %d %d %d %d", a, b, c, d)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) SetWiFiCredentials(ssid, password string) error {
	if err := utils.ValidateNumberInRange(len(ssid), 1, 32); err != nil {
		return errors.InvalidArgumentError("TelloCommander", "ssid",
			"must be greater than 1 and less than 32 chars")
	}
	if len(password) < 1 {
		return errors.InvalidArgumentError("TelloCommander", "password",
			"length must be greater than 1 character")
	}

	utils.Logger.Debugf("Setting WiFi credentials to SSID: %s, Password: %s", ssid, password)
	cmd := fmt.Sprintf("wifi %s %s", ssid, password)

	t.commandQueue.EnqueueControl(cmd)
	return nil
}

func (t *telloCommander) GetSpeed() (int, error) {
	response, err := t.sendReadCommand("speed?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

func (t *telloCommander) GetBatteryPercentage() (int, error) {
	response, err := t.sendReadCommand("battery?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

func (t *telloCommander) GetTime() (int, error) {
	response, err := t.sendReadCommand("time?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

func (t *telloCommander) GetHeight() (int, error) {
	response, err := t.sendReadCommand("height?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

func (t *telloCommander) GetTemperature() (int, error) {
	response, err := t.sendReadCommand("temp?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

func (t *telloCommander) GetAttitude() (int, int, int, error) {
	response, err := t.sendReadCommand("attitude?")
	if err != nil {
		return 0, 0, 0, err
	}

	// Response format: "pitch roll yaw"
	parts := strings.Fields(response)
	if len(parts) != 3 {
		return 0, 0, 0, errors.NewSDKError(errors.ErrCommandFailed, "TelloCommander",
			fmt.Sprintf("unexpected attitude response format: %s", response))
	}

	pitch, err := utils.ParseInt(parts[0])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse pitch")
	}

	roll, err := utils.ParseInt(parts[1])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse roll")
	}

	yaw, err := utils.ParseInt(parts[2])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse yaw")
	}

	return pitch, roll, yaw, nil
}

func (t *telloCommander) GetBarometer() (int, error) {
	response, err := t.sendReadCommand("baro?")
	if err != nil {
		return 0, err
	}
	// Barometer returns float value in meters, but we return int for consistency
	barometerFloat, err := utils.ParseFloat(response)
	if err != nil {
		return 0, err
	}
	return int(barometerFloat), nil
}

func (t *telloCommander) GetAcceleration() (int, int, int, error) {
	response, err := t.sendReadCommand("acceleration?")
	if err != nil {
		return 0, 0, 0, err
	}

	// Response format: "x y z" (in 0.001g units)
	parts := strings.Fields(response)
	if len(parts) != 3 {
		return 0, 0, 0, errors.NewSDKError(errors.ErrCommandFailed, "TelloCommander",
			fmt.Sprintf("unexpected acceleration response format: %s", response))
	}

	agx, err := utils.ParseInt(parts[0])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse acceleration x")
	}

	agy, err := utils.ParseInt(parts[1])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse acceleration y")
	}

	agz, err := utils.ParseInt(parts[2])
	if err != nil {
		return 0, 0, 0, errors.WrapSDKError(err, errors.ErrCommandFailed, "TelloCommander",
			"failed to parse acceleration z")
	}

	return agx, agy, agz, nil
}

func (t *telloCommander) GetTof() (int, error) {
	response, err := t.sendReadCommand("tof?")
	if err != nil {
		return 0, err
	}
	return utils.ParseInt(response)
}

// SetVideoFrameCallback sets a callback function to be called when video frames are received
func (t *telloCommander) SetVideoFrameCallback(callback VideoFrameCallback) {
	t.videoFrameCallback = callback

	// Start a goroutine to listen for video frames and call the callback
	if t.videoStreamListener != nil {
		go func() {
			frameChan := t.videoStreamListener.GetFrameChannel()
			for frame := range frameChan {
				if t.videoFrameCallback != nil {
					t.videoFrameCallback(frame)
				}
			}
		}()
	}
}

// GetVideoFrameChannel returns a read-only channel for receiving video frames
func (t *telloCommander) GetVideoFrameChannel() <-chan transport.VideoFrame {
	if t.videoStreamListener != nil {
		return t.videoStreamListener.GetFrameChannel()
	}
	return nil
}

// Initialize creates and configures a new TelloCommander with all necessary components
func Initialize() (TelloCommander, error) {
	return InitializeWithInit(true)
}

// InitializeWithInit creates and configures a new TelloCommander with optional initialization
func InitializeWithInit(autoInit bool) (TelloCommander, error) {
	// Use default transport configuration
	cfg := config.DefaultTransportConfig()

	utils.Logger.Info("Initializing Tello SDK...")

	commandClient, err := transport.NewCommandConnectionWithConfig(cfg)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create command connection", err)
	}

	commandQueue := NewPriorityCommandQueue()

	// Use configurable ports from TransportConfig
	stateListener, err := transport.NewStateListener(cfg.LocalStateAddr)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create state listener", err)
	}

	videoStreamListener, err := transport.NewVideoStreamListener(cfg.LocalVideoAddr)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create video stream listener", err)
	}

	commander := NewTelloCommander(commandClient, commandQueue, stateListener, videoStreamListener)

	go func() {
		if err := stateListener.Start(); err != nil {
			utils.Logger.Errorf("Failed to start state listener: %v", err)
		}
	}()

	go func() {
		if err := videoStreamListener.Start(); err != nil {
			utils.Logger.Errorf("Failed to start video stream listener: %v", err)
		}
	}()

	if autoInit {
		if err := commander.Init(); err != nil {
			return nil, errors.WrapSDKError(err, errors.ErrConnectionFailed, "TelloCommander",
				"SDK mode handshake failed")
		}
	}

	utils.Logger.Info("Tello SDK initialized successfully")
	return commander, nil
}

// InitializeOptions holds optional initialization parameters
type InitializeOptions struct {
	TransportConfig  config.TransportConfig
	SafetyConfigPath string
	SafetyPreset     string
	SafetyEnabled    bool
}

// WithTransportConfig specifies custom transport configuration
func WithTransportConfig(cfg config.TransportConfig) func(*InitializeOptions) {
	return func(opts *InitializeOptions) {
		opts.TransportConfig = cfg
	}
}

// WithSafetyConfig specifies a custom safety configuration file
func WithSafetyConfig(configPath string) func(*InitializeOptions) {
	return func(opts *InitializeOptions) {
		opts.SafetyConfigPath = configPath
	}
}

// WithSafetyPreset specifies a safety preset (conservative, aggressive, indoor, outdoor)
func WithSafetyPreset(preset string) func(*InitializeOptions) {
	return func(opts *InitializeOptions) {
		opts.SafetyPreset = preset
	}
}

// WithSafetyDisabled disables safety manager entirely
func WithSafetyDisabled() func(*InitializeOptions) {
	return func(opts *InitializeOptions) {
		opts.SafetyEnabled = false
	}
}

// InitializeWithOptions creates and configures a new TelloCommander with safety options
func InitializeWithOptions(opts ...func(*InitializeOptions)) (TelloCommander, error) {
	// Apply default options
	options := &InitializeOptions{
		TransportConfig: config.DefaultTransportConfig(),
		SafetyEnabled:   true, // Enable safety by default
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	// Validate transport configuration
	if err := options.TransportConfig.Validate(); err != nil {
		return nil, errors.WrapSDKError(err, errors.ErrConfigValidation, "TelloCommander",
			"invalid transport configuration")
	}

	utils.Logger.Info("Initializing Tello SDK...")

	commandClient, err := transport.NewCommandConnectionWithConfig(options.TransportConfig)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create command connection", err)
	}

	commandQueue := NewPriorityCommandQueue()

	// Use configurable ports from TransportConfig
	stateListener, err := transport.NewStateListener(options.TransportConfig.LocalStateAddr)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create state listener", err)
	}

	videoStreamListener, err := transport.NewVideoStreamListener(options.TransportConfig.LocalVideoAddr)
	if err != nil {
		return nil, errors.ConnectionError("TelloCommander", "create video stream listener", err)
	}

	commander := NewTelloCommander(commandClient, commandQueue, stateListener, videoStreamListener)

	// Note: Safety manager wrapping should be done by the caller using the safety package
	// This avoids circular imports between pkg/tello and pkg/safety
	if options.SafetyEnabled {
		utils.Logger.Infof("Safety enabled - wrap commander with safety.NewManager")
		utils.Logger.Infof("Safety config: %s", getSafetyConfigName(options))
	}

	// Start listeners with error handling
	stateListenerErr := make(chan error, 1)
	videoListenerErr := make(chan error, 1)

	go func() {
		if err := stateListener.Start(); err != nil {
			stateListenerErr <- errors.ConnectionError("TelloCommander", "start state listener", err)
		} else {
			close(stateListenerErr)
		}
	}()

	go func() {
		if err := videoStreamListener.Start(); err != nil {
			videoListenerErr <- errors.ConnectionError("TelloCommander", "start video stream listener", err)
		} else {
			close(videoListenerErr)
		}
	}()

	// Wait for listener startup and check for errors
	if err := <-stateListenerErr; err != nil {
		return nil, err
	}
	if err := <-videoListenerErr; err != nil {
		return nil, err
	}

	utils.Logger.Info("Tello SDK initialized successfully")
	return commander, nil
}

// Shutdown gracefully shuts down all components and cleans up resources
func (t *telloCommander) Shutdown() error {
	utils.Logger.Info("Shutting down Tello commander...")

	// Signal shutdown
	t.cancel()

	// Stop accepting new commands and wake up queue processor
	t.commandQueue.Close()

	// Wait for command processor to finish
	t.wg.Wait()

	// Stop listeners
	if t.stateListener != nil {
		t.stateListener.Stop()
	}

	if t.videoStreamListener != nil {
		t.videoStreamListener.Stop()
	}

	// Close command connection
	if t.commandClient != nil {
		if err := t.commandClient.Close(); err != nil {
			utils.Logger.Errorf("Error closing command connection: %v", err)
		}
	}

	utils.Logger.Info("Tello commander shutdown complete")
	return nil
}

func getSafetyConfigName(opts *InitializeOptions) string {
	if opts.SafetyConfigPath != "" {
		return opts.SafetyConfigPath
	}
	if opts.SafetyPreset != "" {
		return opts.SafetyPreset
	}
	return "default"
}
