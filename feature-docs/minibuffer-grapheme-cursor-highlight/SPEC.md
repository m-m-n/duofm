# Feature: minibuffer-grapheme-cursor-highlight

## Overview

ミニバッファのカーソル移動を書記素単位にし、カーソルを含む書記素全体を 1 つの反転範囲として描画する。カーソルが書記素（ZWJ 連結の絵文字など）の途中にあっても、結合文字列を分けて描画せず、収まる入力を切り詰めない。編集操作は既存の rune 単位のまま変えない。

要件定義書: `feature-docs/minibuffer-grapheme-cursor-highlight/REQUIREMENTS.md`

## Objectives

- ミニバッファのカーソルが書記素（ZWJ 連結の絵文字など）の途中にあっても、結合文字列を分けて描画しない
- 結合の分割で描画幅が増え、収まっていた入力が縮小ループで切り詰められることをなくす

## User Stories

### US1: 書記素単位のカーソル移動
ミニバッファの利用者として、家族絵文字などの書記素の上では Left/Right 1 回で書記素 1 つを越えたい。

**Acceptance Criteria:**
- [ ] AC1: familyEmojiLsInput（家族絵文字 7 rune + "ls"）で、カーソル末尾（9）から Left を押すと 8, 7, 0 の順に移動する。Ctrl+B でも同じ。
- [ ] AC2: familyEmojiLsInput で、カーソル 0 から Right を押すと 7, 8, 9 の順に移動する。Ctrl+F でも同じ。
- [ ] AC3: familyEmojiLsInput で SetCursorPos(k)（k = 1..6、書記素の途中）の後、Left で 0、Right で 7 に移動する。
- [ ] AC4: 結合文字（e + U+0301）、VS16（❤️ = U+2764 U+FE0F）、国旗（地域指示記号 2 つ）の各書記素について、Left/Right が書記素 1 つずつ移動する。

### US2: 結合文字列を分けない描画
ミニバッファの利用者として、カーソルが書記素の途中にあっても結合文字列が分かれずに表示され、収まる入力が切り詰められないようにしたい。

**Acceptance Criteria:**
- [ ] AC5: 描画器の色プロファイルを ANSI256 に固定した状態で、familyEmojiLsInput のカーソルが 0..6 のどの位置にあっても、View の出力で家族絵文字の 7 rune すべてが 1 つの反転範囲の中に連続して含まれ、反転範囲の外に家族絵文字の rune が現れない。カーソルが 7 または 8 のときは、その 1 rune（l または s）だけが反転する。
- [ ] AC6: 再現手順（家族絵文字を入力し、カーソルを左へ 1 つずつ動かす）のどの段階でも、View の出力（ANSI を除いた表示テキスト）に家族絵文字が分かれずに含まれる。
- [ ] AC7: familyEmojiLsInput と searchPrompt で、m.width 19..24、カーソルが家族絵文字の途中（rune 1..6）にあるとき、ANSI を除いた表示に prompt と入力全体が切り詰めずに含まれる。
- [ ] AC8: スクロールが起きる幅（familyEmojiLsInput と searchPrompt で m.width 14..18 など）でカーソルが家族絵文字の位置（0..6）にあるとき、家族絵文字は丸ごと 1 つの反転範囲で描画されるか、まったく描画されないかのどちらかで、一部の rune だけが描画されることはない。

### US3: 既存の動作と保証の維持
ミニバッファの利用者として、編集操作と表示の保証は既存のまま使いたい。

**Acceptance Criteria:**
- [ ] AC9: 既存の幅とカーソル位置の表（ZWJ・VS16・結合文字の各表）を含むすべての幅・カーソル位置の組み合わせで、View の結果は改行を含まず表示幅が m.width-2 以下である。
- [ ] AC10: 書記素の途中にカーソルがあるときの Backspace・Delete・文字挿入・Ctrl+K・Ctrl+U は、既存どおり rune 単位で動作する。
- [ ] AC11: 既存の internal/ui/minibuffer_test.go のテスト（線形時間の TS1〜TS4、TS7 を含む）がすべて通り、go vet ./... と gofmt で差分が出ない。
- [ ] AC12: 再発を検出する単体テストが internal/ui/minibuffer_test.go にあり、書記素移動（AC1〜AC4）と書記素全体の反転（AC5, AC8）を検証する。

## Technical Requirements

