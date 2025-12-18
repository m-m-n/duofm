package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// カラースキーム
	primaryColor   = lipgloss.Color("39")  // 青
	secondaryColor = lipgloss.Color("240") // グレー
	highlightColor = lipgloss.Color("205") // ピンク
	errorColor     = lipgloss.Color("196") // 赤

	// スタイル定義（Phase 1.3で拡張）
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)
)
