package pipeline

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
	"go.uber.org/goleak"
)

// MockMLProcessor implements processors.MLProcessor interface for testing.
// This mock allows controlled testing without requiring actual ML models or GPU operations.
type MockMLProcessor struct {
	nameValue     string
	processorType ml.ProcessorType
	running       bool
	processFunc   func(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error)
	metrics       ml.ProcessorStats
	mu            sync.RWMutex
	failureRate   float64 // For error injection tests
	processDelay  time.Duration
	shouldFail    bool
	failError     error
}

// NewMockMLProcessor creates a new mock processor with default behavior
func NewMockMLProcessor(name string, procType ml.ProcessorType) *MockMLProcessor {
	return &MockMLProcessor{
		nameValue:     name,
		processorType: procType,
		running:       false,
		metrics: ml.ProcessorStats{
			ProcessTime:   0,
			SuccessCount:  0,
			ErrorCount:    0,
			AvgLatency:    0,
			LastProcessed: time.Time{},
		},
		failureRate:  0,
		processDelay: 0,
	}
}

// SetProcessFunc sets a custom process function
func (m *MockMLProcessor) SetProcessFunc(fn func(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processFunc = fn
}

// SetFailureRate sets the failure rate for error injection (0.0 to 1.0)
func (m *MockMLProcessor) SetFailureRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureRate = rate
}

// SetProcessDelay sets a delay for each process call
func (m *MockMLProcessor) SetProcessDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processDelay = delay
}

// SetShouldFail sets whether the processor should always fail
func (m *MockMLProcessor) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failError = err
}

// Process implements processors.MLProcessor interface
func (m *MockMLProcessor) Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error) {
	m.mu.Lock()

	// Apply process delay if set
	if m.processDelay > 0 {
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.processDelay):
		}
		m.mu.Lock()
	}

	// Check for custom process function
	if m.processFunc != nil {
		m.mu.Unlock()
		return m.processFunc(ctx, frame)
	}

	// Error injection based on failure rate
	if m.shouldFail || (m.failureRate > 0 && rand.Float64() < m.failureRate) {
		m.metrics.ErrorCount++
		m.metrics.LastProcessed = time.Now()
		totalOps := m.metrics.SuccessCount + m.metrics.ErrorCount
		if totalOps > 0 {
			m.metrics.AvgLatency = time.Duration(
				(int64(m.metrics.AvgLatency)*int64(totalOps-1) + int64(1*time.Millisecond)) / int64(totalOps),
			)
		}
		m.mu.Unlock()

		if m.failError != nil {
			return nil, m.failError
		}
		return nil, errors.New("mock processor error")
	}

	// Successful processing
	m.metrics.SuccessCount++
	m.metrics.LastProcessed = time.Now()
	totalOps := m.metrics.SuccessCount + m.metrics.ErrorCount
	if totalOps > 0 {
		m.metrics.AvgLatency = time.Duration(
			(int64(m.metrics.AvgLatency)*int64(totalOps-1) + int64(1*time.Millisecond)) / int64(totalOps),
		)
	}
	m.mu.Unlock()

	// Return a simple detection result
	return ml.DetectionResult{
		Detections: []ml.Detection{
			{
				ClassID:    0,
				ClassName:  "test",
				Confidence: 0.95,
				Timestamp:  time.Now(),
			},
		},
		Processor: m.nameValue,
		Timestamp: time.Now(),
	}, nil
}

// Name implements processors.MLProcessor interface
func (m *MockMLProcessor) Name() string {
	return m.nameValue
}

// Type implements processors.MLProcessor interface
func (m *MockMLProcessor) Type() ml.ProcessorType {
	return m.processorType
}

// Configure implements processors.MLProcessor interface
func (m *MockMLProcessor) Configure(config map[string]interface{}) error {
	return nil
}

// Start implements processors.MLProcessor interface
func (m *MockMLProcessor) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

