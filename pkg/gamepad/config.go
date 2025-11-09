package gamepad

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ConfigLoader handles loading and validating gamepad configurations
type ConfigLoader struct {
	schema *jsonschema.Schema
}

// NewConfigLoader creates a new config loader with schema validation
func NewConfigLoader() (*ConfigLoader, error) {
	// Get the schema file path
	schemaPath, err := getSchemaPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema path: %w", err)
	}

	// Compile the schema
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &ConfigLoader{
		schema: schema,
	}, nil
}

// LoadConfig loads and validates a gamepad configuration from file
func (cl *ConfigLoader) LoadConfig(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read and parse the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration against schema
	if err := cl.validateJSONData(data); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Apply defaults for optional fields
	cl.applyDefaults(&config)

	utils.Logger.Infof("Successfully loaded gamepad configuration from: %s", configPath)
	return &config, nil
}

// LoadDefaultConfig loads the default configuration
func (cl *ConfigLoader) LoadDefaultConfig() (*Config, error) {
	defaultPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get default config path: %w", err)
	}

	return cl.LoadConfig(defaultPath)
}

// validateJSONData validates JSON data against the schema
func (cl *ConfigLoader) validateJSONData(data []byte) error {
	// Create a temporary file for validation
	tempFile := os.TempDir() + "/gamepad-config-validate.json"
	if err := os.WriteFile(tempFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp validation file: %w", err)
	}
	defer os.Remove(tempFile)

	// Open the temp file for validation
	f, err := os.Open(tempFile)
	if err != nil {
		return fmt.Errorf("failed to open temp validation file: %w", err)
	}
	defer f.Close()

	// Unmarshal JSON data
	instance, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON for validation: %w", err)
	}

	// Validate against schema
	if err := cl.schema.Validate(instance); err != nil {
		return err
	}

	return nil
}

// SaveConfig saves a configuration to file with validation
func (cl *ConfigLoader) SaveConfig(config *Config, configPath string) error {
	// Validate the configuration before saving
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Validate the configuration data
	if err := cl.validateJSONData(configData); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the configuration file
	if err := os.WriteFile(configPath, configData, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	utils.Logger.Infof("Successfully saved gamepad configuration to: %s", configPath)
	return nil
}

// ValidateConfig validates a configuration object against the schema
func (cl *ConfigLoader) ValidateConfig(config *Config) error {
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config for validation: %w", err)
	}

	return cl.validateJSONData(configData)
}

// applyDefaults applies default values for optional fields
func (cl *ConfigLoader) applyDefaults(config *Config) {
	// Set default version if not specified
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	// Apply defaults for optional button mappings
	if config.Mappings.Buttons.FlipLeft == nil {
		config.Mappings.Buttons.FlipLeft = &ButtonMapping{
			Button:    LeftBumper,
			HoldTime:  0,
			DoubleTap: false,
		}
	}

	if config.Mappings.Buttons.FlipRight == nil {
		config.Mappings.Buttons.FlipRight = &ButtonMapping{
			Button:    RightBumper,
			HoldTime:  0,
			DoubleTap: false,
		}
	}

	if config.Mappings.Buttons.StreamToggle == nil {
		config.Mappings.Buttons.StreamToggle = &ButtonMapping{
			Button:    ButtonSelect,
			HoldTime:  0,
			DoubleTap: false,
		}
	}

	// Apply defaults for optional axis properties
	if config.Mappings.Axes.MovementX.Deadzone == nil {
		config.Mappings.Axes.MovementX.Deadzone = &config.Controller.Deadzone
	}

	if config.Mappings.Axes.MovementY.Deadzone == nil {
		config.Mappings.Axes.MovementY.Deadzone = &config.Controller.Deadzone
	}

	if config.Mappings.Axes.Altitude.Deadzone == nil {
		config.Mappings.Axes.Altitude.Deadzone = &config.Controller.Deadzone
	}

	if config.Mappings.Axes.Yaw.Deadzone == nil {
		config.Mappings.Axes.Yaw.Deadzone = &config.Controller.Deadzone
	}

	if config.Mappings.Axes.MovementX.Sensitivity == nil {
		config.Mappings.Axes.MovementX.Sensitivity = &config.Controller.Sensitivity
	}

	if config.Mappings.Axes.MovementY.Sensitivity == nil {
		config.Mappings.Axes.MovementY.Sensitivity = &config.Controller.Sensitivity
	}

	if config.Mappings.Axes.Altitude.Sensitivity == nil {
		config.Mappings.Axes.Altitude.Sensitivity = &config.Controller.Sensitivity
	}

	if config.Mappings.Axes.Yaw.Sensitivity == nil {
		config.Mappings.Axes.Yaw.Sensitivity = &config.Controller.Sensitivity
	}
}

// getSchemaPath returns the path to the JSON schema file
func getSchemaPath() (string, error) {
	// Try to find the schema file relative to the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	schemaPath := filepath.Join(wd, "configs", "gamepad-schema.json")
	if _, err := os.Stat(schemaPath); err == nil {
		return schemaPath, nil
	}

	// Try relative to the package directory
	pkgPath := filepath.Join(wd, "pkg", "gamepad", "..", "..", "configs", "gamepad-schema.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return pkgPath, nil
	}

	return "", fmt.Errorf("gamepad schema file not found")
}

// getDefaultConfigPath returns the path to the default configuration file
func getDefaultConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(wd, "configs", "gamepad-default.json")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Try relative to the package directory
	pkgPath := filepath.Join(wd, "pkg", "gamepad", "..", "..", "configs", "gamepad-default.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return pkgPath, nil
	}

	return "", fmt.Errorf("default gamepad configuration file not found")
}

// LoadConfigFromFile loads a configuration from a specific file path
// This is a convenience function that creates a ConfigLoader and loads the config
func LoadConfigFromFile(configPath string) (*Config, error) {
	loader, err := NewConfigLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	return loader.LoadConfig(configPath)
}

// LoadDefaultConfigFromFile loads the default configuration
// This is a convenience function that creates a ConfigLoader and loads the default config
func LoadDefaultConfigFromFile() (*Config, error) {
	loader, err := NewConfigLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	return loader.LoadDefaultConfig()
}

// ValidateConfigFile validates a configuration file against the schema
func ValidateConfigFile(configPath string) error {
	loader, err := NewConfigLoader()
	if err != nil {
		return fmt.Errorf("failed to create config loader: %w", err)
	}

	_, err = loader.LoadConfig(configPath)
	return err
}
