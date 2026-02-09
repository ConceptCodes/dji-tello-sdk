package ml

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MetricsDashboard displays ML pipeline performance metrics
type MetricsDashboard struct {
	state  MLState
	width  int
	height int
	styles MetricsStyles
}

// MetricsStyles defines styles for the metrics dashboard
type MetricsStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Metric    lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
	Good      lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Separator lipgloss.Style
	Graph     lipgloss.Style
	Bar       lipgloss.Style
}

// NewMetricsDashboard creates a new metrics dashboard
func NewMetricsDashboard(width, height int) *MetricsDashboard {
	styles := createMetricsStyles()

	return &MetricsDashboard{
		state:  NewDefaultMLState(),
		width:  width,
		height: height,
		styles: styles,
	}
}

// createMetricsStyles creates the styles for metrics dashboard
func createMetricsStyles() MetricsStyles {
	// Colors
	colorBrand := lipgloss.Color("205")   // Pink/Magenta
	colorSuccess := lipgloss.Color("42")  // Green
	colorWarning := lipgloss.Color("214") // Orange
	colorError := lipgloss.Color("196")   // Red
	colorDim := lipgloss.Color("240")     // Grey
	colorText := lipgloss.Color("252")    // White-ish
	colorBg := lipgloss.Color("234")      // Dark Grey

	return MetricsStyles{
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

		Metric: lipgloss.NewStyle().
			Foreground(colorText).
			MarginBottom(1),

		Label: lipgloss.NewStyle().
			Foreground(colorDim).
			Width(20),

		Value: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true),

		Good: lipgloss.NewStyle().
			Foreground(colorSuccess),

		Warning: lipgloss.NewStyle().
			Foreground(colorWarning),

		Error: lipgloss.NewStyle().
			Foreground(colorError),

		Separator: lipgloss.NewStyle().
			Foreground(colorDim).
			SetString(" | "),

		Graph: lipgloss.NewStyle().
			Foreground(colorDim),

		Bar: lipgloss.NewStyle().
			Foreground(colorSuccess),
	}
}

// UpdateState updates the dashboard state
func (md *MetricsDashboard) UpdateState(state MLState) {
	md.state = state
}

// UpdateSize updates the dashboard size
func (md *MetricsDashboard) UpdateSize(width, height int) {
	md.width = width
	md.height = height
}

// Render renders the metrics dashboard
func (md *MetricsDashboard) Render() string {
	if !md.state.Active {
		return md.renderInactive()
	}

	// Build content
	var content strings.Builder

	// Title
	title := md.styles.Title.Render("📊 ML Pipeline Metrics")
	content.WriteString(title)
	content.WriteString("\n")

	// Performance metrics
	performance := md.renderPerformanceMetrics()
	content.WriteString(performance)
	content.WriteString("\n")

	// Separator
	separator := md.styles.Separator.Render(strings.Repeat("─", md.width-4))
	content.WriteString(separator)
	content.WriteString("\n")

	// Resource usage
	resources := md.renderResourceUsage()
	content.WriteString(resources)
	content.WriteString("\n")

	// Processor stats
	processors := md.renderProcessorStats()
	content.WriteString(processors)

	// Apply container style
	return md.styles.Container.
		Width(md.width).
		Height(md.height).
		Render(content.String())
}

// renderInactive renders the inactive state
func (md *MetricsDashboard) renderInactive() string {
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Render("ML pipeline inactive")

	return md.styles.Container.
		Width(md.width).
		Height(md.height).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				md.styles.Title.Render("📊 ML Pipeline Metrics"),
				"",
				message,
			),
		)
}

// renderPerformanceMetrics renders performance metrics
func (md *MetricsDashboard) renderPerformanceMetrics() string {
	metrics := md.state.Metrics

	// Format values
	fps := fmt.Sprintf("%.1f", metrics.FPS)
	latency := formatDurationForMetrics(metrics.Latency)
	droppedFrames := fmt.Sprintf("%d", metrics.DroppedFrames)

	// Determine FPS color
	var fpsStyle lipgloss.Style
	if metrics.FPS >= 30 {
		fpsStyle = md.styles.Good
	} else if metrics.FPS >= 15 {
		fpsStyle = md.styles.Warning
	} else {
		fpsStyle = md.styles.Error
	}

	// Determine latency color
	var latencyStyle lipgloss.Style
	if metrics.Latency < 50*time.Millisecond {
		latencyStyle = md.styles.Good
	} else if metrics.Latency < 100*time.Millisecond {
		latencyStyle = md.styles.Warning
	} else {
		latencyStyle = md.styles.Error
	}

	// Build metrics
	lines := []string{
		md.renderMetric("FPS", fps, fpsStyle),
		md.renderMetric("Latency", latency, latencyStyle),
		md.renderMetric("Dropped Frames", droppedFrames, md.styles.Value),
	}

	return strings.Join(lines, "\n")
}

