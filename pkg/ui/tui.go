package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/types"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ui/components/controls"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ui/components/layout"
	mlui "github.com/conceptcodes/dji-tello-sdk-go/pkg/ui/components/ml"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ui/components/safety"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ui/components/telemetry"
)

// Styles
var (
	// Colors
	colorBrand   = lipgloss.Color("205") // Pink/Magenta
	colorSuccess = lipgloss.Color("42")  // Green
	colorWarning = lipgloss.Color("214") // Orange
	colorError   = lipgloss.Color("196") // Red
	colorDim     = lipgloss.Color("240") // Grey
	colorText    = lipgloss.Color("252") // White-ish
	colorBg      = lipgloss.Color("234") // Dark Grey
	colorPanel   = lipgloss.Color("236") // Slightly lighter grey

	// Layout
	styleApp = lipgloss.NewStyle().Padding(1, 2)

	// Header
	styleHeader = lipgloss.NewStyle().
			Foreground(colorText).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorDim).
			Padding(0, 1).
			Bold(true)

	styleBrand = lipgloss.NewStyle().
			Foreground(colorBrand).
			Bold(true).
			MarginRight(2)

	styleStatusConnected = lipgloss.NewStyle().
				Foreground(colorSuccess).
				SetString("● CONNECTED")

	styleStatusDisconnected = lipgloss.NewStyle().
				Foreground(colorError).
				SetString("○ DISCONNECTED")

	// Panels
	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			MarginRight(1)

	styleTitle = lipgloss.NewStyle().
			Foreground(colorBrand).
			Bold(true).
			Padding(0, 1).
			Background(colorBg).
			MarginBottom(1)

	// Telemetry
	styleLabel = lipgloss.NewStyle().
			Foreground(colorDim).
			Width(12)

	styleValue = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	// Logs
	styleLogInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue
	styleLogWarn  = lipgloss.NewStyle().Foreground(colorWarning)
	styleLogError = lipgloss.NewStyle().Foreground(colorError)
	styleLogTime  = lipgloss.NewStyle().Foreground(colorDim).MarginRight(1)

	// Input
	styleInputPrompt = lipgloss.NewStyle().
				Foreground(colorBrand).
				MarginRight(1)
)

type (
	tickMsg       time.Time
	connectionMsg struct {
		connected bool
		err       error
	}
)

type GamepadMsg struct {
	Message string
}

type stateUpdateMsg struct {
	state *types.State
}

type TuiModel struct {
	commander             tello.TelloCommander
	textInput             textinput.Model
	viewport              viewport.Model
	dashboard             *telemetry.Dashboard
	safetyDashboard       *safety.Manager
	mlTrackingVisualizer  *mlui.TrackingVisualizer
	mlMetricsDashboard    *mlui.MetricsDashboard
	layoutManager         *layout.LayoutManager
	helpSystem            *controls.HelpSystem
	logs                  []string
	width                 int
	height                int
	connected             bool
	battery               int
	heightCm              int
	speed                 int
	temp                  int
	flightTime            int
	inputMode             bool // true = typing command, false = flight control
	showEnhancedTelemetry bool // true = enhanced dashboard, false = basic telemetry
	showSafetyDashboard   bool // true = safety dashboard visible
	showMLTracking        bool // true = ML tracking visible
	showMLMetrics         bool // true = ML metrics visible
	layoutMode            bool // true = using layout system, false = legacy layout
	err                   error
}

