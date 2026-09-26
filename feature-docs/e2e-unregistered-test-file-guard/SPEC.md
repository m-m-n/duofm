# Feature: e2e-unregistered-test-file-guard

## Overview

E2E の全体実行で、`TEST_FILES` に登録されていないファイルのテストが `FULL_RUN_LIST` にあると、そのテストが集計から漏れて終了コード 0 で終わる。本機能では、`run_test` が未定義の関数を失敗として数え、run-list ガードが未登録ファイルを全体実行と `--check-list` の両方で報告する。再発は `make test-e2e` のたびに走る bash の自己テストで検出する。

要件の詳細は `feature-docs/e2e-unregistered-test-file-guard/REQUIREMENTS.md` を参照。

## Objectives

- E2E の全体実行で、`TEST_FILES` に登録されていないファイルのテストが実行一覧（`FULL_RUN_LIST`）にあるとき、そのテストが集計から漏れて終了コード 0 で終わる状態を無くす
- `TEST_FILES` への登録漏れを、全体実行と `--check-list` の両方でファイル名つきの失敗として検出する
- 同じ不具合の再発を、`make test-e2e` のたびに走る自己テストで検出する

## User Stories

### US1: 全体実行で登録漏れを検出する
E2E テストを追加する開発者として、`TEST_FILES` への登録を忘れたファイルのテストを `FULL_RUN_LIST` に加えたとき、全体実行が失敗で終わってほしい。

**Acceptance Criteria:**
- [ ] AC-1: 再現手順（`tests/` にテストファイルを追加し、`TEST_FILES` に登録せず、関数名だけを `FULL_RUN_LIST` に加えて全体実行する）で、終了コードが 0 以外になり、Failed が 1 以上になり、出力に未登録ファイル名とそのテスト関数名が失敗として出る。
- [ ] AC-3: `run_test` に未定義の名前を渡すと、TESTS_RUN と TESTS_FAILED がそれぞれちょうど 1 増え、TESTS_PASSED は変わらない。
- [ ] AC-4: 現在のリポジトリ（`tests/*.sh` の 15 ファイルがすべて `TEST_FILES` に登録済み）では、`--check-list` が OK 行を出して終了コード 0 で終わり、全体実行でガード由来の失敗が新しく出ない。

### US2: `--check-list` で登録漏れを検出する
E2E テストを追加する開発者として、`--check-list` で未登録ファイルを知りたい。

**Acceptance Criteria:**
- [ ] AC-2: AC-1 と同じ状態で `--check-list` を実行すると、終了コード 1 で終わり、出力に未登録ファイル名が出る。

### US3: 再発を検出する
開発者として、同じ不具合の再発を `make test-e2e` のたびに検出したい。

**Acceptance Criteria:**
- [ ] AC-5: 自己テストは、修正前の `run_all_tests.sh` と `helpers.sh` に対して失敗し、修正後は成功する。
- [ ] AC-6: `make test-e2e-build && make test-e2e` を実行するたびに自己テストが走り、自己テストが失敗したら `make test-e2e` の終了コードが 0 以外になる。
- [ ] AC-7: 自己テストの実行中に、tmux セッションと duofm プロセスが起動されない。

## Technical Requirements

### Functional Requirements

