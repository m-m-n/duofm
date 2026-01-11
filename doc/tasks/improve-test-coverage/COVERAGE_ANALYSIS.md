# テストカバレッジ分析レポート

**作成日**: 2026-01-11
**現在のカバレッジ**: 80.0%（目標達成済み）

## パッケージ別カバレッジ状況

| パッケージ | カバレッジ | 状態 |
|-----------|-----------|------|
| cmd/duofm | 0.0% | main関数（テスト不要） |
| internal/archive | 80.8% | 良好 |
| internal/config | 74.6% | 改善余地あり |
| internal/fs | 87.9% | 良好 |
| internal/ui | 79.4% | 改善余地あり |

## カバレッジ0%の関数（優先度別）

### 優先度: 低（テスト困難/不要）

以下はテスト追加の必要性が低い関数です：

| 関数 | ファイル | 理由 |
|------|----------|------|
| `main` | cmd/duofm/main.go | エントリーポイント |
| `getDiskSpaceWindows` | internal/fs/diskspace.go | Windowsビルドタグ |
| `openWithXDG` | internal/ui/exec.go | 外部コマンド依存 |
| `openWithCustom` | internal/ui/exec.go | 外部コマンド依存 |
| `handleExecFinished` | internal/ui/model_update.go | TUIメッセージハンドラ |
| `handleShellCommandFinished` | internal/ui/model_update.go | TUIメッセージハンドラ |

### 優先度: 中（テスト可能だが複雑）

| 関数 | ファイル | カバレッジ | 理由 |
|------|----------|-----------|------|
| `GetHistoryPath` | internal/config/config.go | 0.0% | 環境依存 |
| `getConfigDir` | internal/config/config.go | 0.0% | 環境依存 |
| `IsDirectory` | internal/fs/operations.go | 0.0% | 簡単なユーティリティ |
| `SetLoading` | internal/ui/pane.go | 0.0% | 状態設定 |
| `IsLoading` | internal/ui/pane.go | 0.0% | 状態取得 |
| `SetEntries` | internal/ui/pane.go | 0.0% | 状態設定 |
| `GetPaneID` | internal/ui/pane.go | 0.0% | getter |
| `SetPath` | internal/ui/pane.go | 0.0% | setter |
| `GetHistory` | internal/ui/pane.go | 0.0% | getter |
| `GetPendingPath` | internal/ui/pane.go | 0.0% | getter |
| `hasHiddenPrefix` | internal/ui/pane.go | 0.0% | 文字列ユーティリティ |

### 優先度: 高（テスト追加推奨）

以下はテストを追加する価値が高い関数です：

| 関数 | ファイル | カバレッジ | 推奨理由 |
|------|----------|-----------|----------|
| `View` | internal/ui/permission_dialog.go | 0.0% | UI表示の検証に重要 |
| `View` | internal/ui/open_with_dialog.go | 0.0% | UI表示の検証に重要 |
| `save` | internal/ui/bookmark_manager.go | 0.0% | 永続化ロジック |
| `Delete` | internal/ui/bookmark_manager.go | 42.9% | エラーケースの追加 |
| `CursorPos` | internal/ui/archive_name_dialog.go | 0.0% | getter |
| `restoreHistoryOnError` | internal/ui/pane.go | 0.0% | エラー処理 |
| `GetPendingCursorTarget` | internal/ui/pane.go | 0.0% | getter |
| `ClearPendingCursorTarget` | internal/ui/pane.go | 0.0% | 状態操作 |
| `SetCursor` | internal/ui/pane.go | 100.0% | 既テスト済み |
| `SetOnConfirm` | internal/ui/permission_dialog.go | 0.0% | コールバック設定 |
| `renderInputField` | internal/ui/permission_dialog.go | 0.0% | 入力UI |

## 中程度カバレッジの関数（50-70%）

改善の余地がある重要な関数：

| 関数 | ファイル | カバレッジ | 問題点 |
|------|----------|-----------|--------|
| `extract` | internal/archive/archive.go | 53.1% | エラーケース不足 |
| `Compress` | internal/archive/sevenzip_executor.go | 62.1% | エッジケース不足 |
| `Extract` | internal/archive/sevenzip_executor.go | 52.6% | エッジケース不足 |
| `AnalyzeStructure` | internal/archive/smart_extractor.go | 60.0% | 異常系テスト不足 |
| `Compress` | internal/archive/zip_executor.go | 62.1% | エッジケース不足 |
| `Extract` | internal/archive/zip_executor.go | 52.6% | エッジケース不足 |
| `handleDialogMessages` | internal/ui/model_update.go | 60.4% | 分岐網羅不足 |
| `handleContextMenuResult` | internal/ui/model_update.go | 35.9% | メニュー操作の網羅不足 |
| `EnterDirectory` | internal/ui/pane_navigation.go | 60.6% | ナビゲーションエラー |
| `atomicWrite` | internal/ui/shell_history.go | 63.0% | ファイルI/Oエラー |

