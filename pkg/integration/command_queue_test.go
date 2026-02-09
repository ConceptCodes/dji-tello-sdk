package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type SimpleCommandQueue struct {
	commands []string
	mutex    sync.Mutex
	cond     *sync.Cond
	closed   bool
}

func NewSimpleCommandQueue() *SimpleCommandQueue {
	cq := &SimpleCommandQueue{
		commands: make([]string, 0),
		closed:   false,
	}
	cq.cond = sync.NewCond(&cq.mutex)
	return cq
}

func (cq *SimpleCommandQueue) Enqueue(command string) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	if cq.closed {
		return
	}

	cq.commands = append(cq.commands, command)
	cq.cond.Signal()
}

func (cq *SimpleCommandQueue) Dequeue() (string, bool) {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()

	for len(cq.commands) == 0 && !cq.closed {
		cq.cond.Wait()
	}

	if cq.closed && len(cq.commands) == 0 {
		return "", false
	}

	if len(cq.commands) > 0 {
		cmd := cq.commands[0]
		cq.commands = cq.commands[1:]
		return cmd, true
	}

	return "", false
}

func (cq *SimpleCommandQueue) Close() {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()
	cq.closed = true
	cq.cond.Broadcast()
}

func (cq *SimpleCommandQueue) Size() int {
	cq.mutex.Lock()
	defer cq.mutex.Unlock()
	return len(cq.commands)
}

func TestCommandQueue_BasicEnqueue(t *testing.T) {
	queue := NewSimpleCommandQueue()
	defer queue.Close()

	queue.Enqueue("command1")
	queue.Enqueue("command2")
	queue.Enqueue("command3")

	if size := queue.Size(); size != 3 {
		t.Errorf("Expected queue size 3, got %d", size)
	}

	cmd1, ok1 := queue.Dequeue()
	if !ok1 {
		t.Error("Expected to dequeue command1, but got false")
	}
	if cmd1 != "command1" {
		t.Errorf("Expected 'command1', got '%s'", cmd1)
	}

	cmd2, ok2 := queue.Dequeue()
	if !ok2 {
		t.Error("Expected to dequeue command2, but got false")
	}
	if cmd2 != "command2" {
		t.Errorf("Expected 'command2', got '%s'", cmd2)
	}

	cmd3, ok3 := queue.Dequeue()
	if !ok3 {
		t.Error("Expected to dequeue command3, but got false")
	}
	if cmd3 != "command3" {
		t.Errorf("Expected 'command3', got '%s'", cmd3)
	}

	if size := queue.Size(); size != 0 {
		t.Errorf("Expected queue size 0, got %d", size)
	}
}

func TestCommandQueue_ConcurrentAccess(t *testing.T) {
	queue := NewSimpleCommandQueue()
	defer queue.Close()

	const numProducers = 5
	const numConsumers = 3
	const commandsPerProducer = 100

	var wg sync.WaitGroup
	dequeuedCommands := make([]string, 0)
	var dequeuedMutex sync.Mutex

	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < commandsPerProducer; j++ {
				cmd := fmt.Sprintf("producer%d_cmd%d", producerID, j)
				queue.Enqueue(cmd)
			}
		}(i)
	}

	consumerWg := sync.WaitGroup{}
	for i := 0; i < numConsumers; i++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for {
				cmd, ok := queue.Dequeue()
				if !ok {
					return
				}
				dequeuedMutex.Lock()
				dequeuedCommands = append(dequeuedCommands, cmd)
				dequeuedMutex.Unlock()
			}
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	queue.Close()
	consumerWg.Wait()

	expectedTotal := numProducers * commandsPerProducer
	dequeuedMutex.Lock()
	actualTotal := len(dequeuedCommands)
	dequeuedMutex.Unlock()

	if actualTotal != expectedTotal {
		t.Errorf("Expected %d commands, got %d", expectedTotal, actualTotal)
	}
}
