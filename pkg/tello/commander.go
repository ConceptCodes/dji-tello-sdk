package tello

import (
	"fmt"

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
	commandClient       *transport.CommandConnection
	commandQueue        *CommandQueue
	stateListener       *transport.StateListener
	videoStreamListener *transport.VideoStreamListener
}

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
}

func NewTelloCommander(
	commandClient *transport.CommandConnection,
	commandQueue *CommandQueue,
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
			utils.Logger.Errorf("Error sending command '%s': %v", command, err)
			// Decide on error handling: retry, log, or stop processing
		}
	}
}

func (t *telloCommander) sendCommand(cmd string) error {
	utils.Logger.Debugf("Sending command: %s", cmd)

	response, err := t.commandClient.SendCommand(cmd)
	if err != nil {
		utils.Logger.Errorf("Error receiving response for '%s': %v", cmd, err)
		return fmt.Errorf("failed to send command '%s': %w", cmd, err)
	}

	respStr := string(response)
	if respStr != "ok" && respStr != "OK" {
		if respStr == "error" || respStr == "ERROR" {
			return fmt.Errorf("tello returned error for command '%s'", cmd)
		}
		return fmt.Errorf("unexpected response to command '%s': %s", cmd, respStr)
	}

	

	utils.Logger.Infof("Command '%s' sent successfully.", cmd)
	return nil
}

func (t *telloCommander) Init() error {
	utils.Logger.Debugf("Initializing Tello Commander")
	t.commandQueue.Enqueue("command")
	return nil
}

func (t *telloCommander) TakeOff() error {
	utils.Logger.Debugf("Taking off")
	t.commandQueue.Enqueue("takeoff")
	return nil
}

func (t *telloCommander) Land() error {
	utils.Logger.Debugf("Landing")
	t.commandQueue.Enqueue("land")
	return nil
}

func (t *telloCommander) StreamOn() error {
	utils.Logger.Debugf("Starting video stream")
	t.commandQueue.Enqueue("streamon")
	return nil
}

func (t *telloCommander) StreamOff() error {
	utils.Logger.Debugf("Stopping video stream")
	t.commandQueue.Enqueue("streamoff")
	return nil
}

func (t *telloCommander) Emergency() error {
	utils.Logger.Debugf("Emergency landing")
	t.commandQueue.Enqueue("emergency")
	return nil
}

func (t *telloCommander) Up(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying up %d cm", distance)
	cmd := fmt.Sprintf("up %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Down(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying down %d cm", distance)
	cmd := fmt.Sprintf("down %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Left(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying left %d cm", distance)
	cmd := fmt.Sprintf("left %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Right(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying right %d cm", distance)
	cmd := fmt.Sprintf("right %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Forward(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying forward %d cm", distance)
	cmd := fmt.Sprintf("forward %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Backward(distance int) error {
	if err := utils.ValidateNumberInRange(distance, 20, 500); err != nil {
		return err
	}

	utils.Logger.Debugf("Flying backward %d cm", distance)
	cmd := fmt.Sprintf("back %d", distance)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Clockwise(angle int) error {
	if err := utils.ValidateNumberInRange(angle, 1, 3600); err != nil {
		return err
	}

	utils.Logger.Debugf("Rotating clockwise %d degrees", angle)
	cmd := fmt.Sprintf("cw %d", angle)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) CounterClockwise(angle int) error {
	if err := utils.ValidateNumberInRange(angle, 1, 3600); err != nil {
		return err
	}

	utils.Logger.Debugf("Rotating counter-clockwise %d degrees", angle)
	cmd := fmt.Sprintf("ccw %d", angle)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) Flip(direction FlipDirection) error {
	utils.Logger.Debugf("Flipping %s", direction)
	cmd := fmt.Sprintf("flip %s", direction)

	t.commandQueue.Enqueue(cmd)
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

	t.commandQueue.Enqueue(cmd)
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

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) SetSpeed(speed int) error {
	if err := utils.ValidateNumberInRange(speed, 10, 100); err != nil {
		return err
	}

	utils.Logger.Debugf("Setting speed to %d cm/s", speed)
	cmd := fmt.Sprintf("speed %d", speed)

	t.commandQueue.Enqueue(cmd)
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

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) SetWiFiCredentials(ssid, password string) error {
	if err := utils.ValidateNumberInRange(len(ssid), 1, 32); err != nil {
		return err
	}
	if len(password) < 1 {
		return fmt.Errorf("password length must be greater than 1 character")
	}

	utils.Logger.Debugf("Setting WiFi credentials to SSID: %s, Password: %s", ssid, password)
	cmd := fmt.Sprintf("wifi %s %s", ssid, password)

	t.commandQueue.Enqueue(cmd)
	return nil
}

func (t *telloCommander) GetSpeed() (int, error) {
	// TODO: Implement the logic to get the speed from the drone
	return 0, nil
}

func (t *telloCommander) GetBatteryPercentage() (int, error) {
	// TODO: Implement the logic to get the battery percentage from the drone
	return 0, nil
}

func (t *telloCommander) GetTime() (int, error) {
	// TODO: Implement the logic to get the flight time from the drone
	return 0, nil
}

func (t *telloCommander) GetHeight() (int, error) {
	// TODO: Implement the logic to get the height from the drone
	return 0, nil
}

func (t *telloCommander) GetTemperature() (int, error) {
	// TODO: Implement the logic to get the temperature from the drone
	return 0, nil
}

func (t *telloCommander) GetAttitude() (int, int, int, error) {
	// TODO: Implement the logic to get the attitude from the drone
	return 0, 0, 0, nil
}

func (t *telloCommander) GetBarometer() (int, error) {
	// TODO: Implement the logic to get the barometer reading from the drone
	return 0, nil
}

func (t *telloCommander) GetAcceleration() (int, int, int, error) {
	// TODO: Implement the logic to get the acceleration from the drone
	return 0, 0, 0, nil
}

func (t *telloCommander) GetTof() (int, error) {
	// TODO: Implement the logic to get the time of flight distance from the drone
	return 0, nil
}

// var _ TelloCommander = (*telloCommander)(nil)
