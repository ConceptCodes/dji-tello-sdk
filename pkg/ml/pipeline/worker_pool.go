package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
)

// WorkerResult holds the result of a worker's processing
type WorkerResult struct {
	Frame  *ml.EnhancedVideoFrame
	Result ml.MLResult
	Error  error
}

// Worker represents a worker that processes frames using a specific ML processor
type Worker struct {
	processor  processors.MLProcessor
	workerPool *WorkerPool
	inputChan  chan *ml.EnhancedVideoFrame
	outputChan chan WorkerResult
	mu         sync.RWMutex
	running    bool
	metrics    WorkerMetrics
}

// WorkerMetrics tracks metrics for a specific worker
type WorkerMetrics struct {
	ProcessedCount  int64
	ErrorCount      int64
	AvgProcessTime  time.Duration
	LastProcessTime time.Time
	mu              sync.RWMutex
}

// WorkerPool manages a pool of workers for concurrent processing
type WorkerPool struct {
	workers     []*Worker
	jobQueue    chan *ml.EnhancedVideoFrame
	workerQueue chan chan *ml.EnhancedVideoFrame
	mu          sync.RWMutex
	running     bool
	maxWorkers  int
	wg          sync.WaitGroup
	done        chan struct{}
}

// NewWorker creates a new worker with the given processor
func NewWorker(processor processors.MLProcessor, poolSize int) *Worker {
	return &Worker{
		processor:  processor,
		inputChan:  make(chan *ml.EnhancedVideoFrame, poolSize),
		outputChan: make(chan WorkerResult, poolSize),
		running:    false,
		metrics: WorkerMetrics{
			ProcessedCount:  0,
			ErrorCount:      0,
			AvgProcessTime:  0,
			LastProcessTime: time.Time{},
		},
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		return fmt.Errorf("worker is already running")
	}

	w.running = true
	go w.run(ctx)

	return nil
}

// Stop stops the worker
func (w *Worker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}
	w.running = false

	// Context cancellation is handled by the caller
	// Don't close input/output channels here - let the run goroutine handle it
	// to avoid panic from concurrent reads/writes

	return nil
}

// Process processes a frame and returns the result
func (w *Worker) Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error) {
	if !w.processor.IsRunning() {
		return nil, fmt.Errorf("processor is not running")
	}

	startTime := time.Now()

	// Process the frame
	result, err := w.processor.Process(ctx, frame)

	processTime := time.Since(startTime)

	// Update metrics
	w.updateMetrics(processTime, err == nil)

	return result, err
}

// GetProcessor returns the processor associated with this worker
func (w *Worker) GetProcessor() processors.MLProcessor {
	return w.processor
}

// GetMetrics returns the worker's metrics
func (w *Worker) GetMetrics() WorkerMetrics {
	w.metrics.mu.RLock()
	defer w.metrics.mu.RUnlock()

	// Return a copy to avoid returning the lock
	return WorkerMetrics{
		ProcessedCount:  w.metrics.ProcessedCount,
		ErrorCount:      w.metrics.ErrorCount,
		AvgProcessTime:  w.metrics.AvgProcessTime,
		LastProcessTime: w.metrics.LastProcessTime,
	}
}

// IsRunning returns whether the worker is currently running
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// run is the main worker loop
func (w *Worker) run(ctx context.Context) {
	defer func() {
		// Safely close channels when goroutine exits
		close(w.inputChan)
		close(w.outputChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-w.inputChan:
			if !ok {
				// Channel closed, exit
				return
			}
			if frame == nil {
				continue
			}

			// Process the frame
			result, err := w.Process(ctx, frame)
			if err != nil {
				fmt.Printf("Worker processing error: %v\n", err)
				// Still send result with error
			}

			// Send result
			select {
			case w.outputChan <- WorkerResult{Frame: frame, Result: result, Error: err}:
			case <-ctx.Done():
				return
			default:
				// Output channel is full, drop result
			}
		}
	}
}

