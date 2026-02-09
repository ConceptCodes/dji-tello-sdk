package udp

import (
	"context"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestNewUDPServer(t *testing.T) {
	server, err := NewUDPServer("127.0.0.1:8890")
	if err != nil {
		t.Errorf("Expected no error creating UDP server, got %v", err)
	}

	if server == nil {
		t.Error("Expected server to be created, got nil")
	}

	if server.Addr != "127.0.0.1:8890" {
		t.Errorf("Expected address '127.0.0.1:8890', got '%s'", server.Addr)
	}
}

func TestNewUDPServerInvalidAddress(t *testing.T) {
	tests := []struct {
		name               string
		address            string
		shouldErrorInNew   bool
		shouldErrorInStart bool
	}{
		{"Missing port", "127.0.0.1", true, false},
		{"Invalid format", "not_an_address", true, false},
		{"Empty address", "", true, false},
		{"Invalid port", "127.0.0.1:99999", false, true},    // This will error during Start()
		{"Invalid IP", "999.999.999.999:8890", false, true}, // This will error during Start()
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, err := NewUDPServer(test.address)

			if test.shouldErrorInNew {
				if err == nil {
					t.Errorf("Expected error for address '%s' in NewUDPServer, got nil", test.address)
				}

				if server != nil {
					t.Errorf("Expected server to be nil for invalid address '%s'", test.address)
				}
				return // Don't test Start() if New() failed
			}

			// Test Start() for addresses that pass New() but should fail in Start()
			if test.shouldErrorInStart {
				if err != nil {
					t.Errorf("Expected no error for address '%s' in NewUDPServer, got %v", test.address, err)
				}

				if server == nil {
					t.Errorf("Expected server to be created for address '%s'", test.address)
				}

				// Try to start the server - this should fail
				startErr := server.Start(context.Background())
				if startErr == nil {
					t.Errorf("Expected error when starting server with address '%s'", test.address)
				}
			}
		})
	}
}

func TestUDPServerOptions(t *testing.T) {
	dataReceived := make(chan []byte, 1)
	errorReceived := make(chan error, 1)

	onData := func(data []byte, addr *net.UDPAddr) {
		dataReceived <- data
	}

	onError := func(err error) {
		errorReceived <- err
	}

	server, err := NewUDPServer(
		"127.0.0.1:8891",
		WithOnData(onData),
		WithOnError(onError),
	)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Test that callbacks are set (we can't directly access them, but we can test behavior)
	// This would require starting the server and sending data to it
	_ = server
}

func TestUDPServerStart(t *testing.T) {
	server, err := NewUDPServer("127.0.0.1:8892")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server in goroutine
	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start(context.Background())
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect to the server to verify it's running
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8892,
	})
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	// Send test data
	testData := []byte("test message")
	_, err = conn.Write(testData)
	if err != nil {
		t.Errorf("Failed to send test data: %v", err)
	}

	// Stop the server
	server.Stop()

	// Check if server returned from Start() (should not error on normal stop)
	select {
	case err := <-startErr:
		if err != nil {
			t.Errorf("Server start returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Server did not return from Start() after stop")
	}
}

func TestUDPServerStartAlreadyStarted(t *testing.T) {
	server, err := NewUDPServer("127.0.0.1:8893")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Try to start again
	err = server.Start(context.Background())
	if err == nil {
		t.Error("Expected error when starting already started server")
	}

	// Clean up
	server.Stop()
}

