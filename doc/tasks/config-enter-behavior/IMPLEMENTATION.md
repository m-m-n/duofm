# Implementation Plan: Configurable Enter Key Behavior

## Overview

Enterキーでファイルを開く際の動作を設定ファイルで変更可能にする機能を実装する。`less`（デフォルト）、`xdg-open`、カスタムアプリケーションの3つのモードをサポートし、実行モード（フォアグラウンド/バックグラウンド）は設定に応じて自動決定する。

## Objectives

- Enterキーの動作をVキー（view）から分離する
- 設定ファイルでEnterキーの動作を指定可能にする
- 3つのモードをサポート: `less`（デフォルト）、`xdg-open`、`path:/path/to/app`
- 既存の設定ファイルとの後方互換性を維持する

## Prerequisites

### Development Environment
- Go 1.21+
- make

### Dependencies
- 既存の設定システム（`internal/config/`）
- 既存のexec機能（`internal/ui/exec.go`）
- Bubble Tea フレームワーク

### Knowledge Requirements
- 既存のConfig構造体とLoadConfig処理フロー
- rawConfigとConfig間の変換ロジック
- mergerの動作原理（不足設定項目の自動追加）
- tea.ExecProcess（フォアグラウンド）とバックグラウンド実行の違い

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Config Format**: TOML

### Design Approach

EnterBehaviorを型安全に扱うために専用の型を導入し、設定ファイルのパース時に一度だけ解析を行う。解析済みの値をModelに保持し、handleEnter()では型に応じた分岐のみを行う。

### Component Interaction

```
[config.toml]
     |
     v
LoadConfig() --> ParseEnterBehavior() --> EnterBehavior struct
                                               |
                                               v
                            NewModelWithConfig() --> Model.enterBehavior
                                                          |
                                                          v
                            handleEnter() --> 型に基づいて実行関数を選択
                                    |
                    +---------------+---------------+
                    |               |               |
                    v               v               v
            openWithViewer()  openWithXDG()  openWithCustomForeground()
            (foreground)      (background)    (foreground)
```

## Implementation Phases

### Phase 1: EnterBehavior型とパース機能

**Goal**: EnterBehavior型を定義し、設定値をパースする機能を実装する

**Files to Create**:
- `internal/config/enter.go` - EnterBehavior型定義とパース関数
- `internal/config/enter_test.go` - パース機能のユニットテスト

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| EnterBehaviorType | 動作モードを表す列挙型 | - | Less/XDGOpen/Customの3値を持つ |
| EnterBehavior | パース済みの動作設定を保持 | - | TypeとCustomPathを含む |
| ParseEnterBehavior | 文字列をEnterBehaviorに変換 | 設定値文字列 | EnterBehaviorと警告メッセージ |
| DefaultEnterBehavior | デフォルト値を返す | - | Type=EnterBehaviorLess |

**Processing Flow**:

```
1. 設定値を受け取る
   |
   v
2. strings.TrimSpace()で前後の空白を除去
   |
   v
3. 値の判定
   ├─ 空文字列 → デフォルト(less)を返却 + 警告
   ├─ "less" → EnterBehaviorLess
   ├─ "xdg-open" → EnterBehaviorXDGOpen
   ├─ "path:" プレフィックス → パスを抽出
   │   ├─ パスが空 → デフォルト + 警告
   │   └─ パスあり → EnterBehaviorCustom + CustomPath設定
   │                  ※ パスの存在検証は実行時(openWithCustomForeground)で行う
   └─ その他 → デフォルト + 警告
```

**Implementation Steps**:

1. **EnterBehaviorType列挙型の定義**
   - 3つのモードを表す定数を定義
   - 型安全な分岐を可能にする

2. **EnterBehavior構造体の定義**
   - Typeフィールド（列挙型）
   - CustomPathフィールド（カスタムモード用パス）

3. **ParseEnterBehavior関数の実装**
   - 入力文字列の解析と適切なEnterBehaviorの生成
   - 不正な値の場合は警告とデフォルト値を返却

4. **DefaultEnterBehavior関数の実装**
   - デフォルト動作（less）を返す

5. **String()メソッドの実装**
   - デバッグ・ログ用の文字列表現

**Dependencies**:
- Requires: なし（新規コンポーネント）
- Blocks: Phase 2（Config構造体の更新）

