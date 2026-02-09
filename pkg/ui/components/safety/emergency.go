package safety

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// EmergencyVisualization represents an emergency procedure visualization
type EmergencyVisualization struct {
	Width           int
	Height          int
	EmergencyType   string
	EmergencyLevel  string
	ProcedureSteps  []string
	CurrentStep     int
	StepProgress    float64
	TimeRemaining   time.Duration
	IsActive        bool
	ShowCountdown   bool
	ShowProcedure   bool
	Style           lipgloss.Style
	HeaderStyle     lipgloss.Style
	StepStyle       lipgloss.Style
	ActiveStepStyle lipgloss.Style
	CompletedStyle  lipgloss.Style
	CountdownStyle  lipgloss.Style
	WarningStyle    lipgloss.Style
	CriticalStyle   lipgloss.Style
	EmergencyStyle  lipgloss.Style
}

// NewEmergencyVisualization creates a new emergency procedure visualization
func NewEmergencyVisualization(width, height int) *EmergencyVisualization {
	return &EmergencyVisualization{
		Width:          width,
		Height:         height,
		EmergencyType:  "none",
		EmergencyLevel: "none",
		ProcedureSteps: []string{},
		CurrentStep:    0,
		StepProgress:   0.0,
		TimeRemaining:  0,
		IsActive:       false,
		ShowCountdown:  true,
		ShowProcedure:  true,
		Style:          lipgloss.NewStyle().Padding(0, 1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Blink(true),
		StepStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		ActiveStepStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		CompletedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		CountdownStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		CriticalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		EmergencyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Blink(true),
	}
}

// StartEmergency starts an emergency procedure visualization
func (e *EmergencyVisualization) StartEmergency(emergencyType, emergencyLevel string) {
	e.EmergencyType = emergencyType
	e.EmergencyLevel = emergencyLevel
	e.CurrentStep = 0
	e.StepProgress = 0.0
	e.IsActive = true

	// Set procedure steps based on emergency type
	e.ProcedureSteps = e.getProcedureSteps(emergencyType)

	// Set time remaining based on emergency level
	e.TimeRemaining = e.getTimeForLevel(emergencyLevel)
}

// StopEmergency stops the emergency procedure visualization
func (e *EmergencyVisualization) StopEmergency() {
	e.IsActive = false
	e.EmergencyType = "none"
	e.EmergencyLevel = "none"
	e.ProcedureSteps = []string{}
	e.CurrentStep = 0
	e.StepProgress = 0.0
	e.TimeRemaining = 0
}

// UpdateProgress updates the emergency procedure progress
func (e *EmergencyVisualization) UpdateProgress(elapsedTime time.Duration) {
	if !e.IsActive || len(e.ProcedureSteps) == 0 {
		return
	}

	// Update time remaining
	if e.TimeRemaining > 0 {
		e.TimeRemaining -= elapsedTime
		if e.TimeRemaining < 0 {
			e.TimeRemaining = 0
		}
	}

	// Calculate progress based on steps
	totalSteps := len(e.ProcedureSteps)
	if totalSteps > 0 {
		// Simple progress: move to next step every 2 seconds
		stepDuration := 2 * time.Second
		stepsCompleted := int(elapsedTime / stepDuration)

		if stepsCompleted > 0 {
			e.CurrentStep = emergencyMin(e.CurrentStep+stepsCompleted, totalSteps-1)
		}

		// Calculate progress within current step
		timeInStep := elapsedTime % stepDuration
		e.StepProgress = float64(timeInStep) / float64(stepDuration)

		// If we've completed all steps, mark as complete
		if e.CurrentStep >= totalSteps-1 && e.StepProgress >= 1.0 {
			e.IsActive = false
		}
	}
}

// getProcedureSteps returns the procedure steps for a given emergency type
func (e *EmergencyVisualization) getProcedureSteps(emergencyType string) []string {
	switch emergencyType {
	case "battery_low":
		return []string{
			"1. Reduce altitude to safe level",
			"2. Initiate controlled descent",
			"3. Prepare for auto-landing",
			"4. Execute emergency landing",
			"5. Power down safely",
		}
	case "connection_lost":
		return []string{
			"1. Attempt connection re-establishment",
			"2. Initiate hover stabilization",
			"3. Execute return-to-home procedure",
			"4. Prepare for auto-landing",
			"5. Land at home position",
		}
	case "sensor_failure":
		return []string{
			"1. Switch to backup sensors",
			"2. Reduce speed and altitude",
			"3. Initiate controlled descent",
			"4. Prepare for emergency landing",
			"5. Execute safe landing",
		}
	case "collision_imminent":
		return []string{
			"1. Execute immediate stop",
			"2. Assess obstacle position",
			"3. Calculate avoidance maneuver",
			"4. Execute avoidance",
			"5. Resume normal operation",
		}
	case "altitude_limit":
		return []string{
			"1. Reduce vertical velocity",
			"2. Stabilize at safe altitude",
			"3. Initiate controlled descent",
			"4. Return to safe altitude range",
			"5. Resume normal operation",
		}
	case "velocity_limit":
		return []string{
			"1. Reduce throttle input",
			"2. Gradually decrease speed",
			"3. Stabilize at safe velocity",
			"4. Maintain safe speed limits",
			"5. Resume normal operation",
		}
	default:
		return []string{
			"1. Assess emergency situation",
			"2. Stabilize drone position",
			"3. Execute safety procedure",
			"4. Prepare for landing",
			"5. Execute safe landing",
		}
	}
}

// getTimeForLevel returns the time allocation for a given emergency level
func (e *EmergencyVisualization) getTimeForLevel(level string) time.Duration {
	switch level {
	case "emergency":
		return 10 * time.Second
	case "critical":
		return 15 * time.Second
	case "warning":
		return 30 * time.Second
	default:
		return 20 * time.Second
	}
}

// Render renders the emergency procedure visualization
func (e *EmergencyVisualization) Render() string {
	if !e.IsActive {
		return e.renderInactive()
	}

	var builder strings.Builder

	// Add emergency header
	builder.WriteString(e.renderEmergencyHeader())
	builder.WriteString("\n\n")

	// Add countdown if enabled
	if e.ShowCountdown && e.TimeRemaining > 0 {
		builder.WriteString(e.renderCountdown())
		builder.WriteString("\n\n")
	}

	// Add procedure steps if enabled
	if e.ShowProcedure && len(e.ProcedureSteps) > 0 {
		builder.WriteString(e.renderProcedureSteps())
		builder.WriteString("\n\n")
	}

	// Add progress visualization
	builder.WriteString(e.renderProgress())
	builder.WriteString("\n")

	// Add emergency instructions
	builder.WriteString(e.renderInstructions())

	return e.Style.Render(builder.String())
}

// renderInactive renders the inactive state
func (e *EmergencyVisualization) renderInactive() string {
	var builder strings.Builder

	builder.WriteString(e.HeaderStyle.Render("🛡️ EMERGENCY PROCEDURES"))
	builder.WriteString("\n\n")
	builder.WriteString(e.StepStyle.Render("No active emergencies"))
	builder.WriteString("\n\n")
	builder.WriteString(e.StepStyle.Render("System status: NORMAL"))
	builder.WriteString("\n")
	builder.WriteString(e.StepStyle.Render("Ready to handle emergencies"))

	return e.Style.Render(builder.String())
}

// renderEmergencyHeader renders the emergency header
func (e *EmergencyVisualization) renderEmergencyHeader() string {
	// Get appropriate icon and title
	icon := "🚨"
	title := "EMERGENCY PROCEDURE ACTIVE"

	switch e.EmergencyLevel {
	case "warning":
		icon = "⚠️"
		title = "WARNING PROCEDURE ACTIVE"
	case "critical":
		icon = "❌"
		title = "CRITICAL PROCEDURE ACTIVE"
	}

	// Get style based on level
	var style lipgloss.Style
	switch e.EmergencyLevel {
	case "warning":
		style = e.WarningStyle
	case "critical":
		style = e.CriticalStyle
	default:
		style = e.EmergencyStyle
	}

	header := fmt.Sprintf("%s %s - %s", icon, style.Render(title), strings.ToUpper(e.EmergencyType))
	return e.HeaderStyle.Render(header)
}

// renderCountdown renders the emergency countdown
func (e *EmergencyVisualization) renderCountdown() string {
	var builder strings.Builder

	seconds := int(e.TimeRemaining.Seconds())
	minutes := seconds / 60
	remainingSeconds := seconds % 60

	timeStr := fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)

	// Color code based on time remaining
	var timeStyle lipgloss.Style
	if e.TimeRemaining <= 5*time.Second {
		timeStyle = e.EmergencyStyle
	} else if e.TimeRemaining <= 15*time.Second {
		timeStyle = e.CriticalStyle
	} else {
		timeStyle = e.WarningStyle
	}

	builder.WriteString("⏰ TIME REMAINING: ")
	builder.WriteString(timeStyle.Render(timeStr))
	builder.WriteString("\n")

	// Add progress bar
	progressWidth := 30
	progress := 1.0 - float64(e.TimeRemaining)/float64(e.getTimeForLevel(e.EmergencyLevel))
	filled := int(float64(progressWidth) * progress)
	empty := progressWidth - filled

	progressBar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
	builder.WriteString(progressBar)
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("%.0f%%", progress*100))

	return builder.String()
}

