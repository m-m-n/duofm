# Implementation Plan: Symlink Navigation

## Overview

シンボリックリンクの物理パスナビゲーション（"Open link target"）の挙動を修正する。現在の実装はリンク先の親ディレクトリに移動するが、仕様ではリンク先そのものに直接移動すべき。

## Objectives

- "Open link target (physical path)" がリンク先そのものに移動するよう修正
- ドキュメントと実装の整合性を確保
- チェーンされたシンボリックリンクの段階的解決をサポート

## Prerequisites

- Goの開発環境がセットアップ済み
- duofmがビルド可能な状態

## Architecture Overview

本修正は小規模な変更で、既存のアーキテクチャに影響しない。

**変更対象:**
1. `internal/ui/context_menu_dialog.go` - `enter_physical` アクションの実装修正
2. `doc/tasks/context-menu/SPEC.md` - FR3.1の記述修正
3. `doc/tasks/context-menu/要件定義書.md` - FR3.1の記述修正

**変更なし:**
- `internal/ui/pane.go` - Enterキーの論理パスナビゲーションは正常
- `internal/fs/symlink.go` - シンボリックリンク解決ロジックは正常

## Implementation Phases

### Phase 1: コード修正

**Goal**: "Open link target" がリンク先そのものに移動するよう修正

**Files to Modify**:
- `internal/ui/context_menu_dialog.go` - enter_physical アクションのロジック修正

**Implementation Steps**:

1. `context_menu_dialog.go` の `buildMenuItems` 関数内の `enter_physical` アクションを修正

   **現在の実装（137-158行目）:**
   ```go
   items = append(items, MenuItem{
       ID:    "enter_physical",
       Label: "Open link target (physical path)",
       Action: func() error {
           // Navigate to the parent directory of the actual target
           if d.pane != nil {
               targetPath := entry.LinkTarget
               var targetDir string

               if filepath.IsAbs(targetPath) {
                   targetDir = filepath.Dir(targetPath)
               } else {
                   absTarget := filepath.Join(sourcePath, targetPath)
                   targetDir = filepath.Dir(absTarget)
               }

               return d.pane.ChangeDirectory(targetDir)
           }
           return nil
       },
       Enabled: !entry.LinkBroken,
   })
   ```

   **修正後:**
   ```go
   items = append(items, MenuItem{
       ID:    "enter_physical",
       Label: "Open link target (physical path)",
       Action: func() error {
           // Navigate directly to the link target itself
           if d.pane != nil {
               targetPath := entry.LinkTarget

               // Handle relative paths by converting to absolute
               if !filepath.IsAbs(targetPath) {
                   targetPath = filepath.Join(sourcePath, targetPath)
               }

               // Clean the path to resolve any .. components
               targetPath = filepath.Clean(targetPath)

               return d.pane.ChangeDirectory(targetPath)
           }
           return nil
       },
       Enabled: !entry.LinkBroken,
   })
   ```

2. 変更のポイント:
   - `filepath.Dir()` の呼び出しを削除
   - 相対パスの場合は絶対パスに変換
   - `filepath.Clean()` でパスを正規化
   - リンク先そのものに `ChangeDirectory()` を呼び出し

**Dependencies**:
- なし（既存のAPIのみ使用）

**Testing**:
- 手動テスト: `/bin -> /usr/bin` で "Open link target" を実行し、`/usr/bin` に移動することを確認
- 手動テスト: チェーンされたリンクで1段階だけ辿ることを確認

**Estimated Effort**: Small

---

### Phase 2: ドキュメント修正

**Goal**: 仕様書の記述を実装と一致させる

**Files to Modify**:
- `doc/tasks/context-menu/SPEC.md` - FR3.1の記述修正
- `doc/tasks/context-menu/要件定義書.md` - FR3.1の記述修正

**Implementation Steps**:

1. `doc/tasks/context-menu/SPEC.md` を修正

   **現在の記述（50行目付近）:**
   ```markdown
   - "Open link target (physical path)": Open parent directory of the actual file/directory
   ```

   **修正後:**
   ```markdown
   - "Open link target (physical path)": Navigate to the link target directory/file location
   ```

2. `doc/tasks/context-menu/要件定義書.md` を修正

   **現在の記述（54行目付近）:**
   ```markdown
   - **「リンク先を開く（物理パス）」**: リンク先の実体が存在するディレクトリを開く
   ```

   **修正後:**
   ```markdown
   - **「リンク先を開く（物理パス）」**: リンク先のディレクトリ/ファイルに直接移動する
   ```

**Dependencies**:
- Phase 1のコード修正完了後

**Testing**:
- ドキュメントレビュー

**Estimated Effort**: Small

---

