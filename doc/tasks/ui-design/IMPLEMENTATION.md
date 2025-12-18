# Implementation Plan: UI Design (MVP)

## 概要

本ドキュメントは、duofm（TUIデュアルペインファイルマネージャー）のMVP（Minimum Viable Product）実装計画を定義します。

**目標**: 基本的な機能を持つ動作可能なデュアルペインファイルマネージャーを構築する

**技術スタック**:
- Go 1.21+
- Bubble Tea (TUIフレームワーク)
- Lip Gloss (スタイリング)

**参照仕様**: `doc/tasks/ui-design/SPEC.md`

## 実装フェーズ

### Phase 1: 基本構造とセットアップ (1-2日)

#### 1.1 プロジェクト基盤の準備

**タスク**:
- 必要な依存関係のインストール
- プロジェクト構造の作成
- 基本的なエントリーポイントの実装

**ファイル構成**:
```
cmd/duofm/
└── main.go                    # アプリケーションエントリーポイント

internal/ui/
├── model.go                   # Bubble Tea メインモデル
├── keys.go                    # キーバインディング定義
└── styles.go                  # Lip Gloss スタイル定義

internal/fs/
└── types.go                   # 基本的な型定義
```

**実装手順**:

1. **依存関係のインストール**:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go mod tidy
```

2. **`cmd/duofm/main.go` の実装**:
```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/sakura/duofm/internal/ui"
)

func main() {
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
```

3. **`internal/ui/model.go` の基本構造**:
```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
)

// Model はアプリケーション全体の状態を保持
type Model struct {
    leftPane   *Pane
    rightPane  *Pane
    activePane PanePosition
    dialog     Dialog
    width      int
    height     int
    ready      bool
}

type PanePosition int

const (
    LeftPane PanePosition = iota
    RightPane
)

// NewModel は初期モデルを作成
func NewModel() Model {
    return Model{
        leftPane:   nil, // Phase 1.2で実装
        rightPane:  nil, // Phase 1.2で実装
        activePane: LeftPane,
        dialog:     nil,
        ready:      false,
    }
}

// Init はBubble Teaの初期化
func (m Model) Init() tea.Cmd {
    return nil
}

// Update はメッセージを処理してモデルを更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.ready = true
        return m, nil

    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        }
    }

    return m, nil
}

// View はUIをレンダリング
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }
    return "duofm - Coming soon!"
}
```

4. **`internal/ui/keys.go` の初期実装**:
```go
package ui

// キーバインディング定義
const (
    KeyMoveDown  = "j"
    KeyMoveUp    = "k"
    KeyMoveLeft  = "h"
    KeyMoveRight = "l"
    KeyEnter     = "enter"
    KeyCopy      = "c"
    KeyMove      = "m"
    KeyDelete    = "d"
    KeyHelp      = "?"
    KeyEscape    = "esc"
    KeyQuit      = "q"
)
```

5. **`internal/ui/styles.go` の基本実装**:
```go
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
```

**テスト**:
```bash
go build -o duofm ./cmd/duofm
./duofm
# "duofm - Coming soon!" と表示され、'q' で終了できることを確認
```

---

### Phase 2: ペインコンポーネントの実装 (2-3日)

#### 2.1 ディレクトリ読み込みとファイルエントリ

**ファイル**:
```
internal/fs/
├── types.go        # データ型定義
├── reader.go       # ディレクトリ読み込み
└── sort.go         # ソート処理
```

**実装手順**:

1. **`internal/fs/types.go` の実装**:
```go
package fs

import (
    "io/fs"
    "time"
)

// FileEntry はファイル/ディレクトリの情報を保持
type FileEntry struct {
    Name        string
    IsDir       bool
    Size        int64
    ModTime     time.Time
    Permissions fs.FileMode
}

// IsParentDir は親ディレクトリエントリかどうかを判定
func (e FileEntry) IsParentDir() bool {
    return e.Name == ".."
}

// DisplayName は表示用の名前を返す
func (e FileEntry) DisplayName() string {
    if e.IsDir && !e.IsParentDir() {
        return e.Name + "/"
    }
    return e.Name
}
```

2. **`internal/fs/reader.go` の実装**:
```go
package fs

import (
    "fmt"
    "os"
    "path/filepath"
)

// ReadDirectory はディレクトリの内容を読み込む
func ReadDirectory(path string) ([]FileEntry, error) {
    // パスの正規化
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, fmt.Errorf("invalid path: %w", err)
    }

    // ディレクトリの読み込み
    entries, err := os.ReadDir(absPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory: %w", err)
    }

    // FileEntry に変換
    var fileEntries []FileEntry

    // 親ディレクトリエントリを追加
    if absPath != "/" {
        fileEntries = append(fileEntries, FileEntry{
            Name:  "..",
            IsDir: true,
        })
    }

    // 各エントリを処理
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue // エラーは無視して次へ
        }

        fileEntries = append(fileEntries, FileEntry{
            Name:        entry.Name(),
            IsDir:       entry.IsDir(),
            Size:        info.Size(),
            ModTime:     info.ModTime(),
            Permissions: info.Mode(),
        })
    }

    return fileEntries, nil
}

// HomeDirectory はホームディレクトリのパスを返す
func HomeDirectory() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get home directory: %w", err)
    }
    return home, nil
}

// CurrentDirectory は現在の作業ディレクトリを返す
func CurrentDirectory() (string, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get current directory: %w", err)
    }
    return cwd, nil
}
```

3. **`internal/fs/sort.go` の実装**:
```go
package fs

import (
    "sort"
    "strings"
)

// SortEntries はエントリをソート（ディレクトリ優先、アルファベット順）
func SortEntries(entries []FileEntry) {
    sort.Slice(entries, func(i, j int) bool {
        // 親ディレクトリ (..) は常に最初
        if entries[i].IsParentDir() {
            return true
        }
        if entries[j].IsParentDir() {
            return false
        }

        // ディレクトリとファイルを分離
        if entries[i].IsDir != entries[j].IsDir {
            return entries[i].IsDir // ディレクトリを先に
        }

        // 同じタイプ内では名前でソート（大文字小文字を区別しない）
        return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
    })
}
```

**テスト**:
```go
// internal/fs/reader_test.go
package fs

import (
    "os"
    "path/filepath"
    "testing"
)

