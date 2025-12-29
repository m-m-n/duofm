# Implementation Plan: Configuration File (Keybindings)

## Overview

duofmに設定ファイル機能を追加し、ユーザーがキーバインドをカスタマイズできるようにする。TOML形式の設定ファイルを`~/.config/duofm/config.toml`に配置し、アプリケーション起動時に読み込む。

## Objectives

- 設定ファイルパッケージ（internal/config）の新規作成
- TOML形式の設定ファイルの読み込み・パース
- デフォルト設定ファイルの自動生成
- キーバインドの動的マッピング
- エラーハンドリングと警告表示
- ヘルプダイアログのキー表記をPascalCase形式に統一

## Prerequisites

- 既存のキーバインド定義（`internal/ui/keys.go`）
- 既存のModel構造体（`internal/ui/model.go`）
- 既存のステータスバー機能（警告表示用）
- TOMLパーサーライブラリの追加（BurntSushi/toml）

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           main.go                                    │
│  ┌─────────────────┐                                                │
│  │ config.Load()   │──▶ Config struct + warnings                    │
│  └─────────────────┘                                                │
│           │                                                          │
│           ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                         Model                                 │   │
│  │  ┌─────────────────┐    ┌───────────────────────────────┐   │   │
│  │  │ KeybindingMap   │───▶│ key → action lookup          │   │   │
│  │  │ (map[string]    │    │ "j" → ActionMoveDown         │   │   │
│  │  │     Action)     │    │ "ctrl+h" → ActionToggleHidden │   │   │
│  │  └─────────────────┘    └───────────────────────────────┘   │   │
│  │           │                                                   │   │
│  │           ▼                                                   │   │
│  │  ┌─────────────────┐                                         │   │
│  │  │ Update()        │ key press → lookup → execute action     │   │
│  │  └─────────────────┘                                         │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: 設定パッケージの基盤

**Goal**: internal/configパッケージを作成し、基本的な型定義とパス解決を実装

**Files to Create/Modify**:
- `internal/config/config.go` - 新規: Config構造体、LoadConfig関数
- `internal/config/path.go` - 新規: 設定ファイルパス解決
- `go.mod` - 修正: TOMLライブラリ追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Config | 設定全体を保持する構造体 | なし | Keybindingsフィールドを持つ |
| GetConfigPath | XDG準拠のパス解決 | なし | 正しいパスを返す |
| LoadConfig | 設定ファイルの読み込み | パスが解決済み | Config構造体と警告リストを返す |

**Processing Flow**:
```
1. GetConfigPath() → パス決定
2. ファイル存在確認 → 存在しない場合はPhase 2へ
3. TOMLパース → パースエラー時は警告+デフォルト
4. バリデーション → 無効な設定は警告+スキップ
5. Config構造体を返す
```

**Implementation Steps**:

1. **TOMLライブラリの追加**
   - `go get github.com/BurntSushi/toml`

2. **Config構造体の定義** (`config.go`)
   - Keybindingsフィールド: map[string][]string
   - 各アクション名をキー、キー配列を値とする

3. **GetConfigPath関数の実装** (`path.go`)
   - `$XDG_CONFIG_HOME`が設定されている場合はそれを使用
   - 未設定の場合は`~/.config/duofm/config.toml`

4. **LoadConfig関数の骨格実装**
   - ファイル読み込み
   - TOMLパース
   - エラー時はデフォルト設定を返す

**Testing**:
- GetConfigPath: XDG_CONFIG_HOME設定時/未設定時のパス
- LoadConfig: ファイルなし時のデフォルト値

**Estimated Effort**: Small

---

### Phase 2: デフォルト設定ファイルの自動生成

**Goal**: 設定ファイルが存在しない場合にデフォルト設定ファイルを自動生成

**Files to Create/Modify**:
- `internal/config/defaults.go` - 新規: デフォルトキーバインド定義
- `internal/config/generator.go` - 新規: 設定ファイル生成

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| DefaultKeybindings | デフォルトキーバインドのマップ | なし | 28アクションの定義を返す |
| GenerateDefaultConfig | 設定ファイルをディスクに生成 | ディレクトリが作成可能 | config.tomlが生成される |

**Processing Flow**:
```
1. 設定ディレクトリの存在確認
2. ディレクトリが存在しない → os.MkdirAll()で作成
3. デフォルト設定をTOML形式で生成
4. コメント付きでファイルに書き込み
```

**Implementation Steps**:

1. **DefaultKeybindings関数の実装** (`defaults.go`)
   - 28アクション全てのデフォルトキーを定義
   - keys.goの定数と一致させる

2. **GenerateDefaultConfig関数の実装** (`generator.go`)
   - ディレクトリ作成（`os.MkdirAll`）
   - コメント付きTOMLテンプレートの生成
   - セクション分け（Navigation, File operations等）

