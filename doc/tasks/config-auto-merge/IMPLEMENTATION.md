# Implementation Plan: 設定ファイル自動マージ機能

## Overview

duofmの起動時に、設定ファイル(config.toml)に記載のない設定項目をデフォルト値で自動追記する機能を実装する。

## Objectives

- 設定ファイルの不足項目を検出する
- 不足項目をデフォルト値でファイル末尾に追記する
- 既存の設定値を保持しつつ、新しいデフォルト項目を追加する
- ファイル書き込みエラー時も起動を継続する

## Prerequisites

### Development Environment
- Go 1.21+
- BurntSushi/toml パッケージ（既存依存）

### Dependencies
- `internal/config/defaults.go` - デフォルト値定義
- `internal/config/colors.go` - 色設定とキー一覧（GetDefaultColorValue関数を追加）
- `internal/config/config.go` - LoadConfig関数

### Knowledge Requirements
- Go の map 操作
- TOML 形式の文字列生成
- ファイルのアペンドモード書き込み

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Config Format**: TOML (BurntSushi/toml)

### Design Approach
マージロジックを独立したファイル(merger.go)に分離し、LoadConfig()から呼び出す。既存コードへの変更を最小限に抑える。

### Component Interaction

```
LoadConfig(path)
     |
     v
TOML解析 (既存)
     |
     v
MergeConfig(path, rawConfig)  <-- 新規追加
     |
     +---> FindMissingKeybindings()
     +---> FindMissingColors()
     +---> IsMissingHistoryLimit()
     |
     v
generateMergeContent()
     |
     v
ファイル末尾に追記
```

## Implementation Phases

### Phase 1: 不足項目検出関数の実装

**Goal**: 既存設定とデフォルト値を比較し、不足項目を特定する関数群を実装する

**Files to Create**:
- `internal/config/merger.go` - マージロジック

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| FindMissingKeybindings | デフォルトに存在するが設定ファイルに存在しないキーバインディングを検出 | 設定ファイルのkeybindings map | 不足しているkeybindings mapを返す |
| FindMissingColors | デフォルトに存在するが設定ファイルに存在しない色設定を検出 | 設定ファイルのcolors map | 不足している色設定mapを返す |
| IsMissingHistoryLimit | history_limitが設定されているか判定 | *int (nilまたは値) | true: 未設定, false: 設定済み |

**Processing Flow**:

```
FindMissingKeybindings:
1. DefaultKeybindings()を取得
2. 各デフォルトキーについて
   +-- existing に存在する --> スキップ
   +-- existing に存在しない --> 結果mapに追加
3. 結果mapを返す

FindMissingColors:
1. AllColorKeys()を取得
2. 各デフォルトキーについて
   +-- existing に存在する --> スキップ
   +-- existing に存在しない --> GetDefaultColorValue(key)でデフォルト値を取得し結果mapに追加
3. 結果mapを返す
```

**Implementation Steps**:

1. **merger.goファイルの作成**
   - 新規ファイルをinternal/config/に作成
   - パッケージ宣言とインポート

2. **FindMissingKeybindings関数の実装**
   - DefaultKeybindings()との比較ロジック
   - 不足キーのマップ作成

3. **FindMissingColors関数の実装**
   - AllColorKeys()を使用した比較ロジック
   - GetDefaultColorValue(key)でデフォルト値を取得

4. **GetDefaultColorValue関数の追加（colors.go）**
   - キー名からDefaultColors()のフィールド値を返す
   - 存在しないキーには-1を返す

5. **IsMissingHistoryLimit関数の実装**
   - nilチェックによる判定

**Dependencies**:
- Requires: なし（既存のdefaults.go, colors.goを使用）
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:
- FindMissingKeybindings: 空map、一部設定済み、全設定済みの3パターン
- FindMissingColors: 空map、一部設定済み、全設定済みの3パターン
- IsMissingHistoryLimit: nil入力とnon-nil入力の2パターン

**Acceptance Criteria**:
- [ ] FindMissingKeybindingsがデフォルトに存在し設定に存在しないキーのみを返す
- [ ] FindMissingColorsがデフォルトに存在し設定に存在しない色設定のみを返す
- [ ] IsMissingHistoryLimitがnilの場合にtrueを返す

**Estimated Effort**: 小 (1-2 days)

---

### Phase 2: TOML追記内容生成の実装

**Goal**: 不足項目からTOML形式の追記文字列を生成する

**Files to Modify**:
- `internal/config/merger.go` - 追記内容生成関数を追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| mergeResult | 不足項目を保持する内部構造体 | - | Keybindings, Colors, HistoryLimitの不足情報を保持 |
| generateMergeContent | mergeResultからTOML形式文字列を生成 | mergeResultに不足項目が設定済み | TOML形式の文字列を返す（空の場合は空文字列） |

