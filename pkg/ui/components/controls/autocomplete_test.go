package controls

import (
	"testing"
)

func TestNewAutocomplete(t *testing.T) {
	ac := NewAutocomplete(80, 10)
	if ac == nil {
		t.Fatal("NewAutocomplete returned nil")
	}
	if ac.width != 80 || ac.height != 10 {
		t.Errorf("Expected width=80, height=10, got width=%d, height=%d", ac.width, ac.height)
	}
	if ac.maxSuggestions != 5 {
		t.Errorf("Expected maxSuggestions=5, got %d", ac.maxSuggestions)
	}
	if len(ac.allCommands) == 0 {
		t.Error("Expected some commands to be initialized")
	}
}

func TestAutocomplete_UpdatePrefix(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Test with empty prefix
	ac.UpdatePrefix("")
	if ac.IsVisible() {
		t.Error("Autocomplete should not be visible with empty prefix")
	}
	if ac.GetSuggestionsCount() != 0 {
		t.Error("Should have no suggestions with empty prefix")
	}

	// Test with matching prefix
	ac.UpdatePrefix("take")
	if !ac.IsVisible() {
		t.Error("Autocomplete should be visible with matching prefix")
	}
	if ac.GetSuggestionsCount() == 0 {
		t.Error("Should have suggestions with matching prefix")
	}

	// Test prefix is trimmed
	ac.UpdatePrefix("  take  ")
	if ac.GetPrefix() != "take" {
		t.Errorf("Expected prefix 'take', got '%s'", ac.GetPrefix())
	}
}

func TestAutocomplete_SelectNextPrevious(t *testing.T) {
	ac := NewAutocomplete(80, 10)
	ac.UpdatePrefix("take")

	initialCount := ac.GetSuggestionsCount()
	if initialCount == 0 {
		t.Fatal("Need suggestions to test selection")
	}

	// Test SelectNext
	initialIndex := ac.selectedIndex
	ac.SelectNext()
	if ac.selectedIndex != (initialIndex+1)%initialCount {
		t.Error("SelectNext should increment index")
	}

	// Test SelectPrevious
	ac.SelectPrevious()
	if ac.selectedIndex != initialIndex {
		t.Error("SelectPrevious should decrement index")
	}

	// Test wrap around
	for i := 0; i < initialCount; i++ {
		ac.SelectNext()
	}
	if ac.selectedIndex != initialIndex {
		t.Error("SelectNext should wrap around")
	}

	// Test with no suggestions (should not panic)
	ac.Hide()
	ac.SelectNext()
	ac.SelectPrevious()
	// Should not panic
}

func TestAutocomplete_GetSelectedSuggestion(t *testing.T) {
	ac := NewAutocomplete(80, 10)
	ac.UpdatePrefix("take")

	// Test with suggestions
	suggestion, ok := ac.GetSelectedSuggestion()
	if !ok {
		t.Error("Should get selected suggestion when available")
	}
	if suggestion.Command == "" {
		t.Error("Suggestion should have a command")
	}

	// Test without suggestions
	ac.Hide()
	suggestion, ok = ac.GetSelectedSuggestion()
	if ok {
		t.Error("Should not get suggestion when hidden")
	}
}

func TestAutocomplete_GetCompletion(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Test with prefix only
	ac.UpdatePrefix("tak")
	completion := ac.GetCompletion()
	if completion != "tak" {
		t.Errorf("Expected completion 'tak', got '%s'", completion)
	}

	// Test with selected suggestion
	ac.UpdatePrefix("take")
	ac.SelectNext() // Select first suggestion
	completion = ac.GetCompletion()
	if completion == "" {
		t.Error("Completion should not be empty")
	}
}

func TestAutocomplete_Hide(t *testing.T) {
	ac := NewAutocomplete(80, 10)
	ac.UpdatePrefix("take")

	if !ac.IsVisible() {
		t.Error("Autocomplete should be visible")
	}

	ac.Hide()

	if ac.IsVisible() {
		t.Error("Autocomplete should be hidden after Hide()")
	}
	if ac.selectedIndex != 0 {
		t.Error("Selected index should be reset after Hide()")
	}
	if len(ac.suggestions) != 0 {
		t.Error("Suggestions should be cleared after Hide()")
	}
}

