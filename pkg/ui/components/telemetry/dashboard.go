package telemetry

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// Dashboard represents an enhanced telemetry dashboard
type Dashboard struct {
	Width        int
	Height       int
	State        *types.State
	History      []*types.State
	MaxHistory   int
	Graphs       map[string]*Graph
	Horizon      *Horizon
	ShowGraphs   bool
	ShowHorizon  bool
	ShowMetrics  bool
	LastUpdate   time.Time
	UpdateRate   time.Duration
	Style        lipgloss.Style
	HeaderStyle  lipgloss.Style
	MetricStyle  lipgloss.Style
	WarningStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
}

// NewDashboard creates a new enhanced telemetry dashboard
func NewDashboard(width, height int) *Dashboard {
	dash := &Dashboard{
		Width:       width,
		Height:      height,
		State:       &types.State{},
		History:     make([]*types.State, 0),
		MaxHistory:  100,
		Graphs:      make(map[string]*Graph),
		ShowGraphs:  true,
		ShowHorizon: true,
		ShowMetrics: true,
		LastUpdate:  time.Now(),
		UpdateRate:  100 * time.Millisecond,
		Style:       lipgloss.NewStyle().Padding(0, 1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 0, 1, 0),
		MetricStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
	}

	// Initialize graphs
	graphWidth := width/2 - 4
	graphHeight := 8

	dash.Graphs["battery"] = NewGraph(graphWidth, graphHeight, "Battery", "%")
	dash.Graphs["battery"].SetBounds(0, 100)

	dash.Graphs["height"] = NewGraph(graphWidth, graphHeight, "Height", "cm")
	dash.Graphs["height"].SetBounds(0, 500)

	dash.Graphs["temperature"] = NewGraph(graphWidth, graphHeight, "Temperature", "°C")
	dash.Graphs["temperature"].SetBounds(0, 80)

	dash.Graphs["velocity"] = NewGraph(graphWidth, graphHeight, "Velocity", "cm/s")
	dash.Graphs["velocity"].SetBounds(-100, 100)

	// Initialize horizon
	horizonWidth := width/2 - 4
	horizonHeight := 12
	dash.Horizon = NewHorizon(horizonWidth, horizonHeight)

	return dash
}

// UpdateState updates the dashboard with new telemetry state
func (d *Dashboard) UpdateState(state *types.State) {
	d.State = state
	d.LastUpdate = time.Now()

	// Add to history
	d.History = append(d.History, state)
	if len(d.History) > d.MaxHistory {
		d.History = d.History[len(d.History)-d.MaxHistory:]
	}

	// Update graphs
	if d.ShowGraphs {
		d.Graphs["battery"].AddDataPoint(float64(state.Bat))
		d.Graphs["height"].AddDataPoint(float64(state.H))
		d.Graphs["temperature"].AddDataPoint(float64(state.Temph))

		// Calculate total velocity
		velocity := mathSqrt(float64(state.Vgx*state.Vgx + state.Vgy*state.Vgy + state.Vgz*state.Vgz))
		d.Graphs["velocity"].AddDataPoint(velocity)
	}

	// Update horizon
	if d.ShowHorizon {
		d.Horizon.SetOrientation(float64(state.Pitch), float64(state.Roll))
	}
}