// updateMetrics updates the worker's metrics
func (w *Worker) updateMetrics(processTime time.Duration, success bool) {
	w.metrics.mu.Lock()
	defer w.metrics.mu.Unlock()

	w.metrics.LastProcessTime = time.Now()

	if success {
		w.metrics.ProcessedCount++
	} else {
		w.metrics.ErrorCount++
	}

	// Update average process time
	totalOps := w.metrics.ProcessedCount + w.metrics.ErrorCount
	if totalOps > 0 {
		w.metrics.AvgProcessTime = time.Duration(
			(int64(w.metrics.AvgProcessTime)*int64(totalOps-1) + int64(processTime)) / int64(totalOps),
		)
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		workers:     make([]*Worker, 0, maxWorkers),
		jobQueue:    make(chan *ml.EnhancedVideoFrame, maxWorkers*10),
		workerQueue: make(chan chan *ml.EnhancedVideoFrame, maxWorkers),
		running:     false,
		maxWorkers:  maxWorkers,
		done:        make(chan struct{}),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return fmt.Errorf("worker pool is already running")
	}

	wp.running = true

	// Recreate done channel if it was closed
	select {
	case <-wp.done:
		wp.done = make(chan struct{})
	default:
	}

	// Start dispatcher
	wp.wg.Add(1)
	go wp.dispatcher(ctx)

	return nil
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()

	if !wp.running {
		wp.mu.Unlock()
		return nil
	}

	wp.running = false

	// Stop all workers
	for _, worker := range wp.workers {
		worker.Stop()
	}

	// Signal dispatcher to exit
	close(wp.done)

	wp.mu.Unlock()

	// Wait for dispatcher to finish before closing channels
	wp.wg.Wait()

	// Close channels
	close(wp.jobQueue)
	close(wp.workerQueue)

	return nil
}

// AddWorker adds a worker to the pool
func (wp *WorkerPool) AddWorker(worker *Worker) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.workers) >= wp.maxWorkers {
		return fmt.Errorf("worker pool is at maximum capacity")
	}

	wp.workers = append(wp.workers, worker)

	// Add worker's input channel to worker queue
	wp.workerQueue <- worker.inputChan

	return nil
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(frame *ml.EnhancedVideoFrame) error {
	if !wp.running {
		return fmt.Errorf("worker pool is not running")
	}

	select {
	case wp.jobQueue <- frame:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetWorkers returns all workers in the pool
func (wp *WorkerPool) GetWorkers() []*Worker {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	workers := make([]*Worker, len(wp.workers))
	copy(workers, wp.workers)
	return workers
}

// GetMetrics returns aggregated metrics for all workers
func (wp *WorkerPool) GetMetrics() PoolMetrics {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	metrics := PoolMetrics{
		WorkerCount:     len(wp.workers),
		TotalProcessed:  0,
		TotalErrors:     0,
		AvgProcessTime:  0,
		LastProcessTime: time.Time{},
	}

	var totalProcessTime int64
	workerCount := 0

	for _, worker := range wp.workers {
		workerMetrics := worker.GetMetrics()
		metrics.TotalProcessed += workerMetrics.ProcessedCount
		metrics.TotalErrors += workerMetrics.ErrorCount

		if workerMetrics.LastProcessTime.After(metrics.LastProcessTime) {
			metrics.LastProcessTime = workerMetrics.LastProcessTime
		}

		totalProcessTime += int64(workerMetrics.AvgProcessTime)
		if workerMetrics.ProcessedCount > 0 || workerMetrics.ErrorCount > 0 {
			workerCount++
		}
	}

	if workerCount > 0 {
		metrics.AvgProcessTime = time.Duration(totalProcessTime / int64(workerCount))
	}

	return metrics
}

// IsRunning returns whether the worker pool is running
func (wp *WorkerPool) IsRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.running
}

// dispatcher distributes jobs to available workers
func (wp *WorkerPool) dispatcher(ctx context.Context) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-wp.done:
			return
		case job := <-wp.jobQueue:
			if job == nil {
				continue
			}

			// Get a worker channel
			var workerChan chan *ml.EnhancedVideoFrame
			select {
			case workerChan = <-wp.workerQueue:
			case <-ctx.Done():
				return
			case <-wp.done:
				return
			}

			// Send job to worker
			select {
			case workerChan <- job:
				// Job sent successfully
			case <-ctx.Done():
				return
			case <-wp.done:
				return
			default:
				// Worker is busy, put job back in queue
				go func() {
					select {
					case wp.jobQueue <- job:
					case <-ctx.Done():
					case <-wp.done:
					}
				}()
			}

			// Put worker channel back in queue
			select {
			case wp.workerQueue <- workerChan:
			case <-ctx.Done():
				return
			case <-wp.done:
				return
			}
		}
	}
}

// PoolMetrics represents aggregated metrics for the worker pool
type PoolMetrics struct {
	WorkerCount     int
	TotalProcessed  int64
	TotalErrors     int64
	AvgProcessTime  time.Duration
	LastProcessTime time.Time
}
