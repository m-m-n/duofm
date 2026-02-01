# Feature: Preserve Cursor Position After File Operations

## Overview

duofm ではファイル操作（移動・コピー・リネーム・バッチ操作）の完了後に `LoadDirectory()` でディレクトリを再読み込みしていたため、カーソルが先頭（インデックス 0）にリセットされていた。本機能では、ファイル操作後のカーソル位置を適切に保持・計算するよう改善する。

## Objectives

- ファイル操作後にカーソル位置を保持し、作業コンテキストを維持する
- ソースペイン（ファイルが減る側）ではインデックスベースのフォールバックでカーソルを保持する
- 移動先ペイン（ファイルが増える側）ではファイル名マッチでカーソルを保持する
- バッチ操作後はマーク情報を考慮したカーソル位置計算を行う

## User Stories

### US1: 単一ファイル移動後のソースペインカーソル保持

移動したファイルが消えた後、カーソルは同じインデックス位置に留まる（次のファイルが繰り上がる）。末尾を超えた場合は最後のエントリに移動する。

**Acceptance Criteria:**
- [ ] 移動後、ソースペインのカーソルは同じインデックス位置を保持する
- [ ] カーソルが末尾を超えた場合、最後のエントリに移動する
- [ ] エントリが空（`..` のみ）の場合、カーソルは 0 に移動する

### US2: 単一ファイル移動後の移動先ペインカーソル保持

ファイルが増える側ではファイル名マッチでカーソル位置を保持する。

**Acceptance Criteria:**
- [ ] 移動先ペインのカーソルは操作前の位置を保持する
- [ ] ファイル名マッチで正しい位置を復元する

### US3: バッチ移動後のソースペインカーソル位置

複数マークファイルを移動した後、カーソルはマークされていないファイルに移動する。

**Acceptance Criteria:**
- [ ] カーソルは最後のマークファイルから上方向に最初に見つかるマークされていないファイルに移動する
- [ ] マークファイルより上にマークされていないファイルがない場合、下方向に最初のマークされていないファイルに移動する
- [ ] 全ファイルがマークされていた場合、カーソルは 0 に移動する

### US4: コピー後のカーソル保持

コピー操作ではソースペインのファイルは消えないため、ファイル名マッチで両ペインのカーソルを保持する。

**Acceptance Criteria:**
- [ ] ソースペインのカーソルは操作前のファイル位置を保持する
- [ ] 移動先ペインのカーソルは操作前のファイル位置を保持する

### US5: リネーム・ファイル作成・その他操作後のカーソル保持

リネームやファイル作成後は既存のカーソル移動ロジック（`moveCursorToFile` 等）が適用されるため、基盤のリロード処理をカーソル保持対応にする。

**Acceptance Criteria:**
- [ ] 基盤のリロードでカーソル位置が保持される
- [ ] 後続の `moveCursorToFile` が正しく動作する

## Technical Requirements

### Functional Requirements

- **FR1:** `RefreshDirectoryPreserveCursor` にインデックスベースのフォールバックを追加する
- **FR2:** バッチ操作完了時にマーク情報を考慮したカーソル位置計算を行う
- **FR3:** 全操作箇所で `LoadDirectory()` を `RefreshDirectoryPreserveCursor()` に置き換える（先行修正で完了済み）
- **FR4:** スクロール位置をカーソルに合わせて自動調整する

### Non-Functional Requirements

- **NFR1 - Performance:** カーソル位置計算は O(n)（n = エントリ数）以内
- **NFR2 - Maintainability:** 既存の `calculateCursorAfterDeletion` パターンと整合する設計
- **NFR3 - Usability:** カーソル移動は自然で予測可能

## Implementation Approach

### Architecture

#### RefreshDirectoryPreserveCursor の改善

**Current Flow:**
```
RefreshDirectoryPreserveCursor()
  ↓
ファイル名で検索
  ↓
見つからない → cursor = 0 (問題)
```

**New Flow:**
```
RefreshDirectoryPreserveCursor()
  ↓
ファイル名で検索
  ↓
見つかった → そのインデックス
見つからない → 旧カーソルインデックスを保持（clamp to range）
```

