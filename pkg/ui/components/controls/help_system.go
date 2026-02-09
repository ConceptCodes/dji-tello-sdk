package controls

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DetailLevel represents the level of detail in help messages
type DetailLevel int

const (
	DetailLevelBasic DetailLevel = iota
	DetailLevelAdvanced
	DetailLevelExpert
)

// HelpContext represents the current context for help
type HelpContext struct {
	Component  string // Current component (telemetry, safety, layout, etc.)
	Mode       string // Current mode (flight, config, etc.)
	Focused    string // Currently focused element
	LastAction string // Last action performed
}

// HelpTopic represents a help topic with multiple detail levels
type HelpTopic struct {
	ID          string
	Title       string
	Description map[DetailLevel]string
	Shortcuts   map[string]string // key -> description
	Examples    []string
}

// HelpSystem provides context-sensitive help
type HelpSystem struct {
	width       int
	height      int
	context     HelpContext
	topics      map[string]HelpTopic
	showHelp    bool
	autoShow    bool
	detailLevel DetailLevel
	scrollPos   int
	maxLines    int
}

// NewHelpSystem creates a new help system
func NewHelpSystem(width, height int) *HelpSystem {
	hs := &HelpSystem{
		width:       width,
		height:      height,
		showHelp:    false,
		autoShow:    true,
		detailLevel: DetailLevelBasic,
		scrollPos:   0,
		maxLines:    height - 4, // Reserve space for header/footer
		topics:      make(map[string]HelpTopic),
	}

	// Initialize help topics
	hs.initializeTopics()

	return hs
}

