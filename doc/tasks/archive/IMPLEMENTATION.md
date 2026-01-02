# Implementation Plan: Archive Operations

## Overview

アーカイブ機能をduofmに追加し、ユーザーがTUI内で直接ファイル/ディレクトリの圧縮と展開を実行できるようにする。UNIX哲学に基づき、外部CLIツール（tar, gzip, bzip2, xz, zip, 7z）を活用してシンプルかつ堅牢な実装を実現する。

## Objectives

- tar, tar.gz, tar.bz2, tar.xz, zip, 7z形式での圧縮・伸長機能を提供
- 既存のコンテキストメニュー（@キー）にシームレスに統合
- 単一ファイル/ディレクトリおよび複数ファイル（マーク選択）の圧縮をサポート
- スマート展開ロジック（アーカイブ構造に応じた展開方法の自動判定）を実装
- バックグラウンド処理によるUI応答性の維持
- リアルタイム進捗表示と操作のキャンセル可能性を提供
- セキュリティ対策（パストラバーサル、圧縮爆弾、シンボリンク）を実装

## Prerequisites

### Development Environment

- Go 1.21以上
- Linux環境（開発・実行共に）
- make（ビルド自動化）

### Dependencies

**外部CLIツール（実行時）**:
- `tar`, `gzip`, `bzip2`, `xz` - 必須（通常はプリインストール済み）
- `zip`, `unzip` - オプション（zip形式サポート用）
- `7z` - オプション（7z形式サポート用）

**Goライブラリ**:
- `os/exec` - 外部コマンド実行（標準ライブラリ）
- `context` - キャンセル処理（標準ライブラリ）
- `github.com/charmbracelet/bubbletea` - TUIフレームワーク（既存）
- `github.com/charmbracelet/lipgloss` - スタイリング（既存）

### Knowledge Requirements

- Bubble Teaアーキテクチャ（Model-Update-Viewパターン）
- Goのgoroutineとchannel（非同期処理）
- Goのcontext（キャンセル処理）
- `os/exec`パッケージの使用方法
- Linuxコマンドライン（tar, zip, 7z等）の基本

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUIフレームワーク)
- **Styling**: Lip Gloss
- **External Tools**: tar, gzip, bzip2, xz, zip, unzip, 7z（外部CLIツール）

### Design Approach

**UNIX哲学に基づく設計**:
- 外部CLIツールに処理を委譲（Do One Thing Well）
- 薄いラッパー層のみをGoで実装
- 標準入出力を通じた連携

**レイヤード アーキテクチャ**:
```
┌─────────────────────────────────┐
│  UI Layer (Bubble Tea Dialogs)  │  ← ユーザーインタラクション
├─────────────────────────────────┤
│  Controller Layer               │  ← 操作フロー制御
├─────────────────────────────────┤
│  Service Layer (CLI Wrapper)    │  ← コマンド実行ラッパー
├─────────────────────────────────┤
│  External CLI Tools             │  ← tar, zip, 7z等
└─────────────────────────────────┘
```

### Component Interaction

1. **ユーザー操作** → ContextMenuDialog（形式選択）
2. **形式選択** → CompressionLevelDialog（レベル選択）
3. **レベル選択** → ArchiveNameDialog（名前入力）
4. **名前確定** → ArchiveController（操作開始）
5. **Controller** → TaskManager（非同期タスク管理）
6. **TaskManager** → CommandExecutor（CLIコマンド実行）
7. **CommandExecutor** → 外部CLIツール（tar/zip/7z）
8. **進捗更新** → ProgressDialog（UI更新）
9. **完了** → ステータス通知、ファイルリスト更新

## Implementation Phases

### Phase 1: Core Infrastructure (基盤構築)

**Goal**: アーカイブ操作の基盤となるデータ構造、コマンド実行基盤、形式検出を実装する。この段階で外部ツールとの連携が可能になる。

**Files to Create**:
- `internal/archive/format.go` - アーカイブ形式の定義と検出
- `internal/archive/command_availability.go` - 外部コマンドの存在確認
- `internal/archive/command_executor.go` - 外部CLIコマンド実行ラッパー
- `internal/archive/errors.go` - エラー定義
- `internal/archive/progress.go` - 進捗情報の構造体
- `internal/archive/format_test.go` - 形式検出のテスト
- `internal/archive/command_availability_test.go` - コマンド確認のテスト
- `internal/archive/command_executor_test.go` - コマンド実行のテスト

**Files to Modify**:
なし（このフェーズでは新規ファイルのみ）

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ArchiveFormat | 形式の列挙型定義 | なし | tar, tar.gz等の形式が定義されている |
| FormatDetector | ファイル拡張子から形式判定 | ファイルパスが存在する | 対応形式またはエラーを返す |
| CommandAvailability | 外部コマンドの存在確認 | なし | 利用可能な形式リストを返す |
| CommandExecutor | CLIコマンド実行とエラーハンドリング | コマンドとオプションが指定されている | 実行結果またはエラーを返す |
| ProgressUpdate | 進捗情報を保持する構造体 | なし | 処理済みファイル数、進捗率等を保持 |
| ArchiveError | エラーコードと詳細情報を保持 | なし | エラー種別を識別可能な構造化されたエラー |

**errors.goの実装内容**:

```go
// Error codes for archive operations (defined in SPEC.md)
const (
	ErrArchiveSourceNotFound       = "ERR_ARCHIVE_001" // Source file not found
	ErrArchivePermissionDeniedRead = "ERR_ARCHIVE_002" // Permission denied (read)
	ErrArchivePermissionDeniedWrite = "ERR_ARCHIVE_003" // Permission denied (write)
	ErrArchiveDiskSpaceInsufficient = "ERR_ARCHIVE_004" // Disk space insufficient
	ErrArchiveUnsupportedFormat    = "ERR_ARCHIVE_005" // Unsupported format
	ErrArchiveCorrupted            = "ERR_ARCHIVE_006" // Corrupted archive
	ErrArchiveInvalidName          = "ERR_ARCHIVE_007" // Invalid archive name
	ErrArchivePathTraversal        = "ERR_ARCHIVE_008" // Path traversal detected
	ErrArchiveCompressionBomb      = "ERR_ARCHIVE_009" // Compression bomb detected
	ErrArchiveOperationCancelled   = "ERR_ARCHIVE_010" // Operation cancelled
	ErrArchiveIOError              = "ERR_ARCHIVE_011" // I/O error
	ErrArchiveInternalError        = "ERR_ARCHIVE_012" // Internal error
)

// ArchiveError represents a structured error with code and details
type ArchiveError struct {
	Code    string // Error code (ERR_ARCHIVE_XXX)
	Message string // User-friendly message
	Details string // Technical details for logging
	Cause   error  // Underlying error (if any)
}

func (e *ArchiveError) Error() string {
	return e.Message
}

func (e *ArchiveError) Unwrap() error {
	return e.Cause
}

// NewArchiveError creates a new ArchiveError with the given code and message
func NewArchiveError(code, message string, cause error) *ArchiveError {
	return &ArchiveError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
```

**Processing Flow**:
```
1. FormatDetector.DetectFormat(filePath)
   ├─ 拡張子マッピングで判定 → 対応形式を返す
   └─ 未対応拡張子 → エラー

2. CommandAvailability.GetAvailableFormats()
   ├─ 各形式に必要なコマンドを確認（exec.LookPath）
   ├─ 利用可能な形式をリストに追加
   └─ リストを返す

3. CommandExecutor.ExecuteCompress(ctx, sources, output, opts)
   ├─ 形式に応じたコマンドライン構築
   ├─ exec.CommandContext()でコマンド実行
   ├─ 標準出力/エラー出力を監視
   └─ 完了またはエラーを返す
```

**Implementation Steps**:

1. **形式定義とマッピング**
   - `ArchiveFormat`列挙型を定義（Tar, TarGz, TarBz2, TarXz, Zip, SevenZ）
   - 拡張子から形式へのマッピング関数を実装
   - 各形式に必要なコマンドリストをマップで定義

2. **コマンド存在確認**
   - `exec.LookPath()`を使用して外部コマンドの存在を確認
   - 形式ごとに必要なコマンドをチェック（例: TarGzはtar+gzip）
   - 利用可能な形式リストを返す関数を実装

3. **コマンド実行ラッパー**
   - `exec.CommandContext()`を使用してコマンド実行
   - 標準出力/エラー出力のキャプチャ
   - エラーハンドリング（コマンドが見つからない、実行失敗等）
   - コンテキストによるキャンセル対応

**Dependencies**:
- Requires: なし（独立して実装可能）
- Blocks: Phase 2（CLI統合）、Phase 3（UI統合）