#### バッチ操作完了時のカーソル計算

**New Flow:**
```
batchCompleteMsg / batchCancelledMsg
  ↓
マーク情報から非マークファイルの位置を計算
  ↓
ClearMarks()
  ↓
RefreshDirectoryPreserveCursor()
  ↓
計算したカーソル位置を設定
```

### API Design

#### Method 1: RefreshDirectoryPreserveCursor の改善

**Current Signature:**
```go
func (p *Pane) RefreshDirectoryPreserveCursor() error
```

**変更内容:** ファイル名マッチ失敗時、カーソルを 0 ではなく旧インデックスで clamp する。

**Logic:**
```go
// ファイル名で検索
found := false
for i, e := range entries {
    if e.Name == selectedName {
        newCursor = i
        found = true
        break
    }
}

// フォールバック: 旧インデックスを clamp
if !found {
    if oldCursor >= len(entries) {
        newCursor = max(0, len(entries)-1)
    } else {
        newCursor = oldCursor
    }
}
```

#### Method 2: calculateCursorAfterBatchMove (New)

**Purpose:** バッチ移動/コピー後のソースペインのカーソル位置を計算する。

**Signature:**
```go
func (p *Pane) calculateCursorAfterBatchMove(markedFiles map[string]bool) int
```

**Parameters:**
- `markedFiles`: 移動/コピー前のマークファイル名マップ

**Returns:**
- `int`: 新しいカーソル位置（0-indexed）

**Algorithm:**
```
1. 現在のカーソル位置を取得
2. カーソルから上方向に走査
3. マークされていないファイルを見つけたらそのインデックスを返す
4. 見つからなければ下方向に走査
5. マークされていないファイルを見つけたらそのインデックスを返す
6. 全てマークされていた場合は 0 を返す
```

**Detailed Logic:**
```go
func (p *Pane) calculateCursorAfterBatchMove(markedFiles map[string]bool) int {
    cursor := p.cursor

    // 上方向に走査（カーソル位置から）
    for i := cursor; i >= 0; i-- {
        if i < len(p.entries) && !markedFiles[p.entries[i].Name] {
            return i
        }
    }

    // 下方向に走査（カーソル位置+1から）
    for i := cursor + 1; i < len(p.entries); i++ {
        if !markedFiles[p.entries[i].Name] {
            return i
        }
    }

    // 全てマーク済み
    return 0
}
```

**Note:** このメソッドは `RefreshDirectoryPreserveCursor()` の **前に** 呼ぶ必要がある。リフレッシュ後はマーク情報とエントリが変わるため。ただし、リフレッシュ後のエントリでインデックスが有効かの検証が必要。

実際のフローは以下:
1. マーク情報を保存（`markedFiles` のコピー）
2. マークされていないファイル名を特定
3. `ClearMarks()`
4. `RefreshDirectoryPreserveCursor()` → ファイル名マッチで保持
5. もしファイル名マッチに失敗した場合 → インデックスフォールバックが適用

より正確なアプローチ: マークされていないファイルの **名前** をカーソルターゲットとして保存し、`RefreshDirectoryPreserveCursor` のファイル名マッチに使う。

#### Method 3: calculateCursorTargetAfterBatchMove (Revised approach)

**Purpose:** バッチ操作前に、操作後のカーソルターゲットとなるファイル名を決定する。

**Signature:**
```go
func (p *Pane) calculateCursorTargetAfterBatchMove(markedFiles map[string]bool) string
```

**Returns:**
- `string`: カーソルターゲットのファイル名。空文字の場合はフォールバック。

**Algorithm:**
```
1. 現在のカーソル位置を取得
2. カーソルから上方向に走査
3. マークされていないファイルを見つけたらその名前を返す
4. 見つからなければ下方向に走査
5. マークされていないファイルを見つけたらその名前を返す
6. 全てマークされていた場合は空文字を返す（0にフォールバック）
```

### Data Flow

#### 単一ファイル移動