func NewTuiModel(commander tello.TelloCommander) TuiModel {
	ti := textinput.New()
	ti.Placeholder = "Type command (e.g. 'takeoff', 'flip l')..."
	ti.CharLimit = 156
	ti.Width = 50

	vp := viewport.New(80, 15)
	vp.SetContent("Waiting for logs...")

	// Initialize with default size, will be updated on first render
	dashboard := telemetry.NewDashboard(80, 20)
	safetyDashboard := safety.NewManager(40, 20)
	mlTrackingVisualizer := mlui.NewTrackingVisualizer(40, 20)
	mlMetricsDashboard := mlui.NewMetricsDashboard(40, 20)

	// Initialize layout manager with default size
	layoutManager := layout.NewLayoutManager(80, 24)

	// Initialize help system
	helpSystem := controls.NewHelpSystem(80, 24)

	return TuiModel{
		commander:             commander,
		textInput:             ti,
		viewport:              vp,
		dashboard:             dashboard,
		safetyDashboard:       safetyDashboard,
		mlTrackingVisualizer:  mlTrackingVisualizer,
		mlMetricsDashboard:    mlMetricsDashboard,
		layoutManager:         layoutManager,
		helpSystem:            helpSystem,
		logs:                  []string{},
		connected:             false,
		showEnhancedTelemetry: true,  // Start with enhanced telemetry enabled
		showSafetyDashboard:   false, // Safety dashboard starts hidden
		showMLTracking:        false, // ML tracking starts hidden
		showMLMetrics:         false, // ML metrics starts hidden
		layoutMode:            false, // Start with legacy layout for backward compatibility
	}
}

func (m TuiModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.tickCmd(),
		m.connectCmd(),
	)
}

