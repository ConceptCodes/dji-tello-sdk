package transport

import (
	"fmt"
	"testing"
	"time"
)

// MockUDPClient is a mock implementation of UDPClient for testing
type MockUDPClient struct {
	sentData    []byte
	receiveData string
	receiveErr  error
	sendErr     error
	closeErr    error
}

func NewMockUDPClient() *MockUDPClient {
	return &MockUDPClient{}
}

func (m *MockUDPClient) SetReceiveData(data string, err error) {
	m.receiveData = data
	m.receiveErr = err
}

func (m *MockUDPClient) SetSendError(err error) {
	m.sendErr = err
}

func (m *MockUDPClient) SetCloseError(err error) {
	m.closeErr = err
}

func (m *MockUDPClient) Send(data []byte) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentData = data
	return nil
}

func (m *MockUDPClient) Receive(bufferSize int, timeout time.Duration) (string, error) {
	if m.receiveErr != nil {
		return "", m.receiveErr
	}
	return m.receiveData, nil
}

func (m *MockUDPClient) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	return nil
}

func (m *MockUDPClient) GetSentData() []byte {
	return m.sentData
}

func TestNewCommandConnection(t *testing.T) {
	// Test with real UDP client (integration test)
	conn, err := NewCommandConnection()
	if err != nil {
		t.Errorf("Expected no error creating command connection, got %v", err)
	}

	if conn == nil {
		t.Error("Expected connection to be created, got nil")
	}

	// Clean up
	conn.Close()
}

func TestCommandConnectionSendCommand(t *testing.T) {
	// Test with mock client using dependency injection
	mockClient := NewMockUDPClient()
	mockClient.SetReceiveData("ok", nil)

	conn := NewCommandConnectionWithClient(mockClient)

	response, err := conn.SendCommand("command")
	if err != nil {
		t.Errorf("Expected no error sending command, got %v", err)
	}

	if response != "ok" {
		t.Errorf("Expected response 'ok', got '%s'", response)
	}

	// Verify the command was sent correctly
	sentData := mockClient.GetSentData()
	expectedCommand := "command\r\n"
	if string(sentData) != expectedCommand {
		t.Errorf("Expected sent command '%s', got '%s'", expectedCommand, string(sentData))
	}
}

