# Feature: minibuffer-shrink-loop-linear

## Overview

Minibuffer.View のプロンプトの切り詰め、入力の表示範囲の決定、行の組み立てを、プロンプト長 + 入力長に対して線形の時間で終える。結合文字を大量に含む入力やプロンプトでも、ミニバッファの再描画で duofm が止まらないようにする。要件の詳細は [REQUIREMENTS.md](REQUIREMENTS.md) を参照。

## Objectives

- 結合文字を大量に含む入力やプロンプトでも、ミニバッファの再描画がプロンプト長 + 入力長に対して線形の時間で終わり、duofm が操作できる状態のまま保たれる
- レビュー指摘 408ffc35dd15e043（minibuffer-wide-prompt-wrap review round1、未解決）を解消する

## User Stories

### US1: 結合文字を大量に含む文字列をミニバッファに入力する
duofm の利用者として、結合文字を大量に含む文字列をミニバッファに入力しても、再描画がプロンプト長 + 入力長に対して線形の時間で終わり、duofm を操作し続けたい。

**Acceptance Criteria:**
- [ ] AC1: ミニバッファ幅 80、プロンプト「!: 」、入力が U+0301 ×100000 + ASCII ×71 + U+00A9 U+FE0F、カーソルが末尾のとき、View が 10 秒以内に終わる（FR1, FR4）
- [ ] AC2: ミニバッファ幅 40（端末幅 80）、同じプロンプト、入力が U+0301 ×100000 + ASCII ×31 + U+00A9 U+FE0F、カーソルが末尾のとき、View が 10 秒以内に終わる（FR1, FR4）
- [ ] AC3: カーソルが末尾以外にある経路（TS3）と、プロンプトの切り詰めの経路（TS4）の入力でも、View が 10 秒以内に終わる（FR1, FR4）
- [ ] AC4: AC1〜AC3 の入力で、View の結果が改行を含まず、表示幅が m.width-2 以下（FR3）
- [ ] AC5: TS1〜TS4 のテストは、変更前の実装では上限時間を超えて失敗する（実測で確かめる）
- [ ] AC6: go test -race で実測した TS1〜TS4 の所要時間が、10 秒の上限を十分に下回る（NFR2）
- [ ] AC7: internal/ui/minibuffer_test.go と internal/ui/pane_render_test.go の既存テストが通る（FR2, FR3）
- [ ] AC8: go test ./... がすべて通る

## Technical Requirements

### Functional Requirements
- **FR1:** View 全体の線形時間 — Minibuffer.View は、プロンプトの切り詰め、入力の表示範囲の決定、行の組み立てを、プロンプト長 + 入力長に対して線形の時間で終える。カーソルが末尾にある経路、末尾以外の経路、プロンプトだけで内容幅を使い切る経路（プロンプトの切り詰め）のすべてが対象
- **FR2:** 書記素単位の表示範囲 — 表示範囲とプロンプトの切り詰め位置は、書記素単位で決めてよい。書記素の途中で切れていた境界のケースでは、表示内容が現行と変わってよい
- **FR3:** 既存の事後条件の維持 — m.width > 4 のとき、View の結果は改行を含まず、表示幅は m.width-2 以下。カーソル位置の文字（末尾ならカーソルブロック）が表示範囲に入る。プロンプトと入力が内容幅に収まるときは切り詰めない。m.width が 0〜4 のときはパニックせず、改行を含まない
- **FR4:** 再発を検出するテスト — `internal/ui/minibuffer_test.go` に、TS1〜TS4 の入力で View を goroutine で実行し、10 秒以内に終わらなければ失敗にするテストを置く。終わったときは、結果が改行を含まず、表示幅が m.width-2 以下であることも確かめる

### Non-Functional Requirements
- **NFR1 - 変更範囲:** 変更は `internal/ui/minibuffer.go` の View とその補助関数、`internal/ui/minibuffer_test.go` に限る。HandleKey などの編集操作と、cursorPos が rune 単位であることは変えない。本体にテスト用の差し替え口（幅の計測関数の注入など）は作らない
- **NFR2 - 上限時間の余裕:** TS1〜TS4 の各入力で、変更後の View の所要時間を `go test -race` で実測する。10 秒の上限を十分に下回ることを確かめ、その結果を記録する

### Assumptions
- **a1:** HandleKey などの編集操作と、cursorPos が rune 単位であることは変えない
- **a2:** 既存テストが確認している事後条件（1 行、幅の上限、カーソルが見える、収まる内容はそのまま、幅 0〜4 でパニックしない）を維持する
- **a3:** 表示幅の基準は引き続き lipgloss.Width（書記素単位）とし、go-runewidth は判定の基準に使わない
- **a4:** 再発を検出するテストは Go の単体テストだけにし、E2E テストは追加しない
- **a5:** レビュー指摘 abb435ae05afd8f2（カーソルが末尾のときのスクロールで、ZWJ 入力の先頭が削られる）は対象外
- **a6:** 端末幅 80 のとき、ミニバッファ幅はペイン幅の 40 になる。再現手順の「幅 80」は、ミニバッファ幅 80（端末幅 160 相当）とミニバッファ幅 40（端末幅 80）の両方で扱う

## Implementation Approach

### Architecture

変更対象は `internal/ui/minibuffer.go` の Minibuffer.View とその補助関数に限る（NFR1）。画面要素や見た目の追加・変更はない。

**Component Diagram:**
```
Minibuffer.View
├── プロンプトの切り詰め（FR1, FR2）
├── 入力の表示範囲の決定（FR1, FR2）
└── 行の組み立て（FR1）
```