**Testing Approach**:

*Unit Tests*:
- FormatDetector: 各拡張子の正しい形式判定
- CommandAvailability: モック環境でのコマンド存在確認
- CommandExecutor: モックコマンドでの実行と出力キャプチャ

*Integration Tests*:
- 実際のtar/zipコマンドを使用した圧縮・伸長テスト（小規模データ）
- コマンドが存在しない場合のエラーハンドリング

*Manual Testing*:
- [ ] tar形式の検出が正しく動作する
- [ ] tar.gz, tar.bz2, tar.xz, zip, 7z各形式が検出される
- [ ] 利用可能な形式リストが正しく返される
- [ ] 外部コマンド実行が成功する
- [ ] コマンドが存在しない場合にエラーが返される

**Acceptance Criteria**:
- [ ] 6つのアーカイブ形式（tar, tar.gz, tar.bz2, tar.xz, zip, 7z）が定義されている
- [ ] 拡張子から形式を正しく判定できる
- [ ] 外部コマンドの存在を確認し、利用可能な形式リストを取得できる
- [ ] 外部コマンドを実行し、標準出力/エラー出力をキャプチャできる
- [ ] コンテキストによるキャンセルが機能する
- [ ] すべてのテストがパスする

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:
- **Risk**: 外部コマンドのバージョン差異によるオプション非互換
  - **Mitigation**: 汎用的で広くサポートされているオプションのみを使用する
- **Risk**: 出力解析の失敗
  - **Mitigation**: 出力解析はオプション（進捗表示）とし、失敗しても操作自体は成功させる

---

### Phase 2: CLI Integration (CLIツール統合)

**Goal**: 各アーカイブ形式ごとのコマンドラインオプションを実装し、実際の圧縮・伸長処理を動作させる。進捗情報の解析も含む。

**Files to Create**:
- `internal/archive/tar_executor.go` - tar系コマンド（tar, tar.gz, tar.bz2, tar.xz）の実行
- `internal/archive/zip_executor.go` - zip/unzipコマンドの実行
- `internal/archive/sevenzip_executor.go` - 7zコマンドの実行
- `internal/archive/smart_extractor.go` - スマート展開ロジック
- `internal/archive/tar_executor_test.go` - tarコマンドのテスト
- `internal/archive/zip_executor_test.go` - zipコマンドのテスト
- `internal/archive/sevenzip_executor_test.go` - 7zコマンドのテスト
- `internal/archive/smart_extractor_test.go` - スマート展開のテスト

**Files to Modify**:
- `internal/archive/command_executor.go`:
  - 各形式に応じてtar_executor, zip_executor, sevenzip_executorに処理を委譲

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TarExecutor | tar系コマンドの引数構築と実行 | ソースファイルとオプションが指定されている | tarアーカイブが生成される |
| ZipExecutor | zip/unzipコマンドの引数構築と実行 | ソースファイルとオプションが指定されている | zipアーカイブが生成/展開される |
| SevenZipExecutor | 7zコマンドの引数構築と実行 | ソースファイルとオプションが指定されている | 7zアーカイブが生成/展開される |
| SmartExtractor | アーカイブ内容解析と展開戦略決定 | アーカイブファイルが存在する | 展開戦略（直接/ディレクトリ作成）を決定 |
| ProgressParser | コマンド出力から進捗情報を抽出 | コマンドの標準出力が取得できる | 進捗率とファイル数を返す |

**Processing Flow**:
```
1. 圧縮処理フロー:
   ├─ TarExecutor.Compress(sources, output, level)
   │   ├─ 形式に応じたオプション構築（-czvf, -cjvf, -cJvf等）
   │   ├─ 圧縮レベル指定（該当する場合）
   │   ├─ コマンド実行（tar -czvf output.tar.gz sources...）
   │   └─ 標準出力監視（進捗情報抽出）
   └─ 完了またはエラー

2. 伸長処理フロー:
   ├─ SmartExtractor.AnalyzeStructure(archivePath)
   │   ├─ アーカイブ内容リスト取得（tar -tvf, unzip -l, 7z l）
   │   ├─ ルート要素数を判定
   │   └─ 展開戦略を決定（単一ディレクトリ → 直接、複数 → ディレクトリ作成）
   ├─ TarExecutor.Extract(archivePath, destDir, strategy)
   │   ├─ 展開戦略に応じた処理
   │   ├─ コマンド実行（tar -xzvf archivePath -C destDir）
   │   └─ 標準出力監視
   └─ 完了またはエラー
```

**Implementation Steps**:

1. **Tar系コマンドの実装**
   - tar, tar.gz, tar.bz2, tar.xzそれぞれの圧縮オプション定義
   - 圧縮レベル指定（gzip: -N, bzip2: -N, xz: -N）
   - 伸長オプション定義（-xzvf, -xjvf, -xJvf）
   - 進捗情報の抽出（tar -vオプションの出力解析）

2. **Zip/7zコマンドの実装**
   - zipコマンドの圧縮オプション（-r, -N）
   - unzipコマンドの伸長オプション
   - 7zコマンドの圧縮・伸長オプション（a, x, -mx=N）
   - 各コマンドの出力形式に応じた進捗情報抽出

3. **スマート展開ロジック**
   - アーカイブ内容リスト取得（tar -tvf, unzip -l, 7z l）
   - ルート要素のカウント
   - 単一ディレクトリか複数ファイルかの判定
   - 展開先ディレクトリの決定（直接展開 vs ディレクトリ作成）

**Dependencies**:
- Requires: Phase 1（Core Infrastructure）完了
- Blocks: Phase 3（UI統合）、Phase 4（スマート機能）

**Testing Approach**:

*Unit Tests*:
- 各Executorのコマンドライン構築ロジック
- 圧縮レベル指定の正しさ
- 進捗情報抽出の正確性（モック出力使用）

*Integration Tests*:
- 実際のファイルを使用した圧縮・伸長
- 各形式（tar, tar.gz, tar.bz2, tar.xz, zip, 7z）での往復テスト
- スマート展開ロジックの検証（単一ディレクトリ、複数ファイル）

*Manual Testing*:
- [ ] tar形式での圧縮・伸長が成功する
- [ ] tar.gz, tar.bz2, tar.xz各形式での圧縮・伸長が成功する
- [ ] zip形式での圧縮・伸長が成功する
- [ ] 7z形式での圧縮・伸長が成功する
- [ ] 圧縮レベル0-9すべてで圧縮が成功する
- [ ] スマート展開が正しく動作する（単一ディレクトリは直接、複数は新ディレクトリ）
- [ ] ファイル権限とタイムスタンプが保持される
- [ ] シンボリックリンクがリンクとして保存される

**Acceptance Criteria**:
- [ ] tar, tar.gz, tar.bz2, tar.xz, zip, 7zすべての形式で圧縮できる
- [ ] すべての形式で伸長できる
- [ ] 圧縮レベル0-9が正しく適用される
- [ ] スマート展開ロジックが正しく動作する
- [ ] 進捗情報が正しく抽出される
- [ ] ファイル属性（権限、タイムスタンプ、シンボリックリンク）が保持される
- [ ] すべてのテストがパスする

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:
- **Risk**: 進捗情報の出力形式がコマンドバージョンで異なる
  - **Mitigation**: 進捗抽出は失敗しても処理を継続（進捗表示なしで完了）
- **Risk**: スマート展開ロジックの誤判定
  - **Mitigation**: 保守的な判定（不明な場合はディレクトリ作成を選択）

---

### Phase 3: UI Integration (UI統合)

**Goal**: 既存のコンテキストメニューにアーカイブ操作を追加し、必要なダイアログを実装する。ユーザーがGUI操作でアーカイブを作成・展開できるようにする。

**Files to Create**:
- `internal/ui/compression_level_dialog.go` - 圧縮レベル選択ダイアログ
- `internal/ui/compression_level_dialog_test.go` - 圧縮レベル選択のテスト
- `internal/ui/archive_name_dialog.go` - アーカイブ名入力ダイアログ
- `internal/ui/archive_name_dialog_test.go` - 名前入力のテスト
- `internal/ui/archive_progress_dialog.go` - 進捗表示ダイアログ
- `internal/ui/archive_progress_dialog_test.go` - 進捗表示のテスト

**Files to Modify**:
- `internal/ui/context_menu_dialog.go`:
  - "Compress"メニュー項目を追加（サブメニュー付き）
  - "Extract archive"メニュー項目を追加（アーカイブファイル選択時のみ）
  - マーク選択時のラベル変更（"Compress N files"）
- `internal/ui/messages.go`:
  - アーカイブ操作用メッセージ型を追加（archiveOperationMsg等）