func (m TuiModel) connectCmd() tea.Cmd {
	return func() tea.Msg {
		// Try to initialize, but don't fail hard if it times out
		// We want the UI to start regardless
		err := m.commander.Init()
		return connectionMsg{connected: err == nil, err: err}
	}
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case connectionMsg:
		m.connected = msg.connected
		if msg.connected {
			m.logs = append(m.logs, m.formatLog("INFO", "Connected to drone", styleLogInfo))
		} else {
			m.logs = append(m.logs, m.formatLog("WARN", fmt.Sprintf("Connection failed: %v. Retrying...", msg.err), styleLogWarn))
			// Retry connection after 5 seconds
			return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return m.connectCmd()()
			})
		}

	case GamepadMsg:
		m.logs = append(m.logs, m.formatLog("GAMEPAD", msg.Message, styleLogInfo))

	case tea.KeyMsg:
		// Global keys
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Input Mode
		if m.inputMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.inputMode = false
				m.textInput.Blur()
				return m, nil
			case tea.KeyEnter:
				cmdText := m.textInput.Value()
				if cmdText != "" {
					m.logs = append(m.logs, m.formatLog("CMD", cmdText, styleLogInfo))
					m.textInput.SetValue("")
					// Execute command asynchronously
					go m.executeCommand(cmdText)
				}
			}
			m.textInput, tiCmd = m.textInput.Update(msg)
			return m, tiCmd
		}

		// Flight Mode
		switch msg.String() {
		case "/":
			m.inputMode = true
			m.textInput.Focus()
			return m, textinput.Blink
		case "t":
			m.logs = append(m.logs, m.formatLog("INFO", "Taking off...", styleLogInfo))
			go m.commander.TakeOff()
		case "l":
			m.logs = append(m.logs, m.formatLog("INFO", "Landing...", styleLogInfo))
			go m.commander.Land()
		case "w":
			go m.commander.Forward(20)
		case "s":
			go m.commander.Backward(20)
		case "a":
			go m.commander.Left(20)
		case "d":
			go m.commander.Right(20)
		case "up":
			go m.commander.Up(20)
		case "down":
			go m.commander.Down(20)
		case "left":
			go m.commander.CounterClockwise(15)
		case "right":
			go m.commander.Clockwise(15)
		case " ": // Spacebar
			m.logs = append(m.logs, m.formatLog("WARN", "EMERGENCY STOP", styleLogError))
			go m.commander.Emergency()
		case "f1":
			if m.helpSystem != nil {
				// Update help context based on current state
				context := controls.HelpContext{
					Component:  m.getCurrentComponent(),
					Mode:       m.getCurrentMode(),
					Focused:    m.getFocusedElement(),
					LastAction: m.getLastAction(),
				}
				m.helpSystem.UpdateContext(context)
				m.helpSystem.HandleKey("f1")
				m.logs = append(m.logs, m.formatLog("INFO",
					fmt.Sprintf("Help %s", map[bool]string{true: "shown", false: "hidden"}[m.helpSystem.IsVisible()]),
					styleLogInfo))
			}
		case "f2":
			if m.dashboard != nil {
				m.dashboard.ToggleGraphs()
				m.logs = append(m.logs, m.formatLog("INFO", "Toggled graphs display", styleLogInfo))
			}
		case "f3":
			if m.dashboard != nil {
				m.dashboard.ToggleHorizon()
				m.logs = append(m.logs, m.formatLog("INFO", "Toggled horizon display", styleLogInfo))
			}
		case "f4":
			if m.dashboard != nil {
				m.dashboard.ToggleMetrics()
				m.logs = append(m.logs, m.formatLog("INFO", "Toggled metrics display", styleLogInfo))
			}
		case "f5":
			m.showEnhancedTelemetry = !m.showEnhancedTelemetry
			m.logs = append(m.logs, m.formatLog("INFO",
				fmt.Sprintf("Switched to %s telemetry", map[bool]string{true: "enhanced", false: "basic"}[m.showEnhancedTelemetry]),
				styleLogInfo))
		case "f6":
			m.showSafetyDashboard = !m.showSafetyDashboard
			m.logs = append(m.logs, m.formatLog("INFO",
				fmt.Sprintf("Safety dashboard %s", map[bool]string{true: "shown", false: "hidden"}[m.showSafetyDashboard]),
				styleLogInfo))
		case "f7":
			m.layoutMode = !m.layoutMode
			m.logs = append(m.logs, m.formatLog("INFO",
				fmt.Sprintf("Layout mode %s", map[bool]string{true: "enabled", false: "disabled"}[m.layoutMode]),
				styleLogInfo))
			if m.layoutMode && m.layoutManager != nil {
				// Initialize layout with panes
				m.initializeLayout()
			}
		case "f8":
			m.showMLTracking = !m.showMLTracking
			m.logs = append(m.logs, m.formatLog("INFO",
				fmt.Sprintf("ML tracking %s", map[bool]string{true: "shown", false: "hidden"}[m.showMLTracking]),
				styleLogInfo))
		case "f9":
			m.showMLMetrics = !m.showMLMetrics
			m.logs = append(m.logs, m.formatLog("INFO",
				fmt.Sprintf("ML metrics %s", map[bool]string{true: "shown", false: "hidden"}[m.showMLMetrics]),
				styleLogInfo))
		case "ctrl+h":
			if m.layoutMode && m.layoutManager != nil {
				m.layoutManager.SetLayoutType(layout.LayoutHorizontal)
				m.logs = append(m.logs, m.formatLog("INFO", "Switched to horizontal layout", styleLogInfo))
			}
		case "ctrl+v":
			if m.layoutMode && m.layoutManager != nil {
				m.layoutManager.SetLayoutType(layout.LayoutVertical)
				m.logs = append(m.logs, m.formatLog("INFO", "Switched to vertical layout", styleLogInfo))
			}
		case "ctrl+g":
			if m.layoutMode && m.layoutManager != nil {
				m.layoutManager.SetLayoutType(layout.LayoutGrid)
				m.logs = append(m.logs, m.formatLog("INFO", "Switched to grid layout", styleLogInfo))
			}
		case "ctrl+t":
			if m.layoutMode && m.layoutManager != nil {
				m.layoutManager.SetLayoutType(layout.LayoutTabbed)
				m.logs = append(m.logs, m.formatLog("INFO", "Switched to tabbed layout", styleLogInfo))
			}
		case "tab":
			if m.layoutMode && m.layoutManager != nil {
				// Handle tab through layout manager
				cmd := m.layoutManager.HandleKey(msg)
				m.logs = append(m.logs, m.formatLog("INFO", "Cycled pane focus", styleLogInfo))
				return m, cmd
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 40   // Reserve space for telemetry
		m.viewport.Height = msg.Height - 10 // Reserve space for header/footer

		// Update layout manager if in layout mode
		if m.layoutMode && m.layoutManager != nil {
			m.layoutManager.Update(msg.Width, msg.Height)
		}

	case tickMsg:
		// Poll telemetry
		m.updateTelemetry()
		return m, m.tickCmd()

	case stateUpdateMsg:
		if msg.state != nil && m.dashboard != nil {
			m.dashboard.UpdateState(msg.state)
			m.updateTelemetry() // Update basic telemetry values
		}
		return m, nil
	}

	// Update viewport content
	m.viewport.SetContent(strings.Join(m.logs, "\n"))
	m.viewport.GotoBottom()
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m TuiModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Header
	status := styleStatusDisconnected
	if m.connected {
		status = styleStatusConnected
	}

	header := styleHeader.Width(m.width - 4).Render(
		lipgloss.JoinHorizontal(lipgloss.Center,
			styleBrand.Render("TELLO CONTROL"),
			status.Render(),
			lipgloss.NewStyle().MarginLeft(2).Render(fmt.Sprintf("🔋 %d%%", m.battery)),
		),
	)

	// Main Content
	var mainContent string
	if m.layoutMode && m.layoutManager != nil {
		// Use layout system
		// Update pane contents before rendering
		m.updateLayoutPanes()
		mainContent = m.layoutManager.Render()
	} else {
		// Legacy layout (two-pane)
		// Telemetry Panel
		var telemetryPanel string
		if m.showEnhancedTelemetry && m.dashboard != nil {
			// Update dashboard size based on current window
			dashboardWidth := 30
			dashboardHeight := m.height - 8

			// Create a new dashboard with updated size if needed
			if m.dashboard.Width != dashboardWidth || m.dashboard.Height != dashboardHeight {
				m.dashboard = telemetry.NewDashboard(dashboardWidth, dashboardHeight)
			}

			telemetryPanel = stylePanel.
				Width(dashboardWidth).
				Height(dashboardHeight).
				Render(m.dashboard.Render())
		} else {
			// Basic telemetry
			var telemetryContent string
			if m.dashboard != nil {
				telemetryContent = lipgloss.JoinVertical(lipgloss.Left,
					m.renderStat("Height", fmt.Sprintf("%d cm", m.heightCm)),
					m.renderStat("Speed", fmt.Sprintf("%d cm/s", m.speed)),
					m.renderStat("Temp", fmt.Sprintf("%d °C", m.temp)),
					m.renderStat("Time", fmt.Sprintf("%ds", m.flightTime)),
					"",
					styleTitle.Render("CONTROLS"),
					m.renderStat("[W/S]", "Fwd/Back"),
					m.renderStat("[A/D]", "Left/Right"),
					m.renderStat("[↑/↓]", "Up/Down"),
					m.renderStat("[←/→]", "Rotate"),
					m.renderStat("[T/L]", "Takeoff/Land"),
					m.renderStat("[Space]", "Emergency"),
					m.renderStat("[/]", "Command Mode"),
					m.renderStat("[F5]", "Toggle Telemetry"),
				)

				if m.showEnhancedTelemetry {
					telemetryContent = lipgloss.JoinVertical(lipgloss.Left,
						telemetryContent,
						"",
						styleTitle.Render("ENHANCED CONTROLS"),
						m.renderStat("[F2]", "Toggle Graphs"),
						m.renderStat("[F3]", "Toggle Horizon"),
						m.renderStat("[F4]", "Toggle Metrics"),
					)
				}
			} else {
				// Fallback if dashboard is nil
				telemetryContent = lipgloss.JoinVertical(lipgloss.Left,
					m.renderStat("Height", fmt.Sprintf("%d cm", m.heightCm)),
					m.renderStat("Speed", fmt.Sprintf("%d cm/s", m.speed)),
					m.renderStat("Temp", fmt.Sprintf("%d °C", m.temp)),
					m.renderStat("Time", fmt.Sprintf("%ds", m.flightTime)),
					"",
					styleTitle.Render("CONTROLS"),
					m.renderStat("[W/S]", "Fwd/Back"),
					m.renderStat("[A/D]", "Left/Right"),
					m.renderStat("[↑/↓]", "Up/Down"),
					m.renderStat("[←/→]", "Rotate"),
					m.renderStat("[T/L]", "Takeoff/Land"),
					m.renderStat("[Space]", "Emergency"),
					m.renderStat("[/]", "Command Mode"),
					m.renderStat("[F5]", "Toggle Telemetry"),
				)

				if m.showEnhancedTelemetry {
					telemetryContent = lipgloss.JoinVertical(lipgloss.Left,
						telemetryContent,
						"",
						styleTitle.Render("ENHANCED CONTROLS"),
						m.renderStat("[F2]", "Toggle Graphs"),
						m.renderStat("[F3]", "Toggle Horizon"),
						m.renderStat("[F4]", "Toggle Metrics"),
					)
				}
			}

			telemetryPanel = stylePanel.
				Width(30).
				Height(m.height - 8).
				Render(telemetryContent)
		}

		// Logs Panel
		logPanelWidth := m.width - 36
		if m.showMLTracking || m.showMLMetrics {
			logPanelWidth = m.width - 76 // Make room for ML panels
		}

		logPanel := stylePanel.
			Width(logPanelWidth).
			Height(m.height - 8).
			Render(m.viewport.View())

		// Build panels list
		panels := []string{telemetryPanel, logPanel}

		// Add ML tracking panel if enabled
		if m.showMLTracking && m.mlTrackingVisualizer != nil {
			mlTrackingPanel := stylePanel.
				Width(35).
				Height(m.height - 8).
				Render(m.mlTrackingVisualizer.Render())
			panels = append(panels, mlTrackingPanel)
		}

		// Add ML metrics panel if enabled
		if m.showMLMetrics && m.mlMetricsDashboard != nil {
			mlMetricsPanel := stylePanel.
				Width(35).
				Height(m.height - 8).
				Render(m.mlMetricsDashboard.Render())
			panels = append(panels, mlMetricsPanel)
		}

		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, panels...)
	}

	// Footer (Input)
	inputView := ""
	if m.inputMode {
		inputView = styleInputPrompt.Render(">") + m.textInput.View()
	} else {
		inputView = styleDim.Render("Press '/' to type command...")
	}

	footer := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 4).
		Render(inputView)

	return styleApp.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			mainContent,
			footer,
		),
	)
}

