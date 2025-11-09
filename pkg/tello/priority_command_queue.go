package tello

import (
	"sync"
)

const (
	PriorityLow  = 0 // Control commands (takeoff, land, movement, etc.)
	PriorityHigh = 1 // Read commands (speed?, battery?, etc.)
)

type QueuedCommand struct {
	Value    string
	Priority int
}

type PriorityCommandQueue struct {
	highPriority []string
	lowPriority  []string
	mutex        sync.Mutex
	cond         *sync.Cond
}

func NewPriorityCommandQueue() *PriorityCommandQueue {
	pcq := &PriorityCommandQueue{
		highPriority: make([]string, 0),
		lowPriority:  make([]string, 0),
	}
	pcq.cond = sync.NewCond(&pcq.mutex)
	return pcq
}

// EnqueueRead adds a high-priority read command to the queue
func (pcq *PriorityCommandQueue) EnqueueRead(command string) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	pcq.highPriority = append(pcq.highPriority, command)
	pcq.cond.Signal()
}

// EnqueueControl adds a low-priority control command to the queue
func (pcq *PriorityCommandQueue) EnqueueControl(command string) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()
	pcq.lowPriority = append(pcq.lowPriority, command)
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

// Dequeue returns the next command, prioritizing high-priority commands
func (pcq *PriorityCommandQueue) Dequeue() (string, bool) {
	pcq.mutex.Lock()
	defer pcq.mutex.Unlock()

	// Wait for any command to be available
	for len(pcq.highPriority) == 0 && len(pcq.lowPriority) == 0 {
		pcq.cond.Wait()
	}

	// Process high-priority commands first
	if len(pcq.highPriority) > 0 {
		command := pcq.highPriority[0]
		pcq.highPriority = pcq.highPriority[1:]
		return command, true
	}

	// Process low-priority commands
	if len(pcq.lowPriority) > 0 {
		command := pcq.lowPriority[0]
		pcq.lowPriority = pcq.lowPriority[1:]
		return command, true
	}

	return "", false
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
		return pcq.highPriority[0], PriorityHigh, true
	}

	if len(pcq.lowPriority) > 0 {
		return pcq.lowPriority[0], PriorityLow, true
	}

	return "", 0, false
}
