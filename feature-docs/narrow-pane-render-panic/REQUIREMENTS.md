---
title: "narrow-pane-render-panic"
created_date: 2026-09-26
status: draft
---

# narrow-pane-render-panic - 要件定義書

## 1. 概要

### 1.1 背景
ペイン幅が 0 または 1 のとき、ペインヘッダーの区切り線描画で `strings.Repeat("─", p.width-2)` に負の回数が渡り、`strings: negative Repeat count` で panic する。ペイン幅は端末幅 / 2 のため、端末幅が 3 以下のときに発生する。`model_view.go` に最小サイズの確認は無い。

該当箇所:
- `internal/ui/pane_render.go` の `viewInternal`
- `internal/ui/pane_render.go` の `ViewWithBgOutput`
- `internal/ui/pane_render.go` の `ViewDimmedWithDiskSpace`

ミニバッファ表示中は、`Minibuffer.View`（`internal/ui/minibuffer.go`）の入力表示幅（width - プロンプト長 - 2）が 0 以下になると、入力文字列の切り出し `runes[startPos:endPos]` で範囲外を参照する。

### 1.2 目的
- ペイン幅 0〜4 のとき、ペイン全体の描画が panic しないようにする。対象は通常表示、bg 出力分割表示、dimmed 表示、ミニバッファ表示中の表示
- 同じ不具合の再発をテストで検出できるようにする

### 1.3 スコープ

**対象**:
- 区切り線を描く 3 か所（`viewInternal`、`ViewWithBgOutput`、`ViewDimmedWithDiskSpace`）
- ミニバッファ表示中のペイン描画（`ViewWithMinibuffer` / `Minibuffer.View`）
- 上記 2 点の再発を検出する `internal/ui` のユニットテスト

**対象外**:
- `model_view.go` への最小端末サイズの確認や「画面が小さい」旨の表示の追加
- ダイアログの重ね描画（`lipgloss.Place` / `overlayDialogOnPane` / `overlaySortDialogOnPane`）とダイアログ本体の描画
- E2E テストの追加

## 2. ビジネス要件

### 2.1 ビジネス目標
- ペイン幅 0〜4 のとき、ペイン全体の描画が panic しないようにする。対象は通常表示、bg 出力分割表示、dimmed 表示、ミニバッファ表示中の表示
- 同じ不具合の再発をテストで検出できるようにする

### 2.2 対象ユーザー
| ユーザータイプ | 説明 |
|----------------|------|
| duofm の利用者 | 端末幅 3 以下で duofm を使う |

### 2.3 期待される効果
- 端末幅 0〜3 でも duofm の描画が panic しない
- 同じ不具合の再発をテストで検出できる

## 3. ユースケース

### 3.1 ユースケース一覧
| ID | ユースケース名 | アクター | 優先度 |
|----|----------------|----------|--------|
| UC01 | 狭い端末幅で画面を描画する | duofm の利用者 | 高 |
| UC02 | 狭い端末幅でミニバッファを表示する | duofm の利用者 | 高 |

### 3.2 ユースケース詳細

#### UC01: 狭い端末幅で画面を描画する

**アクター**: duofm の利用者

**事前条件**:
- 端末幅が 0〜3

**基本フロー**:
1. duofm を起動する
2. ペインが描画される（通常表示・bg 出力分割表示・dimmed 表示のいずれか）

**代替フロー**:
- なし

**事後条件**:
- panic せずに描画される

#### UC02: 狭い端末幅でミニバッファを表示する

**アクター**: duofm の利用者

**事前条件**:
- 端末幅が 0〜3

**基本フロー**:
1. インクリメンタル検索モード、またはシェルコマンドモードを開始する
2. ミニバッファ表示中のペインが描画される

**代替フロー**:
- なし

**事後条件**:
- panic せずに描画される

## 4. 機能要件

### 4.1 機能一覧
| ID | 機能名 | 説明 | 優先度 |
|----|--------|------|--------|
| FR1 | 区切り線の繰り返し回数を 0 未満にしない | ヘッダーの区切り線を max(0, ペイン幅 - 2) 個の「─」で描画する | 高 |
| FR2 | ペイン幅 0〜4 で各表示が panic しない | 4 つの表示がペイン幅 0〜4 で panic せずに文字列を返す | 高 |
| FR3 | 端末幅 0〜3 でもモデル全体の描画が panic しない | 幅 0〜3 の WindowSizeMsg を受けた Model の View() が panic しない | 高 |
| FR4 | 狭い幅でもヘッダーと行の構成を変えない | ヘッダー行数とミニバッファの行数を保つ | 高 |
| FR5 | 再発検出テスト | internal/ui にユニットテストを追加する | 高 |
| FR6 | ミニバッファの入力表示幅が 0 以下でも panic しない | Minibuffer.View の切り出しで範囲外を参照しない | 高 |

