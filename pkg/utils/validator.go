package utils

import (
	"fmt"
	"math"
)

func ValidateNumberInRange(number, min, max int) error {
	if number < min || number > max {
		return fmt.Errorf("number %d is out of range [%d, %d]", number, min, max)
	}
	return nil
}

func ValidateArcRadius(x1, x2, y1, y2, z1, z2 int, min float64, max float64) error {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	dz := float64(z2 - z1)
	
	// Calculate the Euclidean distance (this represents the arc radius)
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	if distance < min || distance > max {
		return fmt.Errorf("calculated arc radius %.2f is out of the allowed range (%.2f-%.2f units)", distance, min, max)
	}
	return nil
}