## テスト追加推奨事項

### 1. Pane関連のgetter/setter（推奨度: ★★★）

**ファイル**: `internal/ui/pane_test.go`

```go
func TestPane_LoadingState(t *testing.T) {
    pane := NewPane(LeftPane, "/tmp")

    // 初期状態
    if pane.IsLoading() {
        t.Error("should not be loading initially")
    }

    // SetLoading
    pane.SetLoading(true, "Loading...")
    if !pane.IsLoading() {
        t.Error("should be loading after SetLoading(true)")
    }
}

func TestPane_Getters(t *testing.T) {
    pane := NewPane(LeftPane, "/tmp")

    if pane.GetPaneID() != LeftPane {
        t.Error("GetPaneID mismatch")
    }

    if pane.GetHistory() == nil {
        t.Error("GetHistory should not return nil")
    }
}
```

**テスト容易性**: 高
**重要度**: 中（getterなので影響小）

### 2. ダイアログView関数（推奨度: ★★☆）

**ファイル**: `internal/ui/permission_dialog_test.go`

```go
func TestPermissionDialog_View(t *testing.T) {
    dialog, _ := NewPermissionDialog("test.txt", 0644, false)
    dialog.SetActive(true)

    view := dialog.View()

    if !strings.Contains(view, "Permissions:") {
        t.Error("View should contain title")
    }

    if !strings.Contains(view, "test.txt") {
        t.Error("View should contain filename")
    }
}
```

**テスト容易性**: 中（出力文字列の検証が必要）
**重要度**: 中

### 3. BookmarkManager.save（推奨度: ★★★）

**ファイル**: `internal/ui/bookmark_manager_test.go`

saveメソッドは永続化を担当する重要な関数です。モックを使用してテスト可能です。

**テスト容易性**: 中（ファイルI/O操作）
**重要度**: 高（データ永続化）

### 4. Archive Executor関数（推奨度: ★★☆）

**ファイル**: 各executor_test.go

Compress/Extract関数のカバレッジ改善には以下のテストケースが有効：

- 空ファイルの圧縮
- 大量ファイルの圧縮
- シンボリックリンクを含むディレクトリ
- 権限エラーのハンドリング
- キャンセル処理

**テスト容易性**: 低（外部コマンド依存）
**重要度**: 高（コア機能）

### 5. Model Update ハンドラ（推奨度: ★☆☆）

**ファイル**: `internal/ui/model_update_test.go`

handleContextMenuResult, handleDialogMessages などのハンドラ関数は分岐が多く、テストが複雑ですが、TUIの正確な動作には重要です。

**テスト容易性**: 低（状態管理が複雑）
**重要度**: 中

## 実装優先順位

1. **即座に追加可能**（getter/setter、シンプルなユーティリティ）
   - `internal/ui/pane.go` の各getter/setter
   - `internal/fs/operations.go` の `IsDirectory`
   - `internal/ui/archive_name_dialog.go` の `CursorPos`

2. **中程度の労力**（ダイアログView関数）
   - `internal/ui/permission_dialog.go` の `View`
   - `internal/ui/open_with_dialog.go` の `View`

3. **労力が必要**（複雑な状態管理）
   - `internal/ui/bookmark_manager.go` の `save`
   - Model Updateハンドラ群

4. **テスト困難/スキップ推奨**
   - 外部コマンド依存関数（`openWithXDG` など）
   - プラットフォーム固有コード（`getDiskSpaceWindows`）

## まとめ

現在80.0%のカバレッジは十分な水準です。追加テストを実装する場合は以下の優先順位を推奨します：

1. **Paneのgetter/setter**: 実装が簡単で即効性あり
2. **ダイアログView関数**: UI表示の検証に重要
3. **BookmarkManager.save**: データ永続化の信頼性向上

カバレッジを85%以上に引き上げるには、複雑なModel Updateハンドラのテスト追加が必要ですが、これは労力対効果を考慮して判断してください。