### 4.2 機能詳細

#### FR1: 区切り線の繰り返し回数を 0 未満にしない

**説明**: `Pane.viewInternal`、`Pane.ViewWithBgOutput`、`Pane.ViewDimmedWithDiskSpace`（`internal/ui/pane_render.go`）は、ヘッダーの区切り線を max(0, ペイン幅 - 2) 個の「─」で描画する。ペイン幅が 0 または 1 のとき、区切り線は空文字列になる。

**ビジネスルール**:
- 区切り線の「─」の個数 = max(0, ペイン幅 - 2)

#### FR2: ペイン幅 0〜4 で各表示が panic しない

**説明**: ペイン幅が 0、1、2、3、4 のいずれでも、次の表示は panic せずに文字列を返す。

- 通常表示（`View` / `ViewWithDiskSpace`）
- bg 出力分割表示（`ViewWithBgOutput`、focused が true と false の両方）
- dimmed 表示（`ViewDimmedWithDiskSpace`）
- ミニバッファ表示中の通常表示（`ViewWithMinibuffer`）

区切り線以外に狭い幅で panic する箇所がこれらの描画経路で見つかった場合も、この要件の対象に含める。

#### FR3: 端末幅 0〜3 でもモデル全体の描画が panic しない

**説明**: 幅 0〜3 の `tea.WindowSizeMsg` を受け取った Model で、`Model.View()` を呼んでも panic しない。対象はダイアログ非表示の次の 3 状態。

- (a) ミニバッファ非表示
- (b) インクリメンタル検索モード開始後（`startSearch(SearchModeIncremental)`）
- (c) シェルコマンドモード開始後（`startShellCommandMode`）

#### FR4: 狭い幅でもヘッダーと行の構成を変えない

**説明**: ペイン幅が 0〜1 のときも、区切り線の行は 1 行として出力する。4 つの表示すべてでヘッダーは `paneHeaderRows`（3）行のまま変わらない。ミニバッファ表示中は、ミニバッファの行が 1 行になる。

#### FR5: 再発検出テスト

**説明**: `internal/ui` にユニットテストを追加し、次の 3 点を確認する。

1. ペイン幅 0〜4 のそれぞれで FR2 の各表示を描画して panic しないこと
2. `Minibuffer.View` を FR6 の条件の組み合わせで呼んで panic しないこと
3. 端末幅 0〜3 で FR3 の各状態の `Model.View()` が panic しないこと

#### FR6: ミニバッファの入力表示幅が 0 以下でも panic しない

**説明**: `Minibuffer.View`（`internal/ui/minibuffer.go`）は、入力表示幅（width - プロンプト長 - 2）が 0 以下のとき、入力文字列の切り出し（`runes[startPos:endPos]`）で範囲外を参照しない。幅 0〜4、入力が空・非空、カーソル位置 0〜入力長のどの組み合わせでも panic しない。出力は改行を含まない 1 行にする。

**ビジネスルール**:
- 入力表示幅 = width - プロンプト長 - 2
- 入力表示幅が 0 以下のときに表示する内容（入力文字やカーソルを出すかどうか）は規定しない

## 5. 非機能要件

### 5.1 パフォーマンス要件
- 該当なし

### 5.2 セキュリティ要件
- 該当なし

### 5.3 可用性要件
- 該当なし

### 5.4 保守性要件
- NFR2: 既存テストを通す。`go test ./...` が全件成功する。

### 5.5 互換性要件
- NFR1: 既存のペイン描画結果を変えない。ペイン幅が 2 以上のとき、通常表示、bg 出力分割表示、dimmed 表示の描画結果は変更前と同じにする。
- NFR3: 既存のミニバッファ描画結果を変えない。`Minibuffer.View` の入力表示幅が 1 以上のとき、描画結果は変更前と同じにする。

## 6. UI/UX要件

### 6.1 画面設計要件
- 新しい画面要素・レイアウト・見た目の変更はない
- ペイン幅 0〜1 では区切り線の行を空の内容で出力し、行そのものは残す（FR4）

### 6.2 画面遷移
- 該当なし

### 6.3 レスポンシブ対応
- 該当なし

## 7. データ要件

- 該当なし

## 8. 外部連携

- 該当なし

## 9. 制約条件

### 9.1 技術的制約
- ヘッダー行数（`paneHeaderRows` = 3）は、描画位置とマウスのヒットテストの対応を固定する既存テストで使われている

### 9.2 ビジネス上の制約
- なし

