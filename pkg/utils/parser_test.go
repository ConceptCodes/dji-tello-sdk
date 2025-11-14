package utils

import (
	"testing"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"-45", -45, false},
		{"999999", 999999, false},
		{"", 0, true},
		{"abc", 0, true},
		{"12.5", 0, true},
		{"123abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseInt(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', got %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %d for input '%s', got %d", test.expected, test.input, result)
				}
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"123.45", 123.45, false},
		{"0", 0.0, false},
		{"-45.67", -45.67, false},
		{"999999.999", 999999.999, false},
		{"", 0.0, true},
		{"abc", 0.0, true},
		{"123", 123.0, false}, // Integer should parse as float
		{"1e10", 1e10, false}, // Scientific notation
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseFloat(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', got %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %f for input '%s', got %f", test.expected, test.input, result)
				}
			}
		})
	}
}

func TestParseState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *types.State
		hasError bool
	}{
		{
			name:  "Complete valid state",
			input: "pitch:10;roll:-5;yaw:180;vgx:20;vgy:30;vgz:40;templ:20;temph:30;tof:300;h:100;bat:85;baro:1013.25;time:120;agx:0.1;agy:0.2;agz:0.3;",
			expected: &types.State{
				Pitch: 10, Roll: -5, Yaw: 180,
				Vgx: 20, Vgy: 30, Vgz: 40,
				Templ: 20, Temph: 30,
				Tof: 300, H: 100, Bat: 85,
				Baro: 1013.25, Time: 120,
				Agx: 0.1, Agy: 0.2, Agz: 0.3,
			},
			hasError: false,
		},
		{
			name:  "Partial state",
			input: "pitch:15;bat:90;time:200;",
			expected: &types.State{
				Pitch: 15, Bat: 90, Time: 200,
			},
			hasError: false,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: &types.State{},
			hasError: false,
		},
		{
			name:     "Invalid format - missing colon",
			input:    "pitch10;roll:20;",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid format - empty key",
			input:    ":10;roll:20;",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid format - empty value",
			input:    "pitch:;roll:20;",
			expected: nil,
			hasError: true,
		},
		{
			name:  "Invalid numeric values",
			input: "pitch:abc;roll:20;bat:xyz;",
			expected: &types.State{
				Roll: 20, // Valid values should still be parsed
			},
			hasError: false, // Should not error, just skip invalid values
		},
		{
			name:  "Unknown keys should be ignored",
			input: "pitch:10;unknown:123;roll:20;",
			expected: &types.State{
				Pitch: 10, Roll: 20,
			},
			hasError: false,
		},
		{
			name:  "State with extra whitespace",
			input: " pitch : 10 ; roll : 20 ; ",
			expected: &types.State{
				Pitch: 10, Roll: 20,
			},
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseState(test.input)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input '%s', got %v", test.input, err)
				}

				if result == nil && test.expected != nil {
					t.Errorf("Expected state %+v, got nil", test.expected)
					return
				}

				if result != nil && test.expected == nil {
					t.Errorf("Expected nil, got state %+v", result)
					return
				}

				if result != nil && test.expected != nil {
					// Compare individual fields
					if result.Pitch != test.expected.Pitch {
						t.Errorf("Expected Pitch %d, got %d", test.expected.Pitch, result.Pitch)
					}
					if result.Roll != test.expected.Roll {
						t.Errorf("Expected Roll %d, got %d", test.expected.Roll, result.Roll)
					}
					if result.Yaw != test.expected.Yaw {
						t.Errorf("Expected Yaw %d, got %d", test.expected.Yaw, result.Yaw)
					}
					if result.Vgx != test.expected.Vgx {
						t.Errorf("Expected Vgx %d, got %d", test.expected.Vgx, result.Vgx)
					}
					if result.Vgy != test.expected.Vgy {
						t.Errorf("Expected Vgy %d, got %d", test.expected.Vgy, result.Vgy)
					}
					if result.Vgz != test.expected.Vgz {
						t.Errorf("Expected Vgz %d, got %d", test.expected.Vgz, result.Vgz)
					}
					if result.Templ != test.expected.Templ {
						t.Errorf("Expected Templ %d, got %d", test.expected.Templ, result.Templ)
					}
					if result.Temph != test.expected.Temph {
						t.Errorf("Expected Temph %d, got %d", test.expected.Temph, result.Temph)
					}
					if result.Tof != test.expected.Tof {
						t.Errorf("Expected Tof %d, got %d", test.expected.Tof, result.Tof)
					}
					if result.H != test.expected.H {
						t.Errorf("Expected H %d, got %d", test.expected.H, result.H)
					}
					if result.Bat != test.expected.Bat {
						t.Errorf("Expected Bat %d, got %d", test.expected.Bat, result.Bat)
					}
					if result.Baro != test.expected.Baro {
						t.Errorf("Expected Baro %f, got %f", test.expected.Baro, result.Baro)
					}
					if result.Time != test.expected.Time {
						t.Errorf("Expected Time %d, got %d", test.expected.Time, result.Time)
					}
					if result.Agx != test.expected.Agx {
						t.Errorf("Expected Agx %f, got %f", test.expected.Agx, result.Agx)
					}
					if result.Agy != test.expected.Agy {
						t.Errorf("Expected Agy %f, got %f", test.expected.Agy, result.Agy)
					}
					if result.Agz != test.expected.Agz {
						t.Errorf("Expected Agz %f, got %f", test.expected.Agz, result.Agz)
					}
				}
			}
		})
	}
}
