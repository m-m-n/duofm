# Implementation Plan: Shell Command History

## Overview

duofmのシェルコマンドモードに永続的なコマンド履歴とbash風Ctrl+Rインクリメンタル検索機能を追加する。

## Objectives

- シェルコマンド履歴をファイルに永続化し、セッション間で利用可能にする
- bash風のCtrl+Rインクリメンタル履歴検索を実装する
- 履歴サイズを設定可能にし、重複コマンドを自動除去する
- 既存のシェルコマンド機能との後方互換性を維持する

## Prerequisites

### Development Environment

- Go 1.21+
- make (ビルド自動化用)
- テストフレームワーク: Go標準testing

### Dependencies

- 外部依存なし (標準ライブラリのみ使用)
- 既存コンポーネント: Minibuffer, Config

### Knowledge Requirements

- duofmのMinibuffer実装パターン
- Bubble Teaのメッセージングパターン
- TOMLによる設定管理

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Key Libraries**:
  - 標準ライブラリのみ (os, bufio, strings, path/filepath)

### Design Approach

**責務分離**: ShellHistory構造体がコマンド履歴の管理・永続化・検索を担当し、Modelが状態遷移とUIを担当する。

**状態遷移**: 通常モード -> シェルコマンドモード -> 履歴検索モード の3状態間を遷移する。

### Component Interaction

```
Model                      ShellHistory              Config
  |                            |                       |
  +-- historySearching ------->|                       |
  |   (検索モードフラグ)        |                       |
  |                            |                       |
  +-- Add/Search/Close ------->|                       |
  |                            +-- Load -------> File (起動時)
  |                            |                       |
  |                            +-- saveQueue -----> goroutine --> Atomic Write
  |                            |                       |
  +-- history_limit <----------+-----------------------+
```

## Implementation Phases

### Phase 1: Core History Infrastructure

**Goal**: ShellHistory構造体を実装し、履歴の追加・永続化・読み込みを可能にする

**Files to Create**:
- `internal/ui/shell_history.go` - ShellHistory構造体とメソッド
- `internal/ui/shell_history_test.go` - ユニットテスト

**Files to Modify**:
- `internal/config/config.go` - HistoryLimitフィールド追加
- `internal/config/defaults.go` - デフォルト値定義

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ShellHistory | コマンド履歴の管理 | なし | 履歴操作が可能 |
| ShellHistory.Load | ファイルから履歴を読み込む | ファイルパスが指定されている | 履歴がメモリに読み込まれる |
| ShellHistory.Save | 履歴をファイルに保存する | 履歴が存在する | ファイルが更新される (0600権限) |
| ShellHistory.Add | コマンドを履歴に追加する | コマンドが空でない | 重複が除去され先頭に追加される |
| Config.HistoryLimit | 履歴上限の設定値 | なし | デフォルト20000、0で無効 |

**Processing Flow**:

```
履歴追加フロー:
1. 入力コマンドを受け取る
   +-- 空白のみの場合 -> 何もしない
   +-- 有効なコマンドの場合 -> 続行
2. 既存の重複を検索・削除する
3. コマンドを先頭に追加する
4. 履歴上限を超えた場合、末尾を削除する
5. saveQueueに通知を送信 (非ブロッキング)
6. バックグラウンドgoroutineがdebounce後にAtomic Write (tmp + rename)

終了フロー:
1. Close()が呼ばれる
2. doneチャネルを閉じてgoroutineに終了を通知
3. 保留中の保存があればflush (Atomic Write)
4. goroutineの終了を待機

履歴読み込みフロー:
1. 設定ディレクトリのパスを解決する
2. 履歴ファイルの存在を確認する
   +-- 存在しない場合 -> 空の履歴で開始
   +-- 存在する場合 -> 続行
3. ファイルを行単位で読み込む
4. 空行を除外して履歴に格納する
5. 読み込みエラー時 -> 警告を記録し空の履歴で開始
```

**Implementation Steps**:

