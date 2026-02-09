package safety

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Alert represents a safety alert
type Alert struct {
	ID           string
	Level        string // "info", "warning", "critical", "emergency"
	Message      string
	Timestamp    time.Time
	Duration     time.Duration
	Acknowledged bool
	Source       string
}

// AlertSystem represents a safety alert notification system
type AlertSystem struct {
	Alerts         []Alert
	MaxAlerts      int
	ShowAlerts     bool
	AlertDuration  time.Duration
	Style          lipgloss.Style
	InfoStyle      lipgloss.Style
	WarningStyle   lipgloss.Style
	CriticalStyle  lipgloss.Style
	EmergencyStyle lipgloss.Style
}

// NewAlertSystem creates a new safety alert system
func NewAlertSystem() *AlertSystem {
	return &AlertSystem{
		Alerts:        make([]Alert, 0),
		MaxAlerts:     10,
		ShowAlerts:    true,
		AlertDuration: 10 * time.Second,
		Style:         lipgloss.NewStyle().Padding(0, 1),
		InfoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")), // Blue
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")), // Orange
		CriticalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // Red
		EmergencyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Blink(true), // Red blinking
	}
}

// AddAlert adds a new alert
func (a *AlertSystem) AddAlert(level, message, source string) string {
	alert := Alert{
		ID:           fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		Level:        level,
		Message:      message,
		Timestamp:    time.Now(),
		Duration:     a.AlertDuration,
		Acknowledged: false,
		Source:       source,
	}

	a.Alerts = append([]Alert{alert}, a.Alerts...)

	// Limit number of alerts
	if len(a.Alerts) > a.MaxAlerts {
		a.Alerts = a.Alerts[:a.MaxAlerts]
	}

	return alert.ID
}

// AcknowledgeAlert acknowledges an alert by ID
func (a *AlertSystem) AcknowledgeAlert(id string) bool {
	for i, alert := range a.Alerts {
		if alert.ID == id {
			a.Alerts[i].Acknowledged = true
			return true
		}
	}
	return false
}

// AcknowledgeAll acknowledges all alerts
func (a *AlertSystem) AcknowledgeAll() {
	for i := range a.Alerts {
		a.Alerts[i].Acknowledged = true
	}
}

// RemoveAlert removes an alert by ID
func (a *AlertSystem) RemoveAlert(id string) bool {
	for i, alert := range a.Alerts {
		if alert.ID == id {
			a.Alerts = append(a.Alerts[:i], a.Alerts[i+1:]...)
			return true
		}
	}
	return false
}

// ClearAlerts clears all alerts
func (a *AlertSystem) ClearAlerts() {
	a.Alerts = make([]Alert, 0)
}

// Update removes expired alerts
func (a *AlertSystem) Update() {
	now := time.Now()
	var validAlerts []Alert

	for _, alert := range a.Alerts {
		if alert.Timestamp.Add(alert.Duration).After(now) {
			validAlerts = append(validAlerts, alert)
		}
	}

	a.Alerts = validAlerts
}

// GetActiveAlerts returns unacknowledged alerts
func (a *AlertSystem) GetActiveAlerts() []Alert {
	var active []Alert
	for _, alert := range a.Alerts {
		if !alert.Acknowledged {
			active = append(active, alert)
		}
	}
	return active
}

// GetCriticalAlerts returns critical or emergency alerts
func (a *AlertSystem) GetCriticalAlerts() []Alert {
	var critical []Alert
	for _, alert := range a.Alerts {
		if alert.Level == "critical" || alert.Level == "emergency" {
			critical = append(critical, alert)
		}
	}
	return critical
}

// Render renders the alert system
func (a *AlertSystem) Render() string {
	if !a.ShowAlerts || len(a.Alerts) == 0 {
		return ""
	}

	var builder strings.Builder

	// Get active alerts
	activeAlerts := a.GetActiveAlerts()
	if len(activeAlerts) == 0 {
		return ""
	}

	// Show up to 3 most recent active alerts
	displayCount := alertsMinInt(3, len(activeAlerts))

	builder.WriteString("🚨 SAFETY ALERTS\n")
	builder.WriteString(strings.Repeat("─", 40))
	builder.WriteString("\n")

	for i := 0; i < displayCount; i++ {
		alert := activeAlerts[i]

		// Get style based on alert level
		var alertStyle lipgloss.Style
		var icon string

		switch alert.Level {
		case "emergency":
			alertStyle = a.EmergencyStyle
			icon = "🚨"
		case "critical":
			alertStyle = a.CriticalStyle
			icon = "❌"
		case "warning":
			alertStyle = a.WarningStyle
			icon = "⚠️"
		default:
			alertStyle = a.InfoStyle
			icon = "ℹ️"
		}

		// Format timestamp
		timestamp := alert.Timestamp.Format("15:04:05")

		// Calculate time since alert
		age := time.Since(alert.Timestamp)
		ageStr := formatDuration(age)

		// Build alert line
		builder.WriteString(fmt.Sprintf("%s [%s] %s: %s (%s ago)\n",
			icon,
			timestamp,
			alertStyle.Render(strings.ToUpper(alert.Level)),
			alert.Message,
			ageStr))

		// Add source if available
		if alert.Source != "" {
			builder.WriteString(fmt.Sprintf("  Source: %s\n", alert.Source))
		}
	}

	if len(activeAlerts) > displayCount {
		builder.WriteString(fmt.Sprintf("... and %d more alerts\n", len(activeAlerts)-displayCount))
	}

	builder.WriteString("\n")
	builder.WriteString("Press 'A' to acknowledge all alerts\n")

	return a.Style.Render(builder.String())
}

// RenderCompact renders a compact alert notification
func (a *AlertSystem) RenderCompact() string {
	if !a.ShowAlerts {
		return ""
	}

	activeAlerts := a.GetActiveAlerts()
	if len(activeAlerts) == 0 {
		return ""
	}

	criticalAlerts := a.GetCriticalAlerts()

	var builder strings.Builder

	if len(criticalAlerts) > 0 {
		// Show critical alerts in compact form
		builder.WriteString("🚨 ")
		builder.WriteString(a.EmergencyStyle.Render(fmt.Sprintf("%d CRITICAL", len(criticalAlerts))))
	} else if len(activeAlerts) > 0 {
		// Show warning count
		builder.WriteString("⚠️ ")
		builder.WriteString(a.WarningStyle.Render(fmt.Sprintf("%d ALERTS", len(activeAlerts))))
	}

	return builder.String()
}

// ToggleAlerts toggles alert display
func (a *AlertSystem) ToggleAlerts() {
	a.ShowAlerts = !a.ShowAlerts
}

// SetAlertDuration sets the default alert duration
func (a *AlertSystem) SetAlertDuration(duration time.Duration) {
	a.AlertDuration = duration
}

// SetMaxAlerts sets the maximum number of alerts to keep
func (a *AlertSystem) SetMaxAlerts(max int) {
	a.MaxAlerts = max
	if len(a.Alerts) > max {
		a.Alerts = a.Alerts[:max]
	}
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}

// Helper function for min (renamed to avoid conflict)
func alertsMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