// renderResourceUsage renders resource usage metrics
func (md *MetricsDashboard) renderResourceUsage() string {
	metrics := md.state.Metrics

	// Format values
	memory := formatMemory(metrics.MemoryUsage)
	gpuUsage := fmt.Sprintf("%.1f%%", metrics.GPUUsage)

	// Determine memory color
	var memoryStyle lipgloss.Style
	if metrics.MemoryUsage < 100*1024*1024 { // < 100MB
		memoryStyle = md.styles.Good
	} else if metrics.MemoryUsage < 500*1024*1024 { // < 500MB
		memoryStyle = md.styles.Warning
	} else {
		memoryStyle = md.styles.Error
	}

	// Determine GPU color
	var gpuStyle lipgloss.Style
	if metrics.GPUUsage < 50 {
		gpuStyle = md.styles.Good
	} else if metrics.GPUUsage < 80 {
		gpuStyle = md.styles.Warning
	} else {
		gpuStyle = md.styles.Error
	}

	// Build metrics
	lines := []string{
		md.renderMetric("Memory", memory, memoryStyle),
		md.renderMetric("GPU Usage", gpuUsage, gpuStyle),
	}

	return strings.Join(lines, "\n")
}

// renderProcessorStats renders processor statistics
func (md *MetricsDashboard) renderProcessorStats() string {
	if len(md.state.Metrics.ProcessorStats) == 0 {
		return md.styles.Metric.Render("No processor stats available")
	}

	var lines []string
	lines = append(lines, md.styles.Metric.Render("Processors:"))

	// Show top 3 processors by success count
	processors := md.state.Metrics.ProcessorStats
	count := 0
	for name, stats := range processors {
		if count >= 3 { // Limit to 3 processors for space
			break
		}

		successRate := float64(stats.SuccessCount) / float64(stats.SuccessCount+stats.ErrorCount) * 100
		successRateStr := fmt.Sprintf("%.1f%%", successRate)
		latencyStr := formatDurationForMetrics(stats.AvgLatency)

		processorLine := fmt.Sprintf("  %s: %s success, %s avg",
			name, successRateStr, latencyStr)

		// Determine color based on success rate
		var style lipgloss.Style
		if successRate >= 95 {
			style = md.styles.Good
		} else if successRate >= 80 {
			style = md.styles.Warning
		} else {
			style = md.styles.Error
		}

		lines = append(lines, style.Render(processorLine))
		count++
	}

	return strings.Join(lines, "\n")
}

// renderMetric renders a single metric
func (md *MetricsDashboard) renderMetric(label, value string, valueStyle lipgloss.Style) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		md.styles.Label.Render(label),
		valueStyle.Render(value),
	)
}

// renderGraph renders a simple bar graph
func (md *MetricsDashboard) renderGraph(value, max float64, width int) string {
	if max <= 0 {
		max = 1
	}

	percentage := value / max
	if percentage > 1 {
		percentage = 1
	}

	barWidth := int(float64(width) * percentage)
	bar := strings.Repeat("█", barWidth)
	empty := strings.Repeat("░", width-barWidth)

	return md.styles.Bar.Render(bar) + md.styles.Graph.Render(empty)
}

// GetState returns the current ML state
func (md *MetricsDashboard) GetState() MLState {
	return md.state
}

// GetWidth returns the dashboard width
func (md *MetricsDashboard) GetWidth() int {
	return md.width
}

// GetHeight returns the dashboard height
func (md *MetricsDashboard) GetHeight() int {
	return md.height
}

// HandleKey handles keyboard input
func (md *MetricsDashboard) HandleKey(key string) {
	// Handle key bindings for metrics dashboard
	switch key {
	case "m":
		// Toggle metrics display
		// This would be implemented in the parent component
	case "g":
		// Toggle graphs
		// This would be implemented in the parent component
	case "r":
		// Reset metrics
		// This would be implemented in the parent component
	}
}

// GetHelp returns help information for the dashboard
func (md *MetricsDashboard) GetHelp() []string {
	return []string{
		"[M] Toggle metrics",
		"[G] Toggle graphs",
		"[R] Reset metrics",
		"[↑/↓] Scroll metrics",
	}
}

// Helper functions

// formatDurationForMetrics formats a duration for display in metrics
func formatDurationForMetrics(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// formatMemory formats memory usage for display
func formatMemory(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