func TestCommandConnectionSendCommandNilClient(t *testing.T) {
	// Test with nil client using dependency injection
	conn := NewCommandConnectionWithClient(nil)

	_, err := conn.SendCommand("command")
	if err == nil {
		t.Error("Expected error for nil client, got nil")
	}

	expectedError := "UDP client is not initialized"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCommandConnectionClose(t *testing.T) {
	// Test with mock client using dependency injection
	mockClient := NewMockUDPClient()

	conn := NewCommandConnectionWithClient(mockClient)

	// Test closing
	err := conn.Close()
	if err != nil {
		t.Errorf("Expected no error closing connection, got %v", err)
	}
}

func TestCommandConnectionSendError(t *testing.T) {
	// Test send error handling
	mockClient := NewMockUDPClient()
	mockClient.SetSendError(fmt.Errorf("send failed"))

	conn := NewCommandConnectionWithClient(mockClient)

	_, err := conn.SendCommand("command")
	if err == nil {
		t.Error("Expected error for send failure, got nil")
	}

	expectedError := "failed to send command 'command': send failed"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCommandConnectionReceiveError(t *testing.T) {
	// Test receive error handling
	mockClient := NewMockUDPClient()
	mockClient.SetReceiveData("", fmt.Errorf("receive timeout"))

	conn := NewCommandConnectionWithClient(mockClient)

	_, err := conn.SendCommand("command")
	if err == nil {
		t.Error("Expected error for receive failure, got nil")
	}

	expectedError := "failed to receive response for command 'command': receive timeout"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCommandConnectionCloseError(t *testing.T) {
	// Test close error handling
	mockClient := NewMockUDPClient()
	mockClient.SetCloseError(fmt.Errorf("close failed"))

	conn := NewCommandConnectionWithClient(mockClient)

	err := conn.Close()
	if err == nil {
		t.Error("Expected error for close failure, got nil")
	}

	expectedError := "close failed"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// AdvancedMockUDPClient allows for more complex testing scenarios
type AdvancedMockUDPClient struct {
	*MockUDPClient
	sendErrors    []error
	receiveErrors []error
	sendCount     int
	receiveCount  int
}

func NewAdvancedMockUDPClient() *AdvancedMockUDPClient {
	return &AdvancedMockUDPClient{
		MockUDPClient: NewMockUDPClient(),
		sendErrors:    []error{},
		receiveErrors: []error{},
	}
}

func (m *AdvancedMockUDPClient) SetSendErrors(errors []error) {
	m.sendErrors = errors
}

func (m *AdvancedMockUDPClient) SetReceiveErrors(errors []error) {
	m.receiveErrors = errors
}

func (m *AdvancedMockUDPClient) Send(data []byte) error {
	if m.sendCount < len(m.sendErrors) && m.sendErrors[m.sendCount] != nil {
		err := m.sendErrors[m.sendCount]
		m.sendCount++
		return err
	}
	m.sendCount++
	return m.MockUDPClient.Send(data)
}

func (m *AdvancedMockUDPClient) Receive(bufferSize int, timeout time.Duration) (string, error) {
	if m.receiveCount < len(m.receiveErrors) && m.receiveErrors[m.receiveCount] != nil {
		err := m.receiveErrors[m.receiveCount]
		m.receiveCount++
		return "", err
	}
	m.receiveCount++
	return m.MockUDPClient.Receive(bufferSize, timeout)
}

func (m *AdvancedMockUDPClient) GetSendCount() int {
	return m.sendCount
}

func (m *AdvancedMockUDPClient) GetReceiveCount() int {
	return m.receiveCount
}

func TestCommandConnectionSendCommandRetryBehavior(t *testing.T) {
	tests := []struct {
		name             string
		sendErrors       []error
		receiveErrors    []error
		expectedResult   string
		expectError      bool
		expectedSends    int
		expectedReceives int
	}{
		{
			name:             "success on first attempt",
			sendErrors:       []error{nil},
			receiveErrors:    []error{nil},
			expectedResult:   "ok",
			expectError:      false,
			expectedSends:    1,
			expectedReceives: 1,
		},
		{
			name:             "success after network error retry",
			sendErrors:       []error{fmt.Errorf("connection refused"), nil},
			receiveErrors:    []error{nil, nil},
			expectedResult:   "ok",
			expectError:      false,
			expectedSends:    2,
			expectedReceives: 2,
		},
		{
			name:             "success after binding error retry with fallback",
			sendErrors:       []error{fmt.Errorf("address already in use"), nil},
			receiveErrors:    []error{nil, nil},
			expectedResult:   "ok",
			expectError:      false,
			expectedSends:    2,
			expectedReceives: 2,
		},
		{
			name:             "failure after max retries",
			sendErrors:       []error{fmt.Errorf("connection refused"), fmt.Errorf("connection refused"), fmt.Errorf("connection refused"), fmt.Errorf("connection refused")},
			receiveErrors:    []error{nil, nil, nil, nil},
			expectedResult:   "",
			expectError:      true,
			expectedSends:    3, // Default CommandRetries is 3
			expectedReceives: 0, // Won't reach receive on send failures
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewAdvancedMockUDPClient()
			mockClient.SetSendErrors(tt.sendErrors)
			mockClient.SetReceiveErrors(tt.receiveErrors)
			mockClient.SetReceiveData(tt.expectedResult, nil)

			conn := NewCommandConnectionWithClient(mockClient)

			result, err := conn.SendCommand("test")

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("Expected result '%s', got '%s'", tt.expectedResult, result)
				}
			}

			if mockClient.GetSendCount() != tt.expectedSends {
				t.Errorf("Expected %d sends, got %d", tt.expectedSends, mockClient.GetSendCount())
			}

			if mockClient.GetReceiveCount() != tt.expectedReceives {
				t.Errorf("Expected %d receives, got %d", tt.expectedReceives, mockClient.GetReceiveCount())
			}
		})
	}
}

func TestCommandConnectionSendCommandSuccessAfterRetries(t *testing.T) {
	// Test that command succeeds after initial failures
	mockClient := NewAdvancedMockUDPClient()
	// First two attempts fail with network errors, third succeeds
	mockClient.SetSendErrors([]error{
		fmt.Errorf("connection refused"),
		fmt.Errorf("connection timed out"),
		nil, // Success on third attempt
	})
	mockClient.SetReceiveErrors([]error{nil, nil, nil})
	mockClient.SetReceiveData("success", nil)

	conn := NewCommandConnectionWithClient(mockClient)

	result, err := conn.SendCommand("retry_test")

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}

	// Should have attempted 3 times (2 failures + 1 success)
	if mockClient.GetSendCount() != 3 {
		t.Errorf("Expected 3 send attempts, got %d", mockClient.GetSendCount())
	}

	if mockClient.GetReceiveCount() != 3 {
		t.Errorf("Expected 3 receive attempts, got %d", mockClient.GetReceiveCount())
	}
}

func TestCommandConnectionSendCommandDifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		response string
	}{
		{
			name:     "simple command",
			command:  "takeoff",
			response: "ok",
		},
		{
			name:     "command with parameters",
			command:  "forward 100",
			response: "ok",
		},
		{
			name:     "query command",
			command:  "battery?",
			response: "85",
		},
		{
			name:     "empty command",
			command:  "",
			response: "ok",
		},
		{
			name:     "command with special characters",
			command:  "set_speed 50",
			response: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockUDPClient()
			mockClient.SetReceiveData(tt.response, nil)

			conn := NewCommandConnectionWithClient(mockClient)

			result, err := conn.SendCommand(tt.command)

			if err != nil {
				t.Errorf("Expected no error for command '%s', got: %v", tt.command, err)
			}

			if result != tt.response {
				t.Errorf("Expected response '%s', got '%s'", tt.response, result)
			}

			// Verify command formatting (should add \r\n)
			sentData := mockClient.GetSentData()
			expectedData := tt.command + "\r\n"
			if string(sentData) != expectedData {
				t.Errorf("Expected sent data '%q', got '%q'", expectedData, string(sentData))
			}
		})
	}
}