// Helpers

var styleDim = lipgloss.NewStyle().Foreground(colorDim)

func (m TuiModel) renderStat(label, value string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		styleLabel.Render(label),
		styleValue.Render(value),
	)
}

func (m TuiModel) formatLog(level, msg string, style lipgloss.Style) string {
	timestamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("%s %s %s",
		styleLogTime.Render(timestamp),
		style.Render(fmt.Sprintf("[%s]", level)),
		msg,
	)
}

func (m *TuiModel) updateTelemetry() {
	// Update basic telemetry values from dashboard state if available
	if m.dashboard != nil && m.dashboard.State != nil {
		state := m.dashboard.State
		m.battery = state.Bat
		m.heightCm = state.H

		// Calculate speed from velocity components
		speed := int(mathSqrt(float64(state.Vgx*state.Vgx + state.Vgy*state.Vgy + state.Vgz*state.Vgz)))
		m.speed = speed

		m.temp = state.Temph
		m.flightTime = state.Time
	}
}

// updateMLState updates ML visualization state from ML results
func (m *TuiModel) updateMLState(result ml.MLResult) {
	if m.mlTrackingVisualizer != nil {
		// Get current state
		state := m.mlTrackingVisualizer.GetState()

		// Update from ML result
		state.UpdateFromMLResult(result)

		// Update visualizer
		m.mlTrackingVisualizer.UpdateState(state)
	}

	if m.mlMetricsDashboard != nil {
		// Get current state
		state := m.mlMetricsDashboard.GetState()

		// Update from ML result
		state.UpdateFromMLResult(result)

		// Update dashboard
		m.mlMetricsDashboard.UpdateState(state)
	}
}

