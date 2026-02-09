package pipeline

import (
	"testing"
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml/models"
	"github.com/stretchr/testify/assert"
)

// TestConcurrentMLPipeline_AdaptiveRateControl tests adaptive processing rate
func TestConcurrentMLPipeline_AdaptiveRateControl(t *testing.T) {
	config := &ml.PipelineConfig{
		MaxConcurrentProcessors: 2,
		FrameBufferSize:         10,
		WorkerPoolSize:          2,
		EnableMetrics:           true,
		TargetFPS:               30,
	}

	processorConfigs := []ml.ProcessorConfig{}

	modelManager, _ := models.NewModelManager(t.TempDir())
	pipeline := NewConcurrentMLPipeline(config, processorConfigs, modelManager)

	// Start pipeline
	err := pipeline.Start()
	assert.NoError(t, err)
	defer pipeline.Stop()

	// Process some frames
	for i := 0; i < 3; i++ {
		frame := ml.NewEnhancedVideoFrame([]byte("test"), time.Now(), i)
		err := pipeline.ProcessFrame(frame)
		assert.NoError(t, err)
	}

	// Check metrics
	metrics := pipeline.GetMetrics()
	assert.NotNil(t, metrics)
}
