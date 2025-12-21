# Implementation Plan: Directory Access Error Handling

## Overview

ディレクトリアクセスエラー時（権限エラー、存在しないディレクトリなど）の挙動を修正し、パス表示とファイルリストの整合性を保つ。エラー発生時はパスを更新せず、ステータスバーにエラーメッセージを表示する。

## Objectives

1. ディレクトリナビゲーションのアトミック性を確保（読み込み成功時のみパス更新）
2. エラー時はステータスバーにわかりやすいメッセージを表示
3. エラー後も通常操作を継続可能にする
4. 既存機能との互換性を維持

## Prerequisites

- Bubble Tea フレームワークの理解
- 既存の `directoryLoadCompleteMsg` メッセージングパターン
- 現在の `EnterDirectory()` および `LoadDirectory()` の実装

## Architecture Overview

現在の問題:
```
EnterDirectory() {
    p.path = newPath          // パスが先に更新される
    return p.LoadDirectory()  // 読み込み失敗してもパスは更新済み
}
```

修正後の設計:
```
EnterDirectory() {
    1. 新しいパスを計算
    2. 以前のパスをバックアップ
    3. パスを暫定的に更新
    4. 非同期でディレクトリ読み込み開始
}

handleDirectoryLoadComplete() {
    if error {
        パスを元に戻す
        ステータスバーにエラー表示
    } else {
        エントリを更新
    }
}
```

## Implementation Phases

### Phase 1: Pane構造体の拡張

**Goal**: パス復元のためのバックアップ機構を追加

**Files to Create/Modify**:
- `internal/ui/pane.go` - Pane構造体に `pendingPath` フィールドを追加

**Implementation Steps**:

1. Pane構造体にフィールドを追加:
   ```go
   type Pane struct {
       // ... 既存フィールド ...
       pendingPath string // 読み込み中の暫定パス（エラー時に元に戻す）
   }
   ```

2. パス復元用のヘルパーメソッドを追加:
   ```go
   // restorePreviousPath は読み込み失敗時に前のパスに復元する
   func (p *Pane) restorePreviousPath() {
       if p.previousPath != "" {
           p.path = p.previousPath
           p.pendingPath = ""
       }
   }
   ```

**Dependencies**: なし

**Testing**:
- `restorePreviousPath()` が正しくパスを復元することを確認

**Estimated Effort**: Small

---

### Phase 2: ステータスバーメッセージ機能

**Goal**: ステータスバーにエラーメッセージを表示し、5秒後に自動クリアする機能を追加

**Files to Create/Modify**:
- `internal/ui/model.go` - Model構造体にステータスメッセージフィールドを追加
- `internal/ui/messages.go` - 新しいメッセージ型を追加

**Implementation Steps**:

1. `messages.go` にメッセージ型を追加:
   ```go
   // statusMessage はステータスバーに表示するメッセージ
   type statusMessage struct {
       text     string
       isError  bool
   }

   // clearStatusMsg はステータスメッセージをクリアするメッセージ
   type clearStatusMsg struct{}

   // statusMessageClearCmd は指定時間後にclearStatusMsgを送信する
   func statusMessageClearCmd(duration time.Duration) tea.Cmd {
       return tea.Tick(duration, func(t time.Time) tea.Msg {
           return clearStatusMsg{}
       })
   }
   ```

2. `Model` 構造体にフィールドを追加:
   ```go
   type Model struct {
       // ... 既存フィールド ...
       statusMessage     string // ステータスバーに表示するメッセージ
       statusMessageTime time.Time // メッセージ表示開始時刻
       isStatusError     bool   // エラーメッセージかどうか
   }
   ```

3. `renderStatusBar()` を修正してエラーメッセージを表示:
   ```go
   func (m Model) renderStatusBar() string {
       // ステータスメッセージがある場合はそれを優先表示
       if m.statusMessage != "" {
           // エラーメッセージの場合は赤色など
           // ...
       }
       // 通常のステータスバー表示
       // ...
   }
   ```

4. `Update()` でメッセージ処理を追加:
   ```go
   case clearStatusMsg:
       m.statusMessage = ""
       m.isStatusError = false
       return m, nil
   ```

**Dependencies**: Phase 1

