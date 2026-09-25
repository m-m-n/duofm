---
title: "fix-e2e-failures"
created_date: 2026-09-25
status: draft
---

# fix-e2e-failures - 要件定義書

## 1. 概要

### 1.1 背景
`make test-e2e-build && make test-e2e`（run_all_tests.sh による全体実行）で失敗が出ている。
run_all_tests.sh の実行リストには未定義の関数名が含まれ、定義済みでも実行されていないテストがある。

### 1.2 目的
- `make test-e2e-build && make test-e2e`（run_all_tests.sh による全体実行）の結果が Failed: 0 になる。
- 同じ問題が再発したとき、E2E の実行で検出できる。対象は、実行されないテスト関数と、定義の無い実行リストの項目。

### 1.3 スコープ
- 対象: E2E スクリプト（`test/e2e/scripts/tests/*.sh` のうち本フィーチャーで修正するテスト、`test/e2e/scripts/run_all_tests.sh`）
- 対象外: Go のソースコード、`test/e2e/Dockerfile`、`test/e2e/scripts/run_tests.sh`

## 2. ビジネス要件

### 2.1 ビジネス目標
- `make test-e2e-build && make test-e2e` の全体実行で Failed: 0 になる。
- 実行されないテスト関数と、定義の無い実行リストの項目を E2E の実行で検出できる。

### 2.2 対象ユーザー
該当なし

### 2.3 期待される効果
- 2.1 と同じ

## 3. ユースケース

該当なし

## 4. 機能要件

### 4.1 機能一覧
| ID | 機能名 | 説明 |
|----|--------|------|
| FR1 | アーカイブのテストを現在のコンテキストメニューの動作に合わせる | 9 件のアーカイブテストを `@` と番号キーの操作に合わせる |
| FR2 | リネームのテストを拡張子を保持するダイアログに合わせる | 入力済みで拡張子を保持するリネームダイアログに合わせる |
| FR3 | 一括削除は `y` で確定する | 一括削除の確定を `y` で行う |
| FR4 | ソートダイアログは `q` を無視する | `q` でソートダイアログが閉じないことを検証する |
| FR5 | ヘルプのシェルコマンド項目の確認 | `!` の項目があるページまで移動して文言を確認する |
| FR6 | ブックマークのヒント文言 | `d:Delete` を確認する |
| FR7 | 履歴の進む方向がクリアされる経路 | 移動できるディレクトリだけを通る |
| FR8 | 実行リストと実行漏れの検出 | 実行リストを修正し、未定義関数と登録漏れを失敗として扱う |
| FR9 | 修正するテストの判定と後片付けの強化 | 常に成功する判定を厳密にし、作成したファイルを削除する |

### 4.2 機能詳細

#### FR1: アーカイブのテストを現在のコンテキストメニューの動作に合わせる

**説明**: 次の 9 件のアーカイブテストは、`@` でコンテキストメニューを開き、Compress または Extract archive を番号キーで選ぶ。余分な Enter は送らない。`/testdata/user_owned` の出力を確認するテストは、先に反対側のペインを user_owned に向ける。

- test_compress_format_dialog_opens
- test_compress_format_navigation
- test_compression_level_dialog
- test_archive_name_dialog
- test_archive_conflict_dialog
- test_compress_cancel_workflow
- test_compress_complete_workflow
- test_extract_complete_workflow
- test_multifile_compress

#### FR2: リネームのテストを拡張子を保持するダイアログに合わせる

**説明**: test_rename_file と test_navigation_after_rename は、入力済みで拡張子を保持するリネームダイアログの動作に合わせて操作する。リネーム後に after_rename.txt / navren_after.txt が存在することを確認する。

#### FR3: 一括削除は `y` で確定する

**説明**: test_batch_delete_marked_files は削除を `y` で確定し、マークしたファイルが無くなったことを確認する。

#### FR4: ソートダイアログは `q` を無視する

**説明**: test_sort_dialog_q_cancel と test_sort_dialog_q_cancel_with_dropdown は、`q` を押してもソートダイアログが閉じないことを検証する。1 つは通常の状態、もう 1 つはドロップダウンを開いた状態で検証する。アプリのコードは変更しない。

#### FR5: ヘルプのシェルコマンド項目の確認