**Processing Flow**:

```
generateMergeContent:
1. 出力バッファを初期化
2. 不足項目が1つでもあるか確認
   +-- なし --> 空文字列を返す
   +-- あり --> 継続
3. 区切りコメントを追加: "# --- Auto-merged settings (added by duofm) ---"
4. keybindingsセクションの生成
   +-- 不足keybindingsがある --> [keybindings] ヘッダーと各項目を追加
   +-- なし --> スキップ
5. colorsセクションの生成
   +-- 不足colorsがある --> [colors] ヘッダーと各項目を追加
   +-- なし --> スキップ
6. history_limitの生成
   +-- HistoryLimitがnilでない --> history_limit = {値} を追加
   +-- nil --> スキップ
7. 生成した文字列を返す
```

**TOML Output Format**:

```
# --- Auto-merged settings (added by duofm) ---

[keybindings]
action_name = ["Key1", "Key2"]

[colors]
color_name = 123

history_limit = 20000
```

**Implementation Steps**:

1. **mergeResult構造体の定義**
   - 不足項目を保持するフィールド

2. **generateMergeContent関数の実装**
   - 各セクションのTOML形式変換
   - 配列値のフォーマット（例: `["Key1", "Key2"]`）
   - 整数値のフォーマット

3. **空セクションスキップロジック**
   - 不足項目がないセクションは出力しない

**Dependencies**:
- Requires: Phase 1
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:
- 全項目不足時: 全セクションが出力される
- 一部項目不足時: 該当セクションのみ出力
- 不足なし: 空文字列が返る

**Acceptance Criteria**:
- [ ] generateMergeContentが正しいTOML形式を生成する
- [ ] 配列値が `["value1", "value2"]` 形式で出力される
- [ ] 整数値が数値として出力される
- [ ] 空のセクションは出力されない

**Estimated Effort**: 小 (1-2 days)

---

### Phase 3: MergeConfig関数とLoadConfig統合

**Goal**: ファイル書き込みとLoadConfigへの統合を完了する

**Files to Modify**:
- `internal/config/merger.go` - MergeConfig関数を追加
- `internal/config/config.go` - LoadConfig内でMergeConfigを呼び出す

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| MergeConfig | 不足項目検出から追記までの全体制御 | 有効なファイルパスとrawConfig | 不足項目がファイル末尾に追記される（エラー時はerrorを返す） |

**Processing Flow**:

```
MergeConfig(path, rawConfig):
1. 不足項目の収集
   +-- FindMissingKeybindings(rawConfig.Keybindings)
   +-- FindMissingColors(rawConfig.Colors)
   +-- IsMissingHistoryLimit(rawConfig.HistoryLimit)
2. generateMergeContent(mergeResult)
3. 追記内容の確認
   +-- 空 --> nil を返す（何もしない）
   +-- 内容あり --> 継続
4. ファイルをアペンドモードで開く
   +-- エラー --> error を返す
5. 既存内容の末尾が改行でない場合、改行を追加
6. 追記内容を書き込む
7. ファイルを閉じる
8. nil を返す

LoadConfig変更:
1. 既存のTOML解析後
2. MergeConfig(path, &raw) を呼び出す
   +-- エラー --> warningsに追加
3. 既存の処理を継続
```

**Implementation Steps**:

1. **MergeConfig関数の実装**
   - 不足項目の収集
   - ファイル末尾への追記
   - エラーハンドリング

2. **ファイル書き込み処理**
   - アペンドモードでファイルを開く
   - 改行の確保
   - 追記内容の書き込み

3. **LoadConfigへの統合**
   - TOML解析後にMergeConfigを呼び出す
   - エラー時はwarningsに追加して継続

**Dependencies**:
- Requires: Phase 2
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- MergeConfig: 一時ファイルを使用した追記テスト
- 不足項目なし: ファイルが変更されないことを確認
- 既存値保持: カスタム値が変更されないことを確認

*Integration Tests*:
- LoadConfig経由でのマージ動作確認

**Acceptance Criteria**:
- [ ] MergeConfigが不足項目をファイル末尾に追記する
- [ ] 不足項目がない場合はファイルを変更しない
- [ ] 既存のカスタム値が保持される
- [ ] ファイル書き込みエラー時もLoadConfigがConfigを返す

**Estimated Effort**: 小 (1-2 days)

---

## Complete File Structure