### Functional Requirements
- **FR1:** 書記素単位のカーソル移動 — Left と Ctrl+B は、カーソルを直前の書記素の開始位置へ移動する。カーソルが書記素の途中にある場合は、その書記素の開始位置へ移動する。Right と Ctrl+F は、カーソルを含む書記素の終了位置（次の書記素の開始位置）へ移動する。先頭での Left/Ctrl+B と末尾での Right/Ctrl+F はカーソルを動かさない。cursorPos は rune 位置のまま保持する。
- **FR2:** カーソルを含む書記素全体の反転表示 — カーソルが末尾以外にあるとき、View はカーソル位置の rune を含む書記素全体 [開始, 終了) を 1 つの反転範囲として描画する。SetCursorPos などで書記素の途中に置かれた場合も同じ。カーソルが末尾にあるときの末尾のブロックカーソルは変えない。
- **FR3:** 収まる入力の非切り詰め — prompt と入力全体（反転表示を含み、カーソルが末尾なら末尾のブロックカーソルを含む）の実際の表示幅が内容幅に収まる場合、カーソルが書記素の途中にあっても入力を切り詰めずに表示する。
- **FR4:** スクロール時のカーソルの書記素 — 入力が収まらずスクロールする経路では、カーソルを含む書記素を常に丸ごと扱う。カーソルの書記素全体の幅が残り幅を超える場合は入力を描画しない（既存のステップ3と同じ扱い）。カーソル前後の区切り位置はカーソルを含む書記素の開始位置と終了位置とする。カーソルから離れた側の表示窓の端は、既存の width unit 単位の切り方のまま変えない。
- **FR5:** 編集操作は変えない — 文字挿入・スペース挿入・Backspace・Delete・Ctrl+K・Ctrl+U は rune 単位の既存の動作のまま変えない。Ctrl+A・Ctrl+E・SetInput・SetCursorPos・CursorPos の動作（rune 位置）も変えない。

### Non-Functional Requirements
- **NFR1:** 1 行・幅上限 — View の結果は改行を含まず、表示幅は m.width-2 以下（既存の保証を維持する）
- **NFR2:** 表示幅と書記素区切りの基準 — 表示幅の基準は lipgloss.Width のまま。書記素の区切りは lipgloss.Width の幅計算と同じ書記素クラスタの区切り（rivo/uniseg）を用いる
- **NFR3:** 計算量 — View の計算量は入力長に対して線形のまま、guardFit の再測定は最大 3 回のまま維持する。書記素の区切り計算も入力長に対して線形とする
- **NFR4:** 依存 — go.mod に既にある依存（rivo/uniseg v0.4.7 を含む）以外の新しい依存を加えない

## Implementation Approach

### Architecture

**System Architecture:**

変更は `internal/ui/minibuffer.go` のミニバッファに閉じる。

**Component Diagram:**
```
Minibuffer
├── カーソル移動（Left / Ctrl+B / Right / Ctrl+F）  … FR1: 書記素単位
├── 編集操作・Ctrl+A / Ctrl+E / SetInput / SetCursorPos / CursorPos … FR5: 変更なし
└── View
    ├── 反転範囲 = カーソルを含む書記素 [開始, 終了)  … FR2
    ├── 収まる場合: 切り詰めない                      … FR3
    └── スクロール経路（ステップ3/4/5）               … FR4
```

### Data Flow

```
キー入力 → カーソル移動（書記素の開始・終了へ）→ cursorPos（rune 位置）
View → カーソルを含む書記素 [開始, 終了) を算出 → 反転範囲として描画 → 幅に合わせて表示
```

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- タブ補完の NewCursorPos（`internal/ui/model_update_keyboard.go`）: rune 位置で SetCursorPos に渡され、書記素の途中の位置も受け付ける

**External Dependencies:**
- rivo/uniseg v0.4.7: 書記素の区切り。go.mod に indirect で入っており、直接使う場合は go.mod の indirect 指定を外す
- lipgloss: 表示幅の基準（lipgloss.Width）と反転表示（Reverse）

### File Structure

```
internal/ui/
├── minibuffer.go        # カーソル移動と View の反転表示
└── minibuffer_test.go   # 再発を検出する単体テスト
go.mod                   # rivo/uniseg の indirect 指定
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/minibuffer-grapheme-cursor-highlight/**`
- `test-docs/minibuffer-grapheme-cursor-highlight/**`

