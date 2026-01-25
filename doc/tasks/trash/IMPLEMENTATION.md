# Implementation Plan: Trash Dialog

## Overview

ゴミ箱の表示を専用ダイアログ（TrashDialog）に変更する。画面中央に表示され、両ペインを暗転（DialogDisplayScreen）する方式を採用する。

## Objectives

- ゴミ箱操作をダイアログ内に分離し、キーバインド衝突を回避する
- 直感的なゴミ箱管理UI（一覧表示、復元、空にする）を提供する
- 既存のダイアログパターン（HelpDialog、BookmarkManagerDialog）に準拠する

## Prerequisites

### Development Environment
- Go 1.21以上
- make（ビルド自動化用）

### Dependencies
- 既存のダイアログ基盤（`internal/ui/dialog_base.go`、`internal/ui/dialog.go`）
- 既存のトラッシュ操作（`internal/fs/trash.go`、`internal/fs/trashinfo.go`）
- 既存のEmptyTrashDialog、RestoreConflictDialog

### Knowledge Requirements
- Bubble Teaのダイアログパターン
- DialogDisplayScreen型の使用方法（HelpDialogを参照）
- 既存のトラッシュインフラ

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Libraries**:
  - `github.com/charmbracelet/lipgloss` - スタイリング
  - 既存の`internal/fs`パッケージ - ファイル操作

### Design Approach
- TrashDialogをDialogDisplayScreen型で実装
- HelpDialogに類似した全画面オーバーレイ
- ダイアログ内で復元・空にするを完結

### Component Interaction
```
User Input (T key)
    |
    v
Model.handleOpenTrashDialog()
    |
    v
TrashDialog (DialogDisplayScreen)
    |-- j/k: カーソル移動
    |-- Space: マーク切り替え
    |-- R: 復元（restoreConflictResultMsg経由）
    |-- Shift+E: 空にする（emptyTrashResultMsg経由）
    v
ダイアログ結果の処理
    |
    v
ペインリフレッシュ
```

## Implementation Phases

### Phase 1: TrashDialog基本実装

**Goal**: TrashDialogの基本構造と表示機能、ナビゲーションを実現する

**Files to Create**:
- `internal/ui/trash_dialog.go` - TrashDialogの実装（ナビゲーション含む）
- `internal/ui/trash_dialog_test.go` - TrashDialogのテスト

**Files to Modify**:
- `internal/ui/model_update_trash.go`:
  - `handleOpenTrash()` を削除し、`handleOpenTrashDialog()` に置き換え
  - `handleRestore()` を完全削除（TrashDialog内のRキーハンドラに移行）
  - `handleEmptyTrash()` を完全削除（TrashDialog内のShift+Eハンドラに移行）
  - TrashDialogの生成と表示
- `internal/ui/model_update.go`:
  - trashDialogResultMsg等の新メッセージ処理追加
  - Rキーのハンドラから `handleRestore()` 呼び出しを削除し、常に `handleRenameUI()` を呼び出す

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TrashDialog | ゴミ箱アイテムの一覧表示・ナビゲーション | トラッシュディレクトリが存在 | 画面中央にダイアログ表示 |
| TrashItem | 単一のゴミ箱アイテム情報 | .trashinfoが読み取り可能 | 名前・サイズ・削除日時・元パスを保持 |
| loadTrashItems | トラッシュアイテムの読み込み | トラッシュディレクトリにアクセス可 | TrashItemのスライスを返す |
| cursor | 現在選択中のアイテムインデックス | アイテムが1件以上 | 0〜len(items)-1の範囲 |
| scrollOffset | スクロール位置 | なし | 表示範囲の開始位置 |

**Processing Flow**:
```
1. Tキーが押される
2. handleOpenTrashDialog()が呼ばれる
3. トラッシュディレクトリの存在確認
   ├─ 失敗 → エラーメッセージ表示
   └─ 成功 → 続行
4. トラッシュアイテムの読み込み
   └─ .trashinfoファイルを全件パース
5. TrashDialogを生成
   └─ DialogDisplayScreenタイプで作成
6. m.dialogにセットしてダイアログ表示
```

**Implementation Steps**:

1. **TrashItem構造体の定義**
   - 名前、サイズ、削除日時、元パス、マーク状態を保持
   - Key considerations:
     - ディレクトリのサイズは"-"で表示
     - 削除日時のフォーマット（YYYY-MM-DD HH:MM）

2. **TrashDialogの基本構造**
   - BaseDialogを埋め込み
   - DialogDisplayScreen型を使用
   - スクロール対応（HelpDialogを参考）
   - Key considerations:
     - 幅を70に設定（ヘルプダイアログと同様）
     - visibleHeightでスクロール範囲を管理

