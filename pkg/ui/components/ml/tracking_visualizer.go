package ml

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TrackingVisualizer displays object tracking information
type TrackingVisualizer struct {
	state  MLState
	width  int
	height int
	styles TrackingStyles
}

// TrackingStyles defines styles for the tracking visualizer
type TrackingStyles struct {
	Container  lipgloss.Style
	Title      lipgloss.Style
	Track      lipgloss.Style
	TrackID    lipgloss.Style
	ClassName  lipgloss.Style
	Confidence lipgloss.Style
	State      lipgloss.Style
	Stats      lipgloss.Style
	Separator  lipgloss.Style
	Confirmed  lipgloss.Style
	Tentative  lipgloss.Style
	Deleted    lipgloss.Style
}

// NewTrackingVisualizer creates a new tracking visualizer
func NewTrackingVisualizer(width, height int) *TrackingVisualizer {
	styles := createTrackingStyles()

	return &TrackingVisualizer{
		state:  NewDefaultMLState(),
		width:  width,
		height: height,
		styles: styles,
	}
}

// createTrackingStyles creates the styles for tracking visualization
func createTrackingStyles() TrackingStyles {
	// Colors
	colorBrand := lipgloss.Color("205")   // Pink/Magenta
	colorSuccess := lipgloss.Color("42")  // Green
	colorWarning := lipgloss.Color("214") // Orange
	colorError := lipgloss.Color("196")   // Red
	colorDim := lipgloss.Color("240")     // Grey
	colorText := lipgloss.Color("252")    // White-ish
	colorBg := lipgloss.Color("234")      // Dark Grey

	return TrackingStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			Background(colorBg),

		Title: lipgloss.NewStyle().
			Foreground(colorBrand).
			Bold(true).
			Padding(0, 1).
			Background(colorBg).
			MarginBottom(1),

		Track: lipgloss.NewStyle().
			Foreground(colorText).
			MarginBottom(1),

		TrackID: lipgloss.NewStyle().
			Foreground(colorBrand).
			Bold(true),

		ClassName: lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true),

		Confidence: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")), // Blue

		State: lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true),

		Stats: lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1),

		Separator: lipgloss.NewStyle().
			Foreground(colorDim).
			SetString(" | "),

		Confirmed: lipgloss.NewStyle().
			Foreground(colorSuccess).
			SetString("✓"),

		Tentative: lipgloss.NewStyle().
			Foreground(colorWarning).
			SetString("?"),

		Deleted: lipgloss.NewStyle().
			Foreground(colorError).
			SetString("✗"),
	}
}

// UpdateState updates the visualizer state
func (tv *TrackingVisualizer) UpdateState(state MLState) {
	tv.state = state
}

// UpdateSize updates the visualizer size
func (tv *TrackingVisualizer) UpdateSize(width, height int) {
	tv.width = width
	tv.height = height
}

// Render renders the tracking visualizer
func (tv *TrackingVisualizer) Render() string {
	if !tv.state.Active {
		return tv.renderInactive()
	}

	// Calculate available height for tracks
	headerHeight := 3 // Title + stats + separator
	footerHeight := 1 // Last update
	availableHeight := tv.height - headerHeight - footerHeight
	maxTracks := availableHeight / 2 // Each track takes ~2 lines

	// Build content
	var content strings.Builder

	// Title
	title := tv.styles.Title.Render("🎯 Object Tracking")
	content.WriteString(title)
	content.WriteString("\n")

	// Stats
	stats := tv.renderStats()
	content.WriteString(stats)
	content.WriteString("\n")

	// Separator
	separator := tv.styles.Separator.Render(strings.Repeat("─", tv.width-4))
	content.WriteString(separator)
	content.WriteString("\n")

	// Tracks
	tracks := tv.renderTracks(maxTracks)
	content.WriteString(tracks)

	// Last update
	lastUpdate := tv.renderLastUpdate()
	content.WriteString(lastUpdate)

	// Apply container style
	return tv.styles.Container.
		Width(tv.width).
		Height(tv.height).
		Render(content.String())
}

// renderInactive renders the inactive state
func (tv *TrackingVisualizer) renderInactive() string {
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Render("ML tracking inactive")

	return tv.styles.Container.
		Width(tv.width).
		Height(tv.height).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				tv.styles.Title.Render("🎯 Object Tracking"),
				"",
				message,
			),
		)
}