- `internal/ui/model.go`:
  - アーカイブ操作の開始処理を追加
  - 進捗ダイアログの表示制御を追加
  - 完了時のステータス表示とファイルリスト更新

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| CompressionLevelDialog | 圧縮レベル（0-9）の選択UIを提供 | 圧縮形式が選択されている | 圧縮レベルが決定される |
| ArchiveNameDialog | アーカイブ名の入力・編集UIを提供 | デフォルト名が生成されている | 最終的なアーカイブ名が確定する |
| ArchiveProgressDialog | 進捗率、ファイル数、経過時間を表示 | 非同期タスクが実行中 | リアルタイムで進捗が更新される |
| ContextMenuDialog (拡張) | Compress/Extractメニュー項目を追加 | ファイル/ディレクトリが選択されている | アーカイブ操作が開始される |

**Processing Flow**:
```
1. 圧縮操作のUIフロー:
   ユーザーが@キーでメニュー表示
   → "Compress"を選択
   → サブメニューで形式選択（tar/tar.gz/tar.bz2/tar.xz/zip/7z）
   → （tar以外の場合）CompressionLevelDialog表示
   → レベル選択（0-9またはEscでデフォルト6）
   → ArchiveNameDialog表示
   → 名前入力・確定
   → 同名ファイル存在確認
      ├─ 存在する → OverwriteDialog表示（Overwrite/Rename/Cancel）
      └─ 存在しない → 圧縮開始
   → ArchiveProgressDialog表示（バックグラウンド処理開始）
   → 完了後、ステータス通知とファイルリスト更新

2. 伸長操作のUIフロー:
   ユーザーがアーカイブファイルを選択
   → @キーでメニュー表示
   → "Extract archive"を選択
   → ArchiveProgressDialog表示（バックグラウンド処理開始）
   → 完了後、ステータス通知とファイルリスト更新
```

**Implementation Steps**:

1. **圧縮レベル選択ダイアログ**
   - 0-9の選択UI（j/k移動、数字キー直接選択）
   - 各レベルの説明表示（0: None, 3: Fast, 6: Normal, 9: Best）
   - デフォルト値6をハイライト
   - Escキーでデフォルト選択

2. **アーカイブ名入力ダイアログ**
   - 既存のInputDialogを拡張または新規実装
   - デフォルト名の自動生成（単一対象: 元の名前+拡張子、複数: 親ディレクトリ名+拡張子）
   - バリデーション（空文字列、不正文字チェック）
   - エラーメッセージ表示

3. **進捗表示ダイアログ**
   - プログレスバー（0-100%）
   - 処理済み/総ファイル数表示
   - 現在処理中のファイル名（長い場合は省略）
   - 経過時間表示（MM:SS形式）
   - 推定残り時間（計算可能な場合）
   - Escキーでキャンセル可能

4. **コンテキストメニュー拡張**
   - "Compress"メニュー項目追加（サブメニュー付き）
   - サブメニュー項目: tar, tar.gz, tar.bz2, tar.xz, zip, 7z
   - 利用可能な形式のみ表示（CommandAvailability使用）
   - "Extract archive"メニュー項目追加（アーカイブファイル選択時のみ）
   - マーク選択時のラベル変更

**Dependencies**:
- Requires: Phase 1（Core Infrastructure）、Phase 2（CLI Integration）完了
- Blocks: Phase 4（スマート機能）、Phase 5（セキュリティとエラーハンドリング）

**Testing Approach**:

*Unit Tests*:
- 各ダイアログのキーボード入力処理
- バリデーションロジック
- デフォルト値の生成

*Integration Tests*:
- ダイアログ間の遷移フロー
- メニューからの操作開始
- 進捗ダイアログの更新

*Manual Testing*:
- [ ] コンテキストメニューに"Compress"と"Extract archive"が表示される
- [ ] 利用可能な形式のみがサブメニューに表示される
- [ ] 圧縮レベル選択が正しく動作する（0-9選択、Escでデフォルト）
- [ ] アーカイブ名入力が正しく動作する（デフォルト名表示、編集、バリデーション）
- [ ] 進捗ダイアログが表示され、リアルタイムで更新される
- [ ] Escキーで操作をキャンセルできる
- [ ] 完了後にステータス通知が表示される
- [ ] ファイルリストが更新される

**Acceptance Criteria**:
- [ ] コンテキストメニューから圧縮操作を開始できる
- [ ] コンテキストメニューから伸長操作を開始できる
- [ ] 圧縮レベル選択ダイアログが表示され、選択できる
- [ ] アーカイブ名入力ダイアログが表示され、名前を編集できる
- [ ] 進捗ダイアログが表示され、進捗がリアルタイムで更新される
- [ ] 操作をキャンセルできる
- [ ] 完了後にステータス通知とファイルリスト更新が行われる
- [ ] すべてのテストがパスする

**Estimated Effort**: 大 (4-5 days)

**Risks and Mitigation**:
- **Risk**: ダイアログ遷移が複雑になりバグが発生
  - **Mitigation**: 状態遷移図を作成し、各遷移を個別にテスト
- **Risk**: 進捗更新の頻度が高すぎてUI描画が遅延
  - **Mitigation**: 更新頻度を最大10Hz（100ms間隔）に制限

---

### Phase 4: Task Management and Background Processing (タスク管理と非同期処理)

**Goal**: アーカイブ操作をバックグラウンドで実行し、UI応答性を維持する。進捗情報のリアルタイム更新とキャンセル処理を実装する。

**Files to Create**:
- `internal/archive/task_manager.go` - 非同期タスク管理
- `internal/archive/task_manager_test.go` - タスク管理のテスト
- `internal/archive/archive.go` - アーカイブコントローラー（統合API）
- `internal/archive/archive_test.go` - コントローラーのテスト

**Files to Modify**:
- `internal/ui/model.go`:
  - アーカイブタスクの開始と進捗監視
  - タスク完了メッセージの処理
  - キャンセル処理の実装
- `internal/ui/messages.go`:
  - archiveTaskStartMsg, archiveTaskProgressMsg, archiveTaskCompleteMsg追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| TaskManager | 非同期タスクの管理とライフサイクル制御 | タスクパラメータが指定されている | タスクが実行され、進捗が通知される |
| ArchiveController | 圧縮・伸長操作の統合API提供 | 必要なパラメータが揃っている | TaskManagerに処理を委譲し、タスクIDを返す |
| ProgressTracker | 進捗情報の集約と計算 | 総ファイル数とサイズが判明している | 進捗率と推定残り時間を計算 |
| CancellationHandler | キャンセル要求の処理 | タスクが実行中 | タスクを中断し、部分ファイルを削除 |

**Processing Flow**:
```
1. タスク開始フロー:
   ArchiveController.CreateArchive(sources, dest, format, level)
   → TaskManager.StartTask(taskParams)
      ├─ タスクIDを生成（UUID）
      ├─ goroutineで処理を開始
      │   ├─ ProgressTrackerを初期化（総ファイル数計算）
      │   ├─ CommandExecutor呼び出し
      │   ├─ 標準出力監視（進捗情報抽出）
      │   └─ 進捗chanに定期送信（100ms間隔）
      └─ タスクIDを返す

2. 進捗更新フロー:
   進捗chan受信
   → UI Model更新（archiveTaskProgressMsg）
   → ArchiveProgressDialog表示更新
   → 進捗率、ファイル数、経過時間を表示

3. キャンセルフロー:
   ユーザーがEscキー押下
   → TaskManager.CancelTask(taskID)
   → context.Cancel()呼び出し
   → goroutineがキャンセルを検知
   → 部分ファイル削除
   → archiveTaskCompleteMsg(cancelled=true)送信

4. 完了フロー:
   処理完了
   → archiveTaskCompleteMsg送信
   → ProgressDialog非表示
   → ステータス通知表示
   → ファイルリスト更新
```

**Implementation Steps**:

1. **TaskManager実装**
   - タスクID生成（UUIDまたは連番）
   - goroutineでのタスク実行
   - context.Context使用によるキャンセル対応
   - 進捗chanの作成と管理
   - タスク完了通知

2. **ProgressTracker実装**
   - 総ファイル数とサイズの事前計算
   - 処理済みファイル数/バイト数の追跡
   - 進捗率（0-100%）の計算
   - 経過時間の測定
   - 推定残り時間の計算（処理速度から算出）

3. **ArchiveController実装**
   - CreateArchive()メソッド（圧縮操作）
   - ExtractArchive()メソッド（伸長操作）
   - TaskManagerへの処理委譲
   - エラーハンドリングと通知