3. **View関数の実装**
   - タイトル行：「Trash」とアイテム数[N]
   - ヘッダ行：Name、Size、Deleted、Original Path
   - アイテム行：各カラムを適切な幅で表示
   - フッタ行：キーバインドヒント
   - Key considerations:
     - 列幅の固定または動的調整
     - パスが長い場合の省略表示（~を使用）

4. **handleOpenTrashDialog()の実装**
   - 既存のhandleOpenTrash()を置き換え
   - アイテム読み込みとダイアログ生成
   - Key considerations:
     - 空のゴミ箱でも正常に表示

5. **カーソル移動の実装**
   - j/k/Up/Downで上下移動
   - 境界処理（先頭/末尾でラップまたは停止）
   - Key considerations:
     - HelpDialogのスクロール実装を参考

6. **スクロール処理**
   - カーソルが表示範囲外に出たら自動スクロール
   - visibleHeightを超える場合のみスクロール発生
   - Key considerations:
     - スムーズなスクロール体験

**Dependencies**:
- Requires: 既存のトラッシュインフラ
- Blocks: Phase 2（マーク機能）

**Testing Approach**:

*Unit Tests*:
- TrashDialogの生成が正しい
- Viewが正しいフォーマットで出力
- 空のゴミ箱でも正常動作
- カーソル移動が正しい
- スクロール処理が正しい

*Integration Tests*:
- Tキーでダイアログが開く
- 両ペインが暗転する

*Manual Testing*:
- [ ] Tキーでダイアログが画面中央に表示
- [ ] ゴミ箱アイテムが一覧表示される
- [ ] 列（名前・サイズ・削除日時・元パス）が正しく表示
- [ ] Escでダイアログが閉じる
- [ ] j/kでカーソルが上下移動
- [ ] アイテム数が多い場合にスクロール動作

**Acceptance Criteria**:
- [ ] TキーでTrashDialogが画面中央に表示される
- [ ] 両ペインが暗転（DialogDisplayScreen）
- [ ] アイテム一覧が表示される（Name, Size, Deleted, Original Path）
- [ ] アイテム数がタイトルに表示[N]
- [ ] Esc/qでダイアログが閉じる
- [ ] j/k/Up/Downでカーソル移動
- [ ] アイテム数>visibleHeightの場合にスクロール
- [ ] カーソル位置がハイライト表示

**Estimated Effort**: 中 (4-6 days)

**Risks and Mitigation**:
- **Risk**: 列幅の調整が難しい
  - **Mitigation**: 固定幅を採用し、パスは省略表示

---

### Phase 2: マーク機能

**Goal**: ダイアログ内でのマーク機能を実現する

**Files to Modify**:
- `internal/ui/trash_dialog.go`:
  - Spaceでマーク切り替え

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| marked | マーク状態の管理（map[string]bool） | なし | 選択されたアイテムがtrue |

**Processing Flow**:
```
マーク:
1. Spaceキーが押される
2. 現在カーソル位置のアイテムのマーク状態を切り替え
3. カーソルを1つ下に移動
4. 画面を再描画
```

**Implementation Steps**:

1. **マーク機能**
   - Space押下でマーク切り替え
   - マークされたアイテムを視覚的に区別（先頭に*等）
   - マーク後にカーソル自動移動
   - Key considerations:
     - BookmarkDialogのパターンを参考

**Dependencies**:
- Requires: Phase 1（TrashDialog基本実装・ナビゲーション）
- Blocks: Phase 3（復元・空にする機能）

**Testing Approach**:

*Unit Tests*:
- マーク状態の切り替えが正しい

*Manual Testing*:
- [ ] Spaceでマーク切り替え
- [ ] マークされたアイテムに*が表示
- [ ] マーク後にカーソルが下に移動

**Acceptance Criteria**:
- [ ] Spaceでマーク切り替え（視覚的フィードバック）
- [ ] マークされたアイテムが視覚的に区別される

**Estimated Effort**: 小 (1 day)

---

### Phase 3: 復元と空にする機能

**Goal**: ダイアログ内からの復元・空にする操作を実現する

**Files to Modify**:
- `internal/ui/trash_dialog.go`:
  - Rキーで復元処理（TrashDialog.Update内で処理）
  - Shift+Eで空にする処理（TrashDialog.Update内で処理）
- `internal/ui/model_update_trash.go`:
  - 既存の`handleRestore()`を完全削除
  - 既存の`handleEmptyTrash()`を完全削除
  - TrashDialog専用のメッセージハンドラ追加:
    - `trashDialogRestoreMsg` - ダイアログからの復元要求
    - `trashDialogEmptyMsg` - ダイアログからの空にする要求
  - 既存の`executeRestore()`, `executeRestoreWithOverwrite()`, `executeRestoreWithRename()`, `executeEmptyTrash()`は再利用