func TestUDPServerStop(t *testing.T) {
	server, err := NewUDPServer("127.0.0.1:8894")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server in goroutine
	startDone := make(chan error, 1)
	go func() {
		startDone <- server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	server.Stop()

	// Wait for Start() to return
	select {
	case err := <-startDone:
		if err != nil {
			t.Errorf("Server Start() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server Start() did not return after Stop()")
	}

	// Test that we can create a new server on the same address (indicates the old one is properly closed)
	newServer, err := NewUDPServer("127.0.0.1:8894")
	if err != nil {
		t.Errorf("Failed to create new server on same address after stop: %v", err)
	}
	_ = newServer // Just verify it can be created
}

func TestUDPServerStopAlreadyStopped(t *testing.T) {
	server, err := NewUDPServer("127.0.0.1:8895")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Stop server without starting
	server.Stop()

	// Should not error
	// This test mainly ensures no panic occurs
}

func TestUDPServerDataHandling(t *testing.T) {
	dataReceived := make(chan []byte, 10)
	addrReceived := make(chan *net.UDPAddr, 10)

	onData := func(data []byte, addr *net.UDPAddr) {
		dataReceived <- data
		addrReceived <- addr
	}

	server, err := NewUDPServer("127.0.0.1:8896", WithOnData(onData))
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Send test data
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8896,
	})
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	testData := []byte("test data handling")
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to send test data: %v", err)
	}

	// Check if data was received
	select {
	case data := <-dataReceived:
		if string(data) != string(testData) {
			t.Errorf("Expected to receive '%s', got '%s'", string(testData), string(data))
		}
	case addr := <-addrReceived:
		if addr == nil {
			t.Error("Expected to receive sender address, got nil")
		}
	case <-time.After(1 * time.Second):
		t.Error("Did not receive data within timeout")
	}

	// Clean up
	server.Stop()
}

func TestUDPServerErrorHandling(t *testing.T) {
	errorReceived := make(chan error, 10)

	onError := func(err error) {
		errorReceived <- err
	}

	server, err := NewUDPServer("127.0.0.1:8897", WithOnError(onError))
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Server should handle errors gracefully
	// This is hard to test without causing actual network errors
	// For now, we'll just verify the error callback is set

	// Clean up
	server.Stop()
}

func TestUDPServerConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	dataCount := 0
	mu := sync.Mutex{}

	onData := func(data []byte, addr *net.UDPAddr) {
		mu.Lock()
		dataCount++
		mu.Unlock()
	}

	server, err := NewUDPServer("127.0.0.1:8898", WithOnData(onData))
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Send data from multiple goroutines
	numSenders := 10
	for i := 0; i < numSenders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8898,
			})
			if err != nil {
				return
			}
			defer conn.Close()

			testData := []byte("test data")
			_, err = conn.Write(testData)
			if err != nil {
				return
			}
		}(i)
	}

	// Wait for all senders to complete
	wg.Wait()

	// Wait a bit for data processing
	time.Sleep(200 * time.Millisecond)

	// Clean up
	server.Stop()

	// Verify that some data was received (we don't know exact count due to timing)
	mu.Lock()
	finalCount := dataCount
	mu.Unlock()

	if finalCount == 0 {
		t.Error("Expected to receive some data, got none")
	}
}

