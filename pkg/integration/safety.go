package integration

import (
	"fmt"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
)

// CreateSafetyManager creates a safety manager wrapper around the given commander
// Note: This function returns interface{} to avoid circular imports
// The caller should type assert to the appropriate interface
func CreateSafetyManager(commander interface{}, config *safety.Config) (interface{}, error) {
	if config == nil {
		// Load default config
		defaultConfig, err := safety.LoadDefaultConfig()
		if err != nil {
			return commander, fmt.Errorf("failed to load default safety config: %w", err)
		}
		config = defaultConfig
	}

	// Create safety manager - this will work because safety.SafetyManager implements the same interface
	safetyManager := safety.NewSafetyManager(commander, config)
	return safetyManager, nil
}

// CreateSafetyManagerWithPreset creates a safety manager with a preset configuration
func CreateSafetyManagerWithPreset(commander interface{}, preset string) (interface{}, error) {
	config, err := safety.LoadPresetConfig(preset)
	if err != nil {
		return commander, fmt.Errorf("failed to load safety preset '%s': %w", preset, err)
	}

	return CreateSafetyManager(commander, config)
}

// CreateSafetyManagerFromConfig creates a safety manager from a config file
func CreateSafetyManagerFromConfig(commander interface{}, configPath string) (interface{}, error) {
	config, err := safety.LoadConfigFromFile(configPath)
	if err != nil {
		return commander, fmt.Errorf("failed to load safety config from '%s': %w", configPath, err)
	}

	return CreateSafetyManager(commander, config)
}