- `internal/ui/model_update.go`:
  - Rキーハンドラを単純化（常に`handleRenameUI()`を呼び出し、トラッシュ判定を削除）
  - TrashDialog関連メッセージのルーティング

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| trashDialogRestoreMsg | 復元要求メッセージ | 復元対象のtrashName | 復元処理開始 |
| trashDialogEmptyMsg | 空にする要求メッセージ | 確認済み | EmptyTrash処理開始 |
| restoreSelected | 選択/マークアイテムの復元 | アイテムが選択/マーク済み | 復元実行または衝突ダイアログ |

**Processing Flow**:
```
復元:
1. Rキーが押される
2. マークされたアイテムがあるか確認
   ├─ あり → マーク全件を復元対象
   └─ なし → カーソル位置のアイテムを復元対象
3. 各アイテムについて:
   a. .trashinfoから元パスを取得
   b. 元パスに同名ファイルが存在するか確認
      ├─ 存在 → RestoreConflictDialogを表示
      └─ 不在 → 直接復元
4. 復元成功後にダイアログを更新（アイテム削除）
5. 全件完了後にペインをリフレッシュ

空にする:
1. Shift+Eキーが押される
2. EmptyTrashDialogで確認
3. 確認後にEmptyTrash()を実行
4. TrashDialogを閉じる
5. ペインをリフレッシュ
```

**Implementation Steps**:

1. **復元メッセージの定義**
   - trashDialogRestoreMsg構造体
   - 単一/バッチ復元の区別
   - Key considerations:
     - 既存のrestoreSuccessMsg/restoreErrorMsgを再利用

2. **Rキーハンドラ**
   - マーク有無で対象を決定
   - 衝突チェックと適切なダイアログ表示
   - Key considerations:
     - RestoreConflictDialogを子ダイアログとして表示
     - TrashDialogを閉じずに衝突解決

3. **Shift+Eキーハンドラ**
   - EmptyTrashDialogを表示
   - 確認後にEmptyTrash()実行
   - Key considerations:
     - 既存のemptyTrashResultMsgを再利用

4. **ダイアログ更新処理**
   - 復元成功時にアイテムリストから削除
   - 空になった場合のダイアログ自動クローズ
   - Key considerations:
     - 部分的な復元失敗時の状態管理

**Dependencies**:
- Requires: Phase 2（ナビゲーションとマーク）
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- Rキーで復元メッセージが発行される
- Shift+Eで確認ダイアログが表示される
- マーク有無で復元対象が正しく決定される

*Integration Tests*:
- 復元が元の場所にファイルを戻す
- 衝突時にダイアログが表示される
- EmptyTrashで全ファイルが削除される

*Manual Testing*:
- [ ] Rキーで選択アイテムが復元される
- [ ] Spaceでマーク後、Rでバッチ復元
- [ ] 復元先衝突時にダイアログ表示
- [ ] Shift+Eで確認後にゴミ箱が空になる
- [ ] 復元成功後にダイアログ内のアイテムが更新される

**Acceptance Criteria**:
- [ ] Rキーで選択/マークアイテムが元の場所へ復元
- [ ] 復元先衝突時にRestoreConflictDialog表示
- [ ] Shift+EでEmptyTrashDialog表示後に全削除
- [ ] 復元/削除後にダイアログ内リストが更新される
- [ ] TrashDialog内のRキーはリネームにならない

**Estimated Effort**: 中 (3-5 days)

---

## Complete File Structure

```
internal/
├── fs/
│   ├── trash.go              # 既存: ゴミ箱操作のコア機能
│   ├── trash_test.go         # 既存: ゴミ箱操作のテスト
│   ├── trashinfo.go          # 既存: .trashinfoファイルの生成・パース
│   └── trashinfo_test.go     # 既存: .trashinfoのテスト
└── ui/
    ├── trash_dialog.go       # 新規: TrashDialog実装
    ├── trash_dialog_test.go  # 新規: TrashDialogテスト
    ├── restore_conflict_dialog.go    # 既存: 復元時衝突解決ダイアログ
    ├── empty_trash_dialog.go         # 既存: ゴミ箱を空にする確認ダイアログ
    ├── model_update_trash.go         # 修正: handleOpenTrashDialog追加
    └── model_update.go               # 修正: TrashDialogメッセージ処理追加
```

**File Descriptions**:
- `trash_dialog.go`: TrashDialog本体（BaseDialog埋め込み、DialogDisplayScreen型）
- `model_update_trash.go`: ダイアログ生成とメッセージハンドリング
- 既存ファイルは最小限の修正

