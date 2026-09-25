# Feature: header-wrap-hittest

## Overview

git 管理外のディレクトリでペイン幅より長いパスを開くと、ペインヘッダが折り返して複数行になり、マウスのクリック・ドラッグ・ダブルクリックが意図と違うエントリに作用する。
ヘッダ1行目を常に1行に収め、ペインヘッダの行数を1つの定義として描画とヒットテストで共有する。
要件の詳細は `feature-docs/header-wrap-hittest/REQUIREMENTS.md` を参照する。

## Objectives

- git 管理外のディレクトリでペイン幅より長いパスを開いたときも、マウスのクリック・ドラッグ・ダブルクリックがクリックした行のエントリに作用する。
- ペインヘッダの行数を1つの定義とし、描画とヒットテストがその定義を共有する（レビュー指摘 stable_id: 52107dd60886982b）。

## User Stories

### US1: 長いパスのディレクトリでエントリをクリックする
マウス操作のユーザーとして、git 管理外でペイン幅より長いパスのディレクトリを開いているときも、クリックした行のエントリを選択したい。

**Acceptance Criteria:**
- [ ] AC1: git 管理外で、ペイン幅より長いパスのディレクトリを開いてエントリ行をクリックすると、クリックした行のエントリが選択される。（FR1, FR3）
- [ ] AC3: git ブランチなし・長いパスのとき、ヘッダ1行目の表示幅はペインの内容幅以下で、末尾が「...」になる。（FR3, NFR2）
- [ ] AC6: 長いパス・ブランチなしのペインを描画したとき、最初のエントリが現れる行位置と、ヒットテストが最初のエントリ行とみなす行位置が一致することを検証するテストがあり、修正前のコードでは失敗する。（FR1, FR6）

### US2: 長いパスのディレクトリでドラッグ・ダブルクリックする
マウス操作のユーザーとして、同じ条件で、ドラッグ・ダブルクリックを操作した行のエントリに作用させたい。

**Acceptance Criteria:**
- [ ] AC2: 同じ条件で、ドラッグはドラッグした行の範囲のエントリをマークし、ダブルクリックはクリックした行のエントリに Enter 動作を行う。（FR2, FR3）

### US3: 狭いペイン幅でもヘッダが崩れない
マウス操作のユーザーとして、ペイン幅が狭いときもヘッダ1行目が1行に収まり、描画が panic しないようにしたい。

**Acceptance Criteria:**
- [ ] AC4: パス表示幅が4未満でブランチ表示だけを出す幅のとき、ブランチ名がペインの内容幅を超えても、ヘッダ1行目は内容幅以下に収まる。（FR4）
- [ ] AC5: 内容幅が0以下になるペイン幅（例: 0〜4）でヘッダを描画しても panic しない。（FR5）

### US4: 既存の振る舞いを変えない
マウス操作のユーザーとして、パスがペイン幅に収まる場合の振る舞いは今まで通りにしたい。

**Acceptance Criteria:**
- [ ] AC7: パスが短い場合の既存のユニットテストと go test ./... がすべて通る。（NFR1）

## Technical Requirements

