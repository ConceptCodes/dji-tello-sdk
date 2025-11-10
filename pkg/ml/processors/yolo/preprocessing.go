package yolo

import (
	"fmt"
	"image"
	"math"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/yalue/onnxruntime_go"
)

// preprocessFrame converts image frame to ONNX tensor data
func (yp *YOLOProcessor) preprocessFrame(img image.Image) error {
	if yp.inputTensor == nil {
		return fmt.Errorf("input tensor not initialized")
	}

	// Convert image to RGB and resize
	resized := resizeImage(img, yp.inputSize.X, yp.inputSize.Y)

	// Normalize and convert to CHW format
	data := yp.inputTensor.GetData()
	for y := 0; y < yp.inputSize.Y; y++ {
		for x := 0; x < yp.inputSize.X; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()

			// Normalize to [0, 1] and convert to float32
			rf := float32(r>>8) / 255.0
			gf := float32(g>>8) / 255.0
			bf := float32(b>>8) / 255.0

			// CHW format: Channel, Height, Width
			idx := y*yp.inputSize.X + x
			data[idx] = rf                                 // Red channel
			data[yp.inputSize.X*yp.inputSize.Y+idx] = gf   // Green channel
			data[2*yp.inputSize.X*yp.inputSize.Y+idx] = bf // Blue channel
		}
	}

	return nil
}

// postprocessResults converts ONNX output to detections
func (yp *YOLOProcessor) postprocessResults(output *onnxruntime_go.Tensor[float32], originalBounds image.Rectangle) ([]ml.Detection, error) {
	data := output.GetData()
	shape := output.GetShape()

	// YOLOv8 output shape: [1, 84, 8400] where 84 = 4(bbox) + 80(classes)
	if len(shape) != 3 || shape[0] != 1 {
		return nil, nil
	}

	numClasses := shape[1] - 4
	numPredictions := shape[2]

	var detections []ml.Detection

	for i := int64(0); i < numPredictions; i++ {
		// Extract bbox coordinates (center_x, center_y, width, height)
		cx := data[i]
		cy := data[numPredictions+i]
		w := data[2*numPredictions+i]
		h := data[3*numPredictions+i]

		// Find best class
		maxConf := float32(0.0)
		bestClass := 0
		for c := int64(0); c < numClasses; c++ {
			conf := data[(4+c)*numPredictions+i]
			if conf > maxConf {
				maxConf = conf
				bestClass = int(c)
			}
		}

		// Filter by confidence
		if maxConf < yp.config.Confidence {
			continue
		}

		// Convert to corner coordinates and scale to original image size
		x1 := (cx - w/2) * float32(originalBounds.Dx()) / float32(yp.inputSize.X)
		y1 := (cy - h/2) * float32(originalBounds.Dy()) / float32(yp.inputSize.Y)
		x2 := (cx + w/2) * float32(originalBounds.Dx()) / float32(yp.inputSize.X)
		y2 := (cy + h/2) * float32(originalBounds.Dy()) / float32(yp.inputSize.Y)

		// Clamp to image bounds
		x1 = float32(math.Max(0, float64(x1)))
		y1 = float32(math.Max(0, float64(y1)))
		x2 = float32(math.Min(float64(originalBounds.Dx()), float64(x2)))
		y2 = float32(math.Min(float64(originalBounds.Dy()), float64(y2)))

		detection := ml.Detection{
			ClassID:    bestClass,
			ClassName:  yp.getClassName(bestClass),
			Confidence: maxConf,
			Box: image.Rectangle{
				Min: image.Point{X: int(x1), Y: int(y1)},
				Max: image.Point{X: int(x2), Y: int(y2)},
			},
		}

		detections = append(detections, detection)
	}

	// Apply Non-Maximum Suppression
	detections = yp.applyNMS(detections)

	return detections, nil
}

// applyNMS applies Non-Maximum Suppression to detections
func (yp *YOLOProcessor) applyNMS(detections []ml.Detection) []ml.Detection {
	if len(detections) == 0 {
		return detections
	}

	// Sort by confidence
	sorted := make([]ml.Detection, len(detections))
	copy(sorted, detections)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Confidence < sorted[j].Confidence {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var result []ml.Detection
	for i := 0; i < len(sorted); i++ {
		if sorted[i].Confidence == 0 {
			continue
		}

		result = append(result, sorted[i])

		// Suppress overlapping detections
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Confidence == 0 {
				continue
			}

			iou := calculateIoU(sorted[i].Box, sorted[j].Box)
			if iou > yp.config.NMSThreshold {
				sorted[j].Confidence = 0
			}
		}
	}

	return result
}

// getClassName returns class name for given class ID
func (yp *YOLOProcessor) getClassName(classID int) string {
	if classID >= 0 && classID < len(yp.classes) {
		return yp.classes[classID]
	}
	return "unknown"
}

// calculateIoU calculates Intersection over Union
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

// resizeImage resizes an image to the specified dimensions
func resizeImage(img image.Image, width, height int) image.Image {
	// Simple nearest neighbor resize
	srcBounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcBounds.Dx() / width
			srcY := y * srcBounds.Dy() / height
			dst.Set(x, y, img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}

	return dst
}
