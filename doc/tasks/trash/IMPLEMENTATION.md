# Implementation Plan: Trash (Recycle Bin)

## Overview

freedesktop.org Trash Specification準拠のゴミ箱機能を実装する。ファイルを安全にゴミ箱へ移動し、必要に応じて復元できる。

## Objectives

- 誤削除からファイルを保護する
- 削除操作の取り消しを可能にする
- デスクトップ環境（GNOME、KDE等）との互換性を維持する
- キーボード駆動の直感的なゴミ箱操作を提供する

## Prerequisites

### Development Environment
- Go 1.21以上
- make（ビルド自動化用）

### Dependencies
- 既存のdialog infrastructure（`internal/ui/dialog.go`、`internal/ui/confirm_dialog.go`等）
- 既存のTaskManager（`internal/archive/task_manager.go`）をクロスファイルシステム操作に活用
- 既存のファイル操作（`internal/fs/operations.go`）

### Knowledge Requirements
- freedesktop.org Trash Specification
- Bubble Teaアーキテクチャ（Model-Update-View）
- 既存のキーバインドシステム（`internal/ui/actions.go`、`internal/config/defaults.go`）

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Libraries**:
  - `os`、`path/filepath` - ファイル操作
  - `net/url` - URLエンコーディング
  - `time` - ISO 8601タイムスタンプ

### Design Approach
- ゴミ箱操作は`internal/fs/trash.go`に集約
- UIコンポーネントは既存のダイアログパターンを継承
- キーバインドは既存のアクションシステムに統合

### Component Interaction
```
User Input (Delete key)
    |
    v
Model.handleTrashMove()
    |
    v
fs.MoveToTrash()
    |-- Same filesystem: os.Rename
    |-- Cross filesystem: Copy + Delete (via TaskManager)
    v
Generate .trashinfo
    |
    v
Refresh pane
```

## Implementation Phases

### Phase 1: Core Trash Infrastructure

**Goal**: ゴミ箱へのファイル移動とゴミ箱ナビゲーションを実現する

**Files to Create**:
- `internal/fs/trash.go` - ゴミ箱操作のコア機能
- `internal/fs/trash_test.go` - ゴミ箱操作のテスト
- `internal/fs/trashinfo.go` - .trashinfoファイルの生成・パース
- `internal/fs/trashinfo_test.go` - .trashinfoのテスト

**Files to Modify**:
- `internal/ui/actions.go`:
  - 新規アクションの追加（ActionTrash、ActionOpenTrash、ActionRestore、ActionEmptyTrash）
- `internal/config/defaults.go`:
  - デフォルトキーバインドの追加
- `internal/ui/model_update_keyboard.go`:
  - キーハンドラの追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TrashDir | ゴミ箱ディレクトリパスの取得 | なし | `~/.local/share/Trash/`を返す |
| EnsureTrashDirs | files/info/ディレクトリの作成確認 | なし | 両ディレクトリが存在 |
| MoveToTrash | ファイルをゴミ箱へ移動 | ファイルが存在 | ファイルがゴミ箱に移動、.trashinfo生成 |
| GenerateTrashinfo | .trashinfoファイル生成 | 元パスと削除日時 | 有効な.trashinfoファイル |
| ResolveNameCollision | 衝突時の連番付与 | ゴミ箱内の既存ファイル名 | 一意のファイル名 |

**Processing Flow**:
```
1. Delete keyが押される
2. 選択ファイルのパスを取得
3. ゴミ箱ディレクトリを確認/作成
4. 衝突回避のためファイル名を解決
   ├─ 衝突なし → 元のファイル名を使用
   └─ 衝突あり → 連番付与（file.2.txt, file.3.txt...）
5. .trashinfoファイルを生成
6. ファイルを移動
   ├─ 同一FS → os.Rename
   └─ 異なるFS → コピー + 削除
7. 移動失敗時のロールバック
   └─ 失敗 → .trashinfoを削除してエラー返却
8. ペインをリフレッシュ
```