**Testing Approach**:

*Unit Tests*:
- ParseEnterBehavior: 各入力パターン（"less", "xdg-open", "path:/usr/bin/vim", 空文字列, 不正値）
- TrimSpace適用の検証（"  less  " → "less"として処理）
- 警告メッセージの内容検証（strings.Containsで "invalid" や "empty" を確認）
- DefaultEnterBehaviorの返却値検証
- String()メソッドの出力検証

**Acceptance Criteria**:
- [ ] "less"をパースするとEnterBehaviorLessを返す
- [ ] "xdg-open"をパースするとEnterBehaviorXDGOpenを返す
- [ ] "path:/usr/bin/vim"をパースするとEnterBehaviorCustom + CustomPath="/usr/bin/vim"を返す
- [ ] 空文字列・不正値はデフォルトと警告を返す
- [ ] "path:"（パス空）は警告を返す
- [ ] "  less  " (前後空白)はTrimSpace後に正常解析される
- [ ] 全ユニットテストが通過する

**Estimated Effort**: 小 (1-2 days)

---

### Phase 2: Config統合

**Goal**: EnterBehaviorをConfig構造体に組み込み、設定ファイルの読み込み・マージ・生成に対応する

**Files to Modify**:
- `internal/config/config.go`:
  - rawConfigにEnterBehaviorフィールド追加
  - Config構造体にEnterBehaviorフィールド追加
  - LoadConfig()でEnterBehaviorをパースして設定
- `internal/config/defaults.go`:
  - DefaultEnterBehavior()を追加（または enter.go にあれば不要）
- `internal/config/merger.go`:
  - mergeResultにEnterBehaviorフィールド追加
  - IsMissingEnterBehavior()関数追加
  - generateMergedFile()でenter_behavior出力対応
- `internal/config/generator.go`:
  - defaultConfigTemplateにenter_behavior項目追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| rawConfig.EnterBehavior | TOML生値の保持 | TOMLパース成功 | 文字列またはnil |
| Config.EnterBehavior | パース済み動作設定の保持 | LoadConfig完了 | EnterBehavior構造体 |
| IsMissingEnterBehavior | 設定が未設定か判定 | rawConfig | true/false |
| generateMergedFile更新 | enter_behavior出力 | mergeResultにenter_behavior含む | TOML形式で出力 |

**Processing Flow**:

```
1. LoadConfig実行
   ├─ rawConfigにTOML値をデコード
   ├─ enter_behaviorの値を取得
   │   ├─ 値あり → ParseEnterBehavior()で変換
   │   │   ├─ 成功 → Config.EnterBehaviorに設定
   │   │   └─ 警告あり → warningsに追加
   │   └─ 値なし → DefaultEnterBehavior()を使用
   └─ MergeConfig()で不足項目を自動追加
       └─ enter_behaviorが未設定 → ファイルにデフォルト値を追記
```

**Implementation Steps**:

1. **rawConfigの拡張**
   - EnterBehaviorフィールドを追加（文字列ポインタ型）
   - TOMLタグを設定

2. **Config構造体の拡張**
   - EnterBehaviorフィールドを追加

3. **LoadConfig()の更新**
   - rawConfig.EnterBehaviorの値をパースしてConfig.EnterBehaviorに設定
   - 警告メッセージをwarningsに追加

4. **merger機能の拡張**
   - IsMissingEnterBehavior()関数の追加
   - mergeResultにEnterBehaviorフィールド追加
   - generateMergedFile()でenter_behaviorを適切な位置に出力

5. **デフォルト設定テンプレートの更新**
   - generator.goのdefaultConfigTemplateにenter_behavior設定を追加
   - コメントで3つのオプションを説明

**Dependencies**:
- Requires: Phase 1（EnterBehavior型）
- Blocks: Phase 3（UI統合）

**Testing Approach**:

*Unit Tests*:
- LoadConfig: enter_behaviorあり/なしの両パターン
- MergeConfig: enter_behavior不足時の自動追加
- 既存設定の保持（上書きされないこと）

*Integration Tests*:
- 部分的なconfig.tomlの読み込みと再読み込み

