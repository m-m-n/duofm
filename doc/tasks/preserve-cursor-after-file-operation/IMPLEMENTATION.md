# 実装計画: ファイル操作後のカーソル位置保持

## 概要

ファイル操作（移動・コピー・リネーム・バッチ操作）の完了後にカーソル位置を適切に保持・計算する。先行修正で全操作箇所の `LoadDirectory()` は `RefreshDirectoryPreserveCursor()` に置き換え済みであり、本実装では残る2点（フォールバック改善とバッチ操作用カーソル計算）を対応する。

## 目的

- ソースペイン（ファイルが減る側）で、ファイル名マッチ失敗時にインデックスベースのフォールバックを適用する
- バッチ操作後のソースペインで、マーク情報を考慮したカーソルターゲットを決定する

## 前提条件

### 開発環境
- Go 1.21+
- make

### 内部依存
- `internal/ui/pane_filter.go`: `RefreshDirectoryPreserveCursor` の既存実装
- `internal/ui/pane.go`: `Pane` 構造体、`calculateCursorAfterDeletion`（設計パターンの参考）、`SetCursor`、`EnsureCursorVisible`
- `internal/ui/model_update.go`: `batchCompleteMsg` / `batchCancelledMsg` ハンドラ
- `internal/ui/pane_marks.go`: `ClearMarks`、`GetMarkedFiles` 等のマーク操作

### 外部依存
なし（新規ライブラリ追加不要）

## アーキテクチャ概要

### 設計方針

2つの独立した変更を、それぞれテストファーストで実装する。既存の `calculateCursorAfterDeletion` と同様のパターン（純粋な計算メソッド + テーブル駆動テスト）を踏襲する。

### コンポーネント間の関係

```
batchCompleteMsg / batchCancelledMsg ハンドラ (model_update.go)
  |
  +--> calculateCursorTargetAfterBatchMove(markedFiles) -> ファイル名
  |
  +--> ClearMarks()
  |
  +--> RefreshDirectoryPreserveCursor()  (pane_filter.go)
  |      |
  |      +--> ファイル名マッチ -> 成功: そのインデックス
  |      |
  |      +--> 失敗: 旧インデックスを clamp (改善点)
  |
  +--> カーソルターゲットのファイル名でカーソル位置を調整
```

## 実装フェーズ

### フェーズ 1: RefreshDirectoryPreserveCursor のフォールバック改善

**目標**: ファイル名マッチ失敗時に、カーソルを 0 ではなく旧インデックス位置にクランプする

**変更ファイル**:
- `internal/ui/pane_filter.go` - フォールバックロジックの変更
- `internal/ui/pane_filter_test.go` - フォールバックのテスト追加

**主要コンポーネント**:

| コンポーネント | 責務 | 事前条件 | 事後条件 |
|------------|------|---------|---------|
| RefreshDirectoryPreserveCursor | ディレクトリ再読込とカーソル位置復元 | ペインが有効なディレクトリを参照 | カーソルがファイル名マッチまたはインデックスフォールバックで設定される |

**処理フロー**:
```
1. 現在の選択ファイル名と現在のカーソルインデックスを保存
2. ディレクトリを再読込（ソート・フィルタ適用）
3. 新エントリで選択ファイル名を検索
   |- 見つかった -> そのインデックスをカーソルに設定
   |- 見つからない -> 旧インデックスをエントリ数の範囲内にクランプしてカーソルに設定
4. スクロール位置を調整
```

**フォールバックのクランプ動作**:
- 旧インデックスがエントリ数未満 -> 旧インデックスをそのまま使用
- 旧インデックスがエントリ数以上 -> `max(0, len(entries)-1)` にクランプ

**実装ステップ**:

1. **フォールバックテストを追加**
   - `pane_filter_test.go` にテーブル駆動テストを追加
   - テスト対象: ファイル名マッチ失敗時のインデックスフォールバック（範囲内、超過、空ディレクトリ）
   - テスト方法: 一時ディレクトリを使い、ファイルを削除してから `RefreshDirectoryPreserveCursor` を呼び出し、カーソル位置を検証