// renderStats renders tracking statistics
func (tv *TrackingVisualizer) renderStats() string {
	totalTracks := tv.state.GetTrackCount()
	confirmedTracks := tv.state.GetConfirmedTrackCount()
	tentativeTracks := tv.state.GetTentativeTrackCount()

	stats := []string{
		fmt.Sprintf("Total: %d", totalTracks),
		fmt.Sprintf("Confirmed: %d", confirmedTracks),
		fmt.Sprintf("Tentative: %d", tentativeTracks),
	}

	return tv.styles.Stats.Render(strings.Join(stats, " • "))
}

// renderTracks renders the track list
func (tv *TrackingVisualizer) renderTracks(maxTracks int) string {
	if len(tv.state.Tracks) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Render("No active tracks")
	}

	// Sort tracks by ID for consistent display
	tracks := tv.state.Tracks
	if len(tracks) > maxTracks {
		tracks = tracks[:maxTracks]
	}

	var trackLines []string
	for _, track := range tracks {
		trackLines = append(trackLines, tv.renderTrack(track))
	}

	return strings.Join(trackLines, "\n")
}

// renderTrack renders a single track
func (tv *TrackingVisualizer) renderTrack(track TrackVisualization) string {
	// State indicator
	var stateIndicator string
	switch track.State {
	case "confirmed":
		stateIndicator = tv.styles.Confirmed.Render()
	case "tentative":
		stateIndicator = tv.styles.Tentative.Render()
	case "deleted":
		stateIndicator = tv.styles.Deleted.Render()
	default:
		stateIndicator = " "
	}

	// Track info
	trackID := tv.styles.TrackID.Render(fmt.Sprintf("ID:%d", track.ID))
	className := tv.styles.ClassName.Render(track.ClassName)
	confidence := tv.styles.Confidence.Render(fmt.Sprintf("%.1f%%", track.Confidence*100))

	// Additional info
	age := fmt.Sprintf("age:%d", track.Age)
	hits := fmt.Sprintf("hits:%d", track.Hits)
	misses := fmt.Sprintf("misses:%d", track.Misses)

	// Box info (simplified)
	boxInfo := fmt.Sprintf("[%d,%d %dx%d]",
		track.Box.Min.X, track.Box.Min.Y,
		track.Box.Dx(), track.Box.Dy())

	// Build track line
	parts := []string{
		stateIndicator,
		trackID,
		className,
		confidence,
		age,
		hits,
		misses,
		boxInfo,
	}

	return tv.styles.Track.Render(strings.Join(parts, " "))
}

// renderLastUpdate renders the last update time
func (tv *TrackingVisualizer) renderLastUpdate() string {
	if tv.state.LastUpdate.IsZero() {
		return ""
	}

	elapsed := time.Since(tv.state.LastUpdate)
	elapsedStr := "just now"
	if elapsed > time.Second {
		elapsedStr = fmt.Sprintf("%.0fs ago", elapsed.Seconds())
	}

	return tv.styles.Stats.
		MarginTop(1).
		Render(fmt.Sprintf("Updated %s", elapsedStr))
}

// GetState returns the current ML state
func (tv *TrackingVisualizer) GetState() MLState {
	return tv.state
}

// GetWidth returns the visualizer width
func (tv *TrackingVisualizer) GetWidth() int {
	return tv.width
}

// GetHeight returns the visualizer height
func (tv *TrackingVisualizer) GetHeight() int {
	return tv.height
}

// HandleKey handles keyboard input
func (tv *TrackingVisualizer) HandleKey(key string) {
	// Handle key bindings for tracking visualizer
	switch key {
	case "t":
		// Toggle track display
		tv.state.Config.ShowTracks = !tv.state.Config.ShowTracks
	case "c":
		// Toggle confidence display
		tv.state.Config.ShowConfidence = !tv.state.Config.ShowConfidence
	case "v":
		// Toggle velocity display
		tv.state.Config.ShowVelocity = !tv.state.Config.ShowVelocity
	case "p":
		// Toggle predictions display
		tv.state.Config.ShowPredictions = !tv.state.Config.ShowPredictions
	}
}

// GetHelp returns help information for the visualizer
func (tv *TrackingVisualizer) GetHelp() []string {
	return []string{
		"[T] Toggle tracks",
		"[C] Toggle confidence",
		"[V] Toggle velocity",
		"[P] Toggle predictions",
		"[↑/↓] Scroll tracks",
	}
}
