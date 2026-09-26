# Feature: fix-batch-cancel-cursor-flake

## Overview

`Pane.GetMarkedFiles` と `Pane.GetMarkedFilePaths` が、マーク済みファイルをペインの表示順（呼び出した時点の entries の並び）で返すようにする。
entries に無いマーク済みファイルは、その後ろにファイル名の昇順で並べる。
あわせて、順序を検証するユニットテストを追加し、E2E `test_batch_cancel_cursor` を強化する。

要件の詳細は [REQUIREMENTS.md](REQUIREMENTS.md) を参照する。

## Objectives

- E2E の `test_batch_cancel_cursor` が毎回成功する
- マーク済みファイルの処理順に依存する不具合の再発を検出するテストがある

## User Stories

### US1: test_batch_cancel_cursor が毎回成功する
E2E の `test_batch_cancel_cursor` が毎回成功する。

**Acceptance Criteria:**
- [ ] AC4: 強化した `test_batch_cancel_cursor` で、上書きダイアログの表示確認、`dst/01_aaa.txt` があること、`src/02_bbb.txt` が残っていること、`dst/02_bbb.txt` の中身が existing のままであることが、すべて成功する
- [ ] AC5: `make test-e2e` を複数回実行して、`test_batch_cancel_cursor` がすべて成功する。既存のユニットテストと E2E テストもすべて成功する

### US2: 処理順に依存する不具合の再発を検出する
マーク済みファイルの処理順に依存する不具合の再発を検出するテストがある。

**Acceptance Criteria:**
- [ ] AC1: マークを付けた順に関係なく、`GetMarkedFiles` が entries の表示順で返す。ユニットテストで検証を複数回繰り返し、すべて一致する
- [ ] AC2: entries に無いマーク済みファイルは、表示中のマーク済みファイルの後ろに名前の昇順で並ぶ。返す件数は `MarkCount()` と一致する
- [ ] AC3: `GetMarkedFilePaths` の i 番目が `filepath.Join(ペインのパス, GetMarkedFiles の i 番目)` と一致する

## Technical Requirements

### Functional Requirements
- **FR1:** マーク済みファイルを表示順で返す - `Pane.GetMarkedFiles` は、マーク済みファイル名をペインの表示順（呼び出した時点の entries の並び。ディレクトリ優先と現在のソートを反映したもの）で返す。マークを付けた順序や map の反復順には依存しない
- **FR2:** 表示一覧に無いマーク済みファイルの扱い - マーク済みで、呼び出した時点の entries に無い名前（フィルタで非表示になったものなど）は、FR1 の並びの後ろに、ファイル名の昇順で追加する。`GetMarkedFiles` が返す件数は `MarkCount()` と一致する
- **FR3:** 完全パス版も同じ順序にする - `Pane.GetMarkedFilePaths` は、`GetMarkedFiles` と同じ順序で、各名前をペインのパスと結合した完全パスを返す
- **FR4:** 再発を検出するユニットテスト - `internal/ui/pane_mark_test.go` に、マークを付ける順を表示順と変えた状態で `GetMarkedFiles` と `GetMarkedFilePaths` の順序を検証するテストを追加する。map の反復順が毎回変わることを踏まえ、同じ検証を複数回繰り返す
- **FR5:** E2E test_batch_cancel_cursor の強化 - `test_batch_cancel_cursor` で、固定時間の sleep の代わりに、上書きダイアログ（"already exists"）が表示されるまで上限付きで待ってからキー 2（Cancel）を送る。キャンセル後に、`dst/01_aaa.txt` があること、`src/02_bbb.txt` が残っていること、`dst/02_bbb.txt` の中身が existing のままであることを検証する

### Non-Functional Requirements
- **NFR1:** `GetMarkedFiles` と `GetMarkedFilePaths` のシグネチャは変えない
- **NFR2:** 変えるのは返す順序だけ。返す要素の集合と件数は変更前と同じ
- **NFR3:** `go test ./...` と `make test-e2e` がすべて成功する
- **NFR4:** gofmt と go vet で指摘が出ない

## Implementation Approach

### Architecture

該当なし（構成の変更は無い。変更は `internal/ui` の `Pane` が返す順序とテストに限る）

**Component Diagram:**
```
Pane (internal/ui)
├── entries             呼び出した時点の表示一覧（ディレクトリ優先と現在のソートを反映）
├── マーク              map で保持
├── MarkCount()         マーク件数
├── GetMarkedFiles()    FR1, FR2 の順序で名前を返す
└── GetMarkedFilePaths() GetMarkedFiles と同じ順序で完全パスを返す（FR3）
```

### Data Flow

`GetMarkedFiles` の並び:

