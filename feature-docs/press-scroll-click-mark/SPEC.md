# Feature: press-scroll-click-mark

## Overview

非アクティブペインの項目を左押下してそのペインがアクティブになり、アクティブ化でファイル一覧が縮む（bg 分割が現れる）場合、押下処理中に scrollOffset が変わり、同じ行で離しただけのクリックが範囲マークになる。押下前のスクロール位置を保つことで、動かさないクリックではマークが増えないようにする。要件の詳細は `feature-docs/press-scroll-click-mark/REQUIREMENTS.md` を参照。

## Objectives

- 項目の左押下でクリックしたペインのスクロール位置が変わる場合でも、マウスを動かさずに押下・解放したときにマークが増えない。

## User Stories

### US1: 動かさないクリックでマークが増えない
マウスで操作する利用者として、アクティブ化で一覧が縮む非アクティブペインの項目を動かさずにクリックしたとき、マークを増やさずにカーソルをその項目へ移動したい。

**Acceptance Criteria:**
- [ ] AC-1: 80x24 で、右ペインが非アクティブで項目 20 個、cursor 15、scrollOffset 0、右ペインでバックグラウンドコマンドが実行中の状態とする。右ペインの項目 index 7 の上で同じ行の左押下・解放を行った後、右ペインのマーク集合は変わらず、cursor は 7 である。
- [ ] AC-2: AC-1 の状態で、押下後の右ペインの scrollOffset は 0 で、項目 7 は押下した画面行にある。
- [ ] AC-3: バックグラウンドの状態が実行中ではなく bgClosing の場合も、AC-1 と AC-2 が成り立つ。
- [ ] AC-4: 押下前の cursor によりアクティブ化で一覧がスクロールする場合、押下した項目がずれた範囲より上にある場合と下にある場合のどちらでも、同じ行の押下・解放でマークが増えない。
- [ ] AC-5: 押下した項目が縮小後のファイル一覧より下にある場合、表示範囲調整でその項目が最下の表示行に置かれ、同じ行の押下・解放でマークが増えない。
- [ ] AC-6: 押下前から存在したマークは、同じ行の押下・解放で変わらない。
- [ ] AC-7: アクティブ化するペインの親エントリ `..` を押下し同じ行で解放すると、マークが増えず、cursor は `..` にある。

### US2: ドラッグでの範囲マーク
マウスで操作する利用者として、アクティブ化で一覧が縮む非アクティブペインで押下してからドラッグしたとき、押下した項目を起点に範囲をマークしたい。

**Acceptance Criteria:**
- [ ] AC-8: 押下後に別の行へ移動して解放すると anchor..target の範囲がマークされる。離れてから押下行へ戻るドラッグでは、押下行がアンカーに対応する。

### US3: ダブルクリック
マウスで操作する利用者として、アクティブ化で一覧が縮む非アクティブペインの項目をダブルクリックしたとき、その項目に対して Enter 動作を実行したい。

**Acceptance Criteria:**
- [ ] AC-9: AC-1 の状態で、同じ行のダブルクリックは押下した項目に対して既存の Enter 動作を実行する。

### US4: 回帰の検出

**Acceptance Criteria:**
- [ ] AC-10: 新しい回帰単体テストが `Model.Update` を通じて AC-1 の状態を再現し、修正前のコードでは失敗する。
- [ ] AC-11: 既存のマウスの単体テストと E2E のマウステストが通る。

## Technical Requirements