**Testing**:
- ステータスメッセージが表示されることを確認
- 5秒後にメッセージがクリアされることを確認
- 新しいアクションでメッセージがクリアされることを確認

**Estimated Effort**: Medium

---

### Phase 3: エラーメッセージのフォーマット

**Goal**: エラータイプに応じた適切なメッセージを生成する

**Files to Create/Modify**:
- `internal/ui/errors.go` (新規作成) - エラーメッセージフォーマット関数

**Implementation Steps**:

1. `errors.go` ファイルを作成:
   ```go
   package ui

   import (
       "errors"
       "fmt"
       "os"
       "syscall"
   )

   // formatDirectoryError はディレクトリアクセスエラーをユーザー向けメッセージにフォーマット
   func formatDirectoryError(err error, path string) string {
       if err == nil {
           return ""
       }

       // syscallエラーを検出
       var pathErr *os.PathError
       if errors.As(err, &pathErr) {
           var errno syscall.Errno
           if errors.As(pathErr.Err, &errno) {
               switch errno {
               case syscall.EACCES:
                   return fmt.Sprintf("Permission denied: %s", path)
               case syscall.ENOENT:
                   return fmt.Sprintf("No such directory: %s", path)
               case syscall.EIO:
                   return fmt.Sprintf("I/O error: %s", path)
               }
           }
       }

       // os.IsNotExist でも判定
       if os.IsNotExist(err) {
           return fmt.Sprintf("No such directory: %s", path)
       }

       if os.IsPermission(err) {
           return fmt.Sprintf("Permission denied: %s", path)
       }

       // その他のエラー
       return fmt.Sprintf("Cannot access: %s", path)
   }
   ```

**Dependencies**: なし

**Testing**:
- 各エラータイプに対して正しいメッセージが生成されることを確認
- EACCES → "Permission denied: {path}"
- ENOENT → "No such directory: {path}"
- EIO → "I/O error: {path}"
- その他 → "Cannot access: {path}"

**Estimated Effort**: Small

---

### Phase 4: EnterDirectoryの非同期化とエラーハンドリング

**Goal**: ディレクトリ移動をアトミックに行い、エラー時にパスを復元する

**Files to Create/Modify**:
- `internal/ui/pane.go` - `EnterDirectory()` を非同期化
- `internal/ui/model.go` - `directoryLoadCompleteMsg` のハンドリングを修正

**Implementation Steps**:

1. `pane.go` の `EnterDirectory()` を修正:
   ```go
   // EnterDirectoryAsync はディレクトリへの移動を開始し、Cmdを返す
   func (p *Pane) EnterDirectoryAsync() tea.Cmd {
       entry := p.SelectedEntry()
       if entry == nil {
           return nil
       }

       // シンボリックリンクの処理（既存ロジック維持）
       if entry.IsSymlink {
           if entry.LinkBroken {
               return nil
           }
           isDir, err := fs.IsDirectory(entry.LinkTarget)
           if err != nil || !isDir {
               return nil
           }
       }

       // 通常のディレクトリ処理
       if !entry.IsDir && !entry.IsSymlink {
           return nil
       }

       var newPath string
       if entry.IsParentDir() {
           newPath = filepath.Dir(p.path)
       } else if entry.IsSymlink {
           newPath = filepath.Join(p.path, entry.Name)
       } else {
           newPath = filepath.Join(p.path, entry.Name)
       }

       // 現在のパスを記録（復元用）
       p.recordPreviousPath()
       p.pendingPath = newPath
       p.path = newPath

       // ローディング状態を開始
       p.StartLoadingDirectory()

       return LoadDirectoryAsync(newPath)
   }
   ```

2. `messages.go` の `directoryLoadCompleteMsg` を拡張:
   ```go
   type directoryLoadCompleteMsg struct {
       panePath      string
       entries       []fs.FileEntry
       err           error
       attemptedPath string // エラー時にメッセージに表示するパス
   }
   ```

