# Feature: minibuffer-grapheme-bounds-linear

## Overview

ミニバッファの書記素区切り計算 `graphemeClusterBounds`（`internal/ui/minibuffer.go`）は `uniseg.NewGraphemes` の `Next()` で書記素を走査しており、「.」+ 空白 n 個 +「A」のような入力で 1 回の呼び出しが O(n^2) になる。この関数は View（カーソルが末尾以外のときの描画）、Left/Ctrl+B、Right/Ctrl+F から毎回呼ばれる。走査を `uniseg.FirstGraphemeClusterInString` による状態引き継ぎの前進走査に置き換え、入力長に対して線形の時間にする。区切り結果と呼び出し側の動作は変えない。

## Objectives

- ミニバッファの書記素区切り計算（`graphemeClusterBounds`）を入力長に対して線形の時間にし、「.」+ 大量の空白を含む長い入力でもカーソル移動と描画に目立つ遅れが出ないようにする
- minibuffer-grapheme-cursor-highlight の SPEC NFR3（書記素の区切り計算も入力長に対して線形）を満たす

## User Stories

### US1: 長い入力でも遅れなくカーソル移動と描画が行われる
ミニバッファに「.」の後に大量の空白が続く長い入力を貼り付けたユーザーとして、カーソルを末尾付近で左右に動かしたときに、入力長に対して線形の時間でカーソル移動と描画が行われ、操作に目立つ遅れが出ないようにしたい。

**Acceptance Criteria:**
- [ ] AC1: 「.」+ `strings.Repeat(" ", n)` +「A」の入力で、カーソルを末尾の 1 つ手前（「A」の位置）に置いた View が `linearTimeLimit`（10 秒）以内に返り、結果は改行を含まず表示幅が `m.width-2` 以下である（FR1, NFR1）
- [ ] AC2: 同じ入力で、カーソル末尾から Left と Ctrl+B をそれぞれ 1 回押すと `linearTimeLimit` 以内に処理が終わり、`cursorPos` が len-1 になる（FR1, FR2, NFR1）
- [ ] AC3: 同じ入力で、カーソルを len-1 に置いて Right と Ctrl+F をそれぞれ 1 回押すと `linearTimeLimit` 以内に処理が終わり、`cursorPos` が len になる（FR1, FR2, NFR1）
- [ ] AC4: AC1〜AC3 の回帰テストの n は、変更前の実装（`NewGraphemes` による走査）では `linearTimeLimit` を超えて失敗する大きさである（NFR1）
- [ ] AC5: 既存の書記素移動テスト（家族絵文字、結合文字、VS16、国旗、先頭の孤立した結合文字・ZWJ、空入力）と書記素全体の反転テストが変更なしで通る（FR2, NFR3）
- [ ] AC6: `go test ./...`、`go vet ./...` が通り、`gofmt` で差分が出ない（NFR2, NFR3）

## Technical Requirements

### Functional Requirements
- **FR1: 線形走査による書記素区切り** — `internal/ui/minibuffer.go` の `graphemeClusterBounds(runes []rune, p int) (start, end int)` は、`uniseg.NewGraphemes` の `Next()` による走査をやめ、入力全体の先頭から `uniseg.FirstGraphemeClusterInString(rest, state)` を `state=-1` で始め、返された state を次の呼び出しに引き継ぎながら各書記素の rune 数を数える 1 回の前進走査で、rune 位置 p を含む書記素の [start, end) を返す。p を含む書記素が見つかった時点で走査を終える。
- **FR2: 区切り結果と呼び出し側の動作を変えない** — `graphemeClusterBounds` の関数シグネチャ、返す区切り（入力全体を先頭から区切った extended grapheme cluster の境界）、事前条件 `0 <= p < len(runes)`、および範囲外の p に対する防御的な戻り値 `(p, p+1)` は変えない。呼び出し側（View、Left/Ctrl+B、Right/Ctrl+F）の動作も変えない。

### Non-Functional Requirements
- **NFR1 - 計算量:** `graphemeClusterBounds` の 1 回の呼び出しは入力長に対して線形の時間で終わる。「.」+ 空白 n 個 +「A」の入力でカーソルが末尾付近にある場合も同じ。View の計算量は入力長に対して線形のまま、`guardFit` の再測定は最大 3 回のまま維持する。
- **NFR2 - 依存:** `go.mod` に既にある依存（`rivo/uniseg` v0.4.7 を含む）以外の新しい依存を加えない。
- **NFR3 - 既存テストと静的検査:** 既存の `internal/ui/minibuffer_test.go` のテスト（書記素移動・書記素全体の反転・線形時間の TS1〜TS4、TS7 を含む）がすべて変更なしで通り、`go vet ./...` と `gofmt` で差分が出ない。

## Implementation Approach

### Architecture

変更は `internal/ui/minibuffer.go` の `graphemeClusterBounds` の内部実装に閉じる。関数シグネチャと呼び出し側は変えない（FR2）。

**Component Diagram:**
```
View（カーソルが末尾以外）─┐
Left / Ctrl+B ─────────────┼─→ graphemeClusterBounds(runes, p) ─→ uniseg.FirstGraphemeClusterInString
Right / Ctrl+F ────────────┘
```

### Data Flow

```
runes, p
  → rest = 入力全体の文字列, state = -1, pos = 0
  → FirstGraphemeClusterInString(rest, state) で先頭の書記素を取り出し、state を引き継ぐ
  → 書記素の rune 数を pos に加算
  → p を含む書記素に達したら [start, end) を返して走査を終える
```