### 9.3 スケジュール制約
- なし

### 9.4 宣言された変更集合

このフィーチャー固有のパスは手動で列挙せず、create-plan で `workflow.yaml` の各タスクの `files` から導出する（`references/phases/create-plan-phase.md`）。

**デフォルトメンバー**（SPEC作成者が明示的に除外しない限り、常に宣言に含まれる）:
- `feature-docs/narrow-pane-render-panic/**`
- `test-docs/narrow-pane-render-panic/**`

`feature-docs/narrow-pane-render-panic/**` に含まれるもの: `REQUIREMENTS.md`、`SPEC.md`、`IMPLEMENTATION.md`、`workflow.yaml`、`phase-state/`、`tasks/`、`reviews/roundN.yaml`、`VERIFICATION.md`、`retrospect.yaml`、およびデザインステップが生成するデザイン成果物。生成主体は各フェーズドキュメントおよび `references/phase-state.md` を参照（引用のみ、ルールは再掲しない）。

`test-docs/narrow-pane-render-panic/**` に含まれるもの: `{T}.tests.yaml`（パス形式: `test-docs/narrow-pane-render-panic/{T}.tests.yaml`）。生成主体は `implement-phase.md` を参照（引用のみ、ルールは再掲しない）。

**意味論**:
- デフォルトのメンバーは、SPEC作成者が明示的に除外しない限り宣言に含まれる。除外は意図的な絞り込みであり、記載漏れによる省略ではない。
- この宣言はスーパーセット（superset）の主張であり、実際の変更集合は宣言に含まれる（CONTAINED IN）必要がある。実際には生成されないパスが宣言されていても違反にはならない。implementタスクを1つも生成しないフィーチャーは `test-docs/narrow-pane-render-panic/` ディレクトリを生成しないが、宣言された `test-docs/narrow-pane-render-panic/**` は依然として正しい。

## 10. 想定される課題とリスク

### 10.1 技術的課題
| 課題 | 影響度 | 対応策 |
|------|--------|--------|
| ペイン幅 3〜4 では区切り線の回数は正だが、lipgloss の `Width(p.width-2)` と `Padding(0,1)` で内容幅が 0 以下になる | 中 | FR2 の対象として、ペイン幅 3〜4 もテストで確認する |
| ミニバッファの入力が空で入力表示幅が負（例: 幅 4、プロンプト `"/: "` で availableWidth = -1）の場合、修正前のコードでは `len(runes)=0 > -1` が成り立って `runes[2:0]` に達するので、入力が空でも panic すると推定（実行未確認） | 中 | FR6 の対象として、入力が空のケースもテストで確認する |

### 10.2 ビジネスリスク
- なし

## 11. 成功基準

### 11.1 受け入れ基準
- [ ] AC1: ペイン幅 0、1、2、3、4 のそれぞれで、`View()`、`ViewWithBgOutput()`（focused true/false）、`ViewDimmedWithDiskSpace()` を呼ぶテストが panic せずに成功する。（FR1、FR2、FR5）
- [ ] AC2: AC1 のテストを修正前のコードで実行すると、ペイン幅 0 と 1 のケースで失敗する（panic を検出する）。（FR5）
- [ ] AC3: 端末幅 0、1、2、3 の WindowSizeMsg を与えた Model の `View()` を呼ぶテストが、ミニバッファ非表示、インクリメンタル検索モード開始後、シェルコマンドモード開始後のそれぞれで panic せずに成功する。（FR3、FR5）
- [ ] AC4: ペイン幅 0 と 1 で、4 つの表示（`ViewWithMinibuffer` はミニバッファ表示中）とも出力の行数が、同じ高さ・幅 40 で描画したときの行数と一致する。（FR4）
- [ ] AC5: `go test ./...` が全件成功する。ペイン幅 40 で描画位置とヒットテストの揃いを固定している既存テストと、既存の `TestMinibufferView` / `TestMinibufferViewTruncation` を含む。（NFR1、NFR2、NFR3）
- [ ] AC6: `Minibuffer.View` を次の全組み合わせで呼ぶテストが panic せずに成功し、出力に改行を含まない。幅: 0〜4。プロンプト: `"/: "` と `"!: "`。入力: `""`、`"a"`、`"abc"`。カーソル位置: 0、中間、入力長。（FR6、FR5）
- [ ] AC7: AC6 のテストを修正前のコードで実行すると失敗する（例: 幅 4、プロンプト `"/: "`、入力 `"a"`、カーソル 1 で `runes[3:1]` の panic）。（FR5、FR6）
- [ ] AC8: ペイン幅 0〜4 のそれぞれで、ミニバッファを表示した（Show 済み、SetWidth にペイン幅を指定）状態で `ViewWithMinibuffer()` を呼ぶテストが panic せずに成功する。入力は空と非空の両方で行う。（FR2、FR5）

