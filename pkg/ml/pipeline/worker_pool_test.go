package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
)

// MockMLProcessor is a mock implementation of processors.MLProcessor for testing
type MockMLProcessor struct {
	name     string
	procType ml.ProcessorType
	running  bool
}

func NewMockMLProcessor(name string, procType ml.ProcessorType) *MockMLProcessor {
	return &MockMLProcessor{
		name:     name,
		procType: procType,
		running:  false,
	}
}

func (m *MockMLProcessor) Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error) {
	return nil, nil
}

func (m *MockMLProcessor) Name() string {
	return m.name
}

func (m *MockMLProcessor) Type() ml.ProcessorType {
	return m.procType
}

func (m *MockMLProcessor) Configure(config map[string]interface{}) error {
	return nil
}

func (m *MockMLProcessor) Start() error {
	m.running = true
	return nil
}

func (m *MockMLProcessor) Stop() error {
	m.running = false
	return nil
}

func (m *MockMLProcessor) IsRunning() bool {
	return m.running
}

func (m *MockMLProcessor) GetMetrics() ml.ProcessorStats {
	return ml.ProcessorStats{
		ProcessTime:   0,
		SuccessCount:  0,
		ErrorCount:    0,
		AvgLatency:    0,
		LastProcessed: time.Time{},
	}
}

func (m *MockMLProcessor) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// Ensure MockMLProcessor implements MLProcessor interface
var _ processors.MLProcessor = (*MockMLProcessor)(nil)

// TestNewWorker_Simple tests that NewWorker creates a worker correctly
func TestNewWorker_Simple(t *testing.T) {
	// Create a mock processor
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)

	// Create a worker with pool size 10
	poolSize := 10
	worker := NewWorker(processor, poolSize)

	// Verify worker was created
	if worker == nil {
		t.Fatal("NewWorker returned nil")
	}

	// Verify processor is set correctly
	if worker.processor != processor {
		t.Error("Worker processor not set correctly")
	}

	// Verify channels are initialized
	if worker.inputChan == nil {
		t.Error("Worker inputChan not initialized")
	}

	if worker.outputChan == nil {
		t.Error("Worker outputChan not initialized")
	}

	if worker.quit == nil {
		t.Error("Worker quit channel not initialized")
	}

	// Verify worker is not running initially
	if worker.running {
		t.Error("Worker should not be running initially")
	}

	// Verify metrics are initialized
	if worker.metrics.ProcessedCount != 0 {
		t.Error("Worker ProcessedCount should be 0 initially")
	}

	if worker.metrics.ErrorCount != 0 {
		t.Error("Worker ErrorCount should be 0 initially")
	}
}

// TestNewWorkerPool_Simple tests that NewWorkerPool creates a pool correctly
func TestNewWorkerPool_Simple(t *testing.T) {
	// Create a worker pool with max 5 workers
	maxWorkers := 5
	pool := NewWorkerPool(maxWorkers)

	// Verify pool was created
	if pool == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	// Verify workers slice is initialized
	if pool.workers == nil {
		t.Error("Pool workers slice not initialized")
	}

	// Verify jobQueue is initialized
	if pool.jobQueue == nil {
		t.Error("Pool jobQueue not initialized")
	}

	// Verify workerQueue is initialized
	if pool.workerQueue == nil {
		t.Error("Pool workerQueue not initialized")
	}

	// Verify quit channel is initialized
	if pool.quit == nil {
		t.Error("Pool quit channel not initialized")
	}

	// Verify pool is not running initially
	if pool.running {
		t.Error("Pool should not be running initially")
	}

	// Verify maxWorkers is set correctly
	if pool.maxWorkers != maxWorkers {
		t.Errorf("Pool maxWorkers should be %d, got %d", maxWorkers, pool.maxWorkers)
	}

	// Verify workers slice has correct capacity
	if cap(pool.workers) != maxWorkers {
		t.Errorf("Pool workers capacity should be %d, got %d", maxWorkers, cap(pool.workers))
	}
}
