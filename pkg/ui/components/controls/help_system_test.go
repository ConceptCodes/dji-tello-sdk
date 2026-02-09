package controls

import (
	"testing"
)

func TestNewHelpSystem(t *testing.T) {
	hs := NewHelpSystem(80, 24)
	if hs == nil {
		t.Fatal("NewHelpSystem returned nil")
	}
	if hs.width != 80 || hs.height != 24 {
		t.Errorf("Expected width=80, height=24, got width=%d, height=%d", hs.width, hs.height)
	}
	if !hs.autoShow {
		t.Error("autoShow should be true by default")
	}
	if hs.showHelp {
		t.Error("showHelp should be false by default")
	}
	if hs.detailLevel != DetailLevelBasic {
		t.Errorf("Expected detail level Basic, got %v", hs.detailLevel)
	}
}

func TestHelpSystem_ToggleHelp(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Initially hidden
	if hs.IsVisible() {
		t.Error("Help should be hidden initially")
	}

	// Toggle on
	hs.ToggleHelp()
	if !hs.IsVisible() {
		t.Error("Help should be visible after toggle")
	}

	// Toggle off
	hs.ToggleHelp()
	if hs.IsVisible() {
		t.Error("Help should be hidden after second toggle")
	}
}

func TestHelpSystem_CycleDetailLevel(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Start at Basic
	if hs.detailLevel != DetailLevelBasic {
		t.Errorf("Expected DetailLevelBasic, got %v", hs.detailLevel)
	}

	// Cycle to Advanced
	hs.CycleDetailLevel()
	if hs.detailLevel != DetailLevelAdvanced {
		t.Errorf("Expected DetailLevelAdvanced, got %v", hs.detailLevel)
	}

	// Cycle to Expert
	hs.CycleDetailLevel()
	if hs.detailLevel != DetailLevelExpert {
		t.Errorf("Expected DetailLevelExpert, got %v", hs.detailLevel)
	}

	// Cycle back to Basic
	hs.CycleDetailLevel()
	if hs.detailLevel != DetailLevelBasic {
		t.Errorf("Expected DetailLevelBasic after wrap, got %v", hs.detailLevel)
	}
}

func TestHelpSystem_UpdateContext(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	context := HelpContext{
		Component:  "telemetry",
		Mode:       "flight",
		Focused:    "dashboard",
		LastAction: "takeoff",
	}

	hs.UpdateContext(context)

	updatedContext := hs.GetContext()
	if updatedContext.Component != "telemetry" {
		t.Errorf("Expected component 'telemetry', got %s", updatedContext.Component)
	}
	if updatedContext.Mode != "flight" {
		t.Errorf("Expected mode 'flight', got %s", updatedContext.Mode)
	}
	if updatedContext.Focused != "dashboard" {
		t.Errorf("Expected focused 'dashboard', got %s", updatedContext.Focused)
	}
	if updatedContext.LastAction != "takeoff" {
		t.Errorf("Expected last action 'takeoff', got %s", updatedContext.LastAction)
	}
}

func TestHelpSystem_GetCurrentTopic(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Test with telemetry context
	hs.UpdateContext(HelpContext{Component: "telemetry"})
	topic := hs.GetCurrentTopic()
	if topic.ID != "telemetry" {
		t.Errorf("Expected topic ID 'telemetry', got %s", topic.ID)
	}
	if topic.Title != "Telemetry Dashboard" {
		t.Errorf("Expected title 'Telemetry Dashboard', got %s", topic.Title)
	}

	// Test with unknown context (should return general)
	hs.UpdateContext(HelpContext{Component: "unknown"})
	topic = hs.GetCurrentTopic()
	if topic.ID != "general" {
		t.Errorf("Expected topic ID 'general' for unknown context, got %s", topic.ID)
	}
}

