package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPane(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		width    int
		height   int
		isActive bool
		wantErr  bool
	}{
		{
			name:     "有効なディレクトリでペインを作成",
			path:     tmpDir,
			width:    40,
			height:   20,
			isActive: true,
			wantErr:  false,
		},
		{
			name:     "非アクティブなペインを作成",
			path:     tmpDir,
			width:    40,
			height:   20,
			isActive: false,
			wantErr:  false,
		},
		{
			name:     "存在しないディレクトリでエラー",
			path:     filepath.Join(tmpDir, "nonexistent"),
			width:    40,
			height:   20,
			isActive: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane, err := NewPane(tt.path, tt.width, tt.height, tt.isActive)

			if tt.wantErr {
				if err == nil {
					t.Error("NewPane() should return error for invalid path")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewPane() error = %v, wantErr %v", err, tt.wantErr)
			}

			if pane.cursor != 0 {
				t.Errorf("NewPane() cursor = %d, want 0", pane.cursor)
			}

			if pane.width != tt.width {
				t.Errorf("NewPane() width = %d, want %d", pane.width, tt.width)
			}

			if pane.height != tt.height {
				t.Errorf("NewPane() height = %d, want %d", pane.height, tt.height)
			}

			if pane.isActive != tt.isActive {
				t.Errorf("NewPane() isActive = %v, want %v", pane.isActive, tt.isActive)
			}

			if len(pane.entries) == 0 {
				t.Error("NewPane() should load directory entries")
			}
		})
	}
}

func TestPaneMoveCursor(t *testing.T) {
	tmpDir := t.TempDir()

	// テストファイルを作成
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	initialCursor := pane.cursor

	t.Run("カーソルを下に移動", func(t *testing.T) {
		pane.MoveCursorDown()
		if pane.cursor <= initialCursor {
			t.Error("MoveCursorDown() should increase cursor")
		}
	})

	t.Run("カーソルを上に移動", func(t *testing.T) {
		currentCursor := pane.cursor
		pane.MoveCursorUp()
		if pane.cursor >= currentCursor {
			t.Error("MoveCursorUp() should decrease cursor")
		}
	})

	t.Run("カーソルが上限を超えない", func(t *testing.T) {
		// カーソルを最上部に移動
		for i := 0; i < 100; i++ {
			pane.MoveCursorUp()
		}
		if pane.cursor != 0 {
			t.Errorf("Cursor should not go below 0, got %d", pane.cursor)
		}
	})

	t.Run("カーソルが下限を超えない", func(t *testing.T) {
		maxCursor := len(pane.entries) - 1
		// カーソルを最下部に移動
		for i := 0; i < 100; i++ {
			pane.MoveCursorDown()
		}
		if pane.cursor > maxCursor {
			t.Errorf("Cursor should not exceed %d, got %d", maxCursor, pane.cursor)
		}
	})
}

func TestPaneSelectedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("選択中のエントリを取得", func(t *testing.T) {
		entry := pane.SelectedEntry()
		if entry == nil {
			t.Error("SelectedEntry() should return non-nil entry")
		}
	})

	t.Run("無効なカーソル位置ではnilを返す", func(t *testing.T) {
		pane.cursor = -1
		entry := pane.SelectedEntry()
		if entry != nil {
			t.Error("SelectedEntry() should return nil for invalid cursor")
		}

		pane.cursor = len(pane.entries) + 10
		entry = pane.SelectedEntry()
		if entry != nil {
			t.Error("SelectedEntry() should return nil for out of bounds cursor")
		}
	})
}

func TestPaneEnterDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// サブディレクトリとファイルを作成
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("ディレクトリに入る", func(t *testing.T) {
		// サブディレクトリを見つける
		for i, entry := range pane.entries {
			if entry.Name == "subdir" && entry.IsDir {
				pane.cursor = i
				break
			}
		}

		initialPath := pane.Path()
		err := pane.EnterDirectory()
		if err != nil {
			t.Errorf("EnterDirectory() error = %v", err)
		}

		if pane.Path() == initialPath {
			t.Error("EnterDirectory() should change path")
		}
	})

	t.Run("ファイルでは何もしない", func(t *testing.T) {
		// 親ディレクトリに戻る
		pane.path = tmpDir
		pane.LoadDirectory()

		// ファイルを選択
		for i, entry := range pane.entries {
			if entry.Name == "file.txt" && !entry.IsDir {
				pane.cursor = i
				break
			}
		}

		initialPath := pane.Path()
		err := pane.EnterDirectory()
		if err != nil {
			t.Errorf("EnterDirectory() error = %v", err)
		}

		if pane.Path() != initialPath {
			t.Error("EnterDirectory() should not change path for files")
		}
	})
}

func TestPaneChangeDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("指定したディレクトリに移動", func(t *testing.T) {
		err := pane.ChangeDirectory(subDir)
		if err != nil {
			t.Errorf("ChangeDirectory() error = %v", err)
		}

		if pane.Path() != subDir {
			t.Errorf("ChangeDirectory() path = %s, want %s", pane.Path(), subDir)
		}
	})
}

func TestPaneMoveToParent(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(subDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("親ディレクトリに移動", func(t *testing.T) {
		initialPath := pane.Path()
		err := pane.MoveToParent()
		if err != nil {
			t.Errorf("MoveToParent() error = %v", err)
		}

		if pane.Path() == initialPath {
			t.Error("MoveToParent() should change path")
		}

		if !filepath.IsAbs(pane.Path()) {
			t.Error("MoveToParent() should maintain absolute path")
		}
	})

	t.Run("ルートディレクトリでは移動しない", func(t *testing.T) {
		pane.path = "/"
		pane.LoadDirectory()

		err := pane.MoveToParent()
		if err != nil {
			t.Errorf("MoveToParent() error = %v", err)
		}

		if pane.Path() != "/" {
			t.Error("MoveToParent() should not move above root")
		}
	})
}

func TestPaneSetSize(t *testing.T) {
	tmpDir := t.TempDir()
	pane, _ := NewPane(tmpDir, 40, 20, true)

	newWidth := 80
	newHeight := 40

	pane.SetSize(newWidth, newHeight)

	if pane.width != newWidth {
		t.Errorf("SetSize() width = %d, want %d", pane.width, newWidth)
	}

	if pane.height != newHeight {
		t.Errorf("SetSize() height = %d, want %d", pane.height, newHeight)
	}
}

func TestPaneSetActive(t *testing.T) {
	tmpDir := t.TempDir()
	pane, _ := NewPane(tmpDir, 40, 20, false)

	if pane.isActive {
		t.Error("Pane should be inactive initially")
	}

	pane.SetActive(true)
	if !pane.isActive {
		t.Error("SetActive(true) should make pane active")
	}

	pane.SetActive(false)
	if pane.isActive {
		t.Error("SetActive(false) should make pane inactive")
	}
}

func TestPaneView(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	view := pane.View()

	if view == "" {
		t.Error("View() should return non-empty string")
	}

	// ビューにパス情報が含まれているか確認
	// 少なくとも何かしらのコンテンツがあることを確認
	if len(view) < 10 {
		t.Error("View() should return substantial content")
	}
}

func TestPaneLoadDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	initialEntryCount := len(pane.entries)

	// 新しいファイルを追加
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	// ディレクトリを再読み込み
	err = pane.LoadDirectory()
	if err != nil {
		t.Errorf("LoadDirectory() error = %v", err)
	}

	if len(pane.entries) <= initialEntryCount {
		t.Error("LoadDirectory() should reflect new files")
	}

	// カーソルとスクロールオフセットがリセットされることを確認
	if pane.cursor != 0 {
		t.Error("LoadDirectory() should reset cursor to 0")
	}

	if pane.scrollOffset != 0 {
		t.Error("LoadDirectory() should reset scrollOffset to 0")
	}
}