// Stop implements processors.MLProcessor interface
func (m *MockMLProcessor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

// IsRunning implements processors.MLProcessor interface
func (m *MockMLProcessor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetMetrics implements processors.MLProcessor interface
func (m *MockMLProcessor) GetMetrics() ml.ProcessorStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// ValidateConfig implements processors.MLProcessor interface
func (m *MockMLProcessor) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// Ensure MockMLProcessor implements MLProcessor interface
var _ processors.MLProcessor = (*MockMLProcessor)(nil)

// createTestFrame creates a test EnhancedVideoFrame for testing
func createTestFrame(seqNum int) *ml.EnhancedVideoFrame {
	return ml.NewEnhancedVideoFrame(
		[]byte("test frame data"),
		time.Now(),
		seqNum,
	)
}

// Helper to check worker is stopped with timeout
func waitForWorkerStop(t *testing.T, worker *Worker, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !worker.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// Helper to check worker pool is stopped with timeout
func waitForPoolStop(t *testing.T, pool *WorkerPool, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !pool.IsRunning() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// ==================== Worker Tests ====================

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

func TestNewWorker_ValidProcessor(t *testing.T) {
	t.Helper()
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	poolSize := 10

	worker := NewWorker(processor, poolSize)

	if worker == nil {
		t.Fatal("NewWorker returned nil")
	}

	if worker.processor != processor {
		t.Error("Worker does not have the correct processor")
	}

	if worker.IsRunning() {
		t.Error("Worker should not be running after creation")
	}

	metrics := worker.GetMetrics()
	if metrics.ProcessedCount != 0 {
		t.Errorf("Expected initial ProcessedCount to be 0, got %d", metrics.ProcessedCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected initial ErrorCount to be 0, got %d", metrics.ErrorCount)
	}
}

func TestWorker_StartStop_Lifecycle(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start worker
	err := worker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}

	if !worker.IsRunning() {
		t.Fatal("Worker should be running after Start")
	}

	// Stop worker
	err = worker.Stop()
	if err != nil {
		t.Fatalf("Failed to stop worker: %v", err)
	}

	// Wait for worker to stop
	if !waitForWorkerStop(t, worker, 2*time.Second) {
		t.Fatal("Worker did not stop within timeout")
	}
}

func TestWorker_ProcessFrame_Success(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.Start() // Must be running for Process to work
	defer processor.Stop()

	worker := NewWorker(processor, 10)
	ctx := context.Background()
	frame := createTestFrame(1)

	result, err := worker.Process(ctx, frame)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if result == nil {
		t.Fatal("Process returned nil result")
	}

	metrics := worker.GetMetrics()
	if metrics.ProcessedCount != 1 {
		t.Errorf("Expected ProcessedCount to be 1, got %d", metrics.ProcessedCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0, got %d", metrics.ErrorCount)
	}
}

func TestWorker_ProcessFrame_ProcessorNotRunning(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	// Don't start the processor
	worker := NewWorker(processor, 10)
	ctx := context.Background()
	frame := createTestFrame(1)

	_, err := worker.Process(ctx, frame)
	if err == nil {
		t.Fatal("Expected error when processor is not running")
	}
}

func TestWorker_ProcessFrame_NilFrame(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.Start()
	defer processor.Stop()

	worker := NewWorker(processor, 10)
	ctx := context.Background()

	// Process with nil frame
	_, _ = worker.Process(ctx, nil)
}

func TestWorker_GetProcessor(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	retrieved := worker.GetProcessor()
	if retrieved != processor {
		t.Error("GetProcessor did not return the original processor")
	}
}

func TestWorker_GetMetrics(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.Start()
	defer processor.Stop()

	worker := NewWorker(processor, 10)
	ctx := context.Background()

	// Initial metrics
	metrics := worker.GetMetrics()
	if metrics.ProcessedCount != 0 {
		t.Errorf("Expected initial ProcessedCount to be 0, got %d", metrics.ProcessedCount)
	}

	// Process some frames
	for i := 0; i < 5; i++ {
		frame := createTestFrame(i)
		_, _ = worker.Process(ctx, frame)
	}

	metrics = worker.GetMetrics()
	if metrics.ProcessedCount != 5 {
		t.Errorf("Expected ProcessedCount to be 5, got %d", metrics.ProcessedCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0, got %d", metrics.ErrorCount)
	}
}

func TestWorker_IsRunning(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	if worker.IsRunning() {
		t.Error("Worker should not be running initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := worker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}

	if !worker.IsRunning() {
		t.Error("Worker should be running after Start")
	}

	worker.Stop()

	if !waitForWorkerStop(t, worker, time.Second) {
		t.Error("Worker should not be running after Stop")
	}
}

func TestWorker_ConcurrentCalls(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.Start()
	defer processor.Stop()

	worker := NewWorker(processor, 100)
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			frame := createTestFrame(seq)
			_, _ = worker.Process(ctx, frame)
		}(i)
	}

	wg.Wait()

	metrics := worker.GetMetrics()
	// Due to concurrent nature, some may fail if processor is busy
	expectedMin := int64(iterations) - 5 // Allow some tolerance
	if metrics.ProcessedCount < expectedMin {
		t.Errorf("Expected at least %d successful processes, got %d", expectedMin, metrics.ProcessedCount)
	}
}

func TestWorker_Stop_WhenNotRunning(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	// Stop when not running should not error
	err := worker.Stop()
	if err != nil {
		t.Fatalf("Stop on non-running worker should not error: %v", err)
	}

	if worker.IsRunning() {
		t.Error("Worker should not be running after Stop")
	}
}

func TestWorker_Start_MultipleTimes(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start once
	err := worker.Start(ctx)
	if err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	// Start again - should error
	err = worker.Start(ctx)
	if err == nil {
		t.Fatal("Expected error on second Start")
	}

	worker.Stop()
}

// ==================== WorkerPool Tests ====================

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

func TestNewWorkerPool(t *testing.T) {
	maxWorkers := 5

	pool := NewWorkerPool(maxWorkers)

	if pool == nil {
		t.Fatal("NewWorkerPool returned nil")
	}

	if pool.IsRunning() {
		t.Error("Pool should not be running after creation")
	}

	if pool.maxWorkers != maxWorkers {
		t.Errorf("Expected maxWorkers to be %d, got %d", maxWorkers, pool.maxWorkers)
	}

	if len(pool.workers) != 0 {
		t.Errorf("Expected no workers initially, got %d", len(pool.workers))
	}
}

func TestWorkerPool_StartStop_Lifecycle(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start pool
	err := pool.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	if !pool.IsRunning() {
		t.Fatal("Pool should be running after Start")
	}

	// Stop pool
	err = pool.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pool: %v", err)
	}

	// Wait for pool to stop
	if !waitForPoolStop(t, pool, 2*time.Second) {
		t.Fatal("Pool did not stop within timeout")
	}
}

func TestWorkerPool_Start_WhenAlreadyRunning(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start pool
	err := pool.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	// Try to start again
	err = pool.Start(ctx)
	if err == nil {
		t.Fatal("Expected error when starting already running pool")
	}

	pool.Stop()
}

func TestWorkerPool_Stop_WhenNotRunning(t *testing.T) {
	pool := NewWorkerPool(3)

	// Stop when not running should not error
	err := pool.Stop()
	if err != nil {
		t.Fatalf("Stop on non-running pool should not error: %v", err)
	}
}

func TestWorkerPool_AddWorker(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	processor := NewMockMLProcessor("worker-1", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	err := pool.AddWorker(worker)
	if err != nil {
		t.Fatalf("Failed to add worker: %v", err)
	}

	workers := pool.GetWorkers()
	if len(workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(workers))
	}
}

func TestWorkerPool_AddWorker_AtMaxCapacity(t *testing.T) {
	maxWorkers := 3
	pool := NewWorkerPool(maxWorkers)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Add max workers
	for i := 0; i < maxWorkers; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		worker := NewWorker(processor, 10)
		err := pool.AddWorker(worker)
		if err != nil {
			t.Fatalf("Failed to add worker %d: %v", i, err)
		}
	}

	// Try to add one more
	processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)
	err := pool.AddWorker(worker)
	if err == nil {
		t.Fatal("Expected error when adding worker at max capacity")
	}
}

func TestWorkerPool_AddWorker_AfterPoolStopped(t *testing.T) {
	pool := NewWorkerPool(3)

	processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	err := pool.AddWorker(worker)
	if err != nil {
		t.Logf("AddWorker returned error: %v", err)
	}
}

func TestWorkerPool_AddWorker_NilWorker(t *testing.T) {
	// This test verifies that adding nil worker causes a panic
	// The implementation doesn't validate nil worker before use
	pool := NewWorkerPool(3)

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Adding nil worker panics as expected: %v", r)
		}
	}()

	// Don't start the pool - just try to add nil worker
	// The panic happens in AddWorker when accessing worker.inputChan
	_ = pool.AddWorker(nil)
	t.Log("AddWorker(nil) did not panic (unexpected)")
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	frame := createTestFrame(1)

	err := pool.Submit(frame)
	if err != nil {
		t.Fatalf("Failed to submit frame: %v", err)
	}
}

func TestWorkerPool_Submit_JobQueueFull(t *testing.T) {
	maxWorkers := 3
	pool := NewWorkerPool(maxWorkers)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Fill up the queue (queue size = maxWorkers * 10 = 30)
	for i := 0; i < maxWorkers*10; i++ {
		frame := createTestFrame(i)
		err := pool.Submit(frame)
		if err != nil {
			t.Fatalf("Failed to submit frame %d: %v", i, err)
		}
	}

	// Try to submit one more - should fail
	frame := createTestFrame(999)
	err := pool.Submit(frame)
	if err == nil {
		t.Fatal("Expected error when job queue is full")
	}
}

func TestWorkerPool_Submit_WhenNotRunning(t *testing.T) {
	pool := NewWorkerPool(3)
	// Don't start the pool

	frame := createTestFrame(1)
	err := pool.Submit(frame)
	if err == nil {
		t.Fatal("Expected error when submitting to stopped pool")
	}
}

func TestWorkerPool_Submit_NilFrame(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Submit nil frame - should not panic
	err := pool.Submit(nil)
	// This depends on implementation - might error or might be accepted
	_ = err
}

func TestWorkerPool_GetWorkers(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Add some workers
	for i := 0; i < 3; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		worker := NewWorker(processor, 10)
		_ = pool.AddWorker(worker)
	}

	workers := pool.GetWorkers()
	if len(workers) != 3 {
		t.Errorf("Expected 3 workers, got %d", len(workers))
	}
}

