package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/ui"
	"github.com/sakura/duofm/internal/version"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "-version", "--version":
			fmt.Printf("duofm %s\n", version.Version)
			return
		}
	}
	// Ambiguous幅文字（☆、ü、①など）を幅1として扱う
	// 多くのモダンターミナルの実際の表示に合わせる設定
	// TODO: 将来的には設定ファイルで変更可能にする
	runewidth.DefaultCondition.EastAsianWidth = false

	// 設定ファイルの読み込み
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not determine config path: %v\n", err)
	}

	var warnings []string
	var loadResult *config.ConfigLoadResult

	if configPath != "" {
		// 設定ファイルが存在しない場合は自動生成
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := config.GenerateDefaultConfig(configPath); err != nil {
				// 生成に失敗しても警告のみ
				warnings = append(warnings, fmt.Sprintf("Warning: could not generate config: %v", err))
			}
		}
		// 詳細エラー付き設定読み込み（ホットリロード対応）
		loadResult = config.LoadConfigDetailed(configPath)
		warnings = append(warnings, loadResult.Warnings...)
	} else {
		loadResult = &config.ConfigLoadResult{
			Config: &config.Config{
				Keybindings:   config.DefaultKeybindings(),
				Colors:        config.DefaultColors(),
				HistoryLimit:  config.DefaultHistoryLimit,
				RefreshRate:   config.DefaultRefreshRate,
				ShellLogDir:   config.DefaultShellLogDir,
				EnterBehavior: config.DefaultEnterBehavior(),
			},
		}
	}

	cfg := loadResult.Config

	// 重複キーのバリデーション
	validationWarnings := config.ValidateKeybindings(cfg)
	warnings = append(warnings, validationWarnings...)

	// KeybindingMapを生成
	keybindingMap := ui.NewKeybindingMap(cfg)

	// Themeを生成
	theme := ui.NewTheme(cfg.Colors)

	// DirSortStoreを初期化
	var dirSortStore *config.DirSortStore
	if configDir, err := config.GetConfigDir(); err == nil {
		dirSortStore = config.NewDirSortStore(configDir)
		if err := dirSortStore.Load(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Warning: failed to load dir sort settings: %v", err))
		}
	}

	// Modelを作成
	model := ui.NewModelWithConfig(ui.ModelOptions{
		KeybindingMap: keybindingMap,
		Theme:         theme,
		Warnings:      warnings,
		HistoryLimit:  cfg.HistoryLimit,
		RefreshRate:   cfg.RefreshRate,
		EnterBehavior: cfg.EnterBehavior,
		MIMEBehavior:  cfg.MIMEBehavior,
		ConfigPath:    configPath,
		DirSortStore:  dirSortStore,
		ShellLogDir:   cfg.ShellLogDir,
	})

	// 起動時エラーがあればpendingReloadResultに設定
	if loadResult.HasErrors() {
		model.SetPendingReloadResult(loadResult)
	}

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // 代替画面バッファを使用
		tea.WithMouseCellMotion(), // マウスサポート（将来用）
	)

	// ウォッチャー初期化
	if configPath != "" {
		watcher, err := config.NewConfigWatcher(configPath, func(msg interface{}) {
			p.Send(msg)
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: config watch failed: %v\n", err)
		} else {
			watcher.Start()
			defer watcher.Stop()
			model.SetConfigWatcher(watcher)
		}
	}

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
