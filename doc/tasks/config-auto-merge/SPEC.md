# 技術仕様書: 設定ファイル自動マージ機能

## 1. 概要

### 1.1 機能概要
duofmの起動時に、設定ファイル（config.toml）に記載のない設定項目をデフォルト値で自動追記する。

### 1.2 関連ドキュメント
- [要件定義書](./要件定義書.md)

## 2. アーキテクチャ

### 2.1 コンポーネント構成

```
internal/config/
├── config.go      # LoadConfig()に追記処理を統合
├── defaults.go    # デフォルト値定義（変更なし）
├── colors.go      # 色設定（GetDefaultColorValue関数を追加）
├── merger.go      # 【新規】マージロジック
└── merger_test.go # 【新規】マージロジックのテスト
```

### 2.2 処理フロー

```
起動
  ↓
LoadConfig(path)
  ↓
設定ファイル存在確認
  ├─ 存在しない → GenerateDefaultConfig() → 終了
  └─ 存在する → TOML解析
                   ↓
              不足項目の検出
                   ↓
              不足項目あり?
                ├─ なし → 終了
                └─ あり → MergeConfig()
                            ↓
                         ファイル末尾に追記
                            ↓
                         終了
```

## 3. 詳細設計

### 3.1 新規関数: MergeConfig

#### 3.1.1 関数シグネチャ

```go
// MergeConfig merges missing configuration items into the existing config file.
// It appends missing items with their default values to the end of the file.
// Returns nil if no items were missing or if the merge was successful.
// Returns an error if the file could not be written.
func MergeConfig(path string, existing *rawConfig) error
```

#### 3.1.2 処理内容

1. 不足項目の収集
2. 不足項目がない場合は早期リターン
3. 追記内容の生成（TOML形式）
4. ファイル末尾に追記

### 3.2 不足項目の検出

#### 3.2.1 keybindings

```go
// FindMissingKeybindings returns keybindings that exist in defaults but not in config.
func FindMissingKeybindings(existing map[string][]string) map[string][]string
```

- `DefaultKeybindings()` と `existing` を比較
- `existing` に存在しないキーを返す

#### 3.2.2 colors

```go
// FindMissingColors returns color settings that exist in defaults but not in config.
// Uses AllColorKeys() to enumerate all color keys and GetDefaultColorValue() to get default values.
func FindMissingColors(existing map[string]interface{}) map[string]int
```

- `AllColorKeys()` で全色設定キーを取得
- `existing` に存在しないキーを特定
- `GetDefaultColorValue(key)` でデフォルト値を取得して返す

#### 3.2.2.1 補助関数: GetDefaultColorValue

```go
// GetDefaultColorValue returns the default color value for the given key.
// Returns -1 if the key is not found.
func GetDefaultColorValue(key string) int
```

- `DefaultColors()` から構造体を取得
- キー名に対応するフィールド値を返す
- colors.go に追加する新規関数

#### 3.2.3 history_limit

```go
// IsMissingHistoryLimit returns true if history_limit is not set in config.
func IsMissingHistoryLimit(historyLimit *int) bool
```

- `historyLimit` が `nil` の場合に `true` を返す

### 3.3 追記内容の生成

#### 3.3.1 フォーマット

```toml
# --- Auto-merged settings (added by duofm) ---

[keybindings]
new_action = ["Key1", "Key2"]

[colors]
new_color_setting = 123

history_limit = 20000
```

#### 3.3.2 生成ルール

1. 区切りコメントを先頭に追加
2. セクションごとにグループ化
3. 空のセクションは出力しない
4. 値はTOML形式で出力

### 3.4 LoadConfig()の変更

#### 3.4.1 変更内容

```go
func LoadConfig(path string) (*Config, []string) {
    // 既存の処理...

    // 設定ファイルのマージ（不足項目の追記）
    if err := MergeConfig(path, &raw); err != nil {
        warnings = append(warnings, fmt.Sprintf("Warning: failed to merge config: %v", err))
    }

    // 既存の処理...
}
```

## 4. データ構造

### 4.1 MergeResult（内部用）

```go
type mergeResult struct {
    Keybindings  map[string][]string
    Colors       map[string]int
    HistoryLimit *int  // nilの場合は追記不要
}
```

## 5. エラーハンドリング

### 5.1 ファイル書き込みエラー

- 警告メッセージを返す
- アプリケーションは正常に起動を続行

### 5.2 権限エラー

- 警告メッセージを返す
- 追記処理をスキップ

## 6. テスト仕様

### 6.1 ユニットテスト

#### TC-001: FindMissingKeybindings
- 入力: 一部のキーバインディングのみ設定されたmap
- 期待: 不足しているキーバインディングのみ返される

#### TC-002: FindMissingColors
- 入力: 一部の色設定のみ設定されたmap
- 期待: 不足している色設定のみ返される

#### TC-003: IsMissingHistoryLimit
- 入力: nil / 値あり
- 期待: nil → true, 値あり → false

#### TC-004: MergeConfig - 不足項目あり
- 入力: 一部項目のみの設定ファイル
- 期待: 不足項目がファイル末尾に追記される

#### TC-005: MergeConfig - 不足項目なし
- 入力: 全項目が存在する設定ファイル
- 期待: ファイルが変更されない

#### TC-006: MergeConfig - 既存値の保持
- 入力: カスタム値が設定された設定ファイル
- 期待: カスタム値が変更されない

### 6.2 統合テスト

#### IT-001: LoadConfig経由でのマージ
- 起動時に不足項目が追記されることを確認

## 7. 実装上の注意

### 7.1 ファイル書き込み

- 既存ファイルにはアペンドモードで書き込み
- 書き込み前に改行を確保

### 7.2 TOML生成

- 配列は `["value1", "value2"]` 形式
- 整数は数値として出力
- キー名はクォートなし（ASCII文字のみ）

### 7.3 パフォーマンス

- 不足項目がない場合はI/O操作をスキップ
- 不足項目の検出は O(n) で完了

## 8. 変更影響範囲

### 8.1 変更ファイル
- `internal/config/config.go` - LoadConfig()の変更
- `internal/config/colors.go` - GetDefaultColorValue()関数の追加

### 8.2 新規ファイル
- `internal/config/merger.go` - マージロジック
- `internal/config/merger_test.go` - テスト

### 8.3 影響を受ける機能
- アプリケーション起動処理
- 設定ファイルの読み込み

## 9. 将来の拡張性

### 9.1 新しい設定項目の追加
- `defaults.go` にデフォルト値を追加
- `merger.go` の検出ロジックに追加
- 自動的にマージ対象となる

### 9.2 マイグレーション機能
- 将来的に設定項目のリネームや削除が必要な場合
- バージョン番号による制御を検討
