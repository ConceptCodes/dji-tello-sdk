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