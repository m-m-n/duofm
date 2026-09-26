---
title: "e2e-rename-initial-value-check"
created_date: 2026-09-27
status: draft
---

# e2e-rename-initial-value-check - 要件定義書

## 1. 概要

### 1.1 背景
E2E のリネームテスト（`test_rename_file` / `test_navigation_after_rename`）は、リネームダイアログの入力欄の初期値を確認している。現状の確認は、ダイアログ背後のファイル一覧や検索フィルタ表示に一致するだけでも成功しうる。また `history_tests.sh` には常に成功する `testdata` の確認がある。

### 1.2 目的
- E2E のリネームテストが、リネームダイアログの入力欄の初期値を実際に検証するようにする。入力欄が期待するベース名をそのまま保持していないときは、確認が失敗するようにする。
- `history_tests.sh` から常に成功する `testdata` の確認を取り除く。

### 1.3 スコープ
- `test/e2e/scripts/tests/file_operation_tests.sh` の `test_rename_file` と `test_navigation_after_rename` の初期値確認
- `test/e2e/scripts/tests/history_tests.sh` の `test_history_forward_cleared` の `testdata` 確認

Go のソースファイル（`*.go`）と `test/e2e/scripts/helpers.sh` の `assert_contains` / `assert_not_contains` の既存の動作はスコープ外とする（NFR1、NFR2）。

## 2. ビジネス要件

### 2.1 ビジネス目標
- E2E のリネームテストが、リネームダイアログの入力欄の初期値を実際に検証するようにし、入力欄が期待するベース名をそのまま保持していないときは確認が失敗するようにする。
- `history_tests.sh` から常に成功する `testdata` の確認を取り除く。

### 2.2 対象ユーザー
該当なし

### 2.3 期待される効果
該当なし

## 3. ユースケース

### 3.1 ユースケース一覧
該当なし

### 3.2 ユースケース詳細
該当なし

## 4. 機能要件

### 4.1 機能一覧
| ID | 機能名 | 説明 | 優先度 |
|----|--------|------|--------|
| FR1 | `test_rename_file` の初期値確認を入力欄に限定し、厳密に一致させる | 入力欄の内容がベース名 `before_rename` と完全に一致することを確認する | - |
| FR2 | `test_navigation_after_rename` の初期値確認を入力欄に限定し、厳密に一致させる | 入力欄の内容がベース名 `navren_before` と完全に一致することを確認する | - |
| FR3 | `history_tests.sh` から常に成功する `testdata` の確認を取り除く | `test_history_forward_cleared` の `assert_contains "testdata"` を削除する | - |

### 4.2 機能詳細

#### FR1: `test_rename_file` の初期値確認を入力欄に限定し、厳密に一致させる

**説明**: `test/e2e/scripts/tests/file_operation_tests.sh` の `test_rename_file` における初期値確認（元は 195 行目の `assert_contains "before_rename"`）を、入力欄の行にしか一致しない形で行い、入力欄の内容がベース名 `before_rename` と完全に一致することを求める。入力欄の左右の枠線文字（`│`）の間には、ベース名以外に空白だけを許す。右の枠線の後ろには固定の拡張子 `.txt` が続く。

**入力**: 該当なし

**出力**: 該当なし

**ビジネスルール**:
- 入力欄の内容が空のとき、確認は失敗する。
- 入力欄の内容が拡張子付きの完全な名前（`before_rename.txt`）のとき、確認は失敗する。
- ベース名の前後に余分な文字があるとき、確認は失敗する。空白以外の 1 文字だけの場合も含む（例: `before_rename.`、`xbefore_rename`、`before_renamex`）。
- ダイアログ背後のファイル一覧に表示される `before_rename.txt` や、検索フィルタ表示の `/before_ren` だけで確認が成功してはならない。

**バリデーション**: 該当なし

**エラーケース**: 該当なし

#### FR2: `test_navigation_after_rename` の初期値確認を入力欄に限定し、厳密に一致させる

**説明**: 同じファイルの `test_navigation_after_rename` における初期値確認（元は 439 行目の `assert_contains "navren_before"`）を FR1 と同じ方法で行い、入力欄の内容がベース名 `navren_before` と完全に一致することを求める。

**入力**: 該当なし

**出力**: 該当なし

**ビジネスルール**:
- 入力欄の内容が空のとき、確認は失敗する。
- 入力欄の内容が完全な名前（`navren_before.txt`）のとき、確認は失敗する。
- ベース名の前後に余分な文字があるとき、確認は失敗する。空白以外の 1 文字だけの場合も含む（例: `navren_before.`、`xnavren_before`、`navren_beforex`）。
- ファイル一覧の `navren_before.txt` や、検索フィルタ表示の `/navren_before` だけで確認が成功してはならない。

**バリデーション**: 該当なし

**エラーケース**: 該当なし

#### FR3: `history_tests.sh` から常に成功する `testdata` の確認を取り除く

