# Implementation Plan: Multi-file Marking

## Overview

複数ファイルのマーク（選択）機能を実装し、マークしたファイルに対して一括でコピー・移動・削除操作を行えるようにする。

## Objectives

- Spaceキーでファイルをマーク/解除できる
- マークされたファイルを視覚的に区別できる（背景色）
- ヘッダーにマーク数と合計サイズを表示
- c/m/dキーでマークファイルに対して一括操作できる
- コンテキストメニューもマークファイルに対応

## Prerequisites

- 既存のファイル操作（コピー、移動、削除）が実装済み
- 上書き確認ダイアログが実装済み
- コンテキストメニューが実装済み

## Architecture Overview

```
Pane (internal/ui/pane.go)
  ├── markedFiles map[string]bool    // マーク状態管理
  ├── ToggleMark() bool              // マーク切り替え
  ├── ClearMarks()                   // マーククリア
  ├── IsMarked(filename string) bool // マーク状態確認
  ├── GetMarkedFiles() []string      // マークファイル取得
  ├── CalculateMarkInfo() MarkInfo   // マーク統計情報
  └── View() with mark highlighting  // マーク表示

Model (internal/ui/model.go)
  ├── handleMarkToggle()             // Spaceキー処理
  ├── handleBatchCopy()              // 一括コピー
  ├── handleBatchMove()              // 一括移動
  └── handleBatchDelete()            // 一括削除
```

## Implementation Phases

### Phase 1: Pane Mark Management

**Goal**: Pane構造体にマーク管理機能を追加

**Files to Create/Modify**:
- `internal/ui/pane.go` - マーク関連メソッド追加
- `internal/ui/pane_mark_test.go` - 単体テスト（新規）

**Implementation Steps**:

1. **MarkInfo構造体の定義**
   ```go
   // MarkInfo holds mark statistics
   type MarkInfo struct {
       Count     int   // Number of marked files
       TotalSize int64 // Total size in bytes
   }
   ```

2. **Pane構造体にmarkedFilesフィールドを追加**
   ```go
   type Pane struct {
       // ... existing fields
       markedFiles map[string]bool  // key: filename, value: marked state
   }
   ```

3. **NewPaneでmarkedFilesを初期化**
   ```go
   pane := &Pane{
       // ... existing
       markedFiles: make(map[string]bool),
   }
   ```

4. **マーク操作メソッドの実装**
   - `ToggleMark() bool`: カーソル位置のファイルのマークを切り替え
     - 親ディレクトリ(..)の場合はfalseを返す
     - マーク切り替え後、trueを返す
   - `ClearMarks()`: すべてのマークをクリア
   - `IsMarked(filename string) bool`: 指定ファイルがマークされているか
   - `GetMarkedFiles() []string`: マークされたファイル名のリストを返す
   - `CalculateMarkInfo() MarkInfo`: マーク数と合計サイズを計算

5. **LoadDirectoryでmarkedFilesをクリア**
   - ディレクトリ変更時にマークをクリアする

**Testing**:
- ToggleMarkの基本動作
- 親ディレクトリでのToggleMark
- ClearMarksの動作
- IsMarkedの正確性
- GetMarkedFilesの結果
- CalculateMarkInfoの計算（ファイルサイズ、ディレクトリは0）

**Dependencies**: なし

---

### Phase 2: Visual Display

**Goal**: マークされたファイルを背景色で視覚的に区別

**Files to Modify**:
- `internal/ui/pane.go` - formatEntry, formatEntryDimmed, renderHeaderLine2

**Implementation Steps**:

1. **マーク用カラー定数の定義**
   ```go
   var (
       // Mark background colors
       markBgColorActive   = lipgloss.Color("136") // Yellow for active pane
       markBgColorInactive = lipgloss.Color("94")  // Dark yellow for inactive pane

       // Cursor + Mark combined colors
       cursorMarkBgColorActive   = lipgloss.Color("30") // Cyan for active pane
       cursorMarkBgColorInactive = lipgloss.Color("23") // Dark cyan for inactive pane
   )
   ```

2. **formatEntryを更新**
   - 既存のカーソルスタイル処理に加えて、マーク状態のスタイルを追加
   - 4つの状態を処理:
     - 通常（マークなし、カーソルなし）
     - カーソルのみ（既存）
     - マークのみ（黄色背景）
     - カーソル + マーク（シアン背景）