2. **既存テストの更新**
   - `pane_test.go` の `TestRefreshDirectoryPreserveCursor` 内「resets cursor to 0 when file deleted」テストケースを更新
   - 変更前: ファイル削除後に cursor=0 を期待
   - 変更後: ファイル削除後に旧インデックスのクランプ値を期待（例: cursor=2 で ccc.txt 削除後、entries=[.., aaa.txt, bbb.txt] なら cursor=1）
   - `TestRefreshDirectoryPreserveCursorWithEmpty` のテストは cursor=0 のまま（`..` のみ残る場合はクランプで 0 になるため変更不要）

3. **フォールバックロジックを変更**
   - `RefreshDirectoryPreserveCursor` 内で旧カーソルインデックスを保存
   - ファイル名検索失敗時のデフォルト値を 0 から旧インデックスのクランプ値に変更

**依存関係**:
- 前提: なし
- 後続: フェーズ 2 で `RefreshDirectoryPreserveCursor` の改善されたフォールバックを活用

**テスト方針**:

*ユニットテスト（pane_filter_test.go に追加）*:

| ID | シナリオ | 初期状態 | 操作 | 期待結果 |
|----|---------|---------|------|---------|
| F-1 | ファイル名マッチ成功 | `[.., a, b, c]` cursor=2(b) | b が存在する状態で再読込 | cursor=2(b) |
| F-2 | ファイル名マッチ失敗・インデックス有効 | `[.., a, b, c]` cursor=2(b) | b を削除して再読込 | cursor=2(c) |
| F-3 | ファイル名マッチ失敗・インデックス超過 | `[.., a, b]` cursor=2(b) | b を削除して再読込 | cursor=1(a) |
| F-4 | 全ファイル削除 | `[.., a]` cursor=1(a) | a を削除して再読込 | cursor=0(..) |

**受入基準**:
- [ ] ファイル名マッチ成功時は従来通りそのインデックスにカーソルが移動する
- [ ] ファイル名マッチ失敗時、旧インデックスが範囲内ならそのインデックスを保持する
- [ ] 旧インデックスが範囲外なら最後のエントリにクランプする
- [ ] エントリが `..` のみの場合は cursor=0 になる
- [ ] 既存テストが全て合格する

**推定工数**: 小

---

### フェーズ 2: バッチ操作用カーソルターゲット計算

**目標**: バッチ移動後のソースペインで、マークされていない最近接ファイルにカーソルを移動する

**作成ファイル**:
- `internal/ui/pane_cursor_test.go` - `calculateCursorTargetAfterBatchMove` のテスト

**変更ファイル**:
- `internal/ui/pane.go` - `calculateCursorTargetAfterBatchMove` メソッドの追加
- `internal/ui/model_update.go` - `batchCompleteMsg` / `batchCancelledMsg` ハンドラでの呼び出し

**主要コンポーネント**:

| コンポーネント | 責務 | 事前条件 | 事後条件 |
|------------|------|---------|---------|
| calculateCursorTargetAfterBatchMove | バッチ操作後のカーソルターゲットファイル名を決定 | entries とマーク情報が有効 | マークされていない最近接ファイルの名前、または空文字を返す |
| batchCompleteMsg ハンドラ | バッチ操作完了時の状態更新 | バッチ操作が正常完了 | カーソルが適切なファイルに位置する |
| batchCancelledMsg ハンドラ | バッチ操作キャンセル時の状態更新 | バッチ操作がキャンセルされた | カーソルが適切なファイルに位置する |

**calculateCursorTargetAfterBatchMove の契約**:

```
calculateCursorTargetAfterBatchMove(markedFiles map[string]bool) -> string

事前条件:
  - p.entries が現在のディレクトリ内容を反映している
  - p.cursor が有効なインデックスである
  - markedFiles にバッチ操作対象のファイル名が含まれている

事後条件:
  - マークされていない最近接ファイルの名前を返す
  - 見つからない場合は空文字を返す

走査アルゴリズム:
  1. カーソル位置から上方向に走査（カーソル位置を含む）
  2. ".." エントリはスキップ（カーソルターゲットとしない）
  3. マークされていないファイルが見つかればその名前を返す
  4. 上方向で見つからなければ、カーソル位置+1から下方向に走査
  5. マークされていないファイルが見つかればその名前を返す
  6. 全てマーク済みなら空文字を返す
```

