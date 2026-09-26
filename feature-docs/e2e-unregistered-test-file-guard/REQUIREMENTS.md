---
title: "e2e-unregistered-test-file-guard"
created_date: 2026-09-27
status: draft
---

# e2e-unregistered-test-file-guard - 要件定義書

## 1. 概要

### 1.1 背景

`test/e2e/scripts/run_all_tests.sh` の `run_all` は、`TEST_FILES` に登録されたファイルだけを source する。一方、run-list ガードの未定義エントリ判定は、`tests/*.sh` の静的な名前集合で行っている。`test/e2e/scripts/helpers.sh` の `run_test` は、関数呼び出しの失敗（command not found）を集計しない。

このため、`TEST_FILES` に登録されていないファイルのテストを `FULL_RUN_LIST` に加えると、ガードは通り、本実行では未定義のまま呼ばれ、Total にも Failed にも数えられず、終了コード 0 で終わる。

現状は `tests/` のファイルがすべて `TEST_FILES` に登録されているため、この現象は起きていない。

### 1.2 目的

未登録ファイルのテストが実行一覧にある場合、実行されるか、失敗として数えられるようにする。

### 1.3 スコープ

対象:

- `test/e2e/scripts/helpers.sh` の `run_test`
- `test/e2e/scripts/run_all_tests.sh` の `run_guard_checks`・`run_all`（引数なしの全体実行）・`--check-list`
- 新規の bash 自己テストスクリプト（`test/e2e/scripts/` の下、`tests/` の外）
- `test/e2e/Dockerfile` の既定コマンド

対象外:

- `TEST_FILES` の値が存在しないファイルを指している場合の挙動（A-4）
- 定義されていてもアサーションを 1 件も実行しないテスト関数が集計に出ないこと（A-5）

## 2. ビジネス要件

### 2.1 ビジネス目標

- E2E の全体実行で、`TEST_FILES` に登録されていないファイルのテストが実行一覧（`FULL_RUN_LIST`）にあるとき、そのテストが集計から漏れて終了コード 0 で終わる状態を無くす
- `TEST_FILES` への登録漏れを、全体実行と `--check-list` の両方でファイル名つきの失敗として検出する
- 同じ不具合の再発を、`make test-e2e` のたびに走る自己テストで検出する

### 2.2 対象ユーザー

| ユーザータイプ | 説明 |
|----------------|------|
| 開発者 | E2E テストを追加・実行する |

### 2.3 期待される効果

- 再現手順で現象が起きない
- 再発を検出するテストがある

## 3. ユースケース

### 3.1 ユースケース一覧

| ID | ユースケース名 | アクター |
|----|----------------|----------|
| UC01 | 全体実行で登録漏れを検出する | 開発者 |
| UC02 | `--check-list` で登録漏れを検出する | 開発者 |
| UC03 | `make test-e2e` で自己テストを走らせる | 開発者 |

### 3.2 ユースケース詳細

#### UC01: 全体実行で登録漏れを検出する

**アクター**: 開発者

**事前条件**:
- `test/e2e/scripts/tests/` に、`TEST_FILES` に登録されていないテストファイルがある

**基本フロー**:
1. `make test-e2e-build && make test-e2e` を実行する
2. 未登録ファイル 1 件ごとに、ファイル名を含む失敗行が出て、失敗 1 件として数えられる
3. 未登録ファイルにあるテストが `FULL_RUN_LIST` に載っている場合、そのテストは未定義の関数として失敗 1 件に数えられる

**事後条件**:
- `print_summary` の Failed が 1 以上になり、終了コードが 0 以外になる

#### UC02: `--check-list` で登録漏れを検出する

**アクター**: 開発者

**事前条件**:
- `test/e2e/scripts/tests/` に、`TEST_FILES` に登録されていないテストファイルがある

**基本フロー**:
1. `run_all_tests.sh --check-list` を実行する
2. 未登録ファイル 1 件ごとに、ファイル名を含む行が出る

**事後条件**:
- 終了コード 1 で終わり、OK 行は出ない
- tmux と duofm は起動しない

#### UC03: `make test-e2e` で自己テストを走らせる

**アクター**: 開発者

**基本フロー**:
1. `make test-e2e` を実行する
2. コンテナ内で自己テストが走る
3. 自己テストが成功したら、`run_all_tests.sh` が走る

**代替フロー**:
- 自己テストが失敗したら、`run_all_tests.sh` を実行せず、`make test-e2e` は終了コード 0 以外で終わる

## 4. 機能要件

### 4.1 機能一覧

