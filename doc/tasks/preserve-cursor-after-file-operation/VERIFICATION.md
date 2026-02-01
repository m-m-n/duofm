# 検証ドキュメント: ファイル操作後のカーソル位置保持

## 概要
**機能**: ファイル操作後のカーソル位置保持
**SPEC.md**: `doc/tasks/preserve-cursor-after-file-operation/SPEC.md`
**IMPLEMENTATION.md**: `doc/tasks/preserve-cursor-after-file-operation/IMPLEMENTATION.md`
**Date:** 2026-02-01
**Status:** Implementation Complete
**All Tests:** PASS

## Implementation Summary

ファイル操作後のカーソル位置保持を改善する2つの変更を実装した。

1. `RefreshDirectoryPreserveCursor` のフォールバックを cursor=0 からインデックスベースのクランプに変更
2. バッチ移動操作用の `calculateCursorTargetAfterBatchMove` メソッドを追加し、マークされていない最近接ファイルにカーソルを移動

### Phase Summary
- [x] Phase 1: RefreshDirectoryPreserveCursor のフォールバック改善
- [x] Phase 2: バッチ操作用カーソルターゲット計算

## ビルド検証

### ビルドコマンド
```bash
$ make build
go build -ldflags "..." -o ./duofm ./cmd/duofm
```
**Result:** Build successful

## テスト検証

### テストコマンド
```bash
$ go test ./internal/ui/... -v -cover -run "TestRefreshDirectoryPreserveCursor|TestCalculateCursorTargetAfterBatchMove"
```

### 全テスト実行
```bash
$ go test ./...
ok  github.com/sakura/duofm/internal/ui
ok  github.com/sakura/duofm/internal/config
ok  github.com/sakura/duofm/internal/fs
ok  github.com/sakura/duofm/internal/archive
ok  github.com/sakura/duofm/internal/filter
ok  github.com/sakura/duofm/internal/version
ok  github.com/sakura/duofm/test
```
**Result:** All tests PASS

### カバレッジ目標
- **最小**: 80%（対象ファイル）
- **目標**: `calculateCursorTargetAfterBatchMove` は 100%

### SPEC.md からのテストシナリオ

| ID | シナリオ | 期待結果 | テスト種別 | テストファイル | 結果 |
|----|---------|---------|-----------|-------------|------|
| TS-1 | RefreshDirectoryPreserveCursor - ファイル名マッチ成功 | ファイル名のインデックスにカーソル移動 | Unit | pane_filter_test.go | PASS |
| TS-2 | RefreshDirectoryPreserveCursor - ファイル名マッチ失敗・インデックス有効 | 旧インデックスを保持 | Unit | pane_filter_test.go | PASS |
| TS-3 | RefreshDirectoryPreserveCursor - ファイル名マッチ失敗・インデックス超過 | 最後のエントリにクランプ | Unit | pane_filter_test.go | PASS |
| TS-4 | RefreshDirectoryPreserveCursor - 全ファイル削除 | cursor=0 | Unit | pane_filter_test.go | PASS |
| TS-5 | calculateCursorTargetAfterBatchMove - カーソル上に非マークあり | そのファイル名を返す | Unit | pane_cursor_test.go | PASS |
| TS-6 | calculateCursorTargetAfterBatchMove - カーソル上が全マーク(..)スキップ | 下方向の非マークファイル名を返す | Unit | pane_cursor_test.go | PASS |
| TS-7 | calculateCursorTargetAfterBatchMove - 全ファイルマーク済み | 空文字を返す | Unit | pane_cursor_test.go | PASS |
| TS-8 | calculateCursorTargetAfterBatchMove - カーソルが非マーク上 | そのファイル名を返す | Unit | pane_cursor_test.go | PASS |
| TS-9 | calculateCursorTargetAfterBatchMove - 単一ファイルのみマーク | 空文字を返す | Unit | pane_cursor_test.go | PASS |
| TS-10 | calculateCursorTargetAfterBatchMove - マーク間に非マークファイルあり | カーソル位置のファイル名を返す | Unit | pane_cursor_test.go | PASS |
| TS-11 | calculateCursorTargetAfterBatchMove - カーソル位置0（..上） | 空文字を返す | Unit | pane_cursor_test.go | PASS |

## コード品質検証

### フォーマットチェック
```bash
$ gofmt -l ./internal/ui/pane_filter.go ./internal/ui/pane.go ./internal/ui/model_update.go
```
**Result:** No output (all formatted)

### 静的解析
```bash
$ go vet ./...
```
**Result:** No warnings

## ファイル構成検証

### ファイルサイズ

| ファイル | 行数 | ステータス |
|---------|------|----------|
| `internal/ui/pane_filter.go` | 297 | OK |
| `internal/ui/pane.go` | 498 | OK |
| `internal/ui/model_update.go` | 447 | OK |
| `internal/ui/pane_filter_test.go` | 580 | OK |
| `internal/ui/pane_cursor_test.go` | 118 | OK |

### 新規作成ファイル
- `internal/ui/pane_cursor_test.go` - calculateCursorTargetAfterBatchMove のテスト

### 変更ファイル
- `internal/ui/pane_filter.go` - RefreshDirectoryPreserveCursor のフォールバック改善
- `internal/ui/pane_filter_test.go` - フォールバックテスト追加
- `internal/ui/pane_test.go` - 既存テストの期待値更新
- `internal/ui/pane.go` - calculateCursorTargetAfterBatchMove メソッド追加
- `internal/ui/model_update.go` - batchCompleteMsg/batchCancelledMsg ハンドラ更新（move/copy分岐追加）

