package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/tello"
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

type tickMsg time.Time
type connectionMsg struct {
	connected bool
	err       error
}

type GamepadMsg struct {
	Message string
}

type TuiModel struct {
	commander  tello.TelloCommander
	textInput  textinput.Model
	viewport   viewport.Model
	logs       []string
	width      int
	height     int
	connected  bool
	battery    int
	heightCm   int
	speed      int
	temp       int
	flightTime int
	inputMode  bool // true = typing command, false = flight control
	err        error
}

func NewTuiModel(commander tello.TelloCommander) TuiModel {
	ti := textinput.New()
	ti.Placeholder = "Type command (e.g. 'takeoff', 'flip l')..."
	ti.CharLimit = 156
	ti.Width = 50

	vp := viewport.New(80, 15)
	vp.SetContent("Waiting for logs...")

	return TuiModel{
		commander: commander,
		textInput: ti,
		viewport:  vp,
		logs:      []string{},
		connected: false,
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
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 40   // Reserve space for telemetry
		m.viewport.Height = msg.Height - 10 // Reserve space for header/footer

	case tickMsg:
		// Poll telemetry
		m.updateTelemetry()
		return m, m.tickCmd()
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

	// Telemetry Panel
	telemetryContent := lipgloss.JoinVertical(lipgloss.Left,
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
	)

	telemetryPanel := stylePanel.
		Width(30).
		Height(m.height - 8).
		Render(telemetryContent)

	// Logs Panel
	logPanel := stylePanel.
		Width(m.width - 36).
		Height(m.height - 8).
		Render(m.viewport.View())

	// Main Content
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, telemetryPanel, logPanel)

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
	// This runs in the update loop, so we should be careful not to block
	// Ideally, we'd use commands to fetch data asynchronously
	// For simplicity in this prototype, we'll do quick checks
	// In a real app, these should be tea.Cmds

	// Note: TelloCommander methods are blocking/synchronous usually
	// We should spawn goroutines and send msgs back
	// But for now, let's assume we can read cached state if available
	// or just do it (might cause slight UI lag if connection is bad)

	// Better approach: The commander should push updates or we poll in a goroutine
	// Let's just simulate "connected" if we get a response

	// We can't easily call methods on m.commander here without blocking Update
	// So we'll dispatch a command to fetch data
}

func (m TuiModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
