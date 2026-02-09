package layout

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Splitter represents a draggable splitter between panes
type Splitter struct {
	ID        string
	Position  int  // Position in pixels
	Min       int  // Minimum position
	Max       int  // Maximum position
	Vertical  bool // True for vertical splitter, false for horizontal
	Dragging  bool
	LastMouse tea.MouseMsg
}

// NewSplitter creates a new splitter
func NewSplitter(id string, position, min, max int, vertical bool) *Splitter {
	return &Splitter{
		ID:       id,
		Position: position,
		Min:      min,
		Max:      max,
		Vertical: vertical,
		Dragging: false,
	}
}

// Update handles mouse events for dragging
func (s *Splitter) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return s.handleMouse(msg)
	}
	return nil
}

// handleMouse processes mouse events
func (s *Splitter) handleMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg.Type {
	case tea.MouseLeft:
		// Check if mouse is over splitter
		if s.isOver(msg.X, msg.Y) {
			s.Dragging = true
			s.LastMouse = msg
			return nil
		}
	case tea.MouseRelease:
		s.Dragging = false
	case tea.MouseMotion:
		if s.Dragging {
			// Calculate delta
			var delta int
			if s.Vertical {
				delta = msg.X - s.LastMouse.X
			} else {
				delta = msg.Y - s.LastMouse.Y
			}

			// Update position with constraints
			newPos := s.Position + delta
			if newPos < s.Min {
				newPos = s.Min
			}
			if newPos > s.Max {
				newPos = s.Max
			}

			s.Position = newPos
			s.LastMouse = msg
			return func() tea.Msg {
				return SplitterMovedMsg{
					ID:       s.ID,
					Position: s.Position,
				}
			}
		}
	}
	return nil
}

// isOver checks if mouse coordinates are over the splitter
func (s *Splitter) isOver(x, y int) bool {
	// Splitter is 3 pixels wide/high for easier clicking
	if s.Vertical {
		return x >= s.Position-1 && x <= s.Position+1
	} else {
		return y >= s.Position-1 && y <= s.Position+1
	}
}

// Render renders the splitter
func (s *Splitter) Render(width, height int) string {
	if s.Vertical {
		return s.renderVertical(width, height)
	} else {
		return s.renderHorizontal(width, height)
	}
}

// renderVertical renders a vertical splitter
func (s *Splitter) renderVertical(width, height int) string {
	splitterStyle := lipgloss.NewStyle().
		Width(1).
		Height(height).
		Background(lipgloss.Color("240"))

	if s.Dragging {
		splitterStyle = splitterStyle.Background(lipgloss.Color("205"))
	}

	// Create a vertical line
	line := "│"
	if s.Dragging {
		line = "┃"
	}

	return splitterStyle.Render(line)
}

// renderHorizontal renders a horizontal splitter
func (s *Splitter) renderHorizontal(width, height int) string {
	splitterStyle := lipgloss.NewStyle().
		Width(width).
		Height(1).
		Background(lipgloss.Color("240"))

	if s.Dragging {
		splitterStyle = splitterStyle.Background(lipgloss.Color("205"))
	}

	// Create a horizontal line
	line := "─"
	if s.Dragging {
		line = "━"
	}

	return splitterStyle.Render(line)
}

// SplitterMovedMsg is sent when a splitter is moved
type SplitterMovedMsg struct {
	ID       string
	Position int
}

// SplitterManager manages multiple splitters
type SplitterManager struct {
	splitters map[string]*Splitter
}

// NewSplitterManager creates a new splitter manager
func NewSplitterManager() *SplitterManager {
	return &SplitterManager{
		splitters: make(map[string]*Splitter),
	}
}

// AddSplitter adds a splitter to the manager
func (sm *SplitterManager) AddSplitter(id string, position, min, max int, vertical bool) {
	sm.splitters[id] = NewSplitter(id, position, min, max, vertical)
}

// RemoveSplitter removes a splitter from the manager
func (sm *SplitterManager) RemoveSplitter(id string) {
	delete(sm.splitters, id)
}

// GetSplitter returns a splitter by ID
func (sm *SplitterManager) GetSplitter(id string) (*Splitter, bool) {
	splitter, exists := sm.splitters[id]
	return splitter, exists
}

// Update updates all splitters
func (sm *SplitterManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, splitter := range sm.splitters {
		if cmd := splitter.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

// Render renders all splitters
func (sm *SplitterManager) Render(width, height int) string {
	var rendered []string
	for _, splitter := range sm.splitters {
		rendered = append(rendered, splitter.Render(width, height))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// HandleSplitterMessage handles splitter movement messages
func HandleSplitterMessage(msg SplitterMovedMsg, panes []Pane) []Pane {
	// This would be implemented to adjust pane sizes based on splitter movement
	// For now, return the panes unchanged
	return panes
}
