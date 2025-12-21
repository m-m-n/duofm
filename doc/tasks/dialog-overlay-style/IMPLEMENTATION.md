# Implementation Plan: Dialog Overlay Style Improvement

## Overview

ダイアログ表示時の背景ペインの見た目を改善する。█文字での塗りつぶしから、グレー背景・暗いテキストでの再描画に変更し、元のファイルリストを視認可能にする。ダイアログの種類によって背景変更範囲を切り替える。

## Objectives

- 背景ペインのファイルリストを視認可能にする
- グレー背景・暗いテキストで「後ろに引っ込んでいる」効果を出す
- 全体表示ダイアログ（ヘルプ、エラー）: 左右両方の背景変更
- ペイン表示ダイアログ（確認、コンテキストメニュー）: 操作ペインのみ背景変更

## Prerequisites

- 既存のダイアログシステムが正常に動作していること
- lipglossライブラリが使用可能であること

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Model                                │
│                                                             │
│  Dialog.DisplayType() determines rendering:                 │
│                                                             │
│  DialogDisplayPane:           DialogDisplayScreen:          │
│  ┌─────────┬─────────┐       ┌─────────┬─────────┐         │
│  │ Dimmed  │ Normal  │       │ Dimmed  │ Dimmed  │         │
│  │  ┌───┐  │         │       │     ┌───────┐     │         │
│  │  │Dlg│  │         │       │     │  Dlg  │     │         │
│  │  └───┘  │         │       │     └───────┘     │         │
│  └─────────┴─────────┘       └─────────┴─────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Dialog Interface Extension

**Goal**: Dialogインターフェースに表示タイプを追加

**Files to Modify**:
- `internal/ui/dialog.go` - DialogDisplayType型とDisplayType()メソッドを追加

**Implementation Steps**:

1. `DialogDisplayType`型を定義
   ```go
   type DialogDisplayType int

   const (
       DialogDisplayPane   DialogDisplayType = iota // ペインローカル
       DialogDisplayScreen                          // 画面全体
   )
   ```

2. `Dialog`インターフェースに`DisplayType()`メソッドを追加
   ```go
   type Dialog interface {
       Update(msg tea.Msg) (Dialog, tea.Cmd)
       View() string
       IsActive() bool
       DisplayType() DialogDisplayType // 新規追加
   }
   ```

**Dependencies**: なし

**Testing**:
- 型定義のコンパイル確認

**Estimated Effort**: Small

---

### Phase 2: Implement DisplayType for Each Dialog

**Goal**: 各ダイアログにDisplayType()メソッドを実装

**Files to Modify**:
- `internal/ui/help_dialog.go` - DialogDisplayScreen を返す
- `internal/ui/error_dialog.go` - DialogDisplayScreen を返す
- `internal/ui/confirm_dialog.go` - DialogDisplayPane を返す
- `internal/ui/context_menu_dialog.go` - DialogDisplayPane を返す

**Implementation Steps**:

1. `help_dialog.go`に追加:
   ```go
   func (d *HelpDialog) DisplayType() DialogDisplayType {
       return DialogDisplayScreen
   }
   ```

2. `error_dialog.go`に追加:
   ```go
   func (d *ErrorDialog) DisplayType() DialogDisplayType {
       return DialogDisplayScreen
   }
   ```

3. `confirm_dialog.go`に追加:
   ```go
   func (d *ConfirmDialog) DisplayType() DialogDisplayType {
       return DialogDisplayPane
   }
   ```

4. `context_menu_dialog.go`に追加:
   ```go
   func (d *ContextMenuDialog) DisplayType() DialogDisplayType {
       return DialogDisplayPane
   }
   ```

**Dependencies**: Phase 1完了

**Testing**:
- 各ダイアログの`DisplayType()`が正しい値を返すことを確認
- 既存のダイアログテストが通ることを確認

**Estimated Effort**: Small

---

### Phase 3: Add Dimmed View Method to Pane

**Goal**: Paneにグレー背景・暗いテキスト用の描画メソッドを追加

**Files to Modify**:
- `internal/ui/pane.go` - ViewDimmedWithDiskSpace()メソッドを追加

**Implementation Steps**:

1. dimmedスタイル用の定数を定義:
   ```go
   var (
       dimmedBgColor = lipgloss.Color("236") // 濃いグレー背景
       dimmedFgColor = lipgloss.Color("243") // 暗いテキスト
   )
   ```

2. `ViewDimmedWithDiskSpace()`メソッドを追加:
   - 基本的に`ViewWithDiskSpace()`と同じ構造
   - すべてのスタイルで:
     - 背景色を`dimmedBgColor`に設定
     - 前景色を`dimmedFgColor`に設定
   - カーソルハイライトは無効化（背景に溶け込む）
   - ファイルタイプの色付けも無効化（すべて暗いグレー）

3. 実装の詳細:
   ```go
   func (p *Pane) ViewDimmedWithDiskSpace(diskSpace uint64) string {
       var b strings.Builder

       // パス表示（暗いスタイル）
       pathStyle := lipgloss.NewStyle().
           Width(p.width-2).
           Padding(0, 1).
           Bold(true).
           Background(dimmedBgColor).
           Foreground(dimmedFgColor)
       // ...

       // ヘッダー、区切り線、ファイルリストも同様に暗いスタイルを適用
   }
   ```