- **FR1:** `run_test` が未定義の関数を失敗 1 件として数える。`test/e2e/scripts/helpers.sh` の `run_test` は、渡された名前がシェル関数として定義されていないとき、その名前を呼び出さず、TESTS_RUN と TESTS_FAILED をそれぞれ 1 増やし、関数名を含む失敗行を出力する。定義済みかどうかは関数定義の有無で判定し、終了コード（127 など）では判定しない。定義済みの関数に対する挙動（呼び出し、戻り値を集計に使わないこと、tmux セッションの後片付け）は変えない。
- **FR2:** run-list ガードが `TEST_FILES` 未登録のファイルを検出する。`test/e2e/scripts/run_all_tests.sh` の `run_guard_checks` は、`tests/*.sh` のうち、ファイル名（basename）が `TEST_FILES` のどの値にも無いものを未登録ファイルとして集め、1 件以上あれば問題あり（戻り値 1）とする。既存の未定義エントリ（GUARD_UNDEFINED）と一覧漏れ（GUARD_MISSING）の判定は変えない。全体実行と `--check-list` は、引き続きこの同じガードを使う。
- **FR3:** 全体実行で未登録ファイルを失敗として報告する。引数なしの全体実行（`run_all`）は、未登録ファイル 1 件ごとに TESTS_RUN と TESTS_FAILED を 1 ずつ増やし、ファイル名を含む失敗行を出力する。未登録ファイルが 1 件でもあれば `print_summary` の Failed は 1 以上になり、終了コードは 0 以外になる。未登録ファイルにあるテストが `FULL_RUN_LIST` に載っている場合、そのテストは FR1 によっても失敗 1 件として数えられる。
- **FR4:** `--check-list` で未登録ファイルを報告する。`--check-list` は、未登録ファイル 1 件ごとにファイル名を含む行を出力し、終了コード 1 で終わる。OK 行（`OK: run list matches defined tests ...`）と終了コード 0 になるのは、未定義エントリ・一覧漏れ・未登録ファイルがすべて 0 件のときだけとする。tmux と duofm は起動しない（現状どおり）。
- **FR5:** 再発を検出する bash の自己テスト。bash の自己テストスクリプトを `test/e2e/scripts/` の下の `tests/` 以外の場所に新しく置く。自己テストは、`run_all_tests.sh` と `helpers.sh` の複製を一時ディレクトリに置き、tmux も duofm も起動しないテスト関数を持つ検証用の `tests/` ディレクトリと、検証用の `TEST_FILES` / `FULL_RUN_LIST` を用意して、TS-1 から TS-5 を検証する。シナリオごとに結果を出力し、1 つでも失敗すれば終了コード 0 以外で終わる。
- **FR6:** `make test-e2e` で自己テストを毎回走らせる。E2E イメージ（`test/e2e/Dockerfile`）の既定コマンドは、自己テストを先に実行し、成功したときに `run_all_tests.sh` を実行する。`make test-e2e`（`docker run --rm duofm-e2e-test`）を実行するたびに、コンテナ内で自己テストが走る。自己テストが失敗したら `make test-e2e` は終了コード 0 以外で終わる。

### Non-Functional Requirements

- **NFR1 - 自己テストは tmux と duofm を起動しない:** 自己テストとその検証用テスト関数は、tmux セッションも duofm も起動しない。`run_test` 内の後片付け（`tmux kill-session ... || true`）は起動にあたらない。
- **NFR2 - 書き込み範囲:** 自己テストが書き込むのは、自分で作った一時ディレクトリ（mktemp）の下だけとする。`/testdata` や `/e2e/scripts` は変更しない。コンテナ内では testuser として動く。
- **NFR3 - 依存を増やさない:** 自己テストが使うのは、E2E イメージに既にある bash と基本コマンドだけとする。新しいパッケージは追加しない。
- **NFR4 - 既存の CLI を保つ:** `run_all_tests.sh` のカテゴリ指定実行・`--list`・`--help` の挙動と、`tests/*.sh` を単体で実行したときの挙動（未定義の関数を数える FR1 は除く）は変えない。

## Implementation Approach

### Architecture

**Component Diagram:**

| コンポーネント | 変更内容 | 要件 |
|----------------|----------|------|
| `test/e2e/scripts/helpers.sh` の `run_test` | 未定義の名前を呼び出さず、失敗 1 件として数える | FR1 |
| `test/e2e/scripts/run_all_tests.sh` の `run_guard_checks` | 未登録ファイルを集め、1 件以上なら戻り値 1 | FR2 |
| `test/e2e/scripts/run_all_tests.sh` の `run_all` | 未登録ファイル 1 件ごとに失敗 1 件を数え、ファイル名を含む失敗行を出す | FR3 |
| `test/e2e/scripts/run_all_tests.sh` の `--check-list` | 未登録ファイル 1 件ごとにファイル名を含む行を出し、終了コード 1 | FR4 |
| 自己テストスクリプト（新規） | TS-1 から TS-5 を検証する | FR5 |
| `test/e2e/Dockerfile` の既定コマンド | 自己テストを先に実行し、成功したら `run_all_tests.sh` を実行する | FR6 |

### Data Flow

自己テスト（FR5）:

