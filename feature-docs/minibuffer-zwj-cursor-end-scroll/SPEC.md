# Feature: minibuffer-zwj-cursor-end-scroll

## Overview

ミニバッファの `Minibuffer.View` で、入力全体が残りの幅に収まるかどうかを、rune ごとの表示幅の合計ではなく入力文字列全体の実際の表示幅（`lipgloss.Width`）で判定する。
実際の表示幅で収まる入力は、ZWJ 連結の絵文字などの結合文字列を途中で切らずにそのまま表示する。
ミニバッファの 1 行表示（改行を含まず、表示幅 `m.width-2` 以下）は維持する。

## Objectives

- 実際の表示幅で残りの幅に収まる入力は、ZWJ 連結の絵文字などの結合文字列を途中で切らずに、ミニバッファにそのまま表示する
- ミニバッファの 1 行表示（改行を含まず、表示幅 `m.width-2` 以下）を保つ

## Acceptance Criteria

- [ ] **AC1**（FR1, FR2）: プロンプトが `(search): `（10 桁）、入力が家族の絵文字（U+1F468 U+200D U+1F469 U+200D U+1F467 U+200D U+1F466）+ `ls` で、カーソルが末尾のとき、`m.width` が 24（再現条件）でも 19〜23 でも、View の結果に `👨‍👩‍👧‍👦ls` がそのまま含まれる（これらの幅は、残りの幅が 5〜10 で、実際の幅 + 1 = 5 は収まるが、rune ごとの合計 + 1 = 11 は収まらない範囲）
- [ ] **AC2**（FR1, FR3）: AC1 と同じプロンプト・入力・幅（`m.width` 19〜24）で、カーソルが `l` または `s` にあるとき、View の結果に `👨‍👩‍👧‍👦` が途中で切れずに含まれ、`s` も含まれる
- [ ] **AC3**（FR4）: AC1 と同じプロンプトと入力で、`m.width` 18〜30 の各幅とカーソル位置 0〜9 のすべての組み合わせについて、View の結果が改行を含まず、表示幅が `m.width-2` 以下である
- [ ] **AC4**（NFR1, NFR2）: 既存のミニバッファのテストと `go test ./...` がすべて通る。変更は `minibuffer.go` と `minibuffer_test.go` だけである
- [ ] **AC5**（FR2）: 再現手順（`m.width=24`、`(search): `、家族の絵文字 + `ls`、カーソル末尾）で、先頭の code point が削られない

## Technical Requirements

### Functional Requirements

- **FR1:** 収まるかどうかは実際の表示幅で判定する — `Minibuffer.View` の残りの幅 >= 1 の経路では、入力全体が収まるかどうかを、入力文字列全体の実際の表示幅（`lipgloss.Width`）で判定する。カーソルが末尾にあるときは、これに 1 桁のカーソルブロックを足す。rune ごとの表示幅の合計は判定に使わない。
- **FR2:** カーソルが末尾のとき、収まる入力はそのまま表示する — カーソルが末尾にあり、入力全体の実際の表示幅に 1 を足した値が残りの幅以下のときは、入力の先頭から末尾までを切り詰めずに表示する。スクロール開始位置は 0 にし、先頭の code point を削らない。
- **FR3:** カーソルが末尾以外のときも、収まる入力はそのまま表示する — カーソルが末尾以外にあり、入力全体の実際の表示幅が残りの幅以下のときは、表示範囲を入力全体（開始位置 0、終了位置は入力の末尾）にする。ただし、カーソルが書記素の途中の rune にあり、カーソル位置の反転表示で結合が分かれて描画幅が増える場合は、既存の縮小ループによる 1 行の保証を優先する。（前提 a3 に基づく）
- **FR4:** 1 行の保証を維持する — 変更後も、`m.width > 4` のとき View の結果は改行を含まず、表示幅は `m.width-2` 以下である。組み立てた行を実際の表示幅で測って縮める既存の処理（D2 の縮小ループ）は残す。

### Non-Functional Requirements

- **NFR1 - 変更範囲:** 変更は `internal/ui/minibuffer.go` の `Minibuffer.View` の表示計算と、`internal/ui/minibuffer_test.go` のテスト追加に限る。`HandleKey` などの編集操作は変えない。`cursorPos` が rune 単位であることも変えない。
- **NFR2 - 既存テストの維持:** 既存のミニバッファのテスト（`TestMinibufferView`、`TestMinibufferView_NarrowWidths_NoPanicNoLineBreak`、`TestMinibufferView_WidthInputCursorTable_BoundedSingleLine`、`TestMinibufferView_BoundaryWidths_TruncationAndBound`、`TestMinibufferView_ScrolledInput_CursorCharacterVisible`、`TestMinibufferView_GraphemeClusters_BoundedSingleLine`、`TestMinibufferView_Width40_PromptAndInputFitUntruncated` など）と `go test ./...` が、変更後も変更なしで通る。

## Assumptions