// Render renders the entire dashboard
func (d *Dashboard) Render() string {
	var builder strings.Builder

	// Add header
	builder.WriteString(d.renderHeader())
	builder.WriteString("\n\n")

	// Calculate layout
	leftWidth := d.Width/2 - 2
	rightWidth := d.Width/2 - 2

	// Create left and right columns
	leftColumn := strings.Builder{}
	rightColumn := strings.Builder{}

	// Left column: Horizon and metrics
	if d.ShowHorizon {
		leftColumn.WriteString(d.Horizon.Render())
		leftColumn.WriteString("\n\n")
	}

	if d.ShowMetrics {
		leftColumn.WriteString(d.renderMetrics())
	}

	// Right column: Graphs
	if d.ShowGraphs {
		rightColumn.WriteString(d.renderGraphs())
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
	builder.WriteString(d.renderFooter())

	return d.Style.Render(builder.String())
}

// renderHeader renders the dashboard header
func (d *Dashboard) renderHeader() string {
	status := "DISCONNECTED"
	statusStyle := d.ErrorStyle

	if d.State != nil && d.LastUpdate.Add(2*time.Second).After(time.Now()) {
		status = "CONNECTED"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	}

	header := fmt.Sprintf("📊 ENHANCED TELEMETRY DASHBOARD  %s", statusStyle.Render(status))
	return d.HeaderStyle.Render(header)
}

// renderMetrics renders key metrics display
func (d *Dashboard) renderMetrics() string {
	if d.State == nil {
		return d.MetricStyle.Render("No telemetry data available")
	}

	var builder strings.Builder

	// Battery with warning/error colors
	batteryStr := fmt.Sprintf("Battery: %d%%", d.State.Bat)
	batteryStyle := d.MetricStyle
	if d.State.Bat < 20 {
		batteryStyle = d.ErrorStyle
	} else if d.State.Bat < 40 {
		batteryStyle = d.WarningStyle
	}
	builder.WriteString(batteryStyle.Render(batteryStr))
	builder.WriteString("\n")

	// Height
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Height: %d cm", d.State.H)))
	builder.WriteString("\n")

	// Temperature with warning
	tempStr := fmt.Sprintf("Temperature: %d°C", d.State.Temph)
	tempStyle := d.MetricStyle
	if d.State.Temph > 60 {
		tempStyle = d.ErrorStyle
	} else if d.State.Temph > 50 {
		tempStyle = d.WarningStyle
	}
	builder.WriteString(tempStyle.Render(tempStr))
	builder.WriteString("\n")

	// Flight time
	minutes := d.State.Time / 60
	seconds := d.State.Time % 60
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Flight Time: %02d:%02d", minutes, seconds)))
	builder.WriteString("\n")

	// Velocity
	velocity := mathSqrt(float64(d.State.Vgx*d.State.Vgx + d.State.Vgy*d.State.Vgy + d.State.Vgz*d.State.Vgz))
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Velocity: %.1f cm/s", velocity)))
	builder.WriteString("\n")

	// Orientation
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Pitch: %d°  Roll: %d°  Yaw: %d°",
		d.State.Pitch, d.State.Roll, d.State.Yaw)))
	builder.WriteString("\n")

	// Barometer
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Barometer: %.1f m", d.State.Baro)))
	builder.WriteString("\n")

	// Acceleration
	accel := mathSqrt(d.State.Agx*d.State.Agx + d.State.Agy*d.State.Agy + d.State.Agz*d.State.Agz)
	builder.WriteString(d.MetricStyle.Render(fmt.Sprintf("Acceleration: %.2f g", accel)))

	return builder.String()
}

// renderGraphs renders all telemetry graphs
func (d *Dashboard) renderGraphs() string {
	var builder strings.Builder

	// Render graphs in a grid
	graphs := []string{"battery", "height", "temperature", "velocity"}

	for i, graphName := range graphs {
		if graph, ok := d.Graphs[graphName]; ok {
			builder.WriteString(graph.Render())
			if i < len(graphs)-1 {
				builder.WriteString("\n\n")
			}
		}
	}

	return builder.String()
}

// renderFooter renders dashboard footer with update info
func (d *Dashboard) renderFooter() string {
	updateAge := time.Since(d.LastUpdate)
	ageStr := fmt.Sprintf("Last update: %.1fs ago", updateAge.Seconds())

	if updateAge > 5*time.Second {
		ageStr = d.ErrorStyle.Render(ageStr)
	} else if updateAge > 2*time.Second {
		ageStr = d.WarningStyle.Render(ageStr)
	} else {
		ageStr = d.MetricStyle.Render(ageStr)
	}

	footer := fmt.Sprintf("%s | F2: Toggle Graphs | F3: Toggle Horizon | F4: Toggle Metrics", ageStr)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Faint(true).
		Render(footer)
}

// ToggleGraphs toggles graph display
func (d *Dashboard) ToggleGraphs() {
	d.ShowGraphs = !d.ShowGraphs
}

// ToggleHorizon toggles horizon display
func (d *Dashboard) ToggleHorizon() {
	d.ShowHorizon = !d.ShowHorizon
}

// ToggleMetrics toggles metrics display
func (d *Dashboard) ToggleMetrics() {
	d.ShowMetrics = !d.ShowMetrics
}

// ClearHistory clears all historical data
func (d *Dashboard) ClearHistory() {
	d.History = make([]*types.State, 0)
	for _, graph := range d.Graphs {
		graph.Clear()
	}
}

// SetStyle sets the main dashboard style
func (d *Dashboard) SetStyle(style lipgloss.Style) *Dashboard {
	d.Style = style
	return d
}

// SetHeaderStyle sets the header style
func (d *Dashboard) SetHeaderStyle(style lipgloss.Style) *Dashboard {
	d.HeaderStyle = style
	return d
}

// SetMetricStyle sets the metric style
func (d *Dashboard) SetMetricStyle(style lipgloss.Style) *Dashboard {
	d.MetricStyle = style
	return d
}

// SetWarningStyle sets the warning style
func (d *Dashboard) SetWarningStyle(style lipgloss.Style) *Dashboard {
	d.WarningStyle = style
	return d
}

// SetErrorStyle sets the error style
func (d *Dashboard) SetErrorStyle(style lipgloss.Style) *Dashboard {
	d.ErrorStyle = style
	return d
}

// Helper function for square root
func mathSqrt(x float64) float64 {
	// Simple square root approximation
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