3. **生成されるファイルのフォーマット**
   - 100行以内
   - 各セクションにコメント
   - 各アクションに説明コメント

**Testing**:
- ディレクトリ作成の確認
- 生成されたファイルのTOML有効性
- 100行以内の確認

**Estimated Effort**: Small

---

### Phase 3: キーバインド設定のパースとバリデーション

**Goal**: TOML設定からキーバインド設定をパースし、バリデーションを行う

**Files to Create/Modify**:
- `internal/config/keybindings.go` - 新規: キーバインド解析
- `internal/config/parser.go` - 新規: キー文字列パーサー
- `internal/config/config_test.go` - 新規: ユニットテスト

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ParseKey | キー文字列の解析 | 有効なキー文字列 | 正規化されたキー文字列 |
| ValidateAction | アクション名の検証 | 文字列 | 有効/無効の判定 |
| ValidateKeybindings | 重複キー検出 | パース済みConfig | 警告リストを返す |

**Processing Flow**:
```
1. TOMLからKeybindingsセクションを抽出
2. 各アクション名を検証 → 無効なら警告してスキップ
3. 各キー文字列をパース → 無効ならデフォルトを使用
4. 重複キー割り当てを検出 → 警告して後勝ち
5. 最終的なキーバインドマップを返す
```

**Implementation Steps**:

1. **ParseKey関数の実装** (`parser.go`)
   - **原則**: 結果の文字で書く（キーボードレイアウト非依存）
   - アルファベット: "J", "N"（大文字）
   - 記号: "?", "@", "!", "~", "/", "-", "="（そのまま）
   - 特殊キー: "Enter", "Esc", "Space", "Tab", "Backspace"（PascalCase）
   - 矢印キー: "Up", "Down", "Left", "Right"
   - ファンクションキー: "F1"〜"F12"
   - 修飾キー: "Ctrl+H", "Ctrl+=", "Shift+N", "Alt+X"
   - 複合修飾キー: "Ctrl+Shift+N", "Alt+Shift+X"

2. **ValidateAction関数の実装** (`keybindings.go`)
   - 有効なアクション名のリストと照合
   - 無効な場合は警告メッセージを生成

3. **ValidateKeybindings関数の実装**
   - 全キーバインドを走査
   - 同一キーが複数アクションに割り当てられているか検出
   - 重複がある場合は警告を生成

4. **BuildKeybindingMap関数の実装**
   - 設定と検証結果からkey→actionのマップを構築
   - 未定義アクションはデフォルト値を使用
   - 無効なキーはスキップ

**Testing**:
- ParseKey: 各フォーマットのテスト
- ValidateAction: 有効/無効アクションのテスト
- ValidateKeybindings: 重複検出のテスト
- 空の配列によるアクション無効化

**Estimated Effort**: Medium

---

### Phase 4: Actionタイプとキーバインドマップの定義

**Goal**: アクションを表す型と、ランタイムで使用するキーバインドマップを定義

**Files to Create/Modify**:
- `internal/ui/actions.go` - 新規: Action型とアクション定数
- `internal/ui/keybinding_map.go` - 新規: KeybindingMap構造体

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Action | アクションを表すenum型 | なし | 28アクションを定義 |
| KeybindingMap | key→actionのマッピング | Config読み込み済み | キー検索が可能 |
| DefaultKeybindingMap | デフォルトマップ生成 | なし | 標準キーバインドのマップ |

**Processing Flow**:
```
1. Config.Keybindingsからアクションごとにキーを取得
2. 各キーに対してkey→actionのエントリを作成
3. KeybindingMapに格納
4. 検索時はキー文字列で引く
```

**Implementation Steps**:

1. **Action型の定義** (`actions.go`)
   - iota使用のenum型
   - ActionNone, ActionMoveDown, ActionMoveUp, ... の28+1アクション
   - String()メソッドで名前を返す

2. **KeybindingMap構造体の定義** (`keybinding_map.go`)
   - 内部にmap[string]Actionを持つ
   - GetAction(key string) Action メソッド
   - HasKey(key string) bool メソッド

3. **NewKeybindingMap関数の実装**
   - Config.Keybindingsを受け取る
   - 各アクションのキーをマップに登録
   - 未定義アクションはデフォルト値を使用

4. **DefaultKeybindingMap関数の実装**
   - keys.goの定数と同等のマップを生成
   - Configなしで動作する後方互換性用

**Testing**:
- Action.String()のテスト
- KeybindingMap.GetActionのテスト
- 未定義キーでActionNoneを返すこと

**Estimated Effort**: Small

---

### Phase 5: Model統合

**Goal**: ModelにKeybindingMapを統合し、キー処理を動的化

