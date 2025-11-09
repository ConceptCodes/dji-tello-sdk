package tello

import (
	"testing"
	"time"
)

func TestPriorityCommandQueue(t *testing.T) {
	pq := NewPriorityCommandQueue()

	// Test initial state
	if !pq.IsEmpty() {
		t.Error("Queue should be empty initially")
	}

	if pq.Size() != 0 {
		t.Error("Queue size should be 0 initially")
	}

	// Test enqueueing control commands
	pq.EnqueueControl("takeoff")
	pq.EnqueueControl("land")

	if pq.Size() != 2 {
		t.Errorf("Expected queue size 2, got %d", pq.Size())
	}

	if pq.LowPrioritySize() != 2 {
		t.Errorf("Expected low priority size 2, got %d", pq.LowPrioritySize())
	}

	// Test enqueueing read commands (high priority)
	pq.EnqueueRead("speed?")
	pq.EnqueueRead("battery?")

	if pq.Size() != 4 {
		t.Errorf("Expected queue size 4, got %d", pq.Size())
	}

	if pq.HighPrioritySize() != 2 {
		t.Errorf("Expected high priority size 2, got %d", pq.HighPrioritySize())
	}

	// Test dequeue order - high priority commands should come first
	cmd, ok := pq.Dequeue()
	if !ok || cmd != "speed?" {
		t.Errorf("Expected first command to be 'speed?', got '%s'", cmd)
	}

	cmd, ok = pq.Dequeue()
	if !ok || cmd != "battery?" {
		t.Errorf("Expected second command to be 'battery?', got '%s'", cmd)
	}

	// Now low priority commands
	cmd, ok = pq.Dequeue()
	if !ok || cmd != "takeoff" {
		t.Errorf("Expected third command to be 'takeoff', got '%s'", cmd)
	}

	cmd, ok = pq.Dequeue()
	if !ok || cmd != "land" {
		t.Errorf("Expected fourth command to be 'land', got '%s'", cmd)
	}

	// Queue should be empty now
	if !pq.IsEmpty() {
		t.Error("Queue should be empty after dequeuing all commands")
	}
}

func TestPriorityCommandQueuePeek(t *testing.T) {
	pq := NewPriorityCommandQueue()

	// Test peek on empty queue
	cmd, priority, ok := pq.Peek()
	if ok {
		t.Error("Peek on empty queue should return false")
	}

	// Add commands
	pq.EnqueueControl("land")
	pq.EnqueueRead("speed?")

	// Peek should return high priority command
	cmd, priority, ok = pq.Peek()
	if !ok {
		t.Error("Peek should return true when queue has items")
	}
	if cmd != "speed?" {
		t.Errorf("Expected peek to return 'speed?', got '%s'", cmd)
	}
	if priority != PriorityHigh {
		t.Errorf("Expected priority %d, got %d", PriorityHigh, priority)
	}

	// Peek again - should still be the same command (peek doesn't remove)
	cmd, priority, ok = pq.Peek()
	if !ok || cmd != "speed?" || priority != PriorityHigh {
		t.Error("Peek should return same command multiple times")
	}
}

func TestPriorityCommandQueueWithPriority(t *testing.T) {
	pq := NewPriorityCommandQueue()

	// Test EnqueueWithPriority
	pq.EnqueueWithPriority("takeoff", PriorityLow)
	pq.EnqueueWithPriority("speed?", PriorityHigh)
	pq.EnqueueWithPriority("land", PriorityLow)

	// Should process high priority first
	cmd, ok := pq.Dequeue()
	if !ok || cmd != "speed?" {
		t.Errorf("Expected 'speed?' first, got '%s'", cmd)
	}

	// Then low priority in FIFO order
	cmd, ok = pq.Dequeue()
	if !ok || cmd != "takeoff" {
		t.Errorf("Expected 'takeoff' second, got '%s'", cmd)
	}

	cmd, ok = pq.Dequeue()
	if !ok || cmd != "land" {
		t.Errorf("Expected 'land' third, got '%s'", cmd)
	}
}

func TestPriorityCommandQueueConcurrency(t *testing.T) {
	pq := NewPriorityCommandQueue()

	// Test concurrent access
	done := make(chan bool, 2)

	// Goroutine 1: enqueue control commands
	go func() {
		for i := 0; i < 10; i++ {
			pq.EnqueueControl("control")
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: enqueue read commands
	go func() {
		for i := 0; i < 10; i++ {
			pq.EnqueueRead("read")
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done

	// Dequeue all commands and verify priority order
	readCount := 0
	controlCount := 0
	totalCount := 0

	for totalCount < 20 {
		cmd, ok := pq.Dequeue()
		if !ok {
			break
		}

		totalCount++
		if cmd == "read" {
			readCount++
		} else if cmd == "control" {
			controlCount++
		}

		// All read commands should come before control commands
		if cmd == "control" && readCount < 10 {
			t.Errorf("Control command appeared before all read commands were processed")
		}
	}

	if readCount != 10 {
		t.Errorf("Expected 10 read commands, got %d", readCount)
	}

	if controlCount != 10 {
		t.Errorf("Expected 10 control commands, got %d", controlCount)
	}
}
