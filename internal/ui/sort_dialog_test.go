package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSortDialog(t *testing.T) {
	current := SortConfig{Field: SortBySize, Order: SortDesc}
	dialog := NewSortDialog(current)

	if dialog.config.Field != SortBySize {
		t.Errorf("config.Field = %v, want SortBySize", dialog.config.Field)
	}
	if dialog.config.Order != SortDesc {
		t.Errorf("config.Order = %v, want SortDesc", dialog.config.Order)
	}
	if dialog.originalConfig.Field != SortBySize {
		t.Errorf("originalConfig.Field = %v, want SortBySize", dialog.originalConfig.Field)
	}
	if dialog.focusedItem != FocusTargetSortBy {
		t.Errorf("focusedItem = %d, want FocusTargetSortBy", dialog.focusedItem)
	}
	if !dialog.active {
		t.Error("dialog should be active")
	}
}

func TestSortDialog_HandleKey_TabNavigation(t *testing.T) {
	tests := []struct {
		name         string
		startFocused FocusTarget
		key          string
		wantFocused  FocusTarget
	}{
		{"Tab from SortBy to Order", FocusTargetSortBy, "tab", FocusTargetOrder},
		{"Tab from Order to OK", FocusTargetOrder, "tab", FocusTargetOK},
		{"Tab from OK stays at OK", FocusTargetOK, "tab", FocusTargetOK},
		{"Shift+Tab from OK to Order", FocusTargetOK, "shift+tab", FocusTargetOrder},
		{"Shift+Tab from Order to SortBy", FocusTargetOrder, "shift+tab", FocusTargetSortBy},
		{"Shift+Tab from SortBy stays at SortBy", FocusTargetSortBy, "shift+tab", FocusTargetSortBy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedItem = tt.startFocused

			dialog.HandleKey(tt.key)

			if dialog.focusedItem != tt.wantFocused {
				t.Errorf("focusedItem = %d, want %d", dialog.focusedItem, tt.wantFocused)
			}
		})
	}
}

func TestSortDialog_HandleKey_EnterExpandsDropdown(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	// Enter should expand the focused dropdown
	dialog.HandleKey("enter")

	if !dialog.fieldDropdown.IsExpanded() {
		t.Error("Enter should expand the focused dropdown")
	}
}

func TestSortDialog_HandleKey_SpaceExpandsDropdown(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	// Space should also expand the focused dropdown
	dialog.HandleKey(" ")

	if !dialog.fieldDropdown.IsExpanded() {
		t.Error("Space should expand the focused dropdown")
	}
}

func TestSortDialog_HandleKey_EnterExpandsDropdownFirst(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})

	// Enter on fresh dialog should expand the dropdown, not confirm
	confirmed, cancelled := dialog.HandleKey("enter")

	if confirmed || cancelled {
		t.Error("Enter should expand dropdown, not confirm/cancel")
	}
	if !dialog.fieldDropdown.IsExpanded() {
		t.Error("Enter should expand focused dropdown")
	}
}

func TestSortDialog_HandleKey_EscapeCancelsDialog(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)

	// Change the selection
	dialog.config.Field = SortBySize
	dialog.config.Order = SortDesc

	confirmed, cancelled := dialog.HandleKey("esc")

	if confirmed {
		t.Error("Escape should not confirm")
	}
	if !cancelled {
		t.Error("Escape should cancel dialog when no dropdown is expanded")
	}
	// Config should be restored to original
	if dialog.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName (restored)", dialog.config.Field)
	}
	if dialog.config.Order != SortAsc {
		t.Errorf("config.Order = %v, want SortAsc (restored)", dialog.config.Order)
	}
}

func TestSortDialog_HandleKey_EscapeClosesExpandedDropdown(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	// Expand the dropdown
	dialog.fieldDropdown.Expand()

	// Escape should close dropdown, not cancel dialog
	confirmed, cancelled := dialog.HandleKey("esc")

	if dialog.fieldDropdown.IsExpanded() {
		t.Error("Escape should close expanded dropdown")
	}
	if confirmed || cancelled {
		t.Error("Escape with expanded dropdown should not confirm or cancel dialog")
	}
}

func TestSortDialog_HandleKey_QCancelsDialogEvenWhenDropdownExpanded(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)

	// Expand the dropdown
	dialog.fieldDropdown.Expand()

	confirmed, cancelled := dialog.HandleKey("q")

	if confirmed {
		t.Error("q should not confirm")
	}
	if !cancelled {
		t.Error("q should cancel dialog even when dropdown is expanded")
	}
}