1. **Config拡張**
   - HistoryLimitフィールドをConfig構造体に追加する
   - デフォルト値20000を設定する
   - TOMLパースで値を読み込む

2. **ShellHistory構造体**
   - 履歴保持、検索状態、ファイルパスを管理する構造体を定義する
   - saveQueue (chan struct{}) で非同期保存をトリガー
   - done (chan struct{}) でgoroutine終了を通知
   - コンストラクタでlimitとファイルパスを初期化し、バックグラウンドgoroutineを起動

3. **永続化メソッド**
   - Load: ファイル読み込み、limit超過時はトリム、エラーハンドリング
   - atomicWrite (private): tmpファイルに書き込み後rename、権限0600
   - Close: doneチャネルを閉じ、保留保存をflush、goroutine終了待機

4. **履歴操作メソッド**
   - Add: 重複除去、先頭追加、上限管理、saveQueueに通知
   - IsEnabled: limit > 0 を返す

5. **バックグラウンド保存goroutine**
   - saveQueueからの通知を受け取り、500msのdebounce後にatomicWrite
   - doneチャネルで終了を検知

**Dependencies**:
- Requires: なし
- Blocks: Phase 2, Phase 3

**Testing Approach**:

*Unit Tests*:
- Add: 重複除去、上限超過、空コマンド
- Load: 正常読み込み、ファイル不在、破損ファイル、limit超過時トリム
- atomicWrite: tmpファイル作成、rename、権限0600確認、親ディレクトリ作成
- Close: 保留保存のflush、goroutine終了
- Debounce: 複数の連続Addが1回の書き込みにまとめられる
- IsEnabled: limit=0, limit>0

*Integration Tests*:
- なし (Phase 1は独立したコンポーネント)

*Manual Testing*:
- [ ] ファイル権限が0600であることを確認

**Acceptance Criteria**:
- [ ] ShellHistory構造体が定義されている
- [ ] Addで重複コマンドが除去される
- [ ] Loadでファイルが読み込まれる (limit超過時はトリム)
- [ ] Atomic Writeでファイルが0600権限で作成される (tmp + rename)
- [ ] Close()で保留中の保存がflushされる
- [ ] Debounceで連続Addが1回の書き込みにまとめられる
- [ ] Config.HistoryLimitが設定可能 (デフォルト20000)
- [ ] HistoryLimit=0で履歴機能が無効化される

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: ファイル権限がOSによって異なる可能性
  - **Mitigation**: Unix系OSを前提とし、Windows対応は将来課題とする
- **Risk**: 非同期書き込みの複雑さ
  - **Mitigation**: シンプルなgoroutine + channelパターンで実装、Closeで確実にflush

---

### Phase 2: Search Functionality

**Goal**: Ctrl+Rによるインクリメンタル履歴検索を実装する

**Files to Create**:
- `internal/ui/history_searcher.go` - HistorySearcher構造体とメソッド
- `internal/ui/history_searcher_test.go` - HistorySearcherのテスト

**Files to Modify**:
- `internal/ui/shell_history.go` - Commands()メソッド追加
- `internal/ui/model_update_keyboard.go` - Ctrl+Rハンドリング
- `internal/ui/minibuffer.go` - SetInputメソッド追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| HistorySearcher | 検索状態を管理する独立した構造体 | ShellHistoryへの参照 | 検索操作が可能 |
| HistorySearcher.SetPattern | パターンを設定しマッチを検索する | なし | matchesが更新される |
| HistorySearcher.Current | 現在のマッチを返す | パターン設定済み | マッチしたコマンドまたは空文字列 |
| HistorySearcher.Next | 次の一致コマンドを返す | パターン設定済み | 次のマッチまたは空文字列 |
| HistorySearcher.Reset | 検索状態をリセットする | なし | 検索状態がクリア |
| ShellHistory.Commands | 履歴のコピーを返す | なし | []string (検索用) |
| Minibuffer.SetInput | 入力を外部から設定する | なし | 入力が設定される |