```
Before: entries=[.., a.txt, b.txt, c.txt, d.txt], cursor=2 (b.txt selected)
Operation: Move b.txt to other pane
After reload: entries=[.., a.txt, c.txt, d.txt]

RefreshDirectoryPreserveCursor:
  selectedName = "b.txt"
  Name search: not found
  Fallback: oldCursor=2, len(entries)=4 → cursor=2 (now c.txt)

Result: cursor=2 → c.txt (correct)
```

#### 単一ファイル移動（末尾）

```
Before: entries=[.., a.txt, b.txt], cursor=2 (b.txt selected)
Operation: Move b.txt to other pane
After reload: entries=[.., a.txt]

RefreshDirectoryPreserveCursor:
  selectedName = "b.txt"
  Name search: not found
  Fallback: oldCursor=2, len(entries)=2 → cursor=1 (clamp to len-1)

Result: cursor=1 → a.txt (correct)
```

#### バッチ移動

```
Before: entries=[.., a.txt, *b.txt, *c.txt, d.txt, *e.txt], cursor=4 (d.txt)
  Marked: b.txt, c.txt, e.txt

calculateCursorTargetAfterBatchMove:
  cursor=4 (d.txt, not marked) → return "d.txt"

After ClearMarks + RefreshDirectoryPreserveCursor:
  entries=[.., a.txt, d.txt]
  Name match: "d.txt" found at index 2
  cursor=2 → d.txt (correct)
```

#### バッチ移動（カーソル上が全てマーク済み）

```
Before: entries=[.., *a.txt, *b.txt, *c.txt, d.txt, e.txt], cursor=1 (a.txt)
  Marked: a.txt, b.txt, c.txt

calculateCursorTargetAfterBatchMove:
  cursor=1 (a.txt, marked)
  Search up from 1: i=1(marked), i=0(..) → ".." is not marked → return ".."

  ※ ".." は特殊。上方向に見つからない場合は下方向へ:
  Search down from 2: i=2(marked), i=3(marked), i=4(d.txt, not marked) → return "d.txt"

After ClearMarks + RefreshDirectoryPreserveCursor:
  entries=[.., d.txt, e.txt]
  Name match: "d.txt" found at index 1
  cursor=1 → d.txt (correct)
```

**注意:** `..`（親ディレクトリ）はマーク対象外なので、上方向走査で `..` に到達した場合はスキップして下方向走査に移行する。

#### バッチ移動（全ファイルマーク済み）

```
Before: entries=[.., *a.txt, *b.txt, *c.txt], cursor=1
  Marked: a.txt, b.txt, c.txt

calculateCursorTargetAfterBatchMove:
  Search up: all marked (skip ..)
  Search down: all marked
  return "" (empty)

After ClearMarks + RefreshDirectoryPreserveCursor:
  entries=[..]
  Name match: "" → not found
  Fallback: oldCursor=1, len(entries)=1 → clamp to 0

Result: cursor=0 → .. (correct)
```

### Affected Files

```
internal/ui/
├── pane_filter.go               # RefreshDirectoryPreserveCursor の改善
├── pane_filter_test.go          # フォールバックテスト追加
├── pane.go                      # calculateCursorTargetAfterBatchMove 追加
├── pane_cursor_test.go          # バッチカーソル計算テスト（新規）
├── model_update.go              # batchCompleteMsg/batchCancelledMsg のカーソル保持
└── model_update_dialog.go       # （先行修正で完了済み）
```

### Dependencies

**Internal Dependencies:**
- `internal/ui/pane_filter.go`: RefreshDirectoryPreserveCursor
- `internal/ui/pane.go`: calculateCursorAfterDeletion（既存パターンの参考）
- `internal/fs`: ReadDirectory（変更なし）

**External Dependencies:**
None

## Test Scenarios

### Unit Tests

#### Test 1: RefreshDirectoryPreserveCursor - ファイル名マッチ成功

**Description:** ファイル名が存在する場合、そのインデックスにカーソルが移動する
**Expected:** ファイル名のインデックスが返る

#### Test 2: RefreshDirectoryPreserveCursor - ファイル名マッチ失敗・インデックス有効

**Description:** ファイルが消えたが、旧インデックスが有効範囲内
**Expected:** 旧インデックスがそのまま使われる