```
mktemp で一時ディレクトリを作る
  → run_all_tests.sh と helpers.sh を複製する
  → 検証用の tests/ と TEST_FILES / FULL_RUN_LIST を用意する
  → TS-1 から TS-5 を実行し、シナリオごとに結果を出力する
  → 1 つでも失敗すれば終了コード 0 以外
```

E2E イメージの既定コマンド（FR6）:

```
自己テスト ── 成功 → run_all_tests.sh
          └─ 失敗 → run_all_tests.sh を実行せず、終了コード 0 以外
```

### API Design

該当なし。

### Database Schema

該当なし。

### Dependencies

**Internal Dependencies:**
- `test/e2e/scripts/helpers.sh`: `run_test` と集計変数（TESTS_RUN / TESTS_PASSED / TESTS_FAILED）
- `test/e2e/scripts/run_all_tests.sh`: `run_all`・`check_list`・`run_guard_checks`・`print_summary`、`TEST_FILES`・`FULL_RUN_LIST`
- `test/e2e/Dockerfile`: E2E イメージの既定コマンド
- `Makefile`: `test-e2e-build` / `test-e2e` ターゲット

**External Dependencies:**
- 追加なし（E2E イメージに既にある bash と基本コマンドだけを使う。NFR3）

### File Structure

```
test/e2e/
├── Dockerfile               # 既定コマンド (FR6)
└── scripts/
    ├── helpers.sh           # run_test (FR1)
    ├── run_all_tests.sh     # run_guard_checks / run_all / --check-list (FR2, FR3, FR4)
    ├── {自己テストスクリプト}  # 新規 (FR5)。tests/ の外に置く
    └── tests/
        └── *.sh             # 未登録ファイル判定の対象 (FR2)
```

## Declared Change Set

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

上記のフィーチャー固有のパスに加えて、次の 2 つのワークフロー生成エントリをデフォルトで宣言する:

- `feature-docs/e2e-unregistered-test-file-guard/**`
- `test-docs/e2e-unregistered-test-file-guard/**`

`feature-docs/e2e-unregistered-test-file-guard/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/e2e-unregistered-test-file-guard/**` に含まれるもの: `test-docs/e2e-unregistered-test-file-guard/{T}.tests.yaml`（タスクごとのテスト記録）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

この 2 つのデフォルトエントリは、SPEC作成者が明示的に除外しない限り宣言に含まれる。除外は意図的な絞り込みであり、記載漏れによる省略ではない。

この宣言はスーパーセット（superset）の主張であり、検証時に観測される実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。一致は求めない。implementタスクを1つも生成しないフィーチャーは `test-docs/e2e-unregistered-test-file-guard/` ディレクトリを生成しないが、宣言された `test-docs/e2e-unregistered-test-file-guard/**` は依然として正しい。実際には生成されないパスが宣言されていても違反にはならない。

## Test Scenarios

### Unit Tests
- [ ] TS-5: `helpers.sh` を読み込み、定義されていない名前で `run_test` を呼ぶ - TESTS_RUN と TESTS_FAILED がそれぞれ 1 増え、TESTS_PASSED は変わらない。出力にその名前が失敗として出る。（FR1, AC-3）

### Integration Tests
自己テスト（FR5）が、複製したランナーと検証用の `tests/`・`TEST_FILES`・`FULL_RUN_LIST` で検証する。

- [ ] TS-1: 未登録ファイルのテストが実行一覧にある全体実行。検証用 `tests/` に、登録済みファイル（成功するテスト 1 件。TESTS_RUN と TESTS_PASSED を 1 ずつ増やす）と未登録ファイル（テスト 1 件）を置き、`FULL_RUN_LIST` に両方のテストを載せる - 終了コードは 0 以外。Total 3 / Passed 1 / Failed 2（未登録ファイル 1 件 + 未定義として呼ばれたテスト 1 件）。出力に未登録ファイル名とそのテスト関数名が出る。（FR1, FR2, FR3, AC-1）
- [ ] TS-2: 未登録ファイルのテストが実行一覧にある `--check-list`。TS-1 と同じ検証用の構成 - 終了コードは 1。出力に未登録ファイル名が出て、OK 行は出ない。（FR2, FR4, AC-2）
- [ ] TS-3: すべて登録済みの構成（誤検出が無いこと）。検証用 `tests/` のファイルがすべて `TEST_FILES` に登録され、定義されたテストがすべて `FULL_RUN_LIST` に載っている - `--check-list` は OK 行を出して終了コード 0。全体実行は Failed 0 で終了コード 0。（FR2, FR3, FR4, AC-4）
- [ ] TS-4: 未登録ファイルのテストが実行一覧に無い場合。未登録ファイルのテストを `FULL_RUN_LIST` に載せない - `--check-list` は終了コード 1 で、出力に未登録ファイル名と、既存の一覧漏れ行（`Missing from run list`）の両方が出る。（FR2, FR4）

