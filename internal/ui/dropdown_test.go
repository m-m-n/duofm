package ui

import (
	"strings"
	"testing"
)

func TestNewDropdown(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	dropdown := NewDropdown(options, 0)

	if len(dropdown.options) != 3 {
		t.Errorf("options length = %d, want 3", len(dropdown.options))
	}
	if dropdown.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", dropdown.selectedIndex)
	}
	if dropdown.expanded {
		t.Error("dropdown should not be expanded initially")
	}
	if dropdown.highlightedIndex != 0 {
		t.Errorf("highlightedIndex = %d, want 0", dropdown.highlightedIndex)
	}
}

func TestDropdown_Expand(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 1)

	dropdown.Expand()

	if !dropdown.expanded {
		t.Error("dropdown should be expanded after Expand()")
	}
	if dropdown.highlightedIndex != 1 {
		t.Errorf("highlightedIndex = %d, want 1 (same as selectedIndex)", dropdown.highlightedIndex)
	}
}

func TestDropdown_Collapse(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 0)
	dropdown.Expand()
	dropdown.highlightedIndex = 1 // simulate navigation

	dropdown.Collapse()

	if dropdown.expanded {
		t.Error("dropdown should not be expanded after Collapse()")
	}
}

func TestDropdown_IsExpanded(t *testing.T) {
	dropdown := NewDropdown([]DropdownOption{{Label: "A", Value: "a"}}, 0)

	if dropdown.IsExpanded() {
		t.Error("IsExpanded should return false initially")
	}

	dropdown.Expand()
	if !dropdown.IsExpanded() {
		t.Error("IsExpanded should return true after Expand")
	}

	dropdown.Collapse()
	if dropdown.IsExpanded() {
		t.Error("IsExpanded should return false after Collapse")
	}
}

func TestDropdown_SelectedValue(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 1)

	if dropdown.SelectedValue() != "size" {
		t.Errorf("SelectedValue = %q, want %q", dropdown.SelectedValue(), "size")
	}
}

func TestDropdown_SelectedLabel(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 1)

	if dropdown.SelectedLabel() != "Size" {
		t.Errorf("SelectedLabel = %q, want %q", dropdown.SelectedLabel(), "Size")
	}
}

func TestDropdown_SetSelectedIndex(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	dropdown := NewDropdown(options, 0)

	dropdown.SetSelectedIndex(2)
	if dropdown.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2", dropdown.selectedIndex)
	}

	// Boundary: negative
	dropdown.SetSelectedIndex(-1)
	if dropdown.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 (clamped)", dropdown.selectedIndex)
	}

	// Boundary: too large
	dropdown.SetSelectedIndex(10)
	if dropdown.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2 (clamped)", dropdown.selectedIndex)
	}
}

func TestDropdown_HandleKey_CursorNavigation(t *testing.T) {
	tests := []struct {
		name             string
		startHighlighted int
		key              string
		wantHighlighted  int
		wantAction       DropdownAction
	}{
		{"j from 0 to 1", 0, "j", 1, DropdownActionNone},
		{"down from 0 to 1", 0, "down", 1, DropdownActionNone},
		{"k from 1 to 0", 1, "k", 0, DropdownActionNone},
		{"up from 1 to 0", 1, "up", 0, DropdownActionNone},
		{"j from last stays at last", 2, "j", 2, DropdownActionNone},
		{"k from first stays at first", 0, "k", 0, DropdownActionNone},
		{"down from last stays at last", 2, "down", 2, DropdownActionNone},
		{"up from first stays at first", 0, "up", 0, DropdownActionNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := []DropdownOption{
				{Label: "Name", Value: "name"},
				{Label: "Size", Value: "size"},
				{Label: "Date", Value: "date"},
			}
			dropdown := NewDropdown(options, 0)
			dropdown.Expand()
			dropdown.highlightedIndex = tt.startHighlighted

			action := dropdown.HandleKey(tt.key)

			if dropdown.highlightedIndex != tt.wantHighlighted {
				t.Errorf("highlightedIndex = %d, want %d", dropdown.highlightedIndex, tt.wantHighlighted)
			}
			if action != tt.wantAction {
				t.Errorf("action = %v, want %v", action, tt.wantAction)
			}
		})
	}
}

func TestDropdown_HandleKey_Enter(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	dropdown := NewDropdown(options, 0)
	dropdown.Expand()
	dropdown.highlightedIndex = 2 // highlight Date

	action := dropdown.HandleKey("enter")

	if action != DropdownActionSelected {
		t.Errorf("action = %v, want DropdownActionSelected", action)
	}
	if dropdown.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2", dropdown.selectedIndex)
	}
	if dropdown.expanded {
		t.Error("dropdown should collapse after selection")
	}
}