```
1. 呼び出した時点の entries を先頭から順に見て、マーク済みの名前を結果に追加する（FR1）
2. entries に無いマーク済みの名前を集め、ファイル名の昇順に並べて結果の末尾に追加する（FR2）
3. 結果の件数は MarkCount() と一致する
```

`GetMarkedFilePaths` の並び:

```
GetMarkedFiles の i 番目 → filepath.Join(ペインのパス, 名前) → 結果の i 番目（FR3）
```

### API Design

該当なし（`GetMarkedFiles` と `GetMarkedFilePaths` のシグネチャは変えない。NFR1）

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- `ApplyFilter`: allEntries から entries を作り直すが、マークは解除しない。このため entries に無いマーク済みファイルが起こりうる（A-filter-keeps-marks）
- `LoadDirectory`: マークを解除する（A-load-clears-marks）
- ディレクトリ優先: 固定の挙動で、entries に反映されている（A-dirs-first）

**External Dependencies:**
- なし

### File Structure

```
internal/ui/
├── pane_marks.go          # GetMarkedFiles / GetMarkedFilePaths（FR1〜FR3）
└── pane_mark_test.go      # 順序を検証するユニットテスト（FR4）
test/e2e/scripts/
├── run_all_tests.sh
└── tests/
    └── cursor_preserve_tests.sh   # test_batch_cancel_cursor（FR5）
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/fix-batch-cancel-cursor-flake/**`
- `test-docs/fix-batch-cancel-cursor-flake/**`

`feature-docs/fix-batch-cancel-cursor-flake/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/fix-batch-cancel-cursor-flake/**` covers `test-docs/fix-batch-cancel-cursor-flake/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/fix-batch-cancel-cursor-flake/` directory at all; the declared
`test-docs/fix-batch-cancel-cursor-flake/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] T1 (FR1, FR4 / AC1): 表示順と逆の順でマークし、`GetMarkedFiles` が entries の表示順で返すことを複数回繰り返して検証する
- [ ] T2 (FR1, FR4 / AC1): ディレクトリとファイルを混ぜてマークし、`GetMarkedFiles` の並びが entries の並び（ディレクトリ優先）と一致することを検証する
- [ ] T3 (FR2, NFR2 / AC2): マーク済みファイルの一部を entries に無い状態（フィルタで非表示）にし、表示中のものは表示順、非表示のものはその後ろに名前の昇順で並ぶこと、件数が `MarkCount()` と一致することを検証する
- [ ] T4 (FR1, FR4 / AC1): マークした後に entries の並びが変わった場合、呼び出した時点の entries の順に従うことを検証する
- [ ] T5 (FR3 / AC3): `GetMarkedFilePaths` の順序と内容が、`GetMarkedFiles` の各要素をペインのパスと結合したものと一致することを検証する
- [ ] T6 (FR1, FR2, FR4, NFR2 / AC1, AC2): マークが 0 件と 1 件の場合に、今までどおりの結果になることを検証する

すべて `internal/ui/pane_mark_test.go` に置く。

### Integration Tests
- なし

### E2E Tests
**Existing E2E tests**: `test/e2e/scripts/run_all_tests.sh`（`test_batch_cancel_cursor` は `test/e2e/scripts/tests/cursor_preserve_tests.sh`）
**Run command**: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] T7 (FR5 / AC4): `test_batch_cancel_cursor`: 01_aaa と 02_bbb をマークして m を押し、"already exists" が表示されるまで上限付きで待ってから 2 を送る。`dst/01_aaa.txt` があること、`src/02_bbb.txt` が残っていること、`dst/02_bbb.txt` の中身が existing であることを検証する
- [ ] T8 (NFR3 / AC5): `make test-e2e` を複数回実行して、`test_batch_cancel_cursor` を含む全テストが毎回成功することを確認する。回数は verify で決める（A-repeat-e2e）

### Edge Cases
- [ ] マークが 0 件と 1 件: 今までどおりの結果になる（T6）
- [ ] entries に無いマーク済みファイル: 表示中のマーク済みファイルの後ろに名前の昇順で並び、件数は `MarkCount()` と一致する（T3）
- [ ] マーク後に entries の並びが変わる: 呼び出した時点の entries の順に従う（T4）
- [ ] ディレクトリとファイルの混在: entries の並び（ディレクトリ優先）と一致する（T2）

### Performance Tests
- 該当なし

## Security Considerations

該当なし

## Error Handling

該当なし

## Performance Optimization

該当なし

## Success Criteria

- [ ] AC1〜AC5 をすべて満たす
- [ ] T1〜T8 がすべて成功する
- [ ] `go test ./...` と `make test-e2e` がすべて成功する（NFR3）
- [ ] gofmt と go vet で指摘が出ない（NFR4）

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

## Implementation Phases (if applicable)

該当なし

## References

- 要件定義書: `feature-docs/fix-batch-cancel-cursor-flake/REQUIREMENTS.md`