**説明**: `test/e2e/scripts/tests/history_tests.sh` の `test_history_forward_cleared` にある 167〜168 行目の `assert_contains "testdata"` を削除する。直後の `assert_not_contains "/testdata/dir1"` は残す。

**入力**: 該当なし

**出力**: 該当なし

**ビジネスルール**:
- `assert_not_contains "/testdata/dir1"` は残す。

**バリデーション**: 該当なし

**エラーケース**: 該当なし

## 5. 非機能要件

| ID | 要件 |
|----|------|
| NFR1 | Go のソースファイル（`*.go`）を変更しない。 |
| NFR2 | `test/e2e/scripts/helpers.sh` の `assert_contains` / `assert_not_contains` の既存の動作を変更しない（すべての E2E テストが使っている）。 |

### 5.1 パフォーマンス要件
該当なし

### 5.2 セキュリティ要件
該当なし

### 5.3 可用性要件
該当なし

### 5.4 保守性要件
該当なし

### 5.5 互換性要件
該当なし

## 6. UI/UX要件

### 6.1 画面設計要件
該当なし（変更は E2E テストのシェルスクリプトだけで、ユーザーインターフェースや見た目の変更はない）

### 6.2 画面遷移
該当なし

### 6.3 レスポンシブ対応
該当なし

## 7. データ要件

### 7.1 データモデル概要
該当なし

### 7.2 データ項目
該当なし

### 7.3 データ保持期間
該当なし

## 8. 外部連携

### 8.1 連携システム
該当なし

### 8.2 API仕様要件
該当なし

## 9. 制約条件

### 9.1 技術的制約
- Go のソースファイル（`*.go`）を変更しない（NFR1）。
- `test/e2e/scripts/helpers.sh` の `assert_contains` / `assert_not_contains` の既存の動作を変更しない（NFR2）。

### 9.2 ビジネス上の制約
該当なし

### 9.3 スケジュール制約
該当なし

### 9.4 宣言された変更集合

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

**デフォルトメンバー**（SPEC作成者が明示的に除外しない限り、常に宣言に含まれる）:
- `feature-docs/e2e-rename-initial-value-check/**`
- `test-docs/e2e-rename-initial-value-check/**`

`feature-docs/e2e-rename-initial-value-check/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/e2e-rename-initial-value-check/**` に含まれるもの: `{T}.tests.yaml`（パス形式: `test-docs/e2e-rename-initial-value-check/{T}.tests.yaml`）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

**意味論**:
- デフォルトのメンバーは、SPEC作成者が明示的に除外しない限り宣言に含まれる。除外は意図的な絞り込みであり、記載漏れによる省略ではない。
- この宣言はスーパーセット（superset）の主張であり、実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。実際には生成されないパスが宣言されていても違反にはならない。implementタスクを1つも生成しないフィーチャーは `test-docs/e2e-rename-initial-value-check/` ディレクトリを生成しないが、宣言された `test-docs/e2e-rename-initial-value-check/**` は依然として正しい。

## 10. 想定される課題とリスク

### 10.1 技術的課題
該当なし

### 10.2 ビジネスリスク
該当なし

## 11. 成功基準

### 11.1 受け入れ基準
- [ ] AC-1（FR1、FR2）: `test_rename_file` と `test_navigation_after_rename` の初期値確認が、入力欄の行にしか一致しない形で行われている。
- [ ] AC-2（FR1、FR2）: 入力欄の初期値が期待するベース名と異なるとき（例: 確認の前に入力欄を空にした場合、拡張子付きの完全な名前を保持している場合）、両テストの初期値確認が失敗する。
- [ ] AC-3（FR3）: `history_tests.sh` の常に成功する `assert_contains "testdata"`（167 行目）がなくなり、`assert_not_contains "/testdata/dir1"` は残っている。
- [ ] AC-4（FR1、FR2、FR3）: `make test-e2e` で file-ops と history のテストがすべて成功する。
- [ ] AC-5（NFR1）: 変更に `*.go` ファイルが含まれない。
- [ ] AC-6（FR1、FR2）: 初期値の一致は厳密で、入力欄の中でベース名の前後に余分な文字（空白以外の 1 文字を含む）を許さない。`test_rename_file` の確認は入力欄が `before_rename.`、`xbefore_rename`、`before_renamex` を表示しているときに失敗し、`test_navigation_after_rename` の確認は入力欄が `navren_before.`、`xnavren_before`、`navren_beforex` を表示しているときに失敗する。

### 11.2 KPI
該当なし

## 12. テストシナリオ