### Functional Requirements
- **FR1:** クリック行とエントリの対応 — ペインヘッダの内容（パスの長さ、git ブランチの有無、[H] 表示・フィルタ表示の有無）に関わらず、エントリ行を左クリックすると、その行に表示されているエントリにカーソルが移動する。（status: confirmed）
- **FR2:** ドラッグ・ダブルクリックの対応 — ドラッグによる範囲マークとダブルクリックによる Enter 動作も、FR1 と同じ行とエントリの対応で動作する。（status: confirmed）
- **FR3:** ヘッダ1行目を常に1行に収める — ヘッダ1行目（パス、[H] 表示、フィルタ表示、git ブランチ表示）は、git ブランチの有無に関わらず、表示幅がペインの内容幅（p.width-4、0 未満のときは 0）を超えない。超える場合は末尾を「...」で切り詰める。（status: assumed）
- **FR4:** ブランチ単独表示の切り詰め — ペイン幅が狭くパス表示幅が4未満になりブランチ表示だけを出す場合も、ブランチ表示がペインの内容幅を超えるときは切り詰め、ヘッダ1行目を1行に収める。（status: assumed）
- **FR5:** 負の幅での安全性 — ペイン幅が極端に狭く内容幅が0以下になる場合でも、ヘッダ描画は panic せず、切り詰め処理に負の幅を渡さない（幅は0以上に補正する）。（status: assumed）
- **FR6:** ヘッダ行数の定義共有 — ペインヘッダの行数は1つの定義とし、ヒットテスト（最初のエントリ行・ドラッグ行の対応）と描画側の表示件数計算（通常表示・bg 出力分割表示・dimmed 表示）がその定義を参照する。（status: assumed）

### Non-Functional Requirements
- **NFR1 - 既存挙動の維持:** パスがペイン幅に収まる場合の既存のヘッダ表示、ヒットテスト、マウス操作の振る舞いを変えない。既存ユニットテストと test/e2e/scripts/tests/mouse_tests.sh が通る。
- **NFR2 - 表示幅の計算:** 表示幅の計算は go-runewidth を使い、全角文字を含むパスでも表示幅で切り詰める。
- **NFR3 - 表示モード間の一致:** 通常表示・bg 出力分割表示・dimmed 表示のすべてで、ヘッダ1行目の切り詰めとヘッダ行数の定義が同じになる。

## Implementation Approach

### Architecture

**現状:**
- ヒットテスト（`internal/ui/mouse_hittest.go`）はヘッダを固定 3 行（`mousePaneHeaderRows=3` / `mouseFirstEntryRow=4`）とみなしており、描画側と定義を共有していない。
- `internal/ui/pane_render.go` の `renderHeaderLine1` は gitBranch が空のとき displayPath を切り詰めず、lipgloss の `Width(p.width-2)` で折り返す。

**対象コンポーネント:**
- `renderHeaderLine1`（`internal/ui/pane_render.go`）: git ブランチの有無に関わらず、ヘッダ1行目の表示幅を内容幅（p.width-4、0 未満は 0）以下に収める（FR3）。ブランチ単独表示の分岐でも切り詰める（FR4）。
- `truncateStringWithEllipsis`: 先頭を残して末尾を「...」で省略する（A2）。負の幅を渡さない（FR5）。
- ペインヘッダ行数の定義: 1つの定義とし、ヒットテスト（最初のエントリ行・ドラッグ行の対応）と描画側の表示件数計算（通常表示・bg 出力分割表示・dimmed 表示）が参照する（FR6, NFR3）。
- 既存のヒットテスト規則（タイトル行・ステータス行は hitNone、ヘッダ行は hitNonEntry、bg 分割時の区切り以降は hitNonEntry、ドラッグ行のクランプ）は変更しない（A4）。

### Data Flow

該当なし

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- `internal/ui`: ペイン描画、マウスのヒットテスト、Model のマウス処理

**External Dependencies:**
- go-runewidth: 表示幅の計算（NFR2）
- lipgloss: ペインヘッダの描画

### File Structure

```
internal/ui/
├── mouse_hittest.go             # ヒットテスト
├── mouse_hittest_test.go
├── model_update_mouse_test.go
├── pane_render.go               # renderHeaderLine1
└── pane_render_test.go
test/e2e/scripts/tests/
└── mouse_tests.sh
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/header-wrap-hittest/**`
- `test-docs/header-wrap-hittest/**`

