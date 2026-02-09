package ml

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/conceptcodes/dji-tello-sdk-go/pkg/ml"
)

// DetectionList represents a panel showing ML detections
type DetectionList struct {
	Width           int
	Height          int
	Detections      []ml.Detection
	Tracks          []ml.Track
	MaxItems        int
	SortBy          SortOption
	Filter          FilterOption
	ShowConfidence  bool
	ShowBoundingBox bool
	ShowTimestamp   bool
	LastUpdate      time.Time
	UpdateRate      time.Duration
	Style           lipgloss.Style
	HeaderStyle     lipgloss.Style
	ItemStyle       lipgloss.Style
	ConfidenceStyle lipgloss.Style
	WarningStyle    lipgloss.Style
	ErrorStyle      lipgloss.Style
}

// SortOption defines how to sort detections
type SortOption string

const (
	SortByConfidence SortOption = "confidence"
	SortByClass      SortOption = "class"
	SortByTimestamp  SortOption = "timestamp"
	SortBySize       SortOption = "size"
)

// FilterOption defines how to filter detections
type FilterOption string

const (
	FilterAll      FilterOption = "all"
	FilterHighConf FilterOption = "high_confidence"
	FilterTracked  FilterOption = "tracked"
	FilterByClass  FilterOption = "class"
)

// NewDetectionList creates a new detection list panel
func NewDetectionList(width, height int) *DetectionList {
	return &DetectionList{
		Width:           width,
		Height:          height,
		Detections:      make([]ml.Detection, 0),
		Tracks:          make([]ml.Track, 0),
		MaxItems:        height - 4, // Reserve space for header and footer
		SortBy:          SortByConfidence,
		Filter:          FilterAll,
		ShowConfidence:  true,
		ShowBoundingBox: true,
		ShowTimestamp:   true,
		LastUpdate:      time.Now(),
		UpdateRate:      100 * time.Millisecond,
		Style:           lipgloss.NewStyle().Padding(0, 1),
		HeaderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1),
		ItemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		ConfidenceStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
	}
}

// UpdateDetections updates the detection list with new detections
func (dl *DetectionList) UpdateDetections(detections []ml.Detection) {
	dl.Detections = detections
	dl.LastUpdate = time.Now()
}

// UpdateTracks updates the detection list with new tracks
func (dl *DetectionList) UpdateTracks(tracks []ml.Track) {
	dl.Tracks = tracks
	dl.LastUpdate = time.Now()
}

// SetSort sets the sort option
func (dl *DetectionList) SetSort(sortBy SortOption) {
	dl.SortBy = sortBy
}

// SetFilter sets the filter option
func (dl *DetectionList) SetFilter(filter FilterOption) {
	dl.Filter = filter
}

// ToggleConfidence toggles confidence display
func (dl *DetectionList) ToggleConfidence() {
	dl.ShowConfidence = !dl.ShowConfidence
}

// ToggleBoundingBox toggles bounding box display
func (dl *DetectionList) ToggleBoundingBox() {
	dl.ShowBoundingBox = !dl.ShowBoundingBox
}

// ToggleTimestamp toggles timestamp display
func (dl *DetectionList) ToggleTimestamp() {
	dl.ShowTimestamp = !dl.ShowTimestamp
}

// Render renders the detection list panel
func (dl *DetectionList) Render() string {
	if dl.Width <= 0 || dl.Height <= 0 {
		return ""
	}

	// Get filtered and sorted detections
	detections := dl.getFilteredDetections()
	detections = dl.sortDetections(detections)

	// Limit to max items
	if len(detections) > dl.MaxItems {
		detections = detections[:dl.MaxItems]
	}

	// Build header
	header := dl.renderHeader(len(detections))

	// Build detection list
	detectionList := dl.renderDetectionList(detections)

	// Build footer
	footer := dl.renderFooter()

	// Combine all parts
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		detectionList,
		footer,
	)

	return dl.Style.Width(dl.Width).Height(dl.Height).Render(content)
}

// getFilteredDetections returns filtered detections based on current filter
func (dl *DetectionList) getFilteredDetections() []ml.Detection {
	switch dl.Filter {
	case FilterAll:
		return dl.Detections
	case FilterHighConf:
		return dl.filterByConfidence(0.5)
	case FilterTracked:
		return dl.filterTrackedDetections()
	case FilterByClass:
		// For now, return all - could be extended to filter by specific class
		return dl.Detections
	default:
		return dl.Detections
	}
}

// filterByConfidence filters detections by minimum confidence
func (dl *DetectionList) filterByConfidence(minConfidence float32) []ml.Detection {
	filtered := make([]ml.Detection, 0)
	for _, det := range dl.Detections {
		if det.Confidence >= minConfidence {
			filtered = append(filtered, det)
		}
	}
	return filtered
}

// filterTrackedDetections filters detections that are being tracked
func (dl *DetectionList) filterTrackedDetections() []ml.Detection {
	if len(dl.Tracks) == 0 {
		return dl.Detections
	}

	// Create a map of tracked class IDs for quick lookup
	trackedClasses := make(map[int]bool)
	for _, track := range dl.Tracks {
		trackedClasses[track.ClassID] = true
	}

	filtered := make([]ml.Detection, 0)
	for _, det := range dl.Detections {
		if trackedClasses[det.ClassID] {
			filtered = append(filtered, det)
		}
	}
	return filtered
}

