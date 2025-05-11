package tello

import (
	"sync"
	"time"
)

type CommandQueue struct {
	mu             sync.Mutex
	queue          []string
	lastCmdTime    time.Time
	minCmdInterval time.Duration
}

func NewCommandQueue() *CommandQueue {
	return &CommandQueue{
		queue:          make([]string, 0),
		minCmdInterval: time.Microsecond, // 1MHz rate limit
	}
}

func (cq *CommandQueue) Enqueue(command string) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	cq.queue = append(cq.queue, command)
}

func (cq *CommandQueue) Dequeue() (string, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.queue) == 0 {
		return "", false
	}

	// Wait if the time since the last command is less than the minimum interval
	if !cq.lastCmdTime.IsZero() {
		elapsed := time.Since(cq.lastCmdTime)
		if elapsed < cq.minCmdInterval {
			time.Sleep(cq.minCmdInterval - elapsed)
		}
	}

	command := cq.queue[0]
	cq.queue = cq.queue[1:]
	cq.lastCmdTime = time.Now()
	return command, true
}

func (cq *CommandQueue) IsEmpty() bool {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return len(cq.queue) == 0
}

func (cq *CommandQueue) Size() int {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return len(cq.queue)
}
