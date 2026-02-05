# Implementation Plan: Clipboard Copy

## Overview

コンテキストメニューに「Copy file name」と「Copy full path」の2項目を追加し、カーソル位置のファイル名または絶対パスをOSC 52 + 外部コマンドフォールバックでシステムクリップボードに書き込む。

## Objectives

- コンテキストメニューに `copy_name` と `copy_path` の2つのメニュー項目を追加する
- OSC 52エスケープシーケンスおよび外部コマンドフォールバックによるクリップボード書き込みを実装する
- 成功・失敗時のステータスバーフィードバックを実装する

## Prerequisites

- コンテキストメニュー機能が実装済み（`internal/ui/context_menu_dialog.go`）
- コンテキストメニュー結果ハンドリングが実装済み（`internal/ui/model_update.go` の `handleContextMenuResult`）
- ステータスバーメッセージ表示機能が実装済み（`statusMessage` + `statusMessageClearCmd`）

## Architecture Overview

### Component Interaction

```
ContextMenuDialog                Model                         clipboard module
  │                                │                              │
  │ buildMenuItems()               │                              │
  │  ├─ copy_name (new)            │                              │
  │  └─ copy_path (new)            │                              │
  │                                │                              │
  │ contextMenuResultMsg           │                              │
  │ ───────────────────────────>   │                              │
  │   actionID: "copy_name"        │  WriteToClipboard(text)      │
  │                                │ ───────────────────────────> │
  │                                │                              │
  │                                │  1. OSC 52 (/dev/tty)        │
  │                                │  2. External cmd (fallback)  │
  │                                │ <─────────────────────────── │
  │                                │                              │
  │                                │ statusMessage = "Copied: .." │
  │                                │ statusMessageClearCmd(3s)     │
```

### Fallback Strategy

```
1. OSC 52 escape sequence emit (always attempted)
   ├─ Open /dev/tty and write sequence (Bubble Tea互換)
   ├─ /dev/tty open failure is non-fatal (best-effort)
   └─ No reliable way to detect terminal support
2. External command detection and execution (timeout: 5s)
   ├─ wl-copy found → pipe text to stdin
   ├─ xclip found → pipe text to stdin with -selection clipboard
   └─ xsel found → pipe text to stdin with --clipboard --input
3. Result determination
   ├─ External command succeeded → success
   ├─ External command failed → error
   ├─ No external command found, OSC 52 attempted → success (best effort)
   └─ /dev/tty open failed AND no external command → error
```

## Implementation Phases

### Phase 1: Clipboard Module

**Goal**: OSC 52およびフォールバック外部コマンドによるクリップボード書き込み機能を独立モジュールとして作成する。

**Files to Create**:
- `internal/clipboard/clipboard.go` - クリップボード書き込み機能
- `internal/clipboard/clipboard_test.go` - ユニットテスト

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| WriteToClipboard | テキストをクリップボードに書き込む | 有効なテキスト文字列 | OSC 52が/dev/tty経由で出力され、外部コマンドが利用可能なら実行される |
| buildOSC52Sequence | OSC 52エスケープシーケンスを生成する | 任意の文字列 | base64エンコードされたOSC 52シーケンスが返される |
| writeOSC52 | OSC 52シーケンスを/dev/ttyに書き込む | 任意の文字列 | /dev/ttyへの書き込み結果（成功またはエラー）が返される |
| findClipboardCommand | 利用可能な外部クリップボードコマンドを検出する | なし | コマンド名と引数、またはコマンドなしが返される |
| execClipboardCommand | 外部コマンドにテキストをパイプで渡して実行する（5秒タイムアウト） | 有効なコマンドとテキスト | コマンドの実行結果（成功またはエラー）が返される |

