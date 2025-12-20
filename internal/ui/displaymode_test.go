package ui

import (
	"testing"
)

func TestDisplayModeConstants(t *testing.T) {
	// DisplayMode定数が定義されていることを確認
	modes := []DisplayMode{DisplayMinimal, DisplayBasic, DisplayDetail}

	if len(modes) != 3 {
		t.Errorf("Expected 3 display modes, got %d", len(modes))
	}

	// 各モードが異なる値を持つことを確認
	if DisplayMinimal == DisplayBasic || DisplayBasic == DisplayDetail || DisplayMinimal == DisplayDetail {
		t.Error("Display modes should have unique values")
	}
}

func TestPaneToggleDisplayMode(t *testing.T) {
	// テスト用のペインを作成
	pane, err := NewPane("/tmp", 100, 20, true)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	// 初期状態はBasicモード
	if pane.displayMode != DisplayBasic {
		t.Errorf("Initial display mode = %v, want %v", pane.displayMode, DisplayBasic)
	}

	// BasicからDetailに切り替え
	pane.ToggleDisplayMode()
	if pane.displayMode != DisplayDetail {
		t.Errorf("After toggle, display mode = %v, want %v", pane.displayMode, DisplayDetail)
	}

	// DetailからBasicに切り替え
	pane.ToggleDisplayMode()
	if pane.displayMode != DisplayBasic {
		t.Errorf("After second toggle, display mode = %v, want %v", pane.displayMode, DisplayBasic)
	}
}

func TestPaneShouldUseMinimalMode(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		wantBool bool
	}{
		{
			name:     "wide terminal",
			width:    120,
			wantBool: false,
		},
		{
			name:     "medium terminal",
			width:    80,
			wantBool: false,
		},
		{
			name:     "narrow terminal",
			width:    50,
			wantBool: true,
		},
		{
			name:     "very narrow terminal",
			width:    40,
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane, err := NewPane("/tmp", tt.width, 20, true)
			if err != nil {
				t.Fatalf("Failed to create pane: %v", err)
			}

			got := pane.ShouldUseMinimalMode()
			if got != tt.wantBool {
				t.Errorf("ShouldUseMinimalMode() with width %d = %v, want %v", tt.width, got, tt.wantBool)
			}
		})
	}
}

func TestPaneGetEffectiveDisplayMode(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		userMode DisplayMode
		wantMode DisplayMode
	}{
		{
			name:     "wide terminal with Basic mode",
			width:    120,
			userMode: DisplayBasic,
			wantMode: DisplayBasic,
		},
		{
			name:     "wide terminal with Detail mode",
			width:    120,
			userMode: DisplayDetail,
			wantMode: DisplayDetail,
		},
		{
			name:     "narrow terminal forces Minimal",
			width:    50,
			userMode: DisplayBasic,
			wantMode: DisplayMinimal,
		},
		{
			name:     "narrow terminal ignores Detail",
			width:    40,
			userMode: DisplayDetail,
			wantMode: DisplayMinimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane, err := NewPane("/tmp", tt.width, 20, true)
			if err != nil {
				t.Fatalf("Failed to create pane: %v", err)
			}

			pane.displayMode = tt.userMode
			got := pane.GetEffectiveDisplayMode()
			if got != tt.wantMode {
				t.Errorf("GetEffectiveDisplayMode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

func TestPaneDisplayModeIndependence(t *testing.T) {
	// 2つのペインを作成
	leftPane, err := NewPane("/tmp", 100, 20, true)
	if err != nil {
		t.Fatalf("Failed to create left pane: %v", err)
	}

	rightPane, err := NewPane("/tmp", 100, 20, false)
	if err != nil {
		t.Fatalf("Failed to create right pane: %v", err)
	}

	// 左ペインの表示モードを切り替え
	leftPane.ToggleDisplayMode()

	// 右ペインの表示モードは影響を受けない
	if leftPane.displayMode == rightPane.displayMode {
		t.Error("Pane display modes should be independent")
	}

	if leftPane.displayMode != DisplayDetail {
		t.Errorf("Left pane display mode = %v, want %v", leftPane.displayMode, DisplayDetail)
	}

	if rightPane.displayMode != DisplayBasic {
		t.Errorf("Right pane display mode = %v, want %v", rightPane.displayMode, DisplayBasic)
	}
}

func TestPaneSetSizeWithDisplayMode(t *testing.T) {
	pane, err := NewPane("/tmp", 100, 20, true)
	if err != nil {
		t.Fatalf("Failed to create pane: %v", err)
	}

	// 初期状態ではBasicモード
	initialMode := pane.displayMode

	// サイズを狭くする
	pane.SetSize(40, 20)

	// 実効的な表示モードはMinimalになる
	effectiveMode := pane.GetEffectiveDisplayMode()
	if effectiveMode != DisplayMinimal {
		t.Errorf("Effective mode after narrow resize = %v, want %v", effectiveMode, DisplayMinimal)
	}

	// ユーザーモードは保持される
	if pane.displayMode != initialMode {
		t.Errorf("User mode after narrow resize = %v, want %v (should be preserved)", pane.displayMode, initialMode)
	}

	// サイズを広げる
	pane.SetSize(120, 20)

	// 実効的な表示モードは元のユーザーモードに戻る
	effectiveMode = pane.GetEffectiveDisplayMode()
	if effectiveMode != initialMode {
		t.Errorf("Effective mode after wide resize = %v, want %v", effectiveMode, initialMode)
	}
}