**Acceptance Criteria**:
- [ ] enter_behavior設定を含むconfig.tomlを正しくロードできる
- [ ] enter_behavior未設定時はデフォルト(less)が使用される
- [ ] enter_behavior未設定のファイルにMergeConfigでデフォルト値が追記される
- [ ] 既存のenter_behavior値は保持される（上書きされない）
- [ ] defaultConfigTemplateにenter_behavior項目が含まれる
- [ ] 全ユニットテストが通過する

**Estimated Effort**: 中 (3-5 days)

---

### Phase 3: UI統合と動作実装

**Goal**: Model構造体にenterBehaviorを保持し、handleEnter()で設定に応じた動作を実装する

**Files to Modify**:
- `internal/ui/model.go`:
  - Model構造体にenterBehaviorフィールド追加
  - NewModelWithConfig()でenterBehaviorを受け取り設定
- `internal/ui/exec.go`:
  - openWithCustomForeground()関数を追加
- `internal/ui/model_update_keyboard.go`:
  - handleEnter()を修正してenterBehaviorに基づく分岐を実装
- `cmd/duofm/main.go`:
  - Config.EnterBehaviorをNewModelWithConfig()に渡す

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.enterBehavior | 動作設定の保持 | NewModelWithConfig完了 | EnterBehavior構造体 |
| openWithCustomForeground | カスタムアプリでフォアグラウンド実行 | 有効なアプリパスとファイルパス | tea.ExecProcess実行 |
| handleEnter更新 | enterBehaviorに基づく動作選択 | ファイル選択状態 | 適切な実行関数呼び出し |

**Processing Flow**:

```
1. handleEnter()が呼ばれる
2. 選択エントリを取得
   ├─ ディレクトリまたは親ディレクトリ → EnterDirectoryAsync()
   └─ ファイル → enterBehaviorを確認
       ├─ EnterBehaviorLess → openWithViewer() (フォアグラウンド)
       ├─ EnterBehaviorXDGOpen → openWithXDG() (バックグラウンド)
       └─ EnterBehaviorCustom → openWithCustomForeground() (フォアグラウンド)
3. エラー時はステータスメッセージを表示
```

**Implementation Steps**:

1. **Model構造体の拡張**
   - enterBehaviorフィールドを追加
   - フィールドの型はconfig.EnterBehavior

2. **NewModelWithConfig()の更新**
   - 引数にEnterBehaviorを追加
   - フィールドに設定

3. **openWithCustomForeground()の実装**
   - カスタムアプリケーションをフォアグラウンドで実行
   - tea.ExecProcessを使用（openWithViewer/openWithEditorと同様のパターン）
   - **アプリケーションパスの実行時検証**:
     - exec.LookPath(application)を使用
     - 絶対パス: ファイルの存在と実行権限を検証
     - 相対パス/コマンド名: PATH環境変数から検索
     - 見つからない場合: "executable not found" エラーメッセージを返却
     - 権限エラー: "permission denied" エラーメッセージを返却

4. **handleEnter()の修正**
   - ファイル選択時にenterBehaviorのTypeで分岐
   - 各モードに対応する実行関数を呼び出し

5. **main.goの更新**
   - Config.EnterBehaviorをUIに渡す

6. **エラーハンドリング**
   - 実行ファイルが見つからない場合のエラーメッセージ
   - 権限エラーの処理

**Dependencies**:
- Requires: Phase 2（Config統合）
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- openWithCustomForeground: コマンド生成の検証
- handleEnter: 各enterBehaviorモードでの分岐確認

*Manual Testing*:
- [ ] enter_behavior = "less" でファイルを開く（lessで表示）
- [ ] enter_behavior = "xdg-open" でファイルを開く（関連アプリで開く）
- [ ] enter_behavior = "path:/usr/bin/vim" でファイルを開く（vimで開く）
- [ ] ディレクトリでEnterを押すと中に入る（従来動作）
- [ ] Vキーでファイルを開く（lessで表示、従来動作）

**Acceptance Criteria**:
- [ ] enter_behavior = "less" でlessが起動する
- [ ] enter_behavior = "xdg-open" でxdg-openがバックグラウンドで起動する
- [ ] enter_behavior = "path:/path/to/app" でカスタムアプリがフォアグラウンドで起動する
- [ ] 存在しないパスを指定した場合にエラーメッセージが表示される
- [ ] Vキーの動作は変更されない
- [ ] ディレクトリへのEnterは変更されない
- [ ] 全ユニットテストが通過する

