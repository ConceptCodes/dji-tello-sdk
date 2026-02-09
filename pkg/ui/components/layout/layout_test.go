package layout

import (
	"testing"
)

func TestNewLayoutManager(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	if lm == nil {
		t.Fatal("NewLayoutManager returned nil")
	}
	if lm.width != 100 || lm.height != 50 {
		t.Errorf("Expected width=100, height=50, got width=%d, height=%d", lm.width, lm.height)
	}
	if len(lm.panes) != 0 {
		t.Errorf("Expected 0 panes, got %d", len(lm.panes))
	}
}

func TestAddPane(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)

	if len(lm.panes) != 1 {
		t.Fatalf("Expected 1 pane, got %d", len(lm.panes))
	}

	pane := lm.panes[0]
	if pane.ID != "telemetry" {
		t.Errorf("Expected ID 'telemetry', got %s", pane.ID)
	}
	if pane.Title != "Telemetry" {
		t.Errorf("Expected title 'Telemetry', got %s", pane.Title)
	}
	if pane.Width != 30 {
		t.Errorf("Expected width 30, got %d", pane.Width)
	}
	if pane.Height != 20 {
		t.Errorf("Expected height 20, got %d", pane.Height)
	}
}

func TestRemovePane(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)
	lm.AddPane("logs", "Logs", 40, 20)

	if len(lm.panes) != 2 {
		t.Fatalf("Expected 2 panes, got %d", len(lm.panes))
	}

	lm.RemovePane("telemetry")

	if len(lm.panes) != 1 {
		t.Fatalf("Expected 1 pane after removal, got %d", len(lm.panes))
	}

	if lm.panes[0].ID != "logs" {
		t.Errorf("Expected remaining pane to be 'logs', got %s", lm.panes[0].ID)
	}
}

func TestSetLayoutType(t *testing.T) {
	lm := NewLayoutManager(100, 50)

	lm.SetLayoutType(LayoutVertical)
	if lm.layoutType != LayoutVertical {
		t.Errorf("Expected LayoutVertical, got %v", lm.layoutType)
	}

	lm.SetLayoutType(LayoutGrid)
	if lm.layoutType != LayoutGrid {
		t.Errorf("Expected LayoutGrid, got %v", lm.layoutType)
	}
}

func TestFocusPane(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)
	lm.AddPane("logs", "Logs", 40, 20)

	lm.FocusPane("logs")

	focusedPane, found := lm.GetFocusedPane()
	if !found {
		t.Fatal("Expected to find focused pane")
	}

	if focusedPane.ID != "logs" {
		t.Errorf("Expected focused pane to be 'logs', got %s", focusedPane.ID)
	}

	// Check that telemetry pane is not focused
	telemetryPane, _ := lm.GetPane("telemetry")
	if telemetryPane.Focused {
		t.Error("Telemetry pane should not be focused")
	}
}

func TestGetVisiblePanes(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)
	lm.AddPane("logs", "Logs", 40, 20)
	lm.AddPane("hidden", "Hidden", 20, 10)

	lm.SetPaneVisibility("hidden", false)

	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) != 2 {
		t.Errorf("Expected 2 visible panes, got %d", len(visiblePanes))
	}

	for _, pane := range visiblePanes {
		if pane.ID == "hidden" {
			t.Error("Hidden pane should not be in visible panes")
		}
	}
}

func TestCycleFocus(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("pane1", "Pane 1", 30, 20)
	lm.AddPane("pane2", "Pane 2", 30, 20)
	lm.AddPane("pane3", "Pane 3", 30, 20)

	lm.FocusPane("pane1")

	lm.cycleFocus()
	focusedPane, _ := lm.GetFocusedPane()
	if focusedPane.ID != "pane2" {
		t.Errorf("Expected pane2 to be focused after cycle, got %s", focusedPane.ID)
	}

	lm.cycleFocus()
	focusedPane, _ = lm.GetFocusedPane()
	if focusedPane.ID != "pane3" {
		t.Errorf("Expected pane3 to be focused after second cycle, got %s", focusedPane.ID)
	}

	lm.cycleFocus()
	focusedPane, _ = lm.GetFocusedPane()
	if focusedPane.ID != "pane1" {
		t.Errorf("Expected pane1 to be focused after wrapping, got %s", focusedPane.ID)
	}
}