**説明**: test_help_shows_shell_command は `!` の項目があるページまでスクロールまたはページ送りし、項目の文言（'execute shell command'）を確認する。

#### FR6: ブックマークのヒント文言

**説明**: test_bookmark_dialog_opens は 'd:Delete' を確認する。

#### FR7: 履歴の進む方向がクリアされる経路

**説明**: test_history_forward_cleared は、進む方向の履歴がクリアされることを確認する際、移動できるディレクトリだけを通る。

#### FR8: 実行リストと実行漏れの検出

**説明**: run_all_tests.sh は明示的な実行リストを維持する。

- 実行リストから未定義の 3 件（test_sort_dialog_hl_navigation、test_sort_dialog_jk_navigation、test_sort_dialog_confirm）を削除する。
- 定義済みで実行されていなかった 9 件のテストを実行リストに追加する。
- 未定義のテスト関数の呼び出しは失敗として数える。
- `test/e2e/scripts/tests/*.sh` に定義された `test_` 関数が実行リストに無い場合、実行を失敗にする。

#### FR9: 修正するテストの判定と後片付けの強化

**説明**: 本フィーチャーで修正するテストについて、常に成功する判定を厳密にする。修正する各テストは自分が作成したファイルを削除する。

## 5. 非機能要件

| ID | 要件名 | 内容 |
|----|--------|------|
| NFR1 | Go のソースを変更しない | 修正は E2E スクリプトだけに行う。Go のソースは変更しない。 |
| NFR2 | Dockerfile と run_tests.sh を変更しない | `test/e2e/Dockerfile` と `test/e2e/scripts/run_tests.sh`（run_all_tests.sh への追跡済みシンボリックリンク）は変更しない。実行スクリプトの修正は run_all_tests.sh だけに行う。 |
| NFR3 | 単体テストが通り続ける | `go test ./...` が通り続ける。 |
| NFR4 | テストの独立性 | 修正する E2E テストは、他のテストが残したファイルに依存しない。 |

## 6. UI/UX要件

該当なし（デザインステップはスキップ。変更は E2E スクリプトと実行スクリプトに限られ、UI の変更は無い）

## 7. データ要件

該当なし

## 8. 外部連携

該当なし

## 9. 制約条件

### 9.1 技術的制約
- Go のソースは変更しない（NFR1）
- `test/e2e/Dockerfile` と `test/e2e/scripts/run_tests.sh` は変更しない（NFR2）

### 9.2 ビジネス上の制約
該当なし

### 9.3 スケジュール制約
該当なし

### 9.4 宣言された変更集合

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

**デフォルトメンバー**（SPEC作成者が明示的に除外しない限り、常に宣言に含まれる）:
- `feature-docs/fix-e2e-failures/**`
- `test-docs/fix-e2e-failures/**`

`feature-docs/fix-e2e-failures/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/fix-e2e-failures/**` に含まれるもの: `{T}.tests.yaml`（パス形式: `test-docs/fix-e2e-failures/{T}.tests.yaml`）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

**意味論**:
- デフォルトのメンバーは、SPEC作成者が明示的に除外しない限り宣言に含まれる。除外は意図的な絞り込みであり、記載漏れによる省略ではない。
- この宣言はスーパーセット（superset）の主張であり、実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。実際には生成されないパスが宣言されていても違反にはならない。implementタスクを1つも生成しないフィーチャーは `test-docs/fix-e2e-failures/` ディレクトリを生成しないが、宣言された `test-docs/fix-e2e-failures/**` は依然として正しい。

## 10. 想定される課題とリスク

### 10.1 技術的課題（エッジケース）
- 圧縮形式ダイアログの番号は archive.GetAvailableFormats() から決まり、イメージに入っているツールに依存する。テストは 2 = tar.gz、5 = zip を前提とする。internal/archive との照合は行っていない。
- zip が使えない場合、test_multifile_compress は TESTS_RUN を増やすが成功にも失敗にも数えないため、Total/Passed/Failed の合計が合わない。
- メニュー項目が増えるとコンテキストメニューの番号がずれる。無効な項目（デスクトップの無いコンテナでの Open、Open with）も番号を持つ。
- これまで実行されていなかったテストは、実行リストに加えると失敗する可能性がある。test_sort_dialog_tab_navigation には判定が無い。
- リネームで残ったファイルは del1.txt より前に並び、test_batch_delete_marked_files の位置に依存する手順を壊す。
- ペインが同期されていない場合、圧縮の出力先が `/testdata`（書き込み不可）になり、アーカイブではなくエラーになる。

