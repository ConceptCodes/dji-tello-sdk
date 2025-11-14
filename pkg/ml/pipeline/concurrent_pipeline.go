package pipeline

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/models"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/processors/yolo"
)

// ConcurrentMLPipeline manages concurrent processing of video frames through multiple ML processors
type ConcurrentMLPipeline struct {
	// Core components
	frameQueue        chan *ml.EnhancedVideoFrame
	resultQueue       chan ml.MLResult
	workers           map[string]*Worker
	processorRegistry *processors.ProcessorRegistry
	modelManager      *models.ModelManager

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

	// Performance optimizations
	framePool    sync.Pool                // Reuse frame objects
	resultPool   sync.Pool                // Reuse result objects
	batchBuffer  []*ml.EnhancedVideoFrame // Batch processing buffer
	adaptiveRate int32                    // Adaptive processing rate
}

// PipelineMetrics tracks performance metrics for the pipeline
type PipelineMetrics struct {
	fps            float64
	latency        time.Duration
	droppedFrames  int64
	frameCount     int64
	processorStats map[string]*ml.ProcessorStats
	memoryUsage    int64
	gpuUsage       float64
	lastUpdate     time.Time
	mu             sync.RWMutex
}

// NewConcurrentMLPipeline creates a new concurrent ML pipeline
func NewConcurrentMLPipeline(config *ml.PipelineConfig, processorConfigs []ml.ProcessorConfig, modelManager *models.ModelManager) *ConcurrentMLPipeline {
	ctx, cancel := context.WithCancel(context.Background())

	// Create processor registry and register factories
	registry := processors.NewProcessorRegistry()
	registry.RegisterFactory(ml.ProcessorTypeYOLO, yolo.NewYOLOFactory())

	return &ConcurrentMLPipeline{
		frameQueue:        make(chan *ml.EnhancedVideoFrame, config.FrameBufferSize),
		resultQueue:       make(chan ml.MLResult, config.FrameBufferSize),
		workers:           make(map[string]*Worker),
		processorRegistry: registry,
		modelManager:      modelManager,
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

		// Resolve model paths for YOLO processors
		enhancedConfig := p.resolveModelPaths(procConfig.Config)

		// Create processor
		processor, err := p.processorRegistry.CreateProcessor(procConfig.Type, enhancedConfig)
		if err != nil {
			return fmt.Errorf("failed to create processor %s: %w", procConfig.Name, err)
		}

		// Configure processor
		if err := processor.Configure(enhancedConfig); err != nil {
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

// resolveModelPaths resolves model names to file paths in processor configuration
func (p *ConcurrentMLPipeline) resolveModelPaths(config map[string]interface{}) map[string]interface{} {
	enhancedConfig := make(map[string]interface{})

	// Copy existing config
	for k, v := range config {
		enhancedConfig[k] = v
	}

	// Resolve model path if present
	if modelName, ok := config["model"].(string); ok && p.modelManager != nil {
		if modelPath, err := p.modelManager.GetModelPath(modelName); err == nil {
			enhancedConfig["model_path"] = modelPath
		}
	}

	// Also handle explicit model_path
	if modelPath, ok := config["model_path"].(string); ok && p.modelManager != nil {
		if resolvedPath, err := p.modelManager.GetModelPath(modelPath); err == nil {
			enhancedConfig["model_path"] = resolvedPath
		}
	}

	return enhancedConfig
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

			// Store result back into the frame
			if result != nil {
				frame.AddResult(processorName, result)
			}

			// Increment frame counter for metrics
			atomic.AddInt64(&p.metrics.frameCount, 1)

			// Send result to result queue as well
			select {
			case p.resultQueue <- result:
			case <-p.ctx.Done():
				return
			default:
				// Result queue is full, drop result
				atomic.AddInt64(&p.droppedFrames, 1)
			}

			// Increment frame counter for metrics
			atomic.AddInt64(&p.metrics.frameCount, 1)

			// Send result to result queue as well
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

	lastFrameCount := int64(0)
	lastTime := time.Now()

	for {
		select {
		case <-p.ctx.Done():
			return

		case <-ticker.C:
			currentTime := time.Now()
			timeDiff := currentTime.Sub(lastTime).Seconds()

			if timeDiff > 0 {
				// Get current frame count
				currentFrameCount := atomic.LoadInt64(&p.metrics.frameCount)
				framesInPeriod := currentFrameCount - lastFrameCount

				// Calculate FPS
				p.metrics.mu.Lock()
				p.metrics.fps = float64(framesInPeriod) / timeDiff
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

				// Update counters for next period
				lastFrameCount = currentFrameCount
				lastTime = currentTime
			}
		}
	}
}

// NewPipelineMetrics creates a new pipeline metrics instance
func NewPipelineMetrics() *PipelineMetrics {
	return &PipelineMetrics{
		fps:            0,
		latency:        0,
		droppedFrames:  0,
		frameCount:     0,
		processorStats: make(map[string]*ml.ProcessorStats),
		memoryUsage:    0,
		gpuUsage:       0,
		lastUpdate:     time.Now(),
	}
}

// ProcessFrameOptimized adds a frame with adaptive rate control and pooling
func (p *ConcurrentMLPipeline) ProcessFrameOptimized(frame *ml.EnhancedVideoFrame) error {
	if !p.running {
		return fmt.Errorf("pipeline is not running")
	}

	// Adaptive rate control based on current FPS
	currentFPS := p.getCurrentFPS()
	targetFPS := float64(p.config.TargetFPS)

	if currentFPS > targetFPS*1.1 { // 10% tolerance
		// Skip frames to maintain target FPS
		return nil
	}

	// Try to reuse frame from pool
	reusedFrame := p.framePool.Get()
	if reusedFrame != nil {
		// Copy data to reused frame
		reusedFrame.(*ml.EnhancedVideoFrame).Data = frame.Data
		reusedFrame.(*ml.EnhancedVideoFrame).Timestamp = frame.Timestamp
		reusedFrame.(*ml.EnhancedVideoFrame).SeqNum = frame.SeqNum
		frame = reusedFrame.(*ml.EnhancedVideoFrame)
	}

	select {
	case p.frameQueue <- frame:
		return nil
	default:
		// Queue is full, drop frame to maintain real-time performance
		atomic.AddInt64(&p.droppedFrames, 1)
		if reusedFrame != nil {
			p.framePool.Put(reusedFrame)
		}
		return fmt.Errorf("frame queue is full, dropping frame")
	}
}

// ProcessBatch processes multiple frames in batch for better throughput
func (p *ConcurrentMLPipeline) ProcessBatch(frames []*ml.EnhancedVideoFrame) error {
	if !p.running {
		return fmt.Errorf("pipeline is not running")
	}

	if len(frames) == 0 {
		return nil
	}

	// Adaptive batch size based on performance
	batchSize := p.calculateOptimalBatchSize(len(frames))
	if batchSize > len(frames) {
		batchSize = len(frames)
	}

	// Process batch
	for i := 0; i < batchSize; i++ {
		frame := frames[i]

		select {
		case p.frameQueue <- frame:
		default:
			// Queue is full, drop remaining frames
			atomic.AddInt64(&p.droppedFrames, int64(batchSize-i))
			return fmt.Errorf("frame queue is full, dropping %d frames", batchSize-i)
		}
	}

	return nil
}

// getCurrentFPS returns the current processing FPS
func (p *ConcurrentMLPipeline) getCurrentFPS() float64 {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()
	return p.metrics.fps
}

// calculateOptimalBatchSize determines optimal batch size based on current performance
func (p *ConcurrentMLPipeline) calculateOptimalBatchSize(requestedSize int) int {
	currentFPS := p.getCurrentFPS()
	targetFPS := float64(p.config.TargetFPS)

	// If we're below target FPS, use smaller batches
	if currentFPS < targetFPS*0.8 {
		return 1 // Process frames individually
	}

	// If we're at or above target, we can use larger batches
	if currentFPS >= targetFPS {
		if requestedSize <= 4 {
			return requestedSize
		}
		return 4 // Cap at 4 for real-time performance
	}

	// Default case
	if requestedSize <= 2 {
		return requestedSize
	}
	return 2
}

// OptimizeForPerformance applies performance optimizations
func (p *ConcurrentMLPipeline) OptimizeForPerformance() {
	// Initialize object pools
	p.framePool = sync.Pool{
		New: func() interface{} {
			return &ml.EnhancedVideoFrame{
				MLResults: make(map[string]ml.MLResult),
			}
		},
	}

	p.resultPool = sync.Pool{
		New: func() interface{} {
			return &ml.DetectionResult{}
		},
	}

	// Pre-allocate batch buffer
	p.batchBuffer = make([]*ml.EnhancedVideoFrame, 0, 8)

	// Set adaptive processing rate
	atomic.StoreInt32(&p.adaptiveRate, 1)
}

// AdaptiveRateControl adjusts processing rate based on performance
func (p *ConcurrentMLPipeline) AdaptiveRateControl() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			currentFPS := p.getCurrentFPS()
			targetFPS := float64(p.config.TargetFPS)

			var newRate int32
			switch {
			case currentFPS < targetFPS*0.7:
				newRate = 1 // Process every frame
			case currentFPS < targetFPS*0.9:
				newRate = 2 // Process every 2nd frame
			case currentFPS > targetFPS*1.2:
				newRate = 3 // Process every 3rd frame
			default:
				newRate = 2 // Default to every 2nd frame
			}

			atomic.StoreInt32(&p.adaptiveRate, newRate)
		}
	}
}