func TestHelpSystem_AddRemoveTopic(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Add custom topic
	customTopic := HelpTopic{
		ID:    "custom",
		Title: "Custom Topic",
		Description: map[DetailLevel]string{
			DetailLevelBasic: "Basic custom help",
		},
		Shortcuts: map[string]string{
			"Ctrl+X": "Custom action",
		},
		Examples: []string{"Example 1", "Example 2"},
	}

	hs.AddTopic(customTopic)

	// Verify topic was added
	hs.UpdateContext(HelpContext{Component: "custom"})
	topic := hs.GetCurrentTopic()
	if topic.ID != "custom" {
		t.Errorf("Expected topic ID 'custom', got %s", topic.ID)
	}

	// Remove topic
	hs.RemoveTopic("custom")

	// Should fall back to general
	hs.UpdateContext(HelpContext{Component: "custom"})
	topic = hs.GetCurrentTopic()
	if topic.ID != "general" {
		t.Errorf("Expected topic ID 'general' after removal, got %s", topic.ID)
	}
}

func TestHelpSystem_HandleKey(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Test F1 toggles help
	hs.HandleKey("f1")
	if !hs.IsVisible() {
		t.Error("F1 should show help")
	}

	hs.HandleKey("f1")
	if hs.IsVisible() {
		t.Error("Second F1 should hide help")
	}

	// Test F2 cycles detail level
	hs.HandleKey("f1") // Show help first
	initialLevel := hs.detailLevel
	hs.HandleKey("f2")
	if hs.detailLevel == initialLevel {
		t.Error("F2 should cycle detail level")
	}

	// Test ESC closes help
	hs.HandleKey("esc")
	if hs.IsVisible() {
		t.Error("ESC should hide help")
	}

	// Test other keys don't affect help when hidden
	hs.HandleKey("up")
	hs.HandleKey("down")
	hs.HandleKey("left")
	hs.HandleKey("right")
	// Should not panic or change state
}

func TestHelpSystem_Render(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Should return empty string when help not visible
	rendered := hs.Render()
	if rendered != "" {
		t.Error("Render should return empty string when help not visible")
	}

	// Show help and render
	hs.ToggleHelp()
	rendered = hs.Render()
	if rendered == "" {
		t.Error("Render should return non-empty string when help visible")
	}

	// Check for expected content
	if len(rendered) == 0 {
		t.Error("Rendered help should have content")
	}

	// Test with different contexts
	hs.UpdateContext(HelpContext{Component: "safety"})
	rendered = hs.Render()
	if rendered == "" {
		t.Error("Render should work with safety context")
	}

	hs.UpdateContext(HelpContext{Component: "layout"})
	rendered = hs.Render()
	if rendered == "" {
		t.Error("Render should work with layout context")
	}

	hs.UpdateContext(HelpContext{Component: "flight"})
	rendered = hs.Render()
	if rendered == "" {
		t.Error("Render should work with flight context")
	}

	hs.UpdateContext(HelpContext{Component: "commands"})
	rendered = hs.Render()
	if rendered == "" {
		t.Error("Render should work with commands context")
	}
}

func TestHelpSystem_AutoShow(t *testing.T) {
	hs := NewHelpSystem(80, 24)

	// Test auto-show for error context
	hs.UpdateContext(HelpContext{Component: "error"})
	if !hs.IsVisible() {
		t.Error("Help should auto-show for error context")
	}

	// Hide and test unknown context
	hs.ToggleHelp()
	hs.UpdateContext(HelpContext{Component: "unknown"})
	if !hs.IsVisible() {
		t.Error("Help should auto-show for unknown context")
	}

	// Test config context
	hs.ToggleHelp()
	hs.UpdateContext(HelpContext{Component: "config"})
	if !hs.IsVisible() {
		t.Error("Help should auto-show for config context")
	}

	// Test normal context doesn't auto-show
	hs.ToggleHelp()
	hs.UpdateContext(HelpContext{Component: "telemetry"})
	if hs.IsVisible() {
		t.Error("Help should not auto-show for telemetry context")
	}

	// Disable auto-show
	hs.SetAutoShow(false)
	hs.UpdateContext(HelpContext{Component: "error"})
	if hs.IsVisible() {
		t.Error("Help should not auto-show when autoShow is false")
	}
}