| ID | 機能名 | 説明 |
|----|--------|------|
| FR1 | `run_test` が未定義の関数を失敗 1 件として数える | 未定義の名前は呼び出さず、TESTS_RUN と TESTS_FAILED を 1 ずつ増やす |
| FR2 | run-list ガードが `TEST_FILES` 未登録のファイルを検出する | `tests/*.sh` のうち `TEST_FILES` に無いファイルを問題ありとする |
| FR3 | 全体実行で未登録ファイルを失敗として報告する | 未登録ファイル 1 件ごとに失敗 1 件を数える |
| FR4 | `--check-list` で未登録ファイルを報告する | 未登録ファイルがあれば終了コード 1 |
| FR5 | 再発を検出する bash の自己テスト | TS-1 から TS-5 を検証する |
| FR6 | `make test-e2e` で自己テストを毎回走らせる | E2E イメージの既定コマンドで自己テストを先に実行する |

### 4.2 機能詳細

#### FR1: `run_test` が未定義の関数を失敗 1 件として数える

**説明**: `test/e2e/scripts/helpers.sh` の `run_test` は、渡された名前がシェル関数として定義されていないとき、その名前を呼び出さず、TESTS_RUN と TESTS_FAILED をそれぞれ 1 増やし、関数名を含む失敗行を出力する。

**入力**:
- テスト関数名: 文字列 - `run_test` に渡す名前

**出力**:
- 未定義のとき: TESTS_RUN を 1、TESTS_FAILED を 1 増やす。関数名を含む失敗行を出力する

**ビジネスルール**:
- 定義済みかどうかは関数定義の有無で判定する。終了コード（127 など）では判定しない
- 定義済みの関数に対する挙動（呼び出し、戻り値を集計に使わないこと、tmux セッションの後片付け）は変えない

#### FR2: run-list ガードが `TEST_FILES` 未登録のファイルを検出する

**説明**: `test/e2e/scripts/run_all_tests.sh` の `run_guard_checks` は、`tests/*.sh` のうち、ファイル名（basename）が `TEST_FILES` のどの値にも無いものを未登録ファイルとして集め、1 件以上あれば問題あり（戻り値 1）とする。

**ビジネスルール**:
- 既存の未定義エントリ（GUARD_UNDEFINED）と一覧漏れ（GUARD_MISSING）の判定は変えない
- 全体実行と `--check-list` は、引き続きこの同じガードを使う

#### FR3: 全体実行で未登録ファイルを失敗として報告する

**説明**: 引数なしの全体実行（`run_all`）は、未登録ファイル 1 件ごとに TESTS_RUN と TESTS_FAILED を 1 ずつ増やし、ファイル名を含む失敗行を出力する。

**ビジネスルール**:
- 未登録ファイルが 1 件でもあれば、`print_summary` の Failed は 1 以上になり、終了コードは 0 以外になる
- 未登録ファイルにあるテストが `FULL_RUN_LIST` に載っている場合、そのテストは FR1 によっても失敗 1 件として数えられる（A-3）

#### FR4: `--check-list` で未登録ファイルを報告する

**説明**: `--check-list` は、未登録ファイル 1 件ごとにファイル名を含む行を出力し、終了コード 1 で終わる。

**ビジネスルール**:
- OK 行（`OK: run list matches defined tests ...`）を出して終了コード 0 になるのは、未定義エントリ・一覧漏れ・未登録ファイルがすべて 0 件のときだけとする
- tmux と duofm は起動しない（現状どおり）

#### FR5: 再発を検出する bash の自己テスト

**説明**: bash の自己テストスクリプトを `test/e2e/scripts/` の下の `tests/` 以外の場所に新しく置く。

**ビジネスルール**:
- `run_all_tests.sh` と `helpers.sh` の複製を一時ディレクトリに置く
- tmux も duofm も起動しないテスト関数を持つ検証用の `tests/` ディレクトリと、検証用の `TEST_FILES` / `FULL_RUN_LIST` を用意して、TS-1 から TS-5 を検証する
- `run_all`・`check_list`・`run_guard_checks`・`run_test` の本番のロジックを実際に通す（A-2）
- シナリオごとに結果を出力し、1 つでも失敗すれば終了コード 0 以外で終わる

#### FR6: `make test-e2e` で自己テストを毎回走らせる

**説明**: E2E イメージ（`test/e2e/Dockerfile`）の既定コマンドは、自己テストを先に実行し、成功したときに `run_all_tests.sh` を実行する。

**ビジネスルール**:
- `make test-e2e`（`docker run --rm duofm-e2e-test`）を実行するたびに、コンテナ内で自己テストが走る
- 自己テストが失敗したら、`make test-e2e` は終了コード 0 以外で終わる
- 自己テストが失敗したときは `run_all_tests.sh` を実行しない（A-1）

## 5. 非機能要件

### 5.1 パフォーマンス要件

該当なし。

