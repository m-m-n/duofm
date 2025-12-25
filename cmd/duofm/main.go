package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/ui"
)

func main() {
	// Ambiguous幅文字（☆、ü、①など）を幅1として扱う
	// 多くのモダンターミナルの実際の表示に合わせる設定
	// TODO: 将来的には設定ファイルで変更可能にする
	runewidth.DefaultCondition.EastAsianWidth = false

	p := tea.NewProgram(
		ui.NewModel(),
		tea.WithAltScreen(),       // 代替画面バッファを使用
		tea.WithMouseCellMotion(), // マウスサポート（将来用）
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