### E2E Tests
**Existing E2E tests**: `test/e2e/scripts/run_all_tests.sh`
**Run command**: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] `make test-e2e` のたびにコンテナ内で自己テストが走り、自己テストが失敗したら終了コード 0 以外で終わる（AC-6）
- [ ] 現在のリポジトリで `--check-list` が OK 行を出して終了コード 0 で終わり、全体実行でガード由来の失敗が新しく出ない（AC-4）
- [ ] 自己テストの実行中に tmux セッションと duofm プロセスが起動されない（AC-7）

### Edge Cases
- [ ] 未登録ファイルのテストが `FULL_RUN_LIST` に無い場合も、`--check-list` は未登録ファイルを報告する（TS-4）
- [ ] `TEST_FILES` の値が存在しないファイルを指している場合の挙動は変えない（対象外、A-4）
- [ ] 定義されていてもアサーションを 1 件も実行しないテスト関数が集計に出ないことは対象外（A-5）

### Performance Tests

該当なし。

## Security Considerations

- **Data Protection:** 自己テストが書き込むのは、自分で作った一時ディレクトリ（mktemp）の下だけ。`/testdata` や `/e2e/scripts` は変更しない。コンテナ内では testuser として動く（NFR2）。
- その他の項目（Authentication / Authorization / Input Validation / XSS / SQL Injection / CSRF）: 該当なし。

## Error Handling

### Error Codes

| 事象 | 集計 | 出力 | 終了コード |
|------|------|------|------------|
| `run_test` に未定義の名前が渡された | TESTS_RUN +1、TESTS_FAILED +1 | 関数名を含む失敗行 | - |
| 全体実行で未登録ファイルがある | ファイル 1 件ごとに TESTS_RUN +1、TESTS_FAILED +1 | ファイル名を含む失敗行 | 0 以外 |
| `--check-list` で未登録ファイルがある | - | ファイル 1 件ごとにファイル名を含む行。OK 行は出ない | 1 |
| 自己テストのシナリオが 1 つ以上失敗した | - | シナリオごとの結果 | 0 以外（`make test-e2e` も 0 以外） |

### Error Flow

```
run_test(name) → name が関数として定義されているか
  ├─ 定義済み → 呼び出す（既存の挙動）
  └─ 未定義   → 呼び出さず、TESTS_RUN +1、TESTS_FAILED +1、失敗行を出力
```

## Performance Optimization

該当なし。

## Success Criteria

- [ ] AC-1 から AC-7 を満たす
- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass

## Assumptions

- A-1: FR6 で自己テストが失敗したときは `run_all_tests.sh` を実行せず、終了コード 0 以外で終わる。
- A-2: 複製したランナーに検証用の `TEST_FILES` / `FULL_RUN_LIST` を与える方法（テキスト置換か、main を実行せずにランナーを読み込んで配列を差し替えるか）は実装で決める。要件として求めるのは、`run_all`・`check_list`・`run_guard_checks`・`run_test` の本番のロジックを自己テストが実際に通すことだけとする。
- A-3: 未登録ファイルはファイル 1 件ごとに失敗 1 件として数え、そのファイルにあって `FULL_RUN_LIST` に載っているテストは、それとは別に FR1 で 1 件ずつ数える。
- A-4: `TEST_FILES` の値が存在しないファイルを指している場合の挙動（`set -e` の下で source が失敗して異常終了する）は変えない。対象外とする。
- A-5: 定義されていてもアサーションを 1 件も実行しないテスト関数が集計に出ないことは、対象外とする。

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

なし（`status: tbd` の要件は無い）。

## References

- 要件定義書: `feature-docs/e2e-unregistered-test-file-guard/REQUIREMENTS.md`