func TestReadDirectory(t *testing.T) {
    // テスト用ディレクトリを作成
    tmpDir := t.TempDir()

    // テストファイル/ディレクトリを作成
    os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
    os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755)
    os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)

    entries, err := ReadDirectory(tmpDir)
    if err != nil {
        t.Fatalf("ReadDirectory failed: %v", err)
    }

    // 親ディレクトリ + 3エントリ = 4
    if len(entries) != 4 {
        t.Errorf("Expected 4 entries, got %d", len(entries))
    }

    // 親ディレクトリが最初にあることを確認
    if !entries[0].IsParentDir() {
        t.Error("First entry should be parent directory")
    }
}

func TestSortEntries(t *testing.T) {
    entries := []FileEntry{
        {Name: "file.txt", IsDir: false},
        {Name: "another_dir", IsDir: true},
        {Name: "..", IsDir: true},
        {Name: "zebra_dir", IsDir: true},
    }

    SortEntries(entries)

    // 順序確認: .. -> ディレクトリ（アルファベット順） -> ファイル
    if entries[0].Name != ".." {
        t.Error("Parent dir should be first")
    }
    if entries[1].Name != "another_dir" {
        t.Error("Directories should come before files")
    }
}
```

#### 2.2 ペインコンポーネントの実装

**ファイル**:
```
internal/ui/
└── pane.go    # ペイン実装
```

**実装手順**:

1. **`internal/ui/pane.go` の実装**:
```go
package ui

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/sakura/duofm/internal/fs"
)

// Pane は1つのファイルリストペインを表現
type Pane struct {
    path         string
    entries      []fs.FileEntry
    cursor       int
    scrollOffset int
    width        int
    height       int
    isActive     bool
}

// NewPane は新しいペインを作成
func NewPane(path string, width, height int, isActive bool) (*Pane, error) {
    pane := &Pane{
        path:         path,
        width:        width,
        height:       height,
        isActive:     isActive,
        cursor:       0,
        scrollOffset: 0,
    }

    if err := pane.LoadDirectory(); err != nil {
        return nil, err
    }

    return pane, nil
}

// LoadDirectory はディレクトリを読み込む
func (p *Pane) LoadDirectory() error {
    entries, err := fs.ReadDirectory(p.path)
    if err != nil {
        return err
    }

    fs.SortEntries(entries)
    p.entries = entries
    p.cursor = 0
    p.scrollOffset = 0

    return nil
}

// MoveCursorDown はカーソルを下に移動
func (p *Pane) MoveCursorDown() {
    if p.cursor < len(p.entries)-1 {
        p.cursor++
        p.adjustScroll()
    }
}

// MoveCursorUp はカーソルを上に移動
func (p *Pane) MoveCursorUp() {
    if p.cursor > 0 {
        p.cursor--
        p.adjustScroll()
    }
}

// adjustScroll はスクロール位置を調整
func (p *Pane) adjustScroll() {
    visibleLines := p.height - 3 // ヘッダーとボーダーを除く

    // カーソルが表示範囲外に出たらスクロール
    if p.cursor < p.scrollOffset {
        p.scrollOffset = p.cursor
    } else if p.cursor >= p.scrollOffset+visibleLines {
        p.scrollOffset = p.cursor - visibleLines + 1
    }
}

// SelectedEntry は選択中のエントリを返す
func (p *Pane) SelectedEntry() *fs.FileEntry {
    if p.cursor < 0 || p.cursor >= len(p.entries) {
        return nil
    }
    return &p.entries[p.cursor]
}

// EnterDirectory はディレクトリに入る
func (p *Pane) EnterDirectory() error {
    entry := p.SelectedEntry()
    if entry == nil || !entry.IsDir {
        return nil // ファイルの場合は何もしない
    }

    var newPath string
    if entry.IsParentDir() {
        // 親ディレクトリに移動
        newPath = filepath.Dir(p.path)
    } else {
        // サブディレクトリに移動
        newPath = filepath.Join(p.path, entry.Name)
    }

    p.path = newPath
    return p.LoadDirectory()
}

// MoveToParent は親ディレクトリに移動
func (p *Pane) MoveToParent() error {
    if p.path == "/" {
        return nil // ルートより上には行けない
    }

    p.path = filepath.Dir(p.path)
    return p.LoadDirectory()
}

// Path は現在のパスを返す
func (p *Pane) Path() string {
    return p.path
}

// View はペインをレンダリング
func (p *Pane) View() string {
    var b strings.Builder

    // パス表示（ホームディレクトリは ~ に置換）
    displayPath := p.formatPath()
    pathStyle := lipgloss.NewStyle().
        Width(p.width - 2).
        Padding(0, 1).
        Bold(true)

    if p.isActive {
        pathStyle = pathStyle.Foreground(lipgloss.Color("39"))
    } else {
        pathStyle = pathStyle.Foreground(lipgloss.Color("240"))
    }

    b.WriteString(pathStyle.Render(displayPath))
    b.WriteString("\n")

    // 区切り線
    border := strings.Repeat("─", p.width-2)
    b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(border))
    b.WriteString("\n")

    // ファイルリスト
    visibleLines := p.height - 3
    endIdx := p.scrollOffset + visibleLines
    if endIdx > len(p.entries) {
        endIdx = len(p.entries)
    }

    for i := p.scrollOffset; i < endIdx; i++ {
        entry := p.entries[i]
        line := p.formatEntry(entry, i == p.cursor)
        b.WriteString(line)
        b.WriteString("\n")
    }

    // 空行で埋める
    for i := endIdx - p.scrollOffset; i < visibleLines; i++ {
        b.WriteString(strings.Repeat(" ", p.width))
        b.WriteString("\n")
    }

    return b.String()
}

// formatPath はパスを表示用にフォーマット
func (p *Pane) formatPath() string {
    home, _ := fs.HomeDirectory()
    if strings.HasPrefix(p.path, home) {
        return "~" + strings.TrimPrefix(p.path, home)
    }
    return p.path
}

// formatEntry はエントリを1行にフォーマット
func (p *Pane) formatEntry(entry fs.FileEntry, isCursor bool) string {
    displayName := entry.DisplayName()

    // 表示幅を調整（パディングを考慮）
    maxWidth := p.width - 4
    if len(displayName) > maxWidth {
        displayName = displayName[:maxWidth-3] + "..."
    }

    style := lipgloss.NewStyle().
        Width(p.width - 2).
        Padding(0, 1)

    if isCursor {
        if p.isActive {
            style = style.Background(lipgloss.Color("39")).
                Foreground(lipgloss.Color("15"))
        } else {
            style = style.Background(lipgloss.Color("240")).
                Foreground(lipgloss.Color("15"))
        }
    } else if entry.IsDir {
        style = style.Foreground(lipgloss.Color("39"))
    }

    return style.Render(displayName)
}

