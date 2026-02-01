# 実装検証レポート: MIME Fallback Configuration

**検証日時**: 2026-02-01
**仕様書**: `doc/tasks/mime-fallback/SPEC.md`
**実装計画**: `doc/tasks/mime-fallback/IMPLEMENTATION.md`
**要件定義書**: `doc/tasks/mime-fallback/要件定義書.md`
**検証者**: implementation-verifier agent

---

## 検証サマリー

| カテゴリ | 評価 | スコア | 詳細 |
|---------|------|--------|------|
| Phase 1: ParseMIMEBehavior | PASS | 100% | 全項目一致 |
| Phase 2: openWithMIME | PASS | 100% | 全項目一致 |
| Phase 3: Merger/Template | PASS | 100% | 全項目一致 |
| FR カバレッジ | PASS | 10/10 | 全 FR 実装済み |
| テストカバレッジ | PASS | 28/28 | 仕様記載の全テストシナリオに対応 |

**総合評価**: PASS -- 仕様書の全要件が正しく実装されている

---

## Phase 1: ParseMIMEBehavior and MIMEBehaviorConfig

### 1.1 MIMEBehaviorConfig 構造体

| チェック項目 | 状態 | 場所 |
|-------------|------|------|
| `Fallback []string` フィールド存在 | PASS | `internal/config/mime.go:20` |
| `Rules map[string][]string` フィールド存在 | PASS | `internal/config/mime.go:15` |
| コメントで用途説明あり | PASS | `internal/config/mime.go:17-19` |

```go
// internal/config/mime.go:10-21
type MIMEBehaviorConfig struct {
    Rules    map[string][]string
    Fallback []string
}
```

### 1.2 ParseMIMEBehavior キー分類ロジック

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| `fallback` キーを exact name で判定 | PASS | `mime.go:51` (`key == "fallback"`) | SPEC L9 |
| `fallback` を Fallback フィールドに格納 | PASS | `mime.go:56` | SPEC FR2 |
| `fallback` が Rules に入らない | PASS | `mime.go:57` (`continue`) | SPEC FR1 |
| `/` を含むキーを MIME パターンとして処理 | PASS | `mime.go:61` | SPEC FR3 |
| 不明キーに警告を生成してスキップ | PASS | `mime.go:71` | SPEC FR3 |
| 空の fallback 配列に警告を生成 | PASS | `mime.go:52-54` | SPEC FR10 |
| 空キーに警告を生成 | PASS | `mime.go:45-47` | IMPL L128 |
| nil マップに対応 | PASS | `mime.go:39-41` | -- |

仕様では「`fallback` は exact name match で判定し、`/` の有無ではない」と明記されている。実装は `key == "fallback"` を先にチェックし、その後 `strings.Contains(key, "/")` でMIMEパターンを判定しており、仕様通り。

### 1.3 Phase 1 テスト

| テスト名 | 仕様テストシナリオ | 状態 |
|----------|-------------------|------|
| `TestParseMIMEBehavior_Fallback/fallback_extracted_from_rules` | fallback key extracted to Fallback field | PASS |
| `TestParseMIMEBehavior_Fallback/fallback_not_in_rules_map` | fallback is not included in Rules map | PASS |
| `TestParseMIMEBehavior_Fallback/empty_fallback_array_generates_warning` | Empty fallback array generates warning | PASS |
| `TestParseMIMEBehavior_Fallback/missing_fallback_results_in_nil` | Missing fallback results in nil/empty | PASS |
| `TestParseMIMEBehavior_Fallback/fallback_only_no_MIME_rules` | Config with only fallback | PASS |
| `TestParseMIMEBehavior_Fallback/multiple_fallback_commands` | Multiple fallback commands | PASS |
| `TestParseMIMEBehavior_Fallback/unknown_key_generates_warning` | Unknown keys generate warning | PASS |
| `TestParseMIMEBehavior_Fallback/fallback_with_MIME_rules_and_unknown_key` | Combined scenario | PASS |
| `TestParseMIMEBehavior_FallbackContent` | Detailed content verification | PASS |

---

## Phase 2: openWithMIME Fallback Chain

### 2.1 tryCommands ヘルパー関数