**Implementation Steps**:

1. **ゴミ箱ディレクトリ管理**
   - ゴミ箱パスの取得と検証
   - files/info/ディレクトリの存在確認と作成
   - Key considerations:
     - XDG_DATA_HOME環境変数のサポート
     - パーミッションエラーの適切な処理

2. **Trashinfo生成・パース**
   - INI形式のファイル生成
   - URLエンコーディング処理
   - ISO 8601タイムスタンプ生成（ローカルタイム、タイムゾーンサフィックスなし）
   - Key considerations:
     - 特殊文字（スペース、日本語等）の正確なエンコード
     - パース時のエラー耐性
     - DeletionDateはローカルタイムで記録（freedesktop仕様準拠）

3. **MoveToTrash実装**
   - 名前衝突解決ロジック
   - 同一ファイルシステム判定と適切な移動方法選択
   - ロールバック処理（移動失敗時に.trashinfoを削除）
   - Key considerations:
     - シンボリックリンクはリンク自体を移動（ターゲットを追跡しない）
     - 複数ファイル選択時の一括処理
     - 移動失敗時は生成した.trashinfoを削除してクリーンな状態に戻す

4. **キーバインドとUI統合**
   - ActionTrash、ActionOpenTrashの追加
   - Deleteキー、Tキーのハンドラ実装
   - Key considerations:
     - 既存のActionDeleteとの区別（dキーは直接削除を維持）

**Dependencies**:
- Requires: なし（フェーズ1は独立）
- Blocks: Phase 2（復元機能はゴミ箱インフラに依存）

**Testing Approach**:

*Unit Tests*:
- Trashinfo生成が正しいフォーマットを出力
- URLエンコードが特殊文字を正しく処理
- ISO 8601タイムスタンプが正確
- 名前衝突時に正しい連番が付与される

*Integration Tests*:
- 単一ファイルの移動が成功
- ディレクトリの再帰的移動が成功
- クロスファイルシステム移動が成功
- 権限エラーが適切に処理される
- 移動失敗時に.trashinfoがロールバック削除される

*Manual Testing*:
- [ ] Deleteキーでファイルがゴミ箱へ移動
- [ ] Tキーでゴミ箱ディレクトリが開く
- [ ] 同名ファイル衝突時に連番が付与される

**Acceptance Criteria**:
- [ ] DeleteキーでファイルがTrash/files/に移動される
- [ ] 対応する.trashinfoがTrash/info/に生成される
- [ ] 同名衝突時に.2, .3...の連番が付与される
- [ ] Tキーでゴミ箱ディレクトリに移動できる
- [ ] 同一FS移動は100ms未満で完了

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:
- **Risk**: クロスファイルシステム判定の複雑さ
  - **Mitigation**: 既存のMoveFile()のフォールバックロジックを活用
- **Risk**: XDG_DATA_HOME未設定時の動作
  - **Mitigation**: デフォルト値（~/.local/share）へのフォールバック

---

### Phase 2: Restore and Management

**Goal**: ゴミ箱からの復元とゴミ箱管理機能を実現する

**Files to Create**:
- `internal/ui/restore_conflict_dialog.go` - 復元時衝突解決ダイアログ
- `internal/ui/restore_conflict_dialog_test.go` - ダイアログのテスト
- `internal/ui/empty_trash_dialog.go` - ゴミ箱を空にする確認ダイアログ
- `internal/ui/empty_trash_dialog_test.go` - ダイアログのテスト

**Files to Modify**:
- `internal/fs/trash.go`:
  - RestoreFromTrash関数の追加
  - EmptyTrash関数の追加
- `internal/fs/trash_test.go`:
  - 復元とEmptyのテスト追加
- `internal/ui/pane.go`:
  - ゴミ箱内判定メソッド
- `internal/ui/pane_render.go`:
  - ゴミ箱表示時の追加列レンダリング
