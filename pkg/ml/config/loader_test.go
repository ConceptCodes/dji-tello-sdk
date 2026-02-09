package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewConfigManager tests that NewConfigManager creates a config manager with the given path
func TestNewConfigManager(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewConfigManager(tempDir)

	if cm == nil {
		t.Fatal("NewConfigManager returned nil")
	}

	if cm.configDir != tempDir {
		t.Errorf("expected configDir to be %s, got %s", tempDir, cm.configDir)
	}

	if cm.schemas == nil {
		t.Error("schemas map should be initialized")
	}

	if cm.schemaLoader == nil {
		t.Error("schemaLoader should be initialized")
	}
}

// TestLoadProcessorConfig_Valid tests loading a valid JSON config file
func TestLoadProcessorConfig_Valid(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewConfigManager(tempDir)

	// Create a valid JSON config file
	configContent := `{
		"model": "test-model.onnx",
		"confidence": 0.5,
		"threshold": 0.4
	}`

	configFile := filepath.Join(tempDir, "test-config.json")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config file: %v", err)
	}

	// Load the config
	config, err := cm.LoadProcessorConfig("test-config.json")
	if err != nil {
		t.Fatalf("LoadProcessorConfig failed: %v", err)
	}

	// Verify the config was loaded correctly
	if config == nil {
		t.Fatal("config should not be nil")
	}

	if config["model"] != "test-model.onnx" {
		t.Errorf("expected model to be 'test-model.onnx', got %v", config["model"])
	}

	if config["confidence"] != 0.5 {
		t.Errorf("expected confidence to be 0.5, got %v", config["confidence"])
	}
}

// TestLoadProcessorConfig_Invalid tests handling of invalid JSON
func TestLoadProcessorConfig_Invalid(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewConfigManager(tempDir)

	// Create an invalid JSON config file
	invalidContent := `{
		"model": "test-model.onnx",
		"confidence": 0.5,
		"threshold": 0.4
	` // Missing closing brace

	configFile := filepath.Join(tempDir, "invalid-config.json")
	if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to create test config file: %v", err)
	}

	// Try to load the invalid config
	_, err := cm.LoadProcessorConfig("invalid-config.json")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestLoadProcessorConfig_MissingFile tests that loading a missing file returns an error
func TestLoadProcessorConfig_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	cm := NewConfigManager(tempDir)

	// Try to load a non-existent config file
	_, err := cm.LoadProcessorConfig("non-existent-config.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
