package layout

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Pane represents a resizable pane in the layout
type Pane struct {
	ID        string
	Title     string
	Content   string
	Width     int
	Height    int
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
	Resizable bool
	Focused   bool
	Visible   bool
}

// LayoutManager manages the multi-pane layout system
type LayoutManager struct {
	panes       []Pane
	layoutType  LayoutType
	splitterPos int // Position of splitter in pixels
	focusedPane string
	width       int
	height      int
}

// LayoutType defines the type of layout arrangement
type LayoutType int

const (
	LayoutHorizontal LayoutType = iota
	LayoutVertical
	LayoutGrid
	LayoutTabbed
)

// NewLayoutManager creates a new layout manager
func NewLayoutManager(width, height int) *LayoutManager {
	return &LayoutManager{
		panes:      []Pane{},
		layoutType: LayoutHorizontal,
		width:      width,
		height:     height,
	}
}

// AddPane adds a new pane to the layout
func (lm *LayoutManager) AddPane(id, title string, minWidth, minHeight int) {
	pane := Pane{
		ID:        id,
		Title:     title,
		Width:     minWidth,
		Height:    minHeight,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		MaxWidth:  -1, // Unlimited
		MaxHeight: -1, // Unlimited
		Resizable: true,
		Focused:   false,
		Visible:   true,
	}
	lm.panes = append(lm.panes, pane)
	// Don't recalculate layout here - let it be done explicitly
}

// RemovePane removes a pane from the layout
func (lm *LayoutManager) RemovePane(id string) {
	for i, pane := range lm.panes {
		if pane.ID == id {
			lm.panes = append(lm.panes[:i], lm.panes[i+1:]...)
			lm.recalculateLayout()
			return
		}
	}
}

// SetLayoutType changes the layout arrangement
func (lm *LayoutManager) SetLayoutType(layoutType LayoutType) {
	lm.layoutType = layoutType
	lm.recalculateLayout()
}

// SetPaneContent updates the content of a pane
func (lm *LayoutManager) SetPaneContent(id, content string) {
	for i, pane := range lm.panes {
		if pane.ID == id {
			lm.panes[i].Content = content
			return
		}
	}
}

// SetPaneVisibility shows or hides a pane
func (lm *LayoutManager) SetPaneVisibility(id string, visible bool) {
	for i, pane := range lm.panes {
		if pane.ID == id {
			lm.panes[i].Visible = visible
			lm.recalculateLayout()
			return
		}
	}
}

// FocusPane sets focus to a specific pane
func (lm *LayoutManager) FocusPane(id string) {
	for i, pane := range lm.panes {
		lm.panes[i].Focused = (pane.ID == id)
	}
	lm.focusedPane = id
}

// ResizePane resizes a pane by the given delta
func (lm *LayoutManager) ResizePane(id string, deltaWidth, deltaHeight int) {
	for i, pane := range lm.panes {
		if pane.ID == id && pane.Resizable {
			newWidth := pane.Width + deltaWidth
			newHeight := pane.Height + deltaHeight

			// Apply constraints
			if newWidth < pane.MinWidth {
				newWidth = pane.MinWidth
			}
			if pane.MaxWidth > 0 && newWidth > pane.MaxWidth {
				newWidth = pane.MaxWidth
			}
			if newHeight < pane.MinHeight {
				newHeight = pane.MinHeight
			}
			if pane.MaxHeight > 0 && newHeight > pane.MaxHeight {
				newHeight = pane.MaxHeight
			}

			lm.panes[i].Width = newWidth
			lm.panes[i].Height = newHeight
			// Don't recalculate layout for manual resizes
			return
		}
	}
}

// GetPane returns a pane by ID
func (lm *LayoutManager) GetPane(id string) (Pane, bool) {
	for _, pane := range lm.panes {
		if pane.ID == id {
			return pane, true
		}
	}
	return Pane{}, false
}

// GetFocusedPane returns the currently focused pane
func (lm *LayoutManager) GetFocusedPane() (Pane, bool) {
	for _, pane := range lm.panes {
		if pane.Focused {
			return pane, true
		}
	}
	return Pane{}, false
}

