package processors

import (
	"context"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// MLProcessor defines the interface for all ML processors
type MLProcessor interface {
	// Process processes a video frame and returns ML results
	Process(ctx context.Context, frame *ml.EnhancedVideoFrame) (ml.MLResult, error)

	// Name returns the processor name
	Name() string

	// Type returns the processor type
	Type() ml.ProcessorType

	// Configure configures the processor with given configuration
	Configure(config map[string]interface{}) error

	// Start starts the processor (initialize resources, load models, etc.)
	Start() error

	// Stop stops the processor and releases resources
	Stop() error

	// IsRunning returns whether the processor is currently running
	IsRunning() bool

	// GetMetrics returns processor-specific metrics
	GetMetrics() ml.ProcessorStats

	// ValidateConfig validates the processor configuration
	ValidateConfig(config map[string]interface{}) error
}

// ProcessorFactory defines the interface for creating processors
type ProcessorFactory interface {
	// CreateProcessor creates a new processor instance
	CreateProcessor(config map[string]interface{}) (MLProcessor, error)

	// GetProcessorType returns the processor type this factory creates
	GetProcessorType() ml.ProcessorType

	// GetDefaultConfig returns default configuration for this processor
	GetDefaultConfig() map[string]interface{}
}

// ProcessorRegistry manages processor registration and creation
type ProcessorRegistry struct {
	factories  map[ml.ProcessorType]ProcessorFactory
	processors map[string]MLProcessor
}

// NewProcessorRegistry creates a new processor registry
func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		factories:  make(map[ml.ProcessorType]ProcessorFactory),
		processors: make(map[string]MLProcessor),
	}
}

// RegisterFactory registers a processor factory
func (pr *ProcessorRegistry) RegisterFactory(processorType ml.ProcessorType, factory ProcessorFactory) {
	pr.factories[processorType] = factory
}

// CreateProcessor creates a new processor of the given type
func (pr *ProcessorRegistry) CreateProcessor(processorType ml.ProcessorType, config map[string]interface{}) (MLProcessor, error) {
	factory, exists := pr.factories[processorType]
	if !exists {
		return nil, &ProcessorError{
			Type:    "UnknownProcessorType",
			Message: "No factory registered for processor type: " + string(processorType),
		}
	}

	return factory.CreateProcessor(config)
}

// GetProcessorTypes returns all registered processor types
func (pr *ProcessorRegistry) GetProcessorTypes() []ml.ProcessorType {
	types := make([]ml.ProcessorType, 0, len(pr.factories))
	for processorType := range pr.factories {
		types = append(types, processorType)
	}
	return types
}

// GetDefaultConfig returns default configuration for a processor type
func (pr *ProcessorRegistry) GetDefaultConfig(processorType ml.ProcessorType) (map[string]interface{}, error) {
	factory, exists := pr.factories[processorType]
	if !exists {
		return nil, &ProcessorError{
			Type:    "UnknownProcessorType",
			Message: "No factory registered for processor type: " + string(processorType),
		}
	}

	return factory.GetDefaultConfig(), nil
}

// RegisterProcessor registers a processor instance by name
func (pr *ProcessorRegistry) RegisterProcessor(name string, processor MLProcessor) {
	pr.processors[name] = processor
}

// GetProcessor gets a registered processor by name
func (pr *ProcessorRegistry) GetProcessor(name string) (MLProcessor, bool) {
	processor, exists := pr.processors[name]
	return processor, exists
}

// ListProcessors returns all registered processor names
func (pr *ProcessorRegistry) ListProcessors() []string {
	names := make([]string, 0, len(pr.processors))
	for name := range pr.processors {
		names = append(names, name)
	}
	return names
}

// StopAll stops all registered processors
func (pr *ProcessorRegistry) StopAll() error {
	var errors []error

	for name, processor := range pr.processors {
		if processor.IsRunning() {
			if err := processor.Stop(); err != nil {
				errors = append(errors, &ProcessorError{
					Type:    "StopError",
					Message: "Failed to stop processor " + name + ": " + err.Error(),
				})
			}
		}
	}

	if len(errors) > 0 {
		return &ProcessorError{
			Type:    "MultipleErrors",
			Message: "Multiple errors occurred while stopping processors",
		}
	}

	return nil
}

// ProcessorError represents an error in processor operations
type ProcessorError struct {
	Type    string
	Message string
}

// Error implements error interface
func (pe *ProcessorError) Error() string {
	return pe.Type + ": " + pe.Message
}

// BaseProcessor provides common functionality for all processors
type BaseProcessor struct {
	name          string
	processorType ml.ProcessorType
	config        map[string]interface{}
	running       bool
	startTime     time.Time
	metrics       ml.ProcessorStats
}

// NewBaseProcessor creates a new base processor
func NewBaseProcessor(name string, processorType ml.ProcessorType) *BaseProcessor {
	return &BaseProcessor{
		name:          name,
		processorType: processorType,
		config:        make(map[string]interface{}),
		running:       false,
		metrics: ml.ProcessorStats{
			ProcessTime:   0,
			SuccessCount:  0,
			ErrorCount:    0,
			AvgLatency:    0,
			LastProcessed: time.Time{},
		},
	}
}

// Name returns the processor name
func (bp *BaseProcessor) Name() string {
	return bp.name
}

// Type returns the processor type
func (bp *BaseProcessor) Type() ml.ProcessorType {
	return bp.processorType
}

// Configure configures the processor
func (bp *BaseProcessor) Configure(config map[string]interface{}) error {
	bp.config = config
	return nil
}

// GetConfig returns the current configuration
func (bp *BaseProcessor) GetConfig() map[string]interface{} {
	return bp.config
}

// Start marks the processor as running
func (bp *BaseProcessor) Start() error {
	bp.running = true
	bp.startTime = time.Now()
	return nil
}

// Stop marks the processor as stopped
func (bp *BaseProcessor) Stop() error {
	bp.running = false
	return nil
}

// IsRunning returns whether the processor is running
func (bp *BaseProcessor) IsRunning() bool {
	return bp.running
}

// GetMetrics returns processor metrics
func (bp *BaseProcessor) GetMetrics() ml.ProcessorStats {
	return bp.metrics
}

// UpdateMetrics updates processor metrics
func (bp *BaseProcessor) UpdateMetrics(processTime time.Duration, success bool) {
	bp.metrics.ProcessTime = processTime
	bp.metrics.LastProcessed = time.Now()

	if success {
		bp.metrics.SuccessCount++
	} else {
		bp.metrics.ErrorCount++
	}

	// Calculate average latency
	totalOps := bp.metrics.SuccessCount + bp.metrics.ErrorCount
	if totalOps > 0 {
		bp.metrics.AvgLatency = time.Duration(
			(int64(bp.metrics.AvgLatency)*int64(totalOps-1) + int64(processTime)) / int64(totalOps),
		)
	}
}

// ValidateConfig performs basic configuration validation
func (bp *BaseProcessor) ValidateConfig(config map[string]interface{}) error {
	// Base validation - can be overridden by specific processors
	if config == nil {
		return &ProcessorError{
			Type:    "InvalidConfig",
			Message: "Configuration cannot be nil",
		}
	}
	return nil
}

// GetUptime returns how long the processor has been running
func (bp *BaseProcessor) GetUptime() time.Duration {
	if !bp.running {
		return 0
	}
	return time.Since(bp.startTime)
}