- `internal/ui/model_update_keyboard.go`:
  - R、Shift+Eキーハンドラ

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| RestoreFromTrash | ファイルを元の場所へ復元 | ファイルがゴミ箱内、.trashinfoあり | ファイルが元パスに復元、.trashinfo削除 |
| EmptyTrash | ゴミ箱内全ファイル削除 | なし | files/とinfo/が空 |
| IsInTrash | 現在のパスがゴミ箱内か判定 | なし | true/false |
| RestoreConflictDialog | 衝突時の選択肢提示 | 復元先に同名ファイル存在 | ユーザー選択（上書き/リネーム/スキップ） |
| EmptyTrashDialog | 確認ダイアログ表示 | なし | ユーザー確認（Yes/No） |

**Processing Flow**:
```
Restore Flow:
1. Rキーが押される
2. ゴミ箱内か確認
   └─ No → 何もしない
3. .trashinfoから元パスを読み取り
4. 元パスに同名ファイルが存在するか確認
   ├─ No → 直接復元
   └─ Yes → 衝突ダイアログ表示
5. ユーザー選択に応じて処理
   ├─ 上書き → 既存ファイル削除後に復元
   ├─ リネーム → 連番付与して復元
   └─ スキップ → 操作中止
6. .trashinfoファイル削除
7. ペインをリフレッシュ

Empty Trash Flow:
1. Shift+Eキーが押される
2. ゴミ箱内か確認
   └─ No → 何もしない
3. 確認ダイアログ表示
4. ユーザー確認
   ├─ Yes → files/とinfo/内の全ファイル削除
   └─ No → 操作中止
5. ペインをリフレッシュ
```

**Implementation Steps**:

1. **ゴミ箱内判定**
   - IsInTrashメソッドの実装
   - Key considerations:
     - パス比較時の正規化（末尾スラッシュ等）

2. **RestoreFromTrash実装**
   - .trashinfoパースと元パス取得
   - 元ディレクトリが存在しない場合の作成
   - Key considerations:
     - 元ディレクトリの権限確認
     - クロスファイルシステム復元

3. **衝突解決ダイアログ**
   - 既存のdialog_base.goパターンに従う
   - O/R/Sキーで選択
   - Key considerations:
     - 複数ファイル復元時のバッチ処理

4. **EmptyTrash実装**
   - files/とinfo/の全エントリ削除
   - Key considerations:
     - 削除中のエラー継続処理

5. **ゴミ箱表示用追加列**
   - 元パスと削除日時の列表示
   - .trashinfoからのメタデータ読み取り（全件読み取り方式）
   - Key considerations:
     - シンプルな全件読み取りを採用（大量ファイル時の遅延は許容）
     - 列幅の動的調整

**Dependencies**:
- Requires: Phase 1（ゴミ箱インフラ）
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- .trashinfoパースが正しく元パスを取得
- 元ディレクトリ不在時に作成される
- 衝突解決の各オプションが正しく動作

*Integration Tests*:
- 復元が元の場所に正しくファイルを戻す
- 衝突時のダイアログが正しく表示される
- EmptyTrashが全ファイルを削除

*Manual Testing*:
- [ ] Rキーで選択ファイルが元の場所へ復元
- [ ] 復元先に同名ファイル存在時にダイアログ表示
- [ ] Shift+Eで確認後にゴミ箱が空になる
- [ ] ゴミ箱表示時に元パスと削除日時が表示

**Acceptance Criteria**:
- [ ] Rキーで.trashinfoに記録された元パスへ復元される
- [ ] 復元先衝突時に上書き/リネーム/スキップが選択できる
- [ ] Shift+Eで確認ダイアログが表示される
- [ ] 確認後にゴミ箱内の全ファイルが削除される
- [ ] ゴミ箱内で元パスと削除日時が列として表示される

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:
- **Risk**: 大量ファイルのメタデータ読み取りによるパフォーマンス低下
  - **Mitigation**: 遅延読み込みまたはキャッシュの検討