`feature-docs/header-wrap-hittest/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/header-wrap-hittest/**` covers `test-docs/header-wrap-hittest/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/header-wrap-hittest/` directory at all; the declared
`test-docs/header-wrap-hittest/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS1 (FR3, NFR2): renderHeaderLine1: gitBranch 空・幅40・ペイン幅を超える長いパスで、戻り値の runewidth 表示幅が p.width-4 以下で、末尾が「...」になる。
- [ ] TS2 (FR3, NFR2): renderHeaderLine1: gitBranch 空で [H] 表示とフィルタ表示が付き、合計がペイン幅を超える場合も表示幅が p.width-4 以下になる。
- [ ] TS3 (FR3, NFR2): renderHeaderLine1: 全角文字を含む長いパスで、表示幅が p.width-4 以下になる。
- [ ] TS4 (FR4): renderHeaderLine1: 狭い幅（例: 15）で長いブランチ名のとき（ブランチ単独表示の分岐）、表示幅が p.width-4 以下になる。
- [ ] TS5 (FR5): renderHeaderLine1 と truncateStringWithEllipsis: ペイン幅 0〜4（内容幅が0以下）で、gitBranch の有無どちらでも panic しない。
- [ ] TS6 (FR1, FR6): ペイン描画（通常表示・bg 出力分割表示・dimmed 表示）: gitBranch 空・長いパスで描画し、最初のエントリが現れる行の位置が共有されたヘッダ行数の定義と一致し、ヒットテストの最初のエントリ行（タイトル行を除く）と一致する。
- [ ] TS7 (FR1, FR2, FR3): Model のマウス処理: 長いパス・gitBranch 空のペインで、エントリ i の行をクリックするとカーソルが i に移動する。ドラッグ・ダブルクリックも同じ行とエントリの対応で動く。

### Integration Tests
- [ ] TS8 (NFR1, regression): 既存の mouse_hittest_test.go、model_update_mouse_test.go、pane_render_test.go と go test ./... が通る。

### E2E Tests
**Existing E2E tests**: test/e2e/scripts/tests/mouse_tests.sh
**Run command**: Not detected
- [ ] Existing E2E tests pass without regression
- E2E テストの追加は必須としない（A3）。

### Edge Cases
- [ ] ブランチ単独表示: パス表示幅が4未満でブランチ名が内容幅を超える場合、ヘッダ1行目を内容幅以下に切り詰める（FR4）。
- [ ] 内容幅が0以下: ペイン幅 0〜4 で panic せず、切り詰め処理に負の幅を渡さない（FR5）。
- [ ] 全角文字: 全角文字を含むパスを表示幅で切り詰める（NFR2）。

### Performance Tests
該当なし

## Security Considerations

該当なし

## Error Handling

### Error Codes

該当なし

### Error Flow

- 内容幅が0以下になるペイン幅では、幅を0以上に補正し、panic しない（FR5）。

## Performance Optimization

該当なし

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] AC1〜AC7 を満たす
- [ ] Code review is completed

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

### Assumptions
- A1: 長いパスはヘッダ1行目を常に1行に収めるため、末尾を「...」で切り詰めて表示する（git 管理外、ブランチ単独表示の狭い幅も含む）。ヘッダ行数は描画とヒットテストで1つの定義を共有する。
- A2: 切り詰めは既存の truncateStringWithEllipsis と同じく先頭を残して末尾を省略する。長いパスの末尾（カレントディレクトリ名）は見えなくなる。
- A3: 再発検出テストは Go のユニットテスト（internal/ui の *_test.go）で用意する。E2E テストの追加は必須としない。
- A4: 既存のヒットテスト規則（タイトル行・ステータス行は hitNone、ヘッダ行は hitNonEntry、bg 分割時の区切り以降は hitNonEntry、ドラッグ行のクランプ）は変更しない。
- A5: 設計工程は省略する。

## Implementation Phases (if applicable)

該当なし

## References

- 要件定義書: `feature-docs/header-wrap-hittest/REQUIREMENTS.md`
- `internal/ui/mouse_hittest.go`
- `internal/ui/pane_render.go`（renderHeaderLine1）
- レビュー指摘 stable_id: 52107dd60886982b
