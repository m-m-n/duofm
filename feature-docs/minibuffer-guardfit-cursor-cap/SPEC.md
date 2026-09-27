# Feature: minibuffer-guardfit-cursor-cap

## Overview

ミニバッファの guardFit が再計測上限に達したとき、カーソル単独なら内容幅に収まる場合は、表示範囲をカーソル単独に縮めてカーソル位置の書記素（末尾ならカーソルブロック）を表示する。カーソル単独でも収まらない場合は、入力もカーソルも表示しない。

## Objectives

- ミニバッファの guardFit が再計測上限に達しても、カーソル単独なら内容幅に収まるときは、カーソル位置の書記素（末尾ならカーソルブロック）を表示する
- レビュー指摘 9925cb2f380c82a9（minibuffer-shrink-loop-linear の review round1、medium、未解決）を解消し、SPEC FR3・IMPLEMENTATION.md CD4 からの逸脱をなくす

## User Experience

- 幅の見積もりが実測とずれる入力でも、カーソル単独で収まる限りカーソル位置（末尾ならカーソルブロック）が見え続ける

## Technical Requirements

### Functional Requirements

- **FR1: 再計測上限に達したときはカーソル単独に縮める** — 入力の表示範囲を決める 3 経路（カーソルが末尾の経路、カーソルが末尾以外で右側を削ってから左側を削る経路、カーソルの書記素を右端に寄せて左側をスクロールする経路）で、guardFit が再計測上限（最初の計測の後 3 回）に達しても行が内容幅を超えるとき、表示範囲をカーソル単独（カーソル位置の書記素、末尾ならカーソルブロック）に縮める。プロンプトとカーソル単独で組み立てた行を lipgloss.Width で計測し、内容幅以下ならその行を表示する。カーソル位置の書記素は 1 つの反転区間として表示し、末尾ならカーソルブロックを表示する
- **FR2: カーソル単独でも収まらないときは入力を表示しない** — FR1 のカーソル単独の行も内容幅を超えるときは、入力もカーソルも表示しない（現行の挙動）。shrink がそれ以上削れなくなって guardFit が失敗したときも同じ扱いにする
- **FR3: 既存の事後条件の維持** — m.width > 4 のとき、View の結果は改行を含まず、表示幅は m.width-2 以下。プロンプトと入力が内容幅に収まるときは切り詰めない。プロンプトが内容幅以上のときの切り詰めの経路（最後は空のプロンプト）は変えない。カーソル位置の書記素が残り幅より広いときは入力を表示しない。m.width が 0〜4 のときはパニックせず、改行を含まない
- **FR4: 再発を検出するテスト** — internal/ui/minibuffer_test.go に、View 経由の回帰テストを 3 経路それぞれに 1 つずつ置く。共通の条件は SetPrompt("؀")（U+0600）、SetWidth(10)、Show()（内容幅 6、プロンプト幅 1、入力に使える幅 5）。書記素は g1 = "😀︎"（U+1F600 U+FE0E）、g2 = "❤‍" + g1、g3 = "❤‍" + g2、g4 = "❤‍" + g3（"❤‍" は U+2764 U+200D）。
    1. カーソルが末尾の経路：入力 g1+g2+g3+g4、カーソル 20
    2. 右側を削ってから左側を削る経路：入力 g2+g3+g4+"xyz"、カーソル 18（"x"）
    3. 左側をスクロールする経路：入力 "z"+g1+g2+g3+g4+"x"、カーソル 21（"x"）

    いずれも fix256ColorProfile の下で View を実行し、reversedRuns による反転区間がちょうど 1 つ（末尾では空白 1 文字、それ以外では "x"）であること、プロンプトが残ること、g1〜g4 の文字が残らないこと、assertBoundedSingleLine を満たすことを確かめる。末尾の経路では反転区間を必ず検証する。これに加えて、カーソル単独でも収まらない状態で、入力もカーソルも表示されないことを確かめる View 経由のテストを置く（FR2）。本体をテストのためだけに分割しない

### Non-Functional Requirements

- **NFR1: 変更範囲** — 変更は internal/ui/minibuffer.go の View とその補助関数（guardFit を含む）、internal/ui/minibuffer_test.go に限る。HandleKey などの編集操作と、cursorPos が rune 単位であることは変えない。本体にテスト専用の差し替え口（幅の計測関数の注入など）は作らない
- **NFR2: 線形時間の維持** — カーソル単独に縮めるときの追加の計測は 1 回とする。既存の線形時間のテスト（TestMinibufferView_LinearTime_TS1〜TS4）は引き続き 10 秒の上限内で通る
- **NFR3: 既存テストの維持** — internal/ui/minibuffer_test.go と internal/ui/pane_render_test.go の既存テストを変更せずに通す。go test ./... がすべて通る