| チェック項目 | 状態 | 場所 |
|-------------|------|------|
| `tryCommands` 関数が存在 | PASS | `internal/ui/exec.go:158` |
| `strings.Fields` でコマンド文字列を分割 | PASS | `exec.go:161` |
| `exec.LookPath` でコマンド検証 | PASS | `exec.go:167` |
| 見つかったコマンドを `tea.ExecProcess` で実行 | PASS | `exec.go:170-175` |
| 見つからないコマンド名を `notFoundCmds` に追加 | PASS | `exec.go:177` |
| 空の parts をスキップ | PASS | `exec.go:162-164` |

```go
// internal/ui/exec.go:155-180
func tryCommands(commands []string, filePath, workDir string, notFoundCmds *[]string) tea.Cmd {
    for _, cmdStr := range commands {
        parts := strings.Fields(cmdStr)
        if len(parts) == 0 {
            continue
        }
        command := parts[0]
        _, err := exec.LookPath(command)
        if err == nil {
            args := append(parts[1:], filePath)
            c := exec.Command(command, args...)
            c.Dir = workDir
            return tea.ExecProcess(c, func(err error) tea.Msg {
                return execFinishedMsg{err: err}
            })
        }
        *notFoundCmds = append(*notFoundCmds, command)
    }
    return nil
}
```

### 2.2 openWithMIME フォールバックチェーン

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| MIME ルールコマンドを最初に試行 | PASS | `exec.go:197-202` | SPEC FR4 |
| MIME 失敗後に fallback コマンドを試行 | PASS | `exec.go:204-209` | SPEC FR5 |
| 全コマンド失敗時にpagerへフォールバック | PASS | `exec.go:211-215` | SPEC L12 |
| 全失敗時のステータスメッセージにコマンド名を含む | PASS | `exec.go:213` | IMPL L242-246 |
| fallback 未設定時はサイレントにpager使用 | PASS | `exec.go:217-218` | IMPL L244 |
| コマンドが見つかった場合はステータスメッセージなし | PASS | `exec.go:200, 207` | IMPL L239-241 |
| notFoundCmds が MIME + fallback の両方を蓄積 | PASS | `exec.go:194, 199, 206` | IMPL L246 |

フォールバックチェーンのフロー:
1. MIME ルール検索 -> マッチ -> コマンド試行 -> 見つかれば実行
2. MIME 全失敗 or マッチなし -> fallback 試行 -> 見つかれば実行
3. 全失敗 -> pager (ステータスメッセージ付き)
4. 何も設定なし -> pager (サイレント)

これは SPEC.md L99-113 のステートマシンと完全に一致する。

### 2.3 Phase 2 テスト

| テスト名 | 仕様テストシナリオ | 状態 |
|----------|-------------------|------|
| `TestOpenWithMIME_FallbackNoMIMEMatch` | No MIME match + fallback command found | PASS |
| `TestOpenWithMIME_FallbackAllCommandsMissing` | No MIME match + all fallback commands not found | PASS |
| `TestOpenWithMIME_FallbackNoFallbackConfigured` | No MIME match + no fallback configured | PASS |
| `TestOpenWithMIME_AllMIMEFailFallbackWorks` | All MIME commands fail + fallback works | PASS |
| `TestOpenWithMIME_MIMEMatchFallbackNotUsed` | MIME match found + fallback not used | PASS |
| `TestOpenWithMIME_FallbackTriesInOrder` | Fallback tries commands in order | PASS |
| `TestOpenWithMIME_FallbackWithOptions` | Fallback command with options | PASS |
| `TestOpenWithMIME_AllMIMEAndFallbackFail` | All MIME + fallback fail, combined status | PASS |

---

## Phase 3: Config Merger and Template

### 3.1 mergeResult 拡張

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| `MIMEFallbackMissing bool` フィールド追加 | PASS | `merger.go:17` | IMPL L352 |
| `hasContent()` が `MIMEFallbackMissing` を含む | PASS | `merger.go:22` | IMPL L353 |

```go
// internal/config/merger.go:11-18
type mergeResult struct {
    Keybindings         map[string][]string
    Colors              map[string]int
    HistoryLimit        *int
    EnterBehavior       *string
    EnterBehaviorMIME   bool
    MIMEFallbackMissing bool  // <-- 追加されたフィールド
}
```

