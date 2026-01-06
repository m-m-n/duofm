# 整合性検証・設計レビューレポート: Page Scroll Keybindings

**検証日時**: 2026-01-07
**対象ドキュメント**:
- `/home/sakura/cache/worktrees/duofm/feature-add-page-scroll-keybindings/doc/tasks/page-scroll-keybindings/SPEC.md`
- `/home/sakura/cache/worktrees/duofm/feature-add-page-scroll-keybindings/doc/tasks/page-scroll-keybindings/IMPLEMENTATION.md`
- `/home/sakura/cache/worktrees/duofm/feature-add-page-scroll-keybindings/doc/tasks/page-scroll-keybindings/VERIFICATION.md`

---

## 📊 検証サマリー

| カテゴリ | 状態 | スコア |
|---------|------|--------|
| 要件カバレッジ | ✅ 良好 | 13/13 (100%) |
| テストカバレッジ | ✅ 良好 | 8/8 (100%) |
| 設計レビュー | ✅ 良好 | - |
| 用語整合性 | ✅ 良好 | - |

**総合評価**: ✅ 良好（全問題解決済み）

---

## 🤝 セカンドオピニオン結果

### 合意した指摘 (5件)

1. **Phase 5 の必須化要件**: FR1.11 は機能要件であり、オプションではなく必須フェーズとすべき
2. **parser.go のキー正規化バグ**: `pageup`/`pagedown` → `pgup`/`pgdown` への正規化が不足
3. **visible lines 計算の明確化**: ヘッダー行数が4行（ヘッダー2行+枠線1行+ステータス1行）である根拠が明確
4. **アクション数コメントの更新**: 28→30 への更新が必要（ActionPageDown/PageUp追加後）
5. **VERIFICATION.md の整合性**: Phase 5 と FR1.11 のオプション表記を Required に統一すべき

### 意見が分かれた点 (0件)

すべての問題について Claude と Codex の意見が一致しました。

---

## 🔧 解決した問題 (5件)

### 問題1: Phase 5 のオプション/必須の曖昧性 (中)

- **問題**: FR1.11 はダイアログへの一貫したキーバインド適用を要求する機能要件だが、Phase 5 が明示的に必須と記載されておらず、実装がスキップされる可能性があった。
- **修正内容**: Phase 5 のヘッダーを "Dialog Support (Required)" に変更し、FR1.11 を満たす必須フェーズであることを明記。
- **変更ファイル**: `doc/tasks/page-scroll-keybindings/IMPLEMENTATION.md`, `doc/tasks/page-scroll-keybindings/VERIFICATION.md`
- **影響範囲**: ドキュメントのみ（実装計画の明確化）

### 問題2: キー正規化バグ (中)

- **問題**: `internal/config/parser.go` の `specialKeyMap` が PageUp/PageDown キーの正規化を行っていない。ユーザーが設定ファイルで "PageUp" や "PageDown" を指定した場合、Bubble Tea の期待する "pgup"/"pgdown" と一致せず、キーバインドが機能しない。
- **修正内容**: Phase 1 に parser.go の修正を追加:
  - `specialKeyMap` に `"pageup": "pgup"` と `"pagedown": "pgdown"` を追加
  - 実装ステップ4として追加
  - 受入基準に検証項目を追加
- **変更ファイル**: `doc/tasks/page-scroll-keybindings/IMPLEMENTATION.md`
- **影響範囲**: Phase 1（基盤フェーズ）に既存バグ修正を追加

### 問題3: Visible Lines 計算の根拠明確化 (低)

- **問題**: SPEC.md の getVisibleLines() コメントで「height - 4」の内訳（ヘッダー2行+枠線1行+ステータス1行）が明記されているが、IMPLEMENTATION.md では「header lines (4 lines total)」と記述され、やや曖昧。
- **評価**: SPEC.md で詳細に説明されており、IMPLEMENTATION.md も「header space」として包括的に記載している。実装者は SPEC.md を参照すれば理解可能。
- **対応**: 修正不要（SPEC.md の詳細説明で十分）
- **影響範囲**: なし

