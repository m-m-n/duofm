# Feature: minibuffer-grapheme-scroll-boundary

## Overview

ミニバッファの入力が残りの幅に収まらずスクロールするとき、表示範囲の開始位置と終了位置を書記素の境目に揃える。ZWJ 連結の絵文字などの結合文字列を途中で切らずに表示し、1 行表示を保つ。

## Objectives

- ミニバッファの入力が残りの幅に収まらずスクロールするときも、表示範囲の開始位置と終了位置を書記素の境目に揃え、ZWJ 連結の絵文字などの結合文字列を途中で切らずに表示する
- ミニバッファの 1 行表示（改行を含まず、表示幅 m.width-2 以下）を保つ

## Acceptance Criteria

- [ ] **AC1** (FR1, FR4): プロンプト `(search): `（10 桁）、m.width=24（残り 10 桁）、入力 `aaaaa` + 家族の絵文字（U+1F468 U+200D U+1F469 U+200D U+1F467 U+200D U+1F466）+ `bbbbb`、カーソル末尾のとき、ANSI を除いた View の結果に `👨‍👩‍👧‍👦bbbbb` が含まれ、1 行に収まる
- [ ] **AC2** (FR2, FR4): AC1 と同じプロンプト・幅・入力で、カーソルが位置 0 のとき、ANSI を除いた View の結果に `aaaaa👨‍👩‍👧‍👦` が含まれ、1 行に収まる
- [ ] **AC3** (FR1, FR2, FR3, FR4): AC1 と同じプロンプトと入力で、m.width 15〜27 の各幅とカーソル位置 0〜17 のすべての組み合わせについて、ANSI を除いた View の結果が、家族の絵文字の結合文字列全体を含むか、U+1F468・U+1F469・U+1F467・U+1F466・U+200D のどれも含まないかのどちらかであり、改行を含まず、表示幅が m.width-2 以下である
- [ ] **AC4** (NFR1, NFR2): 既存のミニバッファのテストと go test ./... がすべて変更なしで通る。変更は minibuffer.go と minibuffer_test.go だけである（前提 a6 の例外を除く）

## Technical Requirements

### Functional Requirements

- **FR1:** スクロール時の開始位置は書記素の境目にする。Minibuffer.View の残りの幅 >= 1 の経路で入力全体が収まらずスクロールするとき、表示範囲の開始位置（displayStart）は常に書記素の境目にする。書記素の途中の rune から表示を始めない。書記素の区切りは表示幅の基準（lipgloss.Width、書記素単位）と同じものを使う。
- **FR2:** スクロール時の終了位置は書記素の境目にする。FR1 と同じ条件で、表示範囲の終了位置（displayEnd）は常に書記素の境目にする。書記素の途中の rune で表示を終えない。
- **FR3:** カーソルを含む書記素は丸ごと扱う（status: assumed、前提 a3 に基づく）。カーソルが書記素の途中の rune（ZWJ などを含む）にあるときも、表示範囲はカーソルを含む書記素全体を含める。その書記素の表示幅が残りの幅を超えるときは、既存のカーソル位置の文字が収まらない場合（CD3）と同じく入力を表示しない。
- **FR4:** 1 行の保証と既存の事後条件を維持する。変更後も、m.width > 4 のとき View の結果は改行を含まず、表示幅は m.width-2 以下である。組み立てた行を実際の表示幅で測って縮める既存の処理（guardFit）は残し、縮めるときも書記素単位で取り除く。スクロール時に、表示幅が許す限りカーソル位置の文字（カーソル末尾のときは入力の最後の文字）が見えることも維持する。

### Non-Functional Requirements

- **NFR1 - 変更範囲:** 変更は internal/ui/minibuffer.go の Minibuffer.View の表示範囲の計算（および、それが使う単位分割の補助関数）と、internal/ui/minibuffer_test.go のテスト追加に限る。HandleKey などの編集操作と、cursorPos が rune 単位であることは変えない。（前提 a6）
- **NFR2 - 既存テストの維持:** 既存のミニバッファのテスト（TestMinibufferView_ScrolledInput_CursorCharacterVisible、TestMinibufferView_GraphemeClusters_BoundedSingleLine、TestMinibufferView_ZWJInput* 系、TestMinibufferView_LinearTime_TS1〜TS4、TestMinibufferView_TS7_JoinedSequenceGuard、TestMinibufferView_VS16* 系など）と go test ./... が、変更なしで通る。LinearTime 系のテストが固定している、入力長に対して線形の時間で終わる性質を保つ。

## Implementation Approach

### Change Scope

- `internal/ui/minibuffer.go`: Minibuffer.View の表示範囲の計算と、それが使う単位分割の補助関数（NFR1）
- `internal/ui/minibuffer_test.go`: テスト追加（NFR1）
- `go.mod` / `go.sum`: 書記素の区切りのために間接依存の github.com/rivo/uniseg を直接使う場合に限り、go.mod の `// indirect` 表記の削除と、それに伴う go.sum の変更を許す（前提 a6）

### Dependencies

**External Dependencies:**
- github.com/rivo/uniseg: 既存の間接依存。書記素の区切りのために直接使う場合がある（前提 a6）

## Declared Change Set

