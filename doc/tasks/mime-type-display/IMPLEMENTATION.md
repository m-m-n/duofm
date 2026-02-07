# Implementation Plan: MIME Type Display in Status Bar

## Overview

ステータスバーのN/M位置表示の右側に、カーソル下エントリのMIMEタイプ/種別を色付きで表示する。

## Implementation Steps

### Step 1: ヘルパー関数の追加 (internal/ui/model_view.go)

**目的:** エントリからMIMEタイプ表示用のラベルと色を返すヘルパー関数を追加する。

**ファイル:** `internal/ui/model_view.go`

**変更内容:**

1. import に `"github.com/sakura/duofm/internal/config"` と `"github.com/sakura/duofm/internal/fs"` を追加
2. 以下のヘルパー関数を追加:

```go
// entryTypeLabel はエントリの種別ラベルと表示色を返す。
// 判定優先順位: シンボリックリンク > ディレクトリ > 通常ファイル（MIMEタイプ）
func (m Model) entryTypeLabel(entry *fs.FileEntry) (label string, fg lipgloss.Color) {
    if entry == nil {
        return "", m.theme.StatusFg
    }
    switch {
    case entry.IsSymlink:
        return "SymbolicLink", m.theme.SymlinkFg
    case entry.IsDir:
        return "Directory", m.theme.DirectoryFg
    default:
        return config.GetMIMEType(entry.Name), m.theme.StatusFg
    }
}
```

**テスト:** Step 3 で作成するテストで検証する。

---

### Step 2: renderStatusBar の修正 (internal/ui/model_view.go)

**目的:** ステータスバーにMIMEタイプ表示を追加する。

**ファイル:** `internal/ui/model_view.go`

**変更内容:**

`renderStatusBar()` メソッドの通常表示部分（ステータスメッセージがない場合のブロック、行336〜365）を修正する。

現在のレイアウト:
```
| {posInfo}                    {hints} |
```

変更後のレイアウト:
```
| {posInfo} {coloredTypeInfo}          {hints} |
```

具体的な変更:

1. `posInfo` 生成の後で、`activePane.SelectedEntry()` を呼んでカーソル位置のエントリを取得
2. エントリが nil でなければ `entryTypeLabel()` でラベルと色を取得
3. `[{label}]` 形式の文字列を生成
4. lipgloss でラベル部分に前景色を適用（背景はステータスバーの `StatusBg` を維持）
5. パディング計算にMIMEタイプ表示の幅を加算
6. 幅が足りない場合はMIMEタイプ表示を省略

変更後のコードイメージ:

```go
activePane := m.getActivePane()

// 選択位置情報
posInfo := fmt.Sprintf("%d/%d", activePane.cursor+1, len(activePane.entries))

// MIMEタイプ / 種別情報
typeInfo := ""
typeInfoWidth := 0
if entry := activePane.SelectedEntry(); entry != nil {
    label, fg := m.entryTypeLabel(entry)
    typeInfo = lipgloss.NewStyle().
        Foreground(fg).
        Background(m.theme.StatusBg).
        Render(fmt.Sprintf(" [%s]", label))
    typeInfoWidth = runewidth.StringWidth(fmt.Sprintf(" [%s]", label))
}

// キーヒント（動的に変更）
hints := "?:help q:quit"
if activePane != nil && activePane.CanToggleMode() {
    hints = "i:info " + hints
}

// スペースで埋める（typeInfoWidth を加算）
padding := m.width - runewidth.StringWidth(posInfo) - typeInfoWidth - runewidth.StringWidth(hints) - 4
if padding < 0 {
    // 幅が足りない場合はMIMEタイプ表示を省略
    typeInfo = ""
    typeInfoWidth = 0
    padding = m.width - runewidth.StringWidth(posInfo) - runewidth.StringWidth(hints) - 4
    if padding < 0 {
        padding = 0
    }
}

// ステータスバーの組み立て
// typeInfo は既に lipgloss でレンダリング済み（ANSI色付き）なので、
// 全体を style.Render() に渡すと色が上書きされる。
// そのため、左部分・typeInfo・右部分を個別にレンダリングして結合する。
statusBg := lipgloss.Color("240")
statusFg := lipgloss.Color("15")

leftPart := lipgloss.NewStyle().
    Background(statusBg).
    Foreground(statusFg).
    Render(fmt.Sprintf(" %s", posInfo))

rightPart := lipgloss.NewStyle().
    Background(statusBg).
    Foreground(statusFg).
    Render(fmt.Sprintf("%s%s ", strings.Repeat(" ", padding), hints))

statusBar := leftPart + typeInfo + rightPart

// Width を指定して幅を揃える
return lipgloss.NewStyle().
    Width(m.width).
    Render(statusBar)
```

