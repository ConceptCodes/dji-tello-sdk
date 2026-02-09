package safety

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// ProximityWarning represents a proximity warning visualization
type ProximityWarning struct {
	Width         int
	Height        int
	CurrentState  *types.State
	Obstacles     []Obstacle
	WarningZones  []WarningZone
	ShowRadar     bool
	ShowWarnings  bool
	ShowDistances bool
	RadarRange    int // cm
	Style         lipgloss.Style
	HeaderStyle   lipgloss.Style
	SafeStyle     lipgloss.Style
	WarningStyle  lipgloss.Style
	CriticalStyle lipgloss.Style
	RadarStyle    lipgloss.Style
	ObstacleStyle lipgloss.Style
	DistanceStyle lipgloss.Style
}

// Obstacle represents a detected obstacle
type Obstacle struct {
	ID         string
	X          int // cm relative to drone
	Y          int // cm relative to drone
	Z          int // cm relative to drone
	Distance   int // cm
	Type       string
	Confidence float64
	Size       int // cm approximate size
}

// WarningZone represents a warning zone around the drone
type WarningZone struct {
	Distance int
	Level    string // "safe", "warning", "critical"
}

// NewProximityWarning creates a new proximity warning visualization
func NewProximityWarning(width, height int) *ProximityWarning {
	return &ProximityWarning{
		Width:        width,
		Height:       height,
		CurrentState: &types.State{},
		Obstacles:    make([]Obstacle, 0),
		WarningZones: []WarningZone{
			{Distance: 200, Level: "safe"},
			{Distance: 100, Level: "warning"},
			{Distance: 50, Level: "critical"},
		},
		ShowRadar:     true,
		ShowWarnings:  true,
		ShowDistances: true,
		RadarRange:    300, // cm
		Style:         lipgloss.NewStyle().Padding(0, 1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Padding(0, 0, 1, 0),
		SafeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")), // Green
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // Orange
		CriticalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // Red
		RadarStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")), // Gray
		ObstacleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		DistanceStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // White
	}
}

// UpdateState updates the proximity warning with new drone state
func (p *ProximityWarning) UpdateState(state *types.State) {
	p.CurrentState = state
}

// AddObstacle adds a detected obstacle
func (p *ProximityWarning) AddObstacle(obstacle Obstacle) {
	// Check if obstacle already exists
	for i, existing := range p.Obstacles {
		if existing.ID == obstacle.ID {
			p.Obstacles[i] = obstacle
			return
		}
	}

	// Add new obstacle
	p.Obstacles = append(p.Obstacles, obstacle)

	// Keep only closest obstacles (max 10)
	if len(p.Obstacles) > 10 {
		// Sort by distance and keep closest
		p.sortObstaclesByDistance()
		p.Obstacles = p.Obstacles[:10]
	}
}

// RemoveObstacle removes an obstacle by ID
func (p *ProximityWarning) RemoveObstacle(id string) {
	for i, obstacle := range p.Obstacles {
		if obstacle.ID == id {
			p.Obstacles = append(p.Obstacles[:i], p.Obstacles[i+1:]...)
			return
		}
	}
}

// ClearObstacles clears all obstacles
func (p *ProximityWarning) ClearObstacles() {
	p.Obstacles = make([]Obstacle, 0)
}

// sortObstaclesByDistance sorts obstacles by distance (closest first)
func (p *ProximityWarning) sortObstaclesByDistance() {
	for i := 0; i < len(p.Obstacles); i++ {
		for j := i + 1; j < len(p.Obstacles); j++ {
			if p.Obstacles[j].Distance < p.Obstacles[i].Distance {
				p.Obstacles[i], p.Obstacles[j] = p.Obstacles[j], p.Obstacles[i]
			}
		}
	}
}

// Render renders the proximity warning visualization
func (p *ProximityWarning) Render() string {
	var builder strings.Builder

	// Add header
	builder.WriteString(p.renderHeader())
	builder.WriteString("\n\n")

	// Calculate layout
	leftWidth := p.Width/2 - 2
	rightWidth := p.Width/2 - 2

	// Create left and right columns
	leftColumn := strings.Builder{}
	rightColumn := strings.Builder{}

	// Left column: Radar visualization
	if p.ShowRadar {
		leftColumn.WriteString(p.renderRadar())
		leftColumn.WriteString("\n\n")
	}

	// Right column: Warnings and distances
	if p.ShowWarnings {
		rightColumn.WriteString(p.renderWarnings())
		rightColumn.WriteString("\n\n")
	}

	if p.ShowDistances {
		rightColumn.WriteString(p.renderDistances())
	}

	// Combine columns
	leftContent := leftColumn.String()
	rightContent := rightColumn.String()

	// Use lipgloss to create columns
	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftWidth).Render(leftContent),
		lipgloss.NewStyle().Width(4).Render(""), // Spacer
		lipgloss.NewStyle().Width(rightWidth).Render(rightContent),
	)

	builder.WriteString(combined)
	builder.WriteString("\n")

	// Add footer
	builder.WriteString(p.renderFooter())

	return p.Style.Render(builder.String())
}

