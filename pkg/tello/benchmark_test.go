package tello

import (
	"context"
	"testing"
)

func BenchmarkNewPriorityCommandQueue(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPriorityCommandQueue()
	}
}

func BenchmarkPriorityCommandQueueEnqueueRead(b *testing.B) {
	queue := NewPriorityCommandQueue()
	commands := []string{"command", "speed?", "battery?", "time?", "height?"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		queue.EnqueueRead(cmd)
	}
}

func BenchmarkPriorityCommandQueueEnqueueControl(b *testing.B) {
	queue := NewPriorityCommandQueue()
	commands := []string{"takeoff", "land", "up 50", "down 50", "left 30"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		queue.EnqueueControl(cmd)
	}
}

func BenchmarkPriorityCommandQueueDequeue(b *testing.B) {
	queue := NewPriorityCommandQueue()

	// Pre-populate queue
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			queue.EnqueueRead("speed?")
		} else {
			queue.EnqueueControl("up 10")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Dequeue(context.Background())
	}
}

func BenchmarkPriorityCommandQueuePeek(b *testing.B) {
	queue := NewPriorityCommandQueue()

	// Pre-populate queue
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			queue.EnqueueRead("speed?")
		} else {
			queue.EnqueueControl("up 10")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Peek()
	}
}