### Phase 3: テスト追加

**Goal**: 修正した挙動を検証するユニットテストを追加

**Files to Create/Modify**:
- `internal/ui/context_menu_dialog_test.go` - テストケース追加

**Implementation Steps**:

1. `enter_physical` アクションのテストケースを追加

   ```go
   func TestEnterPhysical_NavigatesToLinkTarget(t *testing.T) {
       // Setup: Create mock pane and symlink entry
       // entry.LinkTarget = "/usr/share"

       // Action: Execute enter_physical action

       // Assert: Pane.ChangeDirectory was called with "/usr/share" (not "/usr")
   }

   func TestEnterPhysical_ChainedSymlink(t *testing.T) {
       // Setup: link1 -> link2 (where link2 is also a symlink)
       // entry.LinkTarget = "/tmp/link2"

       // Action: Execute enter_physical action

       // Assert: Navigates to /tmp/link2 (not the final target)
   }

   func TestEnterPhysical_RelativePath(t *testing.T) {
       // Setup: entry.LinkTarget = "../share" and sourcePath = "/usr/bin"

       // Action: Execute enter_physical action

       // Assert: Navigates to "/usr/share" (resolved absolute path)
   }
   ```

2. 既存のテストが引き続きパスすることを確認

**Dependencies**:
- Phase 1のコード修正完了後

**Testing**:
- `go test ./internal/ui/...` でテスト実行

**Estimated Effort**: Small

---

## File Structure

```
duofm/
├── internal/
│   └── ui/
│       ├── context_menu_dialog.go      # Phase 1: enter_physical 修正
│       └── context_menu_dialog_test.go # Phase 3: テスト追加
└── doc/
    └── tasks/
        ├── context-menu/
        │   ├── SPEC.md                 # Phase 2: FR3.1 修正
        │   └── 要件定義書.md            # Phase 2: FR3.1 修正
        └── symlink-navigation/
            ├── SPEC.md                 # 本仕様書
            ├── 要件定義書.md
            └── IMPLEMENTATION.md       # 本実装計画
```

## Testing Strategy

### Unit Tests

- `context_menu_dialog_test.go` に `enter_physical` アクションのテストを追加
  - リンク先そのものに移動することを確認
  - チェーンされたリンクで1段階のみ辿ることを確認
  - 相対パスが正しく解決されることを確認
  - リンク切れの場合にアクションが無効化されることを確認

### Integration Tests

- なし（既存の統合テストで十分）

### Manual Testing Checklist

- [ ] `/bin -> /usr/bin` で "Open link target" を実行 → `/usr/bin` に移動
- [ ] `/usr/bin` から `..` を選択 → `/usr` に移動（物理的な親）
- [ ] チェーンされたリンク `link1 -> link2 -> target` で "Open link target" → `link2` に移動
- [ ] リンク切れシンボリックリンクで "Open link target" がグレーアウトされている
- [ ] 相対パスのシンボリックリンクで正しいパスに移動

## Dependencies

### External Libraries

- なし（新規追加なし）

### Internal Dependencies

- `internal/ui/pane.go` の `ChangeDirectory()` メソッド（既存）
- `internal/fs/symlink.go` のシンボリックリンク情報取得（既存）

## Risk Assessment

### Technical Risks

- **リスク**: 相対パスの解決が不正確
  - 軽減策: `filepath.Join()` と `filepath.Clean()` を使用して正しく解決

- **リスク**: 既存機能への影響
  - 軽減策: Enterキー（論理パス）の動作には影響なし、enter_physical のみ修正

### Implementation Risks

- **リスク**: テストカバレッジ不足
  - 軽減策: 複数のテストシナリオ（絶対パス、相対パス、チェーン）を追加

## Performance Considerations

- パフォーマンスへの影響なし
- `filepath.Clean()` の追加は無視できるオーバーヘッド

## Security Considerations

- `filepath.Clean()` でパストラバーサルを防止
- 既存のパーミッションチェックを維持

## Open Questions

すべて解決済み:

- [x] チェーンされたリンクを最終ターゲットまで解決すべきか？ → いいえ、1段階のみ
- [x] 相対パスと絶対パスのどちらで表示するか？ → 絶対パス（現行維持）
- [x] ファイルへのシンボリックリンクの動作は？ → 将来実装（本仕様対象外）

## Future Enhancements

- ファイルへのシンボリックリンク: エディタで開く機能（FR6）
- シンボリックリンクのチェーン全体を表示するビュー

## References

- [仕様書] doc/tasks/symlink-navigation/SPEC.md
- [要件定義書] doc/tasks/symlink-navigation/要件定義書.md
- [コンテキストメニュー仕様] doc/tasks/context-menu/SPEC.md