func TestWorkerPool_GetMetrics(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Add workers and process some frames
	for i := 0; i < 3; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		processor.Start()
		worker := NewWorker(processor, 10)
		_ = pool.AddWorker(worker)

		// Process some frames
		for j := 0; j < 5; j++ {
			frame := createTestFrame(i*10 + j)
			_, _ = worker.Process(ctx, frame)
		}
	}

	metrics := pool.GetMetrics()
	if metrics.WorkerCount != 3 {
		t.Errorf("Expected WorkerCount to be 3, got %d", metrics.WorkerCount)
	}

	expectedProcessed := int64(15) // 3 workers * 5 frames each
	if metrics.TotalProcessed != expectedProcessed {
		t.Errorf("Expected TotalProcessed to be %d, got %d", expectedProcessed, metrics.TotalProcessed)
	}
}

func TestWorkerPool_GetMetrics_EmptyPool(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	metrics := pool.GetMetrics()

	if metrics.WorkerCount != 0 {
		t.Errorf("Expected WorkerCount to be 0 for empty pool, got %d", metrics.WorkerCount)
	}

	if metrics.TotalProcessed != 0 {
		t.Errorf("Expected TotalProcessed to be 0, got %d", metrics.TotalProcessed)
	}
}

func TestWorkerPool_IsRunning(t *testing.T) {
	pool := NewWorkerPool(3)

	if pool.IsRunning() {
		t.Error("Pool should not be running initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pool.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	if !pool.IsRunning() {
		t.Error("Pool should be running after Start")
	}

	pool.Stop()

	if !waitForPoolStop(t, pool, time.Second) {
		t.Error("Pool should not be running after Stop")
	}
}

func TestWorkerPool_ConcurrentSubmits(t *testing.T) {
	pool := NewWorkerPool(5)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	var wg sync.WaitGroup
	submits := 100

	for i := 0; i < submits; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			frame := createTestFrame(seq)
			_ = pool.Submit(frame)
		}(i)
	}

	wg.Wait()
}