### 5.2 セキュリティ要件

- **NFR2 書き込み範囲**: 自己テストが書き込むのは、自分で作った一時ディレクトリ（mktemp）の下だけとする。`/testdata` や `/e2e/scripts` は変更しない。コンテナ内では testuser として動く。

### 5.3 可用性要件

該当なし。

### 5.4 保守性要件

- **NFR3 依存を増やさない**: 自己テストが使うのは、E2E イメージに既にある bash と基本コマンドだけとする。新しいパッケージは追加しない。

### 5.5 互換性要件

- **NFR4 既存の CLI を保つ**: `run_all_tests.sh` のカテゴリ指定実行・`--list`・`--help` の挙動と、`tests/*.sh` を単体で実行したときの挙動（未定義の関数を数える FR1 は除く）は変えない。

### 5.6 実行環境要件

- **NFR1 自己テストは tmux と duofm を起動しない**: 自己テストとその検証用テスト関数は、tmux セッションも duofm も起動しない。`run_test` 内の後片付け（`tmux kill-session ... || true`）は起動にあたらない。

## 6. UI/UX要件

該当なし。

## 7. データ要件

該当なし。

## 8. 外部連携

該当なし。

## 9. 制約条件

### 9.1 技術的制約

- 自己テストは E2E イメージに既にある bash と基本コマンドだけを使う（NFR3）
- 自己テストスクリプトは `test/e2e/scripts/tests/` の外に置く（FR5）

### 9.2 ビジネス上の制約

該当なし。

### 9.3 スケジュール制約

該当なし。

### 9.4 宣言された変更集合

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

**デフォルトメンバー**（SPEC作成者が明示的に除外しない限り、常に宣言に含まれる）:
- `feature-docs/e2e-unregistered-test-file-guard/**`
- `test-docs/e2e-unregistered-test-file-guard/**`

`feature-docs/e2e-unregistered-test-file-guard/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/e2e-unregistered-test-file-guard/**` に含まれるもの: `{T}.tests.yaml`（パス形式: `test-docs/e2e-unregistered-test-file-guard/{T}.tests.yaml`）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

**意味論**:
- デフォルトのメンバーは、SPEC作成者が明示的に除外しない限り宣言に含まれる。除外は意図的な絞り込みであり、記載漏れによる省略ではない。
- この宣言はスーパーセット（superset）の主張であり、実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。実際には生成されないパスが宣言されていても違反にはならない。implementタスクを1つも生成しないフィーチャーは `test-docs/e2e-unregistered-test-file-guard/` ディレクトリを生成しないが、宣言された `test-docs/e2e-unregistered-test-file-guard/**` は依然として正しい。

## 10. 想定される課題とリスク

該当なし。

## 11. 成功基準

### 11.1 受け入れ基準

- [ ] AC-1: 再現手順（`tests/` にテストファイルを追加し、`TEST_FILES` に登録せず、関数名だけを `FULL_RUN_LIST` に加えて全体実行する）で、終了コードが 0 以外になり、Failed が 1 以上になり、出力に未登録ファイル名とそのテスト関数名が失敗として出る。（FR1, FR2, FR3）
- [ ] AC-2: 同じ状態で `--check-list` を実行すると、終了コード 1 で終わり、出力に未登録ファイル名が出る。（FR2, FR4）
- [ ] AC-3: `run_test` に未定義の名前を渡すと、TESTS_RUN と TESTS_FAILED がそれぞれちょうど 1 増え、TESTS_PASSED は変わらない。（FR1）
- [ ] AC-4: 現在のリポジトリ（`tests/*.sh` の 15 ファイルがすべて `TEST_FILES` に登録済み）では、`--check-list` が OK 行を出して終了コード 0 で終わり、全体実行でガード由来の失敗が新しく出ない。（FR2, FR3, FR4）
- [ ] AC-5: 自己テストは、修正前の `run_all_tests.sh` と `helpers.sh` に対して失敗し、修正後は成功する。（FR5）
- [ ] AC-6: `make test-e2e-build && make test-e2e` を実行するたびに自己テストが走り、自己テストが失敗したら `make test-e2e` の終了コードが 0 以外になる。（FR6）
- [ ] AC-7: 自己テストの実行中に、tmux セッションと duofm プロセスが起動されない。（NFR1）

### 11.2 KPI

該当なし。

## 12. テストシナリオ

### 12.1 テスト観点

- [ ] 異常系 TS-1: 未登録ファイルのテストが実行一覧にある全体実行
    - 準備: 検証用 `tests/` に、登録済みファイル（成功するテスト 1 件。TESTS_RUN と TESTS_PASSED を 1 ずつ増やす）と未登録ファイル（テスト 1 件）を置く。`FULL_RUN_LIST` には両方のテストを載せる。
    - 期待: 終了コードは 0 以外。Total 3 / Passed 1 / Failed 2（未登録ファイル 1 件 + 未定義として呼ばれたテスト 1 件）。出力に未登録ファイル名とそのテスト関数名が出る。
    - 対応: FR1, FR2, FR3, AC-1
