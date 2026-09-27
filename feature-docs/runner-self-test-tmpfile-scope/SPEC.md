# Feature: runner-self-test-tmpfile-scope

## Overview

`test/e2e/scripts/runner_self_test.sh`（runner self-test）の `apply_fixture_lists` が作る一時ファイルを、自己テスト自身が作る一時ディレクトリ `SELF_TMPDIR` の中に作るようにする。あわせて、`SELF_TMPDIR` の外に一時ファイルを作る状態に戻ったことを検出する回帰シナリオ TS-6 を同スクリプトに追加する。

## Objectives

- runner self-test は、自身が作る一時ディレクトリ `SELF_TMPDIR` の外に何も書き込まない。スクリプト冒頭のコメント（17 行目）に書かれた約束と一致させる。
- `SELF_TMPDIR` の外に一時ファイルを作る状態に戻ったとき、回帰テストで検出する。

## Technical Requirements

### Functional Requirements

- **FR1: `apply_fixture_lists` は一時ファイルを `SELF_TMPDIR` 内に作る**
  `apply_fixture_lists` が作る 2 つの中間一時ファイル（現行コードでは `test/e2e/scripts/runner_self_test.sh` の 100 行目と 107 行目の `tmp="$(mktemp)"`）を `SELF_TMPDIR` の中に作る。`$TMPDIR` や `/tmp` の直下には作らない。関数の引数、awk によるリスト置換の動作、`chmod +x` による実行権限の復元、fail-closed の戻り値の約束は変えない。

- **FR2: `SELF_TMPDIR` 外の一時ファイルを検出する回帰シナリオ**
  `runner_self_test.sh` に新しいシナリオ TS-6 を追加する。TS-6 は、空の番兵ディレクトリを `TMPDIR` として export した状態で `apply_fixture_lists` を実行する。番兵ディレクトリは `SELF_TMPDIR` の中に置く。TS-6 は一時ファイルの作成から `mv` までの間で失敗を起こさせる（例: 存在しない runner copy のパスを渡して awk を失敗させる）。その後、番兵ディレクトリが空のままであることを確認する。現行コードでは素の `mktemp` が番兵 `TMPDIR` に孤立したファイルを残すので TS-6 は FAIL し、FR1 の適用後は PASS する。TS-6 は既存の `run_scenario` の PASS/FAIL 形式で結果を出し、失敗時は `OVERALL_STATUS` を設定する。

### Non-Functional Requirements

- **NFR1 - 自己テストの書き込み範囲:** TS-6 とその番兵ディレクトリを含め、自己テストの実行全体を通じて、ファイルの作成・変更・残置は `SELF_TMPDIR` の下だけで行う。`SELF_TMPDIR` は既存の EXIT trap が削除する。
- **NFR2 - tmux セッションと duofm プロセスを起動しない:** TS-6 は TS-1〜TS-5 と同様に、tmux セッションも duofm プロセスも起動しない。
- **NFR3 - 既存動作の維持:** TS-1〜TS-5 はアサーションを変えずに PASS し続ける。スクリプトの終了ステータスの約束（全シナリオが PASS したときに限り 0）は変えない。Dockerfile の既定 CMD の連鎖（`runner_self_test.sh && run_all_tests.sh`）は変更せず、E2E イメージ内で非 root ユーザー `testuser` として引き続き動作する。

## Acceptance Criteria

- [ ] **AC-1**（FR1）: 空ディレクトリ D を `TMPDIR` として export し、`apply_fixture_lists` を一時ファイル作成後に失敗させたあとも、D は空のままである。素の `mktemp` の結果と同じ形の名前（`tmp.XXXXXXXXXX`）のファイルが D に現れない。
- [ ] **AC-2**（FR2）: 修正後、`bash test/e2e/scripts/runner_self_test.sh` は `TS-6 ...: PASS` を出力する。`apply_fixture_lists` のどちらかの一時ファイル作成を素の `mktemp` に戻すと、TS-6 は FAIL を出力し、スクリプトは非 0 で終了する。
- [ ] **AC-3**（NFR3）: `bash test/e2e/scripts/runner_self_test.sh` は TS-1〜TS-6 がすべて PASS し、終了ステータス 0 で終わる。
- [ ] **AC-4**（NFR1, NFR3）: `make test-e2e-build && make test-e2e` が成功する。コンテナの既定 CMD は、E2E スイートの前に TS-6 を含む自己テストを実行する。

## Implementation Approach

### Components

- `apply_fixture_lists`（FR1）: 2 箇所の一時ファイル作成先を `SELF_TMPDIR` 内にする。
- TS-6（FR2）: 新規シナリオ。既存の `run_scenario` で実行し、結果を報告する。
- スクリプト冒頭のコメント（8 行目の `Scenarios TS-1 through TS-5`）: シナリオの範囲を TS-6 まで含むように更新する（A-4）。

### Data Flow（TS-6）

```
SELF_TMPDIR 内に空の番兵ディレクトリを作る
  → 番兵ディレクトリを TMPDIR として export する
  → 存在しない runner copy のパスで apply_fixture_lists を呼ぶ
     （一時ファイルの作成後に awk が失敗し、mv は行われない）
  → 番兵ディレクトリにエントリが無いことを確認する
  → run_scenario の PASS/FAIL 形式で報告し、失敗時は OVERALL_STATUS を設定する
```