**Processing Flow**:

```
Ctrl+R検索フロー:
1. シェルコマンドモードでCtrl+Rが押される
   +-- 履歴無効 (limit=0) -> 何もしない
   +-- 履歴有効 -> 続行
2. 検索モードか判定する
   +-- 非検索モード -> HistorySearcherを作成、検索モードを開始、プロンプト変更
   +-- 検索モード -> HistorySearcher.Next()で次のマッチを検索
3. マッチがあればMinibufferに表示する

文字入力フロー (検索モード中):
1. 入力文字を受け取る
2. HistorySearcher.SetPattern()でパターンを更新
3. HistorySearcher.Current()でマッチを取得
4. マッチがあればMinibufferに表示する

Enter/Escフロー:
Enter -> 選択されたコマンドを実行する、HistorySearcherを破棄
Esc -> 検索をキャンセル、HistorySearcherを破棄、シェルコマンドモードに戻る
```

**Implementation Steps**:

1. **Minibuffer拡張**
   - SetInputメソッドを追加する (外部から入力を設定)
   - カーソル位置を末尾に設定する

2. **HistorySearcher構造体** (新規作成)
   - NewHistorySearcher: ShellHistoryへの参照を受け取り初期化
   - SetPattern: パターン設定、マッチインデックス計算
   - Current: 現在のマッチを返す
   - Next: 次のマッチに移動して返す
   - Reset: 状態リセット

3. **ShellHistory拡張**
   - Commands: 履歴のコピーを返す (検索用)

4. **キーボードハンドリング**
   - handleShellCommandInput内でCtrl+Rを検出する
   - 検索モード開始時にHistorySearcherを作成
   - 検索モードフラグ(historySearching)を管理する
   - 検索モード中の文字入力でSetPatternを呼び出す
   - 終了時にHistorySearcherをnilに設定

**Dependencies**:
- Requires: Phase 1
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:
- HistorySearcher.SetPattern: 大文字小文字無視、部分一致、マッチなし
- HistorySearcher.Current: 現在のマッチ取得
- HistorySearcher.Next: 複数マッチ、マッチ終端
- HistorySearcher.Reset: 状態クリア
- ShellHistory.Commands: 履歴のコピーが返される

*Integration Tests*:
- Ctrl+R押下で検索モード開始、HistorySearcher作成
- 文字入力でフィルタリング
- 再度Ctrl+Rで次のマッチ
- Escで検索終了、HistorySearcher破棄

*Manual Testing*:
- [ ] プロンプトが "(bck-i-search): " に変わる
- [ ] 検索結果がMinibufferに表示される

**Acceptance Criteria**:
- [ ] HistorySearcher構造体が定義されている
- [ ] Ctrl+Rで検索モードが開始しHistorySearcherが作成される
- [ ] 入力文字で履歴がフィルタされる
- [ ] 再度Ctrl+Rで次のマッチに移動する
- [ ] 大文字小文字を無視して検索する
- [ ] Enterでコマンドが実行される
- [ ] Escでシェルコマンドモードに戻りHistorySearcherが破棄される

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:
- **Risk**: Minibufferの状態管理が複雑になる
  - **Mitigation**: 検索モードフラグをModelで明確に管理し、HistorySearcherを分離することで複雑さを軽減

---

### Phase 3: Integration and Polish

**Goal**: アプリケーション起動時の履歴読み込みとコマンド実行後の保存を統合する

**Files to Create**:
- なし

**Files to Modify**:
- `internal/ui/model.go` - shellHistoryフィールド追加、初期化
- `internal/ui/model_update.go` - コマンド実行後の履歴保存
- `internal/ui/help_dialog.go` - ヘルプにCtrl+Rを追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.shellHistory | 履歴インスタンスを保持する | なし | Model生成時に初期化 |
| Model.historySearching | 検索モードフラグ | なし | 状態遷移で更新 |
| NewModelWithConfig | 履歴を読み込む | Config.HistoryLimitが設定済み | 履歴がロードされる |
| handleShellCommandInput | コマンド実行時に履歴保存 | コマンドが入力済み | 履歴に追加・保存 |

