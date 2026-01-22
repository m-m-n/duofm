# Implementation Plan: Disable Open/Open with Menu Without Desktop Environment

## Overview

デスクトップ環境が存在しない場合（SSH接続、ヘッドレスサーバーなど）、コンテキストメニューの「Open」および「Open with ...」項目をグレーアウトし選択不可にする機能を実装する。

## Objectives

- 環境変数を使用してデスクトップ環境の有無を検出する
- デスクトップ環境が検出されない場合、メニュー項目を無効化する
- デスクトップ環境が利用可能な場合は通常動作を維持する

## Prerequisites

### Development Environment
- Go 1.21+
- duofm プロジェクトがビルド可能な状態

### Dependencies
- 追加の外部依存なし（標準ライブラリのosパッケージのみ使用）

### Knowledge Requirements
- Go の環境変数アクセス（os.Getenv）
- 既存のコンテキストメニュー実装（MenuItem.Enabledフィールド）

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea
- **Key Components**:
  - `internal/ui/env.go` - 環境検出モジュール（新規）
  - `internal/ui/context_menu_dialog.go` - 既存メニューへの統合

### Design Approach
環境検出を起動時に一度だけ実行し、結果をキャッシュする。コンテキストメニュー生成時にこのキャッシュ値を参照してメニュー項目のEnabled状態を決定する。

### Component Interaction
```
起動時: detectDesktopEnvironment() → hasDesktop（キャッシュ）
メニュー生成時: buildOpenMenuItems() → HasDesktopEnvironment() → Enabled設定
```

## Implementation Phases

### Phase 1: Desktop Environment Detection Module

**Goal**: デスクトップ環境の有無を検出する独立したモジュールを作成する

**Files to Create**:
- `internal/ui/env.go` - 環境検出ロジック
- `internal/ui/env_test.go` - 単体テスト

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| detectDesktopEnvironment | 環境変数をチェックしてデスクトップ環境の有無を判定 | なし | bool値を返す |
| HasDesktopEnvironment | キャッシュされた検出結果を外部に公開 | アプリケーション起動済み | キャッシュ値を返す |
| setDesktopEnvironmentForTest | テスト用にキャッシュ値を上書き（env_test.go内に配置、非公開） | なし | キャッシュ値が更新される |

**Processing Flow**:
```
1. パッケージ初期化時にdetectDesktopEnvironmentを実行
   ├─ DISPLAY環境変数が非空 → true
   └─ WAYLAND_DISPLAY環境変数が非空 → true
   └─ 両方が空または未設定 → false
2. 結果をパッケージ変数にキャッシュ
3. HasDesktopEnvironment()呼び出し時にキャッシュ値を返す
```

**Implementation Steps**:

1. **env.goファイル作成**
   - パッケージ変数でキャッシュを保持
   - 初期化時に検出関数を実行
   - 公開関数でキャッシュ値を返す

2. **テスト用ヘルパー関数作成**
   - `setDesktopEnvironmentForTest`関数を`env_test.go`に配置（非公開）
   - 同一パッケージ内のテストからのみアクセス可能

**Dependencies**:
- Requires: なし
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T1-1 | DISPLAYが設定されている | DISPLAY=":0" | true |
| T1-2 | WAYLAND_DISPLAYが設定されている | WAYLAND_DISPLAY="wayland-0" | true |
| T1-3 | 両方が未設定 | (環境変数なし) | false |
| T1-4 | DISPLAYが空文字列 | DISPLAY="" | false |
| T1-5 | 両方が設定されている | DISPLAY=":0", WAYLAND_DISPLAY="wayland-0" | true |

**Acceptance Criteria**:
- [ ] HasDesktopEnvironment()がDISPLAY設定時にtrueを返す
- [ ] HasDesktopEnvironment()がWAYLAND_DISPLAY設定時にtrueを返す
- [ ] HasDesktopEnvironment()が両方未設定時にfalseを返す
- [ ] 空文字列は未設定として扱う
- [ ] 全ての単体テストが通過する

**Estimated Effort**: 小 (1-2 hours)

---

### Phase 2: Context Menu Integration

**Goal**: コンテキストメニューの「Open」「Open with ...」項目にデスクトップ環境チェックを統合する

**Files to Modify**:
- `internal/ui/context_menu_dialog.go`:
  - `buildOpenMenuItems`関数を修正し、HasDesktopEnvironment()の結果をEnabled設定に反映

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| buildOpenMenuItems | Open/Open with項目を生成 | entryが有効 | Enabledが環境に応じて設定された項目リスト |

**Processing Flow**:
```
1. buildOpenMenuItems呼び出し
2. HasDesktopEnvironment()を呼び出して環境を確認
3. Open項目のEnabled設定
   ├─ デスクトップ環境あり AND markCount == 0 → true
   └─ それ以外 → false
4. Open with項目のEnabled設定
   ├─ デスクトップ環境あり → true
   └─ デスクトップ環境なし → false
```

**Implementation Steps**:

1. **buildOpenMenuItems関数修正**
   - HasDesktopEnvironment()の結果をEnabled条件に追加
   - 既存のmarkCount条件と組み合わせる

2. **無効項目スキップのナビゲーション修正**（FR3対応）
   - moveUp/moveDown関数で無効項目をスキップするロジックを追加
   - 無限ループ防止: 全項目が無効の場合のガード処理を実装
   - スキップ回数がアイテム数を超えた場合は現在位置を維持

**Dependencies**:
- Requires: Phase 1
- Blocks: なし

**Testing Approach**:

*Unit Tests*:
| ID | Scenario | Input | Expected |
|----|----------|-------|----------|
| T2-1 | デスクトップ環境あり、マークなし | hasDesktop=true, markCount=0 | Open=enabled, OpenWith=enabled |
| T2-2 | デスクトップ環境あり、マークあり | hasDesktop=true, markCount>0 | Open=disabled, OpenWith=enabled |
| T2-3 | デスクトップ環境なし、マークなし | hasDesktop=false, markCount=0 | Open=disabled, OpenWith=disabled |
| T2-4 | デスクトップ環境なし、マークあり | hasDesktop=false, markCount>0 | Open=disabled, OpenWith=disabled |

*Integration Tests*:
| ID | Scenario | Expected |
|----|----------|----------|
| T2-5 | メニュー描画時に無効項目がグレー表示 | 視覚的に区別可能 |
| T2-6 | キーボードナビゲーションで無効項目をスキップ | 無効項目に止まらない |

**Acceptance Criteria**:
- [ ] デスクトップ環境がない場合、Open項目が無効化される
- [ ] デスクトップ環境がない場合、Open with項目が無効化される
- [ ] デスクトップ環境がある場合、既存の動作を維持
- [ ] 無効化された項目がグレー表示される
- [ ] 全ての単体テストが通過する

**Estimated Effort**: 小 (1-2 hours)

---

## Complete File Structure

```
internal/ui/
├── env.go                      # Desktop environment detection (新規)
├── env_test.go                 # Tests for env.go (新規)
├── context_menu_dialog.go      # Modified to use HasDesktopEnvironment()
└── context_menu_dialog_test.go # Additional tests for desktop env integration
```

**File Descriptions**:
- `env.go`: 環境変数を使用したデスクトップ環境検出ロジック。起動時に一度検出し結果をキャッシュ
- `env_test.go`: 環境検出の単体テスト。`setDesktopEnvironmentForTest`を含み、テスト時にキャッシュを制御
- `context_menu_dialog.go`: 既存ファイル。buildOpenMenuItemsでHasDesktopEnvironment()を参照
- `context_menu_dialog_test.go`: 既存ファイル。デスクトップ環境有無によるメニュー項目テストを追加

## Testing Strategy

### Unit Testing

**Approach**:
- Go の標準 `testing` パッケージを使用
- Table-driven tests で複数シナリオをカバー
- テスト用ヘルパー関数でキャッシュを制御

**Test Coverage Goals**:
- env.go: 100%
- context_menu_dialog.go (修正部分): 100%

**Key Test Areas**:
1. **Environment Detection** (`internal/ui/env.go`)
   - DISPLAY環境変数のチェック
   - WAYLAND_DISPLAY環境変数のチェック
   - 両方未設定のケース
   - 空文字列のケース

2. **Menu Item Generation** (`internal/ui/context_menu_dialog.go`)
   - デスクトップ環境の有無によるEnabled状態
   - 既存のmarkCount条件との組み合わせ

### Manual Testing Checklist

- [ ] SSH接続環境でduofmを起動し、コンテキストメニューを開く
- [ ] Open項目がグレー表示されていることを確認
- [ ] Open with項目がグレー表示されていることを確認
- [ ] j/kキーでナビゲーション時に無効項目をスキップすることを確認
- [ ] 数字キーで無効項目を選択しても何も起きないことを確認
- [ ] デスクトップ環境でduofmを起動し、Open/Open with項目が通常動作することを確認

## Dependencies

### Internal Dependencies

**Implementation Order**:
1. Phase 1: env.go（依存なし）
2. Phase 2: context_menu_dialog.go修正（Phase 1に依存）

**Component Dependencies**:
- `context_menu_dialog.go` depends on `env.go`

## Risk Assessment

### Technical Risks

1. **環境変数キャッシュのタイミング**
   - **Risk**: パッケージ初期化順序による予期しない動作
   - **Likelihood**: 低
   - **Impact**: 中
   - **Mitigation**: Goの初期化順序（パッケージ変数→init関数）を理解し、適切に実装

### Implementation Risks

1. **既存テストへの影響**
   - **Risk**: 既存のコンテキストメニューテストがデスクトップ環境前提で書かれている
   - **Likelihood**: 中
   - **Impact**: 中
   - **Mitigation**: `setDesktopEnvironmentForTest`関数（env_test.go内、非公開）でテスト時にキャッシュを制御

## Performance Considerations

1. **環境検出のタイミング**
   - 起動時に一度だけ実行し、結果をキャッシュ
   - 毎回の環境変数チェックを回避

## Security Considerations

- 環境変数の読み取りのみで、セキュリティ上の懸念なし

## Open Questions

- なし

## Success Metrics

### Functional Completeness
- [ ] デスクトップ環境検出が正しく動作
- [ ] メニュー項目が適切に無効化される
- [ ] 既存機能が影響を受けない

### Quality Metrics
- [ ] テストカバレッジ目標達成
- [ ] 全テストが通過
- [ ] コードが既存規約に準拠

## References

- **Specification**: `doc/tasks/hide-open-menu-without-desktop/SPEC.md`
- **Related Files**:
  - `internal/ui/context_menu_dialog.go` - 既存コンテキストメニュー実装
  - `internal/ui/context_menu_dialog_test.go` - 既存テスト

## Next Steps

1. **Review and Approval**
   - 実装計画のレビュー
   - 不明点の確認

2. **Begin Implementation**
   - Phase 1: env.go作成
   - Phase 2: context_menu_dialog.go修正

3. **Verification**
   - 単体テスト実行
   - 手動テスト実行