3. `pane.go` の `LoadDirectoryAsync()` を修正:
   ```go
   func LoadDirectoryAsync(panePath string) tea.Cmd {
       return func() tea.Msg {
           entries, err := fs.ReadDirectory(panePath)
           if err != nil {
               return directoryLoadCompleteMsg{
                   panePath:      panePath,
                   entries:       nil,
                   err:           err,
                   attemptedPath: panePath,
               }
           }

           fs.SortEntries(entries)
           return directoryLoadCompleteMsg{
               panePath:      panePath,
               entries:       entries,
               err:           nil,
               attemptedPath: panePath,
           }
       }
   }
   ```

4. `model.go` の `directoryLoadCompleteMsg` ハンドリングを修正:
   ```go
   case directoryLoadCompleteMsg:
       var targetPane *Pane
       // どのペインの読み込みかを判定
       if msg.panePath == m.leftPane.Path() || msg.panePath == m.leftPane.pendingPath {
           targetPane = m.leftPane
       } else if msg.panePath == m.rightPane.Path() || msg.panePath == m.rightPane.pendingPath {
           targetPane = m.rightPane
       }

       if targetPane != nil {
           targetPane.loading = false
           targetPane.loadingProgress = ""

           if msg.err != nil {
               // エラー時: パスを復元してステータスバーにメッセージ表示
               targetPane.restorePreviousPath()
               m.statusMessage = formatDirectoryError(msg.err, msg.attemptedPath)
               m.isStatusError = true
               m.statusMessageTime = time.Now()
               return m, statusMessageClearCmd(5 * time.Second)
           }

           // 成功時: エントリを更新
           entries := msg.entries
           if !targetPane.showHidden {
               entries = filterHiddenFiles(entries)
           }
           targetPane.entries = entries
           targetPane.cursor = 0
           targetPane.scrollOffset = 0
           targetPane.pendingPath = ""
       }
       return m, nil
   ```

5. `model.go` の `KeyEnter` ハンドリングを修正:
   ```go
   case KeyEnter:
       cmd := m.getActivePane().EnterDirectoryAsync()
       // ディスク容量更新は成功時のみ行うよう、directoryLoadCompleteMsg内で実施
       return m, cmd
   ```

**Dependencies**: Phase 1, Phase 2, Phase 3

**Testing**:
- 権限のないディレクトリへの移動でパスが変わらないことを確認
- エラーメッセージがステータスバーに表示されることを確認
- 正常なディレクトリへの移動が引き続き機能することを確認
- 連続して同じエラーディレクトリに入ろうとしてもパスが延長されないことを確認

**Estimated Effort**: Medium

---

### Phase 5: その他のナビゲーション関数の修正

**Goal**: `MoveToParent()`, `ChangeDirectory()`, `NavigateToHome()`, `NavigateToPrevious()` も同様にアトミックに

**Files to Create/Modify**:
- `internal/ui/pane.go` - 各ナビゲーション関数を修正

**Implementation Steps**:

1. 各ナビゲーション関数を非同期化:
   ```go
   // MoveToParentAsync は親ディレクトリへの移動を開始
   func (p *Pane) MoveToParentAsync() tea.Cmd {
       if p.path == "/" {
           return nil
       }
       newPath := filepath.Dir(p.path)
       p.recordPreviousPath()
       p.pendingPath = newPath
       p.path = newPath
       p.StartLoadingDirectory()
       return LoadDirectoryAsync(newPath)
   }

   // ChangeDirectoryAsync は指定パスへの移動を開始
   func (p *Pane) ChangeDirectoryAsync(path string) tea.Cmd {
       p.recordPreviousPath()
       p.pendingPath = path
       p.path = path
       p.StartLoadingDirectory()
       return LoadDirectoryAsync(path)
   }

   // NavigateToHomeAsync はホームディレクトリへの移動を開始
   func (p *Pane) NavigateToHomeAsync() tea.Cmd {
       home, err := fs.HomeDirectory()
       if err != nil {
           return nil
       }
       if p.path == home {
           return nil
       }
       p.recordPreviousPath()
       p.pendingPath = home
       p.path = home
       p.StartLoadingDirectory()
       return LoadDirectoryAsync(home)
   }

   // NavigateToPreviousAsync は直前のディレクトリへの移動を開始
   func (p *Pane) NavigateToPreviousAsync() tea.Cmd {
       if p.previousPath == "" {
           return nil
       }
       current := p.path
       p.pendingPath = p.previousPath
       p.path = p.previousPath
       p.previousPath = current
       p.StartLoadingDirectory()
       return LoadDirectoryAsync(p.path)
   }
   ```

