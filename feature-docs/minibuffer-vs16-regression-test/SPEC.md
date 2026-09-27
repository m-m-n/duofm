# Feature: minibuffer-vs16-regression-test

## Overview

ミニバッファで VS16（U+FE0F）付き絵文字の表示幅を過小評価して 2 行に折り返す不具合について、再発を検出する回帰テストを `internal/ui/minibuffer_test.go` に追加する。
入力側とプロンプト側の両方で、VS16 付き絵文字を含む場合にミニバッファが 1 行で描画されること、および表示内容が正しいことを検査する。
実装（`internal/ui/minibuffer.go`）は変更しない。

## Objectives

- ミニバッファで VS16（U+FE0F）付き絵文字の表示幅を過小評価して 2 行に折り返す不具合が再発したとき、テストで検出できるようにする
- VS16 付き絵文字を含む入力・プロンプトでもミニバッファが 1 行で描画されることをテストで保証する

## Technical Requirements

### Functional Requirements

- **FR1:** 再現条件での 1 行描画の検査 — m.width 37、プロンプト `(reverse-i-search)'日本語': `、入力 U+2764 U+FE0F を 4 組、カーソル末尾の条件で、`Minibuffer.View()` の結果が改行を含まず、表示幅（`lipgloss.Width`）が m.width-2 以下であることを検査するテストを `internal/ui/minibuffer_test.go` に追加する。
- **FR2:** 再現条件での表示内容の検査 — FR1 と同じ条件で、ANSI を除いた表示テキストがプロンプト全体の直後に U+2764 U+FE0F の組を 2 組含むこと、および U+2764 と U+FE0F が切り離されていないこと（U+2764 の直後に必ず U+FE0F があり、U+2764 が直前にない U+FE0F がないこと）を検査する。
- **FR3:** 前提値の検査 — 追加テストは最初に前提値を検査し、ずれていれば `t.Fatalf` で止める。前提値はプロンプトの表示幅 28、U+2764 U+FE0F の組の表示幅 2、その組のルーンごとの幅の合計 1 とする。
- **FR4:** 幅とカーソル位置の網羅 — FR1 と同じプロンプト・入力で、m.width 24〜41 のすべてとカーソル位置 0〜8（VS16 のルーン上を含む）のすべての組み合わせについて、`View()` が panic せず、改行を含まず、表示幅が m.width-2 以下であることを検査する。
- **FR5:** プロンプトに VS16 付き絵文字を含む場合の検査 — プロンプト `(reverse-i-search)'❤️❤️❤️❤️': `（U+2764 U+FE0F を 4 組、表示幅 30）、入力なしで、m.width 24〜34（プロンプト切り詰め経路）のすべてについて改行を含まず表示幅が m.width-2 以下であることを検査する。加えて m.width 29 では表示テキストが `(reverse-i-search)'❤️❤️❤️` を含み、U+2764 と U+FE0F が切り離されていないことを検査する。

### Non-Functional Requirements

- **NFR1:** テストのみの変更 — `internal/ui/minibuffer.go` の実装は変更しない。変更は `internal/ui/minibuffer_test.go` へのテスト追加に限る。
- **NFR2:** 既存の書き方に合わせる — テスト名は `TestMinibufferView_{Scenario}` の形にし、既存ヘルパー（`assertBoundedSingleLine`、`mustNotPanic`、`stripANSI`）を使う。新しい依存は追加しない。gofmt 済みであること。
- **NFR3:** 既存テストへの影響なし — `go test ./...` が全件成功する。

## Acceptance Criteria

- [ ] **AC1**（FR1, FR2, FR3, FR4, FR5）: 追加したテストが現行実装に対して `go test ./internal/ui/...` で成功する。
- [ ] **AC2**（FR1）: 再現手順の条件（m.width 37、プロンプト `(reverse-i-search)'日本語': `、U+2764 U+FE0F を 4 組、カーソル末尾）で `View()` の結果が 1 行に収まることをテストが検査している。
- [ ] **AC3**（FR2, FR5）: `partitionUnits` が単位の幅を単位全体の文字列ではなくルーンごとの幅の合計で見積もるように一時的に変えると、FR2 のテスト（入力側）と FR5 の m.width 29 の内容検査（プロンプト側）が失敗する。この変更はコミットしない。
- [ ] **AC4**（FR1）: AC3 の変更に加えて `guardFit` が常に最初の結果をそのまま返すように一時的に変えると、FR1 のテストが失敗する（改行が入るか、表示幅が m.width-2 を超える）。この変更はコミットしない。
- [ ] **AC5**（NFR2, NFR3）: `go test ./...` が全件成功し、gofmt による差分がない。

