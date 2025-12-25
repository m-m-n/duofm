# Implementation Plan: Sort Toggle

## Overview

ソートトグル機能を実装する。`s`キーでソート設定ダイアログを表示し、ユーザーがソート項目（Name/Size/Date）と順序（Asc/Desc）を選択できるようにする。

## Objectives

- ソートダイアログコンポーネントの実装
- ソートロジックの実装（ディレクトリ優先）
- ペインへのソート設定の統合
- ライブプレビュー機能の実装

## Prerequisites

- 既存のダイアログシステム（`dialog.go`, `context_menu_dialog.go`）
- Pane構造体（`pane.go`）
- fs.FileEntry型（Name, Size, ModTime, IsDirフィールド）

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          Model                                   │
│  ┌─────────────────┐    ┌─────────────────────────────────┐    │
│  │   SortDialog    │───▶│  Pane.sortConfig                │    │
│  │  (2行選択UI)     │    │  Pane.ApplySortAndPreserveCursor│    │
│  └─────────────────┘    └─────────────────────────────────┘    │
│           │                            │                         │
│           ▼                            ▼                         │
│  ┌─────────────────┐    ┌─────────────────────────────────┐    │
│  │ SortConfig      │    │  SortEntries()                  │    │
│  │ (Field + Order) │───▶│  (親dir → dirs → files)         │    │
│  └─────────────────┘    └─────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: ソート型とロジック

**Goal**: ソート設定の型定義とソート関数の実装

**Files to Create/Modify**:
- `internal/ui/sort.go` - 新規: SortField, SortOrder, SortConfig型、SortEntries関数
- `internal/ui/sort_test.go` - 新規: ソートロジックのユニットテスト

**Implementation Steps**:

1. **ソート型の定義** (`sort.go`)
   ```go
   type SortField int
   const (
       SortByName SortField = iota
       SortBySize
       SortByDate
   )

   type SortOrder int
   const (
       SortAsc SortOrder = iota
       SortDesc
   )

   type SortConfig struct {
       Field SortField
       Order SortOrder
   }
   ```

2. **SortEntries関数の実装**
   - 親ディレクトリ（`..`）を分離して先頭に配置
   - ディレクトリとファイルを分離
   - 各グループを指定されたフィールド/順序でソート
   - 結合して返す: `[..] + [dirs] + [files]`

3. **getLessFunc関数の実装**
   - SortConfigに基づいて比較関数を返す
   - 6パターン（Name/Size/Date × Asc/Desc）をサポート

**Testing**:
- 各ソートモード（6パターン）の動作確認
- 親ディレクトリが常に先頭にあることの確認
- ディレクトリがファイルより先にあることの確認
- 空のエントリリストの処理

**Estimated Effort**: Small

---

### Phase 2: ソートダイアログコンポーネント

**Goal**: 2行選択UIのダイアログ実装

**Files to Create/Modify**:
- `internal/ui/sort_dialog.go` - 新規: SortDialog構造体とメソッド
- `internal/ui/sort_dialog_test.go` - 新規: ダイアログのユニットテスト

**Implementation Steps**:

1. **SortDialog構造体の定義**
   ```go
   type SortDialog struct {
       config         SortConfig  // 現在の選択
       originalConfig SortConfig  // キャンセル時の復元用
       focusedRow     int         // 0: Sort by, 1: Order
       active         bool
       width          int
   }
   ```

2. **NewSortDialog関数**
   - 現在のSortConfigを受け取り、ダイアログを初期化
   - originalConfigに保存（キャンセル時の復元用）

3. **Update関数の実装**
   - `h/l/←/→`: 同じ行内で選択変更
   - `j/k/↓/↑`: 行間移動
   - `enter`: 確定（confirmed=true）
   - `esc/q`: キャンセル（cancelled=true、originalConfigに復元）

4. **View関数の実装**
   - 既存のダイアログスタイル（lipgloss）を使用
   - `[ ]`で選択中の項目を強調
   - フォーカス行のハイライト
   - 2行のヘルプテキスト

**Dialog Layout**:
```
┌─ Sort ─────────────────────────┐
│                                │
│  Sort by    [Name]  Size  Date │
│  Order       ↑Asc  [↓Desc]     │
│                                │
│  h/l:change  j/k:row           │
│  Enter:OK  Esc:cancel          │
└────────────────────────────────┘
```

**Testing**:
- キー入力による選択変更
- Enter/Escの動作
- hjklと矢印キーの両対応
- 境界条件（端での移動）

**Estimated Effort**: Medium

---

### Phase 3: Pane統合

**Goal**: PaneにソートConfigを追加し、ソート適用機能を実装

**Files to Create/Modify**:
- `internal/ui/pane.go` - 修正: sortConfigフィールド追加、メソッド追加
- `internal/ui/keys.go` - 修正: KeySort定数追加

**Implementation Steps**:

1. **Pane構造体にsortConfigを追加**
   ```go
   type Pane struct {
       // ... existing fields
       sortConfig SortConfig  // デフォルト: {SortByName, SortAsc}
   }
   ```

2. **NewPane関数の更新**
   - sortConfigをデフォルト値で初期化

3. **LoadDirectory関数の更新**
   - 既存の`fs.SortEntries(entries)`呼び出しを
   - `SortEntries(entries, p.sortConfig)`に変更

4. **ApplySortAndPreserveCursor関数の実装**
   - 現在のカーソル位置のファイル名を記憶
   - ソートを適用
   - 同じファイル名を検索してカーソルを復元
   - 見つからない場合は現在のインデックスを維持