`feature-docs/minibuffer-grapheme-cursor-highlight/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/minibuffer-grapheme-cursor-highlight/**` covers `test-docs/minibuffer-grapheme-cursor-highlight/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/minibuffer-grapheme-cursor-highlight/` directory at all; the declared
`test-docs/minibuffer-grapheme-cursor-highlight/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS1: 家族絵文字上の Left/Ctrl+B 移動 - familyEmojiLsInput、カーソル 9 から Left を 3 回押し、各回の cursorPos が 8, 7, 0 であることを確認する。Ctrl+B でも同様に確認する。（AC1 / FR1）
- [ ] TS2: 家族絵文字上の Right/Ctrl+F 移動 - familyEmojiLsInput、カーソル 0 から Right を 3 回押し、各回の cursorPos が 7, 8, 9 であることを確認する。Ctrl+F でも同様に確認する。（AC2 / FR1）
- [ ] TS3: 書記素の途中からの移動 - 表駆動で SetCursorPos(1..6) の後、Left で 0、Right で 7 になることを確認する。（AC3 / FR1）
- [ ] TS4: 種類の異なる書記素での移動 - e + U+0301、❤️、国旗（例: U+1F1EF U+1F1F5）をそれぞれ含む入力で、Left/Right が書記素 1 つずつ移動することを表駆動で確認する。（AC4 / FR1）
- [ ] TS5: 書記素全体の反転表示 - lipgloss の描画器の色プロファイルを ANSI256 に固定し t.Cleanup で戻す。familyEmojiLsInput と searchPrompt、内容が収まる幅で、カーソル 0..8 の各位置の View 出力から反転範囲を取り出し、0..6 では家族絵文字 7 rune 全体、7 では l、8 では s であることを確認する。（AC5, AC6 / FR1, FR2）
- [ ] TS6: 書記素の途中のカーソルで切り詰めない - familyEmojiLsInput と searchPrompt、m.width 19..24、カーソル 1..6 で、ANSI を除いた表示に prompt + 入力全体が含まれることを確認する。（AC7 / FR3）
- [ ] TS7: スクロール経路でのカーソルの書記素 - ANSI256 固定のもとで、familyEmojiLsInput と searchPrompt、スクロールが起きる幅（m.width 14..18）とカーソル 0..6 の組み合わせで、家族絵文字が丸ごと 1 つの反転範囲で出るか、まったく出ないかのどちらかであることを確認する。（AC8 / FR4）
- [ ] TS8: 1 行・幅上限の維持 - 既存の表駆動テスト（ZWJ・VS16・結合文字）に加え、国旗・結合文字の入力でも全幅・全カーソル位置で assertBoundedSingleLine が通ることを確認する。（AC9 / NFR1）
- [ ] TS9: 編集操作が rune 単位のまま - familyEmojiLsInput で SetCursorPos(3) の後、Backspace で 1 rune だけ削除され cursorPos が 2 になること、Delete で 1 rune だけ削除されることを確認する。既存の編集操作テストが通ることも確認する。（AC10, AC11 / FR5, NFR1, NFR2, NFR3, NFR4）
- [ ] TS10: 既存テストと静的検査 - go test ./...、go vet ./...、gofmt で差分なしを確認する。（AC11, AC12 / FR1, FR2, FR4, FR5, NFR1, NFR2, NFR3, NFR4）

### Integration Tests
該当なし

### E2E Tests
**Existing E2E tests**: None
**Run command**: Not detected
- 再発を検出するテストは internal/ui/minibuffer_test.go の単体テストで満たし、E2E テストは追加しない

### Edge Cases
- [ ] SetCursorPos（タブ補完の NewCursorPos）で書記素の途中に置かれたカーソル
- [ ] 結合文字（e + U+0301）、VS16（❤️・©️）、国旗（地域指示記号 2 つ）、ZWJ 連結など種類の異なる書記素
- [ ] 入力先頭の孤立した幅 0 の rune（先頭の結合文字や ZWJ）は、それ自体を 1 つの書記素として移動・反転する
- [ ] カーソルの書記素全体の幅が残り幅を超える場合（ステップ3）は入力を描画しない
- [ ] スクロール経路（ステップ4/5）でカーソルの書記素の開始・終了を前後の区切りに使う
- [ ] 書記素の途中での文字挿入・Backspace・Delete・Ctrl+K・Ctrl+U は rune 単位のまま（挿入後にカーソルが書記素の途中になっても反転は書記素全体）
- [ ] 空入力、カーソルが先頭・末尾

### Performance Tests
- [ ] 既存の線形時間テスト（internal/ui/minibuffer_test.go の TS1〜TS4、TS7）がすべて通る（NFR3）

## Security Considerations

該当なし

## Error Handling

該当なし

## Performance Optimization

### Performance Goals
- View の計算量は入力長に対して線形のまま、guardFit の再測定は最大 3 回のまま（NFR3）
- 書記素の区切り計算も入力長に対して線形（NFR3）

## Assumptions

- A1: cursorPos・SetCursorPos・CursorPos は rune 位置のまま扱う。タブ補完の NewCursorPos（model_update_keyboard.go）は rune 位置で渡され、書記素の途中の位置も受け付ける。
- A2: 書記素の区切りには go.mod に indirect で入っている rivo/uniseg v0.4.7 を使う。直接使う場合は go.mod の indirect 指定を外す。
- A3: 反転の検証テストでは、lipgloss の描画器の色プロファイルを ANSI256 に固定し、t.Cleanup で元に戻す。
- A4: カーソルから離れた側の表示窓の端は、既存の width unit 単位の切り方のままとし、書記素の境界に合わせない。
- A5: 文字挿入・Backspace・Delete・Ctrl+K・Ctrl+U は rune 単位のまま変えない。
- A6: 「再発を検出するテスト」は internal/ui/minibuffer_test.go の単体テストで満たし、E2E テストは追加しない。
- A7: 反転の見た目（lipgloss の Reverse）と、末尾のブロックカーソル（反転した空白 1 桁）は変えない。

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Performance meets specified goals
- [ ] Code review is completed

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし

## References

- 要件定義書: `feature-docs/minibuffer-grapheme-cursor-highlight/REQUIREMENTS.md`