2. `model.go` の対応するキーハンドリングを修正:
   ```go
   case KeyMoveLeft, KeyArrowLeft:
       if m.activePane == LeftPane {
           return m, m.leftPane.MoveToParentAsync()
       } else {
           m.switchToPane(LeftPane)
       }

   case KeyMoveRight, KeyArrowRight:
       if m.activePane == RightPane {
           return m, m.rightPane.MoveToParentAsync()
       } else {
           m.switchToPane(RightPane)
       }

   case KeyHome:
       return m, m.getActivePane().NavigateToHomeAsync()

   case KeyPrevDir:
       return m, m.getActivePane().NavigateToPreviousAsync()
   ```

**Note**: 同期版の関数（`EnterDirectory()`, `MoveToParent()`, `ChangeDirectory()` 等）は、
`NewPane()` 初期化時や `ToggleHidden()` など内部で使用されているため、削除せず残す。

**Dependencies**: Phase 4

**Testing**:
- h/← キーでの親ディレクトリ移動がエラー時にパスを復元することを確認
- ~ キーでのホームディレクトリ移動が正常に動作することを確認
- - キーでの直前ディレクトリ移動が正常に動作することを確認

**Estimated Effort**: Medium

---

### Phase 6: ユーザーアクションによるステータスメッセージクリア

**Goal**: ユーザーが次のアクションを行ったときにステータスメッセージをクリアする

**Files to Create/Modify**:
- `internal/ui/model.go` - キー入力時にメッセージをクリア

**Implementation Steps**:

1. `Update()` のキー入力処理の冒頭でメッセージをクリア:
   ```go
   case tea.KeyMsg:
       // ダイアログが開いている場合はダイアログに処理を委譲
       if m.dialog != nil {
           var cmd tea.Cmd
           m.dialog, cmd = m.dialog.Update(msg)
           return m, cmd
       }

       // ステータスメッセージがあればクリア
       if m.statusMessage != "" {
           m.statusMessage = ""
           m.isStatusError = false
       }

       // 既存のキー処理...
   ```

**Dependencies**: Phase 2

**Testing**:
- キー入力でステータスメッセージがクリアされることを確認

**Estimated Effort**: Small

---

### Phase 7: ユニットテストの追加

**Goal**: エラーハンドリングの動作を検証するテストを追加

**Files to Create/Modify**:
- `internal/ui/pane_test.go` - パス復元のテスト追加
- `internal/ui/errors_test.go` (新規作成) - エラーメッセージフォーマットのテスト

**Implementation Steps**:

1. `pane_test.go` に権限エラー時のテストを追加:
   ```go
   func TestEnterDirectory_PermissionDenied(t *testing.T) {
       // Setup: 読み取り権限のないディレクトリを作成（可能な環境のみ）
       // テストが実行できない環境ではスキップ
   }

   func TestEnterDirectory_NoPathExtension(t *testing.T) {
       // 連続してエラーディレクトリに入ろうとしても
       // パスが延長されないことを確認
   }

   func TestRestorePreviousPath(t *testing.T) {
       // restorePreviousPath() が正しくパスを復元することを確認
   }
   ```

2. `errors_test.go` を作成:
   ```go
   func TestFormatDirectoryError(t *testing.T) {
       tests := []struct {
           name     string
           err      error
           path     string
           expected string
       }{
           // 各エラータイプのテストケース
       }
   }
   ```

**Dependencies**: Phase 1-6

**Testing**:
- すべてのテストがパスすることを確認

**Estimated Effort**: Medium

---

## File Structure

```
internal/
├── ui/
│   ├── pane.go           # Pane構造体の拡張、非同期ナビゲーション関数
│   ├── model.go          # Model構造体のステータスメッセージ対応、エラーハンドリング
│   ├── messages.go       # statusMessage, clearStatusMsg の追加
│   ├── errors.go         # (新規) formatDirectoryError関数
│   ├── pane_test.go      # 権限エラー時のテスト追加
│   └── errors_test.go    # (新規) エラーメッセージのテスト
└── fs/
    └── reader.go         # (変更なし)
```