上記の機能固有のパスは、手書きの一覧ではなく create-plan で導出する。`workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

機能固有のパスに加え、次の 2 つのワークフロー生成物を既定で宣言する。

- `feature-docs/minibuffer-grapheme-scroll-boundary/**`
- `test-docs/minibuffer-grapheme-scroll-boundary/**`

`feature-docs/minibuffer-grapheme-scroll-boundary/**` は `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、design の成果物を含む。これらはフェーズ文書と `references/phase-state.md` が生成・所有し、本節はその規則を再掲しない。

`test-docs/minibuffer-grapheme-scroll-boundary/**` はタスクごとのテスト記録 `test-docs/minibuffer-grapheme-scroll-boundary/{T}.tests.yaml` を含む。これは `implement-phase.md` が生成・所有し、本節はその規則を再掲しない。

この 2 つの既定の項目は、明示的に取り除かない限り宣言に含まれる。

この宣言は上位集合としての宣言である。verify 時に観測される実際の変更集合は、宣言した集合に含まれていればよく、一致する必要はない。実装タスクを生まない機能では `test-docs/minibuffer-grapheme-scroll-boundary/` は作られないが、宣言した項目が実体化しないことは違反ではない。

## Test Scenarios

### Unit Tests

- [ ] **TS-1** (FR1, FR4, AC1): m.width=24、`(search): `、入力 `aaaaa👨‍👩‍👧‍👦bbbbb`、カーソル末尾。ANSI を除いた結果に `👨‍👩‍👧‍👦bbbbb` が含まれることと assertBoundedSingleLine を確かめる。現行コードでは開始位置が rune 9 になり `👧‍👦bbbbb` と表示されるため失敗する
- [ ] **TS-2** (FR2, FR4, AC2): TS-1 と同じ条件でカーソルを位置 0 に置く。ANSI を除いた結果に `aaaaa👨‍👩‍👧‍👦` が含まれることを確かめる。現行コードでは終了位置が rune 9 になり `aaaaa👨‍👩‍` で切れるため失敗する
- [ ] **TS-3** (FR1, FR2, FR3, FR4, AC3): 表形式のテスト。m.width 15〜27 とカーソル位置 0〜17 のすべての組み合わせで、家族の絵文字が丸ごと含まれるか部品が一つも含まれないかのどちらかであることと、1 行に収まることを確かめる（カーソルが絵文字の途中の rune にある位置 5〜11 を含む）
- [ ] **TS-4** (AC1, AC2, AC3): 前提の確認。プロンプトの表示幅 10、入力の rune 数 17、入力全体の実際の表示幅 12、家族の絵文字の表示幅 2 であることを最初に確かめ、崩れていれば t.Fatalf で止める（既存の checkFamilyEmojiFixture と同じ形）
- [ ] **TS-5** (NFR1, NFR2, AC4): 既存のミニバッファのテストと go test ./... が変更なしで通る

### Integration Tests

なし

### E2E Tests

**Existing E2E tests**: なし
**Run command**: Not detected

E2E テストは追加しない（前提 a5）。

### Edge Cases

- [ ] カーソルが絵文字の最初の rune（位置 5）、ZWJ（位置 6 など）、最後の rune（位置 11）にあるとき
- [ ] 残りの幅が 1（m.width=15）で、表示幅 2 の絵文字が入らないとき
- [ ] 絵文字がちょうど表示範囲の左端・右端にかかる幅のとき（TS-3 の掃引で網羅）
- [ ] 結合文字（e + U+0301）や VS16 の組は現行でも一まとまりで扱われており、変更後も分かれない（既存テストで確認）
- [ ] ZWJ 以外で正の幅の rune が結合する書記素（国旗の地域指示子の組、肌色の修飾子など）にも FR1・FR2 が同じく適用される

## Assumptions

- **a1:** 書記素の区切りは、表示幅の基準と同じく lipgloss.Width（書記素単位）が使う区切りに揃える
- **a2:** カーソル位置の rune だけを反転表示する既存の描画は変えない。書記素の途中にカーソルがあると反転表示で結合が分かれて見えるが、これは修正に含めない（前の機能の前提 a4 と同じ）
- **a3:** カーソルが書記素の途中にあるときも、表示範囲の境界は書記素の境目に揃え、カーソルを含む書記素全体を表示範囲に含める。期待する挙動の文にカーソル位置の限定がないため
- **a4:** 入力全体が実際の幅で収まる場合の経路（前の機能 minibuffer-zwj-cursor-end-scroll の FR1〜FR3）は変えない
- **a5:** 再発を検出するテストは minibuffer_test.go の単体テストだけにする。E2E テストは追加しない
- **a6:** 変更範囲は minibuffer.go と minibuffer_test.go とする。ただし書記素の区切りのために間接依存の github.com/rivo/uniseg を直接使う場合に限り、go.mod の `// indirect` 表記の削除と、それに伴う go.sum の変更を許す
- **a7:** 再現条件は m.width=24（内容幅 20、プロンプト `(search): ` 10 桁、残り 10 桁）、入力 `aaaaa` + 家族の絵文字 + `bbbbb` とする（手順 2 の例に合わせる）
- **a8:** 既存テストで固定されている事後条件（幅 0〜4 でパニックせず改行もしない、1 行の保証、スクロール時にカーソル位置の文字が見える、入力長に対して線形の時間で終わる）を維持する

## Success Criteria

- [ ] すべての機能要件が実装され、テストされている
- [ ] すべてのテストシナリオが通る

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし
