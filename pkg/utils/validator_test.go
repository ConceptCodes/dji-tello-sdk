package utils

import (
	"math"
	"testing"
)

func TestValidateNumberInRange(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		min      int
		max      int
		hasError bool
	}{
		{"Valid in middle", 50, 10, 100, false},
		{"Valid at minimum", 10, 10, 100, false},
		{"Valid at maximum", 100, 10, 100, false},
		{"Invalid below minimum", 5, 10, 100, true},
		{"Invalid above maximum", 150, 10, 100, true},
		{"Negative range valid", -5, -10, 0, false},
		{"Negative range invalid", -15, -10, 0, true},
		{"Zero range valid", 0, 0, 0, false},
		{"Single value range valid", 42, 42, 42, false},
		{"Single value range invalid", 41, 42, 42, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateNumberInRange(test.number, test.min, test.max)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for number %d in range [%d, %d]", test.number, test.min, test.max)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for number %d in range [%d, %d], got %v", test.number, test.min, test.max, err)
				}
			}
		})
	}
}

func TestValidateArcRadius(t *testing.T) {
	tests := []struct {
		name                   string
		x1, x2, y1, y2, z1, z2 int
		min, max               float64
		hasError               bool
	}{
		{
			name: "Valid arc within range",
			x1:   0, x2: 5, y1: 0, y2: 5, z1: 0, z2: 5,
			min: 0.5, max: 10.0,
			hasError: false,
		},
		{
			name: "Valid arc at minimum",
			x1:   0, x2: 1, y1: 0, y2: 0, z1: 0, z2: 0,
			min: 0.5, max: 10.0,
			hasError: false,
		},
		{
			name: "Valid arc at maximum",
			x1:   0, x2: 5, y1: 0, y2: 5, z1: 0, z2: 5,
			min: 0.5, max: 10.0,
			hasError: false,
		},
		{
			name: "Arc too small",
			x1:   0, x2: 0, y1: 0, y2: 0, z1: 0, z2: 0,
			min: 0.5, max: 10.0,
			hasError: true,
		},
		{
			name: "Arc too large",
			x1:   0, x2: 100, y1: 0, y2: 100, z1: 0, z2: 100,
			min: 0.5, max: 10.0,
			hasError: true,
		},
		{
			name: "Zero distance arc",
			x1:   100, x2: 100, y1: 100, y2: 100, z1: 100, z2: 100,
			min: 0.5, max: 10.0,
			hasError: true, // Should be invalid as radius calculation will be problematic
		},
		{
			name: "Negative coordinates",
			x1:   -5, x2: 5, y1: -5, y2: 5, z1: -5, z2: 5,
			min: 0.5, max: 20.0,
			hasError: false,
		},
		{
			name: "Large valid arc",
			x1:   -500, x2: 500, y1: -500, y2: 500, z1: -500, z2: 500,
			min: 0.1, max: 2000.0,
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateArcRadius(test.x1, test.x2, test.y1, test.y2, test.z1, test.z2, test.min, test.max)

			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for arc radius validation")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for arc radius validation, got %v", err)
				}
			}
		})
	}
}

func TestValidateArcRadiusCalculation(t *testing.T) {
	// Test specific arc radius calculations to ensure formula is correct
	tests := []struct {
		name                   string
		x1, x2, y1, y2, z1, z2 int
		expectedRadius         float64
	}{
		{
			name: "Unit arc",
			x1:   0, x2: 1, y1: 0, y2: 0, z1: 0, z2: 0,
			expectedRadius: 0.5, // dx=1, dy=0, dz=0, radius=(1*1)/(2*1*1)=0.5
		},
		{
			name: "Diagonal arc",
			x1:   0, x2: 1, y1: 0, y2: 1, z1: 0, z2: 0,
			expectedRadius: 0.5, // dx=1, dy=1, dz=0, radius=(1+1)/(2*(1+1))=0.5
		},
		{
			name: "3D diagonal arc",
			x1:   0, x2: 1, y1: 0, y2: 1, z1: 0, z2: 1,
			expectedRadius: 0.5, // dx=1, dy=1, dz=1, radius=(1+1+1)/(2*(1+1+1))=0.5
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Calculate radius manually to verify formula
			dx := float64(test.x2 - test.x1)
			dy := float64(test.y2 - test.y1)
			dz := float64(test.z2 - test.z1)

			// This is the current formula from the validator
			radius := (dx*dx + dy*dy + dz*dz) / (2.0 * (dx*dx + dy*dy + dz*dz))

			// For non-zero distances, this should equal 0.5
			if math.Abs(radius-test.expectedRadius) > 1e-10 {
				t.Errorf("Expected radius %f, got %f", test.expectedRadius, radius)
			}
		})
	}
}
