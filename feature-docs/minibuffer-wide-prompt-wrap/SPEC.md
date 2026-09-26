# Feature: minibuffer-wide-prompt-wrap

## Overview

`Minibuffer.View` のプロンプトと入力部分の幅計算を、バイト数・rune 数から lipgloss の表示幅に切り替える。全角文字を含むプロンプトや入力があっても、ミニバッファはどの幅でも 1 行で描画され、ミニバッファ表示時のペインの行数はプロンプトの文字種によらず一定になる。

要件定義書: `feature-docs/minibuffer-wide-prompt-wrap/REQUIREMENTS.md`

## Objectives

- 全角文字を含むプロンプトや入力があっても、ミニバッファをどの幅でも 1 行で描画する
- ミニバッファ表示時のペイン全体の行数を、プロンプトの文字種によらず一定に保つ

## User Stories

### US1: 狭い幅で全角パターンの履歴検索を行う

duofm の利用者として、ミニバッファ幅が 26〜32 程度の端末で、履歴検索（reverse-i-search）に全角文字のパターンを入力したい。ミニバッファが 1 行で描画され、ペインの行数がずれないようにするため。

**Acceptance Criteria:**
- [ ] AC1: 幅 24〜32 で、プロンプト `(reverse-i-search)'日本語': `（入力あり・なし）のとき、View の結果が改行を含まず、表示幅が `m.width-2` 以下である
- [ ] AC5: 両方の経路の幅（例: 26 と 37）で、全角プロンプトを表示したペインの行数が、同じ幅で 1 行に収まる ASCII プロンプトを表示したペインの行数と一致する

### US2: 全角文字を含む入力の編集

duofm の利用者として、全角文字を含む入力を編集しているとき、カーソルがどこにあってもミニバッファが 1 行で描画されるようにしたい。

**Acceptance Criteria:**
- [ ] AC2: 幅 36〜37 で、同じプロンプトと入力 `echo 日本語` のとき、カーソルが先頭・途中・全角文字の上・末尾のどこにあっても、View の結果が改行を含まず、表示幅が `m.width-2` 以下である
- [ ] AC3: `availableWidth >= 1` の経路で ASCII の長い入力を与えたとき、View の結果が改行を含まず、表示幅が `m.width-2` 以下である

### 既存の挙動の維持

**Acceptance Criteria:**
- [ ] AC4: 幅 0〜4 で、パニックせず改行を含まないことを確かめる既存のテストが通る
- [ ] AC6: 幅 40 で、プロンプト `/: ` と入力 `test` がそのまま表示される（既存の `TestMinibufferView` が通る）
- [ ] AC7: `go test ./...` がすべて通る

## Technical Requirements

### Functional Requirements

- **FR1:** 表示幅の基準 — `Minibuffer.View` は、プロンプトと入力部分の幅を、lipgloss が折り返しの判定に使うのと同じ表示幅（`lipgloss.Width` 系）で計算する。バイト数（`len`）や rune 数を表示幅として使わない。
- **FR2:** 1 行の保証 — `m.width > 4` のとき、プロンプト・入力の表示部分・カーソルブロックの合計の表示幅を内容幅（`m.width-4`）以下にする。`availableWidth <= 0` と `>= 1` の両方の経路で、View の結果は改行を含まず、表示幅は `m.width-2` 以下になる。
- **FR3:** プロンプトだけで内容幅を使い切るとき — プロンプトの表示幅が内容幅以上のときは、プロンプトを表示幅で内容幅以下に切り詰めて描画し、入力部分とカーソルブロックは描画しない。
- **FR4:** 入力部分の表示範囲 — 残りの幅（内容幅 − プロンプトの表示幅）が 1 以上のときは、入力部分の表示範囲を表示幅で決める。カーソル位置の文字（カーソルが末尾にあるときは 1 桁のカーソルブロック）が表示範囲に入るようにする。rune 単位の編集位置（`cursorPos`）と表示の桁数を混同しない。
- **FR5:** 全角文字が境界にかかるとき — プロンプトの切り詰めや入力の表示範囲の境界に、全角文字が半分しか入らないときは、その文字を描画しない。
- **FR6:** 幅 0〜4 の既存の挙動を維持 — `m.width` が 0〜4 のとき、View はパニックせず、結果に改行を含まない。
- **FR7:** ペインの行数 — ミニバッファを表示しているとき、ペインの出力行数は、プロンプトに全角文字を含む場合も ASCII だけの場合と同じになる。
- **FR8:** 収まる内容はそのまま表示 — プロンプトと入力が内容幅に収まるときは、どちらも切り詰めずに表示する。

