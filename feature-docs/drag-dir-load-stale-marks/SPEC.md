# Feature: drag-dir-load-stale-marks

## Overview

ディレクトリ読み込みの完了でペインの一覧が置き換わった後に、古い一覧のドラッグのアンカーとベースラインでマークが付かないようにする。一覧を置き換える成功した読み込み完了が、そのペインのドラッグセッションを取り消す。

## Objectives

- ディレクトリ読み込みの完了でペインの一覧が置き換わった後、古い一覧のドラッグのアンカーとベースラインでマークが付かない。

## User Stories

### US1: 一覧の置き換え後に古いアンカーでマークが付かない
マウスで操作する利用者として、ドラッグ中のペインの一覧がディレクトリ読み込みの完了で置き換わったとき、古い一覧のアンカーとベースラインでマークが付かないようにしたい。

**Acceptance Criteria:**
- [ ] AC1: 左ペインで押下して armed になったセッション（移動なし）で、左ペインに対する異なるエントリ一覧での成功した読み込み完了の後、別の行で移動・解放すると、左ペインのマーク集合は完了直後の集合と等しい。完了後の mouseDrag は armed ではない。
- [ ] AC2: すでに active のセッション（左ペインで押下と移動により範囲をマーク済み）で、左ペインに対する成功した読み込み完了の後、別の行で移動・解放すると、マーク集合は完了直後の集合と等しい。完了前に付いたマークは残る。

### US2: 一覧を置き換えない読み込み完了ではドラッグが続く
マウスで操作する利用者として、ドラッグ中のペインの一覧を置き換えない読み込み完了があっても、ドラッグを続けたい。

**Acceptance Criteria:**
- [ ] AC3: 左ペインのドラッグセッションで、右ペインに対する成功した読み込み完了では左のセッションは変わらず、その後の移動で元のアンカーからの左ペインの範囲がマークされる。
- [ ] AC4: 左ペインのドラッグセッションで、左ペインに対するエラーの読み込み完了ではセッションは変わらず、その後の移動で元のアンカーからの範囲がマークされる。
- [ ] AC5: ドラッグ中のペインに対する古い完了（panePath が空でない pendingPath と異なる）では、セッションとペインは変わらない。

### US3: 回帰の検出

**Acceptance Criteria:**
- [ ] AC6: AC1・AC2 のテストは修正前のコードでは失敗し、修正後は通る。
- [ ] AC7: 既存の単体テストと E2E テストが通る。gofmt で差分がなく、go vet で指摘がない。

## Technical Requirements

### Functional Requirements
- **FR1:** 一覧の置き換えでドラッグセッションを取り消す — `handleDirectoryLoadComplete` がドラッグセッション（armed または active）を持つペインのエントリを置き換えたとき、セッションを初期値に戻す（`mouseDrag = dragSession{}`）。その後の移動・解放は、次の押下まで何もマークしない。
- **FR2:** ドラッグ中のペインの読み込みだけが取り消す — 完了した読み込みのペイン（`msg.paneID`）がドラッグセッションのペインである場合にだけ、セッションを取り消す。もう一方のペインの読み込み完了ではドラッグセッションは変わらず、ドラッグは続く。
- **FR3:** 成功した読み込みだけが取り消す — エントリを置き換える成功の完了でだけ、セッションを取り消す。エラーの完了（`msg.err != nil`、エントリは置き換えない）と、pendingPath の確認で無視される古い完了では、どちらもドラッグセッションは変わらない。
- **FR4:** 完了前に付いたマークの維持 — 読み込み完了より前にドラッグで付いたマークは、セッションを取り消したときもそのまま残す。押下時のベースラインへは戻さない。
- **FR5:** 回帰テスト — ペインでの押下（と移動）、そのペインに対する異なるエントリ一覧での成功した読み込み完了、別の行での移動・解放を再現する単体テストを追加する。テストは修正前のコードでは失敗し、修正後は通る。

### Non-Functional Requirements
- **NFR1:** 読み込みを挟まないドラッグの維持 — 操作中に読み込み完了がない場合のクリック、ドラッグ、ダブルクリックの動作は変えない。
- **NFR2:** 読み込み完了のその他の処理の維持 — `handleDirectoryLoadComplete` のエントリ、フィルタのリセット、カーソル、スクロール、pendingPath、git ブランチ、履歴、ディスク空き容量の処理は、成功とエラーのどちらの経路でも従来どおり動く。
- **NFR3:** 既存テストとツールの通過 — 既存の単体テストと E2E テスト（`mouse_tests.sh` と `mark_tests.sh` を含む）が通る。`go test ./...`、gofmt、go vet で失敗や差分がない。

## Implementation Approach

### Architecture