## Acceptance Criteria

- [ ] **AC1**（FR1, FR3, FR4）: カーソルが末尾の経路（プロンプト "؀"、m.width 10、入力 g1+g2+g3+g4、カーソル 20）で、View の結果の反転区間は空白 1 文字の 1 区間だけで、プロンプトが残り、g1〜g4 の文字は残らず、assertBoundedSingleLine を満たす
- [ ] **AC2**（FR1, FR3, FR4）: 右側を削ってから左側を削る経路（プロンプト "؀"、m.width 10、入力 g2+g3+g4+"xyz"、カーソル 18）で、View の結果の反転区間は "x" の 1 区間だけで、プロンプトが残り、g1〜g4 の文字は残らず、assertBoundedSingleLine を満たす
- [ ] **AC3**（FR1, FR3, FR4）: 左側をスクロールする経路（プロンプト "؀"、m.width 10、入力 "z"+g1+g2+g3+g4+"x"、カーソル 21）で、View の結果の反転区間は "x" の 1 区間だけで、プロンプトが残り、g1〜g4 の文字は残らず、assertBoundedSingleLine を満たす
- [ ] **AC4**（FR2, FR3, FR4）: プロンプトとカーソル単独の行も内容幅を超える状態で、入力もカーソルも表示されず（反転区間なし）、assertBoundedSingleLine を満たす
- [ ] **AC5**（FR3, NFR2, NFR3）: internal/ui/minibuffer_test.go と internal/ui/pane_render_test.go の既存テストが変更なしで通り、go test ./... がすべて通る
- [ ] **AC6**（NFR1）: 変更ファイルは internal/ui/minibuffer.go と internal/ui/minibuffer_test.go のみ

## Implementation Approach

### 対象ファイル

```
internal/ui/
├── minibuffer.go        # View とその補助関数（guardFit を含む）
└── minibuffer_test.go   # 回帰テスト（FR4）
```

### 処理の流れ

```
入力の表示範囲を決める 3 経路（末尾 / 右側→左側の削り / 左側スクロール）
  → guardFit（最初の計測 + 最大 3 回の再計測）
    → 内容幅以下                    → その行を表示
    → 再計測上限に達しても内容幅超過 → 表示範囲をカーソル単独に縮める（FR1）
        → プロンプト + カーソル単独の行を lipgloss.Width で 1 回計測（NFR2）
          → 内容幅以下 → その行を表示（カーソル位置の書記素を 1 つの反転区間、末尾ならカーソルブロック）
          → 内容幅超過 → 入力もカーソルも表示しない（FR2）
    → shrink がそれ以上削れず失敗   → 入力もカーソルも表示しない（FR2）
```

プロンプトが内容幅以上のときの切り詰めの経路（最後は空のプロンプト）は変えない（FR3、a2）。

### Design

デザインステップは省略する（既存のミニバッファ描画の縮退処理の不具合修正で、画面要素や見た目の追加・変更がない）。

## Declared Change Set

このセクションは手書きの一覧ではなく、create-plan での導出を示す。機能固有のパスは、create-plan の時点で `workflow.yaml` の各タスクの `files` エントリから導出する（`references/phases/create-plan-phase.md`）。

すべての SPEC は既定で、機能固有のパスに加えて次の 2 つのワークフロー生成エントリを宣言する。

- `feature-docs/minibuffer-guardfit-cursor-cap/**`
- `test-docs/minibuffer-guardfit-cursor-cap/**`