// SetSize はペインサイズを設定
func (p *Pane) SetSize(width, height int) {
    p.width = width
    p.height = height
}

// SetActive はアクティブ状態を設定
func (p *Pane) SetActive(active bool) {
    p.isActive = active
}
```

**テスト**:
```go
// internal/ui/pane_test.go
package ui

import (
    "os"
    "testing"
)

func TestNewPane(t *testing.T) {
    tmpDir := t.TempDir()

    pane, err := NewPane(tmpDir, 40, 20, true)
    if err != nil {
        t.Fatalf("NewPane failed: %v", err)
    }

    if pane.cursor != 0 {
        t.Error("Initial cursor should be 0")
    }

    if len(pane.entries) == 0 {
        t.Error("Entries should not be empty")
    }
}

func TestMoveCursor(t *testing.T) {
    tmpDir := t.TempDir()
    os.WriteFile(tmpDir+"/file1.txt", []byte(""), 0644)
    os.WriteFile(tmpDir+"/file2.txt", []byte(""), 0644)

    pane, _ := NewPane(tmpDir, 40, 20, true)

    initialCursor := pane.cursor
    pane.MoveCursorDown()

    if pane.cursor <= initialCursor {
        t.Error("Cursor should move down")
    }

    pane.MoveCursorUp()
    if pane.cursor != initialCursor {
        t.Error("Cursor should move back up")
    }

    // 境界テスト
    for i := 0; i < 100; i++ {
        pane.MoveCursorUp()
    }
    if pane.cursor != 0 {
        t.Error("Cursor should not go below 0")
    }
}
```

#### 2.3 モデルとペインの統合

**`internal/ui/model.go` の更新**:

```go
// NewModel の更新
func NewModel() Model {
    // 初期ディレクトリの取得
    cwd, err := fs.CurrentDirectory()
    if err != nil {
        cwd = "/"
    }

    home, err := fs.HomeDirectory()
    if err != nil {
        home = "/"
    }

    return Model{
        leftPane:   nil, // Initで初期化
        rightPane:  nil, // Initで初期化
        leftPath:   cwd,
        rightPath:  home,
        activePane: LeftPane,
        dialog:     nil,
        ready:      false,
    }
}

// Init の更新
func (m Model) Init() tea.Cmd {
    return nil
}

// Update の更新（ペイン初期化とナビゲーション）
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

        if !m.ready {
            // 初回のみペインを作成
            paneWidth := msg.Width / 2
            paneHeight := msg.Height - 2 // ステータスバー分を引く

            var err error
            m.leftPane, err = NewPane(m.leftPath, paneWidth, paneHeight, true)
            if err != nil {
                // エラーハンドリングは Phase 3 で実装
                return m, tea.Quit
            }

            m.rightPane, err = NewPane(m.rightPath, paneWidth, paneHeight, false)
            if err != nil {
                return m, tea.Quit
            }

            m.ready = true
        } else {
            // リサイズ時のペインサイズ更新
            paneWidth := msg.Width / 2
            paneHeight := msg.Height - 2
            m.leftPane.SetSize(paneWidth, paneHeight)
            m.rightPane.SetSize(paneWidth, paneHeight)
        }

        return m, nil

    case tea.KeyMsg:
        // ダイアログが開いている場合はダイアログに処理を委譲
        if m.dialog != nil {
            // Phase 3 で実装
            return m, nil
        }

        switch msg.String() {
        case "ctrl+c", KeyQuit:
            return m, tea.Quit

        case KeyMoveDown:
            m.activePane().MoveCursorDown()

        case KeyMoveUp:
            m.activePane().MoveCursorUp()

        case KeyMoveLeft:
            if m.activePane == LeftPane {
                // 左ペインで h -> 親ディレクトリへ
                m.leftPane.MoveToParent()
            } else {
                // 右ペインで h -> 左ペインへ切り替え
                m.switchToPane(LeftPane)
            }

        case KeyMoveRight:
            if m.activePane == RightPane {
                // 右ペインで l -> 親ディレクトリへ
                m.rightPane.MoveToParent()
            } else {
                // 左ペインで l -> 右ペインへ切り替え
                m.switchToPane(RightPane)
            }

        case KeyEnter:
            m.activePane().EnterDirectory()
        }
    }

    return m, nil
}

// activePane は現在アクティブなペインを返す
func (m *Model) activePane() *Pane {
    if m.activePane == LeftPane {
        return m.leftPane
    }
    return m.rightPane
}

// inactivePane は非アクティブなペインを返す
func (m *Model) inactivePane() *Pane {
    if m.activePane == LeftPane {
        return m.rightPane
    }
    return m.leftPane
}

// switchToPane はアクティブペインを切り替え
func (m *Model) switchToPane(pos PanePosition) {
    m.activePane = pos
    m.leftPane.SetActive(pos == LeftPane)
    m.rightPane.SetActive(pos == RightPane)
}

// View の更新
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    // タイトルバー
    title := titleStyle.Render("duofm v0.1.0")

    // 2つのペインを横に並べる
    panes := lipgloss.JoinHorizontal(
        lipgloss.Top,
        m.leftPane.View(),
        m.rightPane.View(),
    )

    // ステータスバー
    statusBar := m.renderStatusBar()

    // 全体を縦に結合
    return lipgloss.JoinVertical(
        lipgloss.Left,
        title,
        panes,
        statusBar,
    )
}

// renderStatusBar はステータスバーをレンダリング
func (m Model) renderStatusBar() string {
    activePane := m.activePane()

    // 選択位置情報
    posInfo := fmt.Sprintf("%d/%d", activePane.cursor+1, len(activePane.entries))

    // キーヒント
    hints := "?:help q:quit"

    // スペースで埋める
    padding := m.width - len(posInfo) - len(hints) - 4
    if padding < 0 {
        padding = 0
    }

    statusBar := fmt.Sprintf(" %s%s%s ",
        posInfo,
        strings.Repeat(" ", padding),
        hints,
    )

    style := lipgloss.NewStyle().
        Width(m.width).
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("15"))

    return style.Render(statusBar)
}
```

**テスト**:
```bash
go build -o duofm ./cmd/duofm
./duofm
# 2つのペインが表示され、hjkl で操作できることを確認
```

---

### Phase 3: ダイアログシステムの実装 (1-2日)

#### 3.1 ダイアログインターフェースと基本実装

**ファイル**:
```
internal/ui/
├── dialog.go           # ダイアログインターフェース
├── confirm_dialog.go   # 確認ダイアログ
├── error_dialog.go     # エラーダイアログ
└── help_dialog.go      # ヘルプダイアログ
```

**実装手順**:

1. **`internal/ui/dialog.go` の実装**:
```go
package ui