func TestSortDialog_HandleKey_DropdownNavigation(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})

	// Expand field dropdown
	dialog.HandleKey("enter")

	// Navigate down (j)
	dialog.HandleKey("j")

	// highlightedIndex should change
	if dialog.fieldDropdown.highlightedIndex != 1 {
		t.Errorf("highlightedIndex = %d, want 1", dialog.fieldDropdown.highlightedIndex)
	}
}

func TestSortDialog_HandleKey_DropdownSelection(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})

	// Expand field dropdown
	dialog.HandleKey("enter")

	// Navigate to Size (index 1)
	dialog.HandleKey("j")

	// Select with Enter
	confirmed, cancelled := dialog.HandleKey("enter")

	if dialog.config.Field != SortBySize {
		t.Errorf("config.Field = %v, want SortBySize", dialog.config.Field)
	}
	if dialog.fieldDropdown.IsExpanded() {
		t.Error("dropdown should collapse after selection")
	}
	// Selecting from dropdown should not confirm/cancel the dialog
	if confirmed || cancelled {
		t.Error("dropdown selection should not confirm or cancel dialog")
	}
}

func TestSortDialog_HandleKey_OrderDropdownSelection(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})

	// Focus on Order dropdown
	dialog.HandleKey("tab")

	// Expand order dropdown
	dialog.HandleKey("enter")

	// Navigate to Desc (index 1)
	dialog.HandleKey("j")

	// Select with Enter
	dialog.HandleKey("enter")

	if dialog.config.Order != SortDesc {
		t.Errorf("config.Order = %v, want SortDesc", dialog.config.Order)
	}
}

func TestSortDialog_Config(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByDate, Order: SortDesc})

	config := dialog.Config()

	if config.Field != SortByDate {
		t.Errorf("Config().Field = %v, want SortByDate", config.Field)
	}
	if config.Order != SortDesc {
		t.Errorf("Config().Order = %v, want SortDesc", config.Order)
	}
}

func TestSortDialog_OriginalConfig(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortAsc})
	// Change config but originalConfig should stay the same
	dialog.config.Field = SortByDate
	dialog.config.Order = SortDesc

	original := dialog.OriginalConfig()

	if original.Field != SortBySize {
		t.Errorf("OriginalConfig().Field = %v, want SortBySize", original.Field)
	}
	if original.Order != SortAsc {
		t.Errorf("OriginalConfig().Order = %v, want SortAsc", original.Order)
	}
}

func TestSortDialog_IsActive(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	if !dialog.IsActive() {
		t.Error("New dialog should be active")
	}

	// Escape to cancel dialog
	dialog.HandleKey("esc")
	if dialog.IsActive() {
		t.Error("After escape, dialog should be inactive")
	}
}

func TestSortDialog_View(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})
	dialog.width = 40

	view := dialog.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestSortDialog_DisplayType(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	if dialog.DisplayType() != DialogDisplayPane {
		t.Errorf("DisplayType() = %v, want DialogDisplayPane", dialog.DisplayType())
	}
}

func TestSortDialog_Update_EnterExpandsDropdown(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := dialog.Update(keyMsg)

	// Enter should expand the dropdown
	if !dialog.fieldDropdown.IsExpanded() {
		t.Error("Enter should expand the focused dropdown")
	}

	// Command should be for live preview (config changed message)
	if cmd == nil {
		t.Error("Update with Enter should return a command")
	}

	msg := cmd()
	_, ok := msg.(sortDialogConfigChangedMsg)
	if !ok {
		t.Fatalf("Expected sortDialogConfigChangedMsg, got %T", msg)
	}
}

func TestSortDialog_Update_Escape(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)
	dialog.config.Field = SortByDate // Change

	keyMsg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := dialog.Update(keyMsg)

	if cmd == nil {
		t.Error("Update with Escape should return a command")
	}

	msg := cmd()
	result, ok := msg.(sortDialogResultMsg)
	if !ok {
		t.Fatalf("Expected sortDialogResultMsg, got %T", msg)
	}

	if result.confirmed {
		t.Error("Expected confirmed = false")
	}
	if !result.cancelled {
		t.Error("Expected cancelled = true")
	}
	if result.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName", result.config.Field)
	}
}

