package telemetry

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Graph represents a real-time ASCII graph for telemetry data
type Graph struct {
	Width      int
	Height     int
	Data       []float64
	MinValue   float64
	MaxValue   float64
	Title      string
	Unit       string
	Style      lipgloss.Style
	ShowLabels bool
}

// NewGraph creates a new telemetry graph
func NewGraph(width, height int, title, unit string) *Graph {
	return &Graph{
		Width:      width,
		Height:     height,
		Data:       make([]float64, 0),
		MinValue:   0,
		MaxValue:   100,
		Title:      title,
		Unit:       unit,
		Style:      lipgloss.NewStyle(),
		ShowLabels: true,
	}
}

// AddDataPoint adds a new data point to the graph
func (g *Graph) AddDataPoint(value float64) {
	g.Data = append(g.Data, value)

	// Keep only the last Width data points
	if len(g.Data) > g.Width {
		g.Data = g.Data[len(g.Data)-g.Width:]
	}

	// Update min/max for scaling
	g.updateBounds()
}

// updateBounds updates the min/max values for scaling
func (g *Graph) updateBounds() {
	if len(g.Data) == 0 {
		return
	}

	min := g.Data[0]
	max := g.Data[0]

	for _, v := range g.Data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Add some padding
	rangeVal := max - min
	if rangeVal < 10 {
		rangeVal = 10
	}

	g.MinValue = min - (rangeVal * 0.1)
	g.MaxValue = max + (rangeVal * 0.1)
}

// Render renders the graph as a string
func (g *Graph) Render() string {
	if len(g.Data) == 0 {
		return g.renderEmpty()
	}

	var builder strings.Builder

	// Add title
	if g.Title != "" {
		builder.WriteString(g.Style.Bold(true).Render(g.Title))
		builder.WriteString("\n")
	}

	// Create graph grid
	grid := make([][]rune, g.Height)
	for i := range grid {
		grid[i] = make([]rune, g.Width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot data points
	scaleY := float64(g.Height-1) / (g.MaxValue - g.MinValue)

	for x, value := range g.Data {
		if x >= g.Width {
			break
		}

		// Calculate y position (inverted because terminal coordinates)
		y := int((value - g.MinValue) * scaleY)
		if y < 0 {
			y = 0
		}
		if y >= g.Height {
			y = g.Height - 1
		}

		// Convert to braille character for higher resolution
		grid[g.Height-1-y][x] = getBrailleChar(x, y, value, g.Data)
	}

	// Convert grid to string with y-axis labels
	for y := 0; y < g.Height; y++ {
		if g.ShowLabels && y == 0 {
			// Top label (max value)
			builder.WriteString(fmt.Sprintf("%5.1f%s ", g.MaxValue, g.Unit))
		} else if g.ShowLabels && y == g.Height-1 {
			// Bottom label (min value)
			builder.WriteString(fmt.Sprintf("%5.1f%s ", g.MinValue, g.Unit))
		} else if g.ShowLabels && y == g.Height/2 {
			// Middle label
			midValue := g.MinValue + (g.MaxValue-g.MinValue)/2
			builder.WriteString(fmt.Sprintf("%5.1f%s ", midValue, g.Unit))
		} else {
			builder.WriteString("       ")
		}

		// Add grid line
		for x := 0; x < g.Width; x++ {
			builder.WriteRune(grid[y][x])
		}
		builder.WriteString("\n")
	}

	// Add x-axis time labels
	if g.ShowLabels {
		builder.WriteString("       ")
		builder.WriteString(strings.Repeat("─", g.Width))
		builder.WriteString("\n")
		builder.WriteString("       now")
		if g.Width > 10 {
			builder.WriteString(strings.Repeat(" ", g.Width-10))
			builder.WriteString("-10s")
		}
	}

	return g.Style.Render(builder.String())
}

// renderEmpty renders an empty graph
func (g *Graph) renderEmpty() string {
	var builder strings.Builder

	if g.Title != "" {
		builder.WriteString(g.Style.Bold(true).Render(g.Title))
		builder.WriteString("\n")
	}

	// Empty grid
	for y := 0; y < g.Height; y++ {
		if g.ShowLabels && y == 0 {
			builder.WriteString(fmt.Sprintf("%5.1f%s ", g.MaxValue, g.Unit))
		} else if g.ShowLabels && y == g.Height-1 {
			builder.WriteString(fmt.Sprintf("%5.1f%s ", g.MinValue, g.Unit))
		} else {
			builder.WriteString("       ")
		}

		for x := 0; x < g.Width; x++ {
			builder.WriteRune('·')
		}
		builder.WriteString("\n")
	}

	if g.ShowLabels {
		builder.WriteString("       ")
		builder.WriteString(strings.Repeat("─", g.Width))
		builder.WriteString("\n")
		builder.WriteString("       waiting for data...")
	}

	return g.Style.Render(builder.String())
}

// getBrailleChar returns the appropriate braille character for a data point
func getBrailleChar(x, y int, value float64, data []float64) rune {
	// Braille patterns for different graph styles
	braillePatterns := []rune{
		'⠁', '⠂', '⠄', '⡀', // Single dots
		'⠃', '⠅', '⠆', '⡄', // Two dots
		'⠇', '⠍', '⠎', '⡆', // Three dots
		'⠏', '⠟', '⠿', '⡿', // Four dots (full)
	}

	// Determine intensity based on value change
	intensity := 0
	if x > 0 && len(data) > 1 {
		prev := data[len(data)-2]
		change := math.Abs(value - prev)
		avgChange := (gMax(data) - gMin(data)) / 10

		if change > avgChange*2 {
			intensity = 3
		} else if change > avgChange {
			intensity = 2
		} else if change > 0 {
			intensity = 1
		}
	}

	if intensity >= len(braillePatterns) {
		intensity = len(braillePatterns) - 1
	}

	return braillePatterns[intensity]
}

// Helper functions for min/max
func gMin(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

func gMax(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

// Clear clears all data from the graph
func (g *Graph) Clear() {
	g.Data = make([]float64, 0)
}

// SetStyle sets the style for the graph
func (g *Graph) SetStyle(style lipgloss.Style) *Graph {
	g.Style = style
	return g
}

// SetBounds sets custom min/max bounds for the graph
func (g *Graph) SetBounds(min, max float64) *Graph {
	g.MinValue = min
	g.MaxValue = max
	return g
}