**Processing Flow**:

```
アプリケーション起動フロー:
1. Configを読み込む
2. HistoryLimitを取得する
   +-- limit=0 -> ShellHistoryをnil/無効状態で初期化
   +-- limit>0 -> ShellHistoryを作成し履歴をロード
3. ロードエラー時 -> 警告を記録、空の履歴で開始
4. Modelにshellhistoryを設定する

コマンド実行フロー:
1. Enterが押される
2. コマンドを取得する
3. シェルコマンドを実行する
4. 履歴にコマンドを追加する (Add)
   -> 内部で非同期保存がトリガーされる (debounce)
5. 保存エラー -> ログに記録、ユーザーには通知しない

アプリケーション終了フロー:
1. 終了シグナルを受信する
2. shellHistory.Close()を呼び出す
3. 保留中の保存がflushされる
4. goroutineが終了する
```

**Implementation Steps**:

1. **Model拡張**
   - shellHistoryフィールドを追加する
   - historySearchingフラグを追加する
   - NewModelWithConfigで履歴を初期化・ロードする

2. **履歴統合**
   - handleShellCommandInputでコマンド実行時に履歴追加 (Add)
   - 非同期保存は内部で自動的にトリガーされる
   - 検索モードからの実行も同様に処理する

3. **終了処理**
   - Model終了時にshellHistory.Close()を呼び出す
   - 保留中の保存を確実にflush

4. **エラーハンドリング**
   - ロードエラー: 警告をconfigWarningsに追加
   - 保存エラー: ログに記録、ユーザー操作を妨げない

5. **ヘルプダイアログ更新**
   - Ctrl+Rキーバインドを追加する

**Dependencies**:
- Requires: Phase 1, Phase 2
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
- Model初期化時に履歴がロードされる
- コマンド実行後に履歴が更新される

*Integration Tests*:
- コマンド実行 -> 再起動 -> 履歴が保持されている
- 重複コマンド実行 -> 履歴に1つだけ存在

*Manual Testing*:
- [ ] duofm起動時に履歴ファイルが読み込まれる
- [ ] コマンド実行後に履歴が保存される
- [ ] 再起動後に履歴が利用可能
- [ ] ヘルプダイアログにCtrl+Rが表示される

**Acceptance Criteria**:
- [ ] 起動時に履歴が自動的に読み込まれる
- [ ] コマンド実行後に履歴が非同期で保存される
- [ ] アプリ終了時にClose()で保留保存がflushされる
- [ ] 履歴読み込みエラー時もアプリが正常起動する
- [ ] ヘルプダイアログにCtrl+Rが記載されている
- [ ] 既存のシェルコマンド機能が正常に動作する

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: 既存のシェルコマンド機能を破壊する可能性
  - **Mitigation**: 既存テストを全て通過することを確認する
- **Risk**: 終了時のClose()呼び出し忘れ
  - **Mitigation**: defer shellHistory.Close() パターンを使用

---

## Complete File Structure

```
internal/
+-- config/
|   +-- config.go           # HistoryLimitフィールド追加
|   +-- defaults.go         # デフォルト値20000
+-- ui/
    +-- shell_history.go    # ShellHistory構造体とメソッド
    +-- shell_history_test.go # ShellHistoryのユニットテスト
    +-- history_searcher.go # HistorySearcher構造体とメソッド
    +-- history_searcher_test.go # HistorySearcherのユニットテスト
    +-- model.go            # shellHistory, historySearcher, historySearchingフィールド
    +-- model_update_keyboard.go # Ctrl+Rハンドリング修正
    +-- minibuffer.go       # SetInputメソッド追加
    +-- help_dialog.go      # Ctrl+Rキーバインド追加

~/.config/duofm/
+-- history                 # 履歴ファイル (0600権限)
+-- config.toml             # history_limit設定
```

