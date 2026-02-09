package safety

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// Dashboard represents a safety monitoring dashboard
type Dashboard struct {
	Width           int
	Height          int
	SafetyStatus    *safety.SafetyStatus
	SafetyConfig    *safety.Config
	CurrentState    *types.State
	ActiveEvents    []safety.SafetyEvent
	EventHistory    []safety.SafetyEvent
	MaxEventHistory int
	ShowDetails     bool
	ShowLimits      bool
	ShowEvents      bool
	LastUpdate      time.Time
	UpdateRate      time.Duration
	Style           lipgloss.Style
	HeaderStyle     lipgloss.Style
	SafeStyle       lipgloss.Style
	WarningStyle    lipgloss.Style
	CriticalStyle   lipgloss.Style
	EmergencyStyle  lipgloss.Style
	LimitStyle      lipgloss.Style
	ValueStyle      lipgloss.Style
}

// NewDashboard creates a new safety monitoring dashboard
func NewDashboard(width, height int) *Dashboard {
	dash := &Dashboard{
		Width:           width,
		Height:          height,
		SafetyStatus:    safety.NewSafetyStatus(),
		SafetyConfig:    loadDefaultConfig(),
		CurrentState:    &types.State{},
		ActiveEvents:    make([]safety.SafetyEvent, 0),
		EventHistory:    make([]safety.SafetyEvent, 0),
		MaxEventHistory: 50,
		ShowDetails:     true,
		ShowLimits:      true,
		ShowEvents:      true,
		LastUpdate:      time.Now(),
		UpdateRate:      100 * time.Millisecond,
		Style:           lipgloss.NewStyle().Padding(0, 1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 0, 1, 0),
		SafeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")), // Green
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // Orange
		CriticalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // Red
		EmergencyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Blink(true), // Red blinking
		LimitStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")), // Gray
		ValueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")), // White
	}

	return dash
}

// loadDefaultConfig loads the default safety configuration
func loadDefaultConfig() *safety.Config {
	return &safety.Config{
		Version: "1.0.0",
		Level:   safety.SafetyLevelNormal,
		Altitude: safety.AltitudeLimits{
			MinHeight:     20,
			MaxHeight:     300,
			TakeoffHeight: 50,
		},
		Velocity: safety.VelocityLimits{
			MaxHorizontal: 100,
			MaxVertical:   80,
			MaxYaw:        100,
		},
		Battery: safety.BatterySafety{
			WarningThreshold:   30,
			CriticalThreshold:  20,
			EmergencyThreshold: 15,
			EnableAutoLand:     true,
			LowBatteryAction:   "land",
		},
		Sensors: safety.SensorSafety{
			MinTOFDistance:      30,
			MaxTiltAngle:        30,
			MaxAcceleration:     2.0,
			BaroPressureDelta:   5.0,
			SensorFailureAction: "land",
		},
		Emergency: safety.EmergencyProcedures{
			ConnectionTimeout:   3000,
			SensorFailureAction: "land",
			EnableAutoLand:      true,
			LowBatteryAction:    "land",
		},
		Behavioral: safety.BehavioralLimits{
			EnableFlips:    true,
			MinFlipHeight:  100,
			MaxFlightTime:  600,
			MaxCommandRate: 10,
		},
	}
}

// UpdateSafetyStatus updates the dashboard with new safety status
func (d *Dashboard) UpdateSafetyStatus(status *safety.SafetyStatus, state *types.State) {
	d.SafetyStatus = status
	d.CurrentState = state
	d.LastUpdate = time.Now()

	// Update active events
	if status != nil {
		d.ActiveEvents = status.ActiveEvents

		// Add to event history
		for _, event := range status.ActiveEvents {
			d.addToEventHistory(event)
		}

		if status.LastEvent != nil {
			d.addToEventHistory(*status.LastEvent)
		}
	}
}

// addToEventHistory adds an event to the history
func (d *Dashboard) addToEventHistory(event safety.SafetyEvent) {
	// Check if event already exists in recent history
	for i, existing := range d.EventHistory {
		if existing.Timestamp.Equal(event.Timestamp) && existing.Message == event.Message {
			// Update existing event
			d.EventHistory[i] = event
			return
		}
	}

	// Add new event
	d.EventHistory = append([]safety.SafetyEvent{event}, d.EventHistory...)

	// Trim history if too long
	if len(d.EventHistory) > d.MaxEventHistory {
		d.EventHistory = d.EventHistory[:d.MaxEventHistory]
	}
}