import tea "github.com/charmbracelet/bubbletea"

// Dialog はモーダルダイアログのインターフェース
type Dialog interface {
    Update(msg tea.Msg) (Dialog, tea.Cmd)
    View() string
    IsActive() bool
}

// DialogResult はダイアログの結果
type DialogResult struct {
    Confirmed bool
    Cancelled bool
}

// dialogResultMsg はダイアログ結果のメッセージ
type dialogResultMsg struct {
    result DialogResult
}
```

2. **`internal/ui/confirm_dialog.go` の実装**:
```go
package ui

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// ConfirmDialog は確認ダイアログ
type ConfirmDialog struct {
    title   string
    message string
    active  bool
}

// NewConfirmDialog は新しい確認ダイアログを作成
func NewConfirmDialog(title, message string) *ConfirmDialog {
    return &ConfirmDialog{
        title:   title,
        message: message,
        active:  true,
    }
}

// Update はメッセージを処理
func (d *ConfirmDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.active {
        return d, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "y", "enter":
            d.active = false
            return d, func() tea.Msg {
                return dialogResultMsg{
                    result: DialogResult{Confirmed: true},
                }
            }

        case "n", "esc":
            d.active = false
            return d, func() tea.Msg {
                return dialogResultMsg{
                    result: DialogResult{Cancelled: true},
                }
            }
        }
    }

    return d, nil
}

// View はダイアログをレンダリング
func (d *ConfirmDialog) View() string {
    if !d.active {
        return ""
    }

    var b strings.Builder

    width := 50

    // タイトル
    titleStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2).
        Bold(true).
        Foreground(lipgloss.Color("39"))
    b.WriteString(titleStyle.Render(d.title))
    b.WriteString("\n\n")

    // メッセージ
    messageStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2)
    b.WriteString(messageStyle.Render(d.message))
    b.WriteString("\n\n")

    // ボタン
    buttonStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2).
        Foreground(lipgloss.Color("240"))
    b.WriteString(buttonStyle.Render("[y] Yes  [n] No"))

    // ボーダーで囲む
    boxStyle := lipgloss.NewStyle().
        Width(width).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("39")).
        Padding(1, 2)

    return boxStyle.Render(b.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *ConfirmDialog) IsActive() bool {
    return d.active
}
```

3. **`internal/ui/error_dialog.go` の実装**:
```go
package ui

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// ErrorDialog はエラーダイアログ
type ErrorDialog struct {
    message string
    active  bool
}

// NewErrorDialog は新しいエラーダイアログを作成
func NewErrorDialog(message string) *ErrorDialog {
    return &ErrorDialog{
        message: message,
        active:  true,
    }
}

// Update はメッセージを処理
func (d *ErrorDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.active {
        return d, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc", "enter":
            d.active = false
            return d, func() tea.Msg {
                return dialogResultMsg{
                    result: DialogResult{Cancelled: true},
                }
            }
        }
    }

    return d, nil
}

// View はダイアログをレンダリング
func (d *ErrorDialog) View() string {
    if !d.active {
        return ""
    }

    var b strings.Builder

    width := 50

    // タイトル
    titleStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2).
        Bold(true).
        Foreground(lipgloss.Color("196"))
    b.WriteString(titleStyle.Render("Error"))
    b.WriteString("\n\n")

    // メッセージ
    messageStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2)
    b.WriteString(messageStyle.Render(d.message))
    b.WriteString("\n\n")

    // ヒント
    hintStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2).
        Foreground(lipgloss.Color("240"))
    b.WriteString(hintStyle.Render("Press Esc to close"))

    // ボーダーで囲む
    boxStyle := lipgloss.NewStyle().
        Width(width).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("196")).
        Padding(1, 2)

    return boxStyle.Render(b.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *ErrorDialog) IsActive() bool {
    return d.active
}
```

4. **`internal/ui/help_dialog.go` の実装**:
```go
package ui

import (
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// HelpDialog はヘルプダイアログ
type HelpDialog struct {
    active bool
}

// NewHelpDialog は新しいヘルプダイアログを作成
func NewHelpDialog() *HelpDialog {
    return &HelpDialog{
        active: true,
    }
}

// Update はメッセージを処理
func (d *HelpDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    if !d.active {
        return d, nil
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc", "?":
            d.active = false
            return d, func() tea.Msg {
                return dialogResultMsg{
                    result: DialogResult{Cancelled: true},
                }
            }
        }
    }

    return d, nil
}

// View はダイアログをレンダリング
func (d *HelpDialog) View() string {
    if !d.active {
        return ""
    }

    var b strings.Builder

    width := 70

    // タイトル
    titleStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2).
        Bold(true).
        Foreground(lipgloss.Color("39"))
    b.WriteString(titleStyle.Render("Keybindings"))
    b.WriteString("\n\n")

    // カテゴリとキーバインディング
    contentStyle := lipgloss.NewStyle().
        Width(width - 4).
        Padding(0, 2)

    content := []string{
        "Navigation",
        "  j/k      : move cursor down/up",
        "  h/l      : move to left/right pane or parent directory",
        "  Enter    : enter directory",
        "  q        : quit",
        "",
        "File Operations",
        "  c        : copy to opposite pane",
        "  m        : move to opposite pane",
        "  d        : delete (with confirmation)",
        "",
        "Help",
        "  ?        : show this help",
        "",
        "",
        "Press Esc or ? to close",
    }

    b.WriteString(contentStyle.Render(strings.Join(content, "\n")))

    // ボーダーで囲む
    boxStyle := lipgloss.NewStyle().
        Width(width).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("39")).
        Padding(1, 2)

    return boxStyle.Render(b.String())
}