**File Descriptions**:
- `shell_history.go`: ShellHistory構造体、Add/Load/Close/Commandsメソッド、バックグラウンドgoroutine
- `history_searcher.go`: HistorySearcher構造体、SetPattern/Current/Next/Resetメソッド
- `shell_history_test.go`: 全メソッドのユニットテスト
- `config.go`: HistoryLimitフィールドとTOMLパース
- `model.go`: shellHistoryインスタンス保持、初期化
- `model_update_keyboard.go`: Ctrl+Rと検索モードの入力処理
- `minibuffer.go`: SetInputメソッド (検索結果の表示用)
- `help_dialog.go`: ヘルプにCtrl+Rを追加

## Testing Strategy

### Unit Testing

**Approach**:
- Go標準testingパッケージを使用
- テーブル駆動テストで複数シナリオをカバー
- 一時ディレクトリでファイル操作をテスト

**Test Coverage Goals**:
- ShellHistory: 90%+ (コアロジック)
- Config拡張: 80%+
- キーボードハンドリング: 70%+

**Key Test Areas**:

1. **ShellHistory** (`internal/ui/shell_history_test.go`)
   - Add: 重複除去、上限超過、空コマンド、トリム
   - Load: 正常、ファイル不在、破損、Unicode
   - Save: ファイル作成、権限、親ディレクトリ作成
   - Search: 大文字小文字無視、部分一致、マッチなし
   - SearchNext: 複数マッチ、終端、パターンなし

2. **Config** (`internal/config/config_test.go`)
   - HistoryLimit読み込み: 明示指定、デフォルト、0

3. **Keyboard** (`internal/ui/model_update_keyboard_test.go`)
   - Ctrl+R: 検索モード開始、次マッチ、無効時
   - Enter: コマンド実行、履歴追加
   - Esc: 検索キャンセル

### Integration Testing

**Scenarios**:
1. シェルコマンド実行 -> 履歴追加 -> 再起動 -> 履歴保持
2. Ctrl+R -> 文字入力 -> フィルタ -> Ctrl+R -> 次マッチ -> Enter -> 実行
3. history_limit=0 -> Ctrl+Rが無視される

**Approach**:
- 一時ディレクトリで統合テスト
- キー入力シミュレーション

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Ctrl+Rで検索モード開始 (プロンプト変化)
- [ ] 文字入力でリアルタイムフィルタリング
- [ ] 再度Ctrl+Rで次のマッチ
- [ ] Enterでコマンド実行
- [ ] Escでキャンセル
- [ ] 再起動後も履歴が保持されている
- [ ] 同一コマンド2回実行で1エントリのみ
- [ ] 長いコマンド (>1000文字) が正常に処理される

## Dependencies

### External Dependencies

なし (Go標準ライブラリのみ)

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: Config拡張 -> ShellHistory基本実装
2. Phase 2: 検索メソッド -> Minibuffer拡張 -> キーボードハンドリング
3. Phase 3: Model統合 -> ヘルプ更新

**Component Dependencies**:
- `shell_history.go` は標準ライブラリのみに依存
- `model.go` は `shell_history.go` に依存
- `model_update_keyboard.go` は `shell_history.go` と `minibuffer.go` に依存

## Risk Assessment

### Technical Risks

1. **Minibuffer状態管理の複雑化**
   - **Risk**: 検索モードと通常入力モードの状態管理
   - **Likelihood**: Medium
   - **Impact**: Medium
   - **Mitigation**: historySearchingフラグをModelで明確に管理し、状態遷移を明確にする

2. **ファイル権限の互換性**
   - **Risk**: Windowsでのファイル権限処理
   - **Likelihood**: Low (Linux/macOSがターゲット)
   - **Impact**: Low
   - **Mitigation**: Unix系OSを前提とし、Windowsでは権限設定をスキップ