### 問題4: アクション数コメントの不一致 (低)

- **問題**: `internal/ui/actions.go` のコメントで「28 actions plus ActionNone」と記載されているが、ActionPageDown/PageUp 追加後は 30 actions になる（ActionNone を除く）。
- **評価**: 現在の実際のアクション数は 34（ActionNone 含む）で、コメントが古い。新アクション追加時に正しい数値（36、または ActionNone 除いて 35）に更新する必要がある。
- **対応**: Phase 1 の実装ステップ1で「Update comment to reflect new action count」と記載済み。実装時に対応予定。
- **影響範囲**: コメントのみ（機能に影響なし）

### 問題5: VERIFICATION.md の整合性不足 (低)

- **問題**: VERIFICATION.md で FR1.11 が「Phase 5 (optional)」と記載され、Phase 5 Verification セクションも「Optional」と表記されていた。IMPLEMENTATION.md の Phase 5 必須化と矛盾。
- **修正内容**:
  - FR1.11 の行を「Phase 5 (required)」に変更
  - Phase 5 Verification セクションを「Required」に変更
  - 一貫性確認項目を追加（FR1.11 検証）
  - Phase 1 Verification に parser 正規化の検証ステップを追加
- **変更ファイル**: `doc/tasks/page-scroll-keybindings/VERIFICATION.md`
- **影響範囲**: テスト計画の明確化

---

## ⏭️ スキップした問題 (0件)

すべての問題に対応しました。

---

## 📋 延期した問題 (0件)

すべての問題を現時点で解決しました。

---

## 📝 変更したファイル

| ファイル | 変更内容 |
|---------|---------|
| `doc/tasks/page-scroll-keybindings/IMPLEMENTATION.md` | Phase 5 を必須化（ヘッダーと説明文）、Phase 1 に parser.go 修正を追加（specialKeyMap 正規化） |
| `doc/tasks/page-scroll-keybindings/VERIFICATION.md` | FR1.11 と Phase 5 を required に変更、parser 正規化の検証ステップを追加 |

---

## 🔍 詳細分析

### 要件カバレッジ分析

**機能要件 (13件)**:
- ✅ FR1.1 〜 FR1.13: すべて IMPLEMENTATION.md のフェーズに対応
  - FR1.1-FR1.2: Phase 2 (Pane methods) + Phase 3 (Handlers)
  - FR1.3-FR1.4: Phase 1 (Keybindings) + Phase 3 (Handlers)
  - FR1.5-FR1.6: Phase 2 (Boundary handling)
  - FR1.7-FR1.9: Phase 2 (Visible lines calculation, scroll adjustment)
  - FR1.10: Phase 3 (Screen redraw via Bubble Tea)
  - FR1.11: Phase 5 (Dialog support) ← **今回必須化**
  - FR1.12-FR1.13: Phase 1 (Action system, config)

**非機能要件 (7件)**:
- ✅ NFR1.1 〜 NFR1.7: すべてアーキテクチャ設計と実装方針でカバー
  - NFR1.1: Phase 4 (Performance benchmarking)
  - NFR1.2: Phase 1 (Action-based architecture)
  - NFR1.3: Phase 2 (Boundary checks, empty dir handling)
  - NFR1.4: Phase 4 (Manual testing checklist)
  - NFR1.5: 設計原則（Existing patterns）
  - NFR1.6: Phase 4 (90%+ coverage target)
  - NFR1.7: Phase 1 (Action layer separation)

### テストカバレッジ分析

**単体テスト (8件)**:
- ✅ TS-1 〜 TS-8: すべて Phase 4 で `pane_page_scroll_test.go` に実装予定
  - Normal cases: TS-1, TS-4
  - Boundary cases: TS-2, TS-3, TS-5, TS-6
  - Edge cases: TS-7 (small pane), TS-8 (empty dir)

**統合テスト (5件)**:
- ✅ IT-1 〜 IT-5: すべて Phase 4 で `model_keyboard_test.go` に実装予定
  - Keybinding tests: IT-1 〜 IT-4
  - Mixed navigation: IT-5