**Processing Flow**:
```
1. /dev/ttyを開いてOSC 52エスケープシーケンスを書き込む（best-effort）
   ├─ /dev/ttyオープン成功 → osc52Attempted = true（書き込み結果に関わらず）
   │   ├─ 書き込み成功（best-effort、エラーは無視）
   │   └─ 書き込み失敗（ログ出力のみ、osc52Attemptedはtrue維持）
   └─ /dev/ttyオープン失敗 → osc52Attempted = false
2. 外部コマンドを検出
   ├─ コマンドあり → context.WithTimeout(5s)でstdinにテキストをパイプして実行
   │   ├─ 成功 → success(nil)を返す
   │   └─ 失敗 → errorを返す
   └─ コマンドなし
       ├─ osc52Attempted == true → success(nil)を返す（OSC 52が動作している可能性）
       └─ osc52Attempted == false → errorを返す（クリップボード手段なし）
```

**Implementation Steps**:

1. **OSC 52シーケンス生成と出力**
   - テキストをbase64エンコードし、OSC 52フォーマット (`\033]52;c;{base64}\a`) でラップする
   - `/dev/tty` を `os.OpenFile("/dev/tty", os.O_WRONLY, 0)` で開き、シーケンスを書き込む
   - `/dev/tty` のオープン失敗はエラーとして記録するが、外部コマンドフォールバックに進む
   - Key considerations:
     - ASCII/Unicodeの両方のファイル名に対応
     - `/dev/tty` への書き込みは Bubble Tea のレンダリングと競合しない
   - テスト容易性のため、OSC 52 の出力先を `io.Writer` パラメータとして受け取る。本番では `/dev/tty` を開いて渡し、テストでは `bytes.Buffer` を渡す

2. **外部コマンド検出**
   - `exec.LookPath` で `wl-copy`, `xclip`, `xsel` を順に検索する
   - 検出順序: wl-copy > xclip > xsel

3. **外部コマンド実行**
   - `context.WithTimeout(ctx, 5*time.Second)` + `exec.CommandContext` でタイムアウト付き実行
   - テキストは stdin パイプで渡す（シェル経由しない）

4. **クリップボード書き込み統合**
   - OSC 52 を `/dev/tty` 経由でまず出力し、外部コマンドが見つかればそちらも実行する
   - OSC 52 出力成功の有無 (`osc52Attempted`) を追跡し、外部コマンドもない場合のエラー判定に使用する

**Dependencies**:
- Requires: なし（独立モジュール）
- Blocks: Phase 3（Model統合）

**Testing Approach**:

*Unit Tests*:
- OSC 52シーケンスの正しいフォーマット検証（ASCII文字列）
- OSC 52シーケンスの正しいフォーマット検証（Unicode文字列）
- OSC 52出力先として `bytes.Buffer` を使用してシーケンス内容を検証
- 外部コマンド検出ロジック（コマンドの存在有無による分岐）
- 外部コマンド実行の成功/失敗パス
- 外部コマンドタイムアウト時のエラー返却
- OSC 52 失敗 + 外部コマンドなしの場合にエラーが返されること
- OSC 52 成功 + 外部コマンドなしの場合に success が返されること

**Acceptance Criteria**:
- [ ] OSC 52エスケープシーケンスが正しいフォーマットで生成される
- [ ] OSC 52 が `/dev/tty` 経由で出力される（`io.Writer` パラメータ経由）
- [ ] 外部コマンドの検出順序がwl-copy > xclip > xselである
- [ ] 外部コマンドに5秒のタイムアウトが設定される
- [ ] 外部コマンドが存在せずOSC 52が出力済みの場合はエラーにならない
- [ ] `/dev/tty` オープン失敗かつ外部コマンドなしの場合はエラーが返される
- [ ] 外部コマンド実行失敗時にエラーが返される
- [ ] 全ユニットテストがパスする

**Estimated Effort**: 小

**Risks and Mitigation**:
- **Risk**: OSC 52の動作確認がプログラム上困難
  - **Mitigation**: `io.Writer` パラメータにバッファを渡してフォーマットのみ検証する
- **Risk**: `/dev/tty` が利用できない環境（Docker等）
  - **Mitigation**: `/dev/tty` オープン失敗は non-fatal として外部コマンドフォールバックに進む

---

### Phase 2: Context Menu Items

**Goal**: コンテキストメニューに `copy_name` と `copy_path` の2項目を追加し、有効/無効条件を正しく設定する。