### Non-Functional Requirements

- **NFR1:** 変更範囲 — 変更は `Minibuffer.View` の表示計算に限る。`HandleKey` などの編集操作と、`cursorPos` が rune 単位であることは変えない。

## Implementation Approach

### Architecture

**Component Diagram:**
```
internal/ui/model_update_keyboard.go   履歴検索のプロンプトを組み立てる（変更なし）
        │  "(reverse-i-search)'" + pattern + "': "
        ▼
internal/ui/minibuffer.go              Minibuffer.View（変更対象）
        │  プロンプト・入力の表示部分・カーソルブロックを 1 行で返す
        ▼
internal/ui/pane_render.go             viewInternal がミニバッファに 1 行を割り当てる（変更なし）
```

### 現在の実装

- `availableWidth` は `m.width - len(m.prompt) - 2` で、`len` はバイト数（`internal/ui/minibuffer.go:206-207`）
- `availableWidth <= 0` かつ `m.width > 4` のとき、プロンプトを `[]rune` で `contentWidth = m.width-4` 個に切り詰めている（`internal/ui/minibuffer.go:238-247`）
- 入力の表示範囲は rune 数（`availableWidth` 個）で計算している（`internal/ui/minibuffer.go:248-256`）
- スタイルは `Width(m.width-2)` と `Padding(0,1)`。内容幅は `m.width-4`（`internal/ui/minibuffer.go:230-237,275-279`）
- `viewInternal` はミニバッファ表示時にファイル一覧を 1 行減らし、`minibuffer.View()` と改行 1 つを書く。`View` が改行を含むと、ペイン全体の行数が増える（`internal/ui/pane_render.go:67-71,109-113`）

### 幅の計算（変更後）

| 項目 | 値 |
|------|----|
| 内容幅 | `m.width-4` |
| プロンプトの幅 | プロンプトの表示幅（FR1） |
| 残りの幅 | 内容幅 − プロンプトの表示幅 |
| View の結果の上限 | 改行なし、表示幅 `m.width-2` 以下（FR2） |

| 条件 | 描画内容 |
|------|----------|
| `m.width` が 0〜4 | パニックせず、改行を含まない（FR6） |
| 残りの幅 <= 0 | プロンプトを表示幅で内容幅以下に切り詰める。入力部分とカーソルブロックは描画しない（FR3, FR5） |
| 残りの幅 >= 1 | 入力部分の表示範囲を表示幅で決め、カーソル位置の文字またはカーソルブロックを範囲に入れる（FR4, FR5） |
| プロンプトと入力が内容幅に収まる | どちらも切り詰めない（FR8） |

- 残りの幅が 1 桁で、カーソル位置の文字が全角のときは、1 行の保証（FR2）を優先し、その文字を描画しない（FR5）

### Data Flow

```
prompt, input, cursorPos (rune 単位), m.width
  → Minibuffer.View（表示幅で切り詰め・表示範囲を決定）
  → 1 行の文字列
  → viewInternal（ミニバッファに 1 行を割り当てる）
```

### Dependencies

**Internal Dependencies:**
- `internal/ui/pane_render.go`: `viewInternal` が `minibuffer.View()` の結果を 1 行として扱う
- `internal/ui/model_update_keyboard.go`: 履歴検索のプロンプトを組み立てる

**External Dependencies:**
- lipgloss v1.1.0: 表示幅の計測（`lipgloss.Width` 系、書記素単位の幅）。判定の基準にする
- go-runewidth v0.0.16: 既存の依存。表示幅の判定の基準には使わない

### File Structure