// renderProcedureSteps renders the emergency procedure steps
func (e *EmergencyVisualization) renderProcedureSteps() string {
	var builder strings.Builder

	builder.WriteString("📋 EMERGENCY PROCEDURE STEPS\n")
	builder.WriteString(strings.Repeat("─", 40))
	builder.WriteString("\n")

	for i, step := range e.ProcedureSteps {
		var stepStyle lipgloss.Style

		if i < e.CurrentStep {
			// Completed step
			stepStyle = e.CompletedStyle
			builder.WriteString("✓ ")
		} else if i == e.CurrentStep {
			// Current step
			stepStyle = e.ActiveStepStyle

			// Add progress indicator for current step
			progressWidth := 10
			filled := int(float64(progressWidth) * e.StepProgress)
			progressBar := "[" + strings.Repeat("▶", filled) + strings.Repeat(" ", progressWidth-filled) + "]"

			builder.WriteString(fmt.Sprintf("▶ %s ", progressBar))
		} else {
			// Future step
			stepStyle = e.StepStyle
			builder.WriteString("○ ")
		}

		builder.WriteString(stepStyle.Render(step))
		builder.WriteString("\n")
	}

	return builder.String()
}

// renderProgress renders the overall progress visualization
func (e *EmergencyVisualization) renderProgress() string {
	var builder strings.Builder

	totalSteps := len(e.ProcedureSteps)
	if totalSteps == 0 {
		return ""
	}

	progress := float64(e.CurrentStep) / float64(totalSteps-1)
	if e.CurrentStep >= totalSteps-1 {
		progress = 1.0
	}

	// Overall progress bar
	progressWidth := 40
	filled := int(float64(progressWidth) * progress)
	empty := progressWidth - filled

	progressBar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"

	builder.WriteString("📊 OVERALL PROGRESS: ")
	builder.WriteString(progressBar)
	builder.WriteString(fmt.Sprintf(" %.0f%%", progress*100))
	builder.WriteString("\n")

	// Step indicator
	builder.WriteString(fmt.Sprintf("Step %d/%d", e.CurrentStep+1, totalSteps))

	return builder.String()
}

