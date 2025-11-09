package pipeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
)

// ConcurrentMLPipeline manages concurrent processing of video frames through multiple ML processors
type ConcurrentMLPipeline struct {
	// Core components
	frameQueue        chan *ml.EnhancedVideoFrame
	resultQueue       chan ml.MLResult
	workers           map[string]*Worker
	processorRegistry *processors.ProcessorRegistry

	// Configuration
	config           *ml.PipelineConfig
	processorConfigs []ml.ProcessorConfig

	// Synchronization
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// Metrics and monitoring
	metrics       *PipelineMetrics
	droppedFrames int64

	// State management
	running   bool
	startTime time.Time
}

// PipelineMetrics tracks performance metrics for the pipeline
type PipelineMetrics struct {
	fps            float64
	latency        time.Duration
	droppedFrames  int64
	processorStats map[string]*ml.ProcessorStats
	memoryUsage    int64
	gpuUsage       float64
	lastUpdate     time.Time
	mu             sync.RWMutex
}

// NewConcurrentMLPipeline creates a new concurrent ML pipeline
func NewConcurrentMLPipeline(config *ml.PipelineConfig, processorConfigs []ml.ProcessorConfig) *ConcurrentMLPipeline {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConcurrentMLPipeline{
		frameQueue:        make(chan *ml.EnhancedVideoFrame, config.FrameBufferSize),
		resultQueue:       make(chan ml.MLResult, config.FrameBufferSize),
		workers:           make(map[string]*Worker),
		processorRegistry: processors.NewProcessorRegistry(),
		config:            config,
		processorConfigs:  processorConfigs,
		ctx:               ctx,
		cancel:            cancel,
		metrics:           NewPipelineMetrics(),
		running:           false,
	}
}

// Start starts the ML pipeline
func (p *ConcurrentMLPipeline) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("pipeline is already running")
	}

	// Initialize processors
	if err := p.initializeProcessors(); err != nil {
		return fmt.Errorf("failed to initialize processors: %w", err)
	}

	// Start workers
	if err := p.startWorkers(); err != nil {
		return fmt.Errorf("failed to start workers: %w", err)
	}

	// Start metrics collection
	if p.config.EnableMetrics {
		p.wg.Add(1)
		go p.metricsCollector()
	}

	p.running = true
	p.startTime = time.Now()

	return nil
}

// Stop stops the ML pipeline
func (p *ConcurrentMLPipeline) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	// Signal cancellation
	p.cancel()

	// Stop all processors
	if err := p.processorRegistry.StopAll(); err != nil {
		return fmt.Errorf("failed to stop processors: %w", err)
	}

	// Wait for all workers to finish
	p.wg.Wait()

	// Close channels
	close(p.frameQueue)
	close(p.resultQueue)

	p.running = false

	return nil
}

// ProcessFrame adds a frame to the processing queue (non-blocking)
func (p *ConcurrentMLPipeline) ProcessFrame(frame *ml.EnhancedVideoFrame) error {
	if !p.running {
		return fmt.Errorf("pipeline is not running")
	}

	select {
	case p.frameQueue <- frame:
		return nil
	default:
		// Queue is full, drop frame to maintain real-time performance
		atomic.AddInt64(&p.droppedFrames, 1)
		return fmt.Errorf("frame queue is full, dropping frame")
	}
}

// GetResults returns a channel for reading ML results
func (p *ConcurrentMLPipeline) GetResults() <-chan ml.MLResult {
	return p.resultQueue
}

// GetMetrics returns current pipeline metrics
func (p *ConcurrentMLPipeline) GetMetrics() ml.PipelineMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	processorStats := make(map[string]float64)
	for name, stats := range p.metrics.processorStats {
		processorStats[name] = float64(stats.ProcessTime.Nanoseconds()) / 1e6 // Convert to milliseconds
	}

	return ml.PipelineMetrics{
		FPS:            p.metrics.fps,
		Latency:        p.metrics.latency,
		DroppedFrames:  atomic.LoadInt64(&p.droppedFrames),
		ProcessorStats: processorStats,
		MemoryUsage:    p.metrics.memoryUsage,
		GPUUsage:       p.metrics.gpuUsage,
		LastUpdate:     p.metrics.lastUpdate,
	}
}