```
internal/ui/
├── minibuffer.go              # Minibuffer.View の表示計算（変更）
├── minibuffer_test.go         # TS1, TS2, TS4
└── pane_render_test.go        # TS3, TS4
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/minibuffer-wide-prompt-wrap/**`
- `test-docs/minibuffer-wide-prompt-wrap/**`

`feature-docs/minibuffer-wide-prompt-wrap/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/minibuffer-wide-prompt-wrap/**` covers `test-docs/minibuffer-wide-prompt-wrap/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/minibuffer-wide-prompt-wrap/` directory at all; the declared
`test-docs/minibuffer-wide-prompt-wrap/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests

- [ ] TS1（`internal/ui/minibuffer_test.go`）: 表形式のテスト。幅（負の分岐 24〜32、正の分岐 36〜37）× 全角プロンプト × 入力（空、全角を含む入力、ASCII の長い入力）× カーソル位置（先頭・途中・全角文字の上・末尾）の組み合わせで、改行を含まないことと、`lipgloss.Width(View()) <= m.width-2` を確かめる（AC1, AC2, AC3 / FR1, FR2, FR3, FR4, FR5）
- [ ] TS2（`internal/ui/minibuffer_test.go`）: 全角文字が切り詰め・表示範囲の境界にかかる幅で、表示幅の上限を超えないことを確かめる（AC1, AC2 / FR1, FR2, FR3, FR4, FR5）
- [ ] TS3（`internal/ui/pane_render_test.go`）: 負の分岐と正の分岐の幅で、全角プロンプトのときの `ViewWithMinibuffer` の行数が、1 行に収まる ASCII プロンプトのときの行数と一致することを確かめる（AC5 / FR7）
- [ ] TS4（`internal/ui/minibuffer_test.go`、`internal/ui/pane_render_test.go`）: 既存の幅 0〜4 のテスト、幅 0/1 の行数テスト、`TestMinibufferView` が変更後も通る（AC4, AC6 / FR6, FR8）

### Integration Tests

該当なし。

### E2E Tests

**Existing E2E tests**: 検出結果の入力なし
**Run command**: 未検出

E2E テストは追加しない。

### Edge Cases

- [ ] 結合文字や絵文字を含むプロンプト・入力: 書記素単位の表示幅で測る。カーソルが書記素の途中の rune にあっても 1 行の保証を維持する
- [ ] 全角文字の上にカーソルがあるとき: 反転表示される文字の 2 桁が表示範囲に収まる
- [ ] カーソルが末尾にあるとき: カーソルブロックの 1 桁を表示幅に含める
- [ ] プロンプトの表示幅が内容幅とちょうど同じ・1 桁少ない・1 桁多いとき
- [ ] 行末の空白は折り返し処理で取り除かれるため、幅の上限は View の結果の表示幅で確かめる

## Security Considerations

該当なし。

## Error Handling

該当なし。

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] AC1〜AC7 を満たす

## Assumptions

- a1: 表示幅は、lipgloss が折り返しの判定に使うのと同じ基準（`lipgloss.Width` 系、書記素単位の幅）で測る。go-runewidth は曖昧幅の扱いが異なることがあるため、判定の基準には使わない
- a2: 切り詰めの境界に全角文字が半分しか入らないときは、その文字を描画しない
- a3: 既存テストが確認している事後条件（幅 0〜4 でパニックせず改行もしない、幅 0/1 の行数が幅 40 と一致する）を維持する。幅 0〜4 でプロンプトを加工しないことは、既存テストの要件ではない
- a4: 修正範囲は `Minibuffer.View` 全体とし、`availableWidth <= 0` と `>= 1` の両方の経路で 1 行を保証する
- a5: 再発を検出するテストは単体テストだけにする（`minibuffer_test.go` と `pane_render_test.go`）。E2E テストは追加しない
- a6: 残りの幅が 1 桁で、カーソル位置の文字が全角のときは、1 行の保証を優先し、a2 に従ってその文字を描画しない
- a7: 端末ごとの曖昧幅の設定（CJK ロケールで曖昧幅を 2 桁として描画するなど）による実際の描画幅の違いは、修正範囲に含めない

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし（すべての要件が resolved）。

## References

- 要件定義書: `feature-docs/minibuffer-wide-prompt-wrap/REQUIREMENTS.md`
