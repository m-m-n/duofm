# Implementation Plan: Configuration Hot-Reload

## Architecture Overview

ホットリロード機能は以下の3層で構成する:

1. **Config層** (`internal/config/`): ファイル監視、詳細エラー付き読み込み、設定ファイル修復
2. **UI層** (`internal/ui/`): エラーダイアログ、メッセージハンドラ、設定反映
3. **エントリポイント** (`cmd/duofm/`): ウォッチャー初期化、`program.Send()` 連携

### メッセージフロー

```
[fsnotify] → ConfigWatcher goroutine → program.Send(configFileChangedMsg)
                                              ↓
                                     Model.Update()
                                              ↓
                                   config.LoadConfigDetailed()
                                        ↙           ↘
                                 成功                エラー
                                  ↓                    ↓
                          設定反映+ステータス    ConfigErrorDialog表示
                                                  ↙           ↘
                                          修復+反映      変更前維持/終了
```

## File Structure

### New Files

| File | Purpose |
|------|---------|
| `internal/config/watcher.go` | fsnotifyベースのファイル監視 |
| `internal/config/watcher_test.go` | ウォッチャーのテスト |
| `internal/config/reload.go` | 詳細エラー付き設定読み込み、部分パース |
| `internal/config/reload_test.go` | リロードのテスト |
| `internal/config/repair.go` | 設定ファイルの修復（構文エラー/値エラー） |
| `internal/config/repair_test.go` | 修復のテスト |
| `internal/ui/config_error_dialog.go` | 設定エラーダイアログ |
| `internal/ui/config_error_dialog_test.go` | ダイアログのテスト |
| `internal/ui/model_update_config.go` | 設定リロード関連メッセージハンドラ |

### Modified Files

| File | Changes |
|------|---------|
| `go.mod` | `github.com/fsnotify/fsnotify` 追加 |
| `cmd/duofm/main.go` | ウォッチャー初期化、起動時エラーダイアログ対応 |
| `internal/ui/model.go` | `configPath` フィールド追加、設定更新メソッド |
| `internal/ui/model_update.go` | `handleConfigMessages()` をカスタムメッセージチェーンに追加 |
| `internal/ui/model_update_dialog.go` | ダイアログ終了時に `checkPendingConfigError()` を呼び出す |
| `internal/ui/pane.go` | `SetTheme()` メソッド追加 |
| `internal/ui/shell_history.go` | `SetLimit()` メソッド追加 |

## Implementation Steps

### Step 1: fsnotify 依存の追加

`go get github.com/fsnotify/fsnotify` を実行して依存を追加する。

**Verification:** `go mod tidy` が成功し、`go.mod` に fsnotify が含まれる。

---

### Step 2: 詳細エラー付き設定読み込み (`internal/config/reload.go`)

既存の `LoadConfig()` を拡張するのではなく、新しい `LoadConfigDetailed()` を作成する。既存の `LoadConfig()` はそのまま維持し、後方互換を保つ。

#### ConfigLoadResult 構造体

```go
// ConfigLoadResult は設定読み込みの詳細な結果を保持する。
type ConfigLoadResult struct {
    Config       *Config
    Warnings     []string
    Errors       []ConfigError
    HasSyntaxErr bool
    SyntaxErrLine int  // 構文エラーの行番号（1始まり）
}

// ConfigError は個別の設定エラーを表す。
type ConfigError struct {
    Field   string // エラーのあるフィールド名 (e.g., "history_limit", "colors.cursor_fg")
    Message string // エラーの説明
    Line    int    // エラーの行番号（取得可能な場合）
}

// HasErrors はエラーが存在するかを返す。
func (r *ConfigLoadResult) HasErrors() bool {
    return len(r.Errors) > 0 || r.HasSyntaxErr
}
```

#### LoadConfigDetailed 関数

```go
func LoadConfigDetailed(path string) *ConfigLoadResult
```

