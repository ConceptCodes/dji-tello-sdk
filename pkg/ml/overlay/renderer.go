package overlay

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// Renderer renders ML results onto images
type Renderer struct {
	config *ml.OverlayConfig
	fonts  map[string]Font
	colors map[string]color.RGBA
}

// Font represents a font for text rendering
type Font interface {
	MeasureText(text string) (width, height int)
	DrawText(img draw.Image, text string, x, y int, col color.Color)
}

// NewRenderer creates a new overlay renderer
func NewRenderer(config *ml.OverlayConfig) *Renderer {
	renderer := &Renderer{
		config: config,
		fonts:  make(map[string]Font),
		colors: make(map[string]color.RGBA),
	}

	// Parse colors
	for name, hex := range config.Colors {
		if col, err := parseHexColor(hex); err == nil {
			renderer.colors[name] = col
		}
	}

	// Set default colors
	if _, exists := renderer.colors["default"]; !exists {
		renderer.colors["default"] = color.RGBA{0, 255, 0, 255} // Green
	}

	return renderer
}

// Render renders ML results onto an image
func (r *Renderer) Render(img draw.Image, results map[string]ml.MLResult, metrics ml.PipelineMetrics) draw.Image {
	// Create a copy of the image to draw on
	resultImg := image.NewRGBA(img.Bounds())
	draw.Draw(resultImg, img.Bounds(), img, image.Point{}, draw.Src)

	// Render FPS if enabled
	if r.config.ShowFPS {
		r.renderFPS(resultImg, metrics.FPS)
	}

	// Render ML results
	for _, result := range results {
		switch res := result.(type) {
		case ml.DetectionResult:
			r.renderDetections(resultImg, res)
		case ml.SLAMResult:
			r.renderSLAMResult(resultImg, res)
		case ml.GestureResult:
			r.renderGestureResult(resultImg, res)
		case ml.DepthResult:
			r.renderDepthResult(resultImg, res)
		}
	}

	return resultImg
}

// renderDetections renders detection results
func (r *Renderer) renderDetections(img draw.Image, result ml.DetectionResult) {
	if !r.config.ShowDetections {
		return
	}

	for _, detection := range result.Detections {
		// Get color for this class
		col := r.getColorForClass(detection.ClassName)

		// Draw bounding box
		r.drawBoundingBox(img, detection.Box, col)

		// Draw confidence if enabled
		if r.config.ShowConfidence {
			r.drawConfidence(img, detection.Box, detection.Confidence, col)
		}

		// Draw class name
		r.drawClassName(img, detection.Box, detection.ClassName, col)
	}
}

// renderSLAMResult renders SLAM results
func (r *Renderer) renderSLAMResult(img draw.Image, result ml.SLAMResult) {
	// Render pose information
	if result.Pose != nil {
		r.renderPose(img, *result.Pose)
	}

	// Render features
	for _, feature := range result.Features {
		r.drawFeature(img, feature)
	}
}

// renderGestureResult renders gesture results
func (r *Renderer) renderGestureResult(img draw.Image, result ml.GestureResult) {
	col := r.colors["default"]

	// Draw bounding box
	r.drawBoundingBox(img, result.BoundingBox, col)

	// Draw gesture name
	r.drawText(img, result.Gesture, result.BoundingBox.Min.X, result.BoundingBox.Min.Y-20, col)

	// Draw confidence if enabled
	if r.config.ShowConfidence {
		confText := strconv.FormatFloat(float64(result.Confidence), 'f', 2, 32)
		r.drawText(img, confText, result.BoundingBox.Min.X, result.BoundingBox.Min.Y-5, col)
	}

	// Draw landmarks
	for i, landmark := range result.Landmarks {
		col := color.RGBA{255, 0, 0, 255} // Red for landmarks
		r.drawPoint(img, landmark, 3, col)

		// Draw landmark index
		r.drawText(img, strconv.Itoa(i), landmark.X+5, landmark.Y-5, col)
	}
}

// renderDepthResult renders depth results
func (r *Renderer) renderDepthResult(img draw.Image, result ml.DepthResult) {
	// Create depth visualization
	if len(result.DepthMap) > 0 {
		r.renderDepthMap(img, result)
	}
}

// renderFPS renders FPS counter
func (r *Renderer) renderFPS(img draw.Image, fps float64) {
	fpsText := "FPS: " + strconv.FormatFloat(fps, 'f', 1, 64)
	col := color.RGBA{255, 255, 255, 255} // White
	r.drawText(img, fpsText, 10, 30, col)
}

// renderPose renders pose information
func (r *Renderer) renderPose(img draw.Image, pose ml.Pose6D) {
	// Create pose text
	poseText := "Position: (" +
		strconv.FormatFloat(float64(pose.Position.X), 'f', 2, 64) + ", " +
		strconv.FormatFloat(float64(pose.Position.Y), 'f', 2, 64) + ", " +
		strconv.FormatFloat(float64(pose.Position.Z), 'f', 2, 64) + ")"

	col := color.RGBA{255, 255, 0, 255} // Yellow
	r.drawText(img, poseText, 10, 60, col)
}

