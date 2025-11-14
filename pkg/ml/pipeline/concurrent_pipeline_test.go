package pipeline

import (
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/models"
	"github.com/stretchr/testify/assert"
)

func TestNewConcurrentMLPipeline(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         100,
		WorkerPoolSize:          4,
		EnableMetrics:           true,
		TargetFPS:               30,
	}

	processorConfigs := []ml.ProcessorConfig{
		{
			Name:    "test_processor",
			Type:    ml.ProcessorTypeYOLO,
			Enabled: true,
			Config:  map[string]interface{}{"threshold": 0.5},
		},
	}

	// Create a mock model manager for testing
	modelManager, _ := models.NewModelManager(t.TempDir())

	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	assert.NotNil(t, pipeline)
	assert.Equal(t, config, pipeline.config)
	assert.Equal(t, processorConfigs, pipeline.processorConfigs)
	assert.False(t, pipeline.IsRunning())
}

func TestConcurrentMLPipeline_StartStop(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         100,
		WorkerPoolSize:          4,
		EnableMetrics:           false, // Disable metrics for simpler test
		TargetFPS:               30,
	}

	// Use empty processor configs to avoid factory registration issues
	processorConfigs := []ml.ProcessorConfig{}

	// Create a mock model manager for testing
	modelManager, _ := models.NewModelManager(t.TempDir())

	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Test start with no processors (should succeed)
	err := pipeline.Start()
	assert.NoError(t, err)
	assert.True(t, pipeline.IsRunning())

	// Test double start
	err = pipeline.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop
	err = pipeline.Stop()
	assert.NoError(t, err)
	assert.False(t, pipeline.IsRunning())
}

func TestConcurrentMLPipeline_ProcessFrame(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         1, // Small buffer for testing
		WorkerPoolSize:          1,
		EnableMetrics:           false,
		TargetFPS:               30,
	}

	processorConfigs := []ml.ProcessorConfig{} // No processors for this test

	modelManager, _ := models.NewModelManager(t.TempDir())
	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Test processing when not running
	frame := ml.NewEnhancedVideoFrame([]byte("test"), time.Now(), 1)
	err := pipeline.ProcessFrame(frame)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start pipeline
	err = pipeline.Start()
	assert.NoError(t, err)

	// Test processing when running
	err = pipeline.ProcessFrame(frame)
	assert.NoError(t, err)

	// Stop pipeline
	err = pipeline.Stop()
	assert.NoError(t, err)
}

func TestConcurrentMLPipeline_GetResults(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         100,
		WorkerPoolSize:          4,
		EnableMetrics:           false,
		TargetFPS:               30,
	}

	processorConfigs := []ml.ProcessorConfig{}

	modelManager, _ := models.NewModelManager(t.TempDir())
	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Get results channel
	resultsChan := pipeline.GetResults()
	assert.NotNil(t, resultsChan)
}

func TestConcurrentMLPipeline_GetMetrics(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         100,
		WorkerPoolSize:          4,
		EnableMetrics:           false,
		TargetFPS:               30,
	}

	processorConfigs := []ml.ProcessorConfig{}

	modelManager, _ := models.NewModelManager(t.TempDir())
	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Get metrics
	metrics := pipeline.GetMetrics()
	assert.Equal(t, float64(0), metrics.FPS)
	assert.Equal(t, int64(0), metrics.DroppedFrames)
	assert.NotNil(t, metrics.ProcessorStats)
}

func TestNewPipelineMetrics(t *testing.T) {
	metrics := NewPipelineMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, float64(0), metrics.fps)
	assert.Equal(t, time.Duration(0), metrics.latency)
	assert.Equal(t, int64(0), metrics.droppedFrames)
	assert.NotNil(t, metrics.processorStats)
	assert.Equal(t, int64(0), metrics.memoryUsage)
	assert.Equal(t, float64(0), metrics.gpuUsage)
}

func TestConcurrentMLPipeline_Integration(t *testing.T) {
	// This is a more comprehensive integration test
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         10,
		WorkerPoolSize:          2,
		EnableMetrics:           true,
		TargetFPS:               30,
	}

	// Use empty processor configs to avoid factory registration issues
	processorConfigs := []ml.ProcessorConfig{}

	modelManager, _ := models.NewModelManager(t.TempDir())
	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Start pipeline
	err := pipeline.Start()
	assert.NoError(t, err)
	assert.True(t, pipeline.IsRunning())

	// Process some frames
	for i := 0; i < 5; i++ {
		frame := ml.NewEnhancedVideoFrame([]byte("test data"), time.Now(), i)
		err := pipeline.ProcessFrame(frame)
		assert.NoError(t, err)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Get metrics
	metrics := pipeline.GetMetrics()
	assert.NotNil(t, metrics)

	// Stop pipeline
	err = pipeline.Stop()
	assert.NoError(t, err)
	assert.False(t, pipeline.IsRunning())
}
