# Feature: MIME Type Display in Status Bar

## Overview

ステータスバーのカーソル位置表示（N/M）の右側に、カーソル下のファイルのMIMEタイプを角括弧付きで表示する。ディレクトリとシンボリックリンクには専用のラベルと色を使用する。

## Domain Rules

- MIMEタイプは拡張子ベース（`mime.TypeByExtension`）で判定する。
- ディレクトリは `[Directory]` と表示し、ディレクトリ色（`DirectoryFg`）を適用する。
- シンボリックリンクは `[SymbolicLink]` と表示し、シンボリックリンク色（`SymlinkFg`）を適用する。
- 通常ファイルは `[text/html]` のようにMIMEタイプを表示し、ステータスバー前景色（`StatusFg`）を適用する。
- 表示されるMIMEタイプは `[enter_behavior_mime]` セクションのキーとしてそのまま使用できる値とする。

## Objectives

- カーソル下のエントリのファイル種別/MIMEタイプをステータスバーに常時表示する。
- ディレクトリとシンボリックリンクを色で視覚的に区別する。

## User Stories

### US1: ファイルのMIMEタイプ確認
ユーザーとして、カーソルを合わせたファイルのMIMEタイプをステータスバーで確認したい。設定ファイルの `[enter_behavior_mime]` セクションにルールを追加する際にそのまま使える値を確認できるようにするため。

**Acceptance Criteria:**
- [ ] ステータスバーのN/M表示の右側にMIMEタイプが表示される
- [ ] 表示形式は `[text/html]` のような角括弧付き
- [ ] 拡張子なしファイルは `[application/octet-stream]` と表示される

### US2: ディレクトリの種別表示
ユーザーとして、カーソルがディレクトリ上にあるとき、ステータスバーで `[Directory]` と青文字で表示されることで、ファイルとディレクトリを視覚的に区別したい。

**Acceptance Criteria:**
- [ ] ディレクトリは `[Directory]` と表示される
- [ ] テーマの `DirectoryFg` 色で表示される

### US3: シンボリックリンクの種別表示
ユーザーとして、カーソルがシンボリックリンク上にあるとき、ステータスバーで `[SymbolicLink]` とシアン文字で表示されることで、通常ファイルと区別したい。

**Acceptance Criteria:**
- [ ] シンボリックリンクは `[SymbolicLink]` と表示される
- [ ] テーマの `SymlinkFg` 色で表示される

## Functional Requirements

- **FR1:** ステータスバーのN/M位置表示の右側にMIMEタイプ情報を表示する。
- **FR2:** 表示形式は `[{type}]` とする（角括弧で囲む）。
- **FR3:** `FileEntry.IsDir` が `true` の場合、`[Directory]` と表示する。
- **FR4:** `FileEntry.IsSymlink` が `true` の場合、`[SymbolicLink]` と表示する。
- **FR5:** 通常ファイルの場合、`config.GetMIMEType(entry.Name)` の戻り値を `[{mimeType}]` として表示する。
- **FR6:** ディレクトリの表示色はテーマの `DirectoryFg` を使用する。
- **FR7:** シンボリックリンクの表示色はテーマの `SymlinkFg` を使用する。
- **FR8:** 通常ファイルの表示色はテーマの `StatusFg` を使用する。
- **FR9:** 親ディレクトリエントリ（`..`）は `[Directory]` と表示する。
- **FR10:** ステータスメッセージ表示中はMIMEタイプ表示も非表示となる（既存の動作を維持）。
- **FR11:** エントリが0件の空ディレクトリではMIMEタイプ表示を省略する。

## Non-Functional Requirements

- **NFR1 - Performance:** MIMEタイプの判定は `mime.TypeByExtension` による拡張子ベースのため、レンダリング遅延は発生しない。
- **NFR2 - Layout:** MIMEタイプ表示の追加により、ステータスバーの中央パディングが減少するが、レイアウトは破綻しない。狭いターミナル幅で収まらない場合はMIMEタイプ表示を省略する。

## Interface Contract

### Input/Output Specification