`feature-docs/minibuffer-guardfit-cursor-cap/**` は `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成する成果物を含む。これらはフェーズ文書と `references/phase-state.md` が生成・所有し、このセクションはそれらを参照するだけで規則を再掲しない。

`test-docs/minibuffer-guardfit-cursor-cap/**` はタスクごとのテスト記録 `test-docs/minibuffer-guardfit-cursor-cap/{T}.tests.yaml` を含む。これは `implement-phase.md` が生成・所有し、このセクションはそれを参照するだけで規則を再掲しない。

この 2 つの既定エントリは、SPEC の作成者が明示的に取り除かない限り宣言に含まれる。記載がないことをもって除外とはみなさず、除外は明示的な絞り込みとして行う。

この宣言は上位集合の宣言である。検証時に観測される実際の変更集合は宣言した集合に含まれていればよく、一致する必要はない。実装タスクを生まない機能は `test-docs/minibuffer-guardfit-cursor-cap/` ディレクトリを生成しないが、その場合も宣言した `test-docs/minibuffer-guardfit-cursor-cap/**` は正しく、実体化しなかった宣言済みパスは違反ではない。

## Test Scenarios

### Unit Tests

- [ ] **TS1**（AC1、internal/ui/minibuffer_test.go）: カーソルが末尾の経路：SetPrompt("؀")、SetWidth(10)、入力 g1+g2+g3+g4、カーソル 20、Show()。fix256ColorProfile の下で View を実行し、reversedRuns が空白 1 文字の 1 区間だけであること、プロンプト U+0600 が残ること、g1〜g4 の文字（U+1F600、U+2764）が残らないこと、assertBoundedSingleLine を満たすことを確かめる
- [ ] **TS2**（AC2、internal/ui/minibuffer_test.go）: 右側を削ってから左側を削る経路：同じプロンプトと幅、入力 g2+g3+g4+"xyz"、カーソル 18（"x"）。reversedRuns が "x" の 1 区間だけであること、プロンプトが残ること、g1〜g4 の文字が残らないこと、assertBoundedSingleLine を満たすことを確かめる
- [ ] **TS3**（AC3、internal/ui/minibuffer_test.go）: 左側をスクロールする経路：同じプロンプトと幅、入力 "z"+g1+g2+g3+g4+"x"、カーソル 21（"x"）。TS2 と同じことを確かめる
- [ ] **TS4**（AC4、internal/ui/minibuffer_test.go）: カーソル単独でも収まらない状態：同じプロンプトと幅で、プロンプトとの境界の結合によって、プロンプト + カーソル単独の行の実際の幅が内容幅 6 を超える入力（例：入力の先頭の書記素 g5 = "❤‍" + g4 にカーソル 0 を置き、スタイルのシーケンスを出さない色プロファイルで実行）。入力もカーソルも表示されず（反転区間なし）、assertBoundedSingleLine を満たすことを確かめる。例の入力が前提を満たさなければ、同じ性質を持つ入力に置き換える
- [ ] **TS5**（AC5、internal/ui/minibuffer_test.go, internal/ui/pane_render_test.go）: 既存テスト（幅 0〜4、全角、ZWJ・VS16・書記素、線形時間 TS1〜TS4、TS7 の結合列のガード、スクロール境界）が変更なしで通る

### Integration Tests

なし

### E2E Tests

**Existing E2E tests**: test/e2e（Docker + tmux の bash スクリプト）
**Run command**: `make test-e2e-build && make test-e2e`

- [ ] 既存の E2E テストが退行なく通る

E2E テストは追加しない（a4）。

### Edge Cases

- [ ] カーソルが末尾：上限到達時、プロンプト + カーソルブロック（1 桁）だけなら収まる
- [ ] カーソルが書記素の途中の rune にある（書記素全体を 1 つの反転区間として表示）
- [ ] カーソル位置の書記素が 2 桁で、残り幅とちょうど同じ
- [ ] プロンプトの末尾が入力側の書記素と結合し、カーソル単独でも実際の幅が見積もりを超える
- [ ] shrink が上限到達前に削れなくなる（すでにカーソル単独で収まらない）
- [ ] m.width 0〜4、プロンプトの幅が内容幅以上（変更なし）

## Assumptions

- **a1**: 縮め方は IMPLEMENTATION.md CD4 の記述に合わせる：再計測上限に達したら表示範囲をカーソル単独に縮め、それも収まらなければ入力を表示しない。再計測上限（3 回）は変えない
- **a2**: プロンプトの切り詰めの経路（プロンプトが内容幅以上）の guardFit 呼び出しは対象外。上限到達時に空のプロンプトになる現行の挙動を維持する
- **a3**: 再発を検出するテストは View 経由で 3 経路それぞれに置き、具体的な入力（プロンプト U+0600 と、VS15・ZWJ で連結した絵文字の書記素 g1〜g4）で再計測上限に到達させ、反転区間でカーソルの表示を確かめる。入力が Go 実行で上限に達しないと判明した場合は、同じ性質（プロンプトとの境界で見積もりが実測より小さくなる）を持つ入力に置き換える。再計測の回数を固定する内部テストは任意
- **a4**: E2E テストは追加しない。Go の単体テストだけにする
- **a5**: 表示幅の基準は引き続き lipgloss.Width とし、go-runewidth は判定の基準に使わない
- **a6**: カーソル単独に縮めたときは、カーソル以外の入力は表示しなくてよい（収まる範囲まで広げ直すことはしない）
- **a7**: HandleKey などの編集操作と、cursorPos が rune 単位であることは変えない

## Success Criteria

- [ ] すべての機能要件が実装され、テストされている
- [ ] すべてのテストシナリオが通る
- [ ] AC1〜AC6 を満たす

## Open Questions

なし（`status: tbd` の要件はない）