**Files to Modify**:
- `internal/ui/context_menu_dialog.go`:
  - `buildMenuItems()` にクリップボードコピー項目の追加を行うヘルパー呼び出しを追加
  - クリップボードコピーメニュー項目を生成するヘルパーメソッドを追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| buildClipboardMenuItems | copy_nameとcopy_pathのメニュー項目を生成する | entry, sourcePath, markCount | 2つのMenuItemが返される |

**Processing Flow**:
```
buildMenuItems()
  ├─ Open operations (既存)
  ├─ File operations: copy, move, delete (既存)
  ├─ Clipboard: copy_name, copy_path (新規 - deleteの後、compressの前)
  ├─ Compress (既存)
  ├─ Extract (既存)
  └─ Symlink operations (既存)
```

**Implementation Steps**:

1. **メニュー項目生成ヘルパーの追加**
   - `buildClipboardMenuItems` メソッドを追加
   - `copy_name`: ラベル "Copy file name"、有効条件は `markCount == 0 && !entry.IsParentDir()`
   - `copy_path`: ラベル "Copy full path"、有効条件は `markCount == 0 && !entry.IsParentDir()`
   - Action は nil（Phase 3でModel側で処理する）

2. **buildMenuItemsへの統合**
   - `buildFileOperationMenuItems` の呼び出しの後、`buildCompressMenuItem` の前に配置

**Dependencies**:
- Requires: なし
- Blocks: Phase 3（Model統合）

**Testing Approach**:

*Unit Tests*:
- 通常ファイルのメニューに `copy_name` と `copy_path` が含まれること
- ディレクトリ（非parent）のメニューに `copy_name` と `copy_path` が含まれること
- 親ディレクトリ（`..`）では両項目が無効であること
- マークファイルがある場合（markCount > 0）では両項目が無効であること
- メニュー内の配置順序が正しいこと（deleteの後、compressの前）
- 既存メニュー項目数の変化に対するテストの更新

**Acceptance Criteria**:
- [ ] `copy_name` メニュー項目が正しい位置に表示される
- [ ] `copy_path` メニュー項目が正しい位置に表示される
- [ ] 親ディレクトリでは両項目が無効
- [ ] マークファイルがある場合は両項目が無効
- [ ] 通常ファイル・ディレクトリでは両項目が有効
- [ ] 既存テストが全てパスする（項目数の更新含む）

**Estimated Effort**: 小

---

### Phase 3: Model Integration

**Goal**: `handleContextMenuResult` に `copy_name` と `copy_path` のハンドリングを追加し、クリップボード書き込みとステータスバーフィードバックを統合する。

**Files to Modify**:
- `internal/ui/model_update.go`:
  - `handleContextMenuResult` に `copy_name` と `copy_path` の分岐を追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| handleContextMenuResult (copy_name分岐) | ファイル名をクリップボードにコピーし、ステータスバーに表示 | 有効なentry | クリップボードにファイル名、ステータスバーにメッセージ |
| handleContextMenuResult (copy_path分岐) | フルパスをクリップボードにコピーし、ステータスバーに表示 | 有効なentry, activePane | クリップボードにフルパス、ステータスバーにメッセージ |

**Processing Flow**:
```
handleContextMenuResult
  ├─ (既存) open, open_with, delete, compress, extract
  ├─ actionID == "copy_name" (新規 - extractの後、copy/moveの前)
  │   ├─ entry取得
  │   ├─ statusMessage = "Copied: {filename}" を即座に設定（楽観的UI）
  │   ├─ tea.Batch(clipboardWriteCmd(text), statusMessageClearCmd(3s)) を返す
  │   └─ return m, cmd, true（ペインリフレッシュなし）
  ├─ actionID == "copy_path" (新規 - copy_nameの後、copy/moveの前)
  │   ├─ entry取得、activePane.Path()と結合
  │   ├─ statusMessage = "Copied: {fullpath}" を即座に設定（楽観的UI）
  │   ├─ tea.Batch(clipboardWriteCmd(text), statusMessageClearCmd(3s)) を返す
  │   └─ return m, cmd, true（ペインリフレッシュなし）
  ├─ (既存) copy, move
  └─ (既存) fallthrough: result.action実行 + ペインリフレッシュ

clipboardResultMsg 受信時
  ├─ err == nil → 何もしない（既に成功メッセージ表示済み）
  └─ err != nil → statusMessage = "Copy failed: {error}", isStatusError = true
```