**処理フロー（batchCompleteMsg / batchCancelledMsg ハンドラ）**:
```
1. 操作種別を確認（msg.operation）
2. move 操作の場合:
   a. マーク情報のコピーを保存
   b. calculateCursorTargetAfterBatchMove でカーソルターゲット名を決定
   c. ClearMarks() でマークをクリア
   d. RefreshDirectoryPreserveCursor() でディレクトリ再読込
      |- ターゲット名が空でない場合
      |   |- エントリ内でターゲット名を検索してカーソルを設定
      |   |- EnsureCursorVisible() でスクロール位置を調整
      |- ターゲット名が空の場合
          |- フォールバック（フェーズ1のインデックスクランプ）に委ねる
3. copy 操作の場合:
   a. ClearMarks() でマークをクリア
   b. RefreshDirectoryPreserveCursor() でディレクトリ再読込
      （ファイルが消えないためファイル名マッチで保持される）
4. 移動先ペインも RefreshDirectoryPreserveCursor() で再読込
5. ステータスメッセージを表示
```

**注意:** コピー操作ではソースペインのファイルが消えないため、`calculateCursorTargetAfterBatchMove` は不要。通常の `RefreshDirectoryPreserveCursor` のファイル名マッチで十分。

**注意:** `ClearMarks()` は `RefreshDirectoryPreserveCursor()` 内部でもマークをクリアする（`p.markedFiles = make(map[string]bool)`）ため、ハンドラでの `ClearMarks()` 呼び出しは技術的には冗長である。ただし move 操作時は `calculateCursorTargetAfterBatchMove` の前にマーク情報を保存し、その後 `ClearMarks()` で明示的にクリアする必要があるため、意図を明確にするために残す。

**実装ステップ**:

1. **calculateCursorTargetAfterBatchMove のテストを作成**
   - `pane_cursor_test.go` にテーブル駆動テストを追加
   - 既存の `pane_delete_test.go` のパターンを踏襲（Pane 構造体を直接構築してメソッドを呼び出す）
   - 全テストシナリオ（後述のテスト表）をカバー

2. **calculateCursorTargetAfterBatchMove を pane.go に追加**
   - `calculateCursorAfterDeletion` の近くに配置
   - 純粋な計算メソッド（副作用なし）として実装

3. **batchCompleteMsg / batchCancelledMsg ハンドラを更新**
   - ハンドラ内でカーソルターゲット計算を呼び出し
   - 計算結果に基づいてカーソル位置を設定

**依存関係**:
- 前提: フェーズ 1（インデックスフォールバックが空文字ターゲット時のセーフティネット）
- 後続: なし

**テスト方針**:

*ユニットテスト（pane_cursor_test.go を新規作成）*:

| ID | シナリオ | entries | cursor | markedFiles | 期待結果 |
|----|---------|---------|--------|------------|---------|
| B-1 | カーソルが非マーク上 | `[.., a, *b, *c, d]` cursor=4(d) | 4 | {b, c} | "d" |
| B-2 | カーソルから上に非マークあり | `[.., a, *b, *c]` cursor=3(*c) | 3 | {b, c} | "a" |
| B-3 | カーソル上が全マーク・下に非マークあり | `[.., *a, *b, c, d]` cursor=1(*a) | 1 | {a, b} | "c" |
| B-4 | 全ファイルマーク済み | `[.., *a, *b]` cursor=1 | 1 | {a, b} | "" |
| B-5 | 単一ファイルのみマーク | `[.., *a]` cursor=1 | 1 | {a} | "" |
| B-6 | マーク間に非マークファイルあり | `[.., *a, b, *c, d, *e]` cursor=4(d) | 4 | {a, c, e} | "d" |
| B-7 | カーソル位置0（..上） | `[.., *a, *b]` cursor=0 | 0 | {a, b} | "" |

テスト手法: `calculateCursorAfterDeletion` テストと同様に、`Pane` 構造体の `entries` と `cursor` を直接設定し、メソッドの戻り値を検証する。`FileEntry` は `Name` フィールドのみ設定すれば十分。`..` エントリは `Name: ".."` で `IsParentDir()` が true を返す。

**受入基準**:
- [ ] カーソル位置が非マークファイル上の場合、そのファイル名を返す
- [ ] カーソルから上方向走査で ".." をスキップする
- [ ] 上方向に非マークファイルがない場合、下方向に走査する
- [ ] 全ファイルがマーク済みの場合、空文字を返す
- [ ] batchCompleteMsg ハンドラでカーソルターゲットが正しく適用される（move 操作時のみ）
- [ ] batchCancelledMsg ハンドラでも同様にカーソルターゲットが適用される（move 操作時のみ）
- [ ] copy 操作時は通常の RefreshDirectoryPreserveCursor のファイル名マッチに委ねる
- [ ] カーソルターゲット設定後に EnsureCursorVisible が呼ばれスクロール位置が正しい
- [ ] 既存テスト（pane_test.go の TestRefreshDirectoryPreserveCursor を含む）を新しいフォールバック動作に合わせて更新し、全て合格する

