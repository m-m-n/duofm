# Feature: narrow-pane-render-panic

## Overview

ペイン幅 0〜4 のとき、ペイン全体（通常表示、bg 出力分割表示、dimmed 表示、ミニバッファ表示中の表示）の描画が panic しないように直す。区切り線の繰り返し回数を 0 未満にしないこと、`Minibuffer.View` で入力表示幅が 0 以下のときに入力文字列の切り出しを避けることで対応し、再発を検出するユニットテストを追加する。要件は `feature-docs/narrow-pane-render-panic/REQUIREMENTS.md` を参照。

## Objectives

- ペイン幅 0〜4 のとき、ペイン全体の描画が panic しないようにする。対象は通常表示、bg 出力分割表示、dimmed 表示、ミニバッファ表示中の表示
- 同じ不具合の再発をテストで検出できるようにする

## User Stories

### US1: 狭いペイン幅でも描画が panic しない
duofm の利用者として、端末幅 0〜3 でもペイン全体を panic せずに描画したい。

**Acceptance Criteria:**
- [ ] AC1: ペイン幅 0、1、2、3、4 のそれぞれで、`View()`、`ViewWithBgOutput()`（focused true/false）、`ViewDimmedWithDiskSpace()` を呼ぶテストが panic せずに成功する。（FR1、FR2、FR5）
- [ ] AC3: 端末幅 0、1、2、3 の WindowSizeMsg を与えた Model の `View()` を呼ぶテストが、ミニバッファ非表示、インクリメンタル検索モード開始後、シェルコマンドモード開始後のそれぞれで panic せずに成功する。（FR3、FR5）
- [ ] AC4: ペイン幅 0 と 1 で、4 つの表示（`ViewWithMinibuffer` はミニバッファ表示中）とも出力の行数が、同じ高さ・幅 40 で描画したときの行数と一致する。（FR4）
- [ ] AC5: `go test ./...` が全件成功する。ペイン幅 40 で描画位置とヒットテストの揃いを固定している既存テストと、既存の `TestMinibufferView` / `TestMinibufferViewTruncation` を含む。（NFR1、NFR2、NFR3）
- [ ] AC6: `Minibuffer.View` を次の全組み合わせで呼ぶテストが panic せずに成功し、出力に改行を含まない。幅: 0〜4。プロンプト: `"/: "` と `"!: "`。入力: `""`、`"a"`、`"abc"`。カーソル位置: 0、中間、入力長。（FR6、FR5）
- [ ] AC8: ペイン幅 0〜4 のそれぞれで、ミニバッファを表示した（Show 済み、SetWidth にペイン幅を指定）状態で `ViewWithMinibuffer()` を呼ぶテストが panic せずに成功する。入力は空と非空の両方で行う。（FR2、FR5）

### US2: 再発をテストで検出する
同じ不具合の再発をテストで検出したい。

**Acceptance Criteria:**
- [ ] AC2: AC1 のテストを修正前のコードで実行すると、ペイン幅 0 と 1 のケースで失敗する（panic を検出する）。（FR5）
- [ ] AC7: AC6 のテストを修正前のコードで実行すると失敗する（例: 幅 4、プロンプト `"/: "`、入力 `"a"`、カーソル 1 で `runes[3:1]` の panic）。（FR5、FR6）

## Technical Requirements

### Functional Requirements
- **FR1:** 区切り線の繰り返し回数を 0 未満にしない - `Pane.viewInternal`、`Pane.ViewWithBgOutput`、`Pane.ViewDimmedWithDiskSpace`（`internal/ui/pane_render.go`）は、ヘッダーの区切り線を max(0, ペイン幅 - 2) 個の「─」で描画する。ペイン幅が 0 または 1 のとき、区切り線は空文字列になる。
- **FR2:** ペイン幅 0〜4 で各表示が panic しない - ペイン幅が 0、1、2、3、4 のいずれでも、次の表示は panic せずに文字列を返す: 通常表示（`View` / `ViewWithDiskSpace`）、bg 出力分割表示（`ViewWithBgOutput`、focused が true と false の両方）、dimmed 表示（`ViewDimmedWithDiskSpace`）、ミニバッファ表示中の通常表示（`ViewWithMinibuffer`）。区切り線以外に狭い幅で panic する箇所がこれらの描画経路で見つかった場合も、この要件の対象に含める。
- **FR3:** 端末幅 0〜3 でもモデル全体の描画が panic しない - 幅 0〜3 の `tea.WindowSizeMsg` を受け取った Model で、`Model.View()` を呼んでも panic しない。対象はダイアログ非表示の次の 3 状態: (a) ミニバッファ非表示、(b) インクリメンタル検索モード開始後（`startSearch(SearchModeIncremental)`）、(c) シェルコマンドモード開始後（`startShellCommandMode`）。
- **FR4:** 狭い幅でもヘッダーと行の構成を変えない - ペイン幅が 0〜1 のときも、区切り線の行は 1 行として出力する。4 つの表示すべてでヘッダーは `paneHeaderRows`（3）行のまま変わらない。ミニバッファ表示中は、ミニバッファの行が 1 行になる。
- **FR5:** 再発検出テスト - `internal/ui` にユニットテストを追加し、次の 3 点を確認する。(1) ペイン幅 0〜4 のそれぞれで FR2 の各表示を描画して panic しないこと。(2) `Minibuffer.View` を FR6 の条件の組み合わせで呼んで panic しないこと。(3) 端末幅 0〜3 で FR3 の各状態の `Model.View()` が panic しないこと。
- **FR6:** ミニバッファの入力表示幅が 0 以下でも panic しない - `Minibuffer.View`（`internal/ui/minibuffer.go`）は、入力表示幅（width - プロンプト長 - 2）が 0 以下のとき、入力文字列の切り出し（`runes[startPos:endPos]`）で範囲外を参照しない。幅 0〜4、入力が空・非空、カーソル位置 0〜入力長のどの組み合わせでも panic しない。出力は改行を含まない 1 行にする。