**Implementation Steps**:

1. **copy_name / copy_path ハンドリングの配置位置**
   - `handleContextMenuResult` 内の既存分岐の中で、`extract` 分岐の後、`copy`/`move` 分岐の前に配置する
   - 具体的な配置順序:
     ```
     open → open_with → delete → compress → extract → copy_name (新規) → copy_path (新規) → copy/move → (fallthrough)
     ```
   - clipboard アクションは `Action: nil` のため、最後のフォールスルーブロック（`result.action != nil` で実行 + ペインリフレッシュ）には到達しないが、明示的に早期 return することでペインリフレッシュを確実に回避する

2. **copy_nameハンドリングの追加**
   - `result.actionID == "copy_name"` の分岐を追加
   - `activePane.SelectedEntry()` でエントリ取得
   - 楽観的UIパターン: `statusMessage` に `Copied: {filename}` を即座に設定
   - `tea.Batch(clipboardWriteCmd(text), statusMessageClearCmd(3 * time.Second))` を返す
   - `clipboardWriteCmd` は `tea.Cmd` として非同期でクリップボード書き込みを実行し、結果を `clipboardResultMsg` で通知する
   - ペインのリフレッシュは不要（`return m, cmd, true` で早期リターン）

3. **copy_pathハンドリングの追加**
   - `result.actionID == "copy_path"` の分岐を追加
   - `activePane.Path()` と `entry.Name` を `filepath.Join` で結合してフルパスを構築
   - 以降は copy_name と同じ流れ

4. **clipboardResultMsg のハンドリング**
   - `handleCustomMessages` に `clipboardResultMsg` の処理を追加
   - `err == nil`: 何もしない（楽観的UIで既に成功メッセージ表示済み）
   - `err != nil`: `statusMessage` を `Copy failed: {error}` に上書き、`isStatusError = true`、`statusMessageClearCmd(3 * time.Second)` を返す

**Dependencies**:
- Requires: Phase 1（clipboard module）、Phase 2（menu items）
- Blocks: なし

**Testing Approach**:

*Integration Tests*:
- copy_name選択時にステータスメッセージが設定されること
- copy_path選択時にステータスメッセージが設定されること
- クリップボード書き込みエラー時にエラーメッセージが設定されること
- ペインリフレッシュが呼ばれないこと（ファイルシステム変更なし）

**Acceptance Criteria**:
- [ ] `copy_name` 選択でファイル名がクリップボードに書き込まれる
- [ ] `copy_path` 選択でフルパスがクリップボードに書き込まれる
- [ ] 成功時に `Copied: {text}` がステータスバーに表示される
- [ ] 失敗時に `Copy failed: {error}` がステータスバーに表示される
- [ ] ステータスメッセージが3秒後にクリアされる
- [ ] ペインの不要なリフレッシュが発生しない

**Estimated Effort**: 小

## Complete File Structure

```
internal/
├── clipboard/
│   ├── clipboard.go          # クリップボード書き込み機能（OSC 52 + 外部コマンド）
│   └── clipboard_test.go     # クリップボードモジュールのユニットテスト
└── ui/
    ├── context_menu_dialog.go # メニュー項目追加（copy_name, copy_path）
    ├── context_menu_dialog_test.go # メニュー項目テスト追加
    └── model_update.go        # handleContextMenuResult にハンドリング追加
```

