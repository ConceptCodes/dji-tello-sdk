package controls

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Suggestion represents a command suggestion
type Suggestion struct {
	Command     string
	Description string
	Category    string
}

// Autocomplete provides command autocomplete functionality
type Autocomplete struct {
	width           int
	height          int
	suggestions     []Suggestion
	selectedIndex   int
	prefix          string
	showSuggestions bool
	maxSuggestions  int
	allCommands     []Suggestion
	filtered        []Suggestion
}

// NewAutocomplete creates a new autocomplete system
func NewAutocomplete(width, height int) *Autocomplete {
	ac := &Autocomplete{
		width:           width,
		height:          height,
		suggestions:     []Suggestion{},
		selectedIndex:   0,
		prefix:          "",
		showSuggestions: false,
		maxSuggestions:  5,
		allCommands:     []Suggestion{},
		filtered:        []Suggestion{},
	}

	// Initialize with all available commands
	ac.initializeCommands()

	return ac
}

// initializeCommands sets up all available commands
func (ac *Autocomplete) initializeCommands() {
	ac.allCommands = []Suggestion{
		// Basic commands
		{Command: "takeoff", Description: "Take off", Category: "flight"},
		{Command: "land", Description: "Land", Category: "flight"},
		{Command: "emergency", Description: "Emergency stop", Category: "safety"},
		{Command: "streamon", Description: "Start video stream", Category: "video"},
		{Command: "streamoff", Description: "Stop video stream", Category: "video"},

		// Movement commands
		{Command: "up", Description: "Ascend [distance]", Category: "movement"},
		{Command: "down", Description: "Descend [distance]", Category: "movement"},
		{Command: "forward", Description: "Move forward [distance]", Category: "movement"},
		{Command: "back", Description: "Move backward [distance]", Category: "movement"},
		{Command: "left", Description: "Move left [distance]", Category: "movement"},
		{Command: "right", Description: "Move right [distance]", Category: "movement"},
		{Command: "cw", Description: "Rotate clockwise [degrees]", Category: "movement"},
		{Command: "ccw", Description: "Rotate counter-clockwise [degrees]", Category: "movement"},

		// Flip commands
		{Command: "flip l", Description: "Flip left", Category: "acrobatics"},
		{Command: "flip r", Description: "Flip right", Category: "acrobatics"},
		{Command: "flip f", Description: "Flip forward", Category: "acrobatics"},
		{Command: "flip b", Description: "Flip backward", Category: "acrobatics"},

		// Speed commands
		{Command: "speed", Description: "Set speed [cm/s]", Category: "configuration"},
		{Command: "speed?", Description: "Get current speed", Category: "query"},

		// Query commands
		{Command: "battery?", Description: "Get battery level", Category: "query"},
		{Command: "time?", Description: "Get flight time", Category: "query"},
		{Command: "height?", Description: "Get height", Category: "query"},
		{Command: "temp?", Description: "Get temperature", Category: "query"},
		{Command: "attitude?", Description: "Get attitude (pitch/roll/yaw)", Category: "query"},
		{Command: "baro?", Description: "Get barometer reading", Category: "query"},
		{Command: "acceleration?", Description: "Get acceleration", Category: "query"},
		{Command: "tof?", Description: "Get time-of-flight distance", Category: "query"},

		// Configuration commands
		{Command: "wifi", Description: "Set WiFi credentials [ssid password]", Category: "configuration"},
		{Command: "mon", Description: "Enable mission pad detection", Category: "configuration"},
		{Command: "moff", Description: "Disable mission pad detection", Category: "configuration"},
		{Command: "mdirection", Description: "Set mission pad direction [0-2]", Category: "configuration"},

		// Advanced commands
		{Command: "go", Description: "Fly to coordinates [x y z speed]", Category: "advanced"},
		{Command: "curve", Description: "Fly curve [x1 y1 z1 x2 y2 z2 speed]", Category: "advanced"},
		{Command: "jump", Description: "Jump to mission pad [x y z speed yaw mid1 mid2]", Category: "advanced"},
	}
}

// UpdatePrefix updates the command prefix and filters suggestions
func (ac *Autocomplete) UpdatePrefix(prefix string) {
	ac.prefix = strings.TrimSpace(prefix)
	ac.showSuggestions = len(ac.prefix) > 0

	if ac.showSuggestions {
		ac.filterSuggestions()
		ac.selectedIndex = 0
	} else {
		ac.suggestions = []Suggestion{}
	}
}

// filterSuggestions filters commands based on prefix
func (ac *Autocomplete) filterSuggestions() {
	ac.filtered = []Suggestion{}

	for _, cmd := range ac.allCommands {
		if strings.HasPrefix(cmd.Command, ac.prefix) {
			ac.filtered = append(ac.filtered, cmd)
		}
	}

	// Limit to max suggestions
	if len(ac.filtered) > ac.maxSuggestions {
		ac.suggestions = ac.filtered[:ac.maxSuggestions]
	} else {
		ac.suggestions = ac.filtered
	}
}