### 検証コマンド
```bash
# 新規ファイルの存在確認
test -f internal/ui/pane_cursor_test.go && echo "OK" || echo "MISSING"

# 変更ファイルに対象関数が含まれるか確認
grep -q "calculateCursorTargetAfterBatchMove" internal/ui/pane.go && echo "OK" || echo "MISSING"
grep -q "oldCursor" internal/ui/pane_filter.go && echo "OK" || echo "MISSING"
```

## SPEC.md 準拠確認

### 成功基準

| ID | SPEC.md の成功基準 | 検証方法 | 結果 |
|----|-------------------|---------|------|
| SC-1 | 全ユニットテスト合格 | `go test ./internal/ui/... -v` | PASS |
| SC-2 | 単一ファイル移動後のカーソル保持が正しく動作する | TS-1 ~ TS-4 のテスト合格 + 手動テスト | PASS (Unit) |
| SC-3 | バッチ操作後のカーソル位置計算が正しく動作する | TS-5 ~ TS-11 のテスト合格 + 手動テスト | PASS (Unit) |
| SC-4 | 既存の削除操作のカーソル動作に影響しない | `go test ./internal/ui/... -run TestCalculateCursorAfterDeletion -v` | PASS |
| SC-5 | 既存テストがすべて合格する | `go test ./... -v` | PASS |

### 機能要件カバレッジ

| 要件 | 実装フェーズ | 検証方法 | 結果 |
|------|-----------|---------|------|
| FR1: RefreshDirectoryPreserveCursor にインデックスフォールバック追加 | フェーズ 1 | TS-1 ~ TS-4 | PASS |
| FR2: バッチ操作完了時のマーク考慮カーソル計算 | フェーズ 2 | TS-5 ~ TS-11 | PASS |
| FR3: 全操作箇所で LoadDirectory を RefreshDirectoryPreserveCursor に置換 | 先行修正で完了済み | 既存テスト合格 | PASS |
| FR4: スクロール位置をカーソルに合わせて自動調整 | フェーズ 1, 2 | adjustScroll / EnsureCursorVisible が呼ばれることを確認 | PASS |

## 手動テストチェックリスト

### 基本機能

- [ ] 単一ファイル移動後、ソースペインのカーソルが同じインデックス位置に留まる（次のファイルが繰り上がる）
- [ ] 単一ファイル移動後、移動先ペインのカーソルが操作前の位置を保持する
- [ ] 末尾ファイル移動後、カーソルが最後のエントリに移動する
- [ ] コピー操作後、ソースペインのカーソルが元の位置を保持する
- [ ] コピー操作後、移動先ペインのカーソルが元の位置を保持する
- [ ] リネーム後、基盤のリロードでカーソル位置が保持される

### バッチ操作

- [ ] バッチ移動後、カーソルがマークされていない最近接ファイル（上方向優先）に位置する
- [ ] マークファイルより上に非マークファイルがない場合、下方向の最初の非マークファイルに位置する
- [ ] 全ファイルバッチ移動後、カーソルが 0（".."）に位置する
- [ ] バッチ操作キャンセル後も、カーソルが適切な位置に移動する

### エッジケース

- [ ] `..` と1ファイルのみのディレクトリで、そのファイルをマーク移動後 cursor=0
- [ ] エントリが `..` のみのディレクトリで cursor=0
- [ ] 連続する非マークファイルの間にマークファイルがある場合の正しいカーソル移動

### 既存動作の非破壊確認

- [ ] 単一ファイル削除後のカーソル動作が変わらない
- [ ] バッチ削除後のカーソル動作が変わらない
- [ ] ディレクトリ移動（Enter, Backspace）のカーソル動作が変わらない
- [ ] 隠しファイル表示切替後のカーソル動作が変わらない

## 性能検証

### 要件
- NFR1: カーソル位置計算は O(n)（n = エントリ数）以内

### 検証方法
- `calculateCursorTargetAfterBatchMove` は entries を最大2回走査するため O(n) を満たす
- `RefreshDirectoryPreserveCursor` のファイル名検索は entries を1回走査するため O(n) を満たす
- コードレビューで確認（ベンチマーク不要）

## 回帰テスト

### 既存テストの全合格確認
```bash
go test ./... -v
```
**Result:** All tests PASS

### 特に注意すべき既存テスト
```bash
# 削除後カーソル計算（影響を受けないことを確認）
go test ./internal/ui/... -run TestCalculateCursorAfterDeletion -v

# フィルタ関連（RefreshDirectoryPreserveCursor の変更の影響確認）
go test ./internal/ui/... -run TestLoadEntriesFromDisk -v

# バッチ操作マネージャ
go test ./internal/ui/... -run "TestBatchOperation" -v
```
**Result:** All PASS

## 検証サマリ

| カテゴリ | 項目数 | 自動 | 手動 |
|---------|-------|------|------|
| ビルド | 1 | PASS | - |
| ユニットテスト | 11 | PASS | - |
| コード品質 | 2 | PASS | - |
| ファイル構成 | 5 | PASS | - |
| SPEC 準拠 | 5 | PASS | 手動テスト待ち |
| 手動テスト | 14 | - | 手動テスト待ち |
| 性能 | 1 | - | コードレビュー済み |
| 回帰テスト | 3 | PASS | - |

**合計**: 自動検証 22 項目 PASS、手動検証 15 項目待ち
