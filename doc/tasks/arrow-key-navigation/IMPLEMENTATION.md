# Implementation Plan: Arrow Key Navigation Support

## Overview

カーソルキー（↑↓←→）によるナビゲーションをduofmに追加する。既存のhjklキーバインドと同一の機能を提供する。

## Objectives

- hjklキーの代替としてカーソルキーナビゲーションを有効化
- 既存のキーバインドとの完全な互換性を維持
- ヘルプドキュメントを更新

## Prerequisites

- 現在のコードベースでhjklナビゲーションが動作していること
- Bubble Teaフレームワークの基本的な理解

## Architecture Overview

アーキテクチャの変更は不要。既存のキー処理にカーソルキーのcase文を追加するのみ。

## Implementation Phases

### Phase 1: キー定数の追加

**Goal**: カーソルキー用の定数を定義

**Files to Modify**:
- `internal/ui/keys.go` - カーソルキー定数を追加

**Implementation Steps**:

1. `keys.go`に以下の定数を追加:
   ```go
   KeyArrowDown  = "down"
   KeyArrowUp    = "up"
   KeyArrowLeft  = "left"
   KeyArrowRight = "right"
   ```

**Dependencies**: なし

**Testing**: コンパイル確認

**Estimated Effort**: Small

---

### Phase 2: メインビューのキー処理更新

**Goal**: メインビューでカーソルキーを処理

**Files to Modify**:
- `internal/ui/model.go` - Update()関数のswitch文を更新

**Implementation Steps**:

1. `KeyMoveDown`のcase文を`KeyMoveDown, KeyArrowDown:`に変更
2. `KeyMoveUp`のcase文を`KeyMoveUp, KeyArrowUp:`に変更
3. `KeyMoveLeft`のcase文を`KeyMoveLeft, KeyArrowLeft:`に変更
4. `KeyMoveRight`のcase文を`KeyMoveRight, KeyArrowRight:`に変更

**Code Changes**:
```go
case KeyMoveDown, KeyArrowDown:
    m.getActivePane().MoveCursorDown()

case KeyMoveUp, KeyArrowUp:
    m.getActivePane().MoveCursorUp()

case KeyMoveLeft, KeyArrowLeft:
    if m.activePane == LeftPane {
        m.leftPane.MoveToParent()
    } else {
        m.switchToPane(LeftPane)
    }

case KeyMoveRight, KeyArrowRight:
    if m.activePane == RightPane {
        m.rightPane.MoveToParent()
    } else {
        m.switchToPane(RightPane)
    }
```

**Dependencies**: Phase 1完了

**Testing**:
- ↑↓でカーソル移動確認
- ←→でペイン切り替え/親ディレクトリ移動確認

**Estimated Effort**: Small

---

### Phase 3: コンテキストメニューの確認・更新

**Goal**: コンテキストメニューでカーソルキーが動作することを確認

**Files to Modify**:
- `internal/ui/context_menu_dialog.go` - 必要に応じて←→を追加

**Implementation Steps**:

1. 現在のコードを確認（↑↓は既に対応済み）
2. ←→キーがページネーションに対応しているか確認
3. 必要に応じて←→の処理を追加

**Current Code** (already handles up/down):
```go
case "j", "down":
    // Move cursor down
case "k", "up":
    // Move cursor up
```

4. ←→をh/lと同様にページ切り替えに対応:
```go
case "h", "left":
    // Previous page (if pagination exists)
case "l", "right":
    // Next page (if pagination exists)
```

**Dependencies**: Phase 1完了

**Testing**: コンテキストメニューで↑↓キー動作確認

**Estimated Effort**: Small

---

### Phase 4: ヘルプダイアログの更新

**Goal**: ヘルプにカーソルキーの説明を追加

**Files to Modify**:
- `internal/ui/help_dialog.go` - ヘルプテキストを更新

**Implementation Steps**:

1. Navigationセクションのテキストを更新:
   - `"  j/k      : move cursor down/up"` → `"  j/k/↑/↓  : move cursor down/up"`
   - `"  h/l      : move to left/right pane or parent directory"` → `"  h/l/←/→  : move to left/right pane or parent directory"`

**Code Changes**:
```go
content := []string{
    "Navigation",
    "  j/k/↑/↓  : move cursor down/up",
    "  h/l/←/→  : move to left/right pane or parent directory",
    "  Enter    : enter directory",
    // ...
}
```

**Dependencies**: なし（並行して実施可能）

**Testing**: ?キーでヘルプを表示し、更新された内容を確認

**Estimated Effort**: Small

---

### Phase 5: テストの追加

**Goal**: カーソルキーナビゲーションのテストを追加

**Files to Modify**:
- `internal/ui/model_test.go` - カーソルキーのテストを追加

**Implementation Steps**:

1. 既存のキーナビゲーションテストを参考にカーソルキーテストを追加
2. テストケース:
   - ↓キーでカーソルが下に移動
   - ↑キーでカーソルが上に移動
   - ←キーで左ペインへ切り替え/親ディレクトリ
   - →キーで右ペインへ切り替え/親ディレクトリ

**Sample Test**:
```go
func TestArrowKeyNavigation(t *testing.T) {
    m := createTestModel()

    // Test down arrow
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
    if m.getActivePane().cursor != 1 {
        t.Error("Down arrow should move cursor down")
    }

    // Test up arrow
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
    if m.getActivePane().cursor != 0 {
        t.Error("Up arrow should move cursor up")
    }
}
```

**Dependencies**: Phase 2完了

**Testing**: `go test ./internal/ui/...`

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── keys.go                  # カーソルキー定数追加
├── model.go                 # Update()でカーソルキー処理
├── model_test.go            # カーソルキーテスト追加
├── context_menu_dialog.go   # ←→キー対応確認
└── help_dialog.go           # ヘルプテキスト更新
```

## Testing Strategy

### Unit Tests
- `model_test.go`: カーソルキーによるカーソル移動テスト
- `model_test.go`: カーソルキーによるペイン切り替えテスト
- `context_menu_dialog_test.go`: コンテキストメニューでのカーソルキーテスト

### Manual Testing Checklist
- [ ] ↑キーでカーソルが上に移動する
- [ ] ↓キーでカーソルが下に移動する
- [ ] 左ペインで←キーを押すと親ディレクトリに移動する
- [ ] 左ペインで→キーを押すと右ペインに切り替わる
- [ ] 右ペインで→キーを押すと親ディレクトリに移動する
- [ ] 右ペインで←キーを押すと左ペインに切り替わる
- [ ] コンテキストメニューで↑↓キーが動作する
- [ ] ヘルプ画面にカーソルキーの説明が表示される
- [ ] 既存のhjklキーが引き続き動作する

## Dependencies

### External Libraries
- なし（Bubble Teaの既存機能を使用）

### Internal Dependencies
- Phase 2はPhase 1に依存
- Phase 5はPhase 2に依存
- Phase 3, Phase 4は並行して実施可能

## Risk Assessment

### Technical Risks
- **リスク: なし** - 既存の動作するコードにcase文を追加するのみ

### Implementation Risks
- **リスク: 低** - 変更範囲が限定的

## Performance Considerations

- パフォーマンスへの影響なし
- キー処理は同期的で即座に完了

## Security Considerations

- セキュリティへの影響なし
- UI機能の追加のみ

## Open Questions

なし - すべての要件が明確化済み

## Verification Checklist

実装完了後、以下を確認:
- [ ] すべてのテストがパス (`go test ./...`)
- [ ] ビルドが成功 (`make build`)
- [ ] 手動テストチェックリストをすべて完了

## References

- [仕様書](./SPEC.md)
- [要件定義書](./要件定義書.md)