// GetVisiblePanes returns all visible panes
func (lm *LayoutManager) GetVisiblePanes() []Pane {
	var visible []Pane
	for _, pane := range lm.panes {
		if pane.Visible {
			visible = append(visible, pane)
		}
	}
	return visible
}

// Update handles window resize events
func (lm *LayoutManager) Update(width, height int) {
	lm.width = width
	lm.height = height
	lm.recalculateLayout()
}

// Render renders the entire layout
func (lm *LayoutManager) Render() string {
	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) == 0 {
		return ""
	}

	switch lm.layoutType {
	case LayoutHorizontal:
		return lm.renderHorizontal(visiblePanes)
	case LayoutVertical:
		return lm.renderVertical(visiblePanes)
	case LayoutGrid:
		return lm.renderGrid(visiblePanes)
	case LayoutTabbed:
		return lm.renderTabbed(visiblePanes)
	default:
		return lm.renderHorizontal(visiblePanes)
	}
}

// recalculateLayout recalculates pane sizes based on layout type
func (lm *LayoutManager) recalculateLayout() {
	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) == 0 {
		return
	}

	switch lm.layoutType {
	case LayoutHorizontal:
		lm.recalculateHorizontal(visiblePanes)
	case LayoutVertical:
		lm.recalculateVertical(visiblePanes)
	case LayoutGrid:
		lm.recalculateGrid(visiblePanes)
	case LayoutTabbed:
		// Tabbed layout doesn't need recalculation
	}
}

// recalculateHorizontal distributes width among horizontal panes
func (lm *LayoutManager) recalculateHorizontal(panes []Pane) {
	totalMinWidth := 0
	for _, pane := range panes {
		totalMinWidth += pane.MinWidth
	}

	availableWidth := lm.width
	if totalMinWidth > availableWidth {
		// Not enough space, use minimum widths
		for i := range lm.panes {
			if lm.panes[i].Visible {
				lm.panes[i].Width = lm.panes[i].MinWidth
			}
		}
		return
	}

	// Distribute extra space proportionally
	extraWidth := availableWidth - totalMinWidth
	widthPerPane := extraWidth / len(panes)

	for i := range lm.panes {
		if lm.panes[i].Visible {
			lm.panes[i].Width = lm.panes[i].MinWidth + widthPerPane
			lm.panes[i].Height = lm.height
		}
	}
}

// recalculateVertical distributes height among vertical panes
func (lm *LayoutManager) recalculateVertical(panes []Pane) {
	totalMinHeight := 0
	for _, pane := range panes {
		totalMinHeight += pane.MinHeight
	}

	availableHeight := lm.height
	if totalMinHeight > availableHeight {
		// Not enough space, use minimum heights
		for i := range lm.panes {
			if lm.panes[i].Visible {
				lm.panes[i].Height = lm.panes[i].MinHeight
			}
		}
		return
	}

	// Distribute extra space proportionally
	extraHeight := availableHeight - totalMinHeight
	heightPerPane := extraHeight / len(panes)

	for i := range lm.panes {
		if lm.panes[i].Visible {
			lm.panes[i].Height = lm.panes[i].MinHeight + heightPerPane
			lm.panes[i].Width = lm.width
		}
	}
}

// recalculateGrid arranges panes in a grid
func (lm *LayoutManager) recalculateGrid(panes []Pane) {
	// Simple 2-column grid for now
	cols := 2
	rows := (len(panes) + 1) / 2

	colWidth := lm.width / cols
	rowHeight := lm.height / rows

	paneIndex := 0
	for i := range lm.panes {
		if lm.panes[i].Visible {
			lm.panes[i].Width = colWidth
			lm.panes[i].Height = rowHeight
			paneIndex++
		}
	}
}

// renderHorizontal renders panes horizontally
func (lm *LayoutManager) renderHorizontal(panes []Pane) string {
	var renderedPanes []string
	for _, pane := range panes {
		renderedPanes = append(renderedPanes, lm.renderPane(pane))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedPanes...)
}

// renderVertical renders panes vertically
func (lm *LayoutManager) renderVertical(panes []Pane) string {
	var renderedPanes []string
	for _, pane := range panes {
		renderedPanes = append(renderedPanes, lm.renderPane(pane))
	}
	return lipgloss.JoinVertical(lipgloss.Left, renderedPanes...)
}

