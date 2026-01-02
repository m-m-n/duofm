# 実装計画: コンテキストメニューのCompress機能バグ修正

**作成日**: 2026-01-02
**仕様書**: `doc/tasks/fix-archive-compress-bug/SPEC.md`
**ステータス**: Ready for Implementation

---

## 1. 概要

### 1.1 目的

`model.go` の `contextMenuResultMsg` ハンドラを修正し、`actionID` ベースの処理を `result.action != nil` の条件チェックの外側に移動する。

### 1.2 修正方針

現在の問題:
- `compress` と `delete` メニュー項目は `Action: nil` で定義されている
- しかし、actionID ベースの処理が `if result.action != nil` ブロック内にあるため到達不可

解決策:
- actionID ベースの処理を条件ブロックの外側に移動
- `result.action != nil` ブロックはカスタムアクション（汎用処理）専用に変更

---

## 2. 実装フェーズ

### フェーズ1: model.go の修正（推定工数: 小）

#### 1.1 変更対象

**ファイル**: `internal/ui/model.go`
**対象行**: 156-238行（contextMenuResultMsg ハンドラ）

#### 1.2 変更内容

**Before (現在の構造)**:
```
if result.action != nil {
    activePane := ...
    markedFiles := ...

    if result.actionID == "delete" { ... }
    if result.actionID == "compress" { ... }
    if result.actionID == "extract" { ... }
    if result.actionID == "copy" || result.actionID == "move" { ... }

    // 汎用アクション実行
    if err := result.action(); err != nil { ... }

    // ペイン再読み込み
}
```

**After (修正後の構造)**:
```
activePane := m.getActivePane()
markedFiles := activePane.GetMarkedFiles()

// 1. actionID ベースの処理（action が nil でも動作）
if result.actionID == "delete" { ... return }
if result.actionID == "compress" { ... return }
if result.actionID == "extract" { ... return }
if result.actionID == "copy" || result.actionID == "move" { ... return }

// 2. 汎用アクション実行（action が必要）
if result.action != nil {
    if err := result.action(); err != nil { ... }

    // ペイン再読み込み
    activePane.LoadDirectory()
    m.getInactivePane().LoadDirectory()
}
```

#### 1.3 具体的な修正手順

1. `if result.action != nil {` の行（166行）を削除
2. `activePane` と `markedFiles` の取得を条件ブロックの外側に移動
3. 各 actionID 処理はそのまま維持（既に return している）
4. 225-233行の汎用アクション処理を新しい `if result.action != nil` ブロックに移動
5. 閉じ括弧の調整

#### 1.4 検証ポイント

- [ ] compress: `NewCompressFormatDialog()` が呼ばれること
- [ ] delete: 確認ダイアログが表示されること
- [ ] extract: セキュリティチェックが実行されること
- [ ] copy/move: ファイル競合チェックが実行されること
- [ ] 汎用アクション: `result.action()` が呼ばれること

---

### フェーズ2: テスト追加（推定工数: 小）

#### 2.1 テストケース

**ファイル**: `internal/ui/model_test.go`（新規または追記）

| テストケース | 説明 | 期待結果 |
|-------------|------|---------|
| TestContextMenu_CompressWithNilAction | action=nil で compress 選択 | CompressFormatDialog が表示される |
| TestContextMenu_DeleteWithNilAction | action=nil で delete 選択 | ConfirmDialog が表示される |
| TestContextMenu_ExtractWithAction | action 付きで extract 選択 | セキュリティチェックが実行される |
| TestContextMenu_CopyWithAction | action 付きで copy 選択 | ファイル競合チェックが実行される |

#### 2.2 テスト実装方針

- `contextMenuResultMsg` を直接送信して `Update` を呼び出す
- 結果の `m.dialog` の型をアサート
- 既存テストパターンに従う

---

### フェーズ3: 既存テスト確認（推定工数: 小）

#### 3.1 実行するテスト

```bash
# model.go のテスト
go test -v ./internal/ui/... -run "Context"

# アーカイブ関連のテスト
go test -v ./internal/archive/...

# 全テスト
go test -v ./...
```

#### 3.2 確認項目

- [ ] 既存のすべてのテストがパスすること
- [ ] 新規テストがパスすること
- [ ] ビルドエラーがないこと

---

## 3. ファイル構成

### 3.1 変更ファイル

| ファイル | 変更種別 | 説明 |
|---------|---------|------|
| `internal/ui/model.go` | 修正 | contextMenuResultMsg ハンドラの restructure |
| `internal/ui/model_test.go` | 追加/修正 | compress アクションのテスト追加 |

### 3.2 変更なしファイル

| ファイル | 理由 |
|---------|------|
| `internal/ui/context_menu_dialog.go` | 現在の Action: nil 設定は正しい設計 |
| `internal/archive/*` | アーカイブロジックに変更なし |

---

## 4. リスク分析

### 4.1 低リスク

- **変更範囲が限定的**: model.go の1関数内のみ
- **ロジック変更なし**: 処理順序の restructure のみ
- **既存テストで検証可能**: 回帰を検出できる

### 4.2 注意点

- `activePane.GetMarkedFiles()` を条件外で呼び出すことによるパフォーマンス影響
  - 影響: 無視できるレベル（メモリアクセスのみ）

- `enter_logical` や `enter_physical` などの他の actionID 処理
  - 確認: これらは context_menu_dialog.go で Action 付きで定義されているため影響なし

---

## 5. 実装チェックリスト

### Phase 1: コード修正

- [ ] model.go の 166行 `if result.action != nil {` を削除
- [ ] `activePane` と `markedFiles` の取得を条件外に移動
- [ ] 汎用アクション処理を新しい `if result.action != nil` ブロックに移動
- [ ] インデント調整
- [ ] コンパイル確認

### Phase 2: テスト追加

- [ ] TestContextMenu_CompressWithNilAction 追加
- [ ] TestContextMenu_DeleteWithNilAction 追加（オプション）
- [ ] テスト実行・パス確認

### Phase 3: 検証

- [ ] `go test ./...` 全テストパス
- [ ] `go build ./...` ビルド成功
- [ ] 手動テスト: コンテキストメニューから Compress 選択

---

## 6. 手動検証手順

### 6.1 Compress 機能テスト

```bash
# duofm を起動
./duofm

# 手順:
# 1. ファイルまたはディレクトリを選択
# 2. @ キーでコンテキストメニューを開く
# 3. "Compress" を選択
# 4. フォーマット選択ダイアログが表示されることを確認
```

### 6.2 複数ファイル圧縮テスト

```bash
# 手順:
# 1. Space キーで複数ファイルをマーク
# 2. @ キーでコンテキストメニューを開く
# 3. "Compress N files" を選択
# 4. フォーマット選択ダイアログが表示されることを確認
```

### 6.3 他のメニュー項目テスト

- Delete: 確認ダイアログ表示
- Copy: ファイル競合チェック（反対ペインに同名ファイルがある場合）
- Move: ファイル競合チェック
- Extract: アーカイブファイル選択時に展開処理

---

## 7. 完了基準

1. [ ] model.go の修正完了
2. [ ] 新規テスト追加・パス
3. [ ] 既存テスト全パス
4. [ ] 手動テストで Compress 機能動作確認
5. [ ] 他のメニュー項目が正常動作することを確認

---

**Last Updated**: 2026-01-02
**Status**: Ready for Implementation