func TestSortDialog_Update_Navigation(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})

	// Tab to move focus
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := dialog.Update(keyMsg)

	if cmd == nil {
		t.Error("Update with Tab should return a command for live preview")
	}

	msg := cmd()
	_, ok := msg.(sortDialogConfigChangedMsg)
	if !ok {
		t.Fatalf("Expected sortDialogConfigChangedMsg, got %T", msg)
	}
}

func TestSortDialog_Update_Inactive(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())
	dialog.active = false

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := dialog.Update(keyMsg)

	if cmd != nil {
		t.Error("Update on inactive dialog should return nil command")
	}
}

func TestSortDialog_Update_NonKeyMsg(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	_, cmd := dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if cmd != nil {
		t.Error("Update with non-key message should return nil command")
	}
}

func TestSortDialog_DropdownSyncWithConfig(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByDate, Order: SortDesc})

	// Dropdown should be initialized with correct selection
	if dialog.fieldDropdown.SelectedIndex() != int(SortByDate) {
		t.Errorf("fieldDropdown.SelectedIndex = %d, want %d", dialog.fieldDropdown.SelectedIndex(), SortByDate)
	}
	if dialog.orderDropdown.SelectedIndex() != int(SortDesc) {
		t.Errorf("orderDropdown.SelectedIndex = %d, want %d", dialog.orderDropdown.SelectedIndex(), SortDesc)
	}
}

func TestSortDialog_ArrowKeysInDropdown(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())

	// Expand dropdown
	dialog.HandleKey("enter")

	// Down arrow should work
	dialog.HandleKey("down")
	if dialog.fieldDropdown.highlightedIndex != 1 {
		t.Errorf("down arrow: highlightedIndex = %d, want 1", dialog.fieldDropdown.highlightedIndex)
	}

	// Up arrow should work
	dialog.HandleKey("up")
	if dialog.fieldDropdown.highlightedIndex != 0 {
		t.Errorf("up arrow: highlightedIndex = %d, want 0", dialog.fieldDropdown.highlightedIndex)
	}
}

// ============================================
// Phase 4: j/k Navigation and OK Button Tests
// ============================================

func TestSortDialog_JKNavigation_BetweenMajorItems(t *testing.T) {
	tests := []struct {
		name       string
		startFocus FocusTarget
		key        string
		wantFocus  FocusTarget
	}{
		{"j from SortBy to Order", FocusTargetSortBy, "j", FocusTargetOrder},
		{"j from Order to OK", FocusTargetOrder, "j", FocusTargetOK},
		{"j from OK stays at OK", FocusTargetOK, "j", FocusTargetOK},
		{"k from OK to Order", FocusTargetOK, "k", FocusTargetOrder},
		{"k from Order to SortBy", FocusTargetOrder, "k", FocusTargetSortBy},
		{"k from SortBy stays at SortBy", FocusTargetSortBy, "k", FocusTargetSortBy},
		{"down from SortBy to Order", FocusTargetSortBy, "down", FocusTargetOrder},
		{"down from Order to OK", FocusTargetOrder, "down", FocusTargetOK},
		{"up from OK to Order", FocusTargetOK, "up", FocusTargetOrder},
		{"up from Order to SortBy", FocusTargetOrder, "up", FocusTargetSortBy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedItem = tt.startFocus

			dialog.HandleKey(tt.key)

			if dialog.focusedItem != tt.wantFocus {
				t.Errorf("focusedItem = %d, want %d", dialog.focusedItem, tt.wantFocus)
			}
		})
	}
}

func TestSortDialog_TabNavigation_ThreeMajorItems(t *testing.T) {
	tests := []struct {
		name       string
		startFocus FocusTarget
		key        string
		wantFocus  FocusTarget
	}{
		{"Tab from SortBy to Order", FocusTargetSortBy, "tab", FocusTargetOrder},
		{"Tab from Order to OK", FocusTargetOrder, "tab", FocusTargetOK},
		{"Tab from OK stays at OK", FocusTargetOK, "tab", FocusTargetOK},
		{"Shift+Tab from OK to Order", FocusTargetOK, "shift+tab", FocusTargetOrder},
		{"Shift+Tab from Order to SortBy", FocusTargetOrder, "shift+tab", FocusTargetSortBy},
		{"Shift+Tab from SortBy stays at SortBy", FocusTargetSortBy, "shift+tab", FocusTargetSortBy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedItem = tt.startFocus

			dialog.HandleKey(tt.key)

			if dialog.focusedItem != tt.wantFocus {
				t.Errorf("focusedItem = %d, want %d", dialog.focusedItem, tt.wantFocus)
			}
		})
	}
}