```
Before: [.., a.txt, b.txt, c.txt], cursor=2 (b.txt)
After:  [.., a.txt, c.txt]
b.txt not found, oldCursor=2, len=3 → cursor=2 (c.txt)
```

#### Test 3: RefreshDirectoryPreserveCursor - ファイル名マッチ失敗・インデックス超過

**Description:** ファイルが消え、旧インデックスが範囲外
**Expected:** 最後のエントリにクランプされる

```
Before: [.., a.txt, b.txt], cursor=2 (b.txt)
After:  [.., a.txt]
b.txt not found, oldCursor=2, len=2 → cursor=1 (a.txt)
```

#### Test 4: RefreshDirectoryPreserveCursor - 全ファイル削除

**Description:** ファイルが全て消え、`..` のみ残る
**Expected:** cursor=0

```
Before: [.., a.txt], cursor=1 (a.txt)
After:  [..]
a.txt not found, oldCursor=1, len=1 → cursor=0 (..)
```

#### Test 5: calculateCursorTargetAfterBatchMove - カーソル上に非マークファイルあり

**Description:** カーソル位置から上方向に非マークファイルがある
**Expected:** そのファイル名が返る

```
entries=[.., a.txt, *b.txt, *c.txt, d.txt], cursor=3 (*c.txt)
Marked: b.txt, c.txt
Search up from 3: i=3(marked), i=2(marked), i=1(a.txt, not marked) → "a.txt"
```

#### Test 6: calculateCursorTargetAfterBatchMove - カーソル上が全てマーク（.. 除く）

**Description:** カーソルから上方向は全てマーク済みで、下に非マークファイルがある
**Expected:** 下方向の最初の非マークファイル名が返る

```
entries=[.., *a.txt, *b.txt, c.txt, d.txt], cursor=1 (*a.txt)
Marked: a.txt, b.txt
Search up: i=1(marked), i=0(..) → skip ..
Search down: i=2(marked), i=3(c.txt, not marked) → "c.txt"
```

#### Test 7: calculateCursorTargetAfterBatchMove - 全ファイルマーク済み

**Description:** 全てのファイルがマークされている
**Expected:** 空文字が返る

```
entries=[.., *a.txt, *b.txt], cursor=1
Marked: a.txt, b.txt
Search up: all marked (skip ..)
Search down: all marked
→ ""
```

#### Test 8: calculateCursorTargetAfterBatchMove - カーソルが非マークファイル上

**Description:** カーソル位置のファイルがマークされていない
**Expected:** そのファイル名がそのまま返る

```
entries=[.., *a.txt, b.txt, *c.txt], cursor=2 (b.txt)
Marked: a.txt, c.txt
cursor=2 (b.txt, not marked) → "b.txt"
```

### Integration Tests

#### Test 9: バッチ操作完了後のカーソル位置統合テスト

**Description:** 実際のバッチ操作完了フローでカーソルが正しい位置に設定される
**Expected:** マーク情報を反映したカーソル位置

### Edge Cases

#### Edge Case 1: エントリが空のディレクトリ

**Description:** エントリが 0 件（通常は発生しないが防御的に対応）
**Expected:** cursor=0

#### Edge Case 2: 単一ファイルのみのディレクトリでマーク移動

**Description:** `..` と 1 ファイルのみで、そのファイルをマーク移動
**Expected:** cursor=0 (`..`)

#### Edge Case 3: 連続する非マークファイルの間にマークファイルがある

**Description:** `[.., a, *b, c, *d, e]` で b, d をマーク移動
**Expected:** カーソル位置に応じた適切な非マークファイルに移動

## Error Handling

- カーソル位置が範囲外になった場合: `max(0, len(entries)-1)` にクランプ
- ディレクトリ読み込み失敗: 既存のエラーハンドリングを継続（カーソル変更なし）
- マーク情報が空の場合: 通常の `RefreshDirectoryPreserveCursor` フォールバック

## Success Criteria

- [ ] 全ユニットテスト合格
- [ ] 単一ファイル移動後のカーソル保持が正しく動作する
- [ ] バッチ操作後のカーソル位置計算が正しく動作する
- [ ] 既存の削除操作のカーソル動作に影響しない
- [ ] 既存テストがすべて合格する