3. **履歴ファイル破損**
   - **Risk**: 書き込み中の異常終了で破損
   - **Likelihood**: Very Low (Atomic Writeで軽減)
   - **Impact**: Low
   - **Mitigation**: Atomic Write (tmp + rename) で書き込み中のクラッシュでも既存ファイルを保護。破損ファイルでも起動可能 (空履歴でフォールバック)

### Implementation Risks

1. **既存機能への影響**
   - **Risk**: シェルコマンドモードの動作変更
   - **Mitigation**: 既存テストを全てパスすることを確認

2. **パフォーマンス劣化**
   - **Risk**: 大量履歴での検索遅延
   - **Mitigation**: 20000エントリ100msの性能要件を満たすことを確認

## Performance Considerations

1. **検索性能**
   - 線形検索で十分 (20000エントリで100ms以内)
   - 大文字小文字無視のため文字列を事前に小文字化してキャッシュ可能 (将来最適化)

2. **ファイルI/O**
   - 非同期Atomic Writeで500ms debounce
   - 連続コマンド実行でも書き込み回数を抑制
   - 大量履歴でも500ms以内でロード可能

3. **メモリ使用量**
   - 20000コマンド x 平均50文字 = 約1MB
   - 問題にならないレベル

## Security Considerations

1. **ファイル権限**
   - 履歴ファイルは0600で作成 (所有者のみ読み書き可能)
   - 設定ディレクトリが存在しない場合は0700で作成

2. **機密データ**
   - ユーザーがパスワード等を入力しないよう注意喚起 (ドキュメント記載)
   - アプリケーション側での検出・除外は行わない

3. **パス固定**
   - 履歴ファイルパスは `~/.config/duofm/history` に固定
   - 設定による変更不可 (パストラバーサル防止)

## Open Questions

### From Specification:
- なし (仕様書で明確に定義されている)

### Implementation-Specific:
- [ ] Windowsでの0600権限処理をどうするか (現時点ではスキップ)
- [ ] 同時起動時の履歴ファイル競合をどう扱うか (現時点では後勝ち)

### To Clarify with User:
- なし

## Future Enhancements

Items deferred to later phases or releases:

### Not in Current Spec:
- 上下矢印での履歴ナビゲーション (bash風)
- 履歴のエクスポート/インポート
- 正規表現による検索
- 履歴クリアコマンド

## Success Metrics

### Functional Completeness
- [ ] 全てのFunctional Requirements (FR1-FR12) が実装されている
- [ ] 全てのUser Stories (US1-US4) が満たされている
- [ ] 全てのテストシナリオがパスする

### Quality Metrics
- [ ] テストカバレッジ80%以上 (ShellHistoryは90%以上)
- [ ] 手動テストで重大なバグなし
- [ ] Goベストプラクティスに準拠

### Performance Metrics
- [ ] 20000エントリの検索が100ms以内
- [ ] 20000エントリの読み込みが500ms以内
- [ ] 20000エントリのAtomic Writeが100ms以内

### User Experience
- [ ] bash風の直感的なCtrl+R操作
- [ ] 明確なエラーメッセージ (ステータスバー)
- [ ] ヘルプダイアログで操作方法を確認可能

## References

- **Specification**: `doc/tasks/shell-command-history/SPEC.md`
- **Existing Implementation**:
  - `internal/ui/exec.go` - シェルコマンド実行
  - `internal/ui/minibuffer.go` - ミニバッファ実装
  - `internal/config/config.go` - 設定管理
- **bash Ctrl+R**: https://www.gnu.org/software/bash/manual/html_node/Commands-For-History.html
- **XDG Base Directory**: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - `/sdd.3-verify-plan` で整合性検証と設計レビューを実行
   - 不明点を確認・解決

2. **Environment Setup**
   - go mod download
   - make test で既存テストが通ることを確認

3. **Begin Implementation**
   - Phase 1から開始
   - TDDアプローチ (テスト先行)
   - 各フェーズ完了後にコミット

4. **Verification**
   - `/sdd.5-check` で実装が計画に準拠しているか確認
   - `/sdd.6-verify` で仕様書の要件が満たされているか確認