// initializeTopics sets up all help topics
func (hs *HelpSystem) initializeTopics() {
	// General TUI help
	hs.topics["general"] = HelpTopic{
		ID:    "general",
		Title: "Tello Control TUI - General Help",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Basic controls for drone operation",
			DetailLevelAdvanced: "Complete TUI reference with all features",
			DetailLevelExpert:   "Advanced usage patterns and troubleshooting",
		},
		Shortcuts: map[string]string{
			"/":          "Enter command mode",
			"T":          "Take off",
			"L":          "Land",
			"Space":      "Emergency stop",
			"WASD":       "Movement controls",
			"Arrow Keys": "Altitude and rotation",
			"F1":         "Show this help",
			"F7":         "Toggle layout mode",
			"Tab":        "Cycle focus between panes",
			"Esc":        "Exit help/close dialogs",
		},
		Examples: []string{
			"Type '/takeoff' to take off",
			"Press 'T' for quick takeoff",
			"Use F1 for context-sensitive help",
		},
	}

	// Telemetry help
	hs.topics["telemetry"] = HelpTopic{
		ID:    "telemetry",
		Title: "Telemetry Dashboard",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Real-time drone telemetry display",
			DetailLevelAdvanced: "Enhanced telemetry with graphs and horizon",
			DetailLevelExpert:   "Advanced sensor data and visualization",
		},
		Shortcuts: map[string]string{
			"F2": "Toggle graphs display",
			"F3": "Toggle artificial horizon",
			"F4": "Toggle detailed metrics",
			"F5": "Switch between basic/enhanced telemetry",
		},
		Examples: []string{
			"Press F5 to switch telemetry modes",
			"Use F2-F4 to customize the display",
		},
	}

	// Safety help
	hs.topics["safety"] = HelpTopic{
		ID:    "safety",
		Title: "Safety Monitoring System",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Safety limits and warnings",
			DetailLevelAdvanced: "Emergency procedures and monitoring",
			DetailLevelExpert:   "Advanced safety configuration and diagnostics",
		},
		Shortcuts: map[string]string{
			"F6":  "Toggle safety dashboard",
			"F9":  "Manual emergency stop",
			"F10": "Toggle proximity warnings",
			"F11": "Toggle limit warnings",
			"F12": "Toggle emergency procedures",
		},
		Examples: []string{
			"Press F6 to show/hide safety dashboard",
			"Use F9 for manual emergency stop",
		},
	}

	// Layout help
	hs.topics["layout"] = HelpTopic{
		ID:    "layout",
		Title: "Multi-pane Layout System",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Basic layout controls",
			DetailLevelAdvanced: "Advanced layout management",
			DetailLevelExpert:   "Custom layouts and configuration",
		},
		Shortcuts: map[string]string{
			"F7":        "Toggle layout mode",
			"Ctrl+H":    "Horizontal layout",
			"Ctrl+V":    "Vertical layout",
			"Ctrl+G":    "Grid layout",
			"Ctrl+T":    "Tabbed layout",
			"Tab":       "Cycle focus between panes",
			"Shift+Tab": "Cycle focus in reverse",
			"Ctrl+Q":    "Close focused pane",
			"Ctrl+F":    "Toggle pane fullscreen",
		},
		Examples: []string{
			"Press F7 to enable multi-pane layout",
			"Use Ctrl+H/V/G/T to switch layouts",
			"Tab to navigate between panes",
		},
	}

	// Flight controls help
	hs.topics["flight"] = HelpTopic{
		ID:    "flight",
		Title: "Flight Controls",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Basic flight controls",
			DetailLevelAdvanced: "Advanced flight patterns and maneuvers",
			DetailLevelExpert:   "Precision flight and automation",
		},
		Shortcuts: map[string]string{
			"W/S":        "Forward/Backward",
			"A/D":        "Left/Right",
			"Up/Down":    "Ascend/Descend",
			"Left/Right": "Rotate counter-clockwise/clockwise",
			"T":          "Take off",
			"L":          "Land",
			"Space":      "Emergency stop",
			"F":          "Flip (requires direction: l/r/f/b)",
			"R":          "Return to home",
		},
		Examples: []string{
			"WASD + Arrow keys for movement",
			"Type '/flip l' to flip left",
			"Space for emergency stop",
		},
	}

	// Command mode help
	hs.topics["commands"] = HelpTopic{
		ID:    "commands",
		Title: "Command Mode Reference",
		Description: map[DetailLevel]string{
			DetailLevelBasic:    "Basic drone commands",
			DetailLevelAdvanced: "Advanced command syntax",
			DetailLevelExpert:   "Scripting and automation",
		},
		Shortcuts: map[string]string{
			"/":       "Enter command mode",
			"Tab":     "Command autocomplete",
			"Up/Down": "Command history",
			"Ctrl+R":  "Search command history",
			"Ctrl+S":  "Save command as macro",
			"Ctrl+P":  "Play macro",
		},
		Examples: []string{
			"takeoff - Take off",
			"land - Land",
			"emergency - Emergency stop",
			"flip l - Flip left",
			"flip r - Flip right",
			"flip f - Flip forward",
			"flip b - Flip backward",
			"streamon - Start video stream",
			"streamoff - Stop video stream",
			"speed 50 - Set speed to 50 cm/s",
			"up 100 - Ascend 100 cm",
			"down 100 - Descend 100 cm",
			"forward 100 - Move forward 100 cm",
			"back 100 - Move backward 100 cm",
			"left 100 - Move left 100 cm",
			"right 100 - Move right 100 cm",
			"cw 90 - Rotate clockwise 90 degrees",
			"ccw 90 - Rotate counter-clockwise 90 degrees",
		},
	}
}

// UpdateContext updates the help context
func (hs *HelpSystem) UpdateContext(context HelpContext) {
	hs.context = context
	if hs.autoShow && hs.shouldAutoShowHelp() {
		hs.showHelp = true
	}
}

// shouldAutoShowHelp determines if help should auto-show
func (hs *HelpSystem) shouldAutoShowHelp() bool {
	// Auto-show help for certain contexts
	switch hs.context.Component {
	case "unknown", "error", "config":
		return true
	}
	return false
}

// ToggleHelp toggles the help display
func (hs *HelpSystem) ToggleHelp() {
	hs.showHelp = !hs.showHelp
	hs.scrollPos = 0
}

// SetDetailLevel sets the help detail level
func (hs *HelpSystem) SetDetailLevel(level DetailLevel) {
	hs.detailLevel = level
}