// renderHeader renders the proximity warning header
func (p *ProximityWarning) renderHeader() string {
	// Check for critical obstacles
	hasCritical := false
	hasWarning := false

	for _, obstacle := range p.Obstacles {
		if obstacle.Distance < 50 {
			hasCritical = true
		} else if obstacle.Distance < 100 {
			hasWarning = true
		}
	}

	status := "CLEAR"
	statusStyle := p.SafeStyle

	if hasCritical {
		status = "CRITICAL"
		statusStyle = p.CriticalStyle
	} else if hasWarning {
		status = "WARNING"
		statusStyle = p.WarningStyle
	}

	obstacleCount := len(p.Obstacles)
	header := fmt.Sprintf("📡 PROXIMITY WARNING  %s  [%d obstacles]", statusStyle.Render(status), obstacleCount)
	return p.HeaderStyle.Render(header)
}

// renderRadar renders the radar visualization
func (p *ProximityWarning) renderRadar() string {
	var builder strings.Builder

	builder.WriteString(p.DistanceStyle.Bold(true).Render("Radar View"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	// Create radar grid
	gridSize := 15
	center := gridSize / 2

	// Initialize grid
	grid := make([][]rune, gridSize)
	for i := range grid {
		grid[i] = make([]rune, gridSize)
		for j := range grid[i] {
			grid[i][j] = '·'
		}
	}

	// Draw warning zones
	for _, zone := range p.WarningZones {
		radius := int(float64(zone.Distance) / float64(p.RadarRange) * float64(center))

		for angle := 0; angle < 360; angle += 5 {
			rad := float64(angle) * math.Pi / 180.0
			x := center + int(float64(radius)*math.Cos(rad))
			y := center + int(float64(radius)*math.Sin(rad))

			if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
				var ch rune
				switch zone.Level {
				case "critical":
					ch = '×'
				case "warning":
					ch = '○'
				default:
					ch = '·'
				}
				grid[y][x] = ch
			}
		}
	}

	// Draw obstacles
	for _, obstacle := range p.Obstacles {
		// Convert obstacle position to grid coordinates
		scale := float64(center) / float64(p.RadarRange)
		x := center + int(float64(obstacle.X)*scale)
		y := center + int(float64(obstacle.Y)*scale)

		if x >= 0 && x < gridSize && y >= 0 && y < gridSize {
			// Choose symbol based on distance
			var ch rune
			if obstacle.Distance < 50 {
				ch = '⨉' // Critical
			} else if obstacle.Distance < 100 {
				ch = '○' // Warning
			} else {
				ch = '·' // Safe
			}
			grid[y][x] = ch
		}
	}

	// Draw drone at center
	grid[center][center] = '✈'

	// Convert grid to string
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			builder.WriteRune(grid[y][x])
			builder.WriteRune(' ')
		}
		builder.WriteString("\n")
	}

	// Add compass
	builder.WriteString("\n")
	builder.WriteString("      N\n")
	builder.WriteString("      ↑\n")
	builder.WriteString("  W ← ✈ → E\n")
	builder.WriteString("      ↓\n")
	builder.WriteString("      S\n")

	return builder.String()
}