4. **UI Model統合**
   - アーカイブタスク開始時のメッセージ処理
   - 進捗更新メッセージの処理
   - タスク完了メッセージの処理
   - キャンセル処理の実装

**Dependencies**:
- Requires: Phase 1（Core Infrastructure）、Phase 2（CLI Integration）、Phase 3（UI統合）完了
- Blocks: Phase 5（セキュリティとエラーハンドリング）

**Testing Approach**:

*Unit Tests*:
- TaskManagerのタスク管理ロジック
- ProgressTrackerの進捗計算
- キャンセル処理の正確性

*Integration Tests*:
- 実際のファイルを使用した非同期圧縮・伸長
- 進捗更新の正確性
- キャンセル処理とクリーンアップ

*Manual Testing*:
- [ ] 圧縮操作がバックグラウンドで実行される
- [ ] 伸長操作がバックグラウンドで実行される
- [ ] 進捗ダイアログがリアルタイムで更新される
- [ ] 操作中もUIが応答する（他の操作が可能）
- [ ] キャンセルが正しく動作し、部分ファイルが削除される
- [ ] 完了後にステータス通知が表示される
- [ ] エラー時に適切なエラーメッセージが表示される

**Acceptance Criteria**:
- [ ] アーカイブ操作がバックグラウンドで実行される
- [ ] UI応答性が維持される（100ms以内のキー入力応答）
- [ ] 進捗情報がリアルタイムで更新される（最大10Hz）
- [ ] キャンセルが1秒以内に完了する
- [ ] 部分ファイルが正しく削除される
- [ ] 完了後の通知とファイルリスト更新が正しく動作する
- [ ] すべてのテストがパスする

**Estimated Effort**: 大 (4-5 days)

**Risks and Mitigation**:
- **Risk**: goroutineリーク
  - **Mitigation**: contextによる確実なキャンセル、defer文での後処理
- **Risk**: 進捗更新の頻度が高すぎてパフォーマンス低下
  - **Mitigation**: 更新頻度を100ms間隔に制限、小ファイルはスキップ

---

### Phase 5: Security and Error Handling (セキュリティとエラーハンドリング)

**Goal**: セキュリティ脅威（パストラバーサル、圧縮爆弾、シンボリンク攻撃）への対策と、包括的なエラーハンドリングを実装する。

**Files to Create**:
- `internal/archive/security.go` - セキュリティ検証
- `internal/archive/security_test.go` - セキュリティ検証のテスト
- `internal/archive/validation.go` - 入力検証
- `internal/archive/validation_test.go` - 入力検証のテスト

**Files to Modify**:
- `internal/archive/smart_extractor.go`:
  - セキュリティ検証を追加
  - パストラバーサルチェック
  - 圧縮爆弾検出
- `internal/archive/tar_executor.go`, `zip_executor.go`, `sevenzip_executor.go`:
  - 展開前のセキュリティ検証呼び出し
  - エラーハンドリング強化
- `internal/ui/model.go`:
  - セキュリティエラーの表示
  - 警告ダイアログの追加

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| PathValidator | パストラバーサル検証 | ファイルパスリストが存在する | 安全なパスのみを許可、危険なパスを拒否 |
| CompressionRatioChecker | 圧縮爆弾検出 | アーカイブファイルと展開後サイズが判明 | 圧縮率が1:1000超で警告ダイアログ表示 |
| DiskSpaceChecker | ディスク容量チェック | 展開先ディレクトリと展開後サイズが判明 | 空き容量不足時に警告ダイアログ表示 |
| SymlinkValidator | シンボリックリンク安全性検証 | シンボリックリンクリストが存在する | 危険なシンボリンク（絶対パス等）を検出 |
| InputValidator | ファイル名等の入力検証 | 入力文字列が存在する | 不正文字やNUL文字を検出 |

**Processing Flow**:
```
1. 展開前のセキュリティ検証:
   SmartExtractor.AnalyzeStructure(archivePath)
   → メタデータコマンド実行（tar -tvf, unzip -l, 7z l）
   → アーカイブ内容リスト取得と総サイズ計算
   → PathValidator.ValidatePaths(paths)
      ├─ ".."を含むパスを検出 → エラー
      ├─ 絶対パスを検出 → 警告
      └─ すべて相対パス → OK
   → SymlinkValidator.ValidateSymlinks(symlinks)
      ├─ 絶対パスシンボリンクを検出 → 警告
      └─ 相対パスシンボリンク → OK
   → CompressionRatioChecker.CheckRatio(archivePath, extractedSize)
      ├─ 圧縮率 > 1:1000 → 警告ダイアログ表示（Continue/Cancel）
      └─ 圧縮率 <= 1:1000 → OK
   → DiskSpaceChecker.CheckSpace(destDir, extractedSize)
      ├─ 空き容量不足 → 警告ダイアログ表示（Continue/Cancel）
      └─ 空き容量十分 → OK
   → 検証通過（またはユーザーがContinue選択） → 展開処理へ

2. エラーハンドリングフロー:
   エラー発生
   → エラー種別判定
      ├─ 一時的エラー（I/Oタイムアウト等） → 最大3回リトライ
      ├─ 恒久的エラー（ファイル不存在等） → 即座に失敗
      └─ セキュリティエラー → 警告ダイアログ表示
   → クリーンアップ（部分ファイル削除）
   → エラーダイアログ表示
```

**Implementation Steps**:

1. **パストラバーサル対策**
   - アーカイブ内の全パスをチェック
   - ".."を含むパスを拒否
   - 絶対パスを検出（警告または拒否）
   - filepath.Clean()で正規化

2. **圧縮爆弾対策**
   - 展開前にメタデータコマンドで展開後サイズを取得
     - tar系: `tar -tvf` / `tar -tzvf` / `tar -tjvf` / `tar -tJvf`
     - zip: `unzip -l`
     - 7z: `7z l`
   - 出力をパースして総サイズを計算
   - 比率が1:1000を超える場合は警告ダイアログを表示（ブロックではなく確認）
   - ユーザーに継続/キャンセルの選択を促す
   - 固定の最大展開サイズ制限は設けない

3. **ディスク容量チェック**
   - 展開後サイズと反対側ペインのディスク空き容量を比較
   - 空き容量が不足する場合は警告ダイアログを表示（ブロックではなく確認）
   - ユーザーに継続/キャンセルの選択を促す

4. **シンボリンク安全性検証**
   - 絶対パスシンボリンクを検出
   - シンボリンクターゲットが展開先ディレクトリ外を指す場合に警告
   - 循環参照の検出（可能な範囲で）

5. **包括的エラーハンドリング**
   - ファイル不存在、権限エラー、ディスク容量不足等の検出
   - ユーザーフレンドリーなエラーメッセージ生成
   - リトライロジック（一時的エラーのみ）

**Dependencies**:
- Requires: Phase 1（Core Infrastructure）、Phase 2（CLI Integration）、Phase 4（Task Management）完了
- Blocks: Phase 6（E2E Testing and Polish）

**Testing Approach**:

*Unit Tests*:
- パストラバーサル検証（".."を含むパス）
- 圧縮爆弾検出（高圧縮率のテストデータ）と警告ダイアログ表示
- ディスク容量チェックと警告ダイアログ表示
- シンボリンク検証（絶対パス、相対パス）
- 入力検証（不正文字、NUL文字）

*Integration Tests*:
- 悪意のあるアーカイブの展開試行（拒否されることを確認）
- セキュリティ警告ダイアログの表示
- エラーハンドリングとクリーンアップ

*Manual Testing*:
- [ ] ".."を含むアーカイブの展開が拒否される
- [ ] 圧縮率が高すぎるアーカイブで警告が表示され、Continue/Cancelが選択できる
- [ ] ディスク容量不足時に警告が表示され、Continue/Cancelが選択できる
- [ ] 絶対パスシンボリンクで警告が表示される
- [ ] ファイル不存在エラーが適切に表示される
- [ ] 権限エラーが適切に表示される
- [ ] ディスク容量不足エラーが適切に表示される
- [ ] エラー時に部分ファイルが削除される

**Acceptance Criteria**:
- [ ] パストラバーサル攻撃が防止される
- [ ] 圧縮爆弾が検出され、警告ダイアログが表示される（ユーザーがContinue/Cancel選択可能）
- [ ] ディスク容量不足が検出され、警告ダイアログが表示される（ユーザーがContinue/Cancel選択可能）
- [ ] 危険なシンボリンクが検出される
- [ ] すべてのエラーケースで適切なメッセージが表示される
- [ ] エラー時にクリーンアップが正しく動作する
- [ ] すべてのセキュリティテストがパスする

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:
- **Risk**: セキュリティチェックが厳しすぎて正当なアーカイブが拒否される
  - **Mitigation**: 警告ダイアログで継続選択肢を提供（ユーザー判断）