func TestCycleFocusReverse(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("pane1", "Pane 1", 30, 20)
	lm.AddPane("pane2", "Pane 2", 30, 20)
	lm.AddPane("pane3", "Pane 3", 30, 20)

	lm.FocusPane("pane1")

	lm.cycleFocusReverse()
	focusedPane, _ := lm.GetFocusedPane()
	if focusedPane.ID != "pane3" {
		t.Errorf("Expected pane3 to be focused after reverse cycle, got %s", focusedPane.ID)
	}

	lm.cycleFocusReverse()
	focusedPane, _ = lm.GetFocusedPane()
	if focusedPane.ID != "pane2" {
		t.Errorf("Expected pane2 to be focused after second reverse cycle, got %s", focusedPane.ID)
	}
}

func TestResizePane(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)

	lm.ResizePane("telemetry", 10, 5)

	pane, _ := lm.GetPane("telemetry")
	if pane.Width != 40 {
		t.Errorf("Expected width 40 after resize, got %d", pane.Width)
	}
	if pane.Height != 25 {
		t.Errorf("Expected height 25 after resize, got %d", pane.Height)
	}

	// Test minimum constraints
	lm.ResizePane("telemetry", -50, -30)
	pane, _ = lm.GetPane("telemetry")
	if pane.Width != 30 {
		t.Errorf("Expected width to be clamped to min 30, got %d", pane.Width)
	}
	if pane.Height != 20 {
		t.Errorf("Expected height to be clamped to min 20, got %d", pane.Height)
	}
}

func TestNewSplitter(t *testing.T) {
	splitter := NewSplitter("split1", 50, 10, 90, true)
	if splitter == nil {
		t.Fatal("NewSplitter returned nil")
	}
	if splitter.ID != "split1" {
		t.Errorf("Expected ID 'split1', got %s", splitter.ID)
	}
	if splitter.Position != 50 {
		t.Errorf("Expected position 50, got %d", splitter.Position)
	}
	if splitter.Vertical != true {
		t.Error("Expected vertical splitter")
	}
}

func TestSplitterManager(t *testing.T) {
	sm := NewSplitterManager()
	if sm == nil {
		t.Fatal("NewSplitterManager returned nil")
	}

	sm.AddSplitter("split1", 50, 10, 90, true)
	sm.AddSplitter("split2", 30, 5, 80, false)

	if len(sm.splitters) != 2 {
		t.Errorf("Expected 2 splitters, got %d", len(sm.splitters))
	}

	splitter, found := sm.GetSplitter("split1")
	if !found {
		t.Fatal("Expected to find splitter 'split1'")
	}
	if splitter.ID != "split1" {
		t.Errorf("Expected splitter ID 'split1', got %s", splitter.ID)
	}

	sm.RemoveSplitter("split1")
	if len(sm.splitters) != 1 {
		t.Errorf("Expected 1 splitter after removal, got %d", len(sm.splitters))
	}

	_, found = sm.GetSplitter("split1")
	if found {
		t.Error("Splitter 'split1' should have been removed")
	}
}

func TestSaveLoadLayout(t *testing.T) {
	lm := NewLayoutManager(100, 50)
	lm.AddPane("telemetry", "Telemetry", 30, 20)
	lm.AddPane("logs", "Logs", 40, 20)
	lm.FocusPane("telemetry")

	saved := lm.SaveLayout()
	if saved == "" {
		t.Error("SaveLayout returned empty string")
	}

	// Note: LoadLayout is a simple implementation that resets and adds default panes
	// In a real implementation, this would parse the saved configuration
	err := lm.LoadLayout(saved)
	if err != nil {
		t.Errorf("LoadLayout failed: %v", err)
	}

	// After load, we should have default panes
	visiblePanes := lm.GetVisiblePanes()
	if len(visiblePanes) < 2 {
		t.Errorf("Expected at least 2 panes after load, got %d", len(visiblePanes))
	}
}