**Estimated Effort**: 中 (3-5 days)

---

## Complete File Structure

```
internal/
├── config/
│   ├── config.go           # Modified: rawConfigとConfigにEnterBehavior追加
│   ├── defaults.go         # Modified: DefaultEnterBehavior()追加（enter.goにある場合不要）
│   ├── enter.go            # NEW: EnterBehavior型、ParseEnterBehavior()
│   ├── enter_test.go       # NEW: enter.goのユニットテスト
│   ├── generator.go        # Modified: defaultConfigTemplateにenter_behavior追加
│   ├── merger.go           # Modified: enter_behaviorのマージ対応
│   └── merger_test.go      # Modified: enter_behaviorマージのテスト追加
└── ui/
    ├── model.go            # Modified: enterBehaviorフィールド追加
    ├── exec.go             # Modified: openWithCustomForeground()追加
    ├── exec_test.go        # Modified: openWithCustomForeground()テスト追加
    └── model_update_keyboard.go  # Modified: handleEnter()更新

cmd/
└── duofm/
    └── main.go             # Modified: EnterBehaviorをUIに渡す
```

**File Descriptions**:

| File | Purpose | Changes |
|------|---------|---------|
| enter.go | EnterBehavior型の定義とパース | 新規作成 |
| enter_test.go | パース機能のテスト | 新規作成 |
| config.go | 設定構造体の定義 | EnterBehaviorフィールド追加 |
| merger.go | 不足設定の自動追加 | enter_behaviorマージ対応 |
| generator.go | デフォルト設定生成 | テンプレート更新 |
| model.go | UIモデル定義 | enterBehaviorフィールド追加 |
| exec.go | 外部コマンド実行 | openWithCustomForeground追加 |
| model_update_keyboard.go | キー入力処理 | handleEnter()分岐追加 |
| main.go | エントリポイント | EnterBehaviorの受け渡し |

## Testing Strategy

### Unit Testing

**Approach**:
- Go標準の`testing`パッケージを使用
- テーブル駆動テストで複数シナリオをカバー
- 外部依存（ファイルシステム、外部コマンド）はモックまたは一時ファイルで対応

**Test Coverage Goals**:
- config/enter.go: 90%+ (パース機能は全パターン網羅)
- config/merger.go: 80%+ (既存テストに追加)
- ui/exec.go: 60%+ (コマンド生成部分のみ)

**Key Test Areas**:

1. **ParseEnterBehavior** (`internal/config/enter_test.go`)
   - 有効な入力: "less", "xdg-open", "path:/usr/bin/vim"
   - 無効な入力: 空文字列, "unknown", "path:"
   - 空白文字処理: "  less  " → TrimSpace適用後に正常解析
   - パス解析: スペースを含むパス, 絶対パス
   - 警告メッセージ検証: strings.Contains()で "invalid" / "empty" を確認

2. **Config Loading** (`internal/config/config_test.go`)
   - enter_behaviorあり/なしのconfig.toml読み込み
   - 不正値での警告生成

3. **Config Merging** (`internal/config/merger_test.go`)
   - enter_behavior未設定時の自動追加
   - 既存値の保持

### Integration Testing

**Scenarios**:
1. 設定ファイルなしでの起動 → デフォルト動作
2. 部分設定ファイルでの起動 → 不足項目の自動追加
3. 完全設定ファイルでの起動 → 設定値の適用

### Manual Testing Checklist

**基本動作**:
- [ ] 設定なし: Enterでファイルを開くとlessが起動
- [ ] enter_behavior = "less": lessが起動
- [ ] enter_behavior = "xdg-open": xdg-openが起動（バックグラウンド）
- [ ] enter_behavior = "path:/usr/bin/vim": vimが起動（フォアグラウンド）

**エッジケース**:
- [ ] enter_behavior = "": デフォルト動作 + 警告
- [ ] enter_behavior = "unknown": デフォルト動作 + 警告
- [ ] enter_behavior = "path:": デフォルト動作 + 警告
- [ ] 存在しないパス: エラーメッセージ表示