// SelectNext selects the next suggestion
func (ac *Autocomplete) SelectNext() {
	if len(ac.suggestions) == 0 {
		return
	}
	ac.selectedIndex = (ac.selectedIndex + 1) % len(ac.suggestions)
}

// SelectPrevious selects the previous suggestion
func (ac *Autocomplete) SelectPrevious() {
	if len(ac.suggestions) == 0 {
		return
	}
	ac.selectedIndex = (ac.selectedIndex - 1 + len(ac.suggestions)) % len(ac.suggestions)
}

// GetSelectedSuggestion returns the currently selected suggestion
func (ac *Autocomplete) GetSelectedSuggestion() (Suggestion, bool) {
	if len(ac.suggestions) == 0 || ac.selectedIndex < 0 || ac.selectedIndex >= len(ac.suggestions) {
		return Suggestion{}, false
	}
	return ac.suggestions[ac.selectedIndex], true
}

// GetCompletion returns the completed command based on selected suggestion
func (ac *Autocomplete) GetCompletion() string {
	if suggestion, ok := ac.GetSelectedSuggestion(); ok {
		return suggestion.Command
	}
	return ac.prefix
}

// Hide hides the autocomplete suggestions
func (ac *Autocomplete) Hide() {
	ac.showSuggestions = false
	ac.suggestions = []Suggestion{}
	ac.selectedIndex = 0
}

// IsVisible returns whether suggestions are visible
func (ac *Autocomplete) IsVisible() bool {
	return ac.showSuggestions && len(ac.suggestions) > 0
}

// Render renders the autocomplete suggestions
func (ac *Autocomplete) Render() string {
	if !ac.IsVisible() {
		return ""
	}

	// Styles
	suggestionStyle := lipgloss.NewStyle().
		Width(ac.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("248"))

	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("76")).
		Width(12).
		Align(lipgloss.Right)

	// Build suggestions list
	var content strings.Builder

	for i, suggestion := range ac.suggestions {
		// Highlight selected suggestion
		cmdStyle := commandStyle
		if i == ac.selectedIndex {
			cmdStyle = selectedStyle.Copy().Width(20)
		}

		line := fmt.Sprintf("%s %s %s",
			cmdStyle.Render(suggestion.Command),
			descStyle.Render(suggestion.Description),
			categoryStyle.Render("["+suggestion.Category+"]"),
		)

		content.WriteString(line)
		if i < len(ac.suggestions)-1 {
			content.WriteString("\n")
		}
	}

	return suggestionStyle.Render(content.String())
}

// HandleKey handles keyboard input for autocomplete
func (ac *Autocomplete) HandleKey(key string) (handled bool, completion string) {
	if !ac.IsVisible() {
		return false, ""
	}

	switch key {
	case "tab":
		if len(ac.suggestions) > 0 {
			ac.SelectNext()
			return true, ac.GetCompletion()
		}
	case "shift+tab":
		if len(ac.suggestions) > 0 {
			ac.SelectPrevious()
			return true, ac.GetCompletion()
		}
	case "up":
		ac.SelectPrevious()
		return true, ac.GetCompletion()
	case "down":
		ac.SelectNext()
		return true, ac.GetCompletion()
	case "enter":
		if suggestion, ok := ac.GetSelectedSuggestion(); ok {
			ac.Hide()
			return true, suggestion.Command
		}
	case "esc":
		ac.Hide()
		return true, ac.prefix
	}

	return false, ""
}

// AddCommand adds a custom command to autocomplete
func (ac *Autocomplete) AddCommand(command, description, category string) {
	ac.allCommands = append(ac.allCommands, Suggestion{
		Command:     command,
		Description: description,
		Category:    category,
	})
}

// RemoveCommand removes a command from autocomplete
func (ac *Autocomplete) RemoveCommand(command string) {
	for i, cmd := range ac.allCommands {
		if cmd.Command == command {
			ac.allCommands = append(ac.allCommands[:i], ac.allCommands[i+1:]...)
			break
		}
	}
}

// SetMaxSuggestions sets the maximum number of suggestions to show
func (ac *Autocomplete) SetMaxSuggestions(max int) {
	ac.maxSuggestions = max
}

// GetSuggestionsCount returns the number of available suggestions
func (ac *Autocomplete) GetSuggestionsCount() int {
	return len(ac.suggestions)
}

// GetFilteredCount returns the number of filtered suggestions
func (ac *Autocomplete) GetFilteredCount() int {
	return len(ac.filtered)
}

// GetPrefix returns the current prefix
func (ac *Autocomplete) GetPrefix() string {
	return ac.prefix
}