// renderWarnings renders proximity warnings
func (p *ProximityWarning) renderWarnings() string {
	var builder strings.Builder

	builder.WriteString(p.DistanceStyle.Bold(true).Render("Proximity Warnings"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	if len(p.Obstacles) == 0 {
		builder.WriteString(p.SafeStyle.Render("No obstacles detected"))
		builder.WriteString("\n")
		builder.WriteString(p.SafeStyle.Render("Clear airspace"))
		return builder.String()
	}

	// Sort obstacles by distance
	p.sortObstaclesByDistance()

	// Show up to 5 closest obstacles
	displayCount := minInt(5, len(p.Obstacles))
	for i := 0; i < displayCount; i++ {
		obstacle := p.Obstacles[i]

		// Get warning level and style
		var warningLevel string
		var warningStyle lipgloss.Style

		if obstacle.Distance < 50 {
			warningLevel = "CRITICAL"
			warningStyle = p.CriticalStyle
		} else if obstacle.Distance < 100 {
			warningLevel = "WARNING"
			warningStyle = p.WarningStyle
		} else {
			warningLevel = "SAFE"
			warningStyle = p.SafeStyle
		}

		// Get direction
		direction := p.getDirection(obstacle.X, obstacle.Y)

		// Format obstacle info
		builder.WriteString(fmt.Sprintf("%s %s: %dcm %s\n",
			p.getObstacleIcon(obstacle.Type),
			warningStyle.Render(warningLevel),
			obstacle.Distance,
			direction))

		// Add type and confidence if available
		if obstacle.Type != "" {
			builder.WriteString(fmt.Sprintf("  Type: %s", obstacle.Type))
			if obstacle.Confidence > 0 {
				builder.WriteString(fmt.Sprintf(" (%.0f%%)", obstacle.Confidence*100))
			}
			builder.WriteString("\n")
		}

		// Add relative position
		builder.WriteString(fmt.Sprintf("  Position: X:%d Y:%d Z:%d cm\n",
			obstacle.X, obstacle.Y, obstacle.Z))
	}

	if len(p.Obstacles) > displayCount {
		builder.WriteString(p.DistanceStyle.Render(fmt.Sprintf("... and %d more obstacles", len(p.Obstacles)-displayCount)))
	}

	return builder.String()
}

// renderDistances renders distance information
func (p *ProximityWarning) renderDistances() string {
	var builder strings.Builder

	builder.WriteString(p.DistanceStyle.Bold(true).Render("Distance Information"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	// Current height
	currentHeight := 0
	if p.CurrentState != nil {
		currentHeight = p.CurrentState.H
	}

	builder.WriteString(fmt.Sprintf("Current Height: %d cm\n", currentHeight))

	// TOF distance
	tofDistance := 0
	if p.CurrentState != nil {
		tofDistance = p.CurrentState.Tof
	}

	tofStyle := p.SafeStyle
	if tofDistance < 50 {
		tofStyle = p.CriticalStyle
	} else if tofDistance < 100 {
		tofStyle = p.WarningStyle
	}

	builder.WriteString(fmt.Sprintf("TOF Distance: %s\n", tofStyle.Render(fmt.Sprintf("%d cm", tofDistance))))

	// Closest obstacle
	if len(p.Obstacles) > 0 {
		p.sortObstaclesByDistance()
		closest := p.Obstacles[0]

		closestStyle := p.SafeStyle
		if closest.Distance < 50 {
			closestStyle = p.CriticalStyle
		} else if closest.Distance < 100 {
			closestStyle = p.WarningStyle
		}

		builder.WriteString(fmt.Sprintf("Closest Obstacle: %s\n",
			closestStyle.Render(fmt.Sprintf("%d cm", closest.Distance))))

		// Time to collision (simplified)
		if p.CurrentState != nil {
			// Calculate approximate time based on velocity toward obstacle
			speed := proximitySqrt(float64(p.CurrentState.Vgx*p.CurrentState.Vgx +
				p.CurrentState.Vgy*p.CurrentState.Vgy +
				p.CurrentState.Vgz*p.CurrentState.Vgz))

			if speed > 0 {
				timeToCollision := float64(closest.Distance) / speed
				if timeToCollision < 5 {
					builder.WriteString(p.CriticalStyle.Render(fmt.Sprintf("Time to collision: %.1f s", timeToCollision)))
				} else if timeToCollision < 10 {
					builder.WriteString(p.WarningStyle.Render(fmt.Sprintf("Time to collision: %.1f s", timeToCollision)))
				} else {
					builder.WriteString(fmt.Sprintf("Time to collision: %.1f s", timeToCollision))
				}
			}
		}
	} else {
		builder.WriteString(p.SafeStyle.Render("No obstacles in range"))
	}

	return builder.String()
}

// renderFooter renders the footer
func (p *ProximityWarning) renderFooter() string {
	controls := "F10: Toggle Radar | F11: Toggle Warnings | F12: Toggle Distances"

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Faint(true).
		Render(controls)
}

// getDirection returns the compass direction from relative coordinates
func (p *ProximityWarning) getDirection(x, y int) string {
	if x == 0 && y == 0 {
		return "directly ahead"
	}

	angle := math.Atan2(float64(y), float64(x)) * 180 / math.Pi

	// Normalize angle to 0-360
	if angle < 0 {
		angle += 360
	}

	// Determine direction
	switch {
	case angle >= 337.5 || angle < 22.5:
		return "E"
	case angle >= 22.5 && angle < 67.5:
		return "NE"
	case angle >= 67.5 && angle < 112.5:
		return "N"
	case angle >= 112.5 && angle < 157.5:
		return "NW"
	case angle >= 157.5 && angle < 202.5:
		return "W"
	case angle >= 202.5 && angle < 247.5:
		return "SW"
	case angle >= 247.5 && angle < 292.5:
		return "S"
	case angle >= 292.5 && angle < 337.5:
		return "SE"
	default:
		return ""
	}
}

// getObstacleIcon returns an icon for the obstacle type
func (p *ProximityWarning) getObstacleIcon(obstacleType string) string {
	switch obstacleType {
	case "person":
		return "👤"
	case "vehicle":
		return "🚗"
	case "tree":
		return "🌳"
	case "building":
		return "🏢"
	case "wall":
		return "🧱"
	case "furniture":
		return "🪑"
	default:
		return "●"
	}
}

// ToggleRadar toggles radar display
func (p *ProximityWarning) ToggleRadar() {
	p.ShowRadar = !p.ShowRadar
}

// ToggleWarnings toggles warnings display
func (p *ProximityWarning) ToggleWarnings() {
	p.ShowWarnings = !p.ShowWarnings
}

// ToggleDistances toggles distances display
func (p *ProximityWarning) ToggleDistances() {
	p.ShowDistances = !p.ShowDistances
}

// SetStyle sets the main style
func (p *ProximityWarning) SetStyle(style lipgloss.Style) *ProximityWarning {
	p.Style = style
	return p
}

// Helper function for square root (renamed to avoid conflict)
func proximitySqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}

	// Babylonian method (one iteration for speed)
	y := x
	for i := 0; i < 3; i++ {
		y = (y + x/y) / 2
	}
	return y
}

// Helper function for min
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