- [ ] 異常系 TS-2: 未登録ファイルのテストが実行一覧にある `--check-list`
    - 準備: TS-1 と同じ検証用の構成。
    - 期待: 終了コードは 1。出力に未登録ファイル名が出て、OK 行は出ない。
    - 対応: FR2, FR4, AC-2
- [ ] 正常系 TS-3: すべて登録済みの構成（誤検出が無いこと）
    - 準備: 検証用 `tests/` のファイルがすべて `TEST_FILES` に登録され、定義されたテストがすべて `FULL_RUN_LIST` に載っている。
    - 期待: `--check-list` は OK 行を出して終了コード 0。全体実行は Failed 0 で終了コード 0。
    - 対応: FR2, FR3, FR4, AC-4
- [ ] 異常系 TS-4: 未登録ファイルのテストが実行一覧に無い場合
    - 準備: 未登録ファイルのテストを `FULL_RUN_LIST` に載せない。
    - 期待: `--check-list` は終了コード 1 で、出力に未登録ファイル名と、既存の一覧漏れ行（`Missing from run list`）の両方が出る。
    - 対応: FR2, FR4
- [ ] 異常系 TS-5: `run_test` に未定義の名前を渡す
    - 準備: `helpers.sh` を読み込み、定義されていない名前で `run_test` を呼ぶ。
    - 期待: TESTS_RUN と TESTS_FAILED がそれぞれ 1 増え、TESTS_PASSED は変わらない。出力にその名前が失敗として出る。
    - 対応: FR1, AC-3

## 13. 用語定義

| 用語 | 定義 |
|------|------|
| `TEST_FILES` | `run_all_tests.sh` が source するテストファイルの一覧 |
| `FULL_RUN_LIST` | 全体実行で実行するテスト関数名の一覧（実行一覧） |
| run-list ガード | `run_guard_checks`。実行一覧と定義されたテストの不一致を検出する |
| 未登録ファイル | `tests/*.sh` のうち、ファイル名（basename）が `TEST_FILES` のどの値にも無いもの |
| GUARD_UNDEFINED | 未定義エントリ。実行一覧にあって定義されていないテスト |
| GUARD_MISSING | 一覧漏れ。定義されていて実行一覧に無いテスト（`Missing from run list`） |
| 全体実行 | 引数なしの `run_all_tests.sh`（`run_all`） |
| 自己テスト | FR5 で新しく置く bash のテストスクリプト |

## 14. 確認事項

### 14.1 確認済み事項

- [x] 修正方針: `run_test` は関数が未定義なら失敗 1 件として TESTS_RUN と TESTS_FAILED に数える。さらに run-list ガードが、`tests/*.sh` のうち `TEST_FILES` に無いファイルを、全体実行と `--check-list` の両方で失敗としてファイル名つきで報告する。
- [x] 再発テストの形: bash の自己テストスクリプトを新しく作り、一時ディレクトリに複製したランナーで、未登録ファイルのテストを `FULL_RUN_LIST` に置いた全体実行と `--check-list` を検証する。tmux と duofm は起動しない。`make test-e2e` の実行経路（コンテナ内）で毎回走らせる。
- [x] デザインステップ: 実施しない。

### 14.2 未確認・保留事項

なし。

### 14.3 前提事項

- A-1: FR6 で自己テストが失敗したときは `run_all_tests.sh` を実行せず、終了コード 0 以外で終わる。
- A-2: 複製したランナーに検証用の `TEST_FILES` / `FULL_RUN_LIST` を与える方法（テキスト置換か、main を実行せずにランナーを読み込んで配列を差し替えるか）は実装で決める。要件として求めるのは、`run_all`・`check_list`・`run_guard_checks`・`run_test` の本番のロジックを自己テストが実際に通すことだけとする。
- A-3: 未登録ファイルはファイル 1 件ごとに失敗 1 件として数え、そのファイルにあって `FULL_RUN_LIST` に載っているテストは、それとは別に FR1 で 1 件ずつ数える。
- A-4: `TEST_FILES` の値が存在しないファイルを指している場合の挙動（`set -e` の下で source が失敗して異常終了する）は変えない。対象外とする。
- A-5: 定義されていてもアサーションを 1 件も実行しないテスト関数が集計に出ないことは、対象外とする。

## 15. 参考資料

- `feature-docs/e2e-unregistered-test-file-guard/SPEC.md`: 実装仕様