// Helper function for square root (copied from dashboard.go)
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

func (m TuiModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// initializeLayout sets up the layout manager with default panes
func (m *TuiModel) initializeLayout() {
	if m.layoutManager == nil {
		return
	}

	// Clear any existing panes
	m.layoutManager = layout.NewLayoutManager(m.width, m.height)

	// Add panes based on what's available
	m.layoutManager.AddPane("telemetry", "📊 Telemetry", 30, 20)
	m.layoutManager.AddPane("logs", "📝 Logs", 40, 20)

	if m.safetyDashboard != nil {
		m.layoutManager.AddPane("safety", "⚠️ Safety", 35, 20)
	}

	if m.mlTrackingVisualizer != nil {
		m.layoutManager.AddPane("ml_tracking", "🎯 ML Tracking", 35, 20)
	}

	if m.mlMetricsDashboard != nil {
		m.layoutManager.AddPane("ml_metrics", "📊 ML Metrics", 35, 20)
	}

	// Set initial focus
	m.layoutManager.FocusPane("telemetry")

	// Update layout with current window size
	m.layoutManager.Update(m.width, m.height)
}

// updateLayoutPanes updates the content of each pane in the layout
func (m *TuiModel) updateLayoutPanes() {
	if m.layoutManager == nil {
		return
	}

	// Update telemetry pane
	var telemetryContent string
	if m.showEnhancedTelemetry && m.dashboard != nil {
		telemetryContent = m.dashboard.Render()
	} else {
		telemetryContent = lipgloss.JoinVertical(lipgloss.Left,
			m.renderStat("Height", fmt.Sprintf("%d cm", m.heightCm)),
			m.renderStat("Speed", fmt.Sprintf("%d cm/s", m.speed)),
			m.renderStat("Temp", fmt.Sprintf("%d °C", m.temp)),
			m.renderStat("Time", fmt.Sprintf("%ds", m.flightTime)),
		)
	}
	m.layoutManager.SetPaneContent("telemetry", telemetryContent)

	// Update logs pane
	m.layoutManager.SetPaneContent("logs", m.viewport.View())

	// Update safety pane if available
	if m.safetyDashboard != nil {
		safetyContent := m.safetyDashboard.Render()
		m.layoutManager.SetPaneContent("safety", safetyContent)
	}

	// Update ML tracking pane if available
	if m.mlTrackingVisualizer != nil {
		mlTrackingContent := m.mlTrackingVisualizer.Render()
		m.layoutManager.SetPaneContent("ml_tracking", mlTrackingContent)
	}

	// Update ML metrics pane if available
	if m.mlMetricsDashboard != nil {
		mlMetricsContent := m.mlMetricsDashboard.Render()
		m.layoutManager.SetPaneContent("ml_metrics", mlMetricsContent)
	}
}

// Helper methods for help context
func (m *TuiModel) getCurrentComponent() string {
	if m.layoutMode && m.layoutManager != nil {
		focusedPane, found := m.layoutManager.GetFocusedPane()
		if found {
			return focusedPane.ID
		}
	}
	return "general"
}

func (m *TuiModel) getCurrentMode() string {
	if m.inputMode {
		return "command"
	}
	return "flight"
}

func (m *TuiModel) getFocusedElement() string {
	if m.layoutMode && m.layoutManager != nil {
		focusedPane, found := m.layoutManager.GetFocusedPane()
		if found {
			return focusedPane.Title
		}
	}
	return "none"
}

func (m *TuiModel) getLastAction() string {
	// This would track the last action performed
	// For now, return based on current state
	if m.connected {
		return "connected"
	}
	return "disconnected"
}

func (m *TuiModel) executeCommand(cmd string) {
	// Simple parser
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])

	// Basic validation and execution
	switch command {
	case "takeoff":
		m.commander.TakeOff()
	case "land":
		m.commander.Land()
	case "emergency":
		m.commander.Emergency()
	case "streamon":
		m.commander.StreamOn()
	case "streamoff":
		m.commander.StreamOff()
	case "up", "down", "left", "right", "forward", "back":
		if len(parts) < 2 {
			m.logs = append(m.logs, m.formatLog("ERROR", fmt.Sprintf("Usage: %s <distance>", command), styleLogError))
			return
		}
		// TODO: Parse distance and call appropriate method
		// For now, just log that we received it
		m.logs = append(m.logs, m.formatLog("INFO", fmt.Sprintf("Command '%s' not fully implemented in TUI yet", command), styleLogWarn))
	case "flip":
		if len(parts) < 2 {
			m.logs = append(m.logs, m.formatLog("ERROR", "Usage: flip <l/r/f/b>", styleLogError))
			return
		}
		direction := parts[1]
		if direction != "l" && direction != "r" && direction != "f" && direction != "b" {
			m.logs = append(m.logs, m.formatLog("ERROR", "Invalid direction. Use l, r, f, or b", styleLogError))
			return
		}
		m.commander.Flip(tello.FlipDirection(direction))
	default:
		m.logs = append(m.logs, m.formatLog("ERROR", fmt.Sprintf("Unknown command: %s", command), styleLogError))
	}
}