// renderInstructions renders emergency instructions
func (e *EmergencyVisualization) renderInstructions() string {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString("💡 EMERGENCY INSTRUCTIONS:\n")
	builder.WriteString(strings.Repeat("─", 40))
	builder.WriteString("\n")

	instructions := e.getEmergencyInstructions(e.EmergencyType)
	for _, instruction := range instructions {
		builder.WriteString("• ")
		builder.WriteString(instruction)
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.WriteString("🚨 DO NOT INTERRUPT EMERGENCY PROCEDURE")

	return builder.String()
}

// getEmergencyInstructions returns instructions for a given emergency type
func (e *EmergencyVisualization) getEmergencyInstructions(emergencyType string) []string {
	switch emergencyType {
	case "battery_low":
		return []string{
			"Maintain clear landing area",
			"Do not attempt manual override",
			"Monitor landing progress",
			"Prepare for immediate battery replacement",
		}
	case "connection_lost":
		return []string{
			"Check controller connection",
			"Clear line of sight to drone",
			"Monitor return-to-home progress",
			"Prepare for manual recovery if needed",
		}
	case "sensor_failure":
		return []string{
			"Avoid sudden movements",
			"Monitor descent carefully",
			"Prepare for rough landing",
			"Inspect sensors after landing",
		}
	case "collision_imminent":
		return []string{
			"Clear area of obstacles",
			"Monitor avoidance maneuver",
			"Be prepared for emergency stop",
			"Resume operation cautiously",
		}
	default:
		return []string{
			"Monitor procedure progress",
			"Maintain safe distance",
			"Be prepared for manual intervention",
			"Follow safety protocols",
		}
	}
}

// ToggleCountdown toggles countdown display
func (e *EmergencyVisualization) ToggleCountdown() {
	e.ShowCountdown = !e.ShowCountdown
}

// ToggleProcedure toggles procedure display
func (e *EmergencyVisualization) ToggleProcedure() {
	e.ShowProcedure = !e.ShowProcedure
}

// SetStyle sets the main style
func (e *EmergencyVisualization) SetStyle(style lipgloss.Style) *EmergencyVisualization {
	e.Style = style
	return e
}

// Helper function for min (renamed to avoid conflict)
func emergencyMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