5. **keys.goにKeySortを追加**
   ```go
   KeySort = "s"
   ```

**Testing**:
- ソート適用後のカーソル位置維持
- ディレクトリ移動後もソート設定が維持されること
- 隠しファイルトグル後もソート設定が維持されること

**Estimated Effort**: Small

---

### Phase 4: Model統合とライブプレビュー

**Goal**: Modelにダイアログ処理を追加、ライブプレビュー実装

**Files to Create/Modify**:
- `internal/ui/model.go` - 修正: sortDialogフィールド追加、キーハンドラ追加
- `internal/ui/messages.go` - 修正: sortDialogResultMsg追加（必要に応じて）

**Implementation Steps**:

1. **Model構造体にsortDialogを追加**
   ```go
   type Model struct {
       // ... existing fields
       sortDialog *SortDialog
   }
   ```

2. **'s'キーハンドラの追加**
   - ダイアログがない場合、新規SortDialogを作成
   - 現在のペインのsortConfigを渡す

3. **handleSortDialogKey関数の実装**
   - ダイアログのUpdateを呼び出す
   - confirmed/cancelledをチェック
   - ライブプレビュー: 選択変更ごとにペインのソートを更新
   - キャンセル時: originalConfigに復元してソート再適用

4. **View関数の更新**
   - sortDialogがactiveの場合、ダイアログをオーバーレイ表示
   - 既存のダイアログ表示パターンに従う

**Implementation Details**:
```go
func (m *Model) handleSortDialogKey(msg tea.KeyMsg) tea.Cmd {
    confirmed, cancelled := m.sortDialog.HandleKey(msg.String())

    if confirmed {
        m.sortDialog = nil
        return nil
    }

    if cancelled {
        // Restore original sort
        m.ActivePane().sortConfig = m.sortDialog.originalConfig
        m.ActivePane().ApplySortAndPreserveCursor()
        m.sortDialog = nil
        return nil
    }

    // Live preview
    m.ActivePane().sortConfig = m.sortDialog.config
    m.ActivePane().ApplySortAndPreserveCursor()
    return nil
}
```

**Testing**:
- 's'キーでダイアログが開くこと
- ライブプレビューが機能すること
- キャンセルで元のソートに戻ること
- 確定でダイアログが閉じること

**Estimated Effort**: Medium

---

### Phase 5: E2Eテスト

**Goal**: E2Eテストの作成

**Files to Create/Modify**:
- `test/e2e/sort_test.go` - 新規: ソート機能のE2Eテスト

**Test Scenarios**:
- [ ] `s`キーでソートダイアログが開く
- [ ] ダイアログに現在のソート設定が表示される
- [ ] hjklと矢印キーでナビゲーションできる
- [ ] Enterで確定してダイアログが閉じる
- [ ] Escでキャンセルして元のソートに戻る
- [ ] 左右ペインで独立したソート設定
- [ ] ディレクトリ移動後もソート設定が維持
- [ ] アプリ再起動でデフォルト（Name Asc）にリセット

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── sort.go              # SortField, SortOrder, SortConfig, SortEntries
├── sort_test.go         # ソートロジックのテスト
├── sort_dialog.go       # SortDialog構造体
├── sort_dialog_test.go  # ダイアログのテスト
├── pane.go              # sortConfigフィールド追加
├── keys.go              # KeySort追加
├── model.go             # sortDialog統合
└── ...

test/e2e/
└── sort_test.go         # E2Eテスト
```

## Testing Strategy

### Unit Tests

**sort.go**:
- SortEntries: 6パターン（Name/Size/Date × Asc/Desc）
- 親ディレクトリの位置
- ディレクトリ/ファイルの順序
- 空リストの処理

**sort_dialog.go**:
- キー入力による状態変更
- Enter/Escの動作
- 境界条件

### Integration Tests

- ペインとソートの連携
- ライブプレビューの動作
- カーソル位置の維持

### Manual Testing Checklist

- [ ] `s`キーでダイアログ表示
- [ ] 各ソートモードでファイル順序が正しい
- [ ] ディレクトリが常にファイルより先
- [ ] `..`が常に先頭
- [ ] ライブプレビューが動作
- [ ] Escで元に戻る
- [ ] 左右ペインで独立動作

## Dependencies

### External Libraries
- `github.com/charmbracelet/lipgloss` - UI スタイリング（既存）
- `github.com/charmbracelet/bubbletea` - TUIフレームワーク（既存）

### Internal Dependencies
- Phase 1 → Phase 3（PaneでSortEntriesを使用）
- Phase 2 → Phase 4（ModelでSortDialogを使用）

## Risk Assessment

### Technical Risks

- **ソート性能**: 大量ファイル（10,000+）でのパフォーマンス
  - Mitigation: `sort.SliceStable`使用、必要に応じてプロファイリング

- **カーソル位置の復元**: ソート後に同名ファイルが見つからない場合
  - Mitigation: 仕様通り現在のインデックスを維持

### Implementation Risks

- **既存機能への影響**: LoadDirectory変更による副作用
  - Mitigation: 既存テストが通ることを確認

## Performance Considerations

- `sort.SliceStable`を使用して安定ソートを保証
- ソートは同期実行（ブロッキング許容）
- 10,000ファイルでも200ms以内を目標

## Security Considerations

- ソート機能は読み取り専用操作
- ファイルシステムへの書き込みなし
- セキュリティリスクなし

## References

- [SPEC.md](./SPEC.md) - 技術仕様書
- [要件定義書.md](./要件定義書.md) - 要件定義
- [既存ダイアログ実装](../../internal/ui/context_menu_dialog.go) - UIパターン参考
