package tracking

import (
	"image"
	"math"
)

// calculateIoU calculates Intersection over Union between two bounding boxes
func calculateIoU(box1, box2 image.Rectangle) float32 {
	x1 := math.Max(float64(box1.Min.X), float64(box2.Min.X))
	y1 := math.Max(float64(box1.Min.Y), float64(box2.Min.Y))
	x2 := math.Min(float64(box1.Max.X), float64(box2.Max.X))
	y2 := math.Min(float64(box1.Max.Y), float64(box2.Max.Y))

	if x2 <= x1 || y2 <= y1 {
		return 0
	}

	intersection := (x2 - x1) * (y2 - y1)
	area1 := float64(box1.Dx()) * float64(box1.Dy())
	area2 := float64(box2.Dx()) * float64(box2.Dy())
	union := area1 + area2 - intersection

	return float32(intersection / union)
}

// calculateIoUDistance calculates IoU distance (1 - IoU) for cost calculation
func calculateIoUDistance(box1, box2 image.Rectangle) float64 {
	iou := calculateIoU(box1, box2)
	return 1.0 - float64(iou)
}

// calculateCenterDistance calculates Euclidean distance between box centers
func calculateCenterDistance(box1, box2 image.Rectangle) float64 {
	center1 := image.Point{
		X: box1.Min.X + box1.Dx()/2,
		Y: box1.Min.Y + box1.Dy()/2,
	}
	center2 := image.Point{
		X: box2.Min.X + box2.Dx()/2,
		Y: box2.Min.Y + box2.Dy()/2,
	}

	dx := float64(center1.X - center2.X)
	dy := float64(center1.Y - center2.Y)

	return math.Sqrt(dx*dx + dy*dy)
}

// calculateBoundingBoxArea calculates the area of a bounding box
func calculateBoundingBoxArea(box image.Rectangle) float64 {
	return float64(box.Dx() * box.Dy())
}

// mergeBoundingBoxes merges two bounding boxes
func mergeBoundingBoxes(box1, box2 image.Rectangle) image.Rectangle {
	x1 := min(box1.Min.X, box2.Min.X)
	y1 := min(box1.Min.Y, box2.Min.Y)
	x2 := max(box1.Max.X, box2.Max.X)
	y2 := max(box1.Max.Y, box2.Max.Y)

	return image.Rect(x1, y1, x2, y2)
}

// expandBoundingBox expands a bounding box by a given factor
func expandBoundingBox(box image.Rectangle, factor float64) image.Rectangle {
	width := int(float64(box.Dx()) * factor)
	height := int(float64(box.Dy()) * factor)

	centerX := box.Min.X + box.Dx()/2
	centerY := box.Min.Y + box.Dy()/2

	newX1 := centerX - width/2
	newY1 := centerY - height/2
	newX2 := centerX + width/2
	newY2 := centerY + height/2

	return image.Rect(newX1, newY1, newX2, newY2)
}

// clampBoundingBox clamps a bounding box to image boundaries
func clampBoundingBox(box image.Rectangle, imageWidth, imageHeight int) image.Rectangle {
	x1 := max(0, box.Min.X)
	y1 := max(0, box.Min.Y)
	x2 := min(imageWidth, box.Max.X)
	y2 := min(imageHeight, box.Max.Y)

	return image.Rect(x1, y1, x2, y2)
}

// isValidBoundingBox checks if a bounding box is valid
func isValidBoundingBox(box image.Rectangle) bool {
	return box.Min.X < box.Max.X && box.Min.Y < box.Max.Y
}

// boundingBoxFromCenter creates a bounding box from center point and dimensions
func boundingBoxFromCenter(centerX, centerY, width, height int) image.Rectangle {
	x1 := centerX - width/2
	y1 := centerY - height/2
	x2 := centerX + width/2
	y2 := centerY + height/2

	return image.Rect(x1, y1, x2, y2)
}

// boundingBoxToCenter converts bounding box to center coordinates and dimensions
func boundingBoxToCenter(box image.Rectangle) (centerX, centerY, width, height int) {
	centerX = box.Min.X + box.Dx()/2
	centerY = box.Min.Y + box.Dy()/2
	width = box.Dx()
	height = box.Dy()
	return
}