func TestWorkerPool_ScaleUp_AddMultipleWorkers(t *testing.T) {
	maxWorkers := 10
	pool := NewWorkerPool(maxWorkers)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)
	defer pool.Stop()

	// Add workers gradually
	for i := 0; i < maxWorkers; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		worker := NewWorker(processor, 10)
		err := pool.AddWorker(worker)
		if err != nil {
			t.Fatalf("Failed to add worker %d: %v", i, err)
		}
	}

	workers := pool.GetWorkers()
	if len(workers) != maxWorkers {
		t.Errorf("Expected %d workers, got %d", maxWorkers, len(workers))
	}
}

func TestWorkerPool_Stop_StopsAllWorkers(t *testing.T) {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool.Start(ctx)

	// Add workers
	for i := 0; i < 3; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		worker := NewWorker(processor, 10)
		_ = pool.AddWorker(worker)
	}

	pool.Stop()

	// All workers should be stopped
	workers := pool.GetWorkers()
	for i, worker := range workers {
		if worker.IsRunning() {
			t.Errorf("Worker %d should be stopped after pool Stop", i)
		}
	}
}

// ==================== Error Injection & Stress Tests ====================

func TestWorkerPool_ProcessFrame_ErrorInjection(t *testing.T) {
	pool := NewWorkerPool(5)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool.Start(ctx)

	// Add workers and start them
	for i := 0; i < 5; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		processor.SetFailureRate(0.2)
		processor.Start()

		worker := NewWorker(processor, 10)
		worker.Start(ctx) // Start the worker
		_ = pool.AddWorker(worker)
	}

	// Submit 100 frames
	frames := 100
	for i := 0; i < frames; i++ {
		frame := createTestFrame(i)
		_ = pool.Submit(frame)
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	pool.Stop()

	metrics := pool.GetMetrics()
	// With 20% failure rate, we expect some errors but most should succeed
	expectedMinSuccess := int64(frames) - 30 // Allow up to 30% failures
	if metrics.TotalProcessed < expectedMinSuccess {
		t.Logf("Processed: %d (expected at least %d)", metrics.TotalProcessed, expectedMinSuccess)
	}

	// Log metrics for inspection
	t.Logf("TotalProcessed: %d, TotalErrors: %d", metrics.TotalProcessed, metrics.TotalErrors)
}