## Testing Strategy

### Unit Testing

**Approach**:
- Go標準の`testing`パッケージを使用
- テーブル駆動テストで複数シナリオをカバー

**Test Coverage Goals**:
- TrashDialog: 80%+
- メッセージハンドリング: 80%+

**Key Test Areas**:

1. **TrashDialog生成・表示** (`internal/ui/trash_dialog_test.go`)
   - 正常なダイアログ生成
   - 空のゴミ箱での生成
   - View出力のフォーマット検証

2. **ナビゲーション** (`internal/ui/trash_dialog_test.go`)
   - カーソル移動の境界処理
   - スクロール処理
   - マーク切り替え

3. **キー入力処理** (`internal/ui/trash_dialog_test.go`)
   - Rキーで復元メッセージ発行
   - Shift+Eで確認ダイアログ表示
   - Escでダイアログクローズ

### Manual Testing Checklist

- [ ] Tキーでダイアログが画面中央に表示
- [ ] 両ペインが暗転
- [ ] j/kでカーソル移動
- [ ] Spaceでマーク切り替え
- [ ] Rキーで復元（ダイアログ内）
- [ ] Rキーでリネーム（通常ファイルリスト）
- [ ] Shift+Eで空にする確認
- [ ] Esc/qでダイアログクローズ
- [ ] 空のゴミ箱でも正常表示

## Dependencies

### Internal Dependencies

**Implementation Order**:
1. Phase 1（TrashDialog基本実装・ナビゲーション）
2. Phase 2（マーク機能）
3. Phase 3（復元と空にする機能）

**Component Dependencies**:
- `trash_dialog.go` depends on `dialog_base.go`
- `trash_dialog.go` depends on `internal/fs/trash.go`, `internal/fs/trashinfo.go`
- `model_update_trash.go` depends on `trash_dialog.go`

## Risk Assessment

### Technical Risks

1. **ダイアログ内での子ダイアログ表示**
   - **Risk**: RestoreConflictDialogをTrashDialog内で表示する際の状態管理
   - **Likelihood**: Medium
   - **Impact**: Medium
   - **Mitigation**: 一旦TrashDialogを閉じてRestoreConflictDialogを表示し、完了後にTrashDialogを再開

2. **大量ファイルのスクロール性能**
   - **Risk**: 1000ファイル超で描画遅延
   - **Likelihood**: Low
   - **Impact**: Low（許容範囲）
   - **Mitigation**: シンプルな実装を維持、必要に応じて仮想化を検討

## Performance Considerations

1. **トラッシュアイテム読み込み**
   - ダイアログ表示時に全件読み込み
   - 大量ファイル時の遅延は許容

2. **描画性能**
   - スクロール時は表示範囲のみ再描画
   - lipglossによる効率的なレンダリング

## Security Considerations

1. **パス検証**
   - 既存のvalidateTrashName()を使用
   - ディレクトリトラバーサル防止

2. **権限確認**
   - 既存のトラッシュ操作の権限チェックを継承

## Open Questions

なし - 全ての要件がSPEC.mdで明確化済み

## Success Metrics

### Functional Completeness
- [ ] TrashDialogが画面中央に表示される
- [ ] 両ペインが暗転（DialogDisplayScreen）
- [ ] j/k/Space/R/Shift+Eが正しく動作
- [ ] Rキーがダイアログ内で復元、ダイアログ外でリネーム

### Quality Metrics
- [ ] テストカバレッジ80%+
- [ ] 手動テストで重大なバグなし

### User Experience
- [ ] 直感的なキーボード操作
- [ ] 明確なエラーメッセージ

## References

- **Specification**: `doc/tasks/trash/SPEC.md`
- **Existing Dialogs**: `internal/ui/help_dialog.go`、`internal/ui/bookmark_dialog.go`
- **Existing Trash Infrastructure**: `internal/fs/trash.go`、`internal/fs/trashinfo.go`

## Next Steps

1. **実装開始**
   - Phase 1から順に実装
   - 既存のダイアログパターンを参考

2. **テスト作成**
   - 各フェーズでユニットテストを作成
   - 手動テストチェックリストで検証

3. **既存機能の完全削除と移行**
   - `handleOpenTrash()` を削除し、`handleOpenTrashDialog()` に置き換え
   - `handleRestore()` を完全削除（ダイアログ方式に完全移行）
   - `handleEmptyTrash()` を完全削除（ダイアログ方式に完全移行）
   - Rキーハンドラのトラッシュ判定ロジックを削除（常にリネーム動作）
   - ペインナビゲーションでゴミ箱に移動する機能は提供しない