3. **formatEntryDimmedを更新**
   - ダイアログ表示時もマーク状態を視認可能に
   - dimmedスタイルでマークを表現（薄いハイライト）

4. **renderHeaderLine2を更新**
   - `CalculateMarkInfo()`を呼び出して実際のマーク情報を表示
   - 既存のプレースホルダー（`markedCount := 0`）を置き換え

**Testing**:
- 手動視覚テスト: マークファイルの背景色確認
- アクティブ/インアクティブペインでの色の違い
- カーソル + マーク状態の表示

**Dependencies**: Phase 1

---

### Phase 3: Key Handling

**Goal**: Spaceキーでマーク切り替え、ディレクトリ変更時にマーククリア

**Files to Modify**:
- `internal/ui/keys.go` - KeyMark定数追加
- `internal/ui/model.go` - Spaceキーハンドラ追加

**Implementation Steps**:

1. **keys.goにKeyMark定数を追加**
   ```go
   KeyMark = " " // Space key for marking files
   ```

2. **model.goにSpaceキーハンドラを追加**
   ```go
   case KeyMark:
       return m.handleMarkToggle()
   ```

3. **handleMarkToggleメソッドの実装**
   ```go
   func (m *Model) handleMarkToggle() (tea.Model, tea.Cmd) {
       activePane := m.getActivePane()
       if activePane.ToggleMark() {
           // マーク切り替え成功 → カーソルを下に移動
           activePane.MoveCursorDown()
       }
       // 親ディレクトリの場合は何もしない
       return m, nil
   }
   ```

4. **ディレクトリ変更時のマーククリア確認**
   - `LoadDirectory()`内で既に`markedFiles`をクリアするので追加実装不要
   - 確認: `EnterDirectory`, `MoveToParent`, `ChangeDirectory`

**Testing**:
- Spaceキーでマークがトグルされる
- マーク後にカーソルが下に移動
- 親ディレクトリでSpaceを押しても何も起こらない
- 最後のファイルでSpaceを押すとマークされるがカーソルは移動しない
- ディレクトリ変更時にマークがクリアされる

**Dependencies**: Phase 1, Phase 2

---

### Phase 4: Batch File Operations

**Goal**: c/m/dキーでマークファイルに対して一括操作

**Files to Modify**:
- `internal/ui/model.go` - handleCopy, handleMove, handleDelete修正

**Implementation Steps**:

1. **一括操作用の状態管理（model.go）**
   ```go
   type BatchOperation struct {
       Files       []string    // List of source file paths
       CurrentIdx  int         // Current file index
       DestPath    string      // Destination directory
       Operation   string      // "copy" or "move"
       Completed   []string    // Successfully completed files
       Failed      []string    // Failed files
   }

   // Model に追加
   type Model struct {
       // ... existing
       batchOp *BatchOperation  // Active batch operation (nil if none)
   }
   ```

2. **handleCopyの修正**
   ```go
   func (m *Model) handleCopy() tea.Cmd {
       activePane := m.getActivePane()
       markedFiles := activePane.GetMarkedFiles()

       if len(markedFiles) > 0 {
           // マークファイルがある場合 → 一括コピー開始
           return m.startBatchOperation(markedFiles, "copy")
       }

       // マークなし → 既存の単一ファイルコピー
       entry := activePane.SelectedEntry()
       if entry != nil && !entry.IsParentDir() {
           // ... existing single file copy logic
       }
       return nil
   }
   ```

3. **startBatchOperationの実装**
   ```go
   func (m *Model) startBatchOperation(files []string, operation string) tea.Cmd {
       srcDir := m.getActivePane().Path()
       destDir := m.getInactivePane().Path()

       // Build full paths
       fullPaths := make([]string, len(files))
       for i, f := range files {
           fullPaths[i] = filepath.Join(srcDir, f)
       }

       m.batchOp = &BatchOperation{
           Files:     fullPaths,
           CurrentIdx: 0,
           DestPath:  destDir,
           Operation: operation,
       }

       // Process first file
       return m.processBatchFile()
   }
   ```

4. **processBatchFileの実装**
   - 現在のファイルの競合をチェック
   - 競合なし → 即座に実行して次へ
   - 競合あり → 上書き確認ダイアログを表示