**重要:** 既存の `renderStatusBar` では `style.Render(statusBar)` で全体にスタイルを適用していたが、MIMEタイプ部分は個別の前景色を持つため、全体を一括レンダリングすると色が上書きされてしまう。代わりに、左部分（posInfo）・MIMEタイプ部分（色付き済み）・右部分（パディング+hints）を個別にレンダリングして結合する方式に変更する。

**注意点:**
- `typeInfo` は lipgloss でレンダリング済みの文字列を使用する（色付き）
- パディング計算には `typeInfoWidth`（ANSI エスケープなしの表示幅）を使用する
- MIMEタイプ部分の背景色は `m.theme.StatusBg` を使用する
- ステータスバー全体のスタイルは既存のハードコード値（`Background: "240"`, `Foreground: "15"`）を維持する（既存設計を踏襲。デフォルトテーマの `StatusBg=240`, `StatusFg=15` と一致する）
- `fmt.Sprintf` で生成する文字列（パディング部分等）はステータスバー全体の `Foreground("15")` が適用される。MIMEタイプ部分のみ `typeInfo`（色適用済み文字列）を直接埋め込むことで個別色が反映される

**テスト:** Step 3 で作成するテストで検証する。

---

### Step 3: ユニットテストの作成 (internal/ui/model_view_test.go)

**目的:** entryTypeLabel 関数とステータスバーのMIMEタイプ表示をテストする。

**ファイル:** `internal/ui/model_view_test.go`（新規作成）

**テストケース:**

1. **entryTypeLabel のテスト:**
   - ディレクトリエントリ → ラベル `"Directory"`, 色 `DirectoryFg`
   - シンボリックリンクエントリ → ラベル `"SymbolicLink"`, 色 `SymlinkFg`
   - `.html` ファイル → ラベル `"text/html"`, 色 `StatusFg`
   - `.txt` ファイル → ラベル `"text/plain"`, 色 `StatusFg`
   - 拡張子なしファイル → ラベル `"application/octet-stream"`, 色 `StatusFg`
   - 親ディレクトリ（`..`）→ ラベル `"Directory"`, 色 `DirectoryFg`
   - シンボリックリンク先がディレクトリ（IsSymlink=true, IsDir=true）→ ラベル `"SymbolicLink"`, 色 `SymlinkFg`
   - nil エントリ → ラベル `""`, 色 `StatusFg`

2. **renderStatusBar のテスト:**
   - ステータスバー文字列に `[text/html]` が含まれる（通常ファイル）
   - ステータスバー文字列に `[Directory]` が含まれる（ディレクトリ）
   - ステータスバー文字列に `[SymbolicLink]` が含まれる（シンボリックリンク）
   - 空ディレクトリ（エントリ0件）ではMIMEタイプ表示なし
   - ステータスメッセージ表示中はMIMEタイプ表示なし

**テスト方針:**
- `entryTypeLabel` はテーブル駆動テストを使用
- Model のテストには最小限のモック（DefaultTheme でテーマを設定、Pane にエントリをセット）

---

### Step 4: ビルド確認・既存テスト通過確認

**目的:** 変更がコンパイルでき、既存テストに影響しないことを確認する。

**コマンド:**
```bash
cd /home/sakura/go/src/duofm && go build ./...
cd /home/sakura/go/src/duofm && go test ./internal/ui/... ./internal/config/...
```

## File Change Summary

| ファイル | 変更種別 | 内容 |
|----------|----------|------|
| `internal/ui/model_view.go` | 修正 | import 追加（`config`, `fs`）、`entryTypeLabel()` 追加、`renderStatusBar()` 修正 |
| `internal/ui/model_view_test.go` | 新規作成 | `entryTypeLabel` と `renderStatusBar` のユニットテスト |

## Dependencies

- `internal/config/mime.go` - `GetMIMEType()` を呼び出す（変更不要）
- `internal/ui/theme.go` - `DirectoryFg`, `SymlinkFg`, `StatusFg` を参照（変更不要）
- `internal/fs/types.go` - `FileEntry` の `IsDir`, `IsSymlink`, `Name` を参照（変更不要）

## Risk Assessment

- **低リスク:** 既存の `renderStatusBar()` の通常表示ブロックのみの変更。ステータスメッセージ表示は触れない。
- **低リスク:** 新しいヘルパー関数は独立しており、既存コードへの影響はない。
- **注意点:** ステータスバーのレンダリング方式が「全体一括 `style.Render()`」から「パーツ個別レンダリング+結合」に変更される。これにより MIMEタイプ部分の色が正しく表示される。既存のステータスバーの見た目（posInfo + padding + hints）には影響しない。
- **注意点:** 既存の `renderStatusBar` はステータスバーの色をハードコード（`"240"`, `"15"`）で指定しているが、MIMEタイプ部分は `m.theme.StatusBg` を使用する。デフォルトテーマでは値が一致するが、カスタムテーマの場合も正しく動作する。
