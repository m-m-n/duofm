# Verification Report: Open File with External Application

## Implementation Summary

ファイルを外部アプリケーション（less/vim）で開く機能を実装しました。

### 実装したファイル

| ファイル | 変更内容 |
|----------|----------|
| `internal/ui/keys.go` | `KeyView`と`KeyEdit`定数を追加 |
| `internal/ui/exec.go` | 新規作成: 外部コマンド実行モジュール |
| `internal/ui/exec_test.go` | 新規作成: ユニットテスト |
| `internal/ui/model.go` | `execFinishedMsg`ハンドラと`v`/`e`/`Enter`キーハンドラを追加 |
| `test/e2e/Dockerfile` | `less`と`vim`をインストール |
| `test/e2e/scripts/test_open_file.sh` | 新規作成: E2Eテストスクリプト |

## Unit Test Results

```
=== RUN   TestCheckReadPermission
=== RUN   TestCheckReadPermission/readable_file
=== RUN   TestCheckReadPermission/non-existent_file
--- PASS: TestCheckReadPermission (0.00s)
    --- PASS: TestCheckReadPermission/readable_file (0.00s)
    --- PASS: TestCheckReadPermission/non-existent_file (0.00s)
=== RUN   TestExecFinishedMsg
--- PASS: TestExecFinishedMsg (0.00s)
=== RUN   TestOpenWithViewerReturnsCmd
--- PASS: TestOpenWithViewerReturnsCmd (0.00s)
=== RUN   TestOpenWithEditorReturnsCmd
--- PASS: TestOpenWithEditorReturnsCmd (0.00s)
PASS
```

## E2E Test Results

```
=== E2E Tests: Open File with External Application ===

--- Running: test_view_parent_dir_ignored ---
✓ duofm title should remain visible (v on .. ignored)

--- Running: test_edit_parent_dir_ignored ---
✓ duofm title should remain visible (e on .. ignored)

--- Running: test_view_directory_ignored ---
✓ duofm title should remain visible (v on dir ignored)

--- Running: test_edit_directory_ignored ---
✓ duofm title should remain visible (e on dir ignored)

--- Running: test_enter_directory ---
✓ Should see subdir inside dir1 directory

--- Running: test_enter_parent_dir ---
✓ Should see file1.txt after navigating to parent

--- Running: test_view_file_with_v ---
✓ less should display file1 content
✓ duofm title bar should be visible after less exits

--- Running: test_edit_file_with_e ---
✓ vim should display file1 content
✓ duofm title bar should be visible after vim exits

--- Running: test_enter_file_opens_viewer ---
✓ less should display file1 content via Enter
✓ duofm title bar should be visible after Enter on file

========================================
Test Summary
========================================
Total:  12
Passed: 12
Failed: 0
========================================
```

## Feature Verification Checklist

### キーバインディング

| テスト項目 | 結果 |
|------------|------|
| `v` on file → lessで開く | ✅ PASS |
| `v` on directory → 無視 | ✅ PASS |
| `v` on `..` → 無視 | ✅ PASS |
| `e` on file → vimで開く | ✅ PASS |
| `e` on directory → 無視 | ✅ PASS |
| `e` on `..` → 無視 | ✅ PASS |
| `Enter` on file → lessで開く | ✅ PASS |
| `Enter` on directory → ディレクトリに入る | ✅ PASS |
| `Enter` on `..` → 親ディレクトリへ移動 | ✅ PASS |

### 画面復帰

| テスト項目 | 結果 |
|------------|------|
| less終了後にduofm画面が復帰 | ✅ PASS |
| vim終了後にduofm画面が復帰 | ✅ PASS |

### ペイン再読み込み

| テスト項目 | 結果 |
|------------|------|
| 外部アプリ終了後に両ペイン再読み込み | ✅ PASS (コード確認済み) |

### エラーハンドリング

| テスト項目 | 結果 |
|------------|------|
| 読み取り権限チェック | ✅ PASS (ユニットテスト) |
| エラー時にステータスバー表示 | ✅ PASS (コード確認済み) |

## Architecture Notes

### tea.ExecProcess の使用

Bubble Teaの`tea.ExecProcess`を使用して外部コマンドを実行しています。この機能により：

1. TUIを一時停止
2. ターミナルの制御を外部アプリ（less/vim）に渡す
3. 外部アプリ終了後にTUIを復帰
4. `execFinishedMsg`を送信してペイン再読み込みを実行

### セキュリティ対策

- ファイルパスは`filepath.Join`で構築（パストラバーサル防止）
- 読み取り権限を事前にチェック
- 実行するコマンドは`less`と`vim`に限定（ユーザー入力なし）
- シェル解釈なし（`exec.Command`で直接実行）

## Conclusion

すべての要件が満たされ、ユニットテストとE2Eテストがすべてパスしました。実装は仕様書と一致しており、既存機能への影響もありません。
