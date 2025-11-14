package safety

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/conceptcodes/dji-tello-sdk-go/configs"
	"github.com/conceptcodes/dji-tello-sdk-go/internal/config"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var (
	// ErrConfigNotFound indicates that no safety configuration file was discovered.
	ErrConfigNotFound = errors.New("safety config not found")
)

// getUserConfigDir returns the user config directory path
func getUserConfigDir() string {
	if dir := os.Getenv("TELLO_CONFIG_DIR"); dir != "" {
		return dir
	}

	// Check standard user config directories
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Try ~/.config/tello/ first (following XDG Base Directory spec)
	configDir := filepath.Join(home, ".config", "tello")
	if _, err := os.Stat(configDir); err == nil {
		return configDir
	}

	// Fallback to ~/.tello/
	return filepath.Join(home, ".tello")
}

// getSystemConfigDir returns the system config directory path
func getSystemConfigDir() string {
	// Try /etc/tello/ first
	if _, err := os.Stat("/etc/tello"); err == nil {
		return "/etc/tello"
	}

	// Fallback to /etc/ for backward compatibility
	if _, err := os.Stat("/etc"); err == nil {
		return "/etc"
	}

	return ""
}

// ConfigLoader handles loading and validating safety configurations
type ConfigLoader struct {
	schema *jsonschema.Schema
}

// NewConfigLoader creates a new config loader with schema validation
func NewConfigLoader() (*ConfigLoader, error) {
	fallbacks, err := getSchemaFallbackPaths()
	if err != nil {
		return nil, err
	}

	schema, err := config.CompileSchema(configs.SafetySchema, fallbacks, "embedded://safety-schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile safety schema: %w", err)
	}

	return &ConfigLoader{
		schema: schema,
	}, nil
}

// LoadConfig loads and validates a safety configuration from file
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

	return cl.loadConfigData(data, configPath)
}

func (cl *ConfigLoader) loadConfigData(data []byte, source string) (*Config, error) {
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config data: %w", err)
	}

	if err := cl.validateJSONData(data); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	cl.applyDefaults(&config)

	src := source
	if src == "" {
		src = "embedded safety-default"
	}
	utils.Logger.Infof("Successfully loaded safety configuration from: %s", src)

	return &config, nil
}

// LoadDefaultConfig loads the default configuration
func (cl *ConfigLoader) LoadDefaultConfig() (*Config, error) {
	if len(configs.SafetyDefault) > 0 {
		return cl.loadConfigData(configs.SafetyDefault, "embedded safety-default.json")
	}

	defaultPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get default config path: %w", err)
	}

	return cl.LoadConfig(defaultPath)
}

// LoadPresetConfig loads a preset configuration by name
func (cl *ConfigLoader) LoadPresetConfig(preset string) (*Config, error) {
	presets := GetPresetConfigs()
	config, exists := presets[preset]
	if !exists {
		return nil, fmt.Errorf("unknown preset: %s. Available presets: %v", preset, GetConfigNames())
	}

	utils.Logger.Infof("Using preset safety configuration: %s", preset)
	return config, nil
}

// validateJSONData validates JSON data against the schema
func (cl *ConfigLoader) validateJSONData(data []byte) error {
	return config.ValidateJSON(cl.schema, data)
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

	utils.Logger.Infof("Successfully saved safety configuration to: %s", configPath)
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

	// Set default safety level if not specified
	if config.Level == "" {
		config.Level = SafetyLevelNormal
	}

	// Apply default altitude limits if not set
	if config.Altitude.MinHeight == 0 {
		config.Altitude.MinHeight = 20
	}
	if config.Altitude.MaxHeight == 0 {
		config.Altitude.MaxHeight = 300
	}
	if config.Altitude.TakeoffHeight == 0 {
		config.Altitude.TakeoffHeight = 50
	}

	// Apply default velocity limits if not set
	if config.Velocity.MaxHorizontal == 0 {
		config.Velocity.MaxHorizontal = 100
	}
	if config.Velocity.MaxVertical == 0 {
		config.Velocity.MaxVertical = 80
	}
	if config.Velocity.MaxYaw == 0 {
		config.Velocity.MaxYaw = 100
	}

	// Apply default battery settings if not set
	if config.Battery.WarningThreshold == 0 {
		config.Battery.WarningThreshold = 30
	}
	if config.Battery.CriticalThreshold == 0 {
		config.Battery.CriticalThreshold = 20
	}
	if config.Battery.EmergencyThreshold == 0 {
		config.Battery.EmergencyThreshold = 15
	}
	if config.Battery.LowBatteryAction == "" {
		config.Battery.LowBatteryAction = "land"
	}

	// Apply default sensor settings if not set
	if config.Sensors.MinTOFDistance == 0 {
		config.Sensors.MinTOFDistance = 30
	}
	if config.Sensors.MaxTiltAngle == 0 {
		config.Sensors.MaxTiltAngle = 30
	}
	if config.Sensors.MaxAcceleration == 0 {
		config.Sensors.MaxAcceleration = 2.0
	}
	if config.Sensors.BaroPressureDelta == 0 {
		config.Sensors.BaroPressureDelta = 5.0
	}
	if config.Sensors.SensorFailureAction == "" {
		config.Sensors.SensorFailureAction = "land"
	}

	// Apply default emergency settings if not set
	if config.Emergency.ConnectionTimeout == 0 {
		config.Emergency.ConnectionTimeout = 3000
	}
	if config.Emergency.SensorFailureAction == "" {
		config.Emergency.SensorFailureAction = "land"
	}
	if config.Emergency.LowBatteryAction == "" {
		config.Emergency.LowBatteryAction = "land"
	}

	// Apply default behavioral settings if not set
	if config.Behavioral.MinFlipHeight == 0 {
		config.Behavioral.MinFlipHeight = 100
	}
	if config.Behavioral.MaxFlightTime == 0 {
		config.Behavioral.MaxFlightTime = 600
	}
	if config.Behavioral.MaxCommandRate == 0 {
		config.Behavioral.MaxCommandRate = 10
	}
}

