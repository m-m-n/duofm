# Verification Document: minibuffer-zwj-cursor-end-scroll

## Overview

**Feature**: minibuffer-zwj-cursor-end-scroll / **SPEC.md**: `feature-docs/minibuffer-zwj-cursor-end-scroll/SPEC.md` / **IMPLEMENTATION.md**: なし（reduced tier で省略。タスクは task0001 の 1 件）

## Build Verification

- Command: workflow.yaml の `project.components.duofm.build_command`（golang:1.25.5-trixie コンテナで make build）
- Expected: exit code 0, no errors

## Test Verification

- Command: workflow.yaml の `project.components.duofm.test_command`（同じコンテナで go test ./...）
- Expected: exit code 0、すべてのテストが通る
- Coverage target: 設けない

### Test Scenarios from SPEC.md

共通の条件: プロンプト `(search): `（表示幅 10）、入力は家族の絵文字（U+1F468 U+200D U+1F469 U+200D U+1F467 U+200D U+1F466）+ `ls`（9 rune、実際の表示幅 4、rune ごとの幅の合計 10）。「表示上の文字列」は View の結果から装飾用エスケープシーケンスを除いた文字列。

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | カーソルが末尾で、m.width 19〜24 の各幅で View を描画する（m.width 24 が再現条件） | 表示上の文字列に `(search): 👨‍👩‍👧‍👦ls` が続けて含まれ、先頭の U+1F468 が削られない。結果は改行を含まず、表示幅は m.width − 2 以下 | Unit |
| TS-2 | カーソルを `l`（rune 位置 7）と `s`（rune 位置 8）に置き、m.width 19〜24 の各幅で View を描画する | 表示上の文字列に `(search): 👨‍👩‍👧‍👦ls` が続けて含まれる（家族の絵文字が途中で切れず、`s` も表示される） | Unit |
| TS-3 | m.width 18〜30 の各幅とカーソル位置 0〜9 のすべての組み合わせで View を描画する（書記素の途中の rune にカーソルがある場合を含む）。m.width 18 でカーソルが末尾の場合は、実際の幅でも収まらずスクロール経路に入る | すべての組み合わせで改行を含まず、表示幅は m.width − 2 以下。m.width 18 でカーソルが末尾の場合、表示上の文字列に `ls` が含まれる | Unit |
| TS-4 | 既存のミニバッファのテストと go test ./... を実行する | 既存テストを変更せずにすべて通る | Unit |

境界条件（上の表に含まれる）: m.width 19（残りの幅 5 = 実際の幅 + 1）、m.width 24（残りの幅 10 = rune ごとの合計 + 1 より 1 少ない）、m.width 18（実際の幅でも収まらない）、書記素の途中の rune（ZWJ を含む）にカーソルがある場合。

## Code Quality Verification

- Format: workflow.yaml の `project.components.duofm.format_command`（gofmt -w .）。実行後に差分が出ない
- Static analysis: workflow.yaml に宣言なし

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | カーソルが末尾で m.width 19〜24 のとき、`👨‍👩‍👧‍👦ls` がそのまま含まれる | TS-1 |
| AC2 | カーソルが `l` または `s` で m.width 19〜24 のとき、`👨‍👩‍👧‍👦` が途中で切れずに含まれ、`s` も含まれる | TS-2 |
| AC3 | m.width 18〜30 とカーソル位置 0〜9 のすべての組み合わせで、改行を含まず表示幅が m.width − 2 以下 | TS-3 |
| AC4 | 既存のミニバッファのテストと go test ./... が通り、変更は minibuffer.go と minibuffer_test.go だけ | TS-4 と、implement の base_commit から統合ブランチまでの差分のファイル一覧（feature-docs/ と test-docs/ 配下のワークフロー生成物を除く） |
| AC5 | 再現手順（m.width 24、`(search): `、家族の絵文字 + `ls`、カーソル末尾）で先頭の code point が削られない | TS-1 の m.width 24 の行 |
| SC-1 | All functional requirements are implemented and tested | 下の Functional Requirements Coverage の全行が満たされる |
| SC-2 | All test scenarios pass | TS-1〜TS-4 が通る |
| SC-3 | Code review is completed | workflow.yaml の review step が completed |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2 |
| FR2 | task0001 | TS-1 |
| FR3 | task0001 | TS-2 |
| FR4 | task0001 | TS-1, TS-3 |
| NFR1 | task0001 | TS-4 と差分のファイル一覧（AC4 と同じ方法） |
| NFR2 | task0001 | TS-4 |

## E2E Testing

E2E テストは追加しない（SPEC a6）。この機能の E2E の検証項目はない。

## Manual Testing (E2E Not Possible)

なし（再現手順は TS-1 の m.width 24 の行で自動で検証する）

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Test Scenarios (TS-1〜TS-4) | 4 | 4 | 0 | 0 |
| Code Quality | 1 | 1 | 0 | 0 |
| SPEC.md Compliance (AC1〜AC5, SC-1〜SC-3) | 8 | 8 | 0 | 0 |