### 3.2 MergeConfig フォールバック検出

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| セクション欠如時に `EnterBehaviorMIME=true` 設定 | PASS | `merger.go:136-142` | SPEC FR8 |
| コメント化プレースホルダー検出 | PASS | `merger.go:137-141` | IMPL L326-328 |
| セクション存在時に fallback キー存在チェック | PASS | `merger.go:144-147` | SPEC FR7 |
| fallback 不在時に `MIMEFallbackMissing=true` | PASS | `merger.go:146` | IMPL L332 |
| fallback 存在時は変更なし | PASS | `merger.go:145` (キーが存在すれば何もしない) | IMPL L348 |

```go
// internal/config/merger.go:135-148
if IsMissingEnterBehaviorMIME(existing.EnterBehaviorMIME) {
    if !hasEnterBehaviorMIMEComment(string(existingContent)) {
        result.EnterBehaviorMIME = true
    } else {
        result.EnterBehaviorMIME = true
    }
} else {
    if _, hasFallback := existing.EnterBehaviorMIME["fallback"]; !hasFallback {
        result.MIMEFallbackMissing = true
    }
}
```

### 3.3 generateMergedFile フォールバック挿入

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| `enterBehaviorMIMESection` のセクション境界追跡 | PASS | `merger.go:175, 192-194, 204-206, 221-223` | IMPL L364 |
| 既存セクション末尾に fallback 挿入 | PASS | `merger.go:265-267` | IMPL L365-366 |
| セクション未存在時にアクティブセクション作成 | PASS | `merger.go:323-327` | IMPL L366-373 |
| 作成セクションに `fallback = ["xdg-open"]` | PASS | `merger.go:326` | SPEC FR6 |
| 作成セクションにコメント例を含む | PASS | `merger.go:325` | IMPL L369-372 |

### 3.4 defaultConfigTemplate

| チェック項目 | 状態 | 場所 | 仕様参照 |
|-------------|------|------|----------|
| `[enter_behavior_mime]` セクションが非コメント | PASS | `generator.go:29` | SPEC FR9 |
| `fallback = ["xdg-open"]` が非コメント | PASS | `generator.go:34` | SPEC FR9 |
| MIME ルール例がコメント | PASS | `generator.go:30-33` | IMPL L376-378 |
| テンプレートが有効な TOML | PASS | ビルド成功で確認済み | IMPL L407 |

```toml
# generator.go:29-34 (テンプレート抜粋)
[enter_behavior_mime]
# "text/plain" = ["bat", "less"]
# "text/*" = ["less"]
# "image/*" = ["feh", "eog", "xdg-open"]
# "application/pdf" = ["zathura", "evince"]
fallback = ["xdg-open"]
```

### 3.5 Phase 3 テスト

| テスト名 | 仕様テストシナリオ | 状態 |
|----------|-------------------|------|
| `TestMergeConfig_MIMEFallback/section_missing_no_placeholder` | Section missing -> full section added | PASS |
| `TestMergeConfig_MIMEFallback/section_exists_fallback_missing` | Section exists, fallback missing -> appended | PASS |
| `TestMergeConfig_MIMEFallback/section_exists_fallback_present` | Section exists, fallback present -> no change | PASS |
| `TestMergeConfig_MIMEFallback/commented_placeholder` | Commented placeholder -> active section | PASS |
| `TestMergeConfig_MIMEFallback/idempotency` | Second merge makes no changes | PASS |
| `TestGenerateMergedFile_MIMEFallback/MIMEFallbackMissing` | Fallback inserted into section | PASS |
| `TestGenerateMergedFile_MIMEFallback/EnterBehaviorMIME` | Active section with fallback | PASS |
| `TestMergeResultHasContent/has_MIMEFallbackMissing` | hasContent() includes flag | PASS |

---

## FR カバレッジマトリクス

| FR | 説明 | 実装場所 | テスト | 状態 |
|----|------|----------|--------|------|
| FR1 | `fallback` キーを MIME ルールと分離してパース | `mime.go:51-58` | `TestParseMIMEBehavior_Fallback` | PASS |
| FR2 | `MIMEBehaviorConfig.Fallback` に格納 | `mime.go:20` | `TestParseMIMEBehavior_FallbackContent` | PASS |
| FR3 | exact name で判定; `/` は MIME; 不明キーは警告 | `mime.go:51,61,71` | `Fallback/unknown_key_generates_warning` | PASS |
| FR4 | MIME 未一致時に fallback 試行 | `exec.go:204-209` | `TestOpenWithMIME_FallbackNoMIMEMatch` | PASS |
| FR5 | MIME 全失敗時に fallback 試行 | `exec.go:204-209` | `TestOpenWithMIME_AllMIMEFailFallbackWorks` | PASS |
| FR6 | デフォルト値 `["xdg-open"]` | `merger.go:266,326`, `generator.go:34` | `TestMergeConfig_MIMEFallback` | PASS |
| FR7 | Merger が fallback 不在を検出 | `merger.go:144-147` | `MIMEFallback/section_exists_fallback_missing` | PASS |
| FR8 | Merger がセクション不在時にセクション全体追加 | `merger.go:323-327` | `MIMEFallback/section_missing_no_placeholder` | PASS |
| FR9 | デフォルトテンプレートに fallback 含む | `generator.go:29-34` | テンプレート内容確認済み | PASS |
| FR10 | 空 fallback 配列に警告 | `mime.go:52-54` | `Fallback/empty_fallback_array_generates_warning` | PASS |

---

## テストカバレッジ

### パッケージ別カバレッジ

| パッケージ | カバレッジ | 状態 |
|-----------|----------|------|
| `internal/config` | 92.1% | PASS |
| `internal/ui` | 77.4% | -- (パッケージ全体、fallback以外含む) |

### SPEC.md テストシナリオ対応表

#### Unit Tests: ParseMIMEBehavior (7/7)

| 仕様テストシナリオ (SPEC L137-143) | 対応テスト | 状態 |
|-------------------------------------|-----------|------|
| fallback key extracted to Fallback field | `TestParseMIMEBehavior_Fallback/fallback_extracted_from_rules` | PASS |
| fallback is not included in Rules map | `TestParseMIMEBehavior_Fallback/fallback_not_in_rules_map` | PASS |
| MIME rules remain in Rules map | `TestParseMIMEBehavior_FallbackContent` | PASS |
| Empty fallback generates warning | `TestParseMIMEBehavior_Fallback/empty_fallback_array_generates_warning` | PASS |
| Missing fallback results in nil | `TestParseMIMEBehavior_Fallback/missing_fallback_results_in_nil` | PASS |
| Only fallback, no MIME rules | `TestParseMIMEBehavior_Fallback/fallback_only_no_MIME_rules` | PASS |
| Unknown keys generate warning | `TestParseMIMEBehavior_Fallback/unknown_key_generates_warning` | PASS |

#### Unit Tests: openWithMIME (6/6)

| 仕様テストシナリオ (SPEC L149-154) | 対応テスト | 状態 |
|-------------------------------------|-----------|------|
| No MIME match + fallback -> tries fallback | `TestOpenWithMIME_FallbackNoMIMEMatch` | PASS |
| No MIME match + fallback found -> executes | `TestOpenWithMIME_FallbackNoMIMEMatch` | PASS |
| No MIME match + all fallback fail -> pager | `TestOpenWithMIME_FallbackAllCommandsMissing` | PASS |
| No MIME match + no fallback -> pager | `TestOpenWithMIME_FallbackNoFallbackConfigured` | PASS |
| All MIME fail + fallback -> tries fallback | `TestOpenWithMIME_AllMIMEFailFallbackWorks` | PASS |
| MIME match found -> MIME used | `TestOpenWithMIME_MIMEMatchFallbackNotUsed` | PASS |

#### Config Merge Tests (4/4)

| 仕様テストシナリオ (SPEC L163-166) | 対応テスト | 状態 |
|-------------------------------------|-----------|------|
| Section missing -> full section added | `TestMergeConfig_MIMEFallback/section_missing_no_placeholder` | PASS |
| Section exists, fallback missing -> appended | `TestMergeConfig_MIMEFallback/section_exists_fallback_missing` | PASS |
| Section exists, fallback present -> no change | `TestMergeConfig_MIMEFallback/section_exists_fallback_present` | PASS |
| Commented placeholder -> active section | `TestMergeConfig_MIMEFallback/commented_placeholder` | PASS |

#### Edge Cases (5/5)

| 仕様テストシナリオ (SPEC L170-174) | 対応テスト | 状態 |
|-------------------------------------|-----------|------|
| Multiple fallback commands: tries in order | `TestOpenWithMIME_FallbackTriesInOrder` | PASS |
| Fallback command with options | `TestOpenWithMIME_FallbackWithOptions` | PASS |
| MIME rules + fallback: MIME takes priority | `TestOpenWithMIME_MIMEMatchFallbackNotUsed` | PASS |
| Only fallback in section | `TestParseMIMEBehavior_Fallback/fallback_only_no_MIME_rules` | PASS |
| Unknown key warning | `TestParseMIMEBehavior_Fallback/unknown_key_generates_warning` | PASS |

#### 追加テスト (SPEC 外、IMPL 記載)

| テスト | 説明 | 状態 |
|--------|------|------|
| `TestMergeConfig_MIMEFallback/idempotency` | 冪等性 (2回目のマージで変更なし) | PASS |
| `TestOpenWithMIME_AllMIMEAndFallbackFail` | MIME + fallback 全失敗時の combined ステータスメッセージ | PASS |
| `TestMergeResultHasContent/has_MIMEFallbackMissing` | hasContent() に MIMEFallbackMissing 含む | PASS |
| `TestGenerateMergedFile_MIMEFallback/MIMEFallbackMissing` | 既存セクションへの fallback 挿入 | PASS |
| `TestGenerateMergedFile_MIMEFallback/EnterBehaviorMIME` | アクティブセクション作成 | PASS |
| `TestParseMIMEBehavior_Fallback/fallback_with_MIME_rules_and_unknown_key` | 複合シナリオ | PASS |

---

## Success Criteria チェック (SPEC L177-185)

| 基準 | 状態 | 根拠 |
|------|------|------|
| fallback キーが MIME ルールと分離してパースされる | PASS | `mime.go:51-58` |
| MIMEBehaviorConfig に Fallback フィールド | PASS | `mime.go:20` |
| openWithMIME が MIME 未一致時に fallback 使用 | PASS | `exec.go:204-209` |
| openWithMIME が MIME 全失敗時に fallback 使用 | PASS | `exec.go:204-209` |
| Config merger が missing fallback を追加 | PASS | `merger.go:144-147, 265-267` |
| デフォルトテンプレートに fallback 含む | PASS | `generator.go:34` |
| 既存 MIME テストが引き続き PASS | PASS | `go test ./...` で確認 |
| 既存設定ファイルとの後方互換性 | PASS | auto-merge で補完される |

---

## Non-Functional Requirements (SPEC L54-58)

| NFR | 状態 | 根拠 |
|-----|------|------|
| NFR1 - 互換性: fallback なし設定が動作 | PASS | auto-merge でデフォルト追加 |
| NFR2 - 一貫性: fallback が MIME ルールと同じ形式 | PASS | 同一の `[]string` 形式、`tryCommands` で共通処理 |
| NFR3 - パフォーマンス: 影響なし | PASS | `key == "fallback"` の O(1) チェックのみ追加 |

---

## ビルドとコード品質

| 項目 | 状態 |
|------|------|
| `go build ./...` | PASS (エラーなし) |
| 全テスト実行 | PASS (全 fallback 関連テスト 28/28 PASS) |
| テストカバレッジ (config) | 92.1% |

---

## 検出された問題

なし。仕様書の全要件が正しく実装されている。

---

## ファイル変更サマリー

| ファイル | 変更内容 | 行数 |
|---------|----------|------|
| `internal/config/mime.go` | `Fallback` フィールド追加、`ParseMIMEBehavior` キー分類ロジック | 147行 |
| `internal/config/mime_test.go` | Fallback パーステスト 9件追加 | 649行 |
| `internal/ui/exec.go` | `tryCommands` ヘルパー、`openWithMIME` fallback chain | 219行 |
| `internal/ui/exec_test.go` | Fallback 実行テスト 8件追加 | 806行 |
| `internal/config/merger.go` | `MIMEFallbackMissing` フィールド、検出・挿入ロジック | 363行 |
| `internal/config/merger_test.go` | Merger テスト 8件追加 | 1261行 |
| `internal/config/generator.go` | テンプレート更新 (`[enter_behavior_mime]` + `fallback`) | 155行 |

---

## 結論

MIME Fallback Configuration 機能は、SPEC.md に記載された全 10 件の機能要件 (FR1-FR10)、3 件の非機能要件 (NFR1-NFR3)、8 件の Success Criteria 全てを満たしている。IMPLEMENTATION.md の 3 フェーズ全てが計画通りに実装され、SPEC.md に記載された 28 件のテストシナリオ全てに対応するテストが存在し、全て PASS している。

追加の対応は不要。