// GetPerformanceStats returns detailed performance statistics
func (p *ConcurrentMLPipeline) GetPerformanceStats() map[string]interface{} {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	stats := map[string]interface{}{
		"current_fps":     p.metrics.fps,
		"target_fps":      p.config.TargetFPS,
		"latency_ms":      p.metrics.latency.Milliseconds(),
		"dropped_frames":  atomic.LoadInt64(&p.droppedFrames),
		"queue_size":      len(p.frameQueue),
		"queue_capacity":  cap(p.frameQueue),
		"adaptive_rate":   atomic.LoadInt32(&p.adaptiveRate),
		"memory_usage_mb": p.metrics.memoryUsage / 1024 / 1024,
		"gpu_usage":       p.metrics.gpuUsage,
		"processor_count": len(p.workers),
		"uptime_seconds":  time.Since(p.startTime).Seconds(),
	}

	// Add per-processor stats
	processorStats := make(map[string]interface{})
	for name, worker := range p.workers {
		procMetrics := worker.GetProcessor().GetMetrics()
		processorStats[name] = map[string]interface{}{
			"process_time_ms": procMetrics.ProcessTime.Milliseconds(),
			"success_count":   procMetrics.SuccessCount,
			"error_count":     procMetrics.ErrorCount,
			"avg_latency_ms":  procMetrics.AvgLatency.Milliseconds(),
		}
	}
	stats["processors"] = processorStats

	return stats
}