**File Descriptions**:
- `internal/clipboard/clipboard.go`: OSC 52エスケープシーケンス生成（/dev/tty経由）と外部コマンドフォールバック（5秒タイムアウト）によるクリップボード書き込み。UIから独立した純粋なクリップボード操作モジュール。
- `internal/clipboard/clipboard_test.go`: OSC 52フォーマット検証、外部コマンド検出、エラーハンドリングのテスト。
- `internal/ui/context_menu_dialog.go`: `buildClipboardMenuItems` メソッド追加と `buildMenuItems` への統合。
- `internal/ui/context_menu_dialog_test.go`: 新メニュー項目の存在・有効/無効条件・配置順序のテスト。既存テストの項目数更新。
- `internal/ui/model_update.go`: `handleContextMenuResult` に `copy_name`/`copy_path` の分岐追加。`clipboardResultMsg` のハンドリング追加。`clipboardWriteCmd` 関数追加。

## Testing Strategy

### Unit Testing

**clipboard module** (`internal/clipboard/`):

| Test | Description | Type |
|------|-------------|------|
| OSC 52 ASCII | ASCII文字列のbase64エンコードとシーケンスフォーマット検証 | Unit |
| OSC 52 Unicode | Unicode文字列のbase64エンコードとシーケンスフォーマット検証 | Unit |
| External cmd detection (wl-copy) | wl-copyが最優先で検出される | Unit |
| External cmd detection (xclip) | wl-copyがない場合xclipが検出される | Unit |
| External cmd detection (xsel) | wl-copy, xclipがない場合xselが検出される | Unit |
| External cmd detection (none) | 全コマンドがない場合のフォールバック動作 | Unit |
| External cmd exec failure | 外部コマンド実行失敗時のエラー返却 | Unit |
| External cmd timeout | 外部コマンドが5秒以内に完了しない場合のタイムアウト | Unit |
| OSC 52 + no external cmd | OSC 52出力済み + 外部コマンドなしの場合はsuccess | Unit |
| No /dev/tty + no external cmd | /dev/ttyオープン失敗 + 外部コマンドなしの場合はerror | Unit |

**context menu items** (`internal/ui/`):

| Test | Description | Type |
|------|-------------|------|
| Menu items presence | copy_name, copy_pathが存在する | Unit |
| Menu items position | deleteの後、compressの前に配置される | Unit |
| Parent dir disabled | `..`で両項目が無効 | Unit |
| Marked files disabled | markCount > 0で両項目が無効 | Unit |
| Regular file enabled | 通常ファイルで両項目が有効 | Unit |
| Directory enabled | ディレクトリ（非parent）で両項目が有効 | Unit |

**model integration** (`internal/ui/`):

| Test | Description | Type |
|------|-------------|------|
| copy_name status message | copy_name選択時にステータスメッセージが設定される | Integration |
| copy_path status message | copy_path選択時にステータスメッセージが設定される | Integration |
| Copy error handling | クリップボードエラー時にエラーメッセージが設定される | Integration |

### E2E Tests (Docker)

- [ ] コンテキストメニューを開き、「Copy file name」を選択し、ステータスバーメッセージを確認
- [ ] コンテキストメニューを開き、「Copy full path」を選択し、ステータスバーメッセージを確認

### Manual Testing (E2E Not Possible)

- [ ] OSC 52対応ターミナルでクリップボードにテキストが実際にコピーされる
- [ ] xclip/xsel/wl-copyがインストールされた環境でクリップボードにテキストが実際にコピーされる
- [ ] クリップボードツールが一切ない環境でエラーにならない

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `encoding/base64` | Go標準ライブラリ | OSC 52のbase64エンコード |
| `os/exec` | Go標準ライブラリ | 外部コマンド検出・実行 |
| `context` | Go標準ライブラリ | 外部コマンドの5秒タイムアウト制御 |

### Internal Dependencies

**Implementation Order** (依存関係順):
1. Phase 1: clipboard module（依存なし）
2. Phase 2: context menu items（依存なし、Phase 1と並行可能）
3. Phase 3: model integration（Phase 1, 2に依存）

**Component Dependencies**:
- `internal/ui/model_update.go` → `internal/clipboard/clipboard.go`
- `internal/ui/context_menu_dialog.go` → 変更のみ（新規依存なし）

## Risk Assessment

### Technical Risks