### Notes

- runner copy のパスが存在しない場合、`apply_fixture_lists` は 0 を返す。awk が失敗し、存在しないファイルへの `grep -q` が 2 を返すので、本番エントリの検査がどちらも発火しないためである。したがって TS-6 は、この強制失敗経路での `apply_fixture_lists` の戻り値をアサートしない。

### Dependencies

**Internal Dependencies:**
- `SELF_TMPDIR` と既存の EXIT trap（`cleanup_self_tmpdir`）: 一時ファイルと番兵ディレクトリの置き場所と、その削除。
- `run_scenario` / `OVERALL_STATUS`: TS-6 の結果報告と失敗の記録。

### File Structure

```
test/e2e/scripts/
└── runner_self_test.sh   # apply_fixture_lists の修正、TS-6 の追加、冒頭コメントの更新
```

## Assumptions

- **A-1:** 回帰テストは別スクリプトではなく、`runner_self_test.sh` 内のシナリオ TS-6 として追加する。
- **A-2:** TS-6 は成功経路ではなく、`mktemp` と `mv` の間で失敗を起こさせることで不具合を観測する。
- **A-3:** 途中の失敗で `SELF_TMPDIR` 内に孤立した一時ファイルは、既存の EXIT trap（`cleanup_self_tmpdir`）が削除する。`apply_fixture_lists` には呼び出しごとの後始末を追加しない。
- **A-4:** 冒頭コメントのシナリオ範囲（8 行目の `Scenarios TS-1 through TS-5`）を TS-6 まで含むように更新する。
- **A-5:** チケットの「追加指示」節に書かれたワークフロー上の作業は、本機能の機能要件ではなく、本仕様には含めない。

## Declared Change Set

本機能固有のパスは手書きの一覧ではなく、create-plan で導出する。上記の機能固有パスは、create-plan の時点で `workflow.yaml` の全タスクの `files` エントリから導出する（`references/phases/create-plan-phase.md`）。

すべての SPEC は、既定で、上記の機能固有パスに加えて次のワークフロー生成エントリ 2 つを宣言する。

- `feature-docs/runner-self-test-tmpfile-scope/**`
- `test-docs/runner-self-test-tmpfile-scope/**`

`feature-docs/runner-self-test-tmpfile-scope/**` は `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、および design ステップが生成する設計成果物を含む。これらは各フェーズ文書と `references/phase-state.md` が生成・所有する。本節はそれらを参照するだけで、その規則は再掲しない。

`test-docs/runner-self-test-tmpfile-scope/**` はタスクごとのテスト記録 `test-docs/runner-self-test-tmpfile-scope/{T}.tests.yaml` を含む。これは `implement-phase.md` が生成・所有する。本節はそれを参照するだけで、その規則は再掲しない。

この 2 つの既定エントリは、SPEC の作成者が明示的に取り除かない限り宣言に含まれる。記載が無いことをもって除外とはみなさない。除外は意図的かつ明示的な絞り込みとして行う。

この宣言は上位集合の宣言である。検証時に観測される実際の変更集合は、宣言した集合と等しい必要はなく、宣言した集合に含まれていればよい。implement タスクを生成しない機能では `test-docs/runner-self-test-tmpfile-scope/` ディレクトリ自体が作られないが、その場合も `test-docs/runner-self-test-tmpfile-scope/**` の宣言は正しい。宣言したパスが実際に作られないことは違反ではない。

## Test Scenarios

### Self-Test Scenarios

- [ ] **TS-6**（FR1, FR2 / AC-1, AC-2）: `SELF_TMPDIR` 内に空の番兵ディレクトリを作る。番兵ディレクトリを `TMPDIR` として export した状態で、存在しない runner copy のパスに対して `apply_fixture_lists` を呼ぶ。一時ファイルの作成後に awk が失敗し、`mv` は行われない。番兵ディレクトリにエントリが無いことを確認する。

### Mutation Check

- [ ] **MUT-1**（AC-2）: 検証時に行う変異確認。100 行目または 107 行目を一時的に `tmp="$(mktemp)"` に戻し、TS-6 が FAIL してスクリプトが非 0 で終了することを確認したあと、修正を元に戻す。

### E2E Tests

**Existing E2E tests**: `test/e2e/scripts/runner_self_test.sh`、`run_all_tests.sh`（E2E コンテナの既定 CMD で連鎖実行）
**Run command**: `make test-e2e-build && make test-e2e`

- [ ] Existing E2E tests pass without regression
- [ ] **REG-1**（NFR3 / AC-3, AC-4）: `bash test/e2e/scripts/runner_self_test.sh` を直接実行した場合と、E2E コンテナの既定 CMD 経由で実行した場合の両方で、TS-1〜TS-6 がすべて PASS し、終了ステータスが 0 になる。

## Success Criteria

- [ ] FR1、FR2 が実装され、テストされている
- [ ] AC-1〜AC-4 を満たす
- [ ] TS-6、MUT-1、REG-1 がすべて期待どおりの結果になる

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

## References

- runner self-test: `test/e2e/scripts/runner_self_test.sh`
