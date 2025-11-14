package tello

import (
	"errors"
	"testing"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/transport"
)

// MockCommandConnection is a mock implementation of CommandConnection for testing
type MockCommandConnection struct {
	responses    map[string]string
	errors       map[string]error
	sentCommands []string
}

func NewMockCommandConnection() *MockCommandConnection {
	return &MockCommandConnection{
		responses:    make(map[string]string),
		errors:       make(map[string]error),
		sentCommands: make([]string, 0),
	}
}

func (m *MockCommandConnection) SetResponse(command, response string) {
	m.responses[command] = response
}

func (m *MockCommandConnection) SetError(command string, err error) {
	m.errors[command] = err
}

func (m *MockCommandConnection) SendCommand(command string) (string, error) {
	m.sentCommands = append(m.sentCommands, command)

	if err, exists := m.errors[command]; exists {
		return "", err
	}

	if response, exists := m.responses[command]; exists {
		return response, nil
	}

	// Default response for control commands
	return "ok", nil
}

func (m *MockCommandConnection) GetSentCommands() []string {
	return m.sentCommands
}

func (m *MockCommandConnection) Close() error {
	return nil
}

func TestNewTelloCommander(t *testing.T) {
	mockConn := NewMockCommandConnection()
	queue := NewPriorityCommandQueue()

	// Create mock state and video listeners (we'll test these separately)
	stateListener := &transport.StateListener{}
	videoListener := &transport.VideoStreamListener{}

	commander := NewTelloCommander(mockConn, queue, stateListener, videoListener)

	if commander == nil {
		t.Error("Expected commander to be created, got nil")
	}
}

// Read Command Tests
func TestGetSpeed(t *testing.T) {
	// Create a mock command connection
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("speed?", "50")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	speed, err := commander.GetSpeed()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if speed != 50 {
		t.Errorf("Expected speed 50, got %d", speed)
	}

	commands := mockConn.GetSentCommands()
	if len(commands) != 1 || commands[0] != "speed?" {
		t.Errorf("Expected 'speed?' command to be sent, got %v", commands)
	}
}

func TestGetBatteryPercentage(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("battery?", "85")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	battery, err := commander.GetBatteryPercentage()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if battery != 85 {
		t.Errorf("Expected battery 85, got %d", battery)
	}
}

func TestGetTime(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("time?", "120")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	time, err := commander.GetTime()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if time != 120 {
		t.Errorf("Expected time 120, got %d", time)
	}
}

func TestGetHeight(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("height?", "100")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	height, err := commander.GetHeight()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if height != 100 {
		t.Errorf("Expected height 100, got %d", height)
	}
}

func TestGetTemperature(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("temp?", "25")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	temp, err := commander.GetTemperature()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if temp != 25 {
		t.Errorf("Expected temperature 25, got %d", temp)
	}
}

func TestGetAttitude(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("attitude?", "10 -5 180")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	pitch, roll, yaw, err := commander.GetAttitude()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if pitch != 10 || roll != -5 || yaw != 180 {
		t.Errorf("Expected attitude (10, -5, 180), got (%d, %d, %d)", pitch, roll, yaw)
	}
}

func TestGetAttitudeInvalidFormat(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("attitude?", "invalid")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	_, _, _, err := commander.GetAttitude()
	if err == nil {
		t.Error("Expected error for invalid attitude format, got nil")
	}
}

func TestGetBarometer(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("baro?", "1013.25")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	baro, err := commander.GetBarometer()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if baro != 1013 { // Should be truncated to int
		t.Errorf("Expected barometer 1013, got %d", baro)
	}
}

func TestGetAcceleration(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("acceleration?", "100 -200 50")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	agx, agy, agz, err := commander.GetAcceleration()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if agx != 100 || agy != -200 || agz != 50 {
		t.Errorf("Expected acceleration (100, -200, 50), got (%d, %d, %d)", agx, agy, agz)
	}
}

func TestGetAccelerationInvalidFormat(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("acceleration?", "invalid")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	_, _, _, err := commander.GetAcceleration()
	if err == nil {
		t.Error("Expected error for invalid acceleration format, got nil")
	}
}

func TestGetTof(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("tof?", "300")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	tof, err := commander.GetTof()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if tof != 300 {
		t.Errorf("Expected TOF 300, got %d", tof)
	}
}

// Control Command Tests
func TestInit(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("command", "ok")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	err := commander.Init()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	commands := mockConn.GetSentCommands()
	if len(commands) != 1 || commands[0] != "command" {
		t.Errorf("Expected 'command' to be sent, got %v", commands)
	}
}

func TestTakeOff(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.TakeOff()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.Size())
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "takeoff" {
		t.Errorf("Expected 'takeoff' command, got '%s'", cmd)
	}
}

func TestLand(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.Land()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.Size())
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "land" {
		t.Errorf("Expected 'land' command, got '%s'", cmd)
	}
}

