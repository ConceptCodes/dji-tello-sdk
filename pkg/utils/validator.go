package utils

import "fmt"

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
	radius := (dx*dx + dy*dy + dz*dz) / (2.0 * (dx*dx + dy*dy + dz*dz))

	if radius < min || radius > max {
		return fmt.Errorf("calculated arc radius %.2f is out of the allowed range (%.2f-%.2f units)", radius, min, max)
	}
	return nil
}
