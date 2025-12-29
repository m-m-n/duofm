package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/config"
	"github.com/sakura/duofm/internal/ui"
)

// version is set via ldflags at build time
var version = "dev"

func main() {
	// Handle version flag
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "-version", "--version":
			fmt.Printf("duofm %s\n", version)
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

	var cfg *config.Config
	var warnings []string

	if configPath != "" {
		// 設定ファイルが存在しない場合は自動生成
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := config.GenerateDefaultConfig(configPath); err != nil {
				// 生成に失敗しても警告のみ
				warnings = append(warnings, fmt.Sprintf("Warning: could not generate config: %v", err))
			}
		}
		cfg, warnings = config.LoadConfig(configPath)
	} else {
		cfg = &config.Config{
			Keybindings: config.DefaultKeybindings(),
			Colors:      config.DefaultColors(),
		}
	}

	// 重複キーのバリデーション
	validationWarnings := config.ValidateKeybindings(cfg)
	warnings = append(warnings, validationWarnings...)

	// KeybindingMapを生成
	keybindingMap := ui.NewKeybindingMap(cfg)

	// Themeを生成
	theme := ui.NewTheme(cfg.Colors)

	p := tea.NewProgram(
		ui.NewModelWithConfig(keybindingMap, theme, warnings),
		tea.WithAltScreen(),       // 代替画面バッファを使用
		tea.WithMouseCellMotion(), // マウスサポート（将来用）
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