**Files to Create/Modify**:
- `internal/ui/model.go` - 修正: keybindingMapフィールド追加、Update関数修正
- `cmd/duofm/main.go` - 修正: 設定読み込み、Model初期化

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.keybindingMap | キーバインドの保持 | Config読み込み済み | キー検索が可能 |
| Model.handleKeyPress | キー入力の処理 | keybindingMap設定済み | アクションを実行 |
| NewModelWithConfig | Config付きModel生成 | Config | 設定済みModel |

**Processing Flow**:
```
1. main.go: config.LoadConfig()を呼び出し
2. main.go: 警告があればModel初期化時に渡す
3. Model.Init(): 警告をステータスバーに表示
4. Model.Update(): キー入力時にkeybindingMapを参照
5. Model.Update(): ActionからハンドラーSwitchへ分岐
```

**Implementation Steps**:

1. **main.goの修正**
   - config.LoadConfig()の呼び出し追加
   - 戻り値のConfigとwarningsを取得
   - NewModelWithConfig(cfg, warnings)を呼び出し

2. **Model構造体にkeybindingMapを追加**
   - keybindingMap *KeybindingMap フィールド
   - configWarnings []string フィールド（初期警告表示用）

3. **NewModelWithConfig関数の実装**
   - 既存のNewModel()を拡張
   - Configからkeybindingマップを生成
   - 警告がある場合はconfigWarningsに保存

4. **Model.Init()の修正**
   - configWarningsがある場合、ステータスメッセージに設定
   - 複数警告がある場合は最初の1つを表示

5. **Model.Update()のキー処理修正**
   - ダイアログ非表示時のキー処理を抽出
   - keybindingMap.GetAction(key)でアクションを取得
   - ActionからSwitchでハンドラーを呼び出し
   - 既存のハードコードキー比較を段階的に置き換え

**Testing**:
- 設定ファイルなしで従来通り動作すること
- カスタムキーバインドが反映されること
- 警告がステータスバーに表示されること

**Estimated Effort**: Large

---

### Phase 6: ヘルプダイアログの表記更新

**Goal**: ヘルプダイアログのキー表記をPascalCase形式に統一

**Files to Create/Modify**:
- `internal/ui/help_dialog.go` - 修正: キー表記をPascalCase形式に更新

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| HelpDialog.View | ヘルプ内容の表示 | なし | PascalCase形式でキーを表示 |

**Implementation Steps**:

1. **help_dialog.goのcontent配列を更新**
   - `j/k/↑/↓` → `J/K/Up/Down`
   - `h/l/←/→` → `H/L/Left/Right`
   - `Enter` → `Enter`（変更なし）
   - `~` → `~`（記号はそのまま）
   - `-` → `-`（変更なし）
   - `q` → `Q`
   - `@` → `@`（記号はそのまま）
   - `c` → `C`
   - `m` → `M`
   - `d` → `D`
   - `!` → `!`（記号はそのまま）
   - `i` → `I`
   - `Ctrl+H` → `Ctrl+H`（変更なし）
   - `?` → `?`（記号はそのまま）

**Testing**:
- ヘルプダイアログの表示確認
- 設定ファイルの表記形式との一致確認

**Estimated Effort**: Small

---

### Phase 7: ユニットテストとE2Eテスト

**Goal**: 包括的なテストの作成

**Files to Create/Modify**:
- `internal/config/config_test.go` - 新規: 設定パッケージのテスト
- `internal/config/parser_test.go` - 新規: パーサーのテスト
- `test/e2e/config_test.go` - 新規: E2Eテスト

**Test Scenarios**:

**Unit Tests (config package)**:
- [ ] LoadConfig: ファイルなし時にデフォルト値を返す
- [ ] LoadConfig: 有効なTOMLを正しくパース
- [ ] LoadConfig: keybindingsセクションなしでもエラーにならない
- [ ] LoadConfig: 空の配列でアクション無効化
- [ ] ParseKey: 単一文字キーの解析（"J", "N"）
- [ ] ParseKey: Ctrl修飾キーの解析（"Ctrl+H"）
- [ ] ParseKey: Shift修飾キーの解析（"Shift+N"）
- [ ] ParseKey: 複合修飾キーの解析（"Ctrl+Shift+N"）
- [ ] ParseKey: ファンクションキーの解析（"F5"）
- [ ] ParseKey: 無効なキー形式でエラー
- [ ] GenerateDefaultConfig: 有効なTOMLを生成
- [ ] GetConfigPath: XDG_CONFIG_HOME設定時のパス
- [ ] GetConfigPath: デフォルトパス
- [ ] ValidateKeybindings: 重複キー検出
- [ ] ValidateKeybindings: 無効アクション名検出