### Non-Functional Requirements
- **NFR1:** 既存のペイン描画結果を変えない - ペイン幅が 2 以上のとき、通常表示、bg 出力分割表示、dimmed 表示の描画結果は変更前と同じにする。
- **NFR2:** 既存テストを通す - `go test ./...` が全件成功する。
- **NFR3:** 既存のミニバッファ描画結果を変えない - `Minibuffer.View` の入力表示幅が 1 以上のとき、描画結果は変更前と同じにする。

## Implementation Approach

### Architecture

新しい構成要素は追加しない。既存の描画関数を修正する。

- 区切り線を描く 3 か所（`internal/ui/pane_render.go` の `viewInternal`、`ViewWithBgOutput`、`ViewDimmedWithDiskSpace`）で、繰り返し回数を 0 未満にしない（FR1）。
- `Minibuffer.View`（`internal/ui/minibuffer.go`）で、入力表示幅が 0 以下のときに入力文字列の切り出しを避ける（FR6）。
- `model_view.go` に最小端末サイズの確認や「画面が小さい」旨の表示は追加しない（A1）。

### Data Flow

該当なし

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- `internal/ui/pane_render.go`: 区切り線を描く 3 つの表示関数
- `internal/ui/minibuffer.go`: `Minibuffer.View`

**External Dependencies:**
- なし

### File Structure

```
internal/ui/
├── pane_render.go         # 区切り線の繰り返し回数（FR1）
├── pane_render_test.go    # TS1、TS2
├── minibuffer.go          # 入力表示幅が 0 以下のときの切り出し（FR6）
├── minibuffer_test.go     # TS5
└── model_basic_test.go    # TS3（WindowSizeMsg を使っている既存テストと同じファイル）
```

## Declared Change Set

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

すべての SPEC は、フィーチャー固有のパスに加えて、次の 2 つのワークフロー生成エントリをデフォルトで宣言する。

- `feature-docs/narrow-pane-render-panic/**`
- `test-docs/narrow-pane-render-panic/**`

`feature-docs/narrow-pane-render-panic/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/narrow-pane-render-panic/**` に含まれるもの: `test-docs/narrow-pane-render-panic/{T}.tests.yaml`（タスクごとのテスト記録）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

この 2 つのデフォルトエントリは、SPEC 作成者が明示的に除外しない限り宣言に含まれる。記載が無いことを除外とはみなさない。除外は意図的な絞り込みとして明示する。

この宣言はスーパーセット（superset）の主張であり、検証時に観測される実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。一致は求めない。implement タスクを 1 つも生成しないフィーチャーは `test-docs/narrow-pane-render-panic/` ディレクトリを生成しないが、宣言された `test-docs/narrow-pane-render-panic/**` は依然として正しい。宣言されたパスが生成されなくても違反にはならない。

## Test Scenarios

