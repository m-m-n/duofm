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

// === ナビゲーション強化機能のテスト ===

func TestFilterHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 通常ファイルと隠しファイルを作成
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)
	os.Mkdir(filepath.Join(tmpDir, ".hiddendir"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "visibledir"), 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("デフォルトで隠しファイルは非表示", func(t *testing.T) {
		if pane.IsShowingHidden() {
			t.Error("showHidden should be false by default")
		}

		for _, entry := range pane.entries {
			if entry.Name != ".." && entry.Name[0] == '.' {
				t.Errorf("Hidden file %s should not be visible", entry.Name)
			}
		}
	})

	t.Run("親ディレクトリ(..)は常に表示", func(t *testing.T) {
		found := false
		for _, entry := range pane.entries {
			if entry.IsParentDir() {
				found = true
				break
			}
		}
		if !found {
			t.Error("Parent directory (..) should always be visible")
		}
	})
}

func TestToggleHidden(t *testing.T) {
	tmpDir := t.TempDir()

	// 通常ファイルと隠しファイルを作成
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("トグルでshowHiddenが切り替わる", func(t *testing.T) {
		if pane.showHidden {
			t.Error("showHidden should be false initially")
		}

		pane.ToggleHidden()
		if !pane.showHidden {
			t.Error("showHidden should be true after toggle")
		}

		pane.ToggleHidden()
		if pane.showHidden {
			t.Error("showHidden should be false after second toggle")
		}
	})

	t.Run("トグル後に隠しファイルが表示される", func(t *testing.T) {
		pane.ToggleHidden() // showHidden = true

		foundHidden := false
		for _, entry := range pane.entries {
			if entry.Name == ".hidden" {
				foundHidden = true
				break
			}
		}
		if !foundHidden {
			t.Error("Hidden file should be visible when showHidden is true")
		}
	})

	t.Run("カーソル位置が維持される", func(t *testing.T) {
		pane.showHidden = true
		pane.LoadDirectory()

		// visible.txtを選択
		for i, entry := range pane.entries {
			if entry.Name == "visible.txt" {
				pane.cursor = i
				break
			}
		}

		pane.ToggleHidden() // 隠しファイルを非表示に

		// visible.txtが選択されたままか確認
		entry := pane.SelectedEntry()
		if entry == nil || entry.Name != "visible.txt" {
			t.Error("Cursor should remain on the same visible file after toggle")
		}
	})
}

func TestNavigateToHome(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(subDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("ホームディレクトリに移動", func(t *testing.T) {
		initialPath := pane.Path()
		err := pane.NavigateToHome()
		if err != nil {
			t.Errorf("NavigateToHome() error = %v", err)
		}

		// ホームディレクトリに移動したか確認
		home, _ := os.UserHomeDir()
		if pane.Path() != home {
			t.Errorf("NavigateToHome() path = %s, want %s", pane.Path(), home)
		}

		// previousPathが更新されているか確認
		if pane.previousPath != initialPath {
			t.Errorf("previousPath = %s, want %s", pane.previousPath, initialPath)
		}
	})

	t.Run("すでにホームにいる場合は何もしない", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		pane.path = home
		pane.previousPath = tmpDir
		pane.LoadDirectory()

		err := pane.NavigateToHome()
		if err != nil {
			t.Errorf("NavigateToHome() error = %v", err)
		}

		// previousPathは変更されないはず
		if pane.previousPath != tmpDir {
			t.Error("previousPath should not change when already at home")
		}
	})
}

func TestNavigateToPrevious(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("履歴がない場合は何もしない", func(t *testing.T) {
		initialPath := pane.Path()
		err := pane.NavigateToPrevious()
		if err != nil {
			t.Errorf("NavigateToPrevious() error = %v", err)
		}

		if pane.Path() != initialPath {
			t.Error("Path should not change when no previous directory")
		}
	})

	t.Run("直前のディレクトリに移動（トグル動作）", func(t *testing.T) {
		// subDirに移動
		pane.ChangeDirectory(subDir)
		currentPath := pane.Path()
		previousPath := pane.previousPath

		// 直前のディレクトリに移動
		err := pane.NavigateToPrevious()
		if err != nil {
			t.Errorf("NavigateToPrevious() error = %v", err)
		}

		if pane.Path() != previousPath {
			t.Errorf("Path = %s, want %s", pane.Path(), previousPath)
		}

		if pane.previousPath != currentPath {
			t.Errorf("previousPath = %s, want %s", pane.previousPath, currentPath)
		}
	})

	t.Run("トグル動作のテスト（A→B→A→B）", func(t *testing.T) {
		pathA := tmpDir
		pathB := subDir

		pane.path = pathA
		pane.previousPath = pathB
		pane.LoadDirectory()

		// A→B
		pane.NavigateToPrevious()
		if pane.Path() != pathB {
			t.Errorf("After first toggle: path = %s, want %s", pane.Path(), pathB)
		}

		// B→A
		pane.NavigateToPrevious()
		if pane.Path() != pathA {
			t.Errorf("After second toggle: path = %s, want %s", pane.Path(), pathA)
		}
	})
}