5. **上書き確認結果のバッチ対応**
   - `overwriteDialogResultMsg`処理を拡張
   - 上書き/リネーム → 処理して次へ
   - キャンセル → バッチ全体を中断

6. **バッチ完了処理**
   - すべてのファイル処理後にマークをクリア
   - 両ペインを再読み込み

7. **handleMoveの同様の修正**

8. **handleDeleteの修正**
   - マークファイルがある場合は確認ダイアログに件数を表示
   - 例: "Delete 3 files?"
   - 確認後、すべてのマークファイルを削除

**Testing**:
- マークファイルがある場合、cキーで一括コピー
- マークファイルがない場合、既存の単一ファイルコピー
- 上書き確認が各ファイルごとに表示される
- キャンセルで残りのファイルが処理されない
- 完了後にマークがクリアされる
- 削除の確認ダイアログに正しいメッセージが表示される

**Dependencies**: Phase 1, Phase 3

---

### Phase 5: Context Menu Integration

**Goal**: コンテキストメニューのコピー/移動/削除をマークファイルに対応

**Files to Modify**:
- `internal/ui/context_menu_dialog.go` - マーク対応
- `internal/ui/model.go` - contextMenuResultMsg処理

**Implementation Steps**:

1. **ContextMenuDialogにマーク情報を渡す**
   ```go
   func NewContextMenuDialogWithMarks(entry *fs.FileEntry, sourcePath, destPath string, pane *Pane, markedFiles []string) *ContextMenuDialog
   ```

2. **buildMenuItemsをマーク対応**
   - マークファイルがある場合:
     - "Copy to other pane" → "Copy 3 files to other pane"
     - "Move to other pane" → "Move 3 files to other pane"
     - "Delete" → "Delete 3 files"
   - マークファイルがない場合: 既存動作

3. **model.goのコンテキストメニュー表示を修正**
   ```go
   case KeyContextMenu:
       activePane := m.getActivePane()
       entry := activePane.SelectedEntry()
       markedFiles := activePane.GetMarkedFiles()

       if entry != nil && !entry.IsParentDir() {
           m.dialog = NewContextMenuDialogWithMarks(
               entry,
               activePane.Path(),
               m.getInactivePane().Path(),
               activePane,
               markedFiles,
           )
       }
   ```

4. **コンテキストメニューからの操作をバッチ処理と統合**
   - `contextMenuResultMsg`処理で、マークファイルがある場合はバッチ操作を開始

**Testing**:
- マークファイルがある場合、コンテキストメニューに件数が表示される
- コンテキストメニューから一括操作が正しく動作する
- マークがない場合は既存動作

**Dependencies**: Phase 4

---

### Phase 6: Edge Cases and Polish

**Goal**: エッジケースの処理と品質向上

**Files to Modify**:
- `internal/ui/pane.go`
- `internal/ui/model.go`

**Implementation Steps**:

1. **隠しファイル表示切り替え時**
   - 非表示になったファイルのマークをクリア
   ```go
   func (p *Pane) ToggleHidden() {
       // ... existing
       // 非表示になったファイルのマークをクリア
       if !p.showHidden {
           for filename := range p.markedFiles {
               if strings.HasPrefix(filename, ".") {
                   delete(p.markedFiles, filename)
               }
           }
       }
   }
   ```

2. **フィルタ適用時**
   - マークは維持（仕様通り）
   - フィルタで非表示になったファイルもマーク状態を保持

3. **リフレッシュ（F5）時**
   - 存在しないファイルのマークをクリア
   ```go
   func (p *Pane) Refresh() error {
       // ... existing refresh
       // 存在しないファイルのマークをクリア
       validMarks := make(map[string]bool)
       for _, entry := range p.entries {
           if p.markedFiles[entry.Name] {
               validMarks[entry.Name] = true
           }
       }
       p.markedFiles = validMarks
   }
   ```

4. **パフォーマンス最適化**
   - `CalculateMarkInfo()`のキャッシュ検討
   - 大量のマーク時の描画効率確認

**Testing**:
- 隠しファイルをマークして表示を切り替え → マークがクリア
- フィルタ適用時にマークが維持される
- 外部でファイルを削除してF5 → マークがクリア
- 1000ファイルマーク時のパフォーマンス