## Implementation Approach

### File Structure

```
internal/ui/
├── minibuffer.go         # 変更しない
└── minibuffer_test.go    # テストを追加する
```

### Dependencies

**Internal Dependencies:**
- `internal/ui/minibuffer_test.go` の既存ヘルパー: `assertBoundedSingleLine`、`mustNotPanic`、`stripANSI`

**External Dependencies:**
- 新しい依存は追加しない

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/minibuffer-vs16-regression-test/**`
- `test-docs/minibuffer-vs16-regression-test/**`

`feature-docs/minibuffer-vs16-regression-test/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/minibuffer-vs16-regression-test/**` covers `test-docs/minibuffer-vs16-regression-test/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/minibuffer-vs16-regression-test/` directory at all; the declared
`test-docs/minibuffer-vs16-regression-test/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests

- [ ] **TS1: 再現条件・カーソル末尾**（FR1, FR2, FR3）
    - 手順: プロンプト `(reverse-i-search)'日本語': `、`SetWidth(37)`、`SetInput("❤️❤️❤️❤️")`、カーソル末尾、`Show()` の後に `View()` を呼ぶ
    - 期待: 改行なし、表示幅 35 以下、表示テキストが `(reverse-i-search)'日本語': ❤️❤️` を含み、U+2764 と U+FE0F が切り離されていない
- [ ] **TS2: 幅 × カーソル位置の表**（FR4）
    - 手順: TS1 と同じプロンプト・入力で m.width 24〜41、カーソル位置 0〜8 をすべて試す
    - 期待: すべての組み合わせで panic せず、改行なし、表示幅 m.width-2 以下
- [ ] **TS3: プロンプト切り詰め経路の幅上限**（FR5）
    - 手順: プロンプト `(reverse-i-search)'❤️❤️❤️❤️': `、入力なしで m.width 24〜34 をすべて試す
    - 期待: すべての幅で改行なし、表示幅 m.width-2 以下
- [ ] **TS4: プロンプト切り詰め経路の表示内容**（FR5）
    - 手順: TS3 のプロンプトで m.width 29 の `View()` を呼ぶ
    - 期待: 表示テキストが `(reverse-i-search)'❤️❤️❤️` を含み、U+2764 と U+FE0F が切り離されていない

### E2E Tests

**Existing E2E tests**: None
**Run command**: Not detected

- E2E テストは追加しない。表示幅の検査は単体テストで行う（A5）。

## Assumptions

- **A1:** 現行実装は再現条件で 1 行に描画する（コード上の計算: contentWidth 33、プロンプト幅 28、残り幅 5。カーソル末尾の経路で単位ごとの見積もり幅 2 の組を末尾から 2 組残し、28+4+1=33 で収まる）。よって本機能はテスト追加のみで、実装変更は不要とした。
- **A2:** 1 行・幅上限の検査だけでは、単位幅の見積もりが退行しても `guardFit` の縮小で 1 行に収まるため退行を検出できない（再現条件では組が 0 個表示になる）。そのため表示内容（組 2 組）も検査対象に含めた。
- **A3:** タスク記述の「入力・プロンプト」に従い、プロンプト側（プロンプト切り詰め経路）の検査も範囲に含めた。m.width 29 の期待値はコード上の計算（contentWidth 25、ASCII 19 桁 + 組 3 組 = 25）による。
- **A4:** lipgloss v1.1.0 の幅計算で U+2764 単体は 1、U+FE0F 単体は 0、U+2764 U+FE0F の組は 2 とした（既存テストが U+00A9 U+FE0F で同じ性質を前提にしている）。FR3 の前提値検査でこれを確認する。
- **A5:** E2E テストは追加しない。表示幅の検査は単体テストで行う。

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] AC1〜AC5 を満たす

## Open Questions

なし

## References

- 対象実装: `internal/ui/minibuffer.go`
- 対象テスト: `internal/ui/minibuffer_test.go`