func TestAutocomplete_HandleKey(t *testing.T) {
	ac := NewAutocomplete(80, 10)
	ac.UpdatePrefix("take")

	// Test Tab
	handled, completion := ac.HandleKey("tab")
	if !handled {
		t.Error("Tab should be handled")
	}
	if completion == "" {
		t.Error("Tab should return completion")
	}

	// Test Shift+Tab
	handled, _ = ac.HandleKey("shift+tab")
	if !handled {
		t.Error("Shift+Tab should be handled")
	}

	// Test Up
	handled, _ = ac.HandleKey("up")
	if !handled {
		t.Error("Up should be handled")
	}

	// Test Down
	handled, _ = ac.HandleKey("down")
	if !handled {
		t.Error("Down should be handled")
	}

	// Test Enter
	handled, completion = ac.HandleKey("enter")
	if !handled {
		t.Error("Enter should be handled")
	}
	if completion == "" {
		t.Error("Enter should return completion")
	}
	if ac.IsVisible() {
		t.Error("Autocomplete should be hidden after Enter")
	}

	// Test Esc
	ac.UpdatePrefix("take")
	handled, completion = ac.HandleKey("esc")
	if !handled {
		t.Error("Esc should be handled")
	}
	if ac.IsVisible() {
		t.Error("Autocomplete should be hidden after Esc")
	}

	// Test other key (should not be handled)
	handled, _ = ac.HandleKey("a")
	if handled {
		t.Error("Other keys should not be handled")
	}

	// Test when not visible
	ac.Hide()
	handled, _ = ac.HandleKey("tab")
	if handled {
		t.Error("Keys should not be handled when not visible")
	}
}

func TestAutocomplete_AddRemoveCommand(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Add custom command
	customCmd := "customcommand"
	ac.AddCommand(customCmd, "Custom command", "custom")

	// Verify command was added
	ac.UpdatePrefix("custom")
	if ac.GetSuggestionsCount() == 0 {
		t.Error("Custom command should be in suggestions")
	}

	// Remove command
	ac.RemoveCommand(customCmd)

	// Verify command was removed
	ac.UpdatePrefix("custom")
	if ac.GetSuggestionsCount() != 0 {
		t.Error("Custom command should be removed from suggestions")
	}
}

func TestAutocomplete_SetMaxSuggestions(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Set max suggestions
	ac.SetMaxSuggestions(3)
	if ac.maxSuggestions != 3 {
		t.Errorf("Expected maxSuggestions=3, got %d", ac.maxSuggestions)
	}

	// Test that suggestions are limited
	ac.UpdatePrefix("") // Update with empty to reset
	ac.UpdatePrefix("") // This should show all commands starting with empty string
	// Note: empty prefix doesn't show suggestions by design
}

func TestAutocomplete_GetFilteredCount(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Test with matching prefix
	ac.UpdatePrefix("take")
	filteredCount := ac.GetFilteredCount()
	if filteredCount == 0 {
		t.Error("Should have filtered commands")
	}

	// Test with non-matching prefix
	ac.UpdatePrefix("xyz123")
	filteredCount = ac.GetFilteredCount()
	if filteredCount != 0 {
		t.Error("Should have no filtered commands with non-matching prefix")
	}
}

func TestAutocomplete_Render(t *testing.T) {
	ac := NewAutocomplete(80, 10)

	// Should return empty string when not visible
	rendered := ac.Render()
	if rendered != "" {
		t.Error("Render should return empty string when not visible")
	}

	// Show suggestions and render
	ac.UpdatePrefix("take")
	rendered = ac.Render()
	if rendered == "" {
		t.Error("Render should return non-empty string when visible")
	}

	// Check for expected content
	if len(rendered) == 0 {
		t.Error("Rendered autocomplete should have content")
	}
}