// drawBoundingBox draws a bounding box
func (r *Renderer) drawBoundingBox(img draw.Image, box image.Rectangle, col color.RGBA) {
	// Draw top edge
	r.drawLine(img, box.Min.X, box.Min.Y, box.Max.X, box.Min.Y, col)
	// Draw right edge
	r.drawLine(img, box.Max.X, box.Min.Y, box.Max.X, box.Max.Y, col)
	// Draw bottom edge
	r.drawLine(img, box.Max.X, box.Max.Y, box.Min.X, box.Max.Y, col)
	// Draw left edge
	r.drawLine(img, box.Min.X, box.Max.Y, box.Min.X, box.Min.Y, col)
}

// drawConfidence draws confidence score
func (r *Renderer) drawConfidence(img draw.Image, box image.Rectangle, confidence float32, col color.RGBA) {
	confText := strconv.FormatFloat(float64(confidence), 'f', 2, 64)
	textX := box.Min.X
	textY := box.Max.Y + 15
	r.drawText(img, confText, textX, textY, col)
}

// drawClassName draws class name
func (r *Renderer) drawClassName(img draw.Image, box image.Rectangle, className string, col color.RGBA) {
	textX := box.Min.X
	textY := box.Min.Y - 5
	r.drawText(img, className, textX, textY, col)
}

// drawFeature draws a feature point
func (r *Renderer) drawFeature(img draw.Image, feature ml.Feature) {
	col := color.RGBA{0, 255, 255, 255} // Cyan
	r.drawPoint(img, feature.Position, 2, col)
}

// drawPoint draws a point
func (r *Renderer) drawPoint(img draw.Image, point image.Point, radius int, col color.RGBA) {
	for y := point.Y - radius; y <= point.Y+radius; y++ {
		for x := point.X - radius; x <= point.X+radius; x++ {
			if math.Sqrt(float64((x-point.X)*(x-point.X)+(y-point.Y)*(y-point.Y))) <= float64(radius) {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, col)
				}
			}
		}
	}
}

// drawLine draws a line using Bresenham's algorithm
func (r *Renderer) drawLine(img draw.Image, x0, y0, x1, y1 int, col color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, col)
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// drawText draws text (simplified implementation)
func (r *Renderer) drawText(img draw.Image, text string, x, y int, col color.RGBA) {
	// This is a simplified text drawing implementation
	// In a real implementation, you would use a proper font library
	// For now, we'll draw a simple background rectangle and skip the actual text rendering

	// Draw background rectangle for text
	textWidth := len(text) * 8 // Approximate width
	textHeight := 16           // Approximate height

	for py := y; py < y+textHeight && py < img.Bounds().Dy(); py++ {
		for px := x; px < x+textWidth && px < img.Bounds().Dx(); px++ {
			if py >= 0 && px >= 0 {
				// Semi-transparent background
				bg := color.RGBA{0, 0, 0, 128}
				img.Set(px, py, bg)
			}
		}
	}

	// TODO: Implement actual text rendering with a font library
}

// renderDepthMap renders depth visualization
func (r *Renderer) renderDepthMap(img draw.Image, result ml.DepthResult) {
	if len(result.DepthMap) == 0 {
		return
	}

	// Create depth visualization by coloring pixels based on depth
	width := result.Width
	height := result.Height

	for y := 0; y < height && y < img.Bounds().Dy(); y++ {
		for x := 0; x < width && x < img.Bounds().Dx(); x++ {
			idx := y*width + x
			if idx < len(result.DepthMap) {
				depth := result.DepthMap[idx]

				// Map depth to color (blue = far, red = near)
				col := r.depthToColor(depth)

				// Blend with original image
				if x < img.Bounds().Dx() && y < img.Bounds().Dy() {
					original := img.At(x, y)
					blended := r.blendColors(original, col, 0.5)
					img.Set(x, y, blended)
				}
			}
		}
	}
}

// depthToColor converts depth value to color
func (r *Renderer) depthToColor(depth float32) color.RGBA {
	// Normalize depth to 0-1 range (assuming max depth of 10m)
	normalized := float64(depth) / 10.0
	if normalized > 1.0 {
		normalized = 1.0
	}

	// Create color gradient from blue (far) to red (near)
	red := uint8(255 * (1.0 - normalized))
	blue := uint8(255 * normalized)

	return color.RGBA{red, 0, blue, 128} // Semi-transparent
}

// blendColors blends two colors
func (r *Renderer) blendColors(c1, c2 color.Color, alpha float64) color.RGBA {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	// Simple alpha blending
	r3 := uint8(float64(r1)*(1-alpha) + float64(r2)*alpha)
	g3 := uint8(float64(g1)*(1-alpha) + float64(g2)*alpha)
	b3 := uint8(float64(b1)*(1-alpha) + float64(b2)*alpha)
	a3 := uint8(float64(a1)*(1-alpha) + float64(a2)*alpha)

	return color.RGBA{r3, g3, b3, a3}
}

// getColorForClass returns color for a given class
func (r *Renderer) getColorForClass(className string) color.RGBA {
	if col, exists := r.colors[className]; exists {
		return col
	}
	return r.colors["default"]
}

// parseHexColor parses hex color string
func parseHexColor(hex string) (color.RGBA, error) {
	var r, g, b, a uint8
	a = 255 // Default to fully opaque

	if len(hex) == 7 {
		// Format: #RRGGBB
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		return color.RGBA{r, g, b, a}, err
	} else if len(hex) == 9 {
		// Format: #RRGGBBAA
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x%02x", &r, &g, &b, &a)
		return color.RGBA{r, g, b, a}, err
	}

	return color.RGBA{}, fmt.Errorf("invalid hex color format: %s", hex)
}

// abs returns absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