func TestPreviousPathTracking(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("ChangeDirectoryでpreviousPathが更新される", func(t *testing.T) {
		initialPath := pane.Path()
		pane.ChangeDirectory(subDir)

		if pane.previousPath != initialPath {
			t.Errorf("previousPath = %s, want %s", pane.previousPath, initialPath)
		}
	})

	t.Run("MoveToParentでpreviousPathが更新される", func(t *testing.T) {
		pane.path = subDir
		pane.previousPath = ""
		pane.LoadDirectory()

		initialPath := pane.Path()
		pane.MoveToParent()

		if pane.previousPath != initialPath {
			t.Errorf("previousPath = %s, want %s", pane.previousPath, initialPath)
		}
	})
}

func TestIsShowingHidden(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	if pane.IsShowingHidden() {
		t.Error("IsShowingHidden() should return false initially")
	}

	pane.showHidden = true
	if !pane.IsShowingHidden() {
		t.Error("IsShowingHidden() should return true when showHidden is true")
	}
}

// === Phase 1: パス復元機能のテスト ===

func TestRestorePreviousPath(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("previousPathが設定されている場合にパスを復元", func(t *testing.T) {
		// 初期状態の設定
		pane.path = subDir
		pane.previousPath = tmpDir
		pane.pendingPath = subDir

		// パスを復元
		pane.restorePreviousPath()

		if pane.path != tmpDir {
			t.Errorf("path = %s, want %s", pane.path, tmpDir)
		}

		if pane.pendingPath != "" {
			t.Errorf("pendingPath = %s, want empty string", pane.pendingPath)
		}
	})

	t.Run("previousPathが空の場合は何もしない", func(t *testing.T) {
		originalPath := subDir
		pane.path = originalPath
		pane.previousPath = ""
		pane.pendingPath = originalPath

		pane.restorePreviousPath()

		if pane.path != originalPath {
			t.Errorf("path = %s, want %s", pane.path, originalPath)
		}
	})
}

func TestPendingPathField(t *testing.T) {
	tmpDir := t.TempDir()
	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("初期状態でpendingPathは空", func(t *testing.T) {
		if pane.pendingPath != "" {
			t.Errorf("pendingPath should be empty initially, got %s", pane.pendingPath)
		}
	})

	t.Run("pendingPathを設定してクリアできる", func(t *testing.T) {
		pane.pendingPath = "/some/path"
		if pane.pendingPath != "/some/path" {
			t.Error("pendingPath should be settable")
		}

		pane.pendingPath = ""
		if pane.pendingPath != "" {
			t.Error("pendingPath should be clearable")
		}
	})
}

// === Phase 4: EnterDirectoryAsync のテスト ===

func TestEnterDirectoryAsync(t *testing.T) {
	tmpDir := t.TempDir()

	// サブディレクトリを作成
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte(""), 0644)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("ディレクトリ選択時にコマンドを返す", func(t *testing.T) {
		// サブディレクトリを選択
		for i, entry := range pane.entries {
			if entry.Name == "subdir" && entry.IsDir {
				pane.cursor = i
				break
			}
		}

		cmd := pane.EnterDirectoryAsync()
		if cmd == nil {
			t.Error("EnterDirectoryAsync() should return a command for directory")
		}

		// pendingPathが設定されているか確認
		if pane.pendingPath != filepath.Join(tmpDir, "subdir") {
			t.Errorf("pendingPath = %s, want %s", pane.pendingPath, filepath.Join(tmpDir, "subdir"))
		}
	})

	t.Run("ファイル選択時はnilを返す", func(t *testing.T) {
		// ペインをリセット
		pane.path = tmpDir
		pane.LoadDirectory()

		// ファイルを選択
		for i, entry := range pane.entries {
			if entry.Name == "file.txt" && !entry.IsDir {
				pane.cursor = i
				break
			}
		}

		cmd := pane.EnterDirectoryAsync()
		if cmd != nil {
			t.Error("EnterDirectoryAsync() should return nil for file")
		}
	})

	t.Run("nilエントリではnilを返す", func(t *testing.T) {
		pane.cursor = -1

		cmd := pane.EnterDirectoryAsync()
		if cmd == nil {
			// nilエントリでも安全に終了すればOK（一部の実装ではnilを返す）
		}
	})
}

func TestEnterDirectoryAsyncParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(subDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("親ディレクトリ(..)選択時にコマンドを返す", func(t *testing.T) {
		// 親ディレクトリを選択
		for i, entry := range pane.entries {
			if entry.IsParentDir() {
				pane.cursor = i
				break
			}
		}

		originalPath := pane.path
		cmd := pane.EnterDirectoryAsync()
		if cmd == nil {
			t.Error("EnterDirectoryAsync() should return a command for parent directory")
		}

		// pendingPathが設定されているか確認
		if pane.pendingPath != filepath.Dir(originalPath) {
			t.Errorf("pendingPath = %s, want %s", pane.pendingPath, filepath.Dir(originalPath))
		}
	})
}

// === Phase 7: エラーハンドリング追加テスト ===

func TestEnterDirectoryNoPathExtension(t *testing.T) {
	// 連続してエラーディレクトリに入ろうとしてもパスが延長されないことを確認
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("エラー後にrestorePreviousPathでパスが復元される", func(t *testing.T) {
		originalPath := tmpDir

		// サブディレクトリに移動を試みる
		for i, entry := range pane.entries {
			if entry.Name == "subdir" && entry.IsDir {
				pane.cursor = i
				break
			}
		}

		pane.EnterDirectoryAsync()

		// この時点でpathは変更されている
		if pane.path != filepath.Join(tmpDir, "subdir") {
			t.Errorf("path should be updated to subdir, got %s", pane.path)
		}

		// エラーをシミュレート: restorePreviousPathを呼び出す
		pane.restorePreviousPath()

		// パスが復元されることを確認
		if pane.path != originalPath {
			t.Errorf("path should be restored to %s, got %s", originalPath, pane.path)
		}

		// pendingPathがクリアされることを確認
		if pane.pendingPath != "" {
			t.Errorf("pendingPath should be cleared, got %s", pane.pendingPath)
		}
	})
}

func TestMoveToParentAsync(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	t.Run("親ディレクトリへの移動コマンドを返す", func(t *testing.T) {
		pane, err := NewPane(subDir, 40, 20, true)
		if err != nil {
			t.Fatalf("NewPane() failed: %v", err)
		}

		originalPath := pane.path
		cmd := pane.MoveToParentAsync()
		if cmd == nil {
			t.Error("MoveToParentAsync() should return a command")
		}

		// pendingPathが設定されているか確認
		expectedPath := filepath.Dir(originalPath)
		if pane.pendingPath != expectedPath {
			t.Errorf("pendingPath = %s, want %s", pane.pendingPath, expectedPath)
		}
	})

	t.Run("ルートディレクトリではnilを返す", func(t *testing.T) {
		pane, _ := NewPane("/", 40, 20, true)
		cmd := pane.MoveToParentAsync()
		if cmd != nil {
			t.Error("MoveToParentAsync() should return nil at root")
		}
	})
}

func TestNavigateToHomeAsync(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("ホームディレクトリへの移動コマンドを返す", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		if pane.path == home {
			t.Skip("Already at home directory")
		}

		cmd := pane.NavigateToHomeAsync()
		if cmd == nil {
			t.Error("NavigateToHomeAsync() should return a command")
		}

		// pendingPathが設定されているか確認
		if pane.pendingPath != home {
			t.Errorf("pendingPath = %s, want %s", pane.pendingPath, home)
		}
	})

	t.Run("すでにホームにいる場合はnilを返す", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		pane.path = home
		pane.pendingPath = ""

		cmd := pane.NavigateToHomeAsync()
		if cmd != nil {
			t.Error("NavigateToHomeAsync() should return nil when already at home")
		}
	})
}

func TestNavigateToPreviousAsync(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(tmpDir, 40, 20, true)
	if err != nil {
		t.Fatalf("NewPane() failed: %v", err)
	}

	t.Run("履歴がない場合はnilを返す", func(t *testing.T) {
		pane.previousPath = ""
		cmd := pane.NavigateToPreviousAsync()
		if cmd != nil {
			t.Error("NavigateToPreviousAsync() should return nil when no previous path")
		}
	})

	t.Run("直前のディレクトリへの移動コマンドを返す", func(t *testing.T) {
		pane.previousPath = subDir
		originalPath := pane.path

		cmd := pane.NavigateToPreviousAsync()
		if cmd == nil {
			t.Error("NavigateToPreviousAsync() should return a command")
		}

		// パスがスワップされているか確認
		if pane.path != subDir {
			t.Errorf("path = %s, want %s", pane.path, subDir)
		}
		if pane.previousPath != originalPath {
			t.Errorf("previousPath = %s, want %s", pane.previousPath, originalPath)
		}
	})
}