### Data Flow

```
プロンプト + 入力 + cursorPos + m.width → Minibuffer.View → 1 行の表示文字列（FR3）
```

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- Minibuffer の編集操作（HandleKey など）: 変更しない。cursorPos は rune 単位のまま（NFR1, a1）

**External Dependencies:**
- lipgloss.Width: 表示幅の基準（書記素単位）。go-runewidth は判定の基準に使わない（a3）

### File Structure

```
internal/
└── ui/
    ├── minibuffer.go            # View とその補助関数（変更対象）
    ├── minibuffer_test.go       # TS1〜TS4 を追加、既存テストを維持
    └── pane_render_test.go      # 既存テストを維持（変更しない）
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/minibuffer-shrink-loop-linear/**`
- `test-docs/minibuffer-shrink-loop-linear/**`

`feature-docs/minibuffer-shrink-loop-linear/**` covers `REQUIREMENTS.md`,
`SPEC.md`, `IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/minibuffer-shrink-loop-linear/**` covers
`test-docs/minibuffer-shrink-loop-linear/{T}.tests.yaml`, the per-task test
record. It is generated and owned by `implement-phase.md`; this section
cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/minibuffer-shrink-loop-linear/` directory at all; the declared
`test-docs/minibuffer-shrink-loop-linear/**` entry is still correct in that
case — a declared path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS1: ミニバッファ幅 80、プロンプト「!: 」、入力 U+0301 ×100000 + 'a' ×71 + U+00A9 U+FE0F、カーソルが末尾。View を goroutine で実行し、10 秒以内に終わること、改行を含まないこと、表示幅が 78 以下であることを確かめる（`internal/ui/minibuffer_test.go`。AC1, AC4, AC5）
- [ ] TS2: ミニバッファ幅 40、プロンプト「!: 」、入力 U+0301 ×100000 + 'a' ×31 + U+00A9 U+FE0F、カーソルが末尾。TS1 と同じことを確かめる（表示幅の上限は 38）（`internal/ui/minibuffer_test.go`。AC2, AC4, AC5）
- [ ] TS3: カーソルが末尾以外の経路。例: ミニバッファ幅 40、プロンプト「!: 」、入力 U+00A9 U+FE0F + 'a' ×32 + U+0301 ×100000、カーソルは rune 位置 2（最初の 'a'）。現行実装では、rune ごとの幅の予算で幅 0 の文字の並び全体が範囲に入り、縮小ループが右端から 1 rune ずつ削る。TS1 と同じことを確かめる（`internal/ui/minibuffer_test.go`。AC3, AC4, AC5）
- [ ] TS4: プロンプトだけで内容幅を使い切る経路。例: ミニバッファ幅 40、プロンプト `"(reverse-i-search)'"` + U+00A9 U+FE0F + 'a' ×16 + U+0301 ×100000 + `"': "`、入力は空。現行実装では、プロンプトの切り詰めループが幅 0 の文字を 1 rune ずつ削る。TS1 と同じことを確かめる（`internal/ui/minibuffer_test.go`。AC3, AC4, AC5）
- [ ] TS5: 既存テスト（幅 0〜4、全角プロンプト・入力、ZWJ・結合文字、カーソル位置の表、幅 40 でそのまま表示）が変更後も通る（`internal/ui/minibuffer_test.go`、`internal/ui/pane_render_test.go`。AC7, AC8）

### Integration Tests
該当なし

### E2E Tests
**Existing E2E tests**: None
**Run command**: Not detected
- E2E テストは追加しない（a4）

### Edge Cases
- [ ] 入力の先頭に幅 0 の文字が長く並び、末尾に VS16 付きの絵文字がある（カーソルは末尾）
- [ ] 収まる範囲の最後に幅 0 の文字が長く並ぶ（カーソルは末尾以外）
- [ ] プロンプト（履歴検索のパターン）に幅 0 の文字が長く並び、VS16 で実際の幅が rune ごとの幅の合計を超える
- [ ] ZWJ で連結した絵文字、全角文字、書記素の途中にある rune 上のカーソル（既存テストの範囲）
- [ ] m.width 0〜4、プロンプトの幅が内容幅とちょうど同じ・1 桁少ない・1 桁多い
- [ ] 変更前の実装で上限時間を超えたとき、goroutine は動き続けるが、テストは上限時間で失敗として終わる

### Performance Tests
- [ ] TS6: `go test -race` で TS1〜TS4 を実行し、それぞれの View の所要時間を記録する（AC6）

## Security Considerations

該当なし

## Error Handling

- m.width が 0〜4 のとき、View はパニックせず、改行を含まない結果を返す（FR3）

## Performance Optimization

### Performance Goals
- View の処理時間: プロンプト長 + 入力長に対して線形（FR1）
- TS1〜TS4 の各入力での View の所要時間: `go test -race` で 10 秒の上限を十分に下回る（NFR2）

### Optimization Strategies
- 表示範囲とプロンプトの切り詰め位置を書記素単位で決めてよい（FR2）

### Caching Strategy
該当なし

## Success Criteria

- [ ] FR1〜FR4、NFR1〜NFR2 を満たす
- [ ] AC1〜AC8 を満たす
- [ ] TS1〜TS6 がすべて通る、または記録される

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし

## Implementation Phases (if applicable)

該当なし

## References

- 要件定義書: [REQUIREMENTS.md](REQUIREMENTS.md)