// getSchemaFallbackPaths returns the ordered schema locations to check on disk.
func getSchemaFallbackPaths() ([]string, error) {
	var locations []string

	if userConfigDir := getUserConfigDir(); userConfigDir != "" {
		locations = append(locations,
			filepath.Join(userConfigDir, "safety-schema.json"),
			filepath.Join(userConfigDir, "schemas", "safety-schema.json"),
		)
	}

	if systemConfigDir := getSystemConfigDir(); systemConfigDir != "" {
		locations = append(locations,
			filepath.Join(systemConfigDir, "safety-schema.json"),
			filepath.Join(systemConfigDir, "schemas", "safety-schema.json"),
		)
	}

	if wd, err := os.Getwd(); err == nil {
		locations = append(locations,
			filepath.Join(wd, "configs", "schemas", "safety-schema.json"),
		)
	} else {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("no schema fallback paths available")
	}

	return locations, nil
}

// getDefaultConfigPath returns the path to the default configuration file
func getDefaultConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(wd, "configs", "safety-default.json")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Try relative to the package directory
	pkgPath := filepath.Join(wd, "pkg", "safety", "..", "..", "configs", "safety-default.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return pkgPath, nil
	}

	return "", fmt.Errorf("default safety configuration file not found")
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
func LoadDefaultConfig() (*Config, error) {
	loader, err := NewConfigLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	return loader.LoadDefaultConfig()
}

// LoadPresetConfigFromFile loads a preset configuration by name
// This is a convenience function that creates a ConfigLoader and loads the preset config
func LoadPresetConfig(preset string) (*Config, error) {
	loader, err := NewConfigLoader()
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	return loader.LoadPresetConfig(preset)
}

// LoadPreset loads a preset configuration by name (alias for LoadPresetConfig)
func LoadPreset(preset string) (*Config, error) {
	return LoadPresetConfig(preset)
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

// LoadAutoConfig discovers and loads the first available safety configuration file.
// If no configuration file is found, it falls back to the embedded default config.
// The returned path will be empty when the embedded default is used.
func LoadAutoConfig() (*Config, string, error) {
	loader, err := NewConfigLoader()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create config loader: %w", err)
	}

	configPath, err := FindAutoConfigPath()
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			config, derr := loader.LoadDefaultConfig()
			if derr != nil {
				return nil, "", derr
			}
			return config, "", nil
		}
		return nil, "", err
	}

	config, err := loader.LoadConfig(configPath)
	if err != nil {
		return nil, "", err
	}

	return config, configPath, nil
}

// FindAutoConfigPath returns the discovered config file path following the order:
//  1. config.json (or safety-config.json) in the current directory
//  2. User/global config directories (TELLO_CONFIG_DIR, ~/.config/tello[/safety], ~/.tello[/safety])
//  3. System directories (/etc/tello, /etc)
//
// If no file is found, ErrConfigNotFound is returned.
func FindAutoConfigPath() (string, error) {
	if path, err := findLocalSafetyConfig(); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	if path, err := findGlobalSafetyConfig(); err != nil {
		return "", err
	} else if path != "" {
		return path, nil
	}

	return "", ErrConfigNotFound
}

func findLocalSafetyConfig() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return findSafetyConfigInDir(wd)
}

func findGlobalSafetyConfig() (string, error) {
	dirs := config.UniqueDirs(collectSafetyConfigDirs()...)
	return config.FindConfigFile(safetyConfigFilenames, dirs)
}

func collectSafetyConfigDirs() []string {
	var dirs []string

	if envDir := os.Getenv("TELLO_CONFIG_DIR"); envDir != "" {
		dirs = append(dirs, envDir, filepath.Join(envDir, "safety"))
	}

	if userDir := getUserConfigDir(); userDir != "" {
		dirs = append(dirs,
			userDir,
			filepath.Join(userDir, "safety"),
		)
	}

	if systemDir := getSystemConfigDir(); systemDir != "" {
		dirs = append(dirs,
			systemDir,
			filepath.Join(systemDir, "tello"),
			filepath.Join(systemDir, "tello", "safety"),
		)
	}

	return dirs
}

var safetyConfigFilenames = []string{
	"config.json",
	"safety-config.json",
	"safety.json",
}

func findSafetyConfigInDir(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}

	return config.FindConfigFile(safetyConfigFilenames, []string{dir})
}

// ConfigExists checks if a configuration file exists
func ConfigExists(filename string) bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}

	configPath := filepath.Join(wd, "configs", filename)
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	// Try relative to the package directory
	pkgPath := filepath.Join(wd, "pkg", "safety", "..", "..", "configs", filename)
	_, err = os.Stat(pkgPath)
	return err == nil
}
