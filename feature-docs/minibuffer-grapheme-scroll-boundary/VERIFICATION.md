# Verification Document: minibuffer-grapheme-scroll-boundary

## Overview

**Feature**: minibuffer-grapheme-scroll-boundary / **SPEC.md**: `feature-docs/minibuffer-grapheme-scroll-boundary/SPEC.md` / **IMPLEMENTATION.md**: なし（reduced tier で省略。タスクは task0001 の 1 件）

## Build Verification

- Command: workflow.yaml の `project.components.duofm.build_command`（golang:1.25.5-trixie コンテナで make build）
- Expected: exit code 0, no errors

## Test Verification

- Command: workflow.yaml の `project.components.duofm.test_command`（同じコンテナで go test ./...）
- Expected: exit code 0、すべてのテストが通る
- Coverage target: 設けない

### Test Scenarios from SPEC.md

共通の条件: プロンプト `(search): `（表示幅 10）、入力は `aaaaa` + 家族の絵文字（U+1F468 U+200D U+1F469 U+200D U+1F467 U+200D U+1F466）+ `bbbbb`（17 rune、実際の表示幅 12、家族の絵文字は rune 位置 5〜11 で表示幅 2）。「表示上の文字列」は View の結果から装飾用エスケープシーケンスを除いた文字列。「1 行の条件」は、改行を含まず表示幅が m.width − 2 以下であること。

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | m.width 24（残りの幅 10）、カーソルが末尾で View を描画する（再現条件） | 表示上の文字列に `👨‍👩‍👧‍👦bbbbb` が続けて含まれ、1 行の条件を満たす | Unit |
| TS-2 | TS-1 と同じ条件で、カーソルを位置 0 に置いて View を描画する | 表示上の文字列に `aaaaa👨‍👩‍👧‍👦` が続けて含まれ、1 行の条件を満たす | Unit |
| TS-3 | m.width 15〜27 の各幅とカーソル位置 0〜17 のすべての組み合わせで View を描画する（カーソルが家族の絵文字の途中の rune にある位置 5〜11 を含む） | すべての組み合わせで、表示上の文字列が家族の絵文字の結合文字列全体を含むか、U+1F468・U+1F469・U+1F467・U+1F466・U+200D のどれも含まないかのどちらかで、1 行の条件を満たす | Unit（表形式） |
| TS-4 | TS-1〜TS-3 のテストの最初に、フィクスチャの前提を確かめる | プロンプトの表示幅 10、入力の rune 数 17、入力全体の実際の表示幅 12、家族の絵文字の表示幅 2。崩れていれば致命的な失敗で即座に止まる | Unit |
| TS-5 | 既存のミニバッファのテストと go test ./... を実行する | 既存テストを変更せずにすべて通る（LinearTime 系が時間内に終わることを含む） | Unit |

### Edge Cases from SPEC.md

| Edge case | How to Verify |
|-----------|---------------|
| カーソルが絵文字の最初の rune（位置 5）、ZWJ（位置 6 など）、最後の rune（位置 11）にあるとき | TS-3（位置 5〜11 を含む）。m.width 15 では task0001 の AC-6 のテストで、プロンプトより後ろに入力の文字も U+200D も表示されないことを確かめる |
| 残りの幅が 1（m.width 15）で、表示幅 2 の絵文字が入らないとき | TS-3 の m.width 15 の行と、task0001 の AC-6 のテスト |
| 絵文字がちょうど表示範囲の左端・右端にかかる幅のとき | TS-3 の掃引 |
| 結合文字（e + U+0301）や VS16 の組が分かれないこと | TS-5（既存の TestMinibufferView_GraphemeClusters_BoundedSingleLine、VS16 系のテスト） |
| ZWJ 以外で正の幅の rune が結合する書記素（国旗、肌色の修飾子） | task0001 の AC-5 のテスト（国旗 U+1F1EF U+1F1F5 と、肌色の修飾子付きの絵文字 U+1F44D U+1F3FD で、m.width 15〜27 × カーソル位置 0〜12 の掃引）。go test ./... に含まれる |

## Code Quality Verification

- Format: workflow.yaml の `project.components.duofm.format_command`（gofmt -w .）。実行後に差分が出ない
- Static analysis: workflow.yaml に宣言なし

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | m.width 24、カーソル末尾で、`👨‍👩‍👧‍👦bbbbb` が含まれ、1 行に収まる | TS-1 |
| AC2 | m.width 24、カーソル位置 0 で、`aaaaa👨‍👩‍👧‍👦` が含まれ、1 行に収まる | TS-2 |
| AC3 | m.width 15〜27 とカーソル位置 0〜17 のすべての組み合わせで、家族の絵文字が丸ごと含まれるか部品が一つも含まれず、1 行に収まる | TS-3 |
| AC4 | 既存のミニバッファのテストと go test ./... が変更なしで通り、変更は minibuffer.go と minibuffer_test.go だけ（前提 a6 の例外を除く） | TS-5 と、implement の base_commit から統合ブランチまでの差分のファイル一覧（feature-docs/ と test-docs/ 配下のワークフロー生成物を除く）。go.mod と go.sum に差分があるときは、github.com/rivo/uniseg の `// indirect` 表記の削除と、それに伴う go.sum の変更だけであること |
| SC-1 | すべての機能要件が実装され、テストされている | 下の Functional Requirements Coverage の全行が満たされる |
| SC-2 | すべてのテストシナリオが通る | TS-1〜TS-5 が通る |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-3, TS-4 |
| FR2 | task0001 | TS-2, TS-3, TS-4 |
| FR3 | task0001 | TS-3, TS-4 |
| FR4 | task0001 | TS-1, TS-2, TS-3, TS-4 |
| NFR1 | task0001 | TS-5 と差分のファイル一覧（AC4 と同じ方法） |
| NFR2 | task0001 | TS-5 |

## E2E Testing

E2E テストは追加しない（SPEC a5）。この機能の E2E の検証項目はない。

## Manual Testing (E2E Not Possible)

なし（再現条件は TS-1・TS-2 で自動で検証する）

## Security / License Verification

- 依存: github.com/rivo/uniseg（MIT、既存の間接依存）を直接使う場合がある。project.license（MIT）と両立する。新しい依存の追加がないことを、go.mod の require の一覧で確かめる

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Test Scenarios (TS-1〜TS-5) | 5 | 5 | 0 | 0 |
| Edge Cases | 5 | 5 | 0 | 0 |
| Code Quality | 1 | 1 | 0 | 0 |
| SPEC.md Compliance (AC1〜AC4, SC-1, SC-2) | 6 | 6 | 0 | 0 |
| License | 1 | 1 | 0 | 0 |
