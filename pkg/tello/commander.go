package tello

import (
	"fmt"
	"strings"

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
}

func NewTelloCommander(
	commandClient CommandConnection,
	commandQueue *PriorityCommandQueue,
	stateListener *transport.StateListener,
	videoStreamListener *transport.VideoStreamListener,
) TelloCommander {
	tc := &telloCommander{
		commandClient:       commandClient,
		commandQueue:        commandQueue,
		stateListener:       stateListener,
		videoStreamListener: videoStreamListener,
	}
	go tc.processCommandQueue()
	return tc
}

func (t *telloCommander) processCommandQueue() {
	for {
		command, ok := t.commandQueue.Dequeue()
		if !ok {
			continue
		}

		err := t.sendCommand(command)
		if err != nil {
			// Only log at this layer, do not wrap or re-log elsewhere
			utils.Logger.Errorf("Failed to process command '%s': %v", command, err)
			continue
		}
	}
}

func (t *telloCommander) sendCommand(cmd string) error {
	utils.Logger.Debugf("Sending command: %s", cmd)

	response, err := t.commandClient.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("send command '%s' failed: %w", cmd, err)
	}

	respStr := string(response)
	if respStr != "ok" && respStr != "OK" {
		if respStr == "error" || respStr == "ERROR" {
			return fmt.Errorf("command '%s' returned error: %w", cmd, err)
		}
		return fmt.Errorf("unexpected response to command '%s': %s", cmd, respStr)
	}

	utils.Logger.Infof("Command '%s' sent successfully.", cmd)
	return nil
}

func (t *telloCommander) sendReadCommand(cmd string) (string, error) {
	utils.Logger.Debugf("Sending read command: %s", cmd)

	response, err := t.commandClient.SendCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("send read command '%s' failed: %w", cmd, err)
	}

	respStr := string(response)
	if respStr == "error" || respStr == "ERROR" {
		return "", fmt.Errorf("read command '%s' returned error", cmd)
	}

	utils.Logger.Debugf("Read command '%s' response: %s", cmd, respStr)
	return respStr, nil
}

func (t *telloCommander) Init() error {
	utils.Logger.Debugf("Initializing SDK mode")
	return t.sendCommand("command")
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
		return fmt.Errorf("x1, y1, z1 cannot all be between -20 and 20 at the same time")
	}
	if (x2 >= -20 && x2 <= 20) && (y2 >= -20 && y2 <= 20) && (z2 >= -20 && z2 <= 20) {
		return fmt.Errorf("x2, y2, z2 cannot all be between -20 and 20 at the same time")
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
		return fmt.Errorf("SSID must be greater than 1 and less than 32 chars")
	}
	if len(password) < 1 {
		return fmt.Errorf("password length must be greater than 1 character")
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
		return 0, 0, 0, fmt.Errorf("unexpected attitude response format: %s", response)
	}

	pitch, err := utils.ParseInt(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse pitch: %w", err)
	}

	roll, err := utils.ParseInt(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse roll: %w", err)
	}

	yaw, err := utils.ParseInt(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse yaw: %w", err)
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
		return 0, 0, 0, fmt.Errorf("unexpected acceleration response format: %s", response)
	}

	agx, err := utils.ParseInt(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse acceleration x: %w", err)
	}

	agy, err := utils.ParseInt(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse acceleration y: %w", err)
	}

	agz, err := utils.ParseInt(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse acceleration z: %w", err)
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
	utils.Logger.Info("Initializing Tello SDK...")

	commandClient, err := transport.NewCommandConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create command connection: %w", err)
	}

	commandQueue := NewPriorityCommandQueue()

	// Use standard Tello SDK ports
	stateListener, err := transport.NewStateListener(":8890")
	if err != nil {
		return nil, fmt.Errorf("failed to create state listener: %w", err)
	}

	videoStreamListener, err := transport.NewVideoStreamListener(":11111")
	if err != nil {
		return nil, fmt.Errorf("failed to create video stream listener: %w", err)
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

	utils.Logger.Info("Tello SDK initialized successfully")
	return commander, nil
}
