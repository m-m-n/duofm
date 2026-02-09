package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sakura/duofm/internal/config"
)

func TestSortDialogConfirmSavesToStore(t *testing.T) {
	storeDir := t.TempDir()
	store := config.NewDirSortStore(storeDir)

	model := NewModel()
	model.dirSortStore = store

	// Setup panes
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, DefaultTheme())
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}
	pane.SetDirSortStore(store)
	model.leftPane = pane
	model.rightPane, _ = NewPane(RightPane, tmpDir, 80, 20, false, DefaultTheme())
	model.ready = true

	// Set sort config on active pane to Size descending
	pane.SetSortConfig(SortConfig{Field: SortBySize, Order: SortDesc})

	// Simulate sort dialog confirmation (not cancelled)
	msg := sortDialogResultMsg{
		config:    DefaultSortConfig(), // original config for restore on cancel
		confirmed: true,
		cancelled: false,
	}

	model.handleDialogMessages(msg)

	// Verify store has the entry
	field, order, found := store.Get(tmpDir)
	if !found {
		t.Fatal("expected sort setting to be saved to store")
	}
	if field != "size" || order != "desc" {
		t.Errorf("expected size/desc, got %s/%s", field, order)
	}
}

func TestSortDialogCancelDoesNotSave(t *testing.T) {
	storeDir := t.TempDir()
	store := config.NewDirSortStore(storeDir)

	model := NewModel()
	model.dirSortStore = store

	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)

	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, DefaultTheme())
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}
	pane.SetDirSortStore(store)
	model.leftPane = pane
	model.rightPane, _ = NewPane(RightPane, tmpDir, 80, 20, false, DefaultTheme())
	model.ready = true

	// Simulate sort dialog cancel
	msg := sortDialogResultMsg{
		config:    DefaultSortConfig(),
		confirmed: false,
		cancelled: true,
	}

	model.handleDialogMessages(msg)

	// Verify store does NOT have the entry
	_, _, found := store.Get(tmpDir)
	if found {
		t.Error("expected no sort setting saved to store after cancel")
	}
}

func TestNavigationAppliesSavedSort(t *testing.T) {
	storeDir := t.TempDir()
	store := config.NewDirSortStore(storeDir)

	// Create directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("content"), 0644)

	// Save a sort setting for subDir
	store.Set(subDir, "date", "desc")

	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}
	pane.SetDirSortStore(store)

	// Navigate to subDir
	pane.ChangeDirectoryAsync(subDir)

	// Verify pane's sort config was set to the saved value
	sc := pane.GetSortConfig()
	if sc.Field != SortByDate || sc.Order != SortDesc {
		t.Errorf("expected Date/Desc, got %v", sc)
	}
}

func TestNavigationAppliesDefaultWhenNoSavedSort(t *testing.T) {
	storeDir := t.TempDir()
	store := config.NewDirSortStore(storeDir)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)

	pane, err := NewPane(LeftPane, tmpDir, 80, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}
	pane.SetDirSortStore(store)

	// Set a non-default sort first
	pane.SetSortConfig(SortConfig{Field: SortBySize, Order: SortDesc})

	// Navigate to subDir (no saved setting)
	pane.ChangeDirectoryAsync(subDir)

	// Verify pane's sort config was reset to default
	sc := pane.GetSortConfig()
	expected := DefaultSortConfig()
	if sc.Field != expected.Field || sc.Order != expected.Order {
		t.Errorf("expected default sort %v, got %v", expected, sc)
	}
}

func TestAllNavigationMethodsApplySavedSort(t *testing.T) {
	storeDir := t.TempDir()
	store := config.NewDirSortStore(storeDir)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("x"), 0644)

	// Save sort settings
	store.Set(subDir, "size", "desc")
	store.Set(tmpDir, "date", "asc")

	t.Run("MoveToParentAsync", func(t *testing.T) {
		pane, _ := NewPane(LeftPane, subDir, 80, 20, true, nil)
		pane.SetDirSortStore(store)
		pane.MoveToParentAsync()
		sc := pane.GetSortConfig()
		if sc.Field != SortByDate || sc.Order != SortAsc {
			t.Errorf("expected Date/Asc for parent, got %v", sc)
		}
	})

	t.Run("NavigateToHomeAsync", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get home directory")
		}
		store.Set(home, "name", "desc")
		pane, _ := NewPane(LeftPane, tmpDir, 80, 20, true, nil)
		pane.SetDirSortStore(store)
		pane.NavigateToHomeAsync()
		sc := pane.GetSortConfig()
		if sc.Field != SortByName || sc.Order != SortDesc {
			t.Errorf("expected Name/Desc for home, got %v", sc)
		}
	})

	t.Run("SyncTo", func(t *testing.T) {
		pane, _ := NewPane(LeftPane, tmpDir, 80, 20, true, nil)
		pane.SetDirSortStore(store)
		pane.SyncTo(subDir)
		sc := pane.GetSortConfig()
		if sc.Field != SortBySize || sc.Order != SortDesc {
			t.Errorf("expected Size/Desc for synced dir, got %v", sc)
		}
	})
}

func TestModelOptionsRefactoring(t *testing.T) {
	t.Run("NewModel still works", func(t *testing.T) {
		model := NewModel()
		if model.refreshRate != config.DefaultRefreshRate {
			t.Errorf("expected default refresh rate %d, got %d", config.DefaultRefreshRate, model.refreshRate)
		}
	})

	t.Run("NewModelWithConfig with options", func(t *testing.T) {
		model := NewModelWithConfig(ModelOptions{
			RefreshRate:  5,
			HistoryLimit: 100,
		})
		if model.refreshRate != 5 {
			t.Errorf("expected refresh rate 5, got %d", model.refreshRate)
		}
	})

	t.Run("DirSortStore passed via options", func(t *testing.T) {
		storeDir := t.TempDir()
		store := config.NewDirSortStore(storeDir)
		model := NewModelWithConfig(ModelOptions{
			DirSortStore: store,
		})
		if model.dirSortStore != store {
			t.Error("expected dirSortStore to be set from options")
		}
	})
}