**Input:**
- `fs.FileEntry` - カーソル位置のエントリ
  - `IsDir`: ディレクトリ判定
  - `IsSymlink`: シンボリックリンク判定
  - `Name`: MIMEタイプ判定用ファイル名

**Output:**
- ステータスバー上の文字列（例: `1/15 [text/html]`）
- エントリ種別に応じた色付き表示

### Status Bar Layout

```
| {N/M} {[MIMEType]}          {hints} |
```

例:
```
| 3/15 [text/html]         ?:help q:quit |
| 1/5  [Directory]         ?:help q:quit |
| 2/10 [SymbolicLink]   i:info ?:help q:quit |
```

## Implementation Details

### renderStatusBar の変更

`internal/ui/model_view.go` の `renderStatusBar()` メソッドを修正する。

1. アクティブペインのカーソル位置エントリを取得
2. エントリの種別に応じてラベルと色を決定:
   - `entry.IsDir` → ラベル: `Directory`, 色: `theme.DirectoryFg`
   - `entry.IsSymlink` → ラベル: `SymbolicLink`, 色: `theme.SymlinkFg`
   - それ以外 → ラベル: `config.GetMIMEType(entry.Name)`, 色: `theme.StatusFg`
3. ラベルを `[{label}]` 形式にフォーマット
4. lipgloss スタイルで色を適用してレンダリング
5. N/M表示の右側に半角スペース1つを挟んで配置
6. パディング計算にMIMEタイプ表示の幅を加算

### 判定の優先順位

シンボリックリンク先がディレクトリの場合の扱い:
- `IsSymlink` を `IsDir` より優先して判定する（シンボリックリンクであることを明示）

## Test Scenarios

### Unit Tests
- [ ] ディレクトリエントリで `[Directory]` が返される
- [ ] シンボリックリンクエントリで `[SymbolicLink]` が返される
- [ ] `.html` ファイルで `[text/html]` が返される
- [ ] `.txt` ファイルで `[text/plain]` が返される
- [ ] 拡張子なしファイルで `[application/octet-stream]` が返される
- [ ] 親ディレクトリ（`..`）で `[Directory]` が返される
- [ ] シンボリックリンク先がディレクトリでも `[SymbolicLink]` が返される
- [ ] 空ディレクトリ（エントリ0件）ではMIMEタイプ表示が省略される

### Integration Tests
- [ ] ステータスバーにMIMEタイプ表示が含まれる
- [ ] ディレクトリにカーソルを合わせると `DirectoryFg` 色で表示される
- [ ] シンボリックリンクにカーソルを合わせると `SymlinkFg` 色で表示される
- [ ] 通常ファイルにカーソルを合わせると `StatusFg` 色で表示される
- [ ] ステータスメッセージ表示中はMIMEタイプ表示が非表示

### E2E Tests
- [ ] ファイルにカーソルを合わせたとき、ステータスバーにMIMEタイプが表示される
- [ ] ディレクトリにカーソルを合わせたとき、ステータスバーに `[Directory]` が表示される

## Success Criteria

- [ ] ステータスバーにMIMEタイプ/種別が正しく表示される
- [ ] ディレクトリは `DirectoryFg` 色で `[Directory]` と表示される
- [ ] シンボリックリンクは `SymlinkFg` 色で `[SymbolicLink]` と表示される
- [ ] 通常ファイルは `StatusFg` 色でMIMEタイプが表示される
- [ ] 既存のステータスバー機能（メッセージ表示、キーヒント）に影響がない
- [ ] 全ユニットテストが通過する
- [ ] 狭いターミナル幅でレイアウトが破綻しない

## Dependencies

**Internal:**
- `internal/ui/model_view.go` - ステータスバーレンダリング
- `internal/config/mime.go` - `GetMIMEType()` 関数
- `internal/ui/theme.go` - `DirectoryFg`, `SymlinkFg`, `StatusFg` 色
- `internal/fs/types.go` - `FileEntry` 構造体

**External:**
- `mime` (Go標準ライブラリ) - 拡張子からMIMEタイプ判定（既存利用）

## Constraints

- Linux only（既存プロジェクトスコープと一致）。
- MIMEタイプの精度は Go 標準ライブラリの `mime.TypeByExtension` に依存する。

## Open Questions

None.