### 10.2 ビジネスリスク
該当なし

## 11. 成功基準

### 11.1 受け入れ基準
- [ ] AC1（FR1〜FR8）: 全体実行 `make test-e2e-build && make test-e2e` の結果が Failed: 0 になる。
- [ ] AC2（FR8）: テストの集計に、これまで実行されていなかった 9 件のテストが含まれる。
- [ ] AC3（FR8）: 実行リストに未定義のテスト関数名がある場合、実行結果に失敗が出る。
- [ ] AC4（FR8）: `tests/*.sh` に定義された `test_` 関数が実行リストに無い場合、実行が失敗する。
- [ ] AC5（FR4）: 2 件の q テストは、`q` の後もソートダイアログが開いている場合だけ成功する。通常の状態とドロップダウンを開いた状態の両方で確認する。
- [ ] AC6（FR1、FR2、FR9）: 全体実行の後、アーカイブとリネームのテストが作成したファイルが `/testdata/user_owned` に残っていない。
- [ ] AC7（NFR1、NFR2、NFR3）: 差分に Go のソースの変更、`test/e2e/Dockerfile` と `test/e2e/scripts/run_tests.sh` の変更が無く、`go test ./...` が通る。

### 11.2 KPI
該当なし

## 12. テストシナリオ

### 12.1 テスト観点
- [ ] TS1（FR1〜FR7）: E2E 全体を実行する。Archive、File Operation、Mark、Sort、Shell、Bookmark、History の各グループで、これまで失敗していたテストがすべて成功する。
- [ ] TS2（FR8）: 実行リストに未定義の名前を一時的に追加する。実行結果でそれが失敗として数えられる。
- [ ] TS3（FR8）: 定義済みのテストを実行リストから一時的に削除する。実行が失敗し、漏れたテスト名が表示される。
- [ ] TS4（FR4）: ソートダイアログを開いて `q` を押す: ダイアログは表示されたまま。ドロップダウンを開いて `q` を押す: ダイアログは表示されたまま。
- [ ] TS5（FR9、NFR4）: 全体実行の後、`/testdata/user_owned` にアーカイブとリネームの出力ファイルが残っていない。

## 13. 用語定義

| 用語 | 定義 |
|------|------|
| 実行リスト | run_all_tests.sh が明示的に列挙して呼び出すテスト関数の一覧 |

## 14. 確認事項

### 14.1 確認済み事項

- [x] A1 修正の方針（requirement.fix-direction）: E2E スクリプトを現在のアプリの動作に合わせて修正する。Go は変更しない（update_tests_only）。
- [x] A2 ソートダイアログの q（requirement.sort-q-cancel）: 2 件の q テストを、`q` でソートダイアログが閉じないことを検証する内容に書き換える（invert_q_tests）。doc/tasks/cancel-key-unification/SPEC.md の FR2 は SortDialog から q を意図的に削除している。
- [x] A3 実行漏れの検出（requirement.runner-coverage）: 明示的な実行リストを維持して修正する。未定義関数の呼び出しは失敗として数える。定義済みで実行リストに無いテストは実行を失敗にする（fix_list_and_guard）。
- [x] A4 起動スクリプト（requirement.entrypoint）: `test/e2e/Dockerfile` と `test/e2e/scripts/run_tests.sh` は変更しない（leave_to_mouse_support）。run_tests.sh は run_all_tests.sh への追跡済みシンボリックリンク（mode 120000）であり、現在の CMD はすでに run_all_tests.sh を実行している。
- [x] A5 判定強化の範囲（requirement.assertion-scope）: 判定の強化とファイルの後片付けは、本フィーチャーで修正するテストだけを対象にする（touched_tests_only）。

A1〜A5 はバッチ実行での回答を前提として記録したもの。

### 14.2 未確認・保留事項
なし

## 15. 参考資料

- doc/tasks/cancel-key-unification/SPEC.md