**E2E Tests**:
- [ ] 初回起動で~/.config/duofm/config.tomlが生成される
- [ ] カスタムmove_down = ["Ctrl+N"]が正しく動作
- [ ] 複数キー割り当て（refresh = ["F5", "Ctrl+R"]）が両方動作
- [ ] アクション無効化（help = []）でヘルプが開かない
- [ ] 設定変更後の再起動で反映される

**Estimated Effort**: Medium

---

## File Structure

```
internal/config/           # 新規パッケージ
├── config.go              # Config構造体、LoadConfig
├── defaults.go            # DefaultKeybindings
├── generator.go           # GenerateDefaultConfig
├── keybindings.go         # キーバインド解析、バリデーション
├── parser.go              # ParseKey関数
├── path.go                # GetConfigPath
├── config_test.go         # ユニットテスト
└── parser_test.go         # パーサーテスト

internal/ui/
├── actions.go             # 新規: Action型
├── keybinding_map.go      # 新規: KeybindingMap
├── model.go               # 修正: keybindingMap統合
├── help_dialog.go         # 修正: キー表記をPascalCase形式に更新
├── keys.go                # 既存（参照のみ）
└── ...

cmd/duofm/
└── main.go                # 修正: config.LoadConfig()呼び出し

test/e2e/
└── config_test.go         # 新規: E2Eテスト
```

## Testing Strategy

### Unit Tests

**config.go**:
- LoadConfig: ファイルなし、パースエラー、正常ケース
- 警告メッセージの生成

**parser.go**:
- ParseKey: 全キーフォーマット（12パターン以上）
- 無効キーのエラーハンドリング

**keybindings.go**:
- アクション名バリデーション
- 重複キー検出
- デフォルト値のフォールバック

### Integration Tests

- Model + KeybindingMap の連携
- カスタムキーバインドでのアクション実行
- 警告表示フロー

### Manual Testing Checklist

- [ ] 設定ファイルなしで起動（従来通り動作）
- [ ] 初回起動でconfig.tomlが生成される
- [ ] 生成されたファイルにコメントがある
- [ ] カスタムキーが動作する
- [ ] 無効化したアクションが反応しない
- [ ] パースエラー時に警告が表示される
- [ ] 重複キー時に警告が表示される

## Dependencies

### External Libraries

- `github.com/BurntSushi/toml` - TOMLパース - 最新版

### Internal Dependencies

- Phase 1 → Phase 2（パス解決後にファイル生成）
- Phase 3 → Phase 4（パース結果からマップ生成）
- Phase 4 → Phase 5（マップをModelに統合）

## Risk Assessment

### Technical Risks

- **TOMLパースエラーの網羅性**: 様々なエラーパターンへの対応
  - Mitigation: BurntSushi/tomlのエラーメッセージを活用、行番号表示

- **キー文字列の正規化**: Bubble Teaのキー表現との不一致
  - Mitigation: Bubble Teaのkey.String()出力を調査し、マッピングを確認

- **キーボードレイアウト依存**: JIS/US配列での記号キーの挙動差異
  - Mitigation: 実装前に各配列で`Ctrl+=`等の修飾キー+記号の動作を検証
  - 検証項目: `Ctrl+Shift+-`（JIS配列）と`Ctrl+=`が同一イベントになるか確認

- **後方互換性の維持**: 既存ユーザーへの影響
  - Mitigation: 設定ファイルなしで従来通り動作することを保証

### Implementation Risks

- **Model.Update()の大規模変更**: 既存キー処理の書き換え
  - Mitigation: 段階的に移行、テスト網羅率を上げる

- **起動時間への影響**: 設定ファイル読み込みのオーバーヘッド
  - Mitigation: 100ms以内の読み込みを検証

## Performance Considerations

- 設定ファイルは起動時に1回のみ読み込み
- KeybindingMapはmap[string]Actionで O(1) アクセス
- 100ms以内の読み込み完了を目標

## Security Considerations

- 設定ファイルのパーミッション確認は不要（読み取り専用）
- シェルコマンド実行に関わる設定はキーバインドのみ（既存機能）
- ファイルパス解決で`../`等のトラバーサルを考慮（XDGパス固定のため問題なし）

## Open Questions

- [ ] エラー発生時の複数警告表示: 1つのみか複数か？（仕様では1つずつ表示）

**実装前検証が必要**:
- [ ] JIS/US配列でのBubble Teaキーイベント確認（`Ctrl+Shift+-` vs `Ctrl+=`）

**解決済み**:
- ~~Shift+文字キーの扱い~~ → PascalCase形式で統一。大文字入力は明示的に"Shift+N"と記載
- ~~記号キーの表記~~ → 結果の文字で書く（例: `"?"`）。キーボードレイアウト非依存

## References

- [SPEC.md](./SPEC.md) - 技術仕様書
- [要件定義書.md](./要件定義書.md) - 要件定義
- [BurntSushi/toml](https://github.com/BurntSushi/toml) - TOMLライブラリ
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
