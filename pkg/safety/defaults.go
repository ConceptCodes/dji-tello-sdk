package safety

import (
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/utils"
)

// DefaultConfig returns default safety configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		Level:   SafetyLevelNormal,
		Altitude: AltitudeLimits{
			MinHeight:     20,
			MaxHeight:     300,
			TakeoffHeight: 50,
		},
		Velocity: VelocityLimits{
			MaxHorizontal: 100,
			MaxVertical:   80,
			MaxYaw:        100,
		},
		Battery: BatterySafety{
			WarningThreshold:   30,
			CriticalThreshold:  20,
			EmergencyThreshold: 15,
			EnableAutoLand:     true,
			LowBatteryAction:   "land",
		},
		Sensors: SensorSafety{
			MinTOFDistance:      30,
			MaxTiltAngle:        30,
			MaxAcceleration:     2.0,
			BaroPressureDelta:   5.0,
			SensorFailureAction: "land",
		},
		Emergency: EmergencyProcedures{
			ConnectionTimeout:   3000,
			SensorFailureAction: "land",
			EnableAutoLand:      true,
			LowBatteryAction:    "land",
		},
		Behavioral: BehavioralLimits{
			EnableFlips:    true,
			MinFlipHeight:  100,
			MaxFlightTime:  600,
			MaxCommandRate: 10,
		},
	}
}

// ConservativeConfig returns conservative safety configuration with restrictive limits
func ConservativeConfig() *Config {
	config := DefaultConfig()
	config.Version = "1.0.0-conservative"
	config.Level = SafetyLevelConservative

	// More restrictive limits
	config.Altitude.MaxHeight = 150
	config.Velocity.MaxHorizontal = 60
	config.Velocity.MaxVertical = 40
	config.Velocity.MaxYaw = 60
	config.Battery.WarningThreshold = 40
	config.Battery.CriticalThreshold = 30
	config.Battery.EmergencyThreshold = 25
	config.Sensors.MinTOFDistance = 50
	config.Sensors.MaxTiltAngle = 20
	config.Sensors.MaxAcceleration = 1.5
	config.Behavioral.EnableFlips = false
	config.Behavioral.MaxFlightTime = 300
	config.Behavioral.MaxCommandRate = 5

	utils.Logger.Info("Created conservative safety configuration")
	return config
}

// AggressiveConfig returns aggressive safety configuration with permissive limits
func AggressiveConfig() *Config {
	config := DefaultConfig()
	config.Version = "1.0.0-aggressive"
	config.Level = SafetyLevelAggressive

	// More permissive limits
	config.Altitude.MaxHeight = 500
	config.Velocity.MaxHorizontal = 100
	config.Velocity.MaxVertical = 100
	config.Velocity.MaxYaw = 100
	config.Battery.WarningThreshold = 25
	config.Battery.CriticalThreshold = 15
	config.Battery.EmergencyThreshold = 10
	config.Sensors.MinTOFDistance = 20
	config.Sensors.MaxTiltAngle = 45
	config.Sensors.MaxAcceleration = 3.0
	config.Behavioral.EnableFlips = true
	config.Behavioral.MinFlipHeight = 50
	config.Behavioral.MaxFlightTime = 900
	config.Behavioral.MaxCommandRate = 20

	utils.Logger.Info("Created aggressive safety configuration")
	return config
}

// IndoorConfig returns indoor-specific safety configuration
func IndoorConfig() *Config {
	config := ConservativeConfig()
	config.Version = "1.0.0-indoor"
	config.Level = SafetyLevelIndoor

	// Indoor-specific limits
	config.Altitude.MaxHeight = 200
	config.Sensors.MinTOFDistance = 50
	config.Velocity.MaxHorizontal = 40
	config.Velocity.MaxVertical = 30
	config.Velocity.MaxYaw = 50
	config.Behavioral.EnableFlips = false
	config.Sensors.MaxTiltAngle = 15
	config.Behavioral.MaxFlightTime = 300

	utils.Logger.Info("Created indoor safety configuration")
	return config
}

// OutdoorConfig returns outdoor-specific safety configuration
func OutdoorConfig() *Config {
	config := DefaultConfig()
	config.Version = "1.0.0-outdoor"
	config.Level = SafetyLevelOutdoor

	// Outdoor-specific limits
	config.Altitude.MaxHeight = 400
	config.Velocity.MaxHorizontal = 100
	config.Velocity.MaxVertical = 80
	config.Velocity.MaxYaw = 100
	config.Behavioral.MaxFlightTime = 900
	config.Sensors.MinTOFDistance = 30
	config.Behavioral.EnableFlips = true

	utils.Logger.Info("Created outdoor safety configuration")
	return config
}

// RacingConfig returns racing/FPV-specific safety configuration
func RacingConfig() *Config {
	config := AggressiveConfig()
	config.Version = "1.0.0-racing"
	config.Level = "racing"

	// Racing-specific limits
	config.Altitude.MaxHeight = 300
	config.Velocity.MaxHorizontal = 100
	config.Velocity.MaxVertical = 100
	config.Velocity.MaxYaw = 100
	config.Behavioral.EnableFlips = true
	config.Behavioral.MinFlipHeight = 30
	config.Sensors.MaxTiltAngle = 60
	config.Sensors.MaxAcceleration = 4.0
	config.Behavioral.MaxCommandRate = 30
	config.Battery.WarningThreshold = 20
	config.Battery.CriticalThreshold = 15
	config.Battery.EmergencyThreshold = 10

	utils.Logger.Info("Created racing safety configuration")
	return config
}

// GetPresetConfigs returns all available preset configurations
func GetPresetConfigs() map[string]*Config {
	return map[string]*Config{
		"default":      DefaultConfig(),
		"conservative": ConservativeConfig(),
		"aggressive":   AggressiveConfig(),
		"indoor":       IndoorConfig(),
		"outdoor":      OutdoorConfig(),
		"racing":       RacingConfig(),
	}
}

// GetConfigNames returns names of all available preset configurations
func GetConfigNames() []string {
	return []string{"default", "conservative", "aggressive", "indoor", "outdoor", "racing"}
}
