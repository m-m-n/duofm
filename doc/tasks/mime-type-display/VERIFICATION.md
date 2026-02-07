# Verification: MIME Type Display in Status Bar

## Build Verification

- [x] `go build ./...` が成功する
- [x] コンパイルエラーがない

## Unit Tests

### entryTypeLabel 関数

- [x] ディレクトリエントリで `("Directory", DirectoryFg)` が返される
- [x] シンボリックリンクエントリで `("SymbolicLink", SymlinkFg)` が返される
- [x] `.html` ファイルで `("text/html", StatusFg)` が返される
- [x] `.txt` ファイルで `("text/plain", StatusFg)` が返される
- [x] 拡張子なしファイルで `("application/octet-stream", StatusFg)` が返される
- [x] 親ディレクトリ（`..`、IsDir=true）で `("Directory", DirectoryFg)` が返される
- [x] シンボリックリンク先がディレクトリ（IsSymlink=true, IsDir=true）で `("SymbolicLink", SymlinkFg)` が返される
- [x] nil エントリで `("", StatusFg)` が返される

### renderStatusBar

- [x] 通常ファイル（.html）のステータスバーに `[text/html]` が含まれる
- [x] ディレクトリのステータスバーに `[Directory]` が含まれる
- [x] シンボリックリンクのステータスバーに `[SymbolicLink]` が含まれる
- [x] 空ディレクトリ（エントリ0件）ではMIMEタイプ表示が省略される
- [x] ステータスメッセージ表示中はMIMEタイプ表示が含まれない

## Test Commands

```bash
# 全テスト実行
go test ./internal/ui/... -run "TestEntryTypeLabel|TestRenderStatusBar" -v

# model_view_test.go のみ
go test ./internal/ui/ -run "TestEntryTypeLabel" -v
go test ./internal/ui/ -run "TestRenderStatusBar" -v

# 既存テストの回帰確認
go test ./internal/ui/... ./internal/config/...
```

## Regression Tests

- [x] `go test ./internal/ui/...` が全て通過する
- [x] `go test ./internal/config/...` が全て通過する
- [x] 既存のステータスバー機能（メッセージ表示、キーヒント）に影響がない

## Manual Verification

### 基本表示確認

1. duofm を起動する
2. 通常ファイル（例: `.go` ファイル）にカーソルを合わせる → ステータスバーに `[text/x-go]` 等のMIMEタイプが表示される
3. ディレクトリにカーソルを合わせる → `[Directory]` が青文字で表示される
4. シンボリックリンクにカーソルを合わせる → `[SymbolicLink]` がシアン文字で表示される
5. 拡張子なしファイル（例: Makefile）にカーソルを合わせる → `[application/octet-stream]` が表示される

### レイアウト確認

6. ターミナル幅を十分に広くする → MIMEタイプ表示、位置情報、キーヒントが全て表示される
7. ターミナル幅を極端に狭くする → MIMEタイプ表示が省略され、レイアウトが破綻しない

### ステータスメッセージとの共存

8. ファイル操作（コピー等）を実行してステータスメッセージを表示する → MIMEタイプ表示が非表示になる
9. ステータスメッセージが消えた後 → MIMEタイプ表示が復帰する

### 色の確認

10. ディレクトリの `[Directory]` がテーマの `directory_fg` 色で表示される
11. シンボリックリンクの `[SymbolicLink]` がテーマの `symlink_fg` 色で表示される
12. 通常ファイルのMIMEタイプがステータスバーの前景色で表示される
