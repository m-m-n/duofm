# Feature: drag-refresh-stale-anchor

## Overview

ドラッグ中のペインの表示一覧（フィルタ適用後の entries）が再読み込みで変わったとき、移動・解放でドラッグセッションを取り消し、古い一覧のアンカーとベースラインでマークが付かないようにする。表示一覧が変わらない再読み込みでは、ドラッグを続ける。

要件の詳細は `feature-docs/drag-refresh-stale-anchor/REQUIREMENTS.md` を参照する。

## Objectives

- ドラッグ中のペインの表示一覧が再読み込みで変わった後に、古い一覧のアンカーとベースラインでマークが付かない。
- ドラッグ中のペインの表示一覧が変わらない再読み込み（定期更新を含む）では、ドラッグが途切れない。

## User Stories

該当なし。受け入れ基準は Success Criteria に記載する。

## Technical Requirements

### Functional Requirements
- **FR1:** 表示一覧の変化でドラッグを取り消す。押下でドラッグセッションを armed にするとき、ドラッグ中のペインの表示一覧（フィルタ適用後の entries）の名前の並びを、ペインの一覧とは独立に保持する。移動・解放では、行を解決する前に、ペインの現在の表示一覧の名前の並びと比較する。違っていればセッションを初期値に戻し（`mouseDrag = dragSession{}`）、その移動・解放ではマークを変えない。
- **FR2:** 経路を問わない。FR1 の判定は、表示一覧を変えた経路に関係なく働く。自動更新（`autoRefreshMsg`）、exec・シェルコマンド・一括操作の後の再読み込み、ダイアログの結果やファイル操作の完了の後の再読み込み、F5/Ctrl+R、削除後の再読み込み、バックグラウンドコマンドの完了による再読み込みを含む。
- **FR3:** 表示一覧が変わらなければドラッグを続ける。ドラッグ中のペインの表示一覧の名前と順序が押下時と同じなら、セッションは変えない。その後の移動で、元のアンカーからの範囲がマークされる。サイズや更新時刻だけが変わった場合、フィルタに一致しないファイルだけが増減した場合、読み込みに失敗して一覧が置き換わらなかった場合も、これに当たる。
- **FR4:** 判定対象はドラッグ中のペインだけ。もう一方のペインの表示一覧が変わっても、ドラッグ中のペインのセッションは変えない。
- **FR5:** 取り消し時のマークの維持。FR1 でセッションを取り消すとき、その時点で残っているマークはそのまま残し、押下時の baseline には戻さない。次の押下までは、移動・解放でマークを変えない。
- **FR6:** 消えたファイルのマークを復活させない。表示一覧が変わらずドラッグが続く場合でも、再読み込みで消えたファイルの名前のマークを、押下時の baseline から付け直さない。例: フィルタで表示されていないマーク済みのファイルが削除された場合。
- **FR7:** 回帰テスト。一時ディレクトリを使うモデルと `autoRefreshMsg` で、次の順の操作を再現する単体テストを追加する: 押下（と移動）→ ドラッグ中のペインのディレクトリでファイルを追加・削除 → 再読み込み → 別の行で移動・解放。テストは修正前のコードでは失敗し、修正後は通る。表示一覧が変わらない定期更新でドラッグが続くことを確かめるテストも追加する。E2E テストは追加しない。既存の E2E は回帰確認にだけ使う。

### Non-Functional Requirements
- **NFR1:** 再読み込みを挟まない操作の維持。操作中に表示一覧が変わらない場合のクリック、ドラッグ、ダブルクリックの動作は変えない。
- **NFR2:** 再読み込み処理の維持。`RefreshDirectoryPreserveCursor` のカーソル・フィルタ・スクロール・マークの扱いは変えない。
- **NFR3:** drag-dir-load-stale-marks の維持。`handleDirectoryLoadComplete` でのドラッグの取り消しの挙動と、その単体テスト（`model_update_mouse_test.go` の TS-1..TS-5）を維持する。
- **NFR4:** 既存テストとツールの通過。`go test ./...` と E2E テスト（`make test-e2e-build && make test-e2e`。`mouse_tests.sh` と `mark_tests.sh` を含む）が通る。`gofmt -w .` で差分がなく、`go vet ./...` で指摘がない。

## Implementation Approach

### Architecture

変更はドラッグセッションの状態の判定ロジックと単体テストに限られる。画面の構成や見た目は変えない。

### Data Flow

```
押下（armed にする）
  → ドラッグ中のペインの表示一覧の名前の並びを、ペインの一覧とは独立に保持する

移動・解放
  → 行を解決する前に、ペインの現在の表示一覧の名前の並びと比較する
      一致   → セッションを変えず、元のアンカーからの範囲で処理する（FR3）
      不一致 → mouseDrag = dragSession{} にし、この移動・解放ではマークを変えない（FR1, FR5）
```

判定対象はドラッグ中のペインだけとする（FR4）。判定は表示一覧を変えた経路に依存しない（FR2）。

### API Design

該当なし。

### Database Schema

該当なし。

### Dependencies

**Internal Dependencies:**
- `handleDirectoryLoadComplete`: 既存のドラッグの取り消しの挙動を維持する（NFR3）。
- `RefreshDirectoryPreserveCursor`: カーソル・フィルタ・スクロール・マークの扱いを維持する（NFR2）。

**External Dependencies:**
- 該当なし。

### File Structure