- **Risk**: セキュリティチェックのパフォーマンスオーバーヘッド
  - **Mitigation**: 大規模アーカイブでのみ詳細チェック、小規模は簡易チェック

---

### Phase 6: E2E Testing and Polish (E2Eテストと仕上げ)

**Goal**: E2Eテストシナリオを実装し、実際のユーザーワークフローをテストする。パフォーマンス最適化とドキュメント整備を行い、プロダクション品質を確保する。

**Files to Create**:
- `tests/e2e/archive_test.sh` - E2Eテストスクリプト
- `doc/archive_feature.md` - 機能ドキュメント（ユーザー向け）

**Files to Modify**:
- すべてのGoファイル:
  - godocコメントの追加・改善
  - コード品質改善（linter対応）
- `README.md`:
  - アーカイブ機能の説明追加
  - 必要な外部ツールの記載

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| E2ETestSuite | 実際のユーザーワークフローをテスト | duofmが起動可能 | すべてのシナリオがパス |
| PerformanceBenchmark | パフォーマンス測定 | テストデータが準備されている | パフォーマンス目標を達成 |
| DocumentationGenerator | godocとユーザードキュメント生成 | コードが完成している | ドキュメントが整備されている |

**Processing Flow**:
```
1. E2Eテストフロー:
   各テストシナリオごとに:
   ├─ duofm起動
   ├─ テストデータ準備
   ├─ キー入力シミュレーション
   ├─ 画面出力の検証（assert_contains）
   ├─ ファイルシステム状態の検証
   └─ duofm終了

2. パフォーマンス測定:
   ベンチマークデータ（100MB）を準備
   → 圧縮時間測定（各形式、各圧縮レベル）
   → 伸長時間測定
   → メモリ使用量測定
   → UI応答時間測定
   → 目標値との比較

3. ドキュメント整備:
   godocコメント追加
   → go doc -allで確認
   → ユーザー向けドキュメント作成
   → README更新
```

**Implementation Steps**:

1. **E2Eテストスクリプト作成**
   - 既存のE2Eテスト構造に倣う
   - 各ユースケースごとにテスト関数作成
   - キー入力シミュレーション（tmux send-keys使用）
   - 画面出力検証（assert_contains, assert_not_contains）
   - ファイルシステム状態検証

2. **テストシナリオ実装**
   - シナリオ1: 単一ディレクトリの圧縮（tar.xz）
   - シナリオ2: 複数ファイルの一括圧縮（zip）
   - シナリオ3: アーカイブの伸長（tar.gz）
   - シナリオ4: スマート展開（単一ディレクトリ vs 複数ファイル）
   - シナリオ5: 上書き確認（Overwrite/Rename）
   - シナリオ6: キャンセル処理
   - シナリオ7: エラーハンドリング（権限エラー等）

3. **パフォーマンステストとチューニング**
   - 100MBデータでの圧縮・伸長時間測定
   - 目標: 圧縮 < 30秒（レベル6）、伸長 < 5秒
   - メモリ使用量測定（目標: < 100MB）
   - UI応答時間測定（目標: < 100ms）
   - ボトルネック特定と最適化

4. **ドキュメント整備**
   - すべての公開関数にgodocコメント追加
   - パッケージレベルのドキュメント
   - ユーザー向け機能説明ドキュメント
   - README更新（インストール、使用方法、依存関係）

**Dependencies**:
- Requires: Phase 1-5すべて完了
- Blocks: なし（最終フェーズ）

**Testing Approach**:

*E2E Tests*:
- 7つのテストシナリオすべてを実装
- CI/CDパイプラインに統合
- 自動テスト実行

*Performance Tests*:
- ベンチマークスイート作成
- 各形式・各レベルでの測定
- 継続的なパフォーマンス監視

*Manual Testing*:
- [ ] E2Eテストシナリオ1-7がすべてパスする
- [ ] パフォーマンス目標が達成される
- [ ] godocが正しく生成される
- [ ] ユーザードキュメントが分かりやすい
- [ ] READMEに必要な情報がすべて記載されている

**Acceptance Criteria**:
- [ ] すべてのE2Eテストがパスする
- [ ] パフォーマンス目標（圧縮 < 30秒、伸長 < 5秒、メモリ < 100MB、UI応答 < 100ms）を達成
- [ ] テストカバレッジが80%以上
- [ ] godocがすべての公開APIに存在する
- [ ] ユーザードキュメントが完成している
- [ ] README更新が完了している
- [ ] メモリリークが存在しない

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:
- **Risk**: パフォーマンス目標未達成
  - **Mitigation**: 早期にベンチマーク実施、ボトルネック特定と最適化
- **Risk**: E2Eテストが不安定（フレーキー）
  - **Mitigation**: 適切な待機時間挿入、画面状態の確実な検証

---

## Complete File Structure

```
duofm/
├── cmd/
│   └── duofm/
│       └── main.go                      # Entry point - initializes and runs app
├── internal/
│   ├── archive/                         # NEW: Archive operations package
│   │   ├── archive.go                   # Archive controller (unified API)
│   │   ├── archive_test.go
│   │   ├── command_availability.go      # External command availability check
│   │   ├── command_availability_test.go
│   │   ├── command_executor.go          # Base CLI command executor
│   │   ├── command_executor_test.go
│   │   ├── errors.go                    # Error definitions
│   │   ├── format.go                    # Format detection and constants
│   │   ├── format_test.go
│   │   ├── progress.go                  # Progress tracking structures
│   │   ├── progress_test.go
│   │   ├── security.go                  # Security validation (path traversal, etc.)
│   │   ├── security_test.go
│   │   ├── sevenzip_executor.go         # 7z command wrapper
│   │   ├── sevenzip_executor_test.go
│   │   ├── smart_extractor.go           # Smart extraction logic
│   │   ├── smart_extractor_test.go
│   │   ├── tar_executor.go              # tar/tar.gz/tar.bz2/tar.xz command wrapper
│   │   ├── tar_executor_test.go
│   │   ├── task_manager.go              # Background task management
│   │   ├── task_manager_test.go
│   │   ├── validation.go                # Input validation
│   │   ├── validation_test.go
│   │   ├── zip_executor.go              # zip/unzip command wrapper
│   │   └── zip_executor_test.go
│   ├── ui/
│   │   ├── model.go                     # MODIFIED: Add archive operation handling
│   │   ├── update.go                    # Update function (message handling)
│   │   ├── view.go                      # View rendering
│   │   ├── pane.go                      # Pane component
│   │   ├── statusbar.go                 # Status bar component
│   │   ├── context_menu_dialog.go       # MODIFIED: Add archive menu items
│   │   ├── context_menu_dialog_test.go  # MODIFIED: Add archive menu tests
│   │   ├── archive_name_dialog.go       # NEW: Archive name input dialog
│   │   ├── archive_name_dialog_test.go  # NEW
│   │   ├── archive_progress_dialog.go   # NEW: Progress display
│   │   ├── archive_progress_dialog_test.go # NEW
│   │   ├── compression_level_dialog.go  # NEW: Compression level selection
│   │   ├── compression_level_dialog_test.go # NEW
│   │   ├── dialog.go                    # Dialog interface
│   │   ├── confirm_dialog.go            # Confirmation dialog
│   │   ├── error_dialog.go              # Error dialog
│   │   ├── input_dialog.go              # Generic input dialog
│   │   ├── overwrite_dialog.go          # Overwrite confirmation
│   │   ├── help_dialog.go               # Help screen
│   │   ├── keys.go                      # Keybinding definitions
│   │   ├── messages.go                  # MODIFIED: Add archive messages
│   │   └── styles.go                    # Lip Gloss styles
│   ├── fs/
│   │   ├── operations.go                # File copy/move/delete
│   │   ├── navigation.go                # Directory traversal
│   │   ├── reader.go                    # Directory reading
│   │   └── sort.go                      # File sorting
│   └── config/
│       └── config.go                    # Configuration handling
├── tests/
│   └── e2e/
│       ├── archive_test.sh              # NEW: E2E tests for archive operations
│       └── helpers.sh                   # E2E test helpers
├── doc/
│   ├── tasks/
│   │   └── archive/
│   │       ├── SPEC.md                  # Feature specification
│   │       ├── 要件定義書.md            # Requirements document (Japanese)
│   │       └── IMPLEMENTATION.md        # This file
│   └── archive_feature.md               # NEW: User-facing documentation
├── go.mod
├── go.sum
├── Makefile
└── README.md                            # MODIFIED: Add archive feature description
```