- **Risk**: 復元先ディレクトリが削除されている場合
  - **Mitigation**: 親ディレクトリを再帰的に作成

---

### Phase 3: Extended (External Device Support)

**Goal**: 外部デバイスでのゴミ箱サポートを実現する

**Files to Modify**:
- `internal/fs/trash.go`:
  - デバイスごとのゴミ箱パス解決
  - .Trash-$UIDディレクトリの処理

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| GetTrashDirForPath | パスに応じたゴミ箱ディレクトリ取得 | ファイルパス | 適切なゴミ箱パス |
| CreateDeviceTrash | 外部デバイス上にゴミ箱作成 | デバイスマウントポイント | .Trash-$UIDが存在 |

**Processing Flow**:
```
1. ファイルパスからマウントポイントを特定
2. マウントポイントがホームと同一FSか確認
   ├─ Yes → ~/.local/share/Trash/を使用
   └─ No → マウントポイント/.Trash-$UID/を使用
3. デバイス固有のゴミ箱が存在しない場合は作成
4. 通常のゴミ箱操作を実行
```

**Implementation Steps**:

1. **マウントポイント検出**
   - ファイルパスからマウントポイントを特定
   - Key considerations:
     - Linux固有の/proc/mountsパース

2. **デバイス別ゴミ箱パス解決**
   - .Trash-$UIDディレクトリの検出・作成
   - Key considerations:
     - セキュリティ（.Trashディレクトリのパーミッション確認）

**Dependencies**:
- Requires: Phase 1、Phase 2
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- マウントポイント検出が正確
- デバイス別ゴミ箱パスが正しく解決される

*Integration Tests*:
- 外部デバイス上のファイルが正しくゴミ箱へ移動
- 外部デバイスからの復元が正しく動作

*Manual Testing*:
- [ ] USBドライブ上のファイルがデバイス内ゴミ箱へ移動
- [ ] デバイス内ゴミ箱からの復元が成功

**Acceptance Criteria**:
- [ ] 外部デバイス上のファイルが.Trash-$UIDへ移動される
- [ ] デバイスごとにゴミ箱が分離される

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: マウントポイント検出の信頼性
  - **Mitigation**: /proc/mountsに加えてstatfsのデバイスID比較

---

## Complete File Structure

```
internal/
├── fs/
│   ├── trash.go              # ゴミ箱操作のコア機能
│   ├── trash_test.go         # ゴミ箱操作のテスト
│   ├── trashinfo.go          # .trashinfoファイルの生成・パース
│   └── trashinfo_test.go     # .trashinfoのテスト
├── ui/
│   ├── actions.go            # (modified) 新規アクション追加
│   ├── restore_conflict_dialog.go     # 復元時衝突解決ダイアログ
│   ├── restore_conflict_dialog_test.go
│   ├── empty_trash_dialog.go          # ゴミ箱を空にする確認ダイアログ
│   ├── empty_trash_dialog_test.go
│   ├── pane.go               # (modified) IsInTrashメソッド追加
│   ├── pane_render.go        # (modified) ゴミ箱表示用列追加
│   └── model_update_keyboard.go # (modified) キーハンドラ追加
└── config/
    └── defaults.go           # (modified) デフォルトキーバインド追加
```

**File Descriptions**:
- `trash.go`: MoveToTrash、RestoreFromTrash、EmptyTrash等のコア操作
- `trashinfo.go`: .trashinfoファイルのINI形式生成・パース、URLエンコード処理
- `restore_conflict_dialog.go`: 復元時衝突解決のUI（上書き/リネーム/スキップ）
- `empty_trash_dialog.go`: ゴミ箱を空にする確認ダイアログ（既存ConfirmDialogをラップ）
- `actions.go`: ActionTrash、ActionOpenTrash、ActionRestore、ActionEmptyTrash追加
- `defaults.go`: Delete→trash、T→open_trash、R→restore、Shift+E→empty_trash