**推定工数**: 小

---

## 完成後のファイル構成

```
internal/ui/
  pane_filter.go               # RefreshDirectoryPreserveCursor のフォールバック改善
  pane_filter_test.go          # フォールバックテスト追加
  pane_test.go                 # 既存テストの期待値更新
  pane.go                      # calculateCursorTargetAfterBatchMove 追加
  pane_cursor_test.go          # バッチカーソル計算テスト（新規作成）
  model_update.go              # batchCompleteMsg/batchCancelledMsg ハンドラ更新（move/copy分岐追加）
```

**ファイル別の変更内容**:

| ファイル | 変更種別 | 内容 |
|--------|---------|------|
| `pane_filter.go` | 変更 | `RefreshDirectoryPreserveCursor` のフォールバックを旧インデックスクランプに変更 |
| `pane_filter_test.go` | 変更 | フォールバック動作のテストケースを追加 |
| `pane_test.go` | 変更 | `TestRefreshDirectoryPreserveCursor` の期待値をフォールバック変更に合わせて更新 |
| `pane.go` | 変更 | `calculateCursorTargetAfterBatchMove` メソッドを追加 |
| `pane_cursor_test.go` | 新規 | `calculateCursorTargetAfterBatchMove` のテーブル駆動テスト |
| `model_update.go` | 変更 | バッチ完了/キャンセルハンドラにカーソルターゲット計算を組み込み（move/copy 分岐追加） |

## テスト戦略

### ユニットテスト

**方針**:
- `calculateCursorTargetAfterBatchMove` は副作用のない純粋関数として設計し、`calculateCursorAfterDeletion` と同様にテーブル駆動テストで検証
- `RefreshDirectoryPreserveCursor` のフォールバックは実際のファイルシステム（一時ディレクトリ）を使って検証
- 既存の `pane_delete_test.go` パターン（Pane 直接構築、テーブル駆動テスト）を踏襲

**カバレッジ目標**:
- `calculateCursorTargetAfterBatchMove`: 100%（全分岐をテーブル駆動テストでカバー）
- `RefreshDirectoryPreserveCursor` のフォールバック部分: 主要パス全カバー

### 手動テスト

- [ ] 単一ファイル移動後、ソースペインのカーソルが次のファイルに位置する
- [ ] 末尾ファイル移動後、カーソルが最後のエントリに位置する
- [ ] バッチ移動後、カーソルがマークされていないファイルに位置する
- [ ] 全ファイルバッチ移動後、カーソルが ".." に位置する
- [ ] コピー操作後、両ペインのカーソルが元の位置を保持する
- [ ] リネーム後、カーソルが新しいファイル名に位置する

## リスク評価

### 技術リスク

1. **RefreshDirectoryPreserveCursor のフォールバック変更による既存動作への影響**
   - **可能性**: 低
   - **影響**: 中
   - **対策**: 既存テストの全合格を確認。フォールバックは「0 にリセット」が「旧インデックスにクランプ」に変わるだけで、正常系（ファイル名マッチ成功）には影響しない

2. **バッチ操作ハンドラでの RefreshDirectoryPreserveCursor 呼び出し順序**
   - **可能性**: 低
   - **影響**: 中
   - **対策**: カーソルターゲット計算は `ClearMarks` / `RefreshDirectoryPreserveCursor` の前に実行する。リフレッシュ後にファイル名で再検索するため、エントリ変動に対して安全

## 未解決事項

### 仕様からの確認事項
- なし（SPEC.md で全ケースが定義済み）

### 実装固有の確認事項
- `batchCancelledMsg` の場合、一部ファイルは既に移動済みのため、マーク情報と実際のエントリに乖離がある。`RefreshDirectoryPreserveCursor` のインデックスフォールバックがセーフティネットとして機能する

## 参照

- **仕様書**: `doc/tasks/preserve-cursor-after-file-operation/SPEC.md`
- **既存パターン**: `internal/ui/pane.go` の `calculateCursorAfterDeletion`
- **既存テストパターン**: `internal/ui/pane_delete_test.go`
