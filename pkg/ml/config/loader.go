package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/xeipuuv/gojsonschema"
)

// ConfigManager manages ML configuration with schema validation
type ConfigManager struct {
	schemaLoader *gojsonschema.SchemaLoader
	schemas      map[string]*gojsonschema.Schema
	configDir    string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configDir string) *ConfigManager {
	return &ConfigManager{
		schemaLoader: gojsonschema.NewSchemaLoader(),
		schemas:      make(map[string]*gojsonschema.Schema),
		configDir:    configDir,
	}
}

// LoadSchema loads a JSON schema from file
func (cm *ConfigManager) LoadSchema(name, filename string) error {
	schemaPath := filepath.Join(cm.configDir, "schemas", filename)

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", filename, err)
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to parse schema %s: %w", filename, err)
	}

	cm.schemas[name] = schema
	cm.schemaLoader.AddSchemas(gojsonschema.NewBytesLoader(schemaBytes))

	return nil
}

// ValidateConfig validates configuration against a schema
func (cm *ConfigManager) ValidateConfig(configData []byte, schemaName string) (*gojsonschema.Result, error) {
	schema, exists := cm.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema %s not found", schemaName)
	}

	documentLoader := gojsonschema.NewBytesLoader(configData)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return result, nil
}

// ValidateConfigFromFile validates configuration file against a schema
func (cm *ConfigManager) ValidateConfigFromFile(filename, schemaName string) (*gojsonschema.Result, error) {
	configPath := filepath.Join(cm.configDir, filename)

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	return cm.ValidateConfig(configBytes, schemaName)
}

// LoadMLConfig loads and validates ML configuration
func (cm *ConfigManager) LoadMLConfig(filename string) (*ml.MLConfig, error) {
	configPath := filepath.Join(cm.configDir, filename)

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	// Validate against ML pipeline schema
	result, err := cm.ValidateConfig(configBytes, "ml-pipeline")
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if !result.Valid() {
		return nil, fmt.Errorf("config is invalid: %v", result.Errors())
	}

	// Parse configuration
	var config ml.MLConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate individual processor configurations
	for _, procConfig := range config.Processors {
		if err := cm.validateProcessorConfig(procConfig); err != nil {
			return nil, fmt.Errorf("processor %s validation failed: %w", procConfig.Name, err)
		}
	}

	return &config, nil
}

// SaveMLConfig saves ML configuration to file
func (cm *ConfigManager) SaveMLConfig(config *ml.MLConfig, filename string) error {
	// Validate before saving
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Validate against schema
	result, err := cm.ValidateConfig(configBytes, "ml-pipeline")
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if !result.Valid() {
		return fmt.Errorf("config is invalid: %v", result.Errors())
	}

	// Save to file
	configPath := filepath.Join(cm.configDir, filename)
	if err := os.WriteFile(configPath, configBytes, 0o644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// LoadProcessorConfig loads and validates a specific processor configuration
func (cm *ConfigManager) LoadProcessorConfig(filename string) (map[string]interface{}, error) {
	configPath := filepath.Join(cm.configDir, filename)

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read processor config file %s: %w", filename, err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse processor config: %w", err)
	}

	return config, nil
}

// validateProcessorConfig validates a specific processor configuration
func (cm *ConfigManager) validateProcessorConfig(procConfig ml.ProcessorConfig) error {
	// Basic validation
	if procConfig.Name == "" {
		return fmt.Errorf("processor name cannot be empty")
	}

	if procConfig.Type == "" {
		return fmt.Errorf("processor type cannot be empty")
	}

	// Validate processor type against known types
	validTypes := []ml.ProcessorType{
		ml.ProcessorTypeYOLO,
		ml.ProcessorTypeFace,
		ml.ProcessorTypeSLAM,
		ml.ProcessorTypeGesture,
		ml.ProcessorTypeSegmentation,
		ml.ProcessorTypeCustom,
	}

	valid := false
	for _, validType := range validTypes {
		if procConfig.Type == validType {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid processor type: %s", procConfig.Type)
	}

	// Validate priority
	if procConfig.Priority < 0 {
		return fmt.Errorf("processor priority cannot be negative")
	}

	return nil
}

// GetDefaultMLConfig returns a default ML configuration
func (cm *ConfigManager) GetDefaultMLConfig() *ml.MLConfig {
	return &ml.MLConfig{
		Processors: []ml.ProcessorConfig{},
		Pipeline: ml.PipelineConfig{
			MaxConcurrentProcessors: 4,
			FrameBufferSize:         30,
			WorkerPoolSize:          2,
			EnableMetrics:           true,
			TargetFPS:               30,
		},
		Overlay: ml.OverlayConfig{
			Enabled:        true,
			ShowFPS:        true,
			ShowDetections: true,
			ShowTracking:   true,
			ShowConfidence: true,
			Colors: map[string]string{
				"default": "#00FF00",
				"person":  "#FF0000",
				"car":     "#0000FF",
				"face":    "#FF00FF",
			},
			LineWidth: 2,
			FontSize:  12,
			FontScale: 0.5,
		},
	}
}

// CreateDefaultConfigs creates default configuration files
func (cm *ConfigManager) CreateDefaultConfigs() error {
	// Create directories if they don't exist
	if err := os.MkdirAll(filepath.Join(cm.configDir, "schemas"), 0o755); err != nil {
		return fmt.Errorf("failed to create schemas directory: %w", err)
	}

	// Create default ML config
	defaultConfig := cm.GetDefaultMLConfig()
	if err := cm.SaveMLConfig(defaultConfig, "ml-pipeline-default.json"); err != nil {
		return fmt.Errorf("failed to create default ML config: %w", err)
	}

	// Create default processor configs
	defaultProcessorConfigs := map[string]map[string]interface{}{
		"yolo-default.json": {
			"model":         "yolov8n.onnx",
			"confidence":    0.5,
			"nms_threshold": 0.4,
			"input_size":    []int{640, 640},
			"classes":       []string{"person", "car", "bicycle"},
		},
		"face-default.json": {
			"use_dnn":       true,
			"confidence":    0.7,
			"scale_factor":  1.1,
			"min_neighbors": 3,
			"min_size":      []int{30, 30},
		},
	}

	for filename, config := range defaultProcessorConfigs {
		configBytes, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal default processor config %s: %w", filename, err)
		}

		configPath := filepath.Join(cm.configDir, filename)
		if err := os.WriteFile(configPath, configBytes, 0o644); err != nil {
			return fmt.Errorf("failed to save default processor config %s: %w", filename, err)
		}
	}

	return nil
}

// ListConfigs lists all available configuration files
func (cm *ConfigManager) ListConfigs() ([]string, error) {
	files, err := os.ReadDir(cm.configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var configs []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			configs = append(configs, file.Name())
		}
	}

	return configs, nil
}

// ConfigExists checks if a configuration file exists
func (cm *ConfigManager) ConfigExists(filename string) bool {
	configPath := filepath.Join(cm.configDir, filename)
	_, err := os.Stat(configPath)
	return err == nil
}

// GetConfigPath returns the full path to a configuration file
func (cm *ConfigManager) GetConfigPath(filename string) string {
	return filepath.Join(cm.configDir, filename)
}