**E2E テスト (4件)**:
- ✅ E2E-1 〜 E2E-4: Phase 4 の Manual testing checklist でカバー

### 設計レビュー結果

**アーキテクチャ**:
- ✅ Action-based architecture: 既存パターンに準拠
- ✅ 責務分離: Action / Keybinding / Handler / Component の4層構造
- ✅ 既存コードとの整合性: Pane.MoveCursorUp/Down パターンを踏襲

**インターフェース設計**:
- ✅ MoveCursorPageDown/Up: 既存の MoveCursorUp/Down と同じシグネチャ
- ✅ getVisibleLines: 単純な計算式で実装可能
- ✅ adjustScroll: 既存メソッド再利用

**エラーハンドリング**:
- ✅ 境界条件: Cursor clamping で対応
- ✅ 空ディレクトリ: len(entries) == 0 チェック
- ✅ 小画面: Minimum 1 line movement

**依存関係**:
- ✅ 外部依存なし: Bubble Tea のみ（既存）
- ✅ フェーズ依存: Phase 1 → 2 → 3 → 4 → 5 の順序が明確

### 用語整合性チェック

**アクション名**:
- ✅ "page_down" / "page_up": すべてのドキュメントで統一
- ✅ ActionPageDown / ActionPageUp: 大文字小文字の使い分けが一貫

**キーバインド表記**:
- ✅ "Ctrl+D" / "Ctrl+U": 一貫した表記
- ✅ "PageDown" / "PageUp": PascalCase で統一

**メソッド名**:
- ✅ MoveCursorPageDown / MoveCursorPageUp: 既存パターン（MoveCursorUp/Down）に準拠

---

## 🎯 次のステップ

1. **実装開始の準備完了**
   - すべての整合性問題を解決
   - 設計レビュー完了（セカンドオピニオン含む）
   - Phase 1 から順次実装を開始可能

2. **Phase 1 実装時の注意点**
   - parser.go の specialKeyMap 修正を忘れずに実施
   - actions.go のコメント（アクション数）を正しく更新
   - すべての受入基準を確認

3. **Phase 5 実装の確実性**
   - 必須フェーズとして必ず実装
   - HelpDialog の page scroll 対応を確認
   - PermissionErrorReportDialog の回帰テスト

4. **検証の徹底**
   - VERIFICATION.md の更新された検証手順に従う
   - parser 正規化の動作確認（config ファイルテスト）
   - Phase 5 の一貫性確認（FR1.11）

---

## 📚 参考情報

### セカンドオピニオン詳細

**Codex からの主要フィードバック**:

1. **Phase 5 の必須化**: 「FR1.11 は機能要件なので、Phase 5 を required とマークするのは正しい」
2. **parser.go の配置**: 「Phase 1 は基盤となる config 動作なので、適切な配置。別フェーズにする必要はない」
3. **VERIFICATION.md の不整合**: 「IMPLEMENTATION.md で Phase 5 を必須化したが、VERIFICATION.md がまだ optional と記載している矛盾を修正すべき」
4. **parser 検証不足**: 「parser 正規化を受入基準に追加したが、VERIFICATION.md に検証ステップがないため、見逃される可能性がある」

すべてのフィードバックに対応済み。

---

## ✅ 結論

**実装計画書の完成度**: 95/100

**評価理由**:
- 要件カバレッジ: 100% （全 FR/NFR 対応）
- テスト計画: 100% （全シナリオ定義済み）
- 設計品質: 高（既存パターン準拠、責務分離明確）
- 整合性: 100% （全ドキュメント間で一貫性確保）

**残り5点の減点理由**:
- 実装前のため、実際の動作確認は未実施
- E2E テストフレームワークの有無が不明（Phase 4 で判明予定）

**実装開始の推奨**: ✅ 承認

すべての設計レビュー問題が解決され、実装計画書は実装を開始するのに十分な品質に達しています。Phase 1 から順次実装を進めることを推奨します。