処理フロー:
1. ファイルが存在しなければデフォルト設定を返す（エラーなし）
2. TOML パースを試行
3. 構文エラーの場合:
   - `errors.As(err, &toml.ParseError{})` で型アサーションし、`ParseError.Position.Line` から行番号を取得する
   - ダイアログの詳細表示には `ParseError.ErrorWithPosition()` を使用する（位置情報付きのフォーマット済みメッセージを返す）
   - `HasSyntaxErr = true`, `SyntaxErrLine` を設定
   - エラー行より前の内容で `partialParse()` を試行（ベストエフォート）
   - 成功すればその結果を Config に使用、不足分はデフォルト値で補完
   - 失敗すればデフォルト設定全体を使用（partialParse はマルチライン構文の途中カット等で失敗しうるため、このフォールバックが主要な安全策となる）
4. パース成功の場合:
   - 各フィールドのバリデーションを実行
   - 値エラーがあれば `Errors` に追加し、該当フィールドのみデフォルト値に置き換え

#### partialParse 関数

```go
func partialParse(content string, upToLine int) (*rawConfig, error)
```

- ファイル内容を行番号 `upToLine` の手前まで切り出す
- ベストエフォートで TOML パースを試行する（マルチライン文字列・配列の途中でカットされた場合など、有効な TOML にならないケースがある）
- パースに失敗した場合は error を返す（呼び出し元はデフォルト設定全体にフォールバックする）

**テスト:**
- 正常なファイルで `HasErrors() == false`
- 構文エラーファイルで `HasSyntaxErr == true` かつ `SyntaxErrLine` が正しい行番号
- 構文エラー時にエラー行前の設定が正しくパースされる（partialParse 成功ケース）
- 構文エラーでマルチライン途中カットの場合にデフォルト設定全体にフォールバックする（partialParse 失敗ケース）
- 値エラーファイルで `Errors` に正しいフィールド名が含まれる
- 値エラー時に正常なフィールドが保持される
- ファイルが存在しない場合にデフォルト設定が返される

---

### Step 3: 設定ファイルの修復 (`internal/config/repair.go`)

#### RepairConfig 関数

```go
func RepairConfig(path string, result *ConfigLoadResult) error
```

処理フロー:
1. 構文エラーの場合:
   - ファイルを読み込む
   - エラー検知行以降を削除する
   - 残った内容を保持したまま、不足する設定項目をデフォルト値で追記する（既存の `MergeConfig` パターンを流用）
   - ファイルに書き戻す
2. 値エラーの場合:
   - ファイルを読み込む
   - エラーのあるフィールドの行を特定し、デフォルト値で書き換える
   - ファイルに書き戻す
3. 修復後のファイルパーミッションは元のファイルと同じに保つ

#### repairSyntaxError 関数

```go
func repairSyntaxError(content string, errLine int) string
```

- `content` の `errLine` 行以降を削除
- 末尾の不完全な行を除去
- 結果が有効な TOML であることを検証

#### repairValueErrors 関数

```go
func repairValueErrors(content string, errors []ConfigError) string
```

- 各エラーの `Line` を元に該当行を特定
- その行をデフォルト値のフォーマットで置換
- 行番号がない場合はキー名でテキスト検索して置換

**テスト:**
- 構文エラーの修復でエラー行以降が削除される
- 構文エラーの修復で不足設定が追記される
- 修復後のファイルが有効なTOMLである
- 値エラーの修復で不正な値のみが置換される
- 値エラーの修復で他の設定が保持される
- ファイルパーミッションが保持される

---

### Step 4: ファイルウォッチャー (`internal/config/watcher.go`)

#### ConfigWatcher 構造体

```go
// MsgSender はメッセージ送信コールバックの型。
// config パッケージが tea パッケージに依存しないよう interface{} を使用する。
// 送信されるメッセージ型: ConfigFileChangedMsg, ConfigWatchLostMsg
type MsgSender func(msg interface{})

type ConfigWatcher struct {
    configPath    string
    configDir     string
    watcher       *fsnotify.Watcher
    sendMsg       MsgSender
    done          chan struct{}
    suppressUntil time.Time    // この時刻まではイベントを無視する（自己書き込み防止）
    mu            sync.Mutex   // suppressUntil の排他制御
}
```