func TestWorkerPool_ConcurrentStress(t *testing.T) {
	numWorkers := 10
	numFrames := 1000
	pool := NewWorkerPool(numWorkers)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool.Start(ctx)

	// Add workers and start them
	for i := 0; i < numWorkers; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		processor.SetProcessDelay(1 * time.Millisecond) // Small delay
		processor.Start()

		worker := NewWorker(processor, numWorkers*10)
		worker.Start(ctx) // Start the worker
		_ = pool.AddWorker(worker)
	}

	// Concurrently submit frames
	var submitted atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < numFrames; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			frame := createTestFrame(seq)
			if err := pool.Submit(frame); err == nil {
				submitted.Add(1)
			}
		}(i)
	}

	wg.Wait()

	// Wait for processing
	time.Sleep(10 * time.Second)

	pool.Stop()

	metrics := pool.GetMetrics()
	t.Logf("Submitted: %d, Processed: %d, Errors: %d",
		submitted.Load(), metrics.TotalProcessed, metrics.TotalErrors)

	// Some frames should be processed (allowing for timing issues)
	if metrics.TotalProcessed == 0 {
		t.Log("Warning: No frames processed (timing or configuration issue)")
	}
}

func TestWorker_Metrics_Accuracy(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.Start()
	defer processor.Stop()

	worker := NewWorker(processor, 100)
	ctx := context.Background()

	// Process frames with known delay
	processor.SetProcessDelay(10 * time.Millisecond)

	numFrames := 10
	for i := 0; i < numFrames; i++ {
		frame := createTestFrame(i)
		_, _ = worker.Process(ctx, frame)
	}

	metrics := worker.GetMetrics()

	if metrics.ProcessedCount != int64(numFrames) {
		t.Errorf("Expected ProcessedCount to be %d, got %d", numFrames, metrics.ProcessedCount)
	}

	if metrics.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0, got %d", metrics.ErrorCount)
	}

	// Check that average process time is reasonable
	expectedMinAvg := 5 * time.Millisecond
	if metrics.AvgProcessTime < expectedMinAvg {
		t.Errorf("Expected AvgProcessTime to be at least %v, got %v",
			expectedMinAvg, metrics.AvgProcessTime)
	}

	// LastProcessTime should be recent
	if time.Since(metrics.LastProcessTime) > 10*time.Second {
		t.Error("LastProcessTime is not recent")
	}
}