// IsActive はダイアログがアクティブかどうかを返す
func (d *HelpDialog) IsActive() bool {
    return d.active
}
```

#### 3.2 モデルへのダイアログ統合

**`internal/ui/model.go` の更新**:

```go
// Update にダイアログ処理を追加
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ダイアログの結果処理
    if result, ok := msg.(dialogResultMsg); ok {
        m.dialog = nil

        // 削除確認の結果処理（Phase 4で実装）
        if result.result.Confirmed {
            // ファイル削除処理
        }

        return m, nil
    }

    // ダイアログが開いている場合はダイアログに処理を委譲
    if m.dialog != nil {
        var cmd tea.Cmd
        m.dialog, cmd = m.dialog.Update(msg)
        return m, cmd
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // ... (既存のコード)

    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", KeyQuit:
            return m, tea.Quit

        case KeyHelp:
            // ヘルプダイアログを表示
            m.dialog = NewHelpDialog()
            return m, nil

        // ... (既存のナビゲーションコード)
        }
    }

    return m, nil
}

// View にダイアログオーバーレイを追加
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    // タイトルバー
    title := titleStyle.Render("duofm v0.1.0")

    // 2つのペインを横に並べる
    panes := lipgloss.JoinHorizontal(
        lipgloss.Top,
        m.leftPane.View(),
        m.rightPane.View(),
    )

    // ステータスバー
    statusBar := m.renderStatusBar()

    // 全体を縦に結合
    mainView := lipgloss.JoinVertical(
        lipgloss.Left,
        title,
        panes,
        statusBar,
    )

    // ダイアログがある場合はオーバーレイ
    if m.dialog != nil && m.dialog.IsActive() {
        return lipgloss.Place(
            m.width,
            m.height,
            lipgloss.Center,
            lipgloss.Center,
            m.dialog.View(),
            lipgloss.WithWhitespaceChars("█"),
            lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
        )
    }

    return mainView
}
```

**テスト**:
```bash
go build -o duofm ./cmd/duofm
./duofm
# ? キーでヘルプが表示されることを確認
# Esc でヘルプが閉じることを確認
```

---

### Phase 4: ファイル操作の実装 (2-3日)

#### 4.1 ファイル操作関数の実装

**ファイル**:
```
internal/fs/
└── operations.go    # コピー、移動、削除
```

**実装手順**:

1. **`internal/fs/operations.go` の実装**:
```go
package fs

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// CopyFile はファイルをコピー
func CopyFile(src, dst string) error {
    // ソースファイルを開く
    sourceFile, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("failed to open source file: %w", err)
    }
    defer sourceFile.Close()

    // ソースの情報を取得
    sourceInfo, err := sourceFile.Stat()
    if err != nil {
        return fmt.Errorf("failed to stat source file: %w", err)
    }

    // 宛先パスを決定
    dstPath := dst
    if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
        dstPath = filepath.Join(dst, filepath.Base(src))
    }

    // 宛先ファイルを作成
    destFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
    if err != nil {
        return fmt.Errorf("failed to create destination file: %w", err)
    }
    defer destFile.Close()

    // コピー実行
    _, err = io.Copy(destFile, sourceFile)
    if err != nil {
        return fmt.Errorf("failed to copy file: %w", err)
    }

    return nil
}

// CopyDirectory はディレクトリを再帰的にコピー
func CopyDirectory(src, dst string) error {
    // ソースディレクトリの情報を取得
    srcInfo, err := os.Stat(src)
    if err != nil {
        return fmt.Errorf("failed to stat source directory: %w", err)
    }

    // 宛先パスを決定
    dstPath := dst
    if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
        dstPath = filepath.Join(dst, filepath.Base(src))
    }

    // 宛先ディレクトリを作成
    if err := os.MkdirAll(dstPath, srcInfo.Mode()); err != nil {
        return fmt.Errorf("failed to create destination directory: %w", err)
    }

    // ディレクトリの内容を読み込む
    entries, err := os.ReadDir(src)
    if err != nil {
        return fmt.Errorf("failed to read directory: %w", err)
    }

    // 各エントリを再帰的にコピー
    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        destPath := filepath.Join(dstPath, entry.Name())

        if entry.IsDir() {
            if err := CopyDirectory(srcPath, destPath); err != nil {
                return err
            }
        } else {
            if err := CopyFile(srcPath, destPath); err != nil {
                return err
            }
        }
    }

    return nil
}

// Copy はファイルまたはディレクトリをコピー
func Copy(src, dst string) error {
    srcInfo, err := os.Stat(src)
    if err != nil {
        return fmt.Errorf("source not found: %w", err)
    }

    if srcInfo.IsDir() {
        return CopyDirectory(src, dst)
    }
    return CopyFile(src, dst)
}

// MoveFile はファイルを移動
func MoveFile(src, dst string) error {
    // 宛先パスを決定
    dstPath := dst
    if dstInfo, err := os.Stat(dst); err == nil && dstInfo.IsDir() {
        dstPath = filepath.Join(dst, filepath.Base(src))
    }

    // os.Rename を試す（同一ファイルシステム内）
    err := os.Rename(src, dstPath)
    if err == nil {
        return nil
    }

    // クロスデバイス移動の場合はコピー→削除
    if err := Copy(src, dst); err != nil {
        return fmt.Errorf("failed to copy during move: %w", err)
    }

    if err := Delete(src); err != nil {
        return fmt.Errorf("failed to delete source after copy: %w", err)
    }

    return nil
}

// DeleteFile はファイルを削除
func DeleteFile(path string) error {
    if err := os.Remove(path); err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    return nil
}

// DeleteDirectory はディレクトリを再帰的に削除
func DeleteDirectory(path string) error {
    if err := os.RemoveAll(path); err != nil {
        return fmt.Errorf("failed to delete directory: %w", err)
    }
    return nil
}

// Delete はファイルまたはディレクトリを削除
func Delete(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return fmt.Errorf("path not found: %w", err)
    }

    if info.IsDir() {
        return DeleteDirectory(path)
    }
    return DeleteFile(path)
}
```

**テスト**:
```go
// internal/fs/operations_test.go
package fs

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCopyFile(t *testing.T) {
    tmpDir := t.TempDir()

    // ソースファイル作成
    srcFile := filepath.Join(tmpDir, "source.txt")
    content := []byte("test content")
    if err := os.WriteFile(srcFile, content, 0644); err != nil {
        t.Fatal(err)
    }

    // コピー先ディレクトリ
    dstDir := filepath.Join(tmpDir, "dest")
    os.Mkdir(dstDir, 0755)

    // コピー実行
    if err := CopyFile(srcFile, dstDir); err != nil {
        t.Fatalf("CopyFile failed: %v", err)
    }

    // 検証
    dstFile := filepath.Join(dstDir, "source.txt")
    copiedContent, err := os.ReadFile(dstFile)
    if err != nil {
        t.Fatal(err)
    }

    if string(copiedContent) != string(content) {
        t.Error("Copied content does not match")
    }
}