func TestSortDialog_OKButton_EnterConfirms(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})
	dialog.focusedItem = FocusTargetOK

	confirmed, cancelled := dialog.HandleKey("enter")

	if !confirmed {
		t.Error("Enter on OK button should confirm dialog")
	}
	if cancelled {
		t.Error("Enter on OK button should not cancel")
	}
	if dialog.IsActive() {
		t.Error("Dialog should be closed after confirmation")
	}
}

func TestSortDialog_OKButton_SpaceDoesNothing(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())
	dialog.focusedItem = FocusTargetOK

	confirmed, cancelled := dialog.HandleKey(" ")

	if confirmed {
		t.Error("Space on OK button should not confirm")
	}
	if cancelled {
		t.Error("Space on OK button should not cancel")
	}
	if !dialog.IsActive() {
		t.Error("Dialog should remain active after Space on OK button")
	}
}

func TestSortDialog_JKNavigation_OnlyWhenDropdownsClosed(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())
	dialog.focusedItem = FocusTargetSortBy

	// Expand the dropdown
	dialog.fieldDropdown.Expand()

	// j should not move focus when dropdown is expanded (dropdown handles it)
	initialFocus := dialog.focusedItem
	dialog.HandleKey("j")

	// Focus should not change because dropdown handles j
	if dialog.focusedItem != initialFocus {
		t.Errorf("j with expanded dropdown changed focusedItem from %d to %d", initialFocus, dialog.focusedItem)
	}
}

func TestSortDialog_EnterExpandsDropdown_NotOK(t *testing.T) {
	tests := []struct {
		name   string
		focus  FocusTarget
		expand bool
	}{
		{"Enter on SortBy expands dropdown", FocusTargetSortBy, true},
		{"Enter on Order expands dropdown", FocusTargetOrder, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedItem = tt.focus

			confirmed, cancelled := dialog.HandleKey("enter")

			if confirmed || cancelled {
				t.Error("Enter on dropdown should not confirm or cancel dialog")
			}

			// Check if the correct dropdown is expanded
			if tt.focus == FocusTargetSortBy && !dialog.fieldDropdown.IsExpanded() {
				t.Error("Sort by dropdown should be expanded")
			}
			if tt.focus == FocusTargetOrder && !dialog.orderDropdown.IsExpanded() {
				t.Error("Order dropdown should be expanded")
			}
		})
	}
}

func TestSortDialog_SpaceExpandsDropdown_NotOK(t *testing.T) {
	tests := []struct {
		name   string
		focus  FocusTarget
		expand bool
	}{
		{"Space on SortBy expands dropdown", FocusTargetSortBy, true},
		{"Space on Order expands dropdown", FocusTargetOrder, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedItem = tt.focus

			dialog.HandleKey(" ")

			// Check if the correct dropdown is expanded
			if tt.focus == FocusTargetSortBy && !dialog.fieldDropdown.IsExpanded() {
				t.Error("Sort by dropdown should be expanded")
			}
			if tt.focus == FocusTargetOrder && !dialog.orderDropdown.IsExpanded() {
				t.Error("Order dropdown should be expanded")
			}
		})
	}
}

func TestSortDialog_View_ShowsOKButton(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())
	dialog.width = 40

	view := dialog.View()

	// View should contain OK button
	if !strings.Contains(view, "OK") {
		t.Error("View should contain OK button")
	}
}

func TestSortDialog_View_HelpTextUpdated(t *testing.T) {
	dialog := NewSortDialog(DefaultSortConfig())
	dialog.width = 40

	view := dialog.View()

	// View should contain updated help text with j/k
	if !strings.Contains(view, "j/k") {
		t.Error("Help text should mention j/k for navigation")
	}
}

func TestSortDialog_Update_OKConfirmation(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})
	dialog.focusedItem = FocusTargetOK

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := dialog.Update(keyMsg)

	if cmd == nil {
		t.Error("Update with Enter on OK should return a command")
	}

	msg := cmd()
	result, ok := msg.(sortDialogResultMsg)
	if !ok {
		t.Fatalf("Expected sortDialogResultMsg, got %T", msg)
	}

	if !result.confirmed {
		t.Error("Expected confirmed = true")
	}
	if result.cancelled {
		t.Error("Expected cancelled = false")
	}
	if result.config.Field != SortBySize {
		t.Errorf("config.Field = %v, want SortBySize", result.config.Field)
	}
}
