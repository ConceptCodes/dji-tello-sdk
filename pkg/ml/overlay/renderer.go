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
		case ml.TrackingResult:
			r.renderTracking(resultImg, res)
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

// renderTracking renders tracking results
func (r *Renderer) renderTracking(img draw.Image, result ml.TrackingResult) {
	if !r.config.ShowTracking {
		return
	}

	for _, track := range result.Tracks {
		// Skip deleted tracks
		if track.State == ml.TrackStateDeleted {
			continue
		}

		// Get color for this track (use different colors for different track IDs)
		col := r.getColorForTrack(track.ID)

		// Draw track bounding box with different styles based on state
		switch track.State {
		case ml.TrackStateTentative:
			r.drawDashedBoundingBox(img, track.Box, col)
		case ml.TrackStateConfirmed:
			r.drawBoundingBox(img, track.Box, col)
		}

		// Draw track ID and class name
		r.drawTrackInfo(img, track, col)

		// Draw prediction if available
		if !track.Prediction.Empty() {
			r.drawPrediction(img, track.Prediction, col)
		}

		// Draw velocity vector if significant
		if track.Velocity.X != 0 || track.Velocity.Y != 0 {
			r.drawVelocityVector(img, track.Box, track.Velocity, col)
		}

		// Draw track trail (simplified - just show age)
		if track.Age > 1 {
			r.drawTrackAge(img, track.Box, track.Age, col)
		}
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

// getColorForTrack returns color for a given track ID
func (r *Renderer) getColorForTrack(trackID int) color.RGBA {
	// Generate consistent color based on track ID
	hue := math.Mod(float64(trackID*137), 360) // Golden angle approximation
	return r.hsvToRGB(hue, 0.8, 1.0)
}

// drawDashedBoundingBox draws a dashed bounding box for tentative tracks
func (r *Renderer) drawDashedBoundingBox(img draw.Image, box image.Rectangle, col color.RGBA) {
	dashLength := 5
	gapLength := 3

	// Draw dashed top edge
	r.drawDashedLine(img, box.Min.X, box.Min.Y, box.Max.X, box.Min.Y, dashLength, gapLength, col)
	// Draw dashed right edge
	r.drawDashedLine(img, box.Max.X, box.Min.Y, box.Max.X, box.Max.Y, dashLength, gapLength, col)
	// Draw dashed bottom edge
	r.drawDashedLine(img, box.Max.X, box.Max.Y, box.Min.X, box.Max.Y, dashLength, gapLength, col)
	// Draw dashed left edge
	r.drawDashedLine(img, box.Min.X, box.Max.Y, box.Min.X, box.Min.Y, dashLength, gapLength, col)
}

// drawDashedLine draws a dashed line
func (r *Renderer) drawDashedLine(img draw.Image, x0, y0, x1, y1, dashLength, gapLength int, col color.RGBA) {
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

	totalLength := dashLength + gapLength
	currentLength := 0
	isDash := true

	for {
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			if isDash {
				img.Set(x0, y0, col)
			}
		}

		currentLength++
		if currentLength >= totalLength {
			currentLength = 0
			isDash = !isDash
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

// drawTrackInfo draws track ID and class information
func (r *Renderer) drawTrackInfo(img draw.Image, track ml.Track, col color.RGBA) {
	text := fmt.Sprintf("ID:%d %s", track.ID, track.ClassName)
	textX := track.Box.Min.X
	textY := track.Box.Min.Y - 5

	// Draw background for better readability
	r.drawTextBackground(img, text, textX, textY, col)
	r.drawText(img, text, textX, textY, col)
}

// drawPrediction draws predicted bounding box
func (r *Renderer) drawPrediction(img draw.Image, prediction image.Rectangle, col color.RGBA) {
	// Use semi-transparent color for prediction
	predCol := color.RGBA{col.R, col.G, col.B, 128}
	r.drawDashedBoundingBox(img, prediction, predCol)
}

// drawVelocityVector draws velocity vector
func (r *Renderer) drawVelocityVector(img draw.Image, box image.Rectangle, velocity ml.Point3D, col color.RGBA) {
	// Calculate center of bounding box
	centerX := (box.Min.X + box.Max.X) / 2
	centerY := (box.Min.Y + box.Max.Y) / 2

	// Scale velocity for visualization
	scale := 10.0
	endX := centerX + int(float64(velocity.X)*scale)
	endY := centerY + int(float64(velocity.Y)*scale)

	// Draw velocity vector
	r.drawArrow(img, centerX, centerY, endX, endY, col)
}

// drawArrow draws an arrow from (x0,y0) to (x1,y1)
func (r *Renderer) drawArrow(img draw.Image, x0, y0, x1, y1 int, col color.RGBA) {
	// Draw main line
	r.drawLine(img, x0, y0, x1, y1, col)

	// Draw arrowhead
	angle := math.Atan2(float64(y1-y0), float64(x1-x0))
	arrowLength := 8
	arrowAngle := math.Pi / 6 // 30 degrees

	// Calculate arrowhead points
	x2 := x1 - int(float64(arrowLength)*math.Cos(angle-arrowAngle))
	y2 := y1 - int(float64(arrowLength)*math.Sin(angle-arrowAngle))
	x3 := x1 - int(float64(arrowLength)*math.Cos(angle+arrowAngle))
	y3 := y1 - int(float64(arrowLength)*math.Sin(angle+arrowAngle))

	// Draw arrowhead lines
	r.drawLine(img, x1, y1, x2, y2, col)
	r.drawLine(img, x1, y1, x3, y3, col)
}

// drawTrackAge draws track age information
func (r *Renderer) drawTrackAge(img draw.Image, box image.Rectangle, age int, col color.RGBA) {
	ageText := fmt.Sprintf("age:%d", age)
	textX := box.Max.X - 40
	textY := box.Max.Y + 15
	r.drawText(img, ageText, textX, textY, col)
}

// drawTextBackground draws background for text
func (r *Renderer) drawTextBackground(img draw.Image, text string, x, y int, col color.RGBA) {
	textWidth := len(text) * 8
	textHeight := 16

	for py := y - textHeight; py < y && py < img.Bounds().Dy(); py++ {
		for px := x; px < x+textWidth && px < img.Bounds().Dx(); px++ {
			if py >= 0 && px >= 0 {
				bg := color.RGBA{0, 0, 0, 180}
				img.Set(px, py, bg)
			}
		}
	}
}

// hsvToRGB converts HSV color to RGB
func (r *Renderer) hsvToRGB(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var red, green, blue float64
	switch {
	case h < 60:
		red, green, blue = c, x, 0
	case h < 120:
		red, green, blue = x, c, 0
	case h < 180:
		red, green, blue = 0, c, x
	case h < 240:
		red, green, blue = 0, x, c
	case h < 300:
		red, green, blue = x, 0, c
	default:
		red, green, blue = c, 0, x
	}

	return color.RGBA{
		uint8((red + m) * 255),
		uint8((green + m) * 255),
		uint8((blue + m) * 255),
		255,
	}
}