### 12.1 テスト観点
- [ ] 正常系 TS-1（AC-1、AC-4）: 修正したリネームテストが成功する。`make test-e2e` を実行する（コンテナ内で `/e2e/scripts/run_all_tests.sh file-ops` を実行してもよい）。`test_rename_file` と `test_navigation_after_rename` の確認がすべて成功する。
- [ ] 異常系 TS-2（AC-2）: 新しい確認が入力欄の内容に依存している。ローカルの未コミットの変更で、(a) 初期値確認の直前に `C-u` を送って入力欄を空にする、(b) 別の変更として、入力欄に拡張子付きの完全な名前を保持させる（例: `End` `.` `t` `x` `t` を送る）。それぞれで file-ops のテストを実行する。ファイル一覧に `before_rename.txt` / `navren_before.txt` が表示されていても、どちらの変更でも両テストの初期値確認が失敗する。
- [ ] 正常系 TS-3（AC-3、AC-4）: history のテストが成功する。`make test-e2e` を実行する（コンテナ内で `/e2e/scripts/run_all_tests.sh history` を実行してもよい）。`test_history_forward_cleared` を含む history のテストがすべて成功し、167 行目の `testdata` 確認がない。
- [ ] 回帰 TS-4（AC-5）: Go のソースが変更されていない。ベースとの差分に `*.go` ファイルが含まれるかを確認し、`make test` を実行する。`*.go` の差分がなく、ユニットテストが成功する。
- [ ] 境界値 TS-5（AC-6）: 厳密一致が前後 1 文字の余分な文字を拒否する。ローカルの未コミットの変更で、各テストの初期値確認の直前に、初期値に 1 文字を加えるキーを送る。1 回の実行につき 1 通り: (a) `End` `.` → `before_rename.` / `navren_before.`、(b) `Home` `x` → `xbefore_rename` / `xnavren_before`、(c) `End` `x` → `before_renamex` / `navren_beforex`。それぞれで file-ops のテストを実行する。どの変更でも、`test_rename_file` と `test_navigation_after_rename` の両方で初期値確認が失敗する（変更した実行での後続の確認の結果は評価しない）。

エッジケース:
- [ ] 入力欄の初期値が空: 確認は失敗する。
- [ ] 入力欄の初期値が拡張子付きの完全な名前（`before_rename.txt` / `navren_before.txt`）: 確認は失敗する。
- [ ] 入力欄の値がベース名の前後に空白以外の余分な文字を 1 文字持つ（`before_rename.`、`xbefore_rename`、`before_renamex`、`navren_before` についても同様）: 確認は失敗する。
- [ ] ダイアログ背後のファイル一覧の行が見えていても隠れていても、確認の結果は入力欄の内容だけで決まる。
- [ ] ダイアログ表示中に検索フィルタ表示（`/before_ren`、`/navren_before`）が画面に残っていても、それに一致して確認が成功してはならない。
- [ ] 値の末尾のカーソルは反転表示の空白として描画される。テキストのキャプチャでは余白の空白と区別できず、これによって確認が失敗してはならない。

## 13. 用語定義

| 用語 | 定義 |
|------|------|
| ベース名 | 拡張子（`.txt`）を除いた名前（例: `before_rename`、`navren_before`） |
| 入力欄 | 拡張子保持リネームダイアログで枠線に囲まれて描画され、ベース名だけを保持する欄 |

## 14. 確認事項

### 14.1 確認済み事項
なし

### 14.2 未確認・保留事項
なし

### 14.3 前提事項
- A-1: 拡張子保持リネームダイアログは、入力欄を角丸の枠線（入力行では `│`）で囲んで描画し、拡張子（`.txt`）を入力欄の右側に分けて表示する（`internal/ui/extension_rename_dialog.go` の `renderInputFieldWithExtension`）。入力欄はベース名だけを保持し、左右に 1 桁の余白を持つ。
- A-2: `TextInput.RenderWithCursor`（`internal/ui/text_input.go`）は、カーソルを反転表示の属性だけで描画する。文字の上にあるカーソルはその文字自身を表示し、値の末尾にあるカーソルは反転表示の空白になる。カーソル記号の文字は挿入されない。`tmux capture-pane -p`（`-e` なし）は属性を落とすため、厳密一致ではベース名の隣のカーソル記号を考慮する必要がない。
- A-3: `history_tests.sh` の 167 行目は「パス表示行に絞る」ではなく「削除」を採用する。続く `dir2` の名前検索と `assert_contains "/testdata/dir2"` によって、表示が `/testdata` に戻っていたことが確認されるため。
- A-4: テスト関数の追加・削除・改名は行わないため、`run_all_tests.sh` / `run_tests.sh` の実行リストは変わらない。
- A-5: E2E コンテナでのリネームダイアログの画面キャプチャに、枠線文字 `│` が現れる。これは前回の verify の実行で、枠線を基準にした確認が `make test-e2e` で成功したことで確かめられている。
- A-6: 入力欄の中でベース名の前後にある空白だけの余分な文字は、厳密一致の対象外とする。テキストのキャプチャでは、欄の余白や値の末尾のカーソルのセルと区別できないため。厳密一致は空白以外の余分な文字をすべて拒否する。

## 15. 参考資料

- `test/e2e/scripts/tests/file_operation_tests.sh`
- `test/e2e/scripts/tests/history_tests.sh`
- `test/e2e/scripts/helpers.sh`
- `internal/ui/extension_rename_dialog.go`
- `internal/ui/text_input.go`