## Testing Strategy

### Unit Testing

**Approach**:
- Go標準の`testing`パッケージを使用
- テーブル駆動テストで複数シナリオをカバー
- ファイルシステム操作は一時ディレクトリを使用

**Test Coverage Goals**:
- ゴミ箱操作（trash.go）: 90%+
- Trashinfoパース（trashinfo.go）: 95%+
- UIコンポーネント: 70%+

**Key Test Areas**:

1. **Trashinfo生成・パース** (`internal/fs/trashinfo_test.go`)
   - 有効な.trashinfoフォーマット生成
   - 特殊文字のURLエンコード
   - ISO 8601タイムスタンプ
   - 不正なファイルのエラーハンドリング

2. **名前衝突解決** (`internal/fs/trash_test.go`)
   - 衝突なし: 元のファイル名使用
   - 初回衝突: .2を付与
   - 複数衝突: カウンタをインクリメント
   - 拡張子付きファイルの正しい処理

3. **ゴミ箱操作** (`internal/fs/trash_test.go`)
   - 単一ファイル移動
   - ディレクトリ再帰移動
   - クロスファイルシステム移動
   - 復元操作
   - EmptyTrash操作

### Integration Testing

**Scenarios**:
1. Delete→Restore往復テスト
2. 複数ファイル選択時のバッチ操作
3. ゴミ箱表示時のメタデータ読み取り

**Approach**:
- 一時ディレクトリでゴミ箱環境を構築
- 実際のファイルシステム操作を検証

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Deleteキーでファイルがゴミ箱へ移動
- [ ] Tキーでゴミ箱ディレクトリが開く
- [ ] Rキーでファイルが元の場所へ復元（ゴミ箱内）
- [ ] Rキーがゴミ箱外で無効
- [ ] Shift+Eで確認後にゴミ箱が空になる
- [ ] 元パス列がゴミ箱内で表示
- [ ] 削除日時列がゴミ箱内で表示
- [ ] Unicode文字を含むファイル名の処理
- [ ] 長いパス名の処理
- [ ] シンボリックリンクの移動（リンク自体を移動）
- [ ] 空のゴミ箱でShift+E押下
- [ ] 復元時に元の親ディレクトリが削除されている場合

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | existing | TUIフレームワーク | already installed |
| github.com/charmbracelet/lipgloss | existing | スタイリング | already installed |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1（ゴミ箱インフラ）- 独立
2. Phase 2（復元と管理）- Phase 1に依存
3. Phase 3（外部デバイス）- Phase 1, 2に依存

**Component Dependencies**:
- `trash.go` depends on `trashinfo.go`
- `restore_conflict_dialog.go` depends on `dialog_base.go`
- `pane_render.go` depends on `trash.go` (IsInTrash)
- `model_update_keyboard.go` depends on `actions.go`

## Risk Assessment

### Technical Risks

1. **クロスファイルシステム操作の複雑さ**
   - **Risk**: 異なるFS間の移動でエラーが発生
   - **Likelihood**: Medium
   - **Impact**: High
   - **Mitigation**:
     - 既存のMoveFile()のフォールバックロジックを活用
     - TaskManagerによるプログレス表示

2. **大量ファイルのメタデータ読み取り**
   - **Risk**: 1000ファイル超のゴミ箱で表示遅延
   - **Likelihood**: Medium
   - **Impact**: Low（許容される遅延）
   - **Mitigation**:
     - シンプルな全件読み取りを採用
     - 大量ファイル時の遅延は許容する設計方針

3. **パス正規化の一貫性**
   - **Risk**: 末尾スラッシュ等でゴミ箱内判定が失敗
   - **Likelihood**: Low
   - **Impact**: Medium
   - **Mitigation**:
     - filepath.Cleanによる正規化
     - 比較前のパス正規化を徹底

### Implementation Risks