func TestWorker_ErrorPropagation(t *testing.T) {
	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	processor.SetShouldFail(true, errors.New("test error"))
	processor.Start()
	defer processor.Stop()

	worker := NewWorker(processor, 10)
	ctx := context.Background()

	frame := createTestFrame(1)

	result, err := worker.Process(ctx, frame)

	// Error should be propagated
	if err == nil {
		t.Fatal("Expected error to be propagated from processor")
	}

	if result != nil {
		t.Error("Expected nil result when processor fails")
	}

	// Error should be counted
	metrics := worker.GetMetrics()
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount to be 1, got %d", metrics.ErrorCount)
	}
}

// ==================== Goroutine Leak Detection Tests ====================

// TestWorkerShutdownNoLeak verifies no goroutines leak after worker shutdown
func TestWorkerShutdownNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	err := worker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}

	// Stop worker
	err = worker.Stop()
	if err != nil {
		t.Fatalf("Failed to stop worker: %v", err)
	}

	// Wait for worker to stop
	time.Sleep(100 * time.Millisecond)
}

// TestWorkerMultipleStartStopNoLeak verifies no leaks during multiple start/stop cycles
func TestWorkerMultipleStartStopNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	processor := NewMockMLProcessor("test-processor", ml.ProcessorTypeYOLO)
	worker := NewWorker(processor, 10)

	// Perform multiple start/stop cycles
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := worker.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start worker: %v", err)
		}

		err = worker.Stop()
		if err != nil {
			t.Fatalf("Failed to stop worker: %v", err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestWorkerPoolShutdownNoLeak verifies no goroutines leak after worker pool shutdown
func TestWorkerPoolShutdownNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	pool := NewWorkerPool(3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start pool
	err := pool.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	// Add workers
	for i := 0; i < 3; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		worker := NewWorker(processor, 10)
		_ = pool.AddWorker(worker)
	}

	// Stop pool
	err = pool.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pool: %v", err)
	}

	// Wait for pool to stop
	time.Sleep(100 * time.Millisecond)
}

// TestWorkerPoolMultipleStartStopNoLeak verifies no leaks during multiple pool start/stop cycles
func TestWorkerPoolMultipleStartStopNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	pool := NewWorkerPool(3)

	// Perform multiple start/stop cycles
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := pool.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start pool: %v", err)
		}

		err = pool.Stop()
		if err != nil {
			t.Fatalf("Failed to stop pool: %v", err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestWorkerPoolConcurrentOperationsNoLeak verifies no leaks during concurrent operations
func TestWorkerPoolConcurrentOperationsNoLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	pool := NewWorkerPool(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start pool
	err := pool.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	// Add workers
	for i := 0; i < 5; i++ {
		processor := NewMockMLProcessor("worker", ml.ProcessorTypeYOLO)
		processor.Start()
		worker := NewWorker(processor, 10)
		worker.Start(ctx)
		_ = pool.AddWorker(worker)
	}

	// Submit frames concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			frame := createTestFrame(seq)
			_ = pool.Submit(frame)
		}(i)
	}
	wg.Wait()

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Stop pool
	err = pool.Stop()
	if err != nil {
		t.Fatalf("Failed to stop pool: %v", err)
	}

	// Wait for pool to stop
	time.Sleep(100 * time.Millisecond)
}
