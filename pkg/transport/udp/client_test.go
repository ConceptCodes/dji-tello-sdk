package udp

import (
	"net"
	"testing"
	"time"
)

func TestNewUDPClient(t *testing.T) {
	// Test with valid address
	client, err := NewUDPClient("127.0.0.1:8889")
	if err != nil {
		t.Errorf("Expected no error creating UDP client, got %v", err)
	}

	if client == nil {
		t.Error("Expected client to be created, got nil")
	}

	// Clean up
	client.Close()
}

func TestNewUDPClientInvalidAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"Invalid IP", "999.999.999.999:8889"},
		{"Invalid port", "127.0.0.1:99999"},
		{"Missing port", "127.0.0.1"},
		{"Empty address", ""},
		{"Invalid format", "not_an_address"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := NewUDPClient(test.address)

			if err == nil {
				t.Errorf("Expected error for address '%s', got nil", test.address)
			}

			if client != nil {
				t.Errorf("Expected client to be nil for invalid address '%s'", test.address)
			}
		})
	}
}

func TestUDPClientSend(t *testing.T) {
	// This test requires a real UDP server
	// We'll create a simple echo server for testing
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0") // Use port 0 for automatic assignment
	if err != nil {
		t.Fatalf("Failed to resolve server address: %v", err)
	}

	server, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.Close()

	// Get the actual server address
	actualAddr := server.LocalAddr().(*net.UDPAddr)

	client, err := NewUDPClient(actualAddr.String())
	if err != nil {
		t.Fatalf("Failed to create UDP client: %v", err)
	}
	defer client.Close()

	// Test sending data
	testData := []byte("test message")
	err = client.Send(testData)
	if err != nil {
		t.Errorf("Expected no error sending data, got %v", err)
	}

	// Verify data was sent (by receiving it on server)
	server.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	buffer := make([]byte, 1024)
	n, _, err := server.ReadFromUDP(buffer)
	if err != nil {
		t.Errorf("Failed to receive sent data: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to receive %d bytes, got %d", len(testData), n)
	}

	received := string(buffer[:n])
	if received != string(testData) {
		t.Errorf("Expected to receive '%s', got '%s'", string(testData), received)
	}
}

func TestUDPClientSendWithNilConnection(t *testing.T) {
	client := &UDPClient{}

	testData := []byte("test")
	err := client.Send(testData)

	if err == nil {
		t.Error("Expected error for nil connection, got nil")
	}

	expectedError := "UDP client connection is not initialized"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUDPClientReceive(t *testing.T) {
	// This test requires a real UDP server to send data
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve server address: %v", err)
	}

	server, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}
	defer server.Close()

	actualAddr := server.LocalAddr().(*net.UDPAddr)

	client, err := NewUDPClient(actualAddr.String())
	if err != nil {
		t.Fatalf("Failed to create UDP client: %v", err)
	}
	defer client.Close()

	// Send test data from server to client
	testData := "test response"
	go func() {
		time.Sleep(50 * time.Millisecond) // Small delay to ensure client is ready to receive

		// First, receive a message from client to get client's address
		buffer := make([]byte, 1024)
		server.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, clientAddr, err := server.ReadFromUDP(buffer)
		if err == nil && n > 0 {
			// Send response back to client's address
			server.WriteToUDP([]byte(testData), clientAddr)
		}
	}()

	// Send a dummy message first to establish client address with server
	client.Send([]byte("hello"))

	// Receive data
	response, err := client.Receive(1024, 1*time.Second)
	if err != nil {
		t.Errorf("Expected no error receiving data, got %v", err)
	}

	if response != testData {
		t.Errorf("Expected to receive '%s', got '%s'", testData, response)
	}
}

func TestUDPClientReceiveTimeout(t *testing.T) {
	client, err := NewUDPClient("127.0.0.1:8889")
	if err != nil {
		t.Fatalf("Failed to create UDP client: %v", err)
	}
	defer client.Close()

	// Try to receive with very short timeout
	_, recvErr := client.Receive(1024, 1*time.Millisecond)

	if recvErr == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check if it's a timeout error (net.Error with Timeout() == true)
	if netErr, ok := recvErr.(net.Error); ok && !netErr.Timeout() {
		t.Errorf("Expected timeout error, got %v", recvErr)
	}
}

func TestUDPClientReceiveWithNilConnection(t *testing.T) {
	client := &UDPClient{}

	_, err := client.Receive(1024, 1*time.Second)

	if err == nil {
		t.Error("Expected error for nil connection, got nil")
	}

	expectedError := "UDP client connection is not initialized"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUDPClientClose(t *testing.T) {
	client, err := NewUDPClient("127.0.0.1:8889")
	if err != nil {
		t.Fatalf("Failed to create UDP client: %v", err)
	}

	// Test closing
	err = client.Close()
	if err != nil {
		t.Errorf("Expected no error closing client, got %v", err)
	}

	// Test closing already closed client - this should return an error
	err = client.Close()
	if err == nil {
		t.Error("Expected error closing already closed client, got nil")
	}
}

func TestUDPClientCloseWithNilConnection(t *testing.T) {
	client := &UDPClient{}

	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error closing nil connection, got %v", err)
	}
}