func TestCopyDirectory(t *testing.T) {
    tmpDir := t.TempDir()

    // ソースディレクトリ構造を作成
    srcDir := filepath.Join(tmpDir, "source")
    os.Mkdir(srcDir, 0755)
    os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
    os.Mkdir(filepath.Join(srcDir, "subdir"), 0755)
    os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644)

    // コピー先
    dstDir := filepath.Join(tmpDir, "dest")

    // コピー実行
    if err := CopyDirectory(srcDir, dstDir); err != nil {
        t.Fatalf("CopyDirectory failed: %v", err)
    }

    // 検証
    expectedDst := filepath.Join(dstDir, "source")
    if _, err := os.Stat(filepath.Join(expectedDst, "file1.txt")); err != nil {
        t.Error("file1.txt not copied")
    }
    if _, err := os.Stat(filepath.Join(expectedDst, "subdir", "file2.txt")); err != nil {
        t.Error("subdir/file2.txt not copied")
    }
}

func TestDelete(t *testing.T) {
    tmpDir := t.TempDir()

    // テストファイル作成
    testFile := filepath.Join(tmpDir, "test.txt")
    os.WriteFile(testFile, []byte("test"), 0644)

    // 削除実行
    if err := Delete(testFile); err != nil {
        t.Fatalf("Delete failed: %v", err)
    }

    // 検証
    if _, err := os.Stat(testFile); !os.IsNotExist(err) {
        t.Error("File should be deleted")
    }
}

func TestMoveFile(t *testing.T) {
    tmpDir := t.TempDir()

    // ソースファイル
    srcFile := filepath.Join(tmpDir, "source.txt")
    os.WriteFile(srcFile, []byte("test"), 0644)

    // 宛先ディレクトリ
    dstDir := filepath.Join(tmpDir, "dest")
    os.Mkdir(dstDir, 0755)

    // 移動実行
    if err := MoveFile(srcFile, dstDir); err != nil {
        t.Fatalf("MoveFile failed: %v", err)
    }

    // 検証
    if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
        t.Error("Source file should not exist")
    }

    dstFile := filepath.Join(dstDir, "source.txt")
    if _, err := os.Stat(dstFile); err != nil {
        t.Error("Destination file should exist")
    }
}
```

#### 4.2 モデルへのファイル操作統合

**`internal/ui/model.go` の更新**:

```go
// Update にファイル操作を追加
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ダイアログの結果処理
    if result, ok := msg.(dialogResultMsg); ok {
        prevDialog := m.dialog
        m.dialog = nil

        // 削除確認の結果
        if result.result.Confirmed {
            if _, ok := prevDialog.(*ConfirmDialog); ok {
                // 削除実行
                entry := m.activePane().SelectedEntry()
                if entry != nil && !entry.IsParentDir() {
                    fullPath := filepath.Join(m.activePane().Path(), entry.Name)

                    if err := fs.Delete(fullPath); err != nil {
                        // エラーダイアログを表示
                        m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
                    } else {
                        // ディレクトリを再読み込み
                        m.activePane().LoadDirectory()
                    }
                }
            }
        }

        return m, nil
    }

    // ダイアログが開いている場合はダイアログに処理を委譲
    if m.dialog != nil {
        var cmd tea.Cmd
        m.dialog, cmd = m.dialog.Update(msg)
        return m, cmd
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // ... (既存のコード)

    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", KeyQuit:
            return m, tea.Quit

        case KeyHelp:
            m.dialog = NewHelpDialog()
            return m, nil

        case KeyCopy:
            // コピー操作
            entry := m.activePane().SelectedEntry()
            if entry != nil && !entry.IsParentDir() {
                srcPath := filepath.Join(m.activePane().Path(), entry.Name)
                dstPath := m.inactivePane().Path()

                if err := fs.Copy(srcPath, dstPath); err != nil {
                    m.dialog = NewErrorDialog(fmt.Sprintf("Failed to copy: %v", err))
                } else {
                    // 対象ペインを再読み込み
                    m.inactivePane().LoadDirectory()
                }
            }
            return m, nil

        case KeyMove:
            // 移動操作
            entry := m.activePane().SelectedEntry()
            if entry != nil && !entry.IsParentDir() {
                srcPath := filepath.Join(m.activePane().Path(), entry.Name)
                dstPath := m.inactivePane().Path()

                if err := fs.MoveFile(srcPath, dstPath); err != nil {
                    m.dialog = NewErrorDialog(fmt.Sprintf("Failed to move: %v", err))
                } else {
                    // 両ペインを再読み込み
                    m.activePane().LoadDirectory()
                    m.inactivePane().LoadDirectory()
                }
            }
            return m, nil

        case KeyDelete:
            // 削除確認ダイアログを表示
            entry := m.activePane().SelectedEntry()
            if entry != nil && !entry.IsParentDir() {
                m.dialog = NewConfirmDialog(
                    "Delete file?",
                    entry.DisplayName(),
                )
            }
            return m, nil

        // ... (既存のナビゲーションコード)
        }
    }

    return m, nil
}
```

**テスト**:
```bash
go build -o duofm ./cmd/duofm
./duofm
# c キーでファイルをコピー
# m キーでファイルを移動
# d キーで削除確認ダイアログが表示されることを確認
```

---

### Phase 5: 統合テストと仕上げ (1-2日)

#### 5.1 統合テストの実装

**ファイル**:
```
test/
└── integration_test.go
```

**実装手順**:

```go
// test/integration_test.go
package test

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/sakura/duofm/internal/ui"
)

// testModel はテスト用のモデル実行
func testModel(t *testing.T, inputs []tea.Msg) ui.Model {
    m := ui.NewModel()

    // ウィンドウサイズを設定
    m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40}).(ui.Model, tea.Cmd)

    // 各入力を処理
    for _, input := range inputs {
        var cmd tea.Cmd
        m, cmd = m.Update(input).(ui.Model, tea.Cmd)

        // コマンドが返された場合は実行
        if cmd != nil {
            msg := cmd()
            if msg != nil {
                m, _ = m.Update(msg).(ui.Model, tea.Cmd)
            }
        }
    }

    return m
}