func TestDropdown_HandleKey_Escape(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 0)
	dropdown.Expand()
	dropdown.highlightedIndex = 1 // highlight Size

	action := dropdown.HandleKey("esc")

	if action != DropdownActionCancelled {
		t.Errorf("action = %v, want DropdownActionCancelled", action)
	}
	if dropdown.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 (unchanged)", dropdown.selectedIndex)
	}
	if dropdown.expanded {
		t.Error("dropdown should collapse after cancel")
	}
}

func TestDropdown_HandleKey_UnknownKey(t *testing.T) {
	dropdown := NewDropdown([]DropdownOption{{Label: "A", Value: "a"}}, 0)
	dropdown.Expand()

	action := dropdown.HandleKey("x")

	if action != DropdownActionNone {
		t.Errorf("action = %v, want DropdownActionNone", action)
	}
}

func TestDropdown_HandleKey_NotExpanded(t *testing.T) {
	dropdown := NewDropdown([]DropdownOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}, 0)

	// Keys should not work when dropdown is closed
	action := dropdown.HandleKey("j")
	if action != DropdownActionNone {
		t.Errorf("action = %v, want DropdownActionNone (dropdown closed)", action)
	}
	if dropdown.highlightedIndex != 0 {
		t.Errorf("highlightedIndex = %d, want 0 (unchanged)", dropdown.highlightedIndex)
	}
}

func TestDropdown_View_Closed(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
	}
	dropdown := NewDropdown(options, 0)

	view := dropdown.View(false) // not focused

	// Should contain selected value and down arrow indicator
	if !strings.Contains(view, "Name") {
		t.Errorf("View should contain selected label 'Name', got: %q", view)
	}
	if !strings.Contains(view, "\u25BC") { // ▼
		t.Errorf("View should contain down arrow indicator, got: %q", view)
	}
}

func TestDropdown_View_Expanded(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	dropdown := NewDropdown(options, 0)
	dropdown.Expand()

	view := dropdown.View(true) // focused

	// Should contain all options
	if !strings.Contains(view, "Name") {
		t.Errorf("View should contain 'Name', got: %q", view)
	}
	if !strings.Contains(view, "Size") {
		t.Errorf("View should contain 'Size', got: %q", view)
	}
	if !strings.Contains(view, "Date") {
		t.Errorf("View should contain 'Date', got: %q", view)
	}
}

func TestDropdown_View_ExpandedHeight(t *testing.T) {
	options := []DropdownOption{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Date", Value: "date"},
	}
	dropdown := NewDropdown(options, 0)

	// Both expanded and collapsed should render something
	closedView := dropdown.View(false)
	dropdown.Expand()
	expandedView := dropdown.View(true)

	if len(closedView) == 0 {
		t.Error("closed view should not be empty")
	}
	if len(expandedView) == 0 {
		t.Error("expanded view should not be empty")
	}
}

func TestDropdown_OptionCount(t *testing.T) {
	options := []DropdownOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	}
	dropdown := NewDropdown(options, 0)

	if dropdown.OptionCount() != 3 {
		t.Errorf("OptionCount = %d, want 3", dropdown.OptionCount())
	}
}

func TestDropdownAction_String(t *testing.T) {
	tests := []struct {
		action DropdownAction
		want   string
	}{
		{DropdownActionNone, "None"},
		{DropdownActionSelected, "Selected"},
		{DropdownActionCancelled, "Cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.action.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewDropdown_EmptyOptions(t *testing.T) {
	dropdown := NewDropdown([]DropdownOption{}, 0)

	if dropdown.SelectedValue() != "" {
		t.Errorf("SelectedValue with empty options should return empty string, got: %q", dropdown.SelectedValue())
	}
	if dropdown.SelectedLabel() != "" {
		t.Errorf("SelectedLabel with empty options should return empty string, got: %q", dropdown.SelectedLabel())
	}
	if dropdown.OptionCount() != 0 {
		t.Errorf("OptionCount with empty options should return 0, got: %d", dropdown.OptionCount())
	}
}

func TestNewDropdown_IndexClamping(t *testing.T) {
	options := []DropdownOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}

	// Test negative index clamping
	dropdown := NewDropdown(options, -5)
	if dropdown.SelectedIndex() != 0 {
		t.Errorf("negative index should be clamped to 0, got: %d", dropdown.SelectedIndex())
	}

	// Test over-range index clamping
	dropdown = NewDropdown(options, 10)
	if dropdown.SelectedIndex() != 1 {
		t.Errorf("over-range index should be clamped to last option, got: %d", dropdown.SelectedIndex())
	}
}
