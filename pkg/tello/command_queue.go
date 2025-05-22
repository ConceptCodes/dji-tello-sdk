package tello

import (
	"time"
)

type CommandQueue struct {
	queue          []string
	lastCmdTime    time.Time
	minCmdInterval time.Duration
}

func NewCommandQueue() *CommandQueue {
	return &CommandQueue{
		queue:          make([]string, 0),
		minCmdInterval: time.Second * 1,
	}
}

func (cq *CommandQueue) Enqueue(command string) {
	cq.queue = append(cq.queue, command)
}

func (cq *CommandQueue) Dequeue() (string, bool) {
	if len(cq.queue) == 0 {
		return "", false
	}

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
	return len(cq.queue) == 0
}

func (cq *CommandQueue) Size() int {
	return len(cq.queue)
}