**Dependencies**: Phase 1-5

---

## File Structure

```
internal/
├── ui/
│   ├── pane.go              # マーク管理追加（markedFiles, ToggleMark, etc.）
│   ├── pane_mark_test.go    # マーク機能の単体テスト（新規）
│   ├── model.go             # Spaceキー処理、バッチ操作
│   ├── keys.go              # KeyMark定数追加
│   └── context_menu_dialog.go # マーク対応
└── fs/
    └── operations.go        # 変更なし（既存の操作を利用）
```

## Testing Strategy

### Unit Tests (pane_mark_test.go)

```go
func TestToggleMark(t *testing.T)
func TestToggleMarkOnParentDir(t *testing.T)
func TestClearMarks(t *testing.T)
func TestIsMarked(t *testing.T)
func TestGetMarkedFiles(t *testing.T)
func TestCalculateMarkInfo(t *testing.T)
func TestCalculateMarkInfoWithDirectory(t *testing.T)
func TestMarksCleared OnDirectoryChange(t *testing.T)
```

### Integration Tests

- Spaceキーでマークがトグルされる
- マーク後にカーソルが下に移動
- マークファイルへの一括コピー
- マークファイルへの一括移動
- マークファイルへの一括削除
- 上書き確認のキャンセルでバッチ中断
- 操作完了後のマーククリア

### Manual Testing Checklist

- [ ] マークファイルが黄色背景で表示される（アクティブペイン）
- [ ] マークファイルが暗い黄色背景で表示される（インアクティブペイン）
- [ ] カーソル + マークがシアン背景で表示される
- [ ] ヘッダーに正しいマーク数と合計サイズが表示される
- [ ] 大量ファイル（100+）でのマーク操作が遅延なく動作
- [ ] ディレクトリ変更でマークがクリアされる

## Dependencies

### External Libraries
- `github.com/charmbracelet/bubbletea` - TUI framework（既存）
- `github.com/charmbracelet/lipgloss` - Styling（既存）

### Internal Dependencies
- Phase 1 → Phase 2, 3, 4, 5, 6
- Phase 2 → 視覚的な確認可能
- Phase 3 → Phase 4
- Phase 4 → Phase 5
- Phase 5 → 機能完成
- Phase 6 → 品質向上

## Risk Assessment

### Technical Risks

1. **バッチ操作中の状態管理**
   - リスク: 上書き確認ダイアログとバッチ状態の整合性
   - 軽減策: BatchOperation構造体で明確に状態を管理

2. **パフォーマンス（大量マーク時）**
   - リスク: 1000ファイルマーク時の描画遅延
   - 軽減策: map使用で O(1) ルックアップ、必要に応じてキャッシュ

3. **マーク状態の一貫性**
   - リスク: フィルタ/リフレッシュ時のマーク状態不整合
   - 軽減策: 明確なルール定義（仕様書）とテスト

### Implementation Risks

1. **既存コードへの影響**
   - リスク: 既存のコピー/移動/削除動作への影響
   - 軽減策: マークがない場合は完全に既存動作を維持

## Performance Considerations

- `markedFiles`はmapで実装 → O(1)のルックアップ
- `CalculateMarkInfo()`は毎フレーム呼ばれる → 効率的な実装が必要
- 大量マーク時も描画はO(n)（表示行数のみ）

## Security Considerations

- ファイル操作は既存の`fs`パッケージを利用
- パス操作は`filepath.Join`を使用
- 権限エラーは既存のエラーハンドリングに従う

## Open Questions

- [x] 隠しファイル表示切り替え時のマーク動作 → 非表示になったらクリア
- [x] フィルタ適用時のマーク動作 → マークは維持（非表示ファイルも）
- [ ] 全選択/全解除のショートカット追加は必要か？ → 将来の拡張として検討

## Future Enhancements

- 全選択（Ctrl+A）/ 全解除のショートカット
- パターンによる選択（*.txt など）
- 選択反転
- マークファイルのコピー先パス入力

## References

- [SPEC.md](./SPEC.md) - 機能仕様書
- [要件定義書.md](./要件定義書.md) - 要件定義
- [Midnight Commander](https://midnight-commander.org/) - 参考実装