**読み込み完了時のドラッグセッションの扱い:**
```
handleDirectoryLoadComplete(msg)
  ├─ pendingPath の確認で古い完了と判定 → 一覧を置き換えない（ドラッグセッションは変えない）
  ├─ msg.err != nil → エントリを置き換えない（ドラッグセッションは変えない）
  └─ 成功 → エントリを置き換える
       └─ msg.paneID がドラッグセッションのペインなら mouseDrag = dragSession{}
```

### Data Flow

```
押下 → mouseDrag が armed → （移動 → active、範囲をマーク）
  → ドラッグ中のペインへの成功した読み込み完了 → mouseDrag = dragSession{}（付いたマークは残す）
  → 移動・解放 → 何もマークしない（次の押下まで）
```

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- `handleDirectoryLoadComplete`: 一覧を置き換える読み込み完了処理
- `directoryLoadCompleteMsg`: 読み込み完了メッセージ（`paneID`、`panePath`、`err`）
- `mouseDrag` / `dragSession`: ドラッグセッションの状態
- `newMouseTestModel`: 単体テストのモデル生成

**External Dependencies:**
- 該当なし

### File Structure

変更するファイルは create-plan で `workflow.yaml` の各タスクの `files` から導出する。

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/drag-dir-load-stale-marks/**`
- `test-docs/drag-dir-load-stale-marks/**`

`feature-docs/drag-dir-load-stale-marks/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/drag-dir-load-stale-marks/**` covers `test-docs/drag-dir-load-stale-marks/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/drag-dir-load-stale-marks/` directory at all; the declared
`test-docs/drag-dir-load-stale-marks/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS-1（FR1、FR5 / AC1、AC6）: `newMouseTestModel` で、左ペインの項目を押下する。LeftPane に対して異なるエントリ一覧の `directoryLoadCompleteMsg`（err は nil、panePath が一致するか pendingPath が空）を送る - mouseDrag がリセットされている。別の行で移動・解放する - 左ペインのマーク集合が完了直後の集合と等しい
- [ ] TS-2（FR1、FR4、FR5 / AC2、AC6）: 左ペインで押下と移動を行う（範囲をマーク）。LeftPane に対して異なるエントリ一覧で成功した完了を送る - 完了前のマークが残り、セッションがリセットされている。別の行で移動・解放する - マークがそれ以上変わらない
- [ ] TS-3（FR2 / AC3）: 左ペインで押下する。RightPane に対して成功した完了を送る - 左のセッションは armed のまま。左ペインで移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-4（FR3 / AC4）: 左ペインで押下する。LeftPane に対して err が nil でない完了を送る - セッションが変わらない。移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-5（FR3 / AC5）: 左ペインで押下する。左の pendingPath にパスを設定し、LeftPane に対して異なる panePath の完了を送る - セッションとペインのエントリが変わらない

### Integration Tests
- 該当なし

### E2E Tests
**Existing E2E tests**: `mouse_tests.sh`、`mark_tests.sh` を含む既存の E2E テスト
**Run command**: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] TS-6（NFR1、NFR2、NFR3 / AC7）: 既存の単体テスト一式（`go test ./...`）と E2E テスト一式（`make test-e2e-build && make test-e2e`）を実行する - すべて通る

### Static Checks
- [ ] TS-7（NFR3 / AC7）: `gofmt -w .` で差分がない。`go vet ./...` で指摘がない

### Edge Cases
- [ ] セッションが armed（移動なし）の場合（TS-1）
- [ ] セッションが active で、完了前にマークが付いている場合（TS-2）
- [ ] もう一方のペインの読み込み完了（TS-3）
- [ ] エラーの読み込み完了（TS-4）
- [ ] pendingPath の確認で無視される古い完了（TS-5）

### Performance Tests
- 該当なし

## Security Considerations

該当なし

## Error Handling

エラーの読み込み完了（`msg.err != nil`）ではドラッグセッションを変えない（FR3）。エラー経路のその他の処理は従来どおり（NFR2）。

## Performance Optimization

該当なし

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Code review is completed

## Assumptions

- A1: 対象はディレクトリ読み込み完了の経路（`handleDirectoryLoadComplete`）だけとする。一覧を置き換える他の経路（autoRefreshMsg、exec・シェルコマンド・一括操作の後の再読み込み）は変更せず、この機能の対象外とする。これらは別タスクとして起票する。
- A2: 古い完了メッセージは、既存の pendingPath の確認で引き続き無視し、ドラッグセッションは変えない。
- A3: ディレクトリ変更時のマークの既存の扱いは、この修正で変えない。

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

## References

- 要件の対応付け: `feature-docs/drag-dir-load-stale-marks/workflow.yaml`
