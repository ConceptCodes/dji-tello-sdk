package configs

import _ "embed"

var (
	// GamepadSchema contains the embedded JSON schema used to validate gamepad configs.
	//go:embed gamepad-schema.json
	GamepadSchema []byte

	// SafetyDefault contains the baseline safety configuration used when no file is found.
	//go:embed safety-default.json
	SafetyDefault []byte

	// SafetySchema contains the validation schema for safety configs.
	//go:embed schemas/safety-schema.json
	SafetySchema []byte
)