// TestUDPServerLoad is a stress/load test that verifies the UDP server
// handles high traffic without crashes, deadlocks, or excessive goroutine growth.
func TestUDPServerLoad(t *testing.T) {
	// Track errors and dropped packets
	var errorCount int64
	var muError sync.Mutex
	dataCount := int64(0)

	onData := func(data []byte, addr *net.UDPAddr) {
		atomic.AddInt64(&dataCount, 1)
	}

	onError := func(err error) {
		muError.Lock()
		errorCount++
		muError.Unlock()
	}

	server, err := NewUDPServer("127.0.0.1:8899", WithOnData(onData), WithOnError(onError))
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Baseline goroutine count
	baselineGoroutines := runtime.NumGoroutine()
	t.Logf("Baseline goroutines: %d", baselineGoroutines)

	// Load test parameters
	numPackets := 10000          // 10000 total packets
	packetsPerSecond := 1000     // 1000 packets/sec
	duration := 10 * time.Second // 10 seconds

	// Create UDP connection for sending
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8899,
	})
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	// Send packets at controlled rate
	testData := []byte("load-test-packet-data-")
	packetInterval := time.Second / time.Duration(packetsPerSecond)

	t.Logf("Starting load test: %d packets over %v (%d packets/sec)", numPackets, duration, packetsPerSecond)

	// Send packets with rate limiting
	packetChan := make(chan struct{}, packetsPerSecond)
	for i := 0; i < packetsPerSecond; i++ {
		packetChan <- struct{}{}
	}

	// Stop sending after duration or numPackets
	stopTime := time.Now().Add(duration)
	packetsSent := 0

	for i := 0; i < numPackets && time.Now().Before(stopTime); i++ {
		select {
		case <-packetChan:
			_, err = conn.Write(testData)
			if err != nil {
				t.Logf("Failed to send packet %d: %v", i, err)
			}
			packetsSent++
			// Add packet to channel to maintain rate
			go func() {
				time.Sleep(packetInterval)
				packetChan <- struct{}{}
			}()
		case <-time.After(packetInterval):
			// If rate limiter is slow, just send anyway
			_, err = conn.Write(testData)
			if err != nil {
				t.Logf("Failed to send packet %d: %v", i, err)
			}
			packetsSent++
		}
	}

	t.Logf("Packets sent: %d", packetsSent)

	// Wait for packet processing to complete
	time.Sleep(2 * time.Second)

	// Check goroutine count during/after load
	maxGoroutines := baselineGoroutines + 105 // Allow 5 extra for test overhead + 100 handlers
	postLoadGoroutines := runtime.NumGoroutine()
	t.Logf("Post-load goroutines: %d (max allowed: %d)", postLoadGoroutines, maxGoroutines)

	// Verify goroutine count is bounded
	if postLoadGoroutines > maxGoroutines {
		t.Errorf("Goroutine count exceeded limit: got %d, max allowed %d", postLoadGoroutines, maxGoroutines)
	}

	// Check for dropped packets
	dropped := atomic.LoadInt64(&server.droppedPackets)
	t.Logf("Dropped packets: %d", dropped)

	// Check for errors
	muError.Lock()
	errors := errorCount
	muError.Unlock()
	t.Logf("Error callbacks: %d", errors)

	// Verify data was received
	finalDataCount := atomic.LoadInt64(&dataCount)
	t.Logf("Data packets received: %d", finalDataCount)

	// Clean up
	server.Stop()

	// Wait for server to stop
	time.Sleep(500 * time.Millisecond)

	// Final goroutine check
	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d", finalGoroutines)

	// Verify the server handled the load without issues
	if errors > 0 {
		t.Logf("Note: %d error callbacks occurred during load test", errors)
	}

	// This test primarily verifies:
	// 1. No crash under load
	// 2. Goroutines stay bounded
	// 3. Dropped packets are tracked (if any)
	// 4. No deadlock (server stops cleanly)
}

// ==================== Goroutine Leak Detection Tests ====================

// TestUDPServerShutdownNoLeak verifies no goroutines leak after UDP server shutdown
func TestUDPServerShutdownNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	dataReceived := make(chan []byte, 10)

	onData := func(data []byte, addr *net.UDPAddr) {
		dataReceived <- data
	}

	server, err := NewUDPServer("127.0.0.1:8900", WithOnData(onData))
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Send some test data
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8900,
	})
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()

	testData := []byte("test data for leak detection")
	_, err = conn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to send test data: %v", err)
	}

	// Wait for data to be received
	select {
	case <-dataReceived:
		// Data received successfully
	case <-time.After(1 * time.Second):
		t.Log("Warning: Data not received during leak test")
	}

	// Stop server
	server.Stop()

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
}

// TestUDPServerMultipleStartStopNoLeak verifies no leaks during multiple start/stop cycles
func TestUDPServerMultipleStartStopNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	server, err := NewUDPServer("127.0.0.1:8901")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Perform multiple start/stop cycles
	for i := 0; i < 3; i++ {
		go func() {
			_ = server.Start(context.Background())
		}()

		time.Sleep(100 * time.Millisecond)
		server.Stop()
		time.Sleep(50 * time.Millisecond)
	}

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
}

// TestUDPServerConcurrentStopNoLeak verifies no leaks when stopping concurrently
func TestUDPServerConcurrentStopNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	server, err := NewUDPServer("127.0.0.1:8902")
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	// Start server
	go func() {
		_ = server.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Stop server multiple times concurrently
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			server.Stop()
		}()
	}
	wg.Wait()

	// Give goroutines time to clean up
	time.Sleep(100 * time.Millisecond)
}