1. **OSC 52のターミナル互換性**
   - **Risk**: OSC 52非対応ターミナルでは無視されるだけだが、ユーザーには成功と表示される
   - **Likelihood**: 中
   - **Impact**: 低（外部コマンドフォールバックで補完される）
   - **Mitigation**: 外部コマンドが見つかった場合は必ず外部コマンドも実行する（belt-and-suspenders）

2. **OSC 52出力がBubble Teaのレンダリングと競合する可能性**
   - **Risk**: `os.Stdout` への直接書き込みが Bubble Tea の画面出力と競合し、画面が乱れる
   - **Likelihood**: 高（altscreen モードでは確実に競合する）
   - **Impact**: 高（画面が乱れる）
   - **Resolution**: `/dev/tty` に直接書き込むことで解決済み。`/dev/tty` への書き込みは Bubble Tea のレンダラーが管理する `os.Stdout` とは独立しており、競合しない。`tea.Printf` / `tea.Println` は altscreen モードでは出力されないため使用不可。

3. **/dev/tty が利用できない環境**
   - **Risk**: Docker コンテナや CI 環境では `/dev/tty` が存在しない場合がある
   - **Likelihood**: 低（通常のターミナル使用では問題なし）
   - **Impact**: 低（外部コマンドフォールバックで補完される）
   - **Mitigation**: `/dev/tty` オープン失敗は non-fatal として外部コマンドにフォールバックする

4. **外部コマンドのハング**
   - **Risk**: 外部コマンド（wl-copy, xclip, xsel）が応答しない場合に TUI がブロックされる
   - **Likelihood**: 低
   - **Impact**: 高（TUI 全体がフリーズする）
   - **Mitigation**: `context.WithTimeout(5*time.Second)` + `exec.CommandContext` で5秒タイムアウトを設定

## Performance Considerations

- クリップボード操作は100ms以内に完了する想定（NFR1）
- `/dev/tty` への OSC 52 書き込みは即座に完了する（ローカルデバイスへの書き込み）
- 外部コマンド実行は `context.WithTimeout(5*time.Second)` + `exec.CommandContext` で5秒タイムアウトを設定する
  - 通常のクリップボードコマンドは数ミリ秒で完了するが、異常時のハング防止のためタイムアウトを設ける
  - タイムアウト超過時は `context.DeadlineExceeded` エラーとして処理される
- **NFR1 との整合性**: 外部コマンドの5秒タイムアウトは NFR1 (100ms) と矛盾するため、クリップボード書き込み全体を `tea.Cmd` のコマンド関数内で非同期に実行する。`handleContextMenuResult` ではステータスメッセージを即座に設定し、`tea.Cmd` として clipboard 操作を返す。clipboard 操作の結果は新しいメッセージ型 `clipboardResultMsg` で Model に通知し、エラー時のみステータスバーを更新する

## Security Considerations

- 外部コマンドはフルパスではなくPATH検索で実行する（`exec.LookPath`）
- 入力テキストはシェルを経由せず、stdinパイプで渡す（コマンドインジェクション防止）

## Open Questions

なし（SPEC.mdに未確認事項なし）。

## Success Metrics

### Functional Completeness
- [ ] 両コンテキストメニュー項目が表示され、正常に機能する
- [ ] OSC 52エスケープシーケンスが出力される
- [ ] 外部コマンドフォールバックが利用可能な場合に動作する
- [ ] 成功・失敗時のステータスバーフィードバックが表示される
- [ ] 親ディレクトリとマークファイルで項目が正しく無効化される
- [ ] 全ユニットテストがパスする
- [ ] 既存コンテキストメニュー機能にリグレッションがない

### Quality Metrics
- [ ] テストカバレッジ: clipboard module 90%以上
- [ ] コードが `gofmt` でフォーマット済み
- [ ] `go vet` で問題なし

## References

- SPEC.md: `doc/tasks/clipboard-copy/SPEC.md`
- 要件定義書: `doc/tasks/clipboard-copy/要件定義書.md`
- OSC 52 Specification: ANSI escape sequence for clipboard access
- 既存コンテキストメニュー: `internal/ui/context_menu_dialog.go`
- 結果ハンドリング: `internal/ui/model_update.go` (`handleContextMenuResult`)