```
internal/config/
├── config.go        # LoadConfig()の変更（MergeConfig呼び出し追加）
├── defaults.go      # 変更なし
├── colors.go        # GetDefaultColorValue()関数を追加
├── merger.go        # 【新規】マージロジック
│   - FindMissingKeybindings()
│   - FindMissingColors()
│   - IsMissingHistoryLimit()
│   - mergeResult (内部構造体)
│   - generateMergeContent()
│   - MergeConfig()
└── merger_test.go   # 【新規】テスト
```

**File Descriptions**:

| File | Purpose |
|------|---------|
| merger.go | 設定ファイルの不足項目検出とマージロジックを実装 |
| merger_test.go | merger.goの全関数に対するユニットテスト |
| config.go | MergeConfig呼び出しを追加（2行程度の変更） |
| colors.go | GetDefaultColorValue()関数を追加 |

## Testing Strategy

### Unit Testing

**Approach**:
- Go の `testing` パッケージを使用
- テーブル駆動テストで複数シナリオをカバー
- 一時ファイルを使用したファイル操作テスト

**Test Coverage Goals**:
- merger.go: 90%+ coverage

**Key Test Areas**:

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| TC-001: FindMissingKeybindings - 一部設定 | 一部のキーバインディングのみ設定 | 不足キーバインディングのみ返される |
| TC-002: FindMissingColors - 一部設定 | 一部の色設定のみ設定 | 不足色設定のみ返される |
| TC-003: IsMissingHistoryLimit - nil | nil | true |
| TC-003b: IsMissingHistoryLimit - 値あり | *int (値あり) | false |
| TC-004: MergeConfig - 不足あり | 一部項目のみの設定 | ファイル末尾に不足項目追記 |
| TC-005: MergeConfig - 不足なし | 全項目設定済み | ファイル変更なし |
| TC-006: MergeConfig - 既存値保持 | カスタム値設定済み | カスタム値が変更されない |

### Integration Testing

| Test Case | Description |
|-----------|-------------|
| IT-001 | LoadConfig経由でのマージ動作確認 |

### Manual Testing Checklist

- [ ] 空のconfig.tomlで起動 → 全デフォルト値が追記される
- [ ] 一部設定のみのconfig.tomlで起動 → 不足項目のみ追記される
- [ ] 全項目設定済みのconfig.tomlで起動 → ファイル変更なし
- [ ] 読み取り専用ファイルで起動 → 警告が出てアプリは正常起動

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/BurntSushi/toml | existing | TOML解析（既存依存） |

### Internal Dependencies

**Implementation Order**:
1. Phase 1: 検出関数（依存なし）
2. Phase 2: TOML生成（Phase 1に依存）
3. Phase 3: 統合（Phase 2に依存）

## Risk Assessment

### Technical Risks

1. **ファイル書き込み権限**
   - **Risk**: 設定ファイルが読み取り専用の場合
   - **Likelihood**: Low
   - **Impact**: Low（警告のみで起動継続）
   - **Mitigation**: エラーハンドリングで警告を出し、処理継続

2. **TOML形式の互換性**
   - **Risk**: 生成したTOMLが再解析時に問題を起こす
   - **Likelihood**: Low
   - **Impact**: Medium
   - **Mitigation**: テストで追記後の再読み込みを検証

## Performance Considerations

1. **不足項目がない場合のI/O回避**
   - 不足項目検出でI/O操作をスキップ
   - 検出処理は O(n) で完了

2. **起動時間への影響**
   - 追記が必要な場合のみファイル操作
   - 影響は最小限

## Security Considerations

1. **パストラバーサル**
   - LoadConfigから渡されるパスをそのまま使用
   - LoadConfig側で既に検証済み

## Open Questions

### From Specification:
- なし（仕様書で明確に定義済み）

## Future Enhancements

### Not in Current Spec:
- 設定項目のリネーム/削除のマイグレーション機能
- バージョン番号による制御

## Success Metrics

### Functional Completeness
- [ ] 全MVP機能が実装される
- [ ] 全テストシナリオが通過する
- [ ] エラーハンドリングが正しく動作する

### Quality Metrics
- [ ] テストカバレッジ 90%+ (merger.go)
- [ ] 手動テストで致命的なバグなし
- [ ] Go標準のコーディング規約に準拠

### Performance Metrics
- [ ] 起動時間への影響が無視できるレベル

## References

- **Specification**: `doc/tasks/config-auto-merge/SPEC.md`
- **Requirements**: `doc/tasks/config-auto-merge/要件定義書.md`
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test

## Next Steps

1. **Review and Approval**: この実装計画のレビュー
2. **Environment Setup**: 開発環境の確認
3. **Begin Implementation**: Phase 1から実装開始
4. **Continuous Integration**: テストの継続実行
