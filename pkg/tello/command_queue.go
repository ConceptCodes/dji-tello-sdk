package tello

import (
	"sync"
)

type CommandQueue struct {
	commands []string
	mutex    sync.Mutex
	cond     *sync.Cond
}

func NewCommandQueue() *CommandQueue {
	cq := &CommandQueue{
		commands: make([]string, 0),
	}
	cq.cond = sync.NewCond(&cq.mutex)
	return cq
}

func (cq *CommandQueue) Enqueue(command string) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()
	cq.commands = append(cq.commands, command)
	cq.cond.Signal()
}

func (cq *CommandQueue) Dequeue() (string, bool) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	for len(cq.commands) == 0 {
		cq.cond.Wait()
	}

	if len(cq.commands) == 0 {
		return "", false
	}

	command := cq.commands[0]
	cq.commands = cq.commands[1:]
	return command, true
}

func (cq *CommandQueue) Size() int {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()
	return len(cq.commands)
}

func (cq *CommandQueue) IsEmpty() bool {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()
	return len(cq.commands) == 0
}
