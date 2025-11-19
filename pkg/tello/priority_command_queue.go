package tello

import (
	"fmt"
	"sync"
)

const (
	PriorityLow  = 0 // Control commands (takeoff, land, movement, etc.)
	PriorityHigh = 1 // Read commands (speed?, battery?, etc.)
)

// CommandResponse holds the result of a command execution
type CommandResponse struct {
	Response string
	Error    error
}

// CommandRequest holds the command and a channel for the response
type CommandRequest struct {
	Command      string
	ResponseChan chan CommandResponse
}

type PriorityCommandQueue struct {
	highPriority []CommandRequest
	lowPriority  []CommandRequest
	mutex        sync.Mutex
	cond         *sync.Cond
	closed       bool
}

func NewPriorityCommandQueue() *PriorityCommandQueue {
	pcq := &PriorityCommandQueue{
		highPriority: make([]CommandRequest, 0),
		lowPriority:  make([]CommandRequest, 0),
		closed:       false,
	}
	pcq.cond = sync.NewCond(&pcq.mutex)
	return pcq
}

// Close closes the queue and wakes up any waiting goroutines
func (pcq *PriorityCommandQueue) Close() {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	pcq.closed = true
	pcq.cond.Broadcast()
}

// EnqueueRead adds a high-priority read command to the queue and returns a response channel
func (pcq *PriorityCommandQueue) EnqueueRead(command string) <-chan CommandResponse {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()

	respChan := make(chan CommandResponse, 1)
	if pcq.closed {
		respChan <- CommandResponse{Error: fmt.Errorf("queue is closed")}
		close(respChan)
		return respChan
	}

	req := CommandRequest{
		Command:      command,
		ResponseChan: respChan,
	}

	pcq.highPriority = append(pcq.highPriority, req)
	pcq.cond.Signal()
	return respChan
}

// EnqueueControl adds a low-priority control command to the queue
func (pcq *PriorityCommandQueue) EnqueueControl(command string) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()

	if pcq.closed {
		return
	}

	req := CommandRequest{
		Command:      command,
		ResponseChan: nil, // Fire and forget
	}

	pcq.lowPriority = append(pcq.lowPriority, req)
	pcq.cond.Signal()
}

// EnqueueWithPriority adds a command with specified priority
func (pcq *PriorityCommandQueue) EnqueueWithPriority(command string, priority int) {
	switch priority {
	case PriorityHigh:
		pcq.EnqueueRead(command)
	case PriorityLow:
		pcq.EnqueueControl(command)
	default:
		// Default to low priority for unknown priority levels
		pcq.EnqueueControl(command)
	}
}

// Dequeue returns the next command request, prioritizing high-priority commands
func (pcq *PriorityCommandQueue) Dequeue() (CommandRequest, bool) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()

	// Wait for any command to be available or queue to be closed
	for len(pcq.highPriority) == 0 && len(pcq.lowPriority) == 0 && !pcq.closed {
		pcq.cond.Wait()
	}

	// If closed and empty, return false
	if pcq.closed && len(pcq.highPriority) == 0 && len(pcq.lowPriority) == 0 {
		return CommandRequest{}, false
	}

	// Process high-priority commands first
	if len(pcq.highPriority) > 0 {
		req := pcq.highPriority[0]
		pcq.highPriority = pcq.highPriority[1:]
		return req, true
	}

	// Process low-priority commands
	if len(pcq.lowPriority) > 0 {
		req := pcq.lowPriority[0]
		pcq.lowPriority = pcq.lowPriority[1:]
		return req, true
	}

	return CommandRequest{}, false
}

// Size returns the total number of queued commands
func (pcq *PriorityCommandQueue) Size() int {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	return len(pcq.highPriority) + len(pcq.lowPriority)
}

// HighPrioritySize returns the number of high-priority commands
func (pcq *PriorityCommandQueue) HighPrioritySize() int {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	return len(pcq.highPriority)
}

// LowPrioritySize returns the number of low-priority commands
func (pcq *PriorityCommandQueue) LowPrioritySize() int {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	return len(pcq.lowPriority)
}

// IsEmpty returns true if there are no queued commands
func (pcq *PriorityCommandQueue) IsEmpty() bool {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	return len(pcq.highPriority) == 0 && len(pcq.lowPriority) == 0
}

// Peek returns the next command without removing it from the queue
func (pcq *PriorityCommandQueue) Peek() (string, int, bool) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()

	if len(pcq.highPriority) > 0 {
		return pcq.highPriority[0].Command, PriorityHigh, true
	}

	if len(pcq.lowPriority) > 0 {
		return pcq.lowPriority[0].Command, PriorityLow, true
	}

	return "", 0, false
}