// renderGrid renders panes in a grid
func (lm *LayoutManager) renderGrid(panes []Pane) string {
	// Simple 2-column grid
	var rows []string
	for i := 0; i < len(panes); i += 2 {
		rowPanes := []string{lm.renderPane(panes[i])}
		if i+1 < len(panes) {
			rowPanes = append(rowPanes, lm.renderPane(panes[i+1]))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rowPanes...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderTabbed renders panes in a tabbed interface
func (lm *LayoutManager) renderTabbed(panes []Pane) string {
	// For now, just show the focused pane
	focusedPane, found := lm.GetFocusedPane()
	if !found && len(panes) > 0 {
		focusedPane = panes[0]
	}
	return lm.renderPane(focusedPane)
}

// renderPane renders a single pane with title bar
func (lm *LayoutManager) renderPane(pane Pane) string {
	// Pane styles
	paneStyle := lipgloss.NewStyle().
		Width(pane.Width).
		Height(pane.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	if pane.Focused {
		paneStyle = paneStyle.BorderForeground(lipgloss.Color("205"))
	}

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Bold(true)

	titleBar := titleStyle.Render(pane.Title)

	// Content area
	contentStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(pane.Width - 2).  // Account for border
		Height(pane.Height - 2) // Account for border and title

	content := contentStyle.Render(pane.Content)

	// Combine
	return paneStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleBar,
			content,
		),
	)
}

// HandleKey handles keyboard input for layout management
func (lm *LayoutManager) HandleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+h":
		lm.SetLayoutType(LayoutHorizontal)
	case "ctrl+v":
		lm.SetLayoutType(LayoutVertical)
	case "ctrl+g":
		lm.SetLayoutType(LayoutGrid)
	case "ctrl+t":
		lm.SetLayoutType(LayoutTabbed)
	case "tab":
		lm.cycleFocus()
	case "shift+tab":
		lm.cycleFocusReverse()
	}
	return nil
}

// cycleFocus cycles focus to the next pane
func (lm *LayoutManager) cycleFocus() {
	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) == 0 {
		return
	}

	// Find current focused pane index
	currentIndex := -1
	for i, pane := range visiblePanes {
		if pane.Focused {
			currentIndex = i
			break
		}
	}

	// Focus next pane
	nextIndex := (currentIndex + 1) % len(visiblePanes)
	lm.FocusPane(visiblePanes[nextIndex].ID)
}

// cycleFocusReverse cycles focus to the previous pane
func (lm *LayoutManager) cycleFocusReverse() {
	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) == 0 {
		return
	}

	// Find current focused pane index
	currentIndex := -1
	for i, pane := range visiblePanes {
		if pane.Focused {
			currentIndex = i
			break
		}
	}

	// Focus previous pane
	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(visiblePanes) - 1
	}
	lm.FocusPane(visiblePanes[prevIndex].ID)
}

// SaveLayout saves the current layout configuration
func (lm *LayoutManager) SaveLayout() string {
	// Simple JSON-like representation
	var layout []string
	for _, pane := range lm.panes {
		layout = append(layout, fmt.Sprintf(
			`{"id":"%s","title":"%s","width":%d,"height":%d,"visible":%v}`,
			pane.ID, pane.Title, pane.Width, pane.Height, pane.Visible,
		))
	}
	return fmt.Sprintf(`{"layoutType":%d,"panes":[%s]}`, lm.layoutType, strings.Join(layout, ","))
}

// LoadLayout loads a layout configuration
func (lm *LayoutManager) LoadLayout(config string) error {
	// Simple implementation - in a real system, you'd parse JSON
	// For now, just reset and add default panes
	lm.panes = []Pane{}
	lm.layoutType = LayoutHorizontal

	// Add default panes
	lm.AddPane("telemetry", "Telemetry", 30, 20)
	lm.AddPane("logs", "Logs", 50, 20)
	lm.AddPane("safety", "Safety", 40, 20)

	lm.FocusPane("telemetry")
	return nil
}
