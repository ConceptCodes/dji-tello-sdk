package safety

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Check some default values
	if config.Altitude.MaxHeight != 300 {
		t.Errorf("Default Altitude.MaxHeight = %d, want 300", config.Altitude.MaxHeight)
	}

	if config.Battery.CriticalThreshold != 20 {
		t.Errorf("Default Battery.CriticalThreshold = %d, want 20", config.Battery.CriticalThreshold)
	}

	if config.Velocity.MaxHorizontal != 100 {
		t.Errorf("Default Velocity.MaxHorizontal = %d, want 100", config.Velocity.MaxHorizontal)
	}

	if config.Velocity.MaxVertical != 80 {
		t.Errorf("Default Velocity.MaxVertical = %d, want 80", config.Velocity.MaxVertical)
	}

	if config.Level != SafetyLevelNormal {
		t.Errorf("Default Level = %v, want %v", config.Level, SafetyLevelNormal)
	}
}

func TestPresetConfigs(t *testing.T) {
	tests := []struct {
		name   string
		preset string
	}{
		{"conservative", "conservative"},
		{"aggressive", "aggressive"},
		{"indoor", "indoor"},
		{"outdoor", "outdoor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := GetPresetConfigs()
			config, exists := configs[tt.preset]
			if !exists {
				t.Errorf("Preset %s not found in GetPresetConfigs()", tt.preset)
				return
			}

			if config == nil {
				t.Errorf("GetPresetConfigs()[%s] returned nil", tt.preset)
			}

			// Check that preset has correct level
			if config.Level != SafetyLevel(tt.preset) {
				t.Errorf("Preset %s Level = %v, want %v", tt.preset, config.Level, tt.preset)
			}
		})
	}
}

func TestGetConfigNames(t *testing.T) {
	names := GetConfigNames()

	if len(names) == 0 {
		t.Error("GetConfigNames() returned empty slice")
	}

	// Check that expected names are present
	expectedNames := []string{"default", "conservative", "aggressive", "indoor", "outdoor", "racing"}
	for _, expected := range expectedNames {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected config name %s not found in GetConfigNames()", expected)
		}
	}
}

func TestConfigLoaderFunctions(t *testing.T) {
	// Test the preset config functions that don't require file loading
	t.Run("ConservativeConfig", func(t *testing.T) {
		config := ConservativeConfig()
		if config == nil {
			t.Error("ConservativeConfig() returned nil")
		}
		if config.Level != SafetyLevelConservative {
			t.Errorf("ConservativeConfig() Level = %v, want %v", config.Level, SafetyLevelConservative)
		}
	})

	t.Run("AggressiveConfig", func(t *testing.T) {
		config := AggressiveConfig()
		if config == nil {
			t.Error("AggressiveConfig() returned nil")
		}
		if config.Level != SafetyLevelAggressive {
			t.Errorf("AggressiveConfig() Level = %v, want %v", config.Level, SafetyLevelAggressive)
		}
	})

	t.Run("IndoorConfig", func(t *testing.T) {
		config := IndoorConfig()
		if config == nil {
			t.Error("IndoorConfig() returned nil")
		}
		if config.Level != SafetyLevelIndoor {
			t.Errorf("IndoorConfig() Level = %v, want %v", config.Level, SafetyLevelIndoor)
		}
	})

	t.Run("OutdoorConfig", func(t *testing.T) {
		config := OutdoorConfig()
		if config == nil {
			t.Error("OutdoorConfig() returned nil")
		}
		if config.Level != SafetyLevelOutdoor {
			t.Errorf("OutdoorConfig() Level = %v, want %v", config.Level, SafetyLevelOutdoor)
		}
	})
}