#### NewConfigWatcher 関数

```go
func NewConfigWatcher(configPath string, sendMsg MsgSender) (*ConfigWatcher, error)
```

- fsnotify.Watcher を作成
- 設定ファイルの親ディレクトリを監視対象に追加（ファイル新規作成の検知のため）
- 設定ファイル自体も監視対象に追加（存在する場合）
- `suppressUntil` をゼロ値で初期化（抑制なし）

#### Start メソッド

```go
func (cw *ConfigWatcher) Start()
```

ゴルーチン内で以下のイベントループを実行:

1. `fsnotify.Event` を受信
2. 設定ファイルに関するイベントのみフィルタリング
3. **抑制チェック**: `suppressUntil` が現在時刻より未来であればイベントを無視する
4. `Write` / `Create` イベント: デバウンスタイマーをリセット（200ms）
5. `Remove` / `Rename` イベント: 1秒後に監視の再登録をリトライ
6. デバウンスタイマー発火後、`sendMsg(ConfigFileChangedMsg{})` を呼ぶ

デバウンス: `time.AfterFunc` を使用。新しいイベントが来たら既存タイマーをリセットする。

#### SuppressFor メソッド

```go
func (cw *ConfigWatcher) SuppressFor(d time.Duration)
```

- `mu` でロックを取得し、`suppressUntil` を `time.Now().Add(d)` に設定する
- RepairConfig によるファイル書き込みの前に呼び出し、自己書き込みによるイベント再発火を防ぐ

#### Stop メソッド

```go
func (cw *ConfigWatcher) Stop()
```

- `done` チャネルを閉じてゴルーチンを停止
- fsnotify.Watcher を Close

#### リトライロジック

```go
func (cw *ConfigWatcher) retryWatch()
```

- 1秒スリープ後にファイル/ディレクトリの監視を再登録
- 失敗時は `sendMsg(ConfigWatchLostMsg{Error: err})` を送信

**テスト:**
- ファイル書き込みで `ConfigFileChangedMsg` が送信される
- ファイル作成で `ConfigFileChangedMsg` が送信される
- ファイル削除後にリトライが実行される
- デバウンスで連続イベントがまとめられる
- `SuppressFor()` 期間中のイベントが無視される
- `SuppressFor()` 期間経過後にイベントが正常に処理される
- `Stop()` でゴルーチンが正常終了する

---

### Step 5: 設定エラーダイアログ (`internal/ui/config_error_dialog.go`)

既存の ConfirmDialog パターンに従い、設定エラー専用のダイアログを作成する。

#### ConfigErrorDialog 構造体

```go
type ConfigErrorDialog struct {
    BaseDialog
    title      string
    errorMsg   string
    details    string
    isStartup  bool   // 起動時かホットリロード時か
    styles     DialogStyles
}
```

#### メッセージ型

```go
// configErrorDialogResultMsg は設定エラーダイアログの結果メッセージ
type configErrorDialogResultMsg struct {
    choice    ConfigErrorChoice
    isStartup bool
}

type ConfigErrorChoice int

const (
    ConfigErrorChoiceFix  ConfigErrorChoice = iota // デフォルト値で修復
    ConfigErrorChoiceQuit                          // アプリ終了（起動時のみ）
    ConfigErrorChoiceKeep                          // 変更前を維持（ホットリロード時のみ）
)
```

#### コンストラクタ

```go
// NewConfigErrorDialog は起動時用の設定エラーダイアログを作成
func NewConfigErrorDialog(errorMsg, details string) *ConfigErrorDialog

// NewConfigErrorDialogForReload はホットリロード時用の設定エラーダイアログを作成
func NewConfigErrorDialogForReload(errorMsg, details string) *ConfigErrorDialog
```

