package transport_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Arrange
	timeout := 5 * time.Second
	ctx := context.Background()

	// Act
	conn, err := transport.New(timeout, ctx)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Fatal("expected conn to be non-nil")
	}
	if conn.LocalAddr() == nil {
		t.Fatal("expected conn.LocalAddr() to be non-nil")
	}
}

func TestSend(t *testing.T) {
	// Arrange
	timeout := 5 * time.Second
	ctx := context.Background()
	data := []byte("test data")

	conn, err := transport.New(timeout, ctx)
	require.NoError(t, err)

	defer conn.Close()

	// Act
	err = conn.Send(data)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReceive(t *testing.T) {
	timeout := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Arrange
	conn, err := transport.New(timeout, ctx)
	require.NoError(t, err, "transport.New should not fail")
	require.NotNil(t, conn, "conn should not be nil")
	defer conn.Close()

	testData := []byte("hello world")
	buf := make([]byte, 1024)

	localAddr := conn.LocalAddr()
	require.NotNil(t, localAddr, "LocalAddr should not be nil")

	senderConn, err := net.DialUDP("udp", nil, localAddr.(*net.UDPAddr))
	require.NoError(t, err, "Failed to dial UDP for sender")
	defer senderConn.Close()

	// Act & Assert
	var receiveErr error
	var n int
	var addr net.Addr
	done := make(chan struct{})

	go func() {
		defer close(done)
		// Give Receive a moment to start listening
		time.Sleep(50 * time.Millisecond)
		_, sendErr := senderConn.Write(testData)
		require.NoError(t, sendErr, "Sender failed to write data")
	}()

	n, addr, receiveErr = conn.Receive(buf)

	<-done

	// Assert
	require.NoError(t, receiveErr, "conn.Receive failed")
	assert.Equal(t, len(testData), n, "Received incorrect number of bytes")
	assert.Equal(t, testData, buf[:n], "Received data does not match sent data")
	assert.NotNil(t, addr, "Received address should not be nil")
	assert.Equal(t, senderConn.LocalAddr().String(), addr.String())
}

func TestClose(t *testing.T) {
	timeout := 5 * time.Second
	ctx := context.Background()

	conn, err := transport.New(timeout, ctx)
	require.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	buf := make([]byte, 10)
	_, _, err = conn.Receive(buf)
	assert.Error(t, err, "Receive should fail after Close")
}

func TestReceive_Timeout(t *testing.T) {
	timeout := 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := transport.New(timeout, ctx)
	require.NoError(t, err)
	defer conn.Close()

	buf := make([]byte, 1024)
	_, _, err = conn.Receive(buf)

	require.Error(t, err, "Expected an error due to timeout")
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Error should be context.DeadlineExceeded")
}