**File Descriptions**:

**Archive Package (`internal/archive/`)**:
- `archive.go`: 圧縮・伸長操作の統合APIを提供するコントローラー
- `command_availability.go`: 外部コマンド（tar, zip, 7z）の存在確認
- `command_executor.go`: 外部CLIコマンド実行の基底実装
- `tar_executor.go`, `zip_executor.go`, `sevenzip_executor.go`: 各形式のコマンドラッパー
- `format.go`: アーカイブ形式の定義と検出ロジック
- `smart_extractor.go`: アーカイブ内容解析とスマート展開戦略決定
- `task_manager.go`: 非同期タスクの管理とライフサイクル制御
- `progress.go`: 進捗情報の構造体と計算ロジック
- `security.go`: セキュリティ検証（パストラバーサル、圧縮爆弾等）
- `validation.go`: 入力検証（ファイル名等）
- `errors.go`: エラー定義

**UI Components (`internal/ui/`)**:
- `archive_name_dialog.go`: アーカイブ名入力ダイアログ
- `archive_progress_dialog.go`: 進捗表示ダイアログ（プログレスバー、ファイル数等）
- `compression_level_dialog.go`: 圧縮レベル選択ダイアログ（0-9）
- `context_menu_dialog.go` (変更): "Compress"と"Extract archive"メニュー追加
- `messages.go` (変更): アーカイブ操作用メッセージ型追加
- `model.go` (変更): アーカイブタスク管理と進捗監視

**Tests**:
- 各パッケージに対応するテストファイル（`*_test.go`）
- `tests/e2e/archive_test.sh`: E2Eテストシナリオ

**Documentation**:
- `doc/archive_feature.md`: ユーザー向け機能説明
- `README.md` (変更): アーカイブ機能の説明と依存関係追加

## Testing Strategy

### Unit Testing

**Approach**:
- Goの標準`testing`パッケージを使用
- テーブル駆動テスト（複数シナリオを効率的にテスト）
- 外部依存（ファイルシステム、外部コマンド）のモック化

**Test Coverage Goals**:
- コアロジック: 80%以上
- セキュリティ関連コード: 90%以上（クリティカルな機能）
- UI コンポーネント: 60%以上（手動テスト補完）

**Key Test Areas**:

1. **Format Detection** (`internal/archive/format_test.go`)
   - 各拡張子の正しい形式判定
   - 非対応拡張子のエラーハンドリング
   - エッジケース（大文字小文字、複数拡張子）

2. **Command Availability** (`internal/archive/command_availability_test.go`)
   - コマンド存在確認（モック環境）
   - 利用可能な形式リストの生成
   - コマンド不在時の適切なフィルタリング

3. **Command Execution** (`internal/archive/*_executor_test.go`)
   - コマンドライン引数の正しい構築
   - 圧縮レベル指定の適用
   - 標準出力/エラー出力のキャプチャ
   - エラーハンドリング（コマンド失敗、不正な出力等）

4. **Smart Extraction** (`internal/archive/smart_extractor_test.go`)
   - 単一ディレクトリアーカイブの判定
   - 複数ファイルアーカイブの判定
   - 展開戦略の正しい決定

5. **Security Validation** (`internal/archive/security_test.go`)
   - パストラバーサル検出（".."を含むパス）
   - 圧縮爆弾検出（高圧縮率データ）と警告ダイアログ表示
   - ディスク容量チェックと警告ダイアログ表示
   - シンボリンク安全性検証（絶対パス、ディレクトリ外参照）

6. **Task Management** (`internal/archive/task_manager_test.go`)
   - タスク開始と終了
   - 進捗更新の正確性
   - キャンセル処理
   - エラーハンドリング

7. **UI Dialogs** (`internal/ui/*_dialog_test.go`)
   - キーボード入力処理
   - 状態遷移
   - バリデーション
   - 表示内容の正確性

### Integration Testing

**Scenarios**:

| シナリオ | テスト内容 | 期待結果 |
|---------|----------|---------|
| End-to-end Compression | 実際のファイルを各形式で圧縮 | アーカイブが正しく生成される |
| End-to-end Extraction | 各形式のアーカイブを伸長 | ファイルが正しく復元される |
| Attribute Preservation | 権限、タイムスタンプ、シンボリンクの保持 | すべての属性が保持される |
| Smart Extraction | 単一/複数ファイルアーカイブの展開 | 正しい展開戦略が適用される |
| Background Processing | 非同期処理とUI応答性 | UI応答 < 100ms |
| Cancellation | 処理中のキャンセル | クリーンアップが正しく動作 |

**Approach**:
- 一時ディレクトリでの実ファイル操作
- 実際の外部コマンド使用（tar, zip, 7z）
- パフォーマンス測定含む

### Manual Testing Checklist

**圧縮操作**:
- [ ] 単一ファイルをtar形式で圧縮できる
- [ ] 単一ディレクトリをtar.xz形式で圧縮できる
- [ ] 複数ファイルをマーク選択してzip形式で圧縮できる
- [ ] 圧縮レベル0-9すべてで圧縮できる
- [ ] デフォルトのアーカイブ名が適切に生成される
- [ ] アーカイブ名を編集できる
- [ ] 同名ファイル存在時に上書き確認ダイアログが表示される
- [ ] 上書き/リネーム/キャンセルがすべて正しく動作する
- [ ] 進捗ダイアログが表示され、リアルタイムで更新される
- [ ] 圧縮中にEscキーでキャンセルできる
- [ ] キャンセル時に部分ファイルが削除される
- [ ] 完了後にステータス通知が表示される
- [ ] 反対側ペインにアーカイブが生成される

**伸長操作**:
- [ ] tar形式のアーカイブを伸長できる
- [ ] tar.gz, tar.bz2, tar.xz形式のアーカイブを伸長できる
- [ ] zip形式のアーカイブを伸長できる
- [ ] 7z形式のアーカイブを伸長できる
- [ ] 単一ディレクトリアーカイブが直接展開される
- [ ] 複数ファイルアーカイブが新ディレクトリに展開される
- [ ] ファイル権限が保持される
- [ ] タイムスタンプが保持される
- [ ] シンボリックリンクがリンクとして復元される
- [ ] 進捗ダイアログが表示され、リアルタイムで更新される
- [ ] 伸長中にEscキーでキャンセルできる
- [ ] 完了後にステータス通知が表示される
- [ ] 反対側ペインに展開されたファイルが表示される

**エラーハンドリング**:
- [ ] 読み取り権限のないファイルで適切なエラーが表示される
- [ ] 書き込み権限のないディレクトリで適切なエラーが表示される
- [ ] ディスク容量不足で適切なエラーが表示される
- [ ] 破損したアーカイブで適切なエラーが表示される
- [ ] 非対応形式で適切なエラーが表示される
- [ ] エラー時に部分ファイルが削除される

**セキュリティ**:
- [ ] ".."を含むアーカイブの展開が拒否される
- [ ] 圧縮率が高すぎるアーカイブで警告が表示され、Continue/Cancelが選択できる
- [ ] ディスク容量不足時に警告が表示され、Continue/Cancelが選択できる
- [ ] 絶対パスシンボリンクで警告が表示される

**パフォーマンス**:
- [ ] 小ファイル（< 10MB）の圧縮が3秒以内に完了する
- [ ] 100MBファイルの圧縮が30秒以内に完了する（レベル6）
- [ ] 100MBアーカイブの伸長が5秒以内に完了する
- [ ] 操作中もUI応答が100ms以内に維持される

**UI統合**:
- [ ] コンテキストメニューに"Compress"が表示される
- [ ] "Compress"選択でサブメニューが表示される
- [ ] 利用可能な形式のみがサブメニューに表示される
- [ ] アーカイブファイル選択時に"Extract archive"が表示される
- [ ] マーク選択時に"Compress N files"と表示される
- [ ] すべてのダイアログが正しく表示される
- [ ] ダイアログ間の遷移が正しく動作する

### E2E Tests

E2Eテストは`tests/e2e/archive_test.sh`に実装します。

**テストシナリオ**:

1. **Compress Single Directory**
   - ディレクトリを選択してtar.xz形式で圧縮
   - デフォルト名を確認して確定
   - 反対側ペインにアーカイブが生成されることを確認

2. **Compress Multiple Files**
   - 複数ファイルをマーク選択
   - zip形式で圧縮
   - アーカイブ名を編集
   - 反対側ペインにアーカイブが生成されることを確認

3. **Extract Archive**
   - アーカイブファイルを選択
   - "Extract archive"を実行
   - 反対側ペインにファイルが展開されることを確認