- **a1:** 再現条件は `m.width=24`（内容幅 20、プロンプト 10 桁、残り 10 桁）とする。手順 3 の『残り 10 桁』に合わせたもので、手順 1 の『内容幅 24』とは食い違う
- **a2:** 対象は、実際の表示幅で入力全体が収まる場合に限る。入力が実際の幅でも収まらずスクロールが必要なときに、表示範囲の境界が書記素の途中になる挙動は、この修正に含めない
- **a3:** 収まるかどうかの判定を実際の表示幅に切り替えるのは、カーソルが末尾のときだけでなく、すべてのカーソル位置とする。期待する挙動の文にカーソル位置の限定がなく、rune ごとの幅で過大評価する原因が同じ判定にあるため
- **a4:** カーソル位置の rune だけを反転表示する既存の描画は変えない。書記素の途中にカーソルがあると結合が分かれて描画されるが、これは修正に含めない
- **a5:** 表示幅は、前の機能と同じく `lipgloss.Width`（書記素単位）で測る
- **a6:** 再発を検出するテストは、`minibuffer_test.go` の単体テストだけにする。E2E テストは追加しない
- **a7:** 既存テストで固定されている事後条件（幅 0〜4 でパニックせず改行もしない、1 行の保証、スクロール時にカーソル位置の文字が見える）を維持する

## Implementation Approach

### 対象

- `internal/ui/minibuffer.go` の `Minibuffer.View` のうち、残りの幅 >= 1 の経路の表示範囲の計算

### 表示範囲の判定

```
w = lipgloss.Width(入力全体)

カーソルが末尾:
  w + 1 <= 残りの幅  → 開始位置 0、入力の末尾まで表示        (FR1, FR2)
カーソルが末尾以外:
  w <= 残りの幅      → 開始位置 0、終了位置は入力の末尾      (FR1, FR3)
上記以外              → 既存のスクロール経路                  (a2)

組み立てた行 → 既存の縮小ループ（D2）で表示幅 m.width-2 以下に  (FR4)
```

### Dependencies

**Internal Dependencies:**
- 既存の縮小ループ（D2）: 組み立てた行を実際の表示幅で測って縮める処理。変更後も残す

**External Dependencies:**
- `lipgloss.Width`: 入力全体の実際の表示幅（書記素単位）の測定

### File Structure

```
internal/
└── ui/
    ├── minibuffer.go        # Minibuffer.View の表示計算を変更
    └── minibuffer_test.go   # テストを追加
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/minibuffer-zwj-cursor-end-scroll/**`
- `test-docs/minibuffer-zwj-cursor-end-scroll/**`

`feature-docs/minibuffer-zwj-cursor-end-scroll/**` covers `REQUIREMENTS.md`,
`SPEC.md`, `IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/minibuffer-zwj-cursor-end-scroll/**` covers
`test-docs/minibuffer-zwj-cursor-end-scroll/{T}.tests.yaml`, the per-task
test record. It is generated and owned by `implement-phase.md`; this
section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/minibuffer-zwj-cursor-end-scroll/` directory at all; the declared
`test-docs/minibuffer-zwj-cursor-end-scroll/**` entry is still correct in
that case — a declared path that never materializes is not a violation.

## Test Scenarios

### Unit Tests

対象ファイル: `internal/ui/minibuffer_test.go`

- [ ] **TS-1**（FR1, FR2, FR4 / AC1, AC5）: 表形式のテスト。プロンプト `(search): `、入力は家族の絵文字 + `ls`、カーソルは末尾で、`m.width` 19〜24 の各幅について、View の結果に `👨‍👩‍👧‍👦ls` が含まれることと、1 行に収まること（`assertBoundedSingleLine`）を確かめる。現行コードではどの幅でも失敗する
- [ ] **TS-2**（FR1, FR3 / AC2）: TS-1 と同じ条件で、カーソルを `l`（rune 7）と `s`（rune 8）に置く。View の結果に `👨‍👩‍👧‍👦` と `s` が含まれることを確かめる
- [ ] **TS-3**（FR4 / AC3）: `m.width` 18〜30 とカーソル位置 0〜9 のすべての組み合わせで、改行を含まず、表示幅が `m.width-2` 以下であることを確かめる（書記素の途中の rune にカーソルがある場合も含む）
- [ ] **TS-4**（NFR1, NFR2 / AC4）: 既存のミニバッファのテストと `go test ./...` が、変更なしで通る

### Integration Tests

なし

### E2E Tests

**Existing E2E tests**: None
**Run command**: Not detected

E2E テストは追加しない（a6）。

### Edge Cases

- [ ] 残りの幅が、実際の幅 + 1 とちょうど同じとき（`m.width=19`、残り 5）
- [ ] 残りの幅が、rune ごとの合計 + 1 よりちょうど 1 少ないとき（`m.width=24`、残り 10）
- [ ] カーソルが書記素の途中の rune（ZWJ を含む）にあるとき。反転表示で結合が分かれるので、1 行の保証が優先される
- [ ] 実際の幅でも収まらないとき（`m.width=18`、残り 4）は、既存のスクロール経路に入る

## Security Considerations

なし

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Code review is completed

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし

## References

- `internal/ui/minibuffer.go`
- `internal/ui/minibuffer_test.go`
