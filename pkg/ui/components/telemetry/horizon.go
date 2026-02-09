package telemetry

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Horizon represents an artificial horizon display for drone orientation
type Horizon struct {
	Width       int
	Height      int
	Pitch       float64 // degrees
	Roll        float64 // degrees
	ShowLabels  bool
	ShowGrid    bool
	Style       lipgloss.Style
	PitchStyle  lipgloss.Style
	RollStyle   lipgloss.Style
	CenterStyle lipgloss.Style
}

// NewHorizon creates a new artificial horizon display
func NewHorizon(width, height int) *Horizon {
	return &Horizon{
		Width:       width,
		Height:      height,
		Pitch:       0,
		Roll:        0,
		ShowLabels:  true,
		ShowGrid:    true,
		Style:       lipgloss.NewStyle(),
		PitchStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("46")),  // Cyan
		RollStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange
		CenterStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),  // White
	}
}

// SetOrientation sets the pitch and roll angles
func (h *Horizon) SetOrientation(pitch, roll float64) *Horizon {
	h.Pitch = pitch
	h.Roll = roll
	return h
}

// Render renders the artificial horizon as a string
func (h *Horizon) Render() string {
	var builder strings.Builder

	// Calculate grid dimensions
	gridWidth := h.Width
	gridHeight := h.Height

	// Create grid
	grid := make([][]rune, gridHeight)
	for i := range grid {
		grid[i] = make([]rune, gridWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Calculate center position
	centerX := gridWidth / 2
	centerY := gridHeight / 2

	// Convert roll to radians for rotation
	rollRad := h.Roll * math.Pi / 180.0

	// Draw horizon line based on pitch and roll
	h.drawHorizonLine(grid, centerX, centerY, rollRad)

	// Draw center indicator
	h.drawCenterIndicator(grid, centerX, centerY)

	// Draw pitch ladder
	if h.ShowGrid {
		h.drawPitchLadder(grid, centerX, centerY, rollRad)
	}

	// Draw roll indicators
	if h.ShowGrid {
		h.drawRollIndicators(grid, centerX, centerY)
	}

	// Convert grid to string
	for y := 0; y < gridHeight; y++ {
		line := strings.Builder{}
		for x := 0; x < gridWidth; x++ {
			line.WriteRune(grid[y][x])
		}

		// Add labels on the sides
		if h.ShowLabels {
			label := ""
			switch y {
			case 0:
				label = "↑"
			case gridHeight / 4:
				label = "10°"
			case gridHeight / 2:
				label = "0°"
			case 3 * gridHeight / 4:
				label = "-10°"
			case gridHeight - 1:
				label = "↓"
			}

			if label != "" {
				styledLabel := h.PitchStyle.Render(label)
				builder.WriteString(fmt.Sprintf("%-3s%s", styledLabel, line.String()))
			} else {
				builder.WriteString(fmt.Sprintf("   %s", line.String()))
			}
		} else {
			builder.WriteString(line.String())
		}

		builder.WriteString("\n")
	}

	// Add roll indicator at bottom
	if h.ShowLabels {
		builder.WriteString("\n")
		builder.WriteString(h.renderRollIndicator())
	}

	return h.Style.Render(builder.String())
}

// drawHorizonLine draws the horizon line based on pitch and roll
func (h *Horizon) drawHorizonLine(grid [][]rune, centerX, centerY int, rollRad float64) {
	// Calculate pitch offset (each degree = 0.5 rows)
	pitchOffset := int(h.Pitch * 0.5)

	for x := 0; x < len(grid[0]); x++ {
		// Calculate rotated y position
		dx := float64(x - centerX)
		rotatedY := dx * math.Sin(rollRad)

		y := centerY - pitchOffset + int(rotatedY)

		if y >= 0 && y < len(grid) {
			// Choose character based on position
			var ch rune
			if y == centerY-pitchOffset {
				ch = '─' // Horizon line
			} else if y < centerY-pitchOffset {
				ch = '░' // Sky (above horizon)
			} else {
				ch = '▓' // Ground (below horizon)
			}

			// Apply roll-based character
			if math.Abs(dx) < 2 {
				ch = '─'
			}

			grid[y][x] = ch
		}
	}
}

// drawCenterIndicator draws the center aircraft indicator
func (h *Horizon) drawCenterIndicator(grid [][]rune, centerX, centerY int) {
	// Draw center aircraft symbol
	if centerY >= 0 && centerY < len(grid) && centerX >= 0 && centerX < len(grid[0]) {
		grid[centerY][centerX] = '✈'
	}

	// Draw wings
	wingLength := 3
	for i := 1; i <= wingLength; i++ {
		if centerX-i >= 0 && centerY < len(grid) {
			grid[centerY][centerX-i] = '─'
		}
		if centerX+i < len(grid[0]) && centerY < len(grid) {
			grid[centerY][centerX+i] = '─'
		}
	}
}

// drawPitchLadder draws pitch reference lines
func (h *Horizon) drawPitchLadder(grid [][]rune, centerX, centerY int, rollRad float64) {
	// Draw pitch lines at 10 degree intervals
	for pitch := -30; pitch <= 30; pitch += 10 {
		if pitch == 0 {
			continue // Skip 0, it's the horizon line
		}

		// Calculate y offset for this pitch line
		pitchOffset := int(float64(pitch) * 0.5)
		y := centerY - pitchOffset

		if y >= 0 && y < len(grid) {
			// Draw rotated line
			for x := 0; x < len(grid[0]); x++ {
				dx := float64(x - centerX)
				rotatedY := dx * math.Sin(rollRad)
				lineY := y + int(rotatedY)

				if lineY >= 0 && lineY < len(grid) {
					// Draw pitch line with label
					if math.Abs(dx) < 5 {
						grid[lineY][x] = '─'
					} else if math.Abs(dx) == 5 {
						// Add pitch number
						if x < centerX {
							// Left side
							numStr := fmt.Sprintf("%d", int(math.Abs(float64(pitch))))
							for i, ch := range numStr {
								if x-len(numStr)+i >= 0 {
									grid[lineY][x-len(numStr)+i] = ch
								}
							}
						} else {
							// Right side
							numStr := fmt.Sprintf("%d", int(math.Abs(float64(pitch))))
							for i, ch := range numStr {
								if x+i < len(grid[0]) {
									grid[lineY][x+i] = ch
								}
							}
						}
					}
				}
			}
		}
	}
}

// drawRollIndicators draws roll angle indicators
func (h *Horizon) drawRollIndicators(grid [][]rune, centerX, centerY int) {
	radius := min(centerX, centerY) - 2

	// Draw roll circle
	for angle := 0; angle < 360; angle += 10 {
		rad := float64(angle) * math.Pi / 180.0
		x := centerX + int(float64(radius)*math.Cos(rad))
		y := centerY - int(float64(radius)*math.Sin(rad))

		if x >= 0 && x < len(grid[0]) && y >= 0 && y < len(grid) {
			if angle%30 == 0 {
				// Major tick
				grid[y][x] = '+'

				// Add angle label for major ticks
				if angle == 0 {
					label := "0°"
					for i, ch := range label {
						if x+i < len(grid[0]) && y-1 >= 0 {
							grid[y-1][x+i] = ch
						}
					}
				} else if angle == 90 {
					label := "90°"
					for i, ch := range label {
						if x+i < len(grid[0]) {
							grid[y][x+i] = ch
						}
					}
				} else if angle == 180 {
					label := "180°"
					for i, ch := range label {
						if x-len(label)+i >= 0 && y+1 < len(grid) {
							grid[y+1][x-len(label)+i] = ch
						}
					}
				} else if angle == 270 {
					label := "270°"
					for i, ch := range label {
						if x-len(label)+i >= 0 {
							grid[y][x-len(label)+i] = ch
						}
					}
				}
			} else {
				// Minor tick
				grid[y][x] = '·'
			}
		}
	}

	// Draw current roll indicator
	rollRad := h.Roll * math.Pi / 180.0
	indicatorLength := radius + 2

	endX := centerX + int(float64(indicatorLength)*math.Cos(rollRad))
	endY := centerY - int(float64(indicatorLength)*math.Sin(rollRad))

	// Draw line from center to indicator
	h.drawLine(grid, centerX, centerY, endX, endY, '▶')
}

// renderRollIndicator renders a text-based roll indicator
func (h *Horizon) renderRollIndicator() string {
	var builder strings.Builder

	rollStr := fmt.Sprintf("Roll: %5.1f°", h.Roll)
	pitchStr := fmt.Sprintf("Pitch: %5.1f°", h.Pitch)

	builder.WriteString(h.RollStyle.Render(rollStr))
	builder.WriteString("  ")
	builder.WriteString(h.PitchStyle.Render(pitchStr))
	builder.WriteString("\n")

	// Add visual roll indicator
	indicatorWidth := 20
	centerPos := indicatorWidth/2 + int(h.Roll/180.0*float64(indicatorWidth/2))

	builder.WriteString("  [")
	for i := 0; i < indicatorWidth; i++ {
		if i == centerPos {
			builder.WriteString("▲")
		} else if i == indicatorWidth/2 {
			builder.WriteString("|")
		} else {
			builder.WriteString(" ")
		}
	}
	builder.WriteString("]")

	return builder.String()
}

// drawLine draws a line between two points using Bresenham's algorithm
func (h *Horizon) drawLine(grid [][]rune, x0, y0, x1, y1 int, ch rune) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < len(grid[0]) && y0 >= 0 && y0 < len(grid) {
			grid[y0][x0] = ch
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

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetStyle sets the main style for the horizon
func (h *Horizon) SetStyle(style lipgloss.Style) *Horizon {
	h.Style = style
	return h
}

// SetPitchStyle sets the style for pitch elements
func (h *Horizon) SetPitchStyle(style lipgloss.Style) *Horizon {
	h.PitchStyle = style
	return h
}

// SetRollStyle sets the style for roll elements
func (h *Horizon) SetRollStyle(style lipgloss.Style) *Horizon {
	h.RollStyle = style
	return h
}

// SetCenterStyle sets the style for center indicator
func (h *Horizon) SetCenterStyle(style lipgloss.Style) *Horizon {
	h.CenterStyle = style
	return h
}

// SetShowLabels toggles label display
func (h *Horizon) SetShowLabels(show bool) *Horizon {
	h.ShowLabels = show
	return h
}

// SetShowGrid toggles grid display
func (h *Horizon) SetShowGrid(show bool) *Horizon {
	h.ShowGrid = show
	return h
}