func TestEmergency(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.Emergency()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "emergency" {
		t.Errorf("Expected 'emergency' command, got '%s'", cmd)
	}
}

func TestMovementCommands(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	// Test all movement commands
	tests := []struct {
		name    string
		method  func(int) error
		command string
		value   int
	}{
		{"Up", commander.Up, "up 50", 50},
		{"Down", commander.Down, "down 30", 30},
		{"Left", commander.Left, "left 40", 40},
		{"Right", commander.Right, "right 60", 60},
		{"Forward", commander.Forward, "forward 100", 100},
		{"Backward", commander.Backward, "back 80", 80},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.method(test.value)
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", test.name, err)
			}

			cmd, ok := queue.Dequeue()
			if !ok || cmd != test.command {
				t.Errorf("Expected '%s' command for %s, got '%s'", test.command, test.name, cmd)
			}
		})
	}
}

func TestMovementCommandsValidation(t *testing.T) {
	commander := &telloCommander{
		commandQueue: NewPriorityCommandQueue(),
	}

	// Test out of range values
	tests := []struct {
		name   string
		method func(int) error
		value  int
	}{
		{"UpTooSmall", commander.Up, 10},  // Below 20
		{"UpTooLarge", commander.Up, 600}, // Above 500
		{"DownTooSmall", commander.Down, 19},
		{"DownTooLarge", commander.Down, 501},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.method(test.value)
			if err == nil {
				t.Errorf("Expected validation error for %s with value %d", test.name, test.value)
			}
		})
	}
}

func TestRotationCommands(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	// Test rotation commands
	err := commander.Clockwise(90)
	if err != nil {
		t.Errorf("Expected no error for clockwise rotation, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "cw 90" {
		t.Errorf("Expected 'cw 90' command, got '%s'", cmd)
	}

	err = commander.CounterClockwise(180)
	if err != nil {
		t.Errorf("Expected no error for counter-clockwise rotation, got %v", err)
	}

	cmd, ok = queue.Dequeue()
	if !ok || cmd != "ccw 180" {
		t.Errorf("Expected 'ccw 180' command, got '%s'", cmd)
	}
}

func TestFlipCommand(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.Flip(FlipLeft)
	if err != nil {
		t.Errorf("Expected no error for flip command, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "flip l" {
		t.Errorf("Expected 'flip l' command, got '%s'", cmd)
	}
}

func TestSetSpeed(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.SetSpeed(75)
	if err != nil {
		t.Errorf("Expected no error for set speed, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "speed 75" {
		t.Errorf("Expected 'speed 75' command, got '%s'", cmd)
	}
}

func TestSetRcControl(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.SetRcControl(50, -30, 0, 90)
	if err != nil {
		t.Errorf("Expected no error for RC control, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "rc 50 -30 0 90" {
		t.Errorf("Expected 'rc 50 -30 0 90' command, got '%s'", cmd)
	}
}

func TestSetWiFiCredentials(t *testing.T) {
	queue := NewPriorityCommandQueue()

	commander := &telloCommander{
		commandQueue: queue,
	}

	err := commander.SetWiFiCredentials("MyWiFi", "password123")
	if err != nil {
		t.Errorf("Expected no error for WiFi credentials, got %v", err)
	}

	cmd, ok := queue.Dequeue()
	if !ok || cmd != "wifi MyWiFi password123" {
		t.Errorf("Expected 'wifi MyWiFi password123' command, got '%s'", cmd)
	}
}

// Error Handling Tests
func TestSendReadCommandError(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetError("speed?", errors.New("network error"))

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	_, err := commander.GetSpeed()
	if err == nil {
		t.Error("Expected error for failed read command, got nil")
	}

	if err == nil {
		t.Error("Expected error for failed read command, got nil")
	}

	// Check that error contains our expected message
	errMsg := err.Error()
	expectedMsg := "network error"
	if !contains(errMsg, expectedMsg) {
		t.Errorf("Expected error to contain '%s', got %v", expectedMsg, err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSendReadCommandErrorResponse(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("speed?", "error")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	_, err := commander.GetSpeed()
	if err == nil {
		t.Error("Expected error for error response, got nil")
	}
}

func TestSendCommandErrorResponse(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("takeoff", "error")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	err := commander.sendCommand("takeoff")
	if err == nil {
		t.Error("Expected error for error response, got nil")
	}
}

func TestSendCommandUnexpectedResponse(t *testing.T) {
	mockConn := NewMockCommandConnection()
	mockConn.SetResponse("takeoff", "unexpected")

	commander := &telloCommander{
		commandClient: mockConn,
		commandQueue:  NewPriorityCommandQueue(),
	}

	err := commander.sendCommand("takeoff")
	if err == nil {
		t.Error("Expected error for unexpected response, got nil")
	}
}