**リグレッション**:
- [ ] Vキーでファイルを開く: lessで表示（変更なし）
- [ ] Eキーでファイルを開く: エディタで表示（変更なし）
- [ ] Enterでディレクトリに入る: 従来通り動作
- [ ] Enterで親ディレクトリ(..)に移動: 従来通り動作

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/BurntSushi/toml | (既存) | TOML設定ファイル解析 |
| github.com/charmbracelet/bubbletea | (既存) | TUIフレームワーク |

### Internal Dependencies

**Implementation Order**:
1. Phase 1: enter.go（依存なし）
2. Phase 2: config.go, merger.go, generator.go（Phase 1に依存）
3. Phase 3: model.go, exec.go, model_update_keyboard.go（Phase 2に依存）

**Component Dependencies**:
- `config.ParseEnterBehavior` → `config.LoadConfig`で使用
- `config.EnterBehavior` → `ui.Model.enterBehavior`に設定
- `ui.openWithCustomForeground` → `handleEnter()`から呼び出し

## Risk Assessment

### Technical Risks

1. **フォアグラウンド実行の中断処理**
   - **Risk**: カスタムアプリのフォアグラウンド実行中にduofmが応答しなくなる可能性
   - **Likelihood**: 低（既存のopenWithViewer/openWithEditorと同じパターン）
   - **Impact**: 中
   - **Mitigation**: tea.ExecProcessを使用し、既存パターンに従う

2. **パス解析のエッジケース**
   - **Risk**: スペースや特殊文字を含むパスの処理
   - **Likelihood**: 低
   - **Impact**: 低
   - **Mitigation**: 単純な文字列分割を使用し、「path:」以降全体をパスとして扱う

### Implementation Risks

1. **既存設定との互換性**
   - **Risk**: MergeConfigが既存のenter_behaviorを上書きする
   - **Mitigation**: IsMissingEnterBehavior()で未設定の場合のみ追加

## Performance Considerations

1. **設定パース**
   - LoadConfig()内で一度だけパースを実行
   - Model.enterBehaviorにキャッシュ
   - handleEnter()ではenum比較のみ

2. **実行オーバーヘッド**
   - 設定値のenum分岐: O(1)
   - 追加のオーバーヘッドなし

## Security Considerations

1. **パス検証**
   - カスタムパスはユーザー責任（設定ファイルに書いた値をそのまま使用）
   - シェルを経由しないためコマンドインジェクションのリスクは低い

2. **実行権限**
   - 実行時にエラーが発生した場合はエラーメッセージを表示
   - 権限エスカレーションは行わない

## Open Questions

### From Specification:
- [x] `path:`は相対パスをサポートすべきか → No（絶対パスまたはPATH内コマンドのみ）
- [x] `path:`のデフォルト実行モード → フォアグラウンド（lessと同様）

### Implementation-Specific:
- [x] enter_behaviorをルートレベルに配置するか、セクション内に配置するか → ルートレベル（history_limitと同様）

## Future Enhancements

Phase 2以降で検討可能な拡張:
- バックグラウンド/フォアグラウンドの明示的指定オプション
- MIME タイプベースの動作切り替え
- 複数のカスタムアプリ定義

## Success Metrics

### Functional Completeness
- [ ] 3つのモード（less, xdg-open, custom）が動作する
- [ ] 設定ファイルの読み込み・マージが正常に機能する
- [ ] エラーハンドリングが適切に行われる

### Quality Metrics
- [ ] ユニットテストカバレッジ: enter.go 90%+
- [ ] 既存テストの全てがパス
- [ ] golangci-lintでエラーなし

### Backward Compatibility
- [ ] 既存のconfig.tomlが変更なしで動作する
- [ ] enter_behavior未設定時は従来通りlessで開く
- [ ] Vキーの動作は変更なし

## References

- **Specification**: `doc/tasks/config-enter-behavior/SPEC.md`
- **Existing exec implementation**: `internal/ui/exec.go`
- **Config implementation**: `internal/config/config.go`
- **Config merger**: `internal/config/merger.go`
- **Key handler**: `internal/ui/model_update_keyboard.go`

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - 計画内容の確認
   - 不明点の解消

2. **Begin Implementation**
   - Phase 1から開始（enter.go）
   - TDDアプローチ（テストを先に記述）
   - フェーズごとにコミット

3. **Verification**
   - `/sdd.6-verify`で仕様書との整合性を確認
   - 手動テストチェックリストを実行
