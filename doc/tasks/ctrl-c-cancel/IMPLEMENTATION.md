# Implementation Plan: Ctrl+C Cancel Operation Support

## Overview

Ctrl+Cを全モーダル状態（ミニバッファ、ダイアログ、コンテキストメニュー）でEscと同様にキャンセルキーとして追加し、通常のファイルリスト表示状態では2回押しで終了する機能を実装します。

## Objectives

- すべてのモーダル状態でCtrl+Cによるキャンセル操作を可能にする
- 通常モードでの安全な終了メカニズム（2回押しで終了）を実装する
- CLI/TUIアプリケーションの一般的な動作との一貫性を維持する

## Prerequisites

- Go 1.21以上
- Bubble Teaフレームワークの理解
- 既存のダイアログ・ミニバッファ実装の理解

## Architecture Overview

既存のキー処理アーキテクチャに従い、以下の変更を行います：

1. **ダイアログ層**: 各ダイアログのUpdate関数のcase文に`"ctrl+c"`を追加
2. **ミニバッファ層**: `model.go`の検索状態ハンドリングに`tea.KeyCtrlC`を追加
3. **通常モード層**: `model.go`に状態管理フィールドとダブルプレス終了ロジックを追加
4. **メッセージ層**: `messages.go`にタイムアウトメッセージとコマンドを追加

```
┌─────────────────────────────────────────────────────────────┐
│                    Key Event Received                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Dialog Active?  │
                    └─────────────────┘
                      │           │
                    Yes          No
                      │           │
                      ▼           ▼
            ┌─────────────┐  ┌─────────────────┐
            │Handle Dialog│  │ Search Active?  │
            │(incl Ctrl+C)│  └─────────────────┘
            └─────────────┘    │           │
                             Yes          No
                               │           │
                               ▼           ▼
                    ┌─────────────┐  ┌─────────────────┐
                    │Handle Search│  │ Normal Mode     │
                    │(incl Ctrl+C)│  │ (Ctrl+C double) │
                    └─────────────┘  └─────────────────┘
```

## Implementation Phases

### Phase 1: メッセージ定義の追加

**Goal**: Ctrl+Cタイムアウト用のメッセージとコマンドを定義

**Files to Create/Modify**:
- `internal/ui/messages.go` - タイムアウトメッセージ型とコマンドの追加

**Implementation Steps**:

1. `ctrlCTimeoutMsg`型を定義
   ```go
   // ctrlCTimeoutMsg はCtrl+C終了確認のタイムアウトを通知
   type ctrlCTimeoutMsg struct{}
   ```

2. `ctrlCTimeoutCmd`関数を追加
   ```go
   // ctrlCTimeoutCmd は指定時間後にctrlCTimeoutMsgを送信するコマンド
   func ctrlCTimeoutCmd(duration time.Duration) tea.Cmd {
       return tea.Tick(duration, func(t time.Time) tea.Msg {
           return ctrlCTimeoutMsg{}
       })
   }
   ```

**Dependencies**:
- なし（既存パターンに従う）

**Testing**:
- コマンドが非nilを返すことを確認

**Estimated Effort**: Small

---

### Phase 2: Modelへの状態フィールド追加

**Goal**: Ctrl+Cの1回押し状態を追跡するフィールドを追加

**Files to Create/Modify**:
- `internal/ui/model.go` - Model構造体への`ctrlCPending`フィールド追加

**Implementation Steps**:

1. Model構造体に`ctrlCPending`フィールドを追加
   ```go
   type Model struct {
       // ... existing fields ...
       ctrlCPending bool // Ctrl+Cが1回押された状態かどうか
   }
   ```

2. NewModel関数で初期化（デフォルトfalseなので明示的な初期化は不要）

**Dependencies**:
- Phase 1完了

**Testing**:
- NewModelで`ctrlCPending`がfalseであることを確認

**Estimated Effort**: Small

---

### Phase 3: 通常モードでのCtrl+Cダブルプレス終了

**Goal**: 通常モードでCtrl+Cを2回押すと終了、1回目はメッセージ表示

**Files to Create/Modify**:
- `internal/ui/model.go` - Update関数の修正

**Implementation Steps**:

1. 現在の`"ctrl+c", KeyQuit`ケースを分離し、通常のq終了とCtrl+C終了を別々に処理