// Render renders the safety dashboard
func (d *Dashboard) Render() string {
	var builder strings.Builder

	// Add header with safety status
	builder.WriteString(d.renderHeader())
	builder.WriteString("\n\n")

	// Calculate layout
	leftWidth := d.Width/2 - 2
	rightWidth := d.Width/2 - 2

	// Create left and right columns
	leftColumn := strings.Builder{}
	rightColumn := strings.Builder{}

	// Left column: Safety status and limits
	leftColumn.WriteString(d.renderSafetyStatus())
	leftColumn.WriteString("\n\n")

	if d.ShowLimits {
		leftColumn.WriteString(d.renderSafetyLimits())
		leftColumn.WriteString("\n\n")
	}

	// Right column: Active events and details
	if d.ShowEvents && len(d.ActiveEvents) > 0 {
		rightColumn.WriteString(d.renderActiveEvents())
		rightColumn.WriteString("\n\n")
	}

	if d.ShowDetails {
		rightColumn.WriteString(d.renderSafetyDetails())
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
	statusText := "SAFE"
	statusStyle := d.SafeStyle

	if d.SafetyStatus != nil {
		if d.SafetyStatus.EmergencyMode {
			statusText = "EMERGENCY"
			statusStyle = d.EmergencyStyle
		} else if !d.SafetyStatus.IsSafe {
			statusText = "UNSAFE"
			statusStyle = d.CriticalStyle
		} else if len(d.ActiveEvents) > 0 {
			// Check if any events are critical or warning
			hasCritical := false
			hasWarning := false

			for _, event := range d.ActiveEvents {
				switch event.Level {
				case "critical", "emergency":
					hasCritical = true
				case "warning":
					hasWarning = true
				}
			}

			if hasCritical {
				statusText = "CRITICAL"
				statusStyle = d.CriticalStyle
			} else if hasWarning {
				statusText = "WARNING"
				statusStyle = d.WarningStyle
			}
		}
	}

	configLevel := "Normal"
	if d.SafetyConfig != nil {
		configLevel = string(d.SafetyConfig.Level)
	}

	header := fmt.Sprintf("🛡️ SAFETY MONITORING  %s  [%s]", statusStyle.Render(statusText), configLevel)
	return d.HeaderStyle.Render(header)
}

// renderSafetyStatus renders the current safety status
func (d *Dashboard) renderSafetyStatus() string {
	var builder strings.Builder

	builder.WriteString(d.ValueStyle.Bold(true).Render("Safety Status"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	// Overall status
	statusIcon := "✅"
	statusText := "All systems normal"
	statusStyle := d.SafeStyle

	if d.SafetyStatus != nil {
		if d.SafetyStatus.EmergencyMode {
			statusIcon = "🚨"
			statusText = "EMERGENCY MODE ACTIVE"
			statusStyle = d.EmergencyStyle
		} else if !d.SafetyStatus.IsSafe {
			statusIcon = "❌"
			statusText = "Safety violations detected"
			statusStyle = d.CriticalStyle
		} else if len(d.ActiveEvents) > 0 {
			statusIcon = "⚠️"
			statusText = "Warnings active"
			statusStyle = d.WarningStyle
		}
	}

	builder.WriteString(fmt.Sprintf("%s %s\n", statusIcon, statusStyle.Render(statusText)))
	builder.WriteString("\n")

	// Active event count
	eventCount := len(d.ActiveEvents)
	eventText := fmt.Sprintf("Active events: %d", eventCount)
	if eventCount == 0 {
		builder.WriteString(d.SafeStyle.Render(eventText))
	} else if eventCount <= 2 {
		builder.WriteString(d.WarningStyle.Render(eventText))
	} else {
		builder.WriteString(d.CriticalStyle.Render(eventText))
	}
	builder.WriteString("\n")

	// Safety enabled
	safetyEnabled := "Enabled"
	if d.SafetyStatus != nil && !d.SafetyStatus.SafetyEnabled {
		safetyEnabled = "DISABLED"
		builder.WriteString(d.CriticalStyle.Render(fmt.Sprintf("Safety system: %s", safetyEnabled)))
	} else {
		builder.WriteString(d.SafeStyle.Render(fmt.Sprintf("Safety system: %s", safetyEnabled)))
	}
	builder.WriteString("\n")

	// Last update
	updateAge := time.Since(d.LastUpdate)
	ageText := fmt.Sprintf("Last update: %.1fs ago", updateAge.Seconds())
	if updateAge > 5*time.Second {
		builder.WriteString(d.CriticalStyle.Render(ageText))
	} else if updateAge > 2*time.Second {
		builder.WriteString(d.WarningStyle.Render(ageText))
	} else {
		builder.WriteString(d.ValueStyle.Render(ageText))
	}

	return builder.String()
}

// renderSafetyLimits renders the current safety limits
func (d *Dashboard) renderSafetyLimits() string {
	var builder strings.Builder

	builder.WriteString(d.ValueStyle.Bold(true).Render("Safety Limits"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	if d.SafetyConfig == nil {
		builder.WriteString(d.ValueStyle.Render("No configuration loaded"))
		return builder.String()
	}

	// Altitude limits
	minHeight := d.SafetyConfig.Altitude.MinHeight
	maxHeight := d.SafetyConfig.Altitude.MaxHeight
	currentHeight := 0
	if d.CurrentState != nil {
		currentHeight = d.CurrentState.H
	}

	heightStatus := d.SafeStyle
	if currentHeight < minHeight {
		heightStatus = d.WarningStyle
	} else if currentHeight > maxHeight {
		heightStatus = d.CriticalStyle
	}

	builder.WriteString(fmt.Sprintf("Altitude: %s/%d/%d cm\n",
		heightStatus.Render(fmt.Sprintf("%d", currentHeight)),
		minHeight,
		maxHeight))

	// Battery limits
	warningThreshold := d.SafetyConfig.Battery.WarningThreshold
	criticalThreshold := d.SafetyConfig.Battery.CriticalThreshold
	currentBattery := 0
	if d.CurrentState != nil {
		currentBattery = d.CurrentState.Bat
	}

	batteryStatus := d.SafeStyle
	if currentBattery <= criticalThreshold {
		batteryStatus = d.CriticalStyle
	} else if currentBattery <= warningThreshold {
		batteryStatus = d.WarningStyle
	}

	builder.WriteString(fmt.Sprintf("Battery: %s/%d/%d%%\n",
		batteryStatus.Render(fmt.Sprintf("%d", currentBattery)),
		warningThreshold,
		criticalThreshold))

	// Velocity limits
	maxHorizontal := d.SafetyConfig.Velocity.MaxHorizontal
	maxVertical := d.SafetyConfig.Velocity.MaxVertical
	currentSpeed := 0
	if d.CurrentState != nil {
		// Calculate total velocity
		vx := d.CurrentState.Vgx
		vy := d.CurrentState.Vgy
		vz := d.CurrentState.Vgz
		currentSpeed = int(mathSqrt(float64(vx*vx + vy*vy + vz*vz)))
	}

	speedStatus := d.SafeStyle
	if currentSpeed > maxHorizontal {
		speedStatus = d.CriticalStyle
	}

	builder.WriteString(fmt.Sprintf("Speed: %s/%d cm/s (V: %d)\n",
		speedStatus.Render(fmt.Sprintf("%d", currentSpeed)),
		maxHorizontal,
		maxVertical))

	// Temperature limits
	currentTemp := 0
	if d.CurrentState != nil {
		currentTemp = d.CurrentState.Temph
	}

	tempStatus := d.SafeStyle
	if currentTemp > 60 {
		tempStatus = d.CriticalStyle
	} else if currentTemp > 50 {
		tempStatus = d.WarningStyle
	}

	builder.WriteString(fmt.Sprintf("Temperature: %s°C\n",
		tempStatus.Render(fmt.Sprintf("%d", currentTemp))))

	// Flight time
	maxFlightTime := d.SafetyConfig.Behavioral.MaxFlightTime
	currentTime := 0
	if d.CurrentState != nil {
		currentTime = d.CurrentState.Time
	}

	timeStatus := d.SafeStyle
	if currentTime > maxFlightTime-60 {
		timeStatus = d.WarningStyle
	} else if currentTime > maxFlightTime {
		timeStatus = d.CriticalStyle
	}

	builder.WriteString(fmt.Sprintf("Flight Time: %s/%d s\n",
		timeStatus.Render(fmt.Sprintf("%d", currentTime)),
		maxFlightTime))

	return builder.String()
}

// renderActiveEvents renders active safety events
func (d *Dashboard) renderActiveEvents() string {
	var builder strings.Builder

	builder.WriteString(d.ValueStyle.Bold(true).Render("Active Events"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	if len(d.ActiveEvents) == 0 {
		builder.WriteString(d.SafeStyle.Render("No active events"))
		return builder.String()
	}

	// Show up to 5 most recent events
	displayCount := min(5, len(d.ActiveEvents))
	for i := 0; i < displayCount; i++ {
		event := d.ActiveEvents[i]

		// Format timestamp
		timestamp := event.Timestamp.Format("15:04:05")

		// Get style based on event level
		var eventStyle lipgloss.Style
		switch event.Level {
		case "emergency":
			eventStyle = d.EmergencyStyle
		case "critical":
			eventStyle = d.CriticalStyle
		case "warning":
			eventStyle = d.WarningStyle
		default:
			eventStyle = d.ValueStyle
		}

		// Get icon based on event type
		icon := "ℹ️"
		switch event.Type {
		case "altitude":
			icon = "📈"
		case "battery":
			icon = "🔋"
		case "sensor":
			icon = "📡"
		case "behavioral":
			icon = "⚡"
		case "connection":
			icon = "📶"
		case "emergency":
			icon = "🚨"
		}

		builder.WriteString(fmt.Sprintf("%s [%s] %s: %s\n",
			icon,
			timestamp,
			eventStyle.Render(strings.ToUpper(event.Level)),
			eventStyle.Render(event.Message)))
	}

	if len(d.ActiveEvents) > displayCount {
		builder.WriteString(d.ValueStyle.Render(fmt.Sprintf("... and %d more events", len(d.ActiveEvents)-displayCount)))
	}

	return builder.String()
}

// renderSafetyDetails renders detailed safety information
func (d *Dashboard) renderSafetyDetails() string {
	var builder strings.Builder

	builder.WriteString(d.ValueStyle.Bold(true).Render("Safety Details"))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 30))
	builder.WriteString("\n")

	if d.SafetyConfig == nil {
		builder.WriteString(d.ValueStyle.Render("No configuration loaded"))
		return builder.String()
	}

	// Configuration level
	builder.WriteString(fmt.Sprintf("Config Level: %s\n", d.SafetyConfig.Level))

	// Auto-land settings
	autoLand := "Disabled"
	if d.SafetyConfig.Battery.EnableAutoLand {
		autoLand = "Enabled"
	}
	builder.WriteString(fmt.Sprintf("Auto-land: %s\n", autoLand))

	// Flip maneuvers
	flipsEnabled := "Disabled"
	if d.SafetyConfig.Behavioral.EnableFlips {
		flipsEnabled = "Enabled"
	}
	builder.WriteString(fmt.Sprintf("Flip maneuvers: %s\n", flipsEnabled))

	// Connection timeout
	builder.WriteString(fmt.Sprintf("Connection timeout: %d ms\n",
		d.SafetyConfig.Emergency.ConnectionTimeout))

	// Command rate limit
	builder.WriteString(fmt.Sprintf("Max command rate: %d/s\n",
		d.SafetyConfig.Behavioral.MaxCommandRate))

	// Sensor limits
	builder.WriteString(fmt.Sprintf("Max tilt: %d°\n",
		d.SafetyConfig.Sensors.MaxTiltAngle))
	builder.WriteString(fmt.Sprintf("Min TOF distance: %d cm\n",
		d.SafetyConfig.Sensors.MinTOFDistance))

	return builder.String()
}

// renderFooter renders the dashboard footer
func (d *Dashboard) renderFooter() string {
	controls := "F6: Toggle Details | F7: Toggle Limits | F8: Toggle Events | F9: Emergency Stop"

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Faint(true).
		Render(controls)
}

// ToggleDetails toggles detailed view
func (d *Dashboard) ToggleDetails() {
	d.ShowDetails = !d.ShowDetails
}

// ToggleLimits toggles limits display
func (d *Dashboard) ToggleLimits() {
	d.ShowLimits = !d.ShowLimits
}

// ToggleEvents toggles events display
func (d *Dashboard) ToggleEvents() {
	d.ShowEvents = !d.ShowEvents
}

// ClearEvents clears all events from history
func (d *Dashboard) ClearEvents() {
	d.ActiveEvents = make([]safety.SafetyEvent, 0)
	d.EventHistory = make([]safety.SafetyEvent, 0)
}

// SetStyle sets the main dashboard style
func (d *Dashboard) SetStyle(style lipgloss.Style) *Dashboard {
	d.Style = style
	return d
}

// Helper function for square root
func mathSqrt(x float64) float64 {
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