### Functional Requirements
- **FR1:** 同じ行での押下・解放でマークが増えない — 非アクティブペインの項目を左ボタンで押下してそのペインがアクティブになり、そのアクティブ化でファイル一覧が縮む場合（そのペインでバックグラウンドコマンドが実行中または終了処理中のため bg 分割が現れる場合）、別の行へ移動せずに同じ画面行で解放すると、そのペインのマーク集合は押下前の状態と等しい。カーソルは押下した項目へ移動する。
- **FR2:** 押下時のスクロール位置維持 — 押下後、押下した項目が [押下前の scrollOffset, 押下前の scrollOffset + アクティブ化後の表示行数) の範囲にある場合、クリックしたペインの scrollOffset は押下前の値と等しい。押下した項目は押下した画面行に留まる。それ以外の場合（押下した項目が縮小後のファイル一覧より下にある場合）は、既存の表示範囲調整を適用し、押下した項目をファイル一覧の最下の表示行に表示する。
- **FR3:** アンカーと一致するドラッグの対応付け — このような押下で開始した armed セッション中、押下行へ戻る移動は押下した項目（アンカー）に対応する。別の行へ移動してから解放すると、既存のドラッグ動作と同じく anchor..target の範囲をマークする。
- **FR4:** アクティブ化するペインでのダブルクリック — FR2 で押下前の scrollOffset が保たれる場合、ダブルクリック判定時間内に同じ行を 2 回目に押下すると同じ項目に当たり、既存の Enter 動作を実行する。押下した項目が縮小後のファイル一覧より下にあり、表示範囲調整で移動した場合は、2 回目の押下が同じ項目に当たることは求めない。
- **FR5:** 回帰テスト — アクティブ化で bg 分割が現れて scrollOffset がずれる非アクティブペインに対し、`Model.Update` で押下・解放を行う単体テストを追加する。テストはマークが増えないことを検証し、修正前のコードでは失敗する。

### Non-Functional Requirements
- **NFR1:** キーボードでのペイン切り替えの維持 — Tab によるキーボードでのペイン切り替えと、bg 分割のアクティブ化時のスクロール動作は変更しない。
- **NFR2:** 共通のスクロール処理の維持 — 共通のスクロール処理（`Pane.adjustScroll`、`Pane.SetBgOutputActive`、`Model.syncPaneBgOutputState`、`Model.switchToPane`）は、マウス以外の呼び出し元に対して現在の動作を保つ。
- **NFR3:** 既存テストの通過 — `internal/ui/model_update_mouse_test.go`、`mouse_hittest_test.go`、`pane_drag_test.go` の既存テストと、E2E の `test/e2e/scripts/tests/mouse_tests.sh`、`mark_tests.sh` がすべて通り続ける。
- **NFR4:** ビルド・静的検査 — `go test ./...` が通る。コードは gofmt で整形済みで、go vet で指摘がない。

## Implementation Approach

### Architecture

**現在の押下処理の経路:**
```
handleEntryPress (internal/ui/model_update_mouse.go)
  ├─ ヒットテストで index を取得し anchor として記録
  ├─ switchToPane (internal/ui/model.go:318-328)
  │    └─ syncPaneBgOutputState (internal/ui/model.go:368-388)
  │         └─ Pane.SetBgOutputActive (internal/ui/pane.go:600-603)
  │              └─ adjustScroll()  ← 以前の cursor で scrollOffset を調整
  ├─ SetCursor(index)
  └─ EnsureCursorVisible

移動・解放
  └─ updateDrag (internal/ui/model_update_mouse.go:160-179)
       └─ clampedEntryIndex (internal/ui/mouse_hittest.go:83-102)  ← 現在の scrollOffset で行を index に変換
```

**採用する方針（keep_scroll）:**
```
handleEntryPress 内で押下前の scrollOffset を保存し、ペイン切り替え後に復元する。
クリック対象には既存の表示範囲調整を適用する。
共通のスクロール処理（adjustScroll / SetBgOutputActive / syncPaneBgOutputState / switchToPane）は変更しない。
```

### Data Flow

```
マウス押下 → handleEntryPress → anchor 記録 → ペイン切り替え → scrollOffset 復元 → SetCursor → 表示範囲調整
マウス移動・解放 → updateDrag → clampedEntryIndex → target == anchor ならクリック / それ以外は anchor..target をマーク
```

### API Design

該当なし

### Database Schema

該当なし

### Dependencies