**Dependencies**: なし（Phase 1, 2と並行可能）

**Testing**:
- `ViewDimmedWithDiskSpace()`が正しい形式の文字列を返すことを確認
- 背景色・前景色が正しく適用されていることを目視確認

**Estimated Effort**: Medium

---

### Phase 4: Update Model View Rendering

**Goal**: model.goのダイアログ描画ロジックを更新

**Files to Modify**:
- `internal/ui/model.go` - View()メソッドのダイアログ描画部分を修正

**Implementation Steps**:

1. 現在のダイアログ描画ロジック（行391-428）を修正:

2. `DialogDisplayType`に応じた分岐を追加:

   ```go
   if m.dialog != nil && m.dialog.IsActive() {
       switch m.dialog.DisplayType() {
       case DialogDisplayScreen:
           // 全体表示: 両方のペインをdimmed
           return m.renderDialogScreen()
       case DialogDisplayPane:
           // ペイン表示: アクティブペインのみdimmed
           return m.renderDialogPane()
       }
   }
   ```

3. `renderDialogScreen()`メソッドを追加:
   - 左右両方のペインを`ViewDimmedWithDiskSpace()`で描画
   - ダイアログを画面全体の中央に配置
   - `lipgloss.Place()`で画面全体サイズにダイアログを配置

4. `renderDialogPane()`メソッドを追加:
   - アクティブペインのみ`ViewDimmedWithDiskSpace()`で描画
   - 反対側は通常の`ViewWithDiskSpace()`で描画
   - ダイアログをアクティブペインの中央に配置
   - `lipgloss.Place()`でペインサイズにダイアログを配置

**Dependencies**: Phase 1, 2, 3 完了

**Testing**:
- ヘルプダイアログ(`?`)で左右両方がグレーになることを確認
- 削除確認(`d`)でアクティブペインのみグレーになることを確認
- コンテキストメニュー(`@`)でアクティブペインのみグレーになることを確認
- ダイアログを閉じると通常表示に戻ることを確認

**Estimated Effort**: Medium

---

## File Structure

```
internal/ui/
├── dialog.go                  # DialogDisplayType追加
├── help_dialog.go             # DisplayType() -> DialogDisplayScreen
├── error_dialog.go            # DisplayType() -> DialogDisplayScreen
├── confirm_dialog.go          # DisplayType() -> DialogDisplayPane
├── context_menu_dialog.go     # DisplayType() -> DialogDisplayPane
├── pane.go                    # ViewDimmedWithDiskSpace()追加
└── model.go                   # renderDialogScreen(), renderDialogPane()追加
```

## Testing Strategy

### Unit Tests

1. **dialog_test.go**:
   - `DialogDisplayType`の値が正しいことを確認
   - 各ダイアログの`DisplayType()`が正しい値を返すことを確認

2. **pane_test.go**:
   - `ViewDimmedWithDiskSpace()`が空でない文字列を返すことを確認
   - 出力に背景色・前景色のANSIコードが含まれることを確認

3. **model_test.go**:
   - ダイアログ表示時の描画が正しく行われることを確認

### Manual Testing Checklist

- [ ] `?`キーでヘルプ表示 → 左右両方グレー、ダイアログ中央
- [ ] 左ペインで`d`キー → 左のみグレー、右は通常、ダイアログ左ペイン中央
- [ ] 右ペインで`d`キー → 右のみグレー、左は通常、ダイアログ右ペイン中央
- [ ] 左ペインで`@`キー → 左のみグレー、右は通常
- [ ] 右ペインで`@`キー → 右のみグレー、左は通常
- [ ] エラー発生時 → 左右両方グレー、ダイアログ中央
- [ ] 各ダイアログを閉じる → 通常表示に即座に戻る
- [ ] グレー背景でもファイル名が読めることを確認

## Dependencies

### External Libraries
- `github.com/charmbracelet/lipgloss` - 既存使用中（追加不要）

### Internal Dependencies
- Phase 1 → Phase 2（インターフェース定義が先）
- Phase 3 は独立して実装可能
- Phase 4 は Phase 1, 2, 3 すべて完了後

## Risk Assessment

### Technical Risks

- **lipglossの背景色がターミナルで正しく表示されない可能性**
  - Mitigation: 複数のターミナルエミュレータでテスト
  - 256色対応ターミナルを前提とする

- **既存のテストが壊れる可能性**
  - Mitigation: 各フェーズ後にテストを実行して確認

### Implementation Risks

- **Paneの描画コードが複雑で重複が増える可能性**
  - Mitigation: スタイル生成を共通関数に切り出す検討

## Performance Considerations

- `ViewDimmedWithDiskSpace()`は`ViewWithDiskSpace()`とほぼ同じ計算量
- イベント駆動なので継続的な負荷なし
- 追加のメモリ割り当ては最小限（スタイルオブジェクトのみ）

## Security Considerations

- セキュリティ上の影響なし（視覚的な変更のみ）

## Open Questions

なし - すべての要件が確認済み

## References

- 仕様書: `doc/tasks/dialog-overlay-style/SPEC.md`
- 要件定義書: `doc/tasks/dialog-overlay-style/要件定義書.md`