#### Update メソッド

```go
func (d *ConfigErrorDialog) Update(msg tea.Msg) (Dialog, tea.Cmd)
```

- 起動時: `f` で修復、`q` で終了
- ホットリロード時: `f` で修復、`k` で変更前維持
- `Esc` は状況に応じて終了 or 変更前維持

#### View メソッド

起動時:
```
Configuration Error

[エラー内容]

[詳細情報]

[f] Fix with defaults  [q] Quit
```

ホットリロード時:
```
Configuration Error

[エラー内容]

[詳細情報]

[f] Fix with defaults  [k] Keep previous
```

**テスト:**
- 起動時ダイアログで `f` キー → `ConfigErrorChoiceFix` が返る
- 起動時ダイアログで `q` キー → `ConfigErrorChoiceQuit` が返る
- ホットリロード時ダイアログで `f` キー → `ConfigErrorChoiceFix` が返る
- ホットリロード時ダイアログで `k` キー → `ConfigErrorChoiceKeep` が返る
- 非アクティブ時に入力を無視する
- View が正しいレイアウトを返す

---

### Step 6: Model への設定関連フィールド追加 (`internal/ui/model.go`)

#### Model 構造体の拡張

`Model` に以下のフィールドを追加:

```go
// Configuration hot-reload
configPath          string                  // 設定ファイルのパス
configWatcher       *config.ConfigWatcher   // ファイルウォッチャーへの参照（SuppressFor 呼び出し用）
pendingReloadResult *config.ConfigLoadResult // ダイアログ判断待ちのリロード結果
pendingConfigError  *config.ConfigLoadResult // ダイアログ競合時の保留中エラー（既存ダイアログ終了後に表示）
```

#### NewModelWithConfig の拡張

`configPath` パラメータを追加:

```go
func NewModelWithConfig(
    keybindingMap *KeybindingMap,
    theme *Theme,
    warnings []string,
    historyLimit int,
    enterBehavior config.EnterBehavior,
    mimeBehavior config.MIMEBehaviorConfig,
    configPath string,  // 追加
) Model
```

#### 既存テストの更新

以下のテストファイルで `NewModelWithConfig` の呼び出しに `configPath` 引数（空文字列 `""`）を追加する:

- `internal/ui/model_basic_test.go` (2箇所)
- `internal/ui/model_history_navigation_test.go` (1箇所)

#### 設定更新メソッド

```go
// applyConfig は新しい設定をModelに反映する
func (m *Model) applyConfig(cfg *config.Config) {
    m.keybindingMap = NewKeybindingMap(cfg)
    m.theme = NewTheme(cfg.Colors)
    m.enterBehavior = cfg.EnterBehavior
    m.mimeBehavior = cfg.MIMEBehavior

    // ペインのテーマを更新
    if m.leftPane != nil {
        m.leftPane.SetTheme(m.theme)
    }
    if m.rightPane != nil {
        m.rightPane.SetTheme(m.theme)
    }

    // history_limit の更新
    if m.shellHistory != nil {
        m.shellHistory.SetLimit(cfg.HistoryLimit)
    }
}
```

---

### Step 7: Pane のテーマ更新サポート (`internal/ui/pane.go`)

#### SetTheme メソッド

```go
// SetTheme はペインのカラーテーマを更新する
func (p *Pane) SetTheme(theme *Theme) {
    if theme != nil {
        p.theme = theme
    }
}
```

Pane は `theme` フィールドへのポインタ参照で描画しているため、ポインタを差し替えるだけで次の View() から新テーマが適用される。

**テスト:**
- `SetTheme()` でテーマが更新される
- nil を渡しても既存テーマが維持される

---

### Step 7.5: ShellHistory の動的 limit 更新 (`internal/ui/shell_history.go`)

#### SetLimit メソッド