### 11.2 KPI
- 該当なし

## 12. テストシナリオ

### 12.1 テスト観点
- [ ] 境界値: ペイン幅 0（端末幅 0〜1）で width-2 = -2
- [ ] 境界値: ペイン幅 1（端末幅 2〜3）で width-2 = -1
- [ ] 境界値: ペイン幅 2 で区切り線の回数がちょうど 0 になる
- [ ] 境界値: ペイン幅 3〜4 で lipgloss の内容幅が 0 以下になる
- [ ] 境界値: ミニバッファの入力表示幅が負（入力が空のケースを含む）
- [ ] 境界値: ミニバッファの入力表示幅がちょうど 0（入力が非空なら `runes[startPos:endPos]` が逆転する）
- [ ] 境界値: 長いプロンプト（`"(reverse-i-search)'': "`、22 文字）では、ペイン幅 24 以下で入力表示幅が 0 以下になる
- [ ] 異常系: bg 出力分割表示で focused が true と false の両方
- [ ] 異常系: コマンド文字列が空でない bg 出力分割表示（ヘッダー文字列の切り詰め経路）
- [ ] 異常系: ミニバッファのカーソルが先頭、中間、末尾
- [ ] 正常系: `go test ./...` が全件成功する（既存の描画結果が変わらない）

| ID | シナリオ | 種別 | 場所 | 対応する受け入れ基準 |
|----|----------|------|------|----------------------|
| TS1 | 狭いペイン幅での各表示の panic 検出 | unit | `internal/ui/pane_render_test.go` | AC1、AC2、AC8 |
| TS2 | 狭いペイン幅でも行数を保つ | unit | `internal/ui/pane_render_test.go` | AC4 |
| TS3 | モデル全体の描画（端末幅 0〜3） | unit | `internal/ui`（`model_basic_test.go` など、WindowSizeMsg を使っている既存テストと同じファイル） | AC3 |
| TS4 | テストスイート全体 | unit | `./...` | AC5 |
| TS5 | `Minibuffer.View` の狭い幅での panic 検出 | unit | `internal/ui/minibuffer_test.go` | AC6、AC7 |

## 13. 用語定義

| 用語 | 定義 |
|------|------|
| ペイン幅 | 1 つのペインの幅。端末幅 / 2 |
| 区切り線 | ペインヘッダーの「─」を並べた行 |
| bg 出力分割表示 | `ViewWithBgOutput` による表示 |
| dimmed 表示 | `ViewDimmedWithDiskSpace` による表示 |
| 入力表示幅 | `Minibuffer.View` で入力文字列に使える幅。width - プロンプト長 - 2 |

## 14. 確認事項

### 14.1 確認済み事項

- [x] ミニバッファ表示中のペイン描画（`ViewWithMinibuffer` / `Minibuffer.View`）を範囲に含めるか: 含める。ペイン幅 0〜4 で入力が空・非空、カーソル位置の組み合わせでも panic しないように直し、再発検出テストを追加する。ダイアログの重ね描画（`lipgloss.Place` / `overlayDialogOnPane`）は範囲外とする。

### 14.2 未確認・保留事項
- なし

### 14.3 前提

- A1: 区切り線を描く 3 か所で繰り返し回数を 0 未満にしないこと、`Minibuffer.View` で入力表示幅が 0 以下のときに切り出しを避けることで直す。`model_view.go` に最小端末サイズの確認や「画面が小さい」旨の表示は追加しない。
- A2: ペイン幅 0〜1 では区切り線の行を空の内容で出力し、行そのものは残す（FR4）。
- A3: ペイン幅 2 以上での 3 つのペイン表示の描画結果と、入力表示幅 1 以上での `Minibuffer.View` の描画結果は、変更前と同じにする（既存テストで固定済みの前提）。
- A4: ミニバッファ表示中のペイン描画（`ViewWithMinibuffer` / `Minibuffer.View`）は範囲に含める。ダイアログの重ね描画（`lipgloss.Place` / `overlayDialogOnPane` / `overlaySortDialogOnPane`）とダイアログ本体の描画は範囲外とする。
- A5: 再発検出は `internal/ui` のユニットテストで行い、E2E テストは追加しない。
- A6: `Minibuffer.View` の入力表示幅が 0 以下のときに表示する内容（入力文字やカーソルを出すかどうか）は規定しない。求めるのは panic しないことと、出力が 1 行であることだけにする。

## 15. 参考資料

- 該当なし