1. **既存キーバインドとの衝突**
   - **Risk**: Delete、R等の既存アクションとの整合性
   - **Mitigation**:
     - ActionDeleteを直接削除として維持（dキー）
     - ActionTrashを新規追加（Deleteキー）
     - Rキーはコンテキスト依存: ゴミ箱内ではrestore、ゴミ箱外ではrename
     - ゴミ箱内ではrenameアクションを無効化

2. **デスクトップ環境との互換性**
   - **Risk**: freedesktop仕様への準拠不足
   - **Mitigation**: 仕様書を厳密に参照、他のファイルマネージャでの動作確認

3. **エラー時の暗黙的フォールバック**
   - **Risk**: ゴミ箱使用不可時に直接削除にフォールバックするとデータ損失の危険
   - **Mitigation**:
     - ゴミ箱操作失敗時はエラー表示して操作を中止
     - 直接削除へのフォールバックは行わない
     - ユーザーが直接削除を望む場合は明示的に`d`キーを使用

## Performance Considerations

1. **ゴミ箱移動**
   - 同一FS: os.Rename（即座）
   - 異なるFS: コピー+削除（TaskManagerで進捗表示）

2. **ゴミ箱表示**
   - .trashinfoの全件読み取り方式を採用
   - 大量ファイル時の遅延は許容（シンプルさを優先）

3. **EmptyTrash**
   - os.RemoveAllによる効率的な削除
   - 大量ファイル時は進捗表示を検討

## Security Considerations

1. **パス検証**
   - ディレクトリトラバーサル防止
   - filepath.Cleanによる正規化

2. **権限確認**
   - ゴミ箱ディレクトリへの書き込み権限
   - 復元先ディレクトリへの書き込み権限

3. **シンボリックリンク**
   - リンク自体を移動（ターゲットを追跡しない）
   - リンク切れの適切な処理

4. **URLエンコード**
   - .trashinfoのPathフィールドを正しくエンコード/デコード
   - 特殊文字によるインジェクション防止

## Open Questions

### From Specification:
- なし（すべての要件が明確化済み）

### Implementation-Specific:
- 外部デバイス対応の詳細仕様（Phase 3で対応）

## Future Enhancements

Items deferred to later phases or releases:

### Out of Scope (per SPEC):
- ゴミ箱の自動削除（期間経過後の自動クリーンアップ）
- ゴミ箱サイズ制限
- ネットワークドライブ対応
- Undo機能（直前の削除を即座に取り消し）

## Success Metrics

### Functional Completeness
- [ ] Phase 1の全機能要件が実装済み
- [ ] Phase 2の全機能要件が実装済み
- [ ] freedesktop.org Trash Specification準拠

### Quality Metrics
- [ ] テストカバレッジが目標達成（90%+ for core）
- [ ] 手動テストで重大なバグなし
- [ ] Goベストプラクティスに準拠

### Performance Metrics
- [ ] 同一FSゴミ箱移動 < 100ms
- [ ] ゴミ箱一覧表示が正常に動作（大量ファイル時の遅延は許容）
- [ ] キーボード入力への応答 < 100ms

### User Experience
- [ ] 直感的なキーボード操作
- [ ] 明確なエラーメッセージ
- [ ] デスクトップ環境との相互運用性

## References

- **Specification**: `doc/tasks/trash/SPEC.md`
- **Requirements**: `doc/tasks/trash/要件定義書.md`
- **freedesktop.org Trash Specification**: https://specifications.freedesktop.org/trash-spec/trashspec-latest.html
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - ステークホルダーレビュー
   - オープンクエスチョンの解決
   - アプローチとタイムラインの確認

2. **Environment Setup**
   - 依存関係の確認
   - 開発環境のセットアップ

3. **Begin Implementation**
   - Phase 1から開始
   - TDDアプローチ（テストを先に書く）
   - インクリメンタルにコミット

4. **Continuous Integration**
   - CIパイプラインでテスト自動実行
   - コード品質チェックの適用