func TestNavigation(t *testing.T) {
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)

    // テストファイルを作成
    os.Mkdir("testdir", 0755)
    os.WriteFile("file1.txt", []byte("test"), 0644)

    inputs := []tea.Msg{
        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, // 下に移動
        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, // 上に移動
    }

    m := testModel(t, inputs)

    // ビューがレンダリングされることを確認
    view := m.View()
    if view == "" {
        t.Error("View should not be empty")
    }
}

func TestCopyOperation(t *testing.T) {
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)

    // ソースディレクトリとファイル
    os.Mkdir("left", 0755)
    os.Mkdir("right", 0755)
    testFile := filepath.Join("left", "test.txt")
    os.WriteFile(testFile, []byte("content"), 0644)

    // モデルを手動で設定（テスト用）
    // 実際の統合テストでは E2E テストフレームワークを使用する方が良い

    t.Skip("Requires E2E testing framework")
}

func TestHelpDialog(t *testing.T) {
    inputs := []tea.Msg{
        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}, // ヘルプを開く
    }

    m := testModel(t, inputs)
    view := m.View()

    // ヘルプダイアログが表示されていることを確認
    if !contains(view, "Keybindings") {
        t.Error("Help dialog should be visible")
    }
}

func contains(s, substr string) bool {
    return len(s) > 0 && len(substr) > 0 &&
           len(s) >= len(substr) &&
           stringContains(s, substr)
}

func stringContains(s, substr string) bool {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return true
        }
    }
    return false
}
```

#### 5.2 Makefile の更新

**`Makefile` の作成/更新**:

```makefile
.PHONY: build test clean install run

BINARY_NAME=duofm
BINARY_PATH=./cmd/duofm
BUILD_DIR=.
GO=go

build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

test:
	$(GO) test -v ./...

test-coverage:
	$(GO) test -v -cover ./...
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out coverage.html

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

run: build
	./$(BINARY_NAME)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

deps:
	$(GO) mod download
	$(GO) mod tidy