**Internal Dependencies:**
- `internal/ui/model_update_mouse.go`: `handleEntryPress`、`updateDrag`
- `internal/ui/mouse_hittest.go`: `clampedEntryIndex`
- `internal/ui/model.go`: `switchToPane`、`syncPaneBgOutputState`
- `internal/ui/pane.go`: `SetBgOutputActive`、`adjustScroll`、`getVisibleLines`、`bgSplitHeights`
- `internal/ui/mouse_doubleclick.go`: clickKey によるダブルクリック判定

**External Dependencies:**
- 該当なし

### File Structure

```
internal/ui/
├── model_update_mouse.go        # 押下・ドラッグ処理
└── model_update_mouse_test.go   # 回帰テスト
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/press-scroll-click-mark/**`
- `test-docs/press-scroll-click-mark/**`

`feature-docs/press-scroll-click-mark/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/press-scroll-click-mark/**` covers `test-docs/press-scroll-click-mark/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/press-scroll-click-mark/` directory at all; the declared
`test-docs/press-scroll-click-mark/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] TS-1（FR1、FR2、FR5 / AC-1、AC-2、AC-10）: 右ペインが非アクティブ、項目 20 個、cursor 15、scrollOffset 0、右ペインで bg が有効。`Update(rowFor(7,0) での押下)` の後に `Update(同じ行での解放)` - マークが変わらず、cursor 7、scrollOffset 0
- [ ] TS-2（FR1、FR2 / AC-3）: TS-1 を、実行中の runner の代わりに bgClosing=true で行う - TS-1 と同じ結果
- [ ] TS-3（FR1、FR2 / AC-4）: 押下前の cursor と押下 index により、押下した項目がずれた範囲より上にある場合と範囲内にある場合のテーブル駆動のバリエーション - 同じ行での解放後にマークがない
- [ ] TS-4（FR1、FR2 / AC-5）: 押下前の scrollOffset + 縮小後の表示行数以上の index を押下する - cursor が最下の表示行にあり、同じ行での解放後にマークがない
- [ ] TS-5（FR1 / AC-6）: 右ペインの項目を事前にマークしてから TS-1 を行う - マーク集合が押下前と等しい
- [ ] TS-6（FR1 / AC-7）: アクティブ化するペインの `..` で押下・解放する - マークがなく、cursor が `..` にある
- [ ] TS-7（FR3 / AC-8）: アクティブ化するペインで押下し、別の行へ移動して解放する - anchor..target がマークされる。2 つ目のケース: 解放前に離れてから押下行へ戻る - 押下行がアンカーに対応する
- [ ] TS-8（FR4 / AC-9）: アクティブ化する bg ペインの同じ行でダブルクリック（偽の時計を使う） - 押下した項目に対して Enter 動作が実行される

### Integration Tests
- 該当なし

### E2E Tests
**Existing E2E tests**: `test/e2e/scripts/tests/mouse_tests.sh`、`test/e2e/scripts/tests/mark_tests.sh`
**Run command**: `make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] TS-9（NFR3、NFR4 / AC-11）: `go test ./...` と `make test-e2e` を実行する - すべて通る

### Edge Cases
- [ ] バックグラウンドの状態が bgClosing の場合（TS-2）
- [ ] 押下した項目が縮小後のファイル一覧より下にある場合（TS-4）
- [ ] 押下前から存在するマークがある場合（TS-5）
- [ ] 親エントリ `..` を押下する場合（TS-6）
- [ ] 離れてから押下行へ戻るドラッグ（TS-7）

### Performance Tests
- 該当なし

## Security Considerations

該当なし

## Error Handling

該当なし

## Performance Optimization

該当なし

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Code review is completed

## Assumptions

- as-1: Tab によるキーボードでのペイン切り替え（bg 分割のアクティブ化時の adjustScroll の動作を含む）は対象外で、変更しない。
- as-2: テストで固定されている既存のマウス動作は変更しない（クリックでのカーソル移動、ドラッグでのマーク、クランプ、ダブルクリック、モーダル表示中の抑止）。
- as-3: 押下した項目が縮小後のファイル一覧より下にある場合、ダブルクリックの 2 回目の押下が同じ項目に当たらないことがある。

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- なし

## References

- 要件定義書: `feature-docs/press-scroll-click-mark/REQUIREMENTS.md`