// sortDetections sorts detections based on current sort option
func (dl *DetectionList) sortDetections(detections []ml.Detection) []ml.Detection {
	switch dl.SortBy {
	case SortByConfidence:
		sort.Slice(detections, func(i, j int) bool {
			return detections[i].Confidence > detections[j].Confidence
		})
	case SortByClass:
		sort.Slice(detections, func(i, j int) bool {
			if detections[i].ClassName == detections[j].ClassName {
				return detections[i].Confidence > detections[j].Confidence
			}
			return detections[i].ClassName < detections[j].ClassName
		})
	case SortByTimestamp:
		sort.Slice(detections, func(i, j int) bool {
			return detections[i].Timestamp.After(detections[j].Timestamp)
		})
	case SortBySize:
		sort.Slice(detections, func(i, j int) bool {
			sizeI := detections[i].Box.Dx() * detections[i].Box.Dy()
			sizeJ := detections[j].Box.Dx() * detections[j].Box.Dy()
			if sizeI == sizeJ {
				return detections[i].Confidence > detections[j].Confidence
			}
			return sizeI > sizeJ
		})
	}
	return detections
}

// renderHeader renders the panel header
func (dl *DetectionList) renderHeader(count int) string {
	status := "🟢"
	if time.Since(dl.LastUpdate) > 2*time.Second {
		status = "🟡"
	}
	if time.Since(dl.LastUpdate) > 5*time.Second {
		status = "🔴"
	}

	title := fmt.Sprintf("🔍 DETECTIONS (%d)", count)
	statusLine := fmt.Sprintf("%s Updated: %s ago", status, formatDuration(time.Since(dl.LastUpdate)))

	return dl.HeaderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			title,
			statusLine,
		),
	)
}

// renderDetectionList renders the list of detections
func (dl *DetectionList) renderDetectionList(detections []ml.Detection) string {
	if len(detections) == 0 {
		return dl.ItemStyle.Render("No detections")
	}

	lines := make([]string, 0, len(detections))
	for i, det := range detections {
		line := dl.renderDetectionItem(i+1, det)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderDetectionItem renders a single detection item
func (dl *DetectionList) renderDetectionItem(index int, det ml.Detection) string {
	parts := make([]string, 0)

	// Index and class name
	classPart := fmt.Sprintf("%2d. %-15s", index, truncateString(det.ClassName, 15))
	parts = append(parts, classPart)

	// Confidence (if enabled)
	if dl.ShowConfidence {
		confStr := fmt.Sprintf("%3.0f%%", det.Confidence*100)
		confPart := dl.ConfidenceStyle.Render(confStr)
		parts = append(parts, confPart)
	}

	// Bounding box (if enabled)
	if dl.ShowBoundingBox {
		boxStr := fmt.Sprintf("[%dx%d@%d,%d]",
			det.Box.Dx(), det.Box.Dy(), det.Box.Min.X, det.Box.Min.Y)
		parts = append(parts, boxStr)
	}

	// Timestamp (if enabled)
	if dl.ShowTimestamp {
		timeStr := formatDuration(time.Since(det.Timestamp))
		parts = append(parts, timeStr+" ago")
	}

	// Add tracked indicator if this detection is being tracked
	if dl.isTracked(det) {
		parts = append(parts, "📌")
	}

	return dl.ItemStyle.Render(strings.Join(parts, " "))
}

// isTracked checks if a detection is being tracked
func (dl *DetectionList) isTracked(det ml.Detection) bool {
	for _, track := range dl.Tracks {
		if track.ClassID == det.ClassID &&
			track.Box.Overlaps(det.Box) &&
			track.Confidence >= det.Confidence*0.8 {
			return true
		}
	}
	return false
}

// renderFooter renders the panel footer with controls
func (dl *DetectionList) renderFooter() string {
	controls := []string{
		"[F8] Toggle Confidence",
		"[F9] Toggle BBox",
		"[F10] Toggle Time",
		"[F11] Cycle Sort",
		"[F12] Cycle Filter",
	}

	return dl.ItemStyle.Foreground(lipgloss.Color("240")).Render(
		strings.Join(controls, " | "),
	)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetStats returns statistics about current detections
func (dl *DetectionList) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if len(dl.Detections) == 0 {
		stats["count"] = 0
		stats["avg_confidence"] = 0.0
		stats["tracked_count"] = 0
		return stats
	}

	// Count by class
	classCount := make(map[string]int)
	totalConfidence := float32(0)
	trackedCount := 0

	for _, det := range dl.Detections {
		classCount[det.ClassName]++
		totalConfidence += det.Confidence
		if dl.isTracked(det) {
			trackedCount++
		}
	}

	stats["count"] = len(dl.Detections)
	stats["avg_confidence"] = totalConfidence / float32(len(dl.Detections))
	stats["tracked_count"] = trackedCount
	stats["class_distribution"] = classCount
	stats["last_update"] = dl.LastUpdate
	stats["age"] = time.Since(dl.LastUpdate).String()

	return stats
}
