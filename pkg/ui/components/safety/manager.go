package safety

import (
	"time"

	"github.com/conceptcodes/dji-tello-sdk-go/pkg/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
)

// Manager represents a comprehensive safety monitoring manager
type Manager struct {
	Dashboard          *Dashboard
	EmergencyViz       *EmergencyVisualization
	ProximityWarning   *ProximityWarning
	Width              int
	Height             int
	CurrentMode        string // "dashboard", "emergency", "proximity", "combined"
	ShowSafety         bool
	LastSafetyUpdate   time.Time
	SafetyUpdateRate   time.Duration
	ActiveEmergency    bool
	EmergencyStartTime time.Time
}

// NewManager creates a new safety monitoring manager
func NewManager(width, height int) *Manager {
	return &Manager{
		Dashboard:          NewDashboard(width, height),
		EmergencyViz:       NewEmergencyVisualization(width, height),
		ProximityWarning:   NewProximityWarning(width, height),
		Width:              width,
		Height:             height,
		CurrentMode:        "dashboard",
		ShowSafety:         true,
		LastSafetyUpdate:   time.Now(),
		SafetyUpdateRate:   100 * time.Millisecond,
		ActiveEmergency:    false,
		EmergencyStartTime: time.Time{},
	}
}

// UpdateSafety updates all safety components with new data
func (m *Manager) UpdateSafety(safetyStatus *safety.SafetyStatus, state *types.State) {
	m.LastSafetyUpdate = time.Now()

	// Update dashboard
	m.Dashboard.UpdateSafetyStatus(safetyStatus, state)

	// Update proximity warning
	m.ProximityWarning.UpdateState(state)

	// Check for emergencies and update emergency visualization
	m.updateEmergencyVisualization(safetyStatus)
}

// updateEmergencyVisualization updates the emergency visualization based on safety status
func (m *Manager) updateEmergencyVisualization(safetyStatus *safety.SafetyStatus) {
	if safetyStatus == nil {
		return
	}

	// Check if we have an emergency
	if safetyStatus.EmergencyMode && !m.ActiveEmergency {
		// Start emergency visualization
		m.ActiveEmergency = true
		m.EmergencyStartTime = time.Now()
		m.EmergencyViz.StartEmergency("emergency", "emergency")
		m.CurrentMode = "emergency"
	} else if !safetyStatus.EmergencyMode && m.ActiveEmergency {
		// Stop emergency visualization
		m.ActiveEmergency = false
		m.EmergencyViz.StopEmergency()
		m.CurrentMode = "dashboard"
	}

	// Update emergency progress if active
	if m.ActiveEmergency {
		elapsed := time.Since(m.EmergencyStartTime)
		m.EmergencyViz.UpdateProgress(elapsed)
	}

	// Check for critical events that might trigger warnings
	if len(safetyStatus.ActiveEvents) > 0 {
		// Look for critical events
		for _, event := range safetyStatus.ActiveEvents {
			if event.Level == "critical" || event.Level == "emergency" {
				// Switch to emergency mode for critical events
				if !m.ActiveEmergency {
					m.ActiveEmergency = true
					m.EmergencyStartTime = time.Now()
					m.EmergencyViz.StartEmergency(event.Type, event.Level)
					m.CurrentMode = "emergency"
				}
				break
			}
		}
	}
}

// AddObstacle adds an obstacle to the proximity warning system
func (m *Manager) AddObstacle(obstacle Obstacle) {
	m.ProximityWarning.AddObstacle(obstacle)
}

// StartEmergency manually starts an emergency procedure
func (m *Manager) StartEmergency(emergencyType, emergencyLevel string) {
	m.ActiveEmergency = true
	m.EmergencyStartTime = time.Now()
	m.EmergencyViz.StartEmergency(emergencyType, emergencyLevel)
	m.CurrentMode = "emergency"
}

// StopEmergency manually stops the emergency procedure
func (m *Manager) StopEmergency() {
	m.ActiveEmergency = false
	m.EmergencyViz.StopEmergency()
	m.CurrentMode = "dashboard"
}

// ToggleMode toggles between safety display modes
func (m *Manager) ToggleMode() {
	modes := []string{"dashboard", "emergency", "proximity", "combined"}
	currentIndex := -1

	for i, mode := range modes {
		if mode == m.CurrentMode {
			currentIndex = i
			break
		}
	}

	if currentIndex >= 0 {
		nextIndex := (currentIndex + 1) % len(modes)
		m.CurrentMode = modes[nextIndex]
	} else {
		m.CurrentMode = "dashboard"
	}
}

// ToggleSafety toggles safety display on/off
func (m *Manager) ToggleSafety() {
	m.ShowSafety = !m.ShowSafety
}

// Render renders the appropriate safety view based on current mode
func (m *Manager) Render() string {
	if !m.ShowSafety {
		return ""
	}

	switch m.CurrentMode {
	case "dashboard":
		return m.Dashboard.Render()
	case "emergency":
		if m.ActiveEmergency {
			return m.EmergencyViz.Render()
		}
		return m.Dashboard.Render()
	case "proximity":
		return m.ProximityWarning.Render()
	case "combined":
		// For combined view, we need a more sophisticated layout
		return m.renderCombinedView()
	default:
		return m.Dashboard.Render()
	}
}

// renderCombinedView renders a combined safety view
func (m *Manager) renderCombinedView() string {
	// This would be a more sophisticated layout combining all views
	// For now, return dashboard view
	return m.Dashboard.Render()
}

// HandleKey handles keyboard input for safety controls
func (m *Manager) HandleKey(key string) {
	switch key {
	case "f6":
		m.Dashboard.ToggleDetails()
	case "f7":
		m.Dashboard.ToggleLimits()
	case "f8":
		m.Dashboard.ToggleEvents()
	case "f9":
		// Emergency stop - would trigger actual emergency procedure
		m.StartEmergency("manual_emergency", "emergency")
	case "f10":
		m.ProximityWarning.ToggleRadar()
	case "f11":
		m.ProximityWarning.ToggleWarnings()
	case "f12":
		m.ProximityWarning.ToggleDistances()
	case "f13":
		m.ToggleMode()
	case "f14":
		m.ToggleSafety()
	}
}

// GetCurrentMode returns the current display mode
func (m *Manager) GetCurrentMode() string {
	return m.CurrentMode
}

// IsEmergencyActive returns whether an emergency is active
func (m *Manager) IsEmergencyActive() bool {
	return m.ActiveEmergency
}

// ClearAll clears all safety data
func (m *Manager) ClearAll() {
	m.Dashboard.ClearEvents()
	m.ProximityWarning.ClearObstacles()
	if m.ActiveEmergency {
		m.StopEmergency()
	}
}