feature-specific のパスは Declared Change Set のとおり create-plan で導出する。

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/drag-refresh-stale-anchor/**`
- `test-docs/drag-refresh-stale-anchor/**`

`feature-docs/drag-refresh-stale-anchor/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/drag-refresh-stale-anchor/**` covers `test-docs/drag-refresh-stale-anchor/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/drag-refresh-stale-anchor/` directory at all; the declared
`test-docs/drag-refresh-stale-anchor/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS-1（FR1, FR7 / AC1, AC8）: 一時ディレクトリを使うモデルで左ペインの項目を押下し、左ペインのディレクトリにファイルを追加して `autoRefreshMsg` を送る。別の行で移動・解放する - セッションが初期値に戻り、左ペインのマーク集合が再読み込み直後の集合と等しい
- [ ] TS-2（FR1, FR5, FR7 / AC2, AC8）: 左ペインで押下と移動を行って範囲をマークし、マークしていないファイルを削除して `autoRefreshMsg` を送る - 再読み込み直後のマークが残る。別の行で移動・解放する - マークがそれ以上変わらない
- [ ] TS-3（FR3 / AC3）: 左ペインで押下し、ファイルを変えずに `autoRefreshMsg` を送る - セッションは armed のまま。移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-4（FR4 / AC4）: 左ペインで押下し、右ペインのディレクトリだけにファイルを追加して `autoRefreshMsg` を送る - 左のセッションは変わらない。移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-5（FR2 / AC5）: 左ペインで押下し、左ペインのディレクトリにファイルを追加して `execFinishedMsg`（および `shellCommandFinishedMsg`）を送る。移動・解放する - マーク集合が再読み込み直後の集合と等しい
- [ ] TS-6（FR3 / AC6）: 左ペインで押下し、名前と順序を変えずに既存ファイルの内容だけを変えて `autoRefreshMsg` を送る。移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-7（FR3 / AC6）: 左ペインにフィルタを適用した状態で押下し、フィルタに一致しないファイルを追加して `autoRefreshMsg` を送る。移動する - 元のアンカーからの範囲がマークされる
- [ ] TS-8（FR6 / AC7）: 左ペインで、フィルタで表示されていないファイルをマークしておく。フィルタ適用中に押下し、そのファイルを削除して `autoRefreshMsg` を送る（表示一覧は変わらない）。移動・解放する - 削除したファイルの名前がマーク集合に含まれない

### Integration Tests
該当なし。

### Static Checks
- [ ] TS-9（NFR4 / AC9）: `go test ./...` がすべて通る。`gofmt -w .` で差分がない。`go vet ./...` で指摘がない

### E2E Tests
**Existing E2E tests**: `mouse_tests.sh`、`mark_tests.sh` を含む既存の E2E テスト
**Run command**: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression

E2E テストは追加しない（FR7）。

### Edge Cases
- [ ] セッションが armed（移動なし）の場合（TS-1）
- [ ] セッションが active で、再読み込み前にマークが付いている場合（TS-2）
- [ ] もう一方のペインだけ一覧が変わる場合（TS-4）
- [ ] 自動更新以外の経路（TS-5）
- [ ] 内容だけが変わる場合（TS-6）
- [ ] フィルタに一致しないファイルだけが増減する場合（TS-7）
- [ ] 表示されていないマーク済みファイルが削除される場合（TS-8）

### Performance Tests
該当なし。

## Security Considerations

該当なし。

## Error Handling

読み込みに失敗して一覧が置き換わらなかった場合は、表示一覧が変わらない場合として扱い、セッションを変えない（FR3）。

## Performance Optimization

該当なし。

## Assumptions

- A1: drag-dir-load-stale-marks の挙動（`handleDirectoryLoadComplete` の成功経路での取り消し、エラーの完了や古い完了ではセッションを変えない）と、その TS-1..TS-5 のテストは維持する。
- A2: `RefreshDirectoryPreserveCursor` のカーソル・フィルタ・スクロール・マークの扱いは変えない。
- A3: ディレクトリ変更時やフィルタ変更時のマークの既存の扱いは、この修正で変えない。

## Success Criteria

- [ ] AC1: 左ペインで押下して armed になった後、左ペインのディレクトリにファイルを追加して表示一覧を変え、`autoRefreshMsg` を処理する。その後、別の行で移動・解放すると、左ペインのマーク集合は再読み込み直後の集合と等しく、セッションは armed ではない。
- [ ] AC2: 左ペインで押下と移動によって範囲をマークした active のセッションで、左ペインのディレクトリのファイルを削除して表示一覧を変え、`autoRefreshMsg` を処理する。その後、別の行で移動・解放すると、マーク集合は再読み込み直後の集合と等しい（再読み込み後に残っているマークはそのまま残る）。
- [ ] AC3: ファイルを変えずに `autoRefreshMsg` を処理した後、左ペインで移動すると、元のアンカーから移動先までの範囲がマークされる。
- [ ] AC4: 右ペインのディレクトリだけにファイルを追加して `autoRefreshMsg` を処理しても、左ペインのセッションは変わらず、その後の移動で元のアンカーからの範囲がマークされる。
- [ ] AC5: 自動更新以外の経路（例: `execFinishedMsg`、`shellCommandFinishedMsg`）でドラッグ中のペインの表示一覧が変わった後も、移動・解放でマークは変わらない。
- [ ] AC6: 名前と順序を変えずにファイルの内容（サイズ）だけを変えた再読み込みの後、およびフィルタに一致しないファイルだけを増減させた再読み込みの後は、移動で元のアンカーからの範囲がマークされる。
- [ ] AC7: 表示一覧が変わらずドラッグが続く場合、再読み込みで消えたファイルの名前はマーク集合に含まれない。
- [ ] AC8: AC1・AC2 のテストは修正前のコードでは失敗し、修正後は通る。
- [ ] AC9: 既存の単体テストと E2E テストが通る。gofmt で差分がなく、go vet で指摘がない。

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし（`status: tbd` の要件はない）。

## References

- 要件定義書: `feature-docs/drag-refresh-stale-anchor/REQUIREMENTS.md`