## Testing Strategy

### Unit Tests

1. **エラーメッセージフォーマット**
   - EACCES → "Permission denied: {path}"
   - ENOENT → "No such directory: {path}"
   - EIO → "I/O error: {path}"
   - その他 → "Cannot access: {path}"

2. **パス復元**
   - `restorePreviousPath()` が正しくパスを復元
   - `previousPath` が更新される

3. **ナビゲーション関数**
   - エラー時にパスが変わらない
   - 成功時にパスが更新される

### Integration Tests

1. **ステータスバー表示**
   - エラーメッセージがステータスバーに表示される
   - 5秒後にメッセージが消える
   - 次のアクションでメッセージが消える

2. **既存機能との互換性**
   - シンボリックリンクのナビゲーション
   - 隠しファイルのトグル
   - ホームディレクトリ/直前ディレクトリへの移動

### Manual Testing Checklist

- [ ] 権限のないディレクトリに入ろうとしてもパスが変わらない
- [ ] エラーメッセージがステータスバーに表示される
- [ ] 5秒後にメッセージが消える
- [ ] 任意のキーを押すとメッセージが消える
- [ ] 連続して同じエラーディレクトリに入ろうとしてもパスが延長されない
- [ ] エラー後に正常なディレクトリに移動できる
- [ ] シンボリックリンクのナビゲーションが正常に動作する
- [ ] h/←、l/→ キーによる親ディレクトリ移動が正常に動作する
- [ ] ~ キーによるホームディレクトリ移動が正常に動作する
- [ ] - キーによる直前ディレクトリ移動が正常に動作する

## Dependencies

### External Libraries

- `github.com/charmbracelet/bubbletea` - TUIフレームワーク（既存）
- `github.com/charmbracelet/lipgloss` - スタイリング（既存）

### Internal Dependencies

- Phase 1 → Phase 4, Phase 5（パス復元機能が必要）
- Phase 2 → Phase 4, Phase 6（ステータスメッセージ機能が必要）
- Phase 3 → Phase 4（エラーメッセージフォーマットが必要）
- Phase 4 → Phase 5（`EnterDirectoryAsync` のパターンを他の関数に適用）

## Risk Assessment

### Technical Risks

- **非同期処理の競合**: 複数のナビゲーション操作が同時に発生する可能性
  - Mitigation: ローディング中はナビゲーション操作を無効化する（既存の `loading` フラグを活用）

- **メッセージのタイミング問題**: clearStatusMsg が予期しないタイミングで到着する可能性
  - Mitigation: `statusMessageTime` で最新のメッセージかどうかを確認

### Implementation Risks

- **既存テストへの影響**: ナビゲーション関数のシグネチャ変更による既存テストの失敗
  - Mitigation: 同期版の関数を残し、非同期版を別名で追加

- **同期版と非同期版の混在**: コードの複雑化
  - Mitigation: 同期版は内部使用（初期化、トグル等）に限定、ユーザーアクションは非同期版を使用

## Performance Considerations

- 追加のファイルシステムコールなし（既存の `ReadDirectory()` でエラーを検出）
- タイマー処理は Bubble Tea の組み込み `tea.Tick` を使用（効率的）
- パス復元は文字列のスワップのみ（O(1)）

## Security Considerations

- 内部エラー詳細（スタックトレースなど）をユーザーに表示しない
- パス情報は既に表示されているため、エラーメッセージに含めても問題なし
- シンボリックリンクのリンク先パスは `LinkTarget` で管理（既存）

## Open Questions

なし - すべての要件は仕様書で明確化されています。

## Future Enhancements

- エラーメッセージのローカライズ対応（日本語メッセージ）
- エラーログの記録（デバッグ用）
- カスタムエラーメッセージ表示時間の設定

## References

- [SPEC.md](./SPEC.md) - 仕様書
- [要件定義書.md](./要件定義書.md) - 要件定義書
- `internal/ui/pane.go:170-216` - 現在の EnterDirectory() 実装
- `internal/ui/model.go:201-229` - 現在の directoryLoadCompleteMsg ハンドリング
- `internal/fs/reader.go:10-78` - ReadDirectory() 実装