2. Ctrl+Cのダブルプレスロジックを実装：
   ```go
   case "ctrl+c":
       if m.ctrlCPending {
           // 2回目のCtrl+C - 終了
           return m, tea.Quit
       }
       // 1回目のCtrl+C - メッセージ表示とタイマー開始
       m.ctrlCPending = true
       m.statusMessage = "Press Ctrl+C again to quit"
       m.isStatusError = false
       return m, ctrlCTimeoutCmd(2 * time.Second)
   ```

3. タイムアウトメッセージのハンドリングを追加：
   ```go
   case ctrlCTimeoutMsg:
       if m.ctrlCPending {
           m.ctrlCPending = false
           m.statusMessage = ""
       }
       return m, nil
   ```

4. 他のキー入力時に`ctrlCPending`をリセット：
   - キー処理の最後（他の処理が完了した後）に追加
   ```go
   // Reset ctrlCPending on any other key
   if m.ctrlCPending {
       m.ctrlCPending = false
       m.statusMessage = ""
   }
   ```

**Dependencies**:
- Phase 1, 2完了

**Testing**:
- 単一Ctrl+Cでメッセージ表示とctrlCPending=trueを確認
- ダブルCtrl+Cで`tea.Quit`を返すことを確認
- タイムアウト後にctrlCPending=falseとメッセージクリアを確認
- 他キー入力後にctrlCPending=falseを確認

**Estimated Effort**: Medium

---

### Phase 4: ミニバッファ（検索）でのCtrl+Cキャンセル

**Goal**: 検索中にCtrl+Cでキャンセルできるようにする

**Files to Create/Modify**:
- `internal/ui/model.go` - 検索状態ハンドリングの修正

**Implementation Steps**:

1. 検索状態のEscハンドリングにCtrl+Cを追加：
   ```go
   // ミニバッファがアクティブな場合（検索中）
   if m.searchState.IsActive {
       switch msg.Type {
       case tea.KeyEnter:
           // 検索を確定
           m.confirmSearch()
           return m, nil

       case tea.KeyEsc, tea.KeyCtrlC:  // Ctrl+Cを追加
           // 検索をキャンセル
           m.cancelSearch()
           return m, nil
       // ...
       }
   }
   ```

**Dependencies**:
- Phase 2完了

**Testing**:
- 検索開始後、Ctrl+Cでミニバッファが閉じることを確認
- 検索がキャンセルされ、前のフィルタ状態が復元されることを確認

**Estimated Effort**: Small

---

### Phase 5: 確認ダイアログでのCtrl+Cキャンセル

**Goal**: 確認ダイアログでCtrl+Cでキャンセルできるようにする

**Files to Create/Modify**:
- `internal/ui/confirm_dialog.go` - Update関数の修正

**Implementation Steps**:

1. キャンセルケースに`"ctrl+c"`を追加：
   ```go
   case "n", "esc", "ctrl+c":  // ctrl+cを追加
       d.active = false
       return d, func() tea.Msg {
           return dialogResultMsg{
               result: DialogResult{Cancelled: true},
           }
       }
   ```

**Dependencies**:
- なし

**Testing**:
- 確認ダイアログ表示中にCtrl+Cでキャンセルされることを確認

**Estimated Effort**: Small

---

### Phase 6: エラーダイアログでのCtrl+Cクローズ

**Goal**: エラーダイアログでCtrl+Cで閉じられるようにする

**Files to Create/Modify**:
- `internal/ui/error_dialog.go` - Update関数の修正

**Implementation Steps**:

1. クローズケースに`"ctrl+c"`を追加：
   ```go
   case "esc", "enter", "ctrl+c":  // ctrl+cを追加
       d.active = false
       return d, func() tea.Msg {
           return dialogResultMsg{
               result: DialogResult{Cancelled: true},
           }
       }
   ```

**Dependencies**:
- なし

**Testing**:
- エラーダイアログ表示中にCtrl+Cで閉じることを確認

**Estimated Effort**: Small

---

### Phase 7: ヘルプダイアログでのCtrl+Cクローズ

**Goal**: ヘルプダイアログでCtrl+Cで閉じられるようにする

**Files to Create/Modify**:
- `internal/ui/help_dialog.go` - Update関数の修正

**Implementation Steps**:

1. クローズケースに`"ctrl+c"`を追加：
   ```go
   case "esc", "?", "ctrl+c":  // ctrl+cを追加
       d.active = false
       return d, func() tea.Msg {
           return dialogResultMsg{
               result: DialogResult{Cancelled: true},
           }
       }
   ```

**Dependencies**:
- なし

**Testing**:
- ヘルプダイアログ表示中にCtrl+Cで閉じることを確認

**Estimated Effort**: Small

---

### Phase 8: コンテキストメニューでのCtrl+Cキャンセル

**Goal**: コンテキストメニューでCtrl+Cでキャンセルできるようにする

**Files to Create/Modify**:
- `internal/ui/context_menu_dialog.go` - Update関数の修正

**Implementation Steps**:

1. キャンセルケースに`"ctrl+c"`を追加：
   ```go
   case "esc", "ctrl+c":  // ctrl+cを追加
       // Cancel and close
       d.active = false
       return d, func() tea.Msg {
           return contextMenuResultMsg{cancelled: true}
       }
   ```

**Dependencies**:
- なし

**Testing**:
- コンテキストメニュー表示中にCtrl+Cでキャンセルされることを確認

**Estimated Effort**: Small

---

### Phase 9: テストの追加

**Goal**: すべての新機能に対するユニットテストを追加

**Files to Create/Modify**:
- `internal/ui/model_test.go` - Ctrl+C関連のテスト追加
- `internal/ui/dialog_test.go` - ダイアログのCtrl+Cテスト追加
- `internal/ui/context_menu_dialog_test.go` - コンテキストメニューのCtrl+Cテスト追加

**Implementation Steps**:

1. `model_test.go`に以下のテストを追加：
   - `TestSingleCtrlCShowsMessage`: 1回目のCtrl+Cでメッセージ表示
   - `TestDoubleCtrlCQuits`: 2回目のCtrl+Cで終了
   - `TestCtrlCTimeoutResetsState`: タイムアウトで状態リセット
   - `TestOtherKeyAfterCtrlCResetsState`: 他キー入力で状態リセット
   - `TestSearchCtrlCCancelsSearch`: 検索中のCtrl+Cでキャンセル

2. `dialog_test.go`に以下のテストを追加：
   - `TestConfirmDialogCtrlCCancels`: 確認ダイアログのCtrl+Cキャンセル
   - `TestErrorDialogCtrlCCloses`: エラーダイアログのCtrl+Cクローズ
   - `TestHelpDialogCtrlCCloses`: ヘルプダイアログのCtrl+Cクローズ

3. `context_menu_dialog_test.go`に以下のテストを追加：
   - `TestContextMenuCtrlCCancels`: コンテキストメニューのCtrl+Cキャンセル

**Dependencies**:
- Phase 1-8完了

**Testing**:
- `go test ./internal/ui/...`で全テストがパスすることを確認

**Estimated Effort**: Medium

---

## File Structure

```
internal/ui/
├── model.go               # ctrlCPendingフィールド追加、Ctrl+Cハンドリング
├── messages.go            # ctrlCTimeoutMsg型とコマンド追加
├── confirm_dialog.go      # "ctrl+c"ケース追加
├── error_dialog.go        # "ctrl+c"ケース追加
├── help_dialog.go         # "ctrl+c"ケース追加
├── context_menu_dialog.go # "ctrl+c"ケース追加
├── model_test.go          # Ctrl+C関連テスト追加
├── dialog_test.go         # ダイアログCtrl+Cテスト追加
└── context_menu_dialog_test.go # コンテキストメニューCtrl+Cテスト追加
```

## Testing Strategy

### Unit Tests