### Dependencies

**Internal Dependencies:**
- `internal/ui/minibuffer.go` の View、Left/Ctrl+B、Right/Ctrl+F: `graphemeClusterBounds` の呼び出し側。動作は変えない（FR2）
- `partitionUnits`: 変更しない（A4）

**External Dependencies:**
- `github.com/rivo/uniseg` v0.4.7: 既存の依存。新しい依存は加えない（NFR2）

### File Structure

```
internal/ui/
├── minibuffer.go        # graphemeClusterBounds の走査を置き換える（FR1, FR2）
└── minibuffer_test.go   # 回帰テスト TS1〜TS3 を追加する
```

## Declared Change Set

上記の機能固有のパスは手書きの一覧ではなく、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

機能固有のパスに加えて、既定で次の 2 つのワークフロー生成物を宣言する。

- `feature-docs/minibuffer-grapheme-bounds-linear/**`
- `test-docs/minibuffer-grapheme-bounds-linear/**`

`feature-docs/minibuffer-grapheme-bounds-linear/**` は `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、デザイン成果物を含む。これらはフェーズ文書と `references/phase-state.md` が生成・所有する。

`test-docs/minibuffer-grapheme-bounds-linear/**` はタスクごとのテスト記録 `test-docs/minibuffer-grapheme-bounds-linear/{T}.tests.yaml` を含む。これは `implement-phase.md` が生成・所有する。

この宣言は上位集合の宣言であり、検証時に観測される実際の変更集合は宣言した集合に含まれていればよい。実装タスクのないフィーチャーでは `test-docs/minibuffer-grapheme-bounds-linear/` は生成されないが、宣言が実体化しないことは違反ではない。

## Test Scenarios

### Unit Tests
- [ ] TS1: 「.」+ `strings.Repeat(" ", n)` +「A」、カーソルを len-1 に置き、既存の `runViewTimed` で View を計時して `linearTimeLimit` 以内に返ること、改行なし・表示幅 `m.width-2` 以下であることを確認する（AC1, AC4）
- [ ] TS2: 同じ入力でカーソル末尾から Left と Ctrl+B を別ゴルーチンで 1 回実行し、`linearTimeLimit` 以内に終わり `cursorPos` が len-1 になることを確認する（panic はテスト失敗として報告する）（AC2, AC4）
- [ ] TS3: 同じ入力でカーソル len-1 から Right と Ctrl+F を別ゴルーチンで 1 回実行し、`linearTimeLimit` 以内に終わり `cursorPos` が len になることを確認する（AC3, AC4）
- [ ] TS4: 既存の書記素移動・反転・線形時間・1 行幅上限のテストが変更なしで通ることを確認する（AC5）

### Static Checks
- [ ] TS5: `go test ./...`、`go vet ./...`、`gofmt` で差分なしを確認する（AC6）

### E2E Tests
**Existing E2E tests**: ミニバッファを扱う既存の E2E テストなし
**Run command**: Not detected
- 再発を検出するテストは `internal/ui/minibuffer_test.go` の単体テストで満たし、E2E テストは追加しない（A3）

### Edge Cases
- [ ] カーソルが末尾にあるとき View は `graphemeClusterBounds` を呼ばない（既存どおり）
- [ ] 国旗（地域指示記号 2 つ）や ZWJ 連結など、状態を引き継がないと区切りが変わる書記素でも、区切りは変更前と一致する
- [ ] 先頭の孤立した結合文字・ZWJ は、それ自体を 1 つの書記素として扱う
- [ ] 空入力（呼び出されない）と、範囲外の p に対する防御的な戻り値 `(p, p+1)`

### Performance Tests
- [ ] TS1〜TS3 の回帰テストで、View・Left/Ctrl+B・Right/Ctrl+F がそれぞれ `linearTimeLimit`（10 秒）以内に終わる（NFR1）

## Assumptions

- **A1:** 回帰テストの n は既存の線形時間テストと同じ 100000 とする。計測値からの外挿で、変更前の実装では 1 回の呼び出しが約 60 秒となり `linearTimeLimit`（10 秒）を超える。n=8000 では変更前でも約 0.4 秒で上限内に収まり、再発を検出できない。
- **A2:** HandleKey の計時は、既存の `runViewTimed` と同じ形（別ゴルーチンで実行し、panic を回収し、`linearTimeLimit` で打ち切る）のヘルパーで行う。上限超過で打ち切ったゴルーチンは放置する（既存テストと同じ扱い）。
- **A3:** 再発を検出するテストは `internal/ui/minibuffer_test.go` の単体テストで満たし、E2E テストは追加しない。既存の E2E スクリプトにミニバッファを扱うテストはない。
- **A4:** `partitionUnits`（`FirstGraphemeClusterInString` を `state=-1` で毎回呼ぶ）は変更しない。
- **A5:** `graphemeClusterBounds` の区切り結果は変更前と同じ（入力全体を先頭から区切る extended grapheme cluster）であり、既存テストがこれを固定している。

## Success Criteria

- [ ] すべての機能要件が実装され、テストされている
- [ ] すべてのテストシナリオが通る
- [ ] 再現手順（ミニバッファに「.」+ 空白 8000 個程度 +「A」を貼り付け、カーソルを末尾付近で左右に動かす）で遅れが起きない
- [ ] 再発を検出するテストがある

## Open Questions

なし

## References

- minibuffer-grapheme-cursor-highlight の SPEC（NFR3）: `feature-docs/minibuffer-grapheme-cursor-highlight/SPEC.md`