### Unit Tests
- [ ] TS1: 狭いペイン幅での各表示の panic 検出（`internal/ui/pane_render_test.go`、AC1・AC2・AC8） - テーブル駆動テスト。ペイン幅 0〜4 × {`View`, `ViewWithBgOutput`(focused=false), `ViewWithBgOutput`(focused=true), `ViewDimmedWithDiskSpace`, `ViewWithMinibuffer`（入力が空 / 非空）} の組み合わせを描画する。各呼び出しを recover で包み、panic 時に `t.Fatalf` を出す（既存の `TestRenderHeaderLine1_NarrowWidths_NoPanicZeroWidth` と同じ書き方）。ペインは `newFilesPane`（`internal/ui/mouse_hittest_test.go`）で作る。bg 分割表示では `SetBgOutputActive(true)` と、行を入れた `NewOutputBuffer` を使う。ミニバッファは `NewMinibuffer` に `SetPrompt("/: ")`、`SetWidth(ペイン幅)`、`Show` を行って渡す。
- [ ] TS2: 狭いペイン幅でも行数を保つ（`internal/ui/pane_render_test.go`、AC4） - ペイン幅 0 と 1 について、4 つの表示の出力行数（末尾の改行を除く）が、同じ高さ・幅 40 での行数と一致することを確認する。
- [ ] TS3: モデル全体の描画（端末幅 0〜3）（`internal/ui` の `model_basic_test.go` など、WindowSizeMsg を使っている既存テストと同じファイル、AC3） - 端末幅 0〜3（高さは通常の値）の `tea.WindowSizeMsg` で Model を初期化する。そのまま、`startSearch(SearchModeIncremental)` の後、`startShellCommandMode()` の後の 3 状態で、`Model.View()` を recover で包んで呼び、panic しないことを確認する。
- [ ] TS4: テストスイート全体（`./...`、AC5） - `go test ./...` を実行する。
- [ ] TS5: `Minibuffer.View` の狭い幅での panic 検出（`internal/ui/minibuffer_test.go`、AC6・AC7） - テーブル駆動テスト。幅 0〜4 × プロンプト {`"/: "`, `"!: "`} × 入力 {`""`, `"a"`, `"abc"`} × カーソル {0, 中間, 入力長} で Show 済みの `Minibuffer.View` を呼ぶ。recover で panic を検出し、出力に `"\n"` が含まれないことを確認する。追加のケースとして、プロンプト `"(reverse-i-search)'': "`、幅 24、入力 `"a"`、カーソル 1 も含める（`runes[2:1]` の経路）。

### Integration Tests
- なし

### E2E Tests
**Existing E2E tests**: None
**Run command**: Not detected
- E2E テストは追加しない（A5）

### Edge Cases
- [ ] ペイン幅 0（端末幅 0〜1）: width-2 = -2
- [ ] ペイン幅 1（端末幅 2〜3）: width-2 = -1
- [ ] ペイン幅 2: 区切り線の回数がちょうど 0 になる境界
- [ ] ペイン幅 3〜4: 区切り線の回数は正だが、lipgloss の `Width(p.width-2)` と `Padding(0,1)` で内容幅が 0 以下になる
- [ ] bg 出力分割表示で、focused が true と false の両方
- [ ] コマンド文字列が空でない bg 出力分割表示（ヘッダー文字列の切り詰め経路）
- [ ] ミニバッファの入力が空で、入力表示幅が負（例: 幅 4、プロンプト `"/: "` で availableWidth = -1）。修正前のコードを読むと、`len(runes)=0 > -1` が成り立って `runes[2:0]` に達するので、入力が空でも panic すると推定（実行未確認）
- [ ] ミニバッファの入力表示幅がちょうど 0（入力が非空なら `runes[startPos:endPos]` が逆転する）
- [ ] ミニバッファのカーソルが先頭、中間、末尾
- [ ] 長いプロンプト（`"(reverse-i-search)'': "`、22 文字）では、ペイン幅 24 以下で入力表示幅が 0 以下になる

### Performance Tests
- なし

## Security Considerations

該当なし

## Error Handling

該当なし（panic を起こさないことが要件であり、新しいエラーの種類は追加しない）

## Performance Optimization

該当なし

## Success Criteria

- [ ] すべての機能要件（FR1〜FR6）が実装され、テストされている
- [ ] すべてのテストシナリオ（TS1〜TS5）が成功する
- [ ] 受け入れ基準 AC1〜AC8 を満たす
- [ ] コードレビューが完了している

## Assumptions

- A1: 区切り線を描く 3 か所で繰り返し回数を 0 未満にしないこと、`Minibuffer.View` で入力表示幅が 0 以下のときに切り出しを避けることで直す。`model_view.go` に最小端末サイズの確認や「画面が小さい」旨の表示は追加しない。
- A2: ペイン幅 0〜1 では区切り線の行を空の内容で出力し、行そのものは残す（FR4）。
- A3: ペイン幅 2 以上での 3 つのペイン表示の描画結果と、入力表示幅 1 以上での `Minibuffer.View` の描画結果は、変更前と同じにする（既存テストで固定済みの前提）。
- A4: ミニバッファ表示中のペイン描画（`ViewWithMinibuffer` / `Minibuffer.View`）は範囲に含める。ダイアログの重ね描画（`lipgloss.Place` / `overlayDialogOnPane` / `overlaySortDialogOnPane`）とダイアログ本体の描画は範囲外とする。
- A5: 再発検出は `internal/ui` のユニットテストで行い、E2E テストは追加しない。
- A6: `Minibuffer.View` の入力表示幅が 0 以下のときに表示する内容（入力文字やカーソルを出すかどうか）は規定しない。求めるのは panic しないことと、出力が 1 行であることだけにする。

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

## References

- 要件定義書: `feature-docs/narrow-pane-render-panic/REQUIREMENTS.md`