#### model_test.go
```go
// TestSingleCtrlCShowsMessage tests that first Ctrl+C shows message
func TestSingleCtrlCShowsMessage(t *testing.T) {
    // Setup: Normal mode
    // Action: Send Ctrl+C key
    // Assert: Status message shown, ctrlCPending=true
}

// TestDoubleCtrlCQuits tests that double Ctrl+C quits application
func TestDoubleCtrlCQuits(t *testing.T) {
    // Setup: Normal mode
    // Action: Send Ctrl+C twice
    // Assert: Returns tea.Quit command
}

// TestCtrlCTimeoutResetsState tests that timeout resets state
func TestCtrlCTimeoutResetsState(t *testing.T) {
    // Setup: Send first Ctrl+C
    // Action: Send ctrlCTimeoutMsg
    // Assert: ctrlCPending=false, status message cleared
}

// TestOtherKeyAfterCtrlCResetsState tests that other key resets state
func TestOtherKeyAfterCtrlCResetsState(t *testing.T) {
    // Setup: Send first Ctrl+C
    // Action: Send 'j' key
    // Assert: ctrlCPending=false, status message cleared
}

// TestSearchCtrlCCancelsSearch tests Ctrl+C cancels search
func TestSearchCtrlCCancelsSearch(t *testing.T) {
    // Setup: Start search mode
    // Action: Send Ctrl+C key
    // Assert: Search cancelled, minibuffer hidden
}
```

#### dialog_test.go
```go
// TestConfirmDialogCtrlCCancels tests Ctrl+C cancels confirm dialog
func TestConfirmDialogCtrlCCancels(t *testing.T) {
    // Setup: Show confirm dialog
    // Action: Send Ctrl+C key
    // Assert: Dialog closed with Cancelled=true
}

// TestErrorDialogCtrlCCloses tests Ctrl+C closes error dialog
func TestErrorDialogCtrlCCloses(t *testing.T) {
    // Setup: Show error dialog
    // Action: Send Ctrl+C key
    // Assert: Dialog closed
}

// TestHelpDialogCtrlCCloses tests Ctrl+C closes help dialog
func TestHelpDialogCtrlCCloses(t *testing.T) {
    // Setup: Show help dialog
    // Action: Send Ctrl+C key
    // Assert: Dialog closed
}
```

### Integration Tests

- 実際のダイアログフローでのCtrl+C動作確認
- 検索中のCtrl+Cでの前フィルタ復元確認

### Manual Testing Checklist

- [ ] ミニバッファ表示中にCtrl+Cでキャンセル
- [ ] 確認ダイアログ表示中にCtrl+Cでキャンセル
- [ ] エラーダイアログ表示中にCtrl+Cでクローズ
- [ ] ヘルプダイアログ表示中にCtrl+Cでクローズ
- [ ] コンテキストメニュー表示中にCtrl+Cでキャンセル
- [ ] 通常画面でCtrl+C → メッセージ表示確認
- [ ] 通常画面でCtrl+C 2回 → アプリケーション終了確認
- [ ] 通常画面でCtrl+C → 2秒待機 → メッセージクリア確認
- [ ] 通常画面でCtrl+C → 他キー → メッセージクリア確認
- [ ] 既存のEscキー動作が変わらないことを確認
- [ ] 既存のqキー終了動作が変わらないことを確認

## Dependencies

### External Libraries
- `github.com/charmbracelet/bubbletea` - TUIフレームワーク（既存）

### Internal Dependencies
- Phase 1（メッセージ定義）→ Phase 2, 3が依存
- Phase 2（フィールド追加）→ Phase 3, 4が依存
- Phase 4-8は独立して実装可能
- Phase 9（テスト）はPhase 1-8完了後

## Risk Assessment

### Technical Risks
- **Ctrl+Cシグナルとの競合**: Bubble Teaが`tea.KeyCtrlC`を適切に処理するため問題なし
  - Mitigation: Bubble Teaのシグナル処理に任せる

### Implementation Risks
- **ステータスメッセージの競合**: 他の機能がステータスメッセージを使用している場合
  - Mitigation: Ctrl+C状態は専用フィールドで管理し、メッセージは上書き可能

## Performance Considerations

- タイマー使用によるオーバーヘッドは最小限（既存のdiskSpaceTickCmdと同様のパターン）
- boolean状態チェックは即座に完了
- 通常操作時のパフォーマンスへの影響なし

## Security Considerations

- セキュリティへの影響なし
- Ctrl+Cハンドリングは標準的なUIパターン

## Open Questions

なし - すべての要件はユーザーと確認済み

## Future Enhancements

- Ctrl+Dでのページダウン機能（将来の機能として検討可能）
- ESCキーのダブルプレス終了オプション（設定可能にする場合）

## References

- [SPEC.md](./SPEC.md) - 機能仕様書
- [要件定義書.md](./要件定義書.md) - 要件定義書
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