```go
// SetLimit は履歴の上限を動的に変更する。
// 新しい limit が現在のエントリ数より小さい場合、古いエントリを切り捨てる。
// limit が 0 の場合、履歴は無効化される（既存エントリはクリアされない）。
func (sh *ShellHistory) SetLimit(newLimit int) {
    sh.mu.Lock()
    defer sh.mu.Unlock()

    sh.limit = newLimit

    // 現在のエントリが新しい上限を超えていれば切り捨て
    if newLimit > 0 && len(sh.commands) > newLimit {
        sh.commands = sh.commands[:newLimit]
        sh.dirty = true
        // Trigger async save
        select {
        case sh.saveQueue <- struct{}{}:
        default:
        }
    }
}
```

**テスト:**
- `SetLimit()` で limit が更新される
- limit を減らした場合に古いエントリが切り捨てられる
- limit を増やした場合に既存エントリが維持される

---

### Step 8: メッセージ型とハンドラ (`internal/ui/model_update_config.go`)

#### メッセージ型

```go
// configFileChangedMsg はファイル監視からの変更通知
type configFileChangedMsg struct{}

// configWatchLostMsg は監視が外れた通知
type configWatchLostMsg struct {
    err error
}

// configStartupErrorMsg は起動時のエラー通知（Init から送信）
type configStartupErrorMsg struct {
    result *config.ConfigLoadResult
}
```

#### ハンドラ

```go
func (m Model) handleConfigMessages(msg tea.Msg) (Model, tea.Cmd, bool)
```

処理するメッセージ:

1. **`configFileChangedMsg`**:
   - `config.LoadConfigDetailed(m.configPath)` を呼ぶ
   - エラーなし → `m.applyConfig()`, ステータス「Config reloaded」表示
   - エラーあり:
     - `m.dialog != nil`（別のダイアログ表示中）の場合 → `pendingConfigError` に保存し、ダイアログ表示は保留する
     - `m.dialog == nil` の場合 → `pendingReloadResult` に保存、`ConfigErrorDialog` (ホットリロード版) を表示

2. **`configErrorDialogResultMsg`**:
   - `ConfigErrorChoiceFix`:
     - ウォッチャーの `SuppressFor(500ms)` を呼び、修復書き込みによるイベント再発火を抑制する
     - `config.RepairConfig()` でファイルを修復
     - 修復後に再度 `config.LoadConfigDetailed()` で読み込み
     - `m.applyConfig()` で反映
   - `ConfigErrorChoiceKeep`:
     - `pendingReloadResult = nil`（何もしない）
   - `ConfigErrorChoiceQuit`:
     - `tea.Quit` を返す

3. **`configWatchLostMsg`**:
   - ステータスバーに「Config file watch lost. Restart to re-enable.」を表示

4. **`configStartupErrorMsg`**:
   - `pendingReloadResult` に保存
   - `ConfigErrorDialog` (起動時版) を表示

#### 保留中エラーの表示

ダイアログが閉じられた際（`m.dialog` が `nil` になったタイミング）に `pendingConfigError` を確認し、保留中のエラーがあれば `ConfigErrorDialog` を表示する。

```go
// checkPendingConfigError はダイアログ終了後に保留中の設定エラーを表示する。
// ダイアログを閉じる処理の後に呼び出す。
func (m *Model) checkPendingConfigError() tea.Cmd {
    if m.dialog == nil && m.pendingConfigError != nil {
        m.pendingReloadResult = m.pendingConfigError
        m.pendingConfigError = nil
        dialog := NewConfigErrorDialogForReload(/* エラー情報 */)
        m.dialog = dialog
        return dialog.Init()
    }
    return nil
}
```

このメソッドは `handleConfigMessages` 内で `configErrorDialogResultMsg` を処理した後、および既存のダイアログ終了ハンドラ（`model_update_dialog.go` で `m.dialog = nil` を設定する箇所）から呼び出す。

#### handleCustomMessages への統合

`model_update.go` の `handleCustomMessages()` に追加:

```go
// 設定リロード関連メッセージ
if newModel, cmd, handled := m.handleConfigMessages(msg); handled {
    return newModel, cmd, true
}
```

---

### Step 9: main.go の統合 (`cmd/duofm/main.go`)

#### 変更内容

1. **ウォッチャーの初期化**: `tea.NewProgram()` の後、`p.Run()` の前にウォッチャーを作成
2. **`program.Send()` の連携**: ウォッチャーのコールバックに `p.Send` を渡す
3. **起動時エラーダイアログ**: `LoadConfig` → `LoadConfigDetailed` に変更し、エラーがあれば `configStartupErrorMsg` 経由でダイアログ表示
4. **クリーンアップ**: `p.Run()` 終了後にウォッチャーを停止

```go
// 設定読み込みを LoadConfigDetailed に変更
loadResult := config.LoadConfigDetailed(configPath)

// ウォッチャー初期化
watcher, err := config.NewConfigWatcher(configPath, func(msg interface{}) {
    p.Send(msg)
})
if err != nil {
    warnings = append(warnings, fmt.Sprintf("Warning: config watch failed: %v", err))
} else {
    watcher.Start()
    defer watcher.Stop()
}
```

**注意**: `p.Send()` は `p.Run()` が呼ばれた後にのみ有効。ウォッチャーの Start() は Run() の前に呼んでも問題ない（イベントはキューされ、Run() 開始後に処理される）。

起動時エラーの場合は、`Model.Init()` からコマンドで `configStartupErrorMsg` を送信する:

```go
func (m Model) Init() tea.Cmd {
    var cmds []tea.Cmd
    // 起動時設定エラーがあればダイアログ表示
    if m.pendingReloadResult != nil && m.pendingReloadResult.HasErrors() {
        cmds = append(cmds, func() tea.Msg {
            return configStartupErrorMsg{result: m.pendingReloadResult}
        })
    }
    // 既存の警告表示ロジック
    ...
    return tea.Batch(cmds...)
}
```

---

### Step 10: 既存の NFR-1.1 の無効化

既存の設定ファイル仕様 (`doc/tasks/config-file/SPEC.md`) には以下の非機能要件がある:

> NFR-1.1: Configuration file is read only once at startup

この制約はホットリロード機能により無効化される。ホットリロード機能の完了後、既存仕様のこの項目にホットリロード機能への参照を追記する。

---

## Implementation Order

依存関係を考慮した実装順序:

| Order | Step | Depends On | FR Coverage |
|-------|------|-----------|-------------|
| 1 | Step 1: fsnotify 依存追加 | - | - |
| 2 | Step 2: 詳細エラー付き設定読み込み | - | FR-2, FR-3, FR-4 |
| 3 | Step 3: 設定ファイル修復 | Step 2 | FR-7 |
| 4 | Step 4: ファイルウォッチャー | Step 1 | FR-1 |
| 5 | Step 5: 設定エラーダイアログ | - | FR-5, FR-6 |
| 6 | Step 6: Model フィールド追加 | - | FR-2 |
| 7 | Step 7: Pane テーマ更新 | - | FR-2 |
| 8 | Step 8: メッセージハンドラ | Step 2,3,5,6,7 | FR-2,3,4,5,6 |
| 9 | Step 9: main.go 統合 | Step 4,8 | FR-1,5 |
| 10 | Step 10: 既存仕様更新 | Step 9 | - |

## Verification

各ステップ完了後:
- `go build ./...` が成功する
- `go test ./...` が成功する
- `go vet ./...` が警告なし

全ステップ完了後:
- 設定ファイルを編集し、保存後に即座に反映されることを確認
- 構文エラーのある設定ファイルでエラーダイアログが表示されることを確認
- 「デフォルト値で修復」でファイルが修復されることを確認
- 「変更前を維持」で設定が変わらないことを確認
- アプリ起動時に壊れた設定ファイルでエラーダイアログが表示されることを確認