4. **Smart Extraction**
   - 単一ディレクトリアーカイブ → 直接展開
   - 複数ファイルアーカイブ → 新ディレクトリに展開

5. **Overwrite Handling**
   - 同名アーカイブを再作成
   - 上書き確認ダイアログが表示されることを確認
   - "Rename"を選択して新しい名前で作成

6. **Cancel Operation**
   - 圧縮開始
   - 進捗ダイアログ表示中にEsc
   - キャンセル通知が表示され、部分ファイルが削除されることを確認

7. **Error Handling**
   - 読み取り権限のないファイルで圧縮試行
   - エラーダイアログが表示されることを確認

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation (Debian/Ubuntu) |
|---------|---------|---------|------------------------------|
| tar | - | tar archive creation/extraction | `sudo apt install tar` (通常はプリインストール済み) |
| gzip | - | gzip compression | `sudo apt install gzip` (通常はプリインストール済み) |
| bzip2 | - | bzip2 compression | `sudo apt install bzip2` (通常はプリインストール済み) |
| xz | - | LZMA compression | `sudo apt install xz-utils` (通常はプリインストール済み) |
| zip | - | zip compression | `sudo apt install zip` (オプション) |
| unzip | - | zip extraction | `sudo apt install unzip` (オプション) |
| 7z | - | 7z handling | `sudo apt install p7zip-full` (オプション) |

**Note**: tar, gzip, bzip2, xzは通常のLinuxシステムではプリインストールされています。zip/unzip/7zはオプションで、インストールされていない場合は該当形式がメニューから非表示になります。

### Internal Dependencies

**Implementation Order** (依存関係を考慮):
1. Phase 1 (Core Infrastructure) - 依存なし
2. Phase 2 (CLI Integration) - Phase 1に依存
3. Phase 3 (UI Integration) - Phase 1, 2に依存
4. Phase 4 (Task Management) - Phase 1, 2, 3に依存
5. Phase 5 (Security and Error Handling) - Phase 1, 2, 4に依存
6. Phase 6 (E2E Testing and Polish) - Phase 1-5すべてに依存

**Component Dependencies**:
- `internal/archive/tar_executor.go` → `internal/archive/command_executor.go`
- `internal/archive/smart_extractor.go` → `internal/archive/tar_executor.go`, `internal/archive/zip_executor.go`
- `internal/archive/task_manager.go` → `internal/archive/archive.go`
- `internal/ui/archive_progress_dialog.go` → `internal/archive/progress.go`
- `internal/ui/context_menu_dialog.go` → `internal/archive/command_availability.go`
- `internal/ui/model.go` → `internal/archive/archive.go`, `internal/archive/task_manager.go`

## Risk Assessment

### Technical Risks

1. **外部コマンドのバージョン差異**
   - **Risk**: tarやzipのバージョンによってオプションや出力形式が異なる
   - **Likelihood**: 低（広くサポートされているオプションのみ使用）
   - **Impact**: 中（一部環境で動作しない可能性）
   - **Mitigation**: 汎用的なオプションのみ使用、複数バージョンでテスト

2. **進捗情報の解析失敗**
   - **Risk**: コマンド出力形式の違いで進捗が取得できない
   - **Likelihood**: 中
   - **Impact**: 低（進捗表示なしでも操作自体は成功）
   - **Mitigation**: 進捗解析失敗を許容、操作自体は継続

3. **大規模ファイルでのパフォーマンス低下**
   - **Risk**: 数GB以上のファイルで処理が遅くなる
   - **Likelihood**: 高
   - **Impact**: 中（ユーザー体験の低下）
   - **Mitigation**: 進捗表示による透明性確保、キャンセル可能性

4. **外部コマンド不在**
   - **Risk**: zip/7zがインストールされていない環境がある
   - **Likelihood**: 高（オプションコマンド）
   - **Impact**: 低（該当形式が非表示になるだけ）
   - **Mitigation**: コマンド存在確認、非表示化で対応

### Implementation Risks

1. **goroutineリークによるメモリリーク**
   - **Risk**: キャンセル処理の不備でgoroutineが残る
   - **Likelihood**: 中
   - **Impact**: 高（長時間使用でメモリ消費増加）
   - **Mitigation**: contextによる確実なキャンセル、defer文での後処理、メモリリークテスト

2. **ファイルディスクリプタリーク**
   - **Risk**: ファイルやパイプのクローズ漏れ
   - **Likelihood**: 中
   - **Impact**: 中（システムリソース枯渇）
   - **Mitigation**: defer文での確実なClose、リソースリークテスト

3. **セキュリティチェックのバイパス**
   - **Risk**: 巧妙なアーカイブでセキュリティ検証を回避
   - **Likelihood**: 低
   - **Impact**: 高（セキュリティ侵害）
   - **Mitigation**: 複数段階の検証、保守的な判定、セキュリティテスト

4. **UI状態の不整合**
   - **Risk**: 非同期処理とUI状態の同期不良
   - **Likelihood**: 中
   - **Impact**: 中（操作不能、誤動作）
   - **Mitigation**: Bubble Teaのメッセージングパターン遵守、状態遷移テスト

## Performance Considerations

### Performance Goals

| 指標 | 目標値 | 測定方法 |
|-----|-------|---------|
| 小ファイル圧縮 (< 10MB) | < 3秒 | ベンチマークテスト |
| 大ファイル圧縮 (100MB, レベル6) | < 30秒 | ベンチマークテスト |
| アーカイブ伸長 (100MB) | < 5秒 | ベンチマークテスト |
| UI応答時間 | < 100ms | 処理中のキー入力応答測定 |
| 進捗更新頻度 | 最大10Hz | 更新間隔測定 |
| メモリ使用量 | < 100MB | 処理中のメモリプロファイリング |

### Optimization Strategies

1. **ストリーミングI/O**
   - 外部コマンドの標準出力/入力をストリーミング処理
   - 大ファイルを全メモリにロードしない
   - `io.Copy`と`bufio`の使用

2. **進捗更新の最適化**
   - 更新頻度を100ms間隔に制限（10Hz）
   - 小ファイル（< 1MB）では進捗更新をスキップ
   - バッチ更新（複数ファイルの進捗をまとめて通知）

3. **非同期処理**
   - goroutineでのバックグラウンド実行
   - channelによる進捗通信（バッファリング）
   - UI更新とファイル処理の分離

4. **メモリ管理**
   - 外部コマンド使用によりGo側のメモリ使用を最小化
   - 標準出力/エラー出力のバッファサイズ制限
   - 不要なデータの早期解放

5. **ディスク I/O最適化**
   - 外部コマンドに委譲（ネイティブ最適化）
   - 一時ファイルの使用最小化
   - アトミック操作（一時ファイル→リネーム）

### Profiling Points

- 総処理時間（圧縮・伸長）
- goroutine数とチャンネルバッファ使用量
- メモリアロケーション（heap profile）
- UI描画時間（render loop）
- 外部コマンド起動オーバーヘッド

## Security Considerations

### Path Traversal Prevention

- アーカイブ内の全パスを展開前に検証
- ".."を含むパスを拒否
- 絶対パスを検出し警告または拒否
- `filepath.Clean()`で正規化後、展開先ディレクトリ内であることを確認

### Compression Bomb Protection

- 展開前にメタデータコマンドで展開後サイズを取得
  - tar系: `tar -tvf` / `tar -tzvf` / `tar -tjvf` / `tar -tJvf`
  - zip: `unzip -l`
  - 7z: `7z l`
- 出力をパースして総サイズを計算
- アーカイブサイズと展開後サイズの比率を計算
- 比率が1:1000を超える場合に警告ダイアログを表示（ブロックではなく確認）
- ユーザーに継続/キャンセルの選択を促す
- 固定の最大展開サイズ制限は設けない
- 展開中の累積サイズ監視

**警告ダイアログUI（圧縮爆弾検出）**:
```
Warning: Large extraction ratio detected

Archive size: 1 MB
Extracted size: 2 GB (ratio: 1:2000)

This may indicate a zip bomb or highly compressed data.
Do you want to continue?

[Continue] [Cancel]
```

### Disk Space Protection

- 展開前に反対側ペインのディスク空き容量を確認
- 展開後サイズが空き容量を超える場合に警告ダイアログを表示（ブロックではなく確認）
- ユーザーに継続/キャンセルの選択を促す

**警告ダイアログUI（ディスク容量不足）**:
```
Warning: Insufficient disk space

Required: 1.2 GB
Available: 500 MB

Do you want to continue anyway?

[Continue] [Cancel]
```

### Symlink Safety

- シンボリックリンクをリンクとして保存（実体をたどらない）
- 展開時に絶対パスシンボリンクを検出し警告
- シンボリンクターゲットが展開先ディレクトリ外を指す場合に警告
- 循環参照の検出（可能な範囲で）