// IsRunning returns whether the pipeline is currently running
func (p *ConcurrentMLPipeline) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// initializeProcessors creates and configures all processors
func (p *ConcurrentMLPipeline) initializeProcessors() error {
	for _, procConfig := range p.processorConfigs {
		if !procConfig.Enabled {
			continue
		}

		// Create processor
		processor, err := p.processorRegistry.CreateProcessor(procConfig.Type, procConfig.Config)
		if err != nil {
			return fmt.Errorf("failed to create processor %s: %w", procConfig.Name, err)
		}

		// Configure processor
		if err := processor.Configure(procConfig.Config); err != nil {
			return fmt.Errorf("failed to configure processor %s: %w", procConfig.Name, err)
		}

		// Start processor
		if err := processor.Start(); err != nil {
			return fmt.Errorf("failed to start processor %s: %w", procConfig.Name, err)
		}

		// Register processor
		p.processorRegistry.RegisterProcessor(procConfig.Name, processor)
	}

	return nil
}

// startWorkers creates and starts worker goroutines
func (p *ConcurrentMLPipeline) startWorkers() error {
	processorNames := p.processorRegistry.ListProcessors()

	for _, name := range processorNames {
		processor, exists := p.processorRegistry.GetProcessor(name)
		if !exists {
			continue
		}

		// Create worker for this processor
		worker := NewWorker(processor, p.config.WorkerPoolSize)
		p.workers[name] = worker

		// Start worker
		p.wg.Add(1)
		go p.runWorker(name, worker)
	}

	return nil
}

// runWorker runs a worker goroutine for a specific processor
func (p *ConcurrentMLPipeline) runWorker(processorName string, worker *Worker) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return

		case frame, ok := <-p.frameQueue:
			if !ok {
				return // Channel closed
			}

			// Process frame
			result, err := worker.Process(p.ctx, frame)
			if err != nil {
				// Log error but continue processing
				fmt.Printf("Worker %s error: %v\n", processorName, err)
				continue
			}

			// Send result
			select {
			case p.resultQueue <- result:
			case <-p.ctx.Done():
				return
			default:
				// Result queue is full, drop result
				atomic.AddInt64(&p.droppedFrames, 1)
			}
		}
	}
}

// metricsCollector collects and updates pipeline metrics
func (p *ConcurrentMLPipeline) metricsCollector() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	frameCount := int64(0)
	lastTime := time.Now()

	for {
		select {
		case <-p.ctx.Done():
			return

		case <-ticker.C:
			currentTime := time.Now()
			timeDiff := currentTime.Sub(lastTime).Seconds()

			if timeDiff > 0 {
				// Calculate FPS
				p.metrics.mu.Lock()
				p.metrics.fps = float64(frameCount) / timeDiff
				p.metrics.latency = time.Since(lastTime)
				p.metrics.droppedFrames = atomic.LoadInt64(&p.droppedFrames)
				p.metrics.lastUpdate = currentTime

				// Update processor stats
				p.metrics.processorStats = make(map[string]*ml.ProcessorStats)
				for name, processor := range p.workers {
					stats := processor.GetProcessor().GetMetrics()
					p.metrics.processorStats[name] = &stats
				}

				p.metrics.mu.Unlock()
			}

			frameCount = 0
			lastTime = currentTime
		}
	}
}

// NewPipelineMetrics creates a new pipeline metrics instance
func NewPipelineMetrics() *PipelineMetrics {
	return &PipelineMetrics{
		fps:            0,
		latency:        0,
		droppedFrames:  0,
		processorStats: make(map[string]*ml.ProcessorStats),
		memoryUsage:    0,
		gpuUsage:       0,
		lastUpdate:     time.Now(),
	}
}