.DEFAULT_GOAL := build
```

#### 5.3 README とドキュメントの更新

**タスク**:
- `README.md` の更新（ビルド手順、使い方）
- `CONTRIBUTING.md` の確認
- コード内のコメント追加

#### 5.4 最終検証チェックリスト

**機能テスト**:
- [ ] アプリケーションが起動する
- [ ] 左ペインに現在のディレクトリが表示される
- [ ] 右ペインにホームディレクトリが表示される
- [ ] hjkl キーでナビゲーションできる
- [ ] Enter でディレクトリに入れる
- [ ] c キーでファイルをコピーできる
- [ ] m キーでファイルを移動できる
- [ ] d キーで削除確認が表示される
- [ ] y で削除が実行される
- [ ] n で削除がキャンセルされる
- [ ] ? でヘルプが表示される
- [ ] Esc でダイアログが閉じる
- [ ] q でアプリケーションが終了する

**エラーハンドリングテスト**:
- [ ] 読み取り権限のないディレクトリでエラーダイアログが表示される
- [ ] 削除権限のないファイルでエラーダイアログが表示される
- [ ] エラーダイアログを閉じてもアプリケーションが動作し続ける

**UI/UXテスト**:
- [ ] ディレクトリが先に表示される
- [ ] ディレクトリ名の末尾に / が付く
- [ ] 親ディレクトリが ../ と表示される
- [ ] ホームディレクトリが ~ と表示される
- [ ] ステータスバーに位置情報が表示される
- [ ] カーソル位置が視覚的に識別できる
- [ ] アクティブペインが識別できる

---

## ファイル構成（最終版）

```
duofm/
├── cmd/
│   └── duofm/
│       └── main.go              # エントリーポイント
│
├── internal/
│   ├── ui/
│   │   ├── model.go             # Bubble Tea メインモデル
│   │   ├── pane.go              # ペインコンポーネント
│   │   ├── dialog.go            # ダイアログインターフェース
│   │   ├── confirm_dialog.go    # 確認ダイアログ
│   │   ├── error_dialog.go      # エラーダイアログ
│   │   ├── help_dialog.go       # ヘルプダイアログ
│   │   ├── keys.go              # キーバインディング定義
│   │   ├── styles.go            # Lip Gloss スタイル
│   │   └── pane_test.go         # ペインのテスト
│   │
│   └── fs/
│       ├── types.go             # 型定義
│       ├── reader.go            # ディレクトリ読み込み
│       ├── sort.go              # ソート処理
│       ├── operations.go        # ファイル操作
│       ├── reader_test.go       # 読み込みのテスト
│       ├── sort_test.go         # ソートのテスト
│       └── operations_test.go   # 操作のテスト
│
├── test/
│   └── integration_test.go      # 統合テスト
│
├── doc/
│   ├── CONTRIBUTING.md          # 貢献ガイド
│   ├── specification/
│   │   └── SPEC.md              # 全体仕様
│   └── tasks/
│       └── ui-design/
│           ├── SPEC.md          # UI設計仕様
│           ├── IMPLEMENTATION.md # 本ドキュメント
│           └── 要件定義書.md
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── CLAUDE.md
```

---

## 実装スケジュール

### Week 1
- **Day 1-2**: Phase 1 (基本構造とセットアップ)
- **Day 3-5**: Phase 2 (ペインコンポーネント)

### Week 2
- **Day 1-2**: Phase 3 (ダイアログシステム)
- **Day 3-5**: Phase 4 (ファイル操作)

### Week 3
- **Day 1-2**: Phase 5 (統合テストと仕上げ)
- **Day 3**: バッファ（予備日）

**合計見積もり**: 約2-3週間

---

## リスク評価と対策

### 高リスク項目

#### 1. Bubble Tea の学習曲線
**リスク**: Bubble Tea の Model-Update-View パターンに慣れていない場合、実装に時間がかかる

**対策**:
- 公式チュートリアルを先に完了する
- 小さなサンプルアプリを作成して理解を深める
- 公式サンプルコード（list-simple, simple-list など）を参照する

**参考リソース**:
- https://github.com/charmbracelet/bubbletea/tree/master/tutorials
- https://github.com/charmbracelet/bubbletea/tree/master/examples

#### 2. ファイル操作のエッジケース
**リスク**: 権限エラー、シンボリックリンク、特殊ファイルなどの処理が複雑

**対策**:
- 各ファイル操作関数で適切なエラーハンドリングを実装
- テストケースで各種エッジケースをカバー
- エラーメッセージをユーザーフレンドリーにする

#### 3. TUI のテスト
**リスク**: TUI の自動テストが難しい

**対策**:
- ビジネスロジックを UI から分離する
- fs パッケージの単体テストを充実させる
- 手動テストチェックリストを作成
- 将来的に E2E テストフレームワーク（expect など）の導入を検討

### 中リスク項目

#### 4. パフォーマンス（大量ファイル）
**リスク**: 10,000+ ファイルのディレクトリで動作が遅くなる可能性

**対策**:
- MVP では基本的な実装に集中
- Phase 2 以降でページネーションや仮想スクロールを検討
- プロファイリングツール（pprof）で計測

#### 5. ターミナル互換性
**リスク**: 異なるターミナルエミュレータで表示が崩れる可能性

**対策**:
- 主要なターミナル（iTerm2, GNOME Terminal, Windows Terminal など）でテスト
- Bubble Tea が互換性を提供してくれることを信頼
- 最小ターミナルサイズ（80x24）を要件に含める

---

## テスト戦略

### 単体テスト (Unit Tests)

**対象**:
- `internal/fs` パッケージの全関数
- `internal/ui/pane.go` のロジック部分

**手法**:
- テーブル駆動テスト
- 一時ディレクトリ（`t.TempDir()`）を使用
- モックファイルシステムは使わず、実際のファイル操作でテスト

**カバレッジ目標**: 70%+

### 統合テスト (Integration Tests)

**対象**:
- Bubble Tea モデル全体の動作
- ダイアログのフロー

**手法**:
- メッセージを順次送信してモデルの状態を検証
- ビューの文字列に期待する要素が含まれているか確認

**カバレッジ目標**: 主要フローをカバー

### 手動テスト (Manual Tests)

**対象**:
- UI/UX の視覚的な確認
- 仕様書のテストシナリオ（Scenario 1-26）

**手法**:
- チェックリストに基づいて手動で実行
- 複数のターミナルエミュレータでテスト
- スクリーンショットを記録

---

## 追加の検討事項

### オープンクエスチョンへの回答案

#### Q1: 大量ファイル（10,000+）のパフォーマンス最適化戦略は?
**A**: MVP では基本実装に集中。Phase 2 以降で以下を検討:
- 仮想スクロール（表示範囲のみレンダリング）
- 遅延読み込み（ページネーション）
- バックグラウンドでの段階的読み込み

#### Q2: カラースキームは?
**A**: MVP ではシンプルな配色（青/グレー/赤）を使用。Phase 3 で設定ファイルによるカスタマイズを検討。

#### Q3: MVP にファイルサイズ表示を含めるべきか?
**A**: MVP では見送り。Phase 2 で実装する。理由:
- MVP の目標は「動作する最小限の機能」
- ファイルサイズ表示は便利だが必須ではない
- 実装は簡単なので Phase 2 で素早く追加できる

#### Q4: TUI コンポーネントのテスト戦略は?
**A**:
- ビジネスロジックとUIを分離
- ロジック部分は単体テストでカバー
- UI 部分は手動テストと簡易的な統合テスト
- 将来的に `expect` ベースの E2E テストを検討

#### Q5: 統合テストのアプローチは?
**A**:
- 基本: Bubble Tea のメッセージベースでのテスト
- 発展: GitHub Actions で自動化
- 将来: E2E テストフレームワークの導入

---

## 成功基準

### MVP 完成の定義

以下の全てが満たされた場合、MVP は完成とみなす:

1. **機能要件**:
   - [ ] デュアルペイン表示
   - [ ] hjkl ナビゲーション
   - [ ] Enter でディレクトリに入る
   - [ ] c でファイルコピー
   - [ ] m でファイル移動
   - [ ] d で削除（確認付き）
   - [ ] ? でヘルプ表示
   - [ ] q で終了

2. **品質要件**:
   - [ ] 全ての単体テストが通る
   - [ ] 手動テストシナリオ（1-26）が全て通る
   - [ ] エラーハンドリングが適切に機能する

3. **ドキュメント要件**:
   - [ ] README.md が更新されている
   - [ ] コードにコメントが付いている
   - [ ] ビルド手順が動作する

4. **パフォーマンス要件**:
   - [ ] 通常のディレクトリ（<1000ファイル）で快適に動作
   - [ ] キー入力に遅延がない

---

## 次のステップ（MVP 完了後）

### Phase 2 の優先順位

1. **ファイル詳細表示**: サイズ、日時、権限
2. **ファイル開く機能**: less, vim との連携
3. **複数ファイルマーク**: Space キーでマーク
4. **ソート機能**: s キーでトグル
5. **隠しファイルトグル**: Ctrl+H

### Phase 3 の優先順位

1. **検索/フィルター**: / キーで検索
2. **高度なナビゲーション**: gg, G, Ctrl+D/U
3. **進捗表示**: 大きなファイルのコピー時
4. **設定ファイル**: ~/.config/duofm/config.toml
5. **ディレクトリ記憶**: 最後に開いたディレクトリを保存

---

## 参考資料

### Bubble Tea リソース
- [公式ドキュメント](https://github.com/charmbracelet/bubbletea)
- [チュートリアル](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [サンプル集](https://github.com/charmbracelet/bubbletea/tree/master/examples)

### Lip Gloss リソース
- [公式ドキュメント](https://github.com/charmbracelet/lipgloss)
- [レイアウト例](https://github.com/charmbracelet/lipgloss/tree/master/examples)

### Go リソース
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)

### 参考実装
- [lazygit](https://github.com/jesseduffield/lazygit) - ヘルプ画面の参考
- [ranger](https://github.com/ranger/ranger) - Vim スタイルファイルマネージャー
- [nnn](https://github.com/jarun/nnn) - 高速ファイルマネージャー

---

## まとめ

本実装計画は、duofm の MVP を約2-3週間で完成させることを目標としています。

**重要なポイント**:
1. **段階的な実装**: Phase 1 → Phase 5 の順に進める
2. **テスト駆動**: 各フェーズでテストを書く
3. **Bubble Tea パターン**: Model-Update-View を理解する
4. **エラーハンドリング**: ユーザーフレンドリーなエラーメッセージ
5. **ドキュメント**: コードと同時にドキュメントを更新

各フェーズを完了するたびに、動作確認を行い、問題があれば早期に修正します。

MVP 完成後は、Phase 2/3 の機能を優先順位に従って追加していきます。