// CycleDetailLevel cycles through detail levels
func (hs *HelpSystem) CycleDetailLevel() {
	hs.detailLevel = (hs.detailLevel + 1) % 3
}

// ScrollUp scrolls the help content up
func (hs *HelpSystem) ScrollUp() {
	if hs.scrollPos > 0 {
		hs.scrollPos--
	}
}

// ScrollDown scrolls the help content down
func (hs *HelpSystem) ScrollDown(lines int) {
	// We'll adjust based on content height
	hs.scrollPos++
}

// GetCurrentTopic returns the most relevant help topic
func (hs *HelpSystem) GetCurrentTopic() HelpTopic {
	// Determine topic based on context
	topicID := hs.context.Component
	if topicID == "" {
		topicID = "general"
	}

	topic, exists := hs.topics[topicID]
	if !exists {
		return hs.topics["general"]
	}
	return topic
}

// Render renders the help system
func (hs *HelpSystem) Render() string {
	if !hs.showHelp {
		return ""
	}

	topic := hs.GetCurrentTopic()

	// Styles
	helpStyle := lipgloss.NewStyle().
		Width(hs.width).
		Height(hs.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true).
		PaddingTop(1).
		PaddingBottom(1)

	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Width(15)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	exampleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("76")).
		Italic(true)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		PaddingTop(1)

	// Build help content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render(topic.Title))
	content.WriteString("\n\n")

	// Description
	desc := topic.Description[hs.detailLevel]
	content.WriteString(descStyle.Render(desc))
	content.WriteString("\n\n")

	// Shortcuts
	content.WriteString(sectionStyle.Render("Keyboard Shortcuts:"))
	content.WriteString("\n")

	for key, description := range topic.Shortcuts {
		content.WriteString(fmt.Sprintf("  %s: %s\n",
			shortcutStyle.Render(key),
			descStyle.Render(description)))
	}
	content.WriteString("\n")

	// Examples
	if len(topic.Examples) > 0 {
		content.WriteString(sectionStyle.Render("Examples:"))
		content.WriteString("\n")
		for _, example := range topic.Examples {
			content.WriteString(fmt.Sprintf("  • %s\n", exampleStyle.Render(example)))
		}
		content.WriteString("\n")
	}

	// Footer
	footer := fmt.Sprintf("Detail level: %s | Press F1 to close | ↑/↓ to scroll | F2 to change detail",
		map[DetailLevel]string{
			DetailLevelBasic:    "Basic",
			DetailLevelAdvanced: "Advanced",
			DetailLevelExpert:   "Expert",
		}[hs.detailLevel])

	content.WriteString(footerStyle.Render(footer))

	return helpStyle.Render(content.String())
}

// HandleKey handles keyboard input for the help system
func (hs *HelpSystem) HandleKey(key string) bool {
	switch key {
	case "f1":
		hs.ToggleHelp()
		return true
	case "f2":
		if hs.showHelp {
			hs.CycleDetailLevel()
			return true
		}
	case "up":
		if hs.showHelp {
			hs.ScrollUp()
			return true
		}
	case "down":
		if hs.showHelp {
			hs.ScrollDown(1)
			return true
		}
	case "esc":
		if hs.showHelp {
			hs.showHelp = false
			return true
		}
	}

	return false
}

// IsVisible returns whether help is currently visible
func (hs *HelpSystem) IsVisible() bool {
	return hs.showHelp
}

// GetContext returns the current help context
func (hs *HelpSystem) GetContext() HelpContext {
	return hs.context
}

// SetAutoShow sets whether help should auto-show
func (hs *HelpSystem) SetAutoShow(autoShow bool) {
	hs.autoShow = autoShow
}

// AddTopic adds a custom help topic
func (hs *HelpSystem) AddTopic(topic HelpTopic) {
	hs.topics[topic.ID] = topic
}

// RemoveTopic removes a help topic
func (hs *HelpSystem) RemoveTopic(id string) {
	delete(hs.topics, id)
}