### Permission Handling

- 展開時にsetuid/setgidビットを無視
- ユーザーのumaskを適用
- world-writableファイルを作成しない（最大0666 & ~umask）

### Input Validation

- ファイル名の不正文字チェック（NUL、制御文字）
- パス長制限の確認（PATH_MAX）
- 圧縮レベルの範囲チェック（0-9）
- すべてのユーザー入力をサニタイズ

## Error Handling

### Error Categories

1. **入力検証エラー**
   - ファイル名が空
   - 無効な文字を含む
   - サポート外の形式
   - **対応**: エラーダイアログ表示、操作中止

2. **ファイルシステムエラー**
   - ファイル不存在
   - 読み取り権限なし
   - 書き込み権限なし
   - ディスク容量不足
   - **対応**: 一時的エラーはリトライ（最大3回）、恒久的エラーは即座に失敗

3. **アーカイブ処理エラー**
   - 破損したアーカイブ
   - 圧縮/伸長失敗
   - 外部コマンドエラー
   - **対応**: エラーダイアログ表示、部分ファイル削除

4. **セキュリティエラー**
   - パストラバーサル検出
   - 圧縮爆弾検出
   - 危険なシンボリンク検出
   - **対応**: 警告ダイアログ表示、ユーザー選択（継続/中止）

5. **システムエラー**
   - メモリ不足
   - I/Oエラー
   - 予期しない内部エラー
   - **対応**: エラーログ記録、エラーダイアログ表示、クリーンアップ

### Error Messages

**ユーザーフレンドリーなメッセージ**:
- 明確で具体的（何が問題か）
- 実行可能（どうすればよいか）
- 専門用語を避ける

**例**:
- ✅ "Cannot compress: Source file not found"
- ❌ "ENOENT: no such file or directory"

### Cleanup Procedures

**エラー時**:
1. 進行中の操作を停止
2. 開いているファイルハンドルをクローズ
3. 部分的に作成されたアーカイブファイルを削除
4. 部分的に展開されたファイルを削除
5. エラー詳細をログに記録
6. ユーザーにエラーダイアログを表示

**キャンセル時**:
1. キャンセルフラグを設定
2. context.Cancel()呼び出し
3. 外部コマンドの終了を待機
4. エラー時と同様のクリーンアップ
5. キャンセル確認通知を表示

## Success Criteria

### Functional Completeness
- [ ] tar, tar.gz, tar.bz2, tar.xz, zip, 7z形式で圧縮できる
- [ ] tar, tar.gz, tar.bz2, tar.xz, zip, 7z形式を伸長できる
- [ ] 単一ファイル/ディレクトリの圧縮が可能
- [ ] 複数ファイル（マーク選択）の一括圧縮が可能
- [ ] スマート展開が正しく動作する
- [ ] 圧縮レベル（0-9）を選択できる
- [ ] アーカイブ名を編集できる
- [ ] 進捗ダイアログが表示され、リアルタイムで更新される
- [ ] 操作をキャンセルできる
- [ ] 同名ファイル存在時に上書き確認ダイアログが表示される
- [ ] すべてのエラーケースで適切なメッセージが表示される
- [ ] ファイル属性（権限、タイムスタンプ、シンボリックリンク）が保持される

### Quality Metrics
- [ ] テストカバレッジが80%以上
- [ ] すべてのユニットテストがパスする
- [ ] すべてのE2Eテストがパスする
- [ ] 重大なバグが存在しない
- [ ] コードがGo標準スタイルに準拠している
- [ ] すべての公開APIにgodocコメントが存在する

### Performance Metrics
- [ ] 小ファイル圧縮（< 10MB）が3秒以内に完了
- [ ] 大ファイル圧縮（100MB、レベル6）が30秒以内に完了
- [ ] アーカイブ伸長（100MB）が5秒以内に完了
- [ ] UI応答時間が100ms以内
- [ ] メモリ使用量が100MB以下
- [ ] メモリリークが存在しない

### Security Metrics
- [ ] パストラバーサル攻撃が防止される
- [ ] 圧縮爆弾が検出され、警告ダイアログが表示される
- [ ] ディスク容量不足が検出され、警告ダイアログが表示される
- [ ] 危険なシンボリンクが検出される
- [ ] setuid/setgidビットが無視される
- [ ] すべてのセキュリティテストがパスする

### User Experience
- [ ] コンテキストメニューから直感的に操作できる
- [ ] 利用可能な形式のみが表示される
- [ ] デフォルト値が適切に設定されている
- [ ] エラーメッセージが分かりやすい
- [ ] 進捗が視覚的に分かりやすい
- [ ] キャンセルが確実に動作する

## References

- **Specification**: `doc/tasks/archive/SPEC.md`
- **Requirements Document**: `doc/tasks/archive/要件定義書.md`
- **Existing Code**:
  - `internal/ui/context_menu_dialog.go` - コンテキストメニューの実装
  - `internal/fs/operations.go` - ファイル操作の実装
  - `internal/ui/input_dialog.go` - 入力ダイアログの実装
  - `internal/ui/overwrite_dialog.go` - 上書き確認ダイアログの実装
  - `internal/ui/model.go` - Bubble Teaモデルの実装
- **External Documentation**:
  - [GNU tar Manual](https://www.gnu.org/software/tar/manual/)
  - [gzip Manual](https://www.gnu.org/software/gzip/manual/)
  - [bzip2 Manual](https://sourceware.org/bzip2/)
  - [XZ Utils](https://tukaani.org/xz/)
  - [Info-ZIP](https://infozip.sourceforge.net/)
  - [p7zip Project](https://github.com/p7zip-project/p7zip)
  - [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
  - [Lip Gloss Documentation](https://github.com/charmbracelet/lipgloss)
  - [Go Context Package](https://pkg.go.dev/context)
  - [Go os/exec Package](https://pkg.go.dev/os/exec)

## Next Steps

実装計画のレビュー後、以下の手順で進めてください:

1. **Review and Approval**
   - 実装計画を確認し、不明点を解決する
   - フェーズ分割とスコープの妥当性を検証する
   - Open Questionsに対する決定を行う

2. **Environment Setup**
   - Go 1.21+がインストールされていることを確認
   - 必要な外部コマンド（tar, gzip, bzip2, xz）を確認
   - オプションコマンド（zip, unzip, 7z）をインストール（オプション）

3. **Begin Implementation**
   - Phase 1（Core Infrastructure）から開始
   - テスト駆動開発（TDD）アプローチを採用（テスト先行）
   - 各コンポーネントを小さなコミットで段階的に実装

4. **Continuous Testing**
   - 各フェーズ完了後にテストを実行
   - パフォーマンスベンチマークを定期的に測定
   - E2Eテストで統合動作を検証

5. **Documentation**
   - 実装と並行してgodocコメントを記述
   - ユーザー向けドキュメントを作成
   - README更新

6. **Code Review and Refinement**
   - コードレビューを実施
   - linter（golangci-lint等）でコード品質をチェック
   - リファクタリングと最適化

7. **Final Testing and Release**
   - すべてのテストがパスすることを確認
   - パフォーマンス目標達成を確認
   - リリースノート作成
   - マージとリリース

## Open Questions

### From Specification:
- [ ] パスワード保護付きアーカイブを将来的にサポートするか？（現在のスコープ外）
- [ ] デフォルトの圧縮形式/レベルを設定ファイルで指定可能にするか？
- [ ] 非常に大きなアーカイブ（> 10GB）で追加の確認ダイアログを表示するか？

### Implementation-Specific:
- [ ] 進捗情報の解析が失敗した場合、進捗表示なしで処理を継続するか、エラーとして扱うか？
  - **提案**: 進捗表示なしで処理継続（操作自体は成功させる）
- [x] 圧縮爆弾の閾値（1:1000）は適切か？設定可能にするか？
  - **決定**: 1:1000を固定値として使用、警告のみ（ブロックではなく確認）
- [x] 最大展開サイズ制限は必要か？
  - **決定**: 固定の上限は設けない（ディスク容量チェックで対応）
- [x] アーカイブ操作のログをどこに保存するか？
  - **決定**: ログファイル出力は削除（標準出力のみ）

### To Clarify with User:
- [ ] zip/7zコマンドが存在しない環境でのユーザー体験をどうするか？（メニューから非表示で十分か、インストールガイドを表示するか？）
- [ ] 同時に複数のアーカイブタスクを実行できるようにするか？（現在は1つのみ）

---

**Last Updated**: 2026-01-02
**Status**: Draft
**Version**: 1.0
