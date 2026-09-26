# Verification Document: minibuffer-wide-prompt-wrap

## Overview

**Feature**: minibuffer-wide-prompt-wrap / **SPEC.md**: `feature-docs/minibuffer-wide-prompt-wrap/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/minibuffer-wide-prompt-wrap/IMPLEMENTATION.md`

統合後（全タスクのマージ後）に verify フェーズが実施する検証項目。タスク単位の受け入れ基準は `feature-docs/minibuffer-wide-prompt-wrap/tasks/` の各タスク計画にある。

## Build Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=2 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main -e GIT_CONFIG_KEY_1=safe.directory -e GIT_CONFIG_VALUE_1=/home/sakura/go/src/duofm docker.io/library/golang:1.25.5-trixie make build`
- Expected: exit code 0, no errors

## Test Verification

- Command: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie go test ./...`
- Expected: exit code 0、全テスト通過
- Coverage target: 数値目標は設けない。`Minibuffer.View` の各区分（幅 0〜4、残りの幅 0 以下、残りの幅 1 以上で入力が収まる／収まらない）を、いずれも少なくとも 1 つのテストが通ること

### Test Scenarios from SPEC.md

SPEC.md の TS1〜TS4 は、ここでは TS-1〜TS-4 として扱う。TS-5〜TS-8 は SPEC.md の Edge Cases・受け入れ基準から追加したもの。

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | SPEC TS1: 表形式のテスト。幅（残りの幅 0 以下の経路 24〜32、1 以上の経路 36〜37）× プロンプト「(reverse-i-search)'日本語': 」× 入力（空、「echo 日本語」、ASCII の長い入力）× カーソル位置（先頭・途中・全角文字の上・末尾）の全組み合わせで View を描画する（`internal/ui/minibuffer_test.go`） | 結果が CR・LF を含まず、結果の表示幅が `m.width-2` 以下 | Unit |
| TS-2 | SPEC TS2: 全角文字が切り詰め・表示範囲の境界にかかる幅（24・26）、プロンプトの表示幅が内容幅とちょうど同じ・1 桁少ない・1 桁多い幅（32・33・31）、残りの幅が 1 でカーソル位置の文字が全角の場合（幅 33）で View を描画する（`internal/ui/minibuffer_test.go`） | 結果が CR・LF を含まず表示幅が `m.width-2` 以下。境界にかかる全角文字は結果に含まれない | Unit |
| TS-3 | SPEC TS3: 幅 26 と 37 で、全角プロンプトのミニバッファを表示した `ViewWithMinibuffer` の出力行数と、1 行に収まる ASCII プロンプト「/: 」のときの出力行数を比べる（`internal/ui/pane_render_test.go`） | 両者の出力行数が一致する | Unit |
| TS-4 | SPEC TS4: 既存の幅 0〜4 のテスト（パニックなし・改行なし）、reverse-i-search の幅 24 のテスト、幅 0/1 の行数テスト、`TestMinibufferView` を変更後のコードで実行する（`internal/ui/minibuffer_test.go`、`internal/ui/pane_render_test.go`） | 既存テストが変更なしで通る | Unit |
| TS-5 | 結合文字・ZWJ で連結した絵文字を含むプロンプトと入力で、幅 24〜37、カーソルが書記素の途中の rune を含む各位置で View を描画する（`internal/ui/minibuffer_test.go`） | 結果が CR・LF を含まず、結果の表示幅が `m.width-2` 以下 | Unit |
| TS-6 | 幅 40、プロンプト「(reverse-i-search)'日本語': 」、入力「ls」で View を描画する（`internal/ui/minibuffer_test.go`） | プロンプト全体と入力「ls」が切り詰めずに結果に含まれる | Unit |
| TS-7 | 残りの幅が 1 以上で入力が収まらないとき、文字がすべて異なる入力（ASCII と全角を含む）でカーソル位置を変えて View を描画する（`internal/ui/minibuffer_test.go`） | カーソル位置の文字が結果に含まれる。カーソルが末尾のときは入力の最後の文字が結果に含まれる | Unit |
| TS-8 | 既存の HandleKey・編集位置のテストを含む `go test ./...` を実行する | 既存テストを変更せずにすべて通る（編集操作と rune 単位の編集位置が変わっていない） | Unit |

## Code Quality Verification

- Format: `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie gofmt -w .` — 実行後に作業ツリーへ差分が出ないこと
- Static analysis: workflow.yaml に宣言されたコマンドなし

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | AC1: 幅 24〜32 で、全角プロンプト（入力あり・なし）の View の結果が改行を含まず、表示幅が `m.width-2` 以下 | TS-1、TS-2 |
| SC-2 | AC2: 幅 36〜37 で、入力「echo 日本語」のとき、カーソル位置によらず View の結果が改行を含まず、表示幅が `m.width-2` 以下 | TS-1、TS-2 |
| SC-3 | AC3: 残りの幅 1 以上の経路で ASCII の長い入力のとき、View の結果が改行を含まず、表示幅が `m.width-2` 以下 | TS-1 |
| SC-4 | AC4: 幅 0〜4 で、パニックせず改行を含まないことを確かめる既存のテストが通る | TS-4 |
| SC-5 | AC5: 両方の経路の幅（26 と 37）で、全角プロンプトのペインの行数が ASCII プロンプトのときと一致する | TS-3 |
| SC-6 | AC6: 幅 40 で、プロンプト「/: 」と入力「test」がそのまま表示される | TS-4、TS-6 |
| SC-7 | AC7: `go test ./...` がすべて通る | TS-8（Test Verification のコマンド） |
| SC-8 | 再現手順で現象が起きない | TS-3、Manual Testing の 1 項目目 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1、TS-2、TS-5（View の結果を lipgloss の表示幅計測で判定） |
| FR2 | task0001 | TS-1、TS-2、TS-5 |
| FR3 | task0001 | TS-1、TS-2 |
| FR4 | task0001 | TS-1、TS-2、TS-7 |
| FR5 | task0001 | TS-1、TS-2 |
| FR6 | task0001 | TS-4 |
| FR7 | task0001 | TS-3 |
| FR8 | task0001 | TS-4、TS-6 |
| NFR1 | task0001 | TS-8、review で変更が View の描画計算に閉じていることを確認 |

## E2E Testing

E2E テストは追加しない（SPEC の仮定 a5）。既存の E2E スイートを回帰確認として実行する。

- [ ] 既存の E2E スイートが通る: `make test-e2e-build && make test-e2e`

## Manual Testing (E2E Not Possible)

- [ ] ミニバッファ幅が 26〜32 程度になる端末幅で duofm を起動し、履歴検索（reverse-i-search）で「日本語」をパターンとして入力したとき、ミニバッファが 1 行で描画され、ペインの行数がずれない（ASCII のパターンのときと同じ位置に描画される）
- [ ] ミニバッファ幅が 36〜37 程度になる端末幅で、全角文字を含む入力を編集し、カーソルを先頭・途中・全角文字の上・末尾へ動かしたとき、ミニバッファが 1 行のままで、カーソル位置の文字が見えている

## Performance / Security Verification (if applicable)

該当なし。

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Test scenarios | 8 | 8 | 0 | 0 |
| Code quality | 1 | 1 | 0 | 0 |
| E2E regression | 1 | 0 | 1 | 0 |
| Manual | 2 | 0 | 0 | 2 |
| Total | 13 | 10 | 1 | 2 |
