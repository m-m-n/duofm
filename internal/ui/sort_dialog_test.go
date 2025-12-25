package ui

import "testing"

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
	if dialog.focusedRow != 0 {
		t.Errorf("focusedRow = %d, want 0", dialog.focusedRow)
	}
	if !dialog.active {
		t.Error("dialog should be active")
	}
}

func TestSortDialog_HandleKey_FieldNavigation(t *testing.T) {
	tests := []struct {
		name       string
		startField SortField
		key        string
		wantField  SortField
		focusedRow int
	}{
		{"h from Size to Name", SortBySize, "h", SortByName, 0},
		{"left from Size to Name", SortBySize, "left", SortByName, 0},
		{"l from Size to Date", SortBySize, "l", SortByDate, 0},
		{"right from Size to Date", SortBySize, "right", SortByDate, 0},
		{"h from Name stays at Name", SortByName, "h", SortByName, 0},
		{"l from Date stays at Date", SortByDate, "l", SortByDate, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(SortConfig{Field: tt.startField, Order: SortAsc})
			dialog.focusedRow = tt.focusedRow

			dialog.HandleKey(tt.key)

			if dialog.config.Field != tt.wantField {
				t.Errorf("Field = %v, want %v", dialog.config.Field, tt.wantField)
			}
		})
	}
}

func TestSortDialog_HandleKey_OrderNavigation(t *testing.T) {
	tests := []struct {
		name       string
		startOrder SortOrder
		key        string
		wantOrder  SortOrder
	}{
		{"h from Desc to Asc", SortDesc, "h", SortAsc},
		{"left from Desc to Asc", SortDesc, "left", SortAsc},
		{"l from Asc to Desc", SortAsc, "l", SortDesc},
		{"right from Asc to Desc", SortAsc, "right", SortDesc},
		{"h from Asc stays at Asc", SortAsc, "h", SortAsc},
		{"l from Desc stays at Desc", SortDesc, "l", SortDesc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(SortConfig{Field: SortByName, Order: tt.startOrder})
			dialog.focusedRow = 1 // Order行

			dialog.HandleKey(tt.key)

			if dialog.config.Order != tt.wantOrder {
				t.Errorf("Order = %v, want %v", dialog.config.Order, tt.wantOrder)
			}
		})
	}
}

func TestSortDialog_HandleKey_RowNavigation(t *testing.T) {
	tests := []struct {
		name  string
		start int
		key   string
		want  int
	}{
		{"j from row 0 to 1", 0, "j", 1},
		{"down from row 0 to 1", 0, "down", 1},
		{"k from row 1 to 0", 1, "k", 0},
		{"up from row 1 to 0", 1, "up", 0},
		{"j from row 1 stays at 1", 1, "j", 1},
		{"k from row 0 stays at 0", 0, "k", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog := NewSortDialog(DefaultSortConfig())
			dialog.focusedRow = tt.start

			dialog.HandleKey(tt.key)

			if dialog.focusedRow != tt.want {
				t.Errorf("focusedRow = %d, want %d", dialog.focusedRow, tt.want)
			}
		})
	}
}

func TestSortDialog_HandleKey_Enter(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortBySize, Order: SortDesc})

	confirmed, cancelled := dialog.HandleKey("enter")

	if !confirmed {
		t.Error("Expected confirmed = true")
	}
	if cancelled {
		t.Error("Expected cancelled = false")
	}
}

func TestSortDialog_HandleKey_Escape(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)

	// 変更
	dialog.config.Field = SortBySize
	dialog.config.Order = SortDesc

	confirmed, cancelled := dialog.HandleKey("esc")

	if confirmed {
		t.Error("Expected confirmed = false")
	}
	if !cancelled {
		t.Error("Expected cancelled = true")
	}
	// configがoriginalに復元されることを確認
	if dialog.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName", dialog.config.Field)
	}
	if dialog.config.Order != SortAsc {
		t.Errorf("config.Order = %v, want SortAsc", dialog.config.Order)
	}
}

func TestSortDialog_HandleKey_Q(t *testing.T) {
	original := SortConfig{Field: SortByName, Order: SortAsc}
	dialog := NewSortDialog(original)
	dialog.config.Field = SortByDate

	confirmed, cancelled := dialog.HandleKey("q")

	if confirmed {
		t.Error("Expected confirmed = false")
	}
	if !cancelled {
		t.Error("Expected cancelled = true")
	}
	// Escと同じ動作
	if dialog.config.Field != SortByName {
		t.Errorf("config.Field = %v, want SortByName", dialog.config.Field)
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
	// 変更しても originalConfig は変わらない
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

	// Enterで終了
	dialog.HandleKey("enter")
	if dialog.IsActive() {
		t.Error("After enter, dialog should be inactive")
	}
}

func TestSortDialog_View(t *testing.T) {
	dialog := NewSortDialog(SortConfig{Field: SortByName, Order: SortAsc})
	dialog.width = 40

	view := dialog.View()

	// 基本的な内容が含まれることを確認
	if view == "" {
		t.Error("View should not be empty")
	}
}
