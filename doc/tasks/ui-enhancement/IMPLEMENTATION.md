# 実装計画: UIブラッシュアップ

## 概要

duofmのTUIを強化し、Total Commanderのような実用的で情報豊富なファイルマネージャーインターフェースを実現します。本実装では、ヘッダー情報の拡張、複数の表示モード切り替え、シンボリックリンクの適切な処理、大規模ディレクトリ読み込み時のローディング表示を追加します。

## 目的

- **情報の可視化**: マーク数、容量、空き容量などの重要な情報を常時表示
- **柔軟な表示**: 端末幅とユーザーの好みに応じて表示モードを切り替え
- **シンボリックリンク対応**: リンク先情報の表示とリンク先への移動を実現
- **応答性向上**: 大量ファイル読み込み時のフィードバック提供
- **実用性の向上**: Total Commanderレベルの操作性を提供

## 前提条件

### 環境
- Go 1.21以上
- Bubble Tea v1.3.10（既存）
- Lip Gloss v1.1.0（既存）
- golang.org/x/sys v0.36.0（既存、ディスク容量取得に使用）

### 既存の実装状況
- 基本的な2ペインUI実装済み（`internal/ui/model.go`, `internal/ui/pane.go`）
- ファイル読み込み機能実装済み（`internal/fs/reader.go`）
- FileEntry構造体定義済み（`internal/fs/types.go`）
- 基本的なキーバインディング実装済み（`internal/ui/keys.go`）

### 実装開始前の準備
- SPEC.mdと要件定義書.mdの理解
- 既存コードベースの構造把握
- Bubble TeaのElm Architectureの理解（Update/View分離）

## アーキテクチャ概要

### 設計原則
1. **レイヤー分離**: UI層（internal/ui）とファイルシステム層（internal/fs）を明確に分離
2. **状態管理**: Bubble Teaのモデルパターンに従い、状態を不変に扱う
3. **非同期処理**: goroutineとBubble Teaのメッセージングで大規模ディレクトリ読み込みを非ブロッキング化
4. **レスポンシブ**: 端末幅に応じた動的なレイアウト調整

### 主要な拡張ポイント

#### 1. データ構造の拡張
- **FileEntry**: 所有者、グループ、シンボリックリンク情報を追加
- **Pane**: 表示モード、ローディング状態、最後の選択モードを追加
- **Model**: ディスク空き容量、最終チェック時刻を追加

#### 2. 新規モジュール
- **internal/fs/diskspace.go**: ディスク容量取得
- **internal/fs/symlink.go**: シンボリックリンク処理
- **internal/fs/owner.go**: ファイル所有者情報取得
- **internal/ui/format.go**: フォーマット関数（サイズ、時刻、パーミッション）
- **internal/ui/messages.go**: Bubble Teaのカスタムメッセージ定義

#### 3. 表示モード切り替えメカニズム
- 端末幅を監視して自動的にMinimalモードに切り替え
- ユーザーが`i`キーでBasic/Detailモードを切り替え
- 各ペインが独立して表示モードを管理

## 実装フェーズ

### Phase 1: データ構造とファイル情報取得の拡張

**目標**: FileEntryに必要な情報を追加し、ファイルシステム層で情報取得機能を実装する

**作成・修正するファイル**:
- `internal/fs/types.go` - FileEntry構造体の拡張
- `internal/fs/owner.go` - ファイル所有者・グループ情報取得（新規作成）
- `internal/fs/symlink.go` - シンボリックリンク情報取得（新規作成）
- `internal/fs/diskspace.go` - ディスク空き容量取得（新規作成）
- `internal/fs/reader.go` - ReadDirectory関数の拡張

**実装ステップ**:

1. **FileEntry構造体の拡張** (`internal/fs/types.go`)
   - 以下のフィールドを追加:
     ```go
     Owner       string      // 所有者名
     Group       string      // グループ名
     IsSymlink   bool        // シンボリックリンクか
     LinkTarget  string      // リンク先パス（シンボリックリンクの場合）
     LinkBroken  bool        // リンク切れか（シンボリックリンクの場合）
     ```
   - DisplayName()メソッドをシンボリックリンク対応に更新
     - シンボリックリンクの場合: `"name -> target"` 形式で返す
     - 長いリンク先は省略（...で切り詰め、最大幅は引数で受け取る）

2. **ファイル所有者・グループ情報取得** (`internal/fs/owner.go` 新規作成)
   ```go
   // GetFileOwnerGroup は指定されたパスのファイルの所有者とグループ名を返す
   func GetFileOwnerGroup(path string) (owner, group string, err error)
   ```
   - Unix/Linux: syscall.Stat_tからUIDとGIDを取得し、user.LookupId/LookupGroupIdで名前解決
   - Windows: 所有者情報は制限的、"N/A"を返すか、Windows APIを使用（将来対応）
   - エラー時はプレースホルダー（"unknown"）を返す

3. **シンボリックリンク情報取得** (`internal/fs/symlink.go` 新規作成)
   ```go
   // GetSymlinkInfo はシンボリックリンクの情報を取得する
   // target: リンク先のパス（絶対パスに解決）
   // isBroken: リンク先が存在しない場合true
   func GetSymlinkInfo(path string) (target string, isBroken bool, err error)
   ```
   - `os.Readlink()`でリンク先を取得
   - 相対パスの場合は絶対パスに変換（`filepath.Join` + `filepath.Clean`）
   - `os.Stat()`でリンク先の存在確認（エラーならisBroken=true）

4. **ディスク空き容量取得** (`internal/fs/diskspace.go` 新規作成)
   ```go
   // GetDiskSpace は指定されたパスが属するパーティションの空き容量を返す
   // 戻り値: freeBytes (利用可能バイト数)
   func GetDiskSpace(path string) (freeBytes uint64, err error)
   ```
   - Unix/Linux/macOS: `syscall.Statfs()`を使用
   - Windows: `syscall.GetDiskFreeSpaceEx()`を使用
   - エラー時は0を返す

5. **ReadDirectory関数の拡張** (`internal/fs/reader.go`)
   - 各エントリに対して以下の情報を追加取得:
     - `entry.Type()&os.ModeSymlink != 0` でシンボリックリンク判定
     - シンボリックリンクの場合: `GetSymlinkInfo()`を呼び出し
     - 所有者・グループ: `GetFileOwnerGroup()`を呼び出し
   - エラーハンドリング: 個別ファイルの情報取得失敗は無視（continue）し、取得可能な情報のみ格納

**依存関係**:
- なし（最初のフェーズ）

**テスト**:
- `internal/fs/owner_test.go`: 所有者・グループ取得のユニットテスト
  - 既存ファイルの所有者取得成功
  - 存在しないファイルでエラーハンドリング
- `internal/fs/symlink_test.go`: シンボリックリンク処理のユニットテスト
  - 有効なシンボリックリンクの情報取得
  - リンク切れの検出
  - 相対パス→絶対パス変換
- `internal/fs/diskspace_test.go`: ディスク容量取得のユニットテスト
  - ルートディレクトリの空き容量取得
  - ホームディレクトリの空き容量取得
- `internal/fs/reader_test.go`: ReadDirectory拡張のテスト
  - シンボリックリンクを含むディレクトリの読み込み
  - 所有者情報の正しい設定

**推定工数**: 中（2-3日）

---

### Phase 2: フォーマット関数の実装

**目標**: ファイルサイズ、タイムスタンプ、パーミッションなどの表示用フォーマット関数を実装する

**作成・修正するファイル**:
- `internal/ui/format.go` - 各種フォーマット関数（新規作成）

**実装ステップ**:

1. **ファイルサイズフォーマット関数**
   ```go
   // FormatSize はバイト数を人間が読みやすい形式に変換（1024進法）
   // 例: 512 B, 1.5 KiB, 2.3 MiB, 1.8 GiB, 3.2 TiB
   func FormatSize(bytes int64) string
   ```
   - 1024進法で単位変換（B, KiB, MiB, GiB, TiB）
   - 小数点1桁まで表示（1.5 KiB）
   - 1024未満はB単位で整数表示（512 B）
   - ディレクトリは"-"を返す（呼び出し側で判定）

2. **タイムスタンプフォーマット関数**
   ```go
   // FormatTimestamp は時刻をISO 8601形式でフォーマット
   // フォーマット: "2024-12-17 22:28" (固定)
   func FormatTimestamp(t time.Time) string
   ```
   - `t.Format("2006-01-02 15:04")` を使用
   - 固定幅16文字
   - カスタマイズ不可（仕様で固定）

3. **パーミッションフォーマット関数**
   ```go
   // FormatPermissions はfs.FileModeをUnix形式の文字列に変換
   // 例: "rwxr-xr-x" (固定10文字)
   func FormatPermissions(mode fs.FileMode) string
   ```
   - ファイルタイプ（1文字）+ 所有者(3) + グループ(3) + その他(3)
   - `-rwxr-xr-x` の形式
   - シンボリックリンク: `l`、ディレクトリ: `d`、通常ファイル: `-`

4. **ディスク容量フォーマット関数**
   ```go
   // FormatDiskSpace はディスク空き容量を人間が読みやすい形式に変換
   // FormatSizeと同じロジック（エイリアス化も可）
   func FormatDiskSpace(bytes uint64) string
   ```

5. **カラム幅計算関数**
   ```go
   // CalculateColumnWidths は表示モードに応じた各カラムの幅を計算
   func CalculateColumnWidths(mode DisplayMode, paneWidth int) ColumnWidths

   type ColumnWidths struct {
       Name        int
       Size        int
       Timestamp   int
       Permissions int
       Owner       int
       Group       int
   }
   ```
   - 固定幅カラム（タイムスタンプ、パーミッション）のサイズを確定
   - 残りをName幅に割り当て
   - 必要最小幅を満たせない場合はMinimalモードフラグを返す

**依存関係**:
- Phase 1完了（FileEntry構造体の拡張）

**テスト**:
- `internal/ui/format_test.go`: フォーマット関数のテーブル駆動テスト
  - FormatSize: 各単位の変換（0 B, 512 B, 1.0 KiB, 1.5 MiB, 2.3 GiB, 1.8 TiB）
  - FormatTimestamp: 固定フォーマット出力
  - FormatPermissions: 各パーミッションパターン（777, 644, 755など）
  - CalculateColumnWidths: 各表示モードでの幅計算

**推定工数**: 小（1日）

---

### Phase 3: 表示モード管理の実装

**目標**: 表示モード（Minimal, Basic, Detail）の切り替えロジックと状態管理を実装する

**作成・修正するファイル**:
- `internal/ui/pane.go` - Pane構造体とメソッドの拡張
- `internal/ui/display_mode.go` - 表示モード定義と関連関数（新規作成）

**実装ステップ**:

1. **DisplayMode型の定義** (`internal/ui/display_mode.go` 新規作成)
   ```go
   type DisplayMode int

   const (
       DisplayMinimal DisplayMode = iota  // 名前のみ
       DisplayBasic                        // 名前 + サイズ + タイムスタンプ
       DisplayDetail                       // 名前 + パーミッション + 所有者 + グループ
   )

   // String は表示モードの文字列表現を返す
   func (d DisplayMode) String() string

   // MinRequiredWidth は各モードで必要な最小端末幅を返す
   func (d DisplayMode) MinRequiredWidth() int
   ```
   - Minimal: 最小幅30カラム（ファイル名のみ）
   - Basic: 最小幅60カラム（名前 + サイズ + タイムスタンプ）
   - Detail: 最小幅70カラム（名前 + パーミッション + 所有者 + グループ）
   - 実装時に実測して調整

2. **Pane構造体の拡張** (`internal/ui/pane.go`)
   ```go
   type Pane struct {
       path              string
       entries           []fs.FileEntry
       cursor            int
       scrollOffset      int
       width             int
       height            int
       isActive          bool
       displayMode       DisplayMode  // 現在の表示モード
       userSelectedMode  DisplayMode  // ユーザーが選択したモード（Basic or Detail）
       loading           bool         // ローディング中フラグ
       loadingProgress   string       // ローディングメッセージ
   }
   ```
   - `userSelectedMode`: 端末幅が広い時に復元するモード（BasicかDetail）
   - `displayMode`: 実際に使用される表示モード（端末幅が狭い場合は自動的にMinimal）

3. **表示モード切り替えメソッド** (`internal/ui/pane.go`)
   ```go
   // ToggleDisplayMode はBasicとDetailを切り替える（端末幅が十分な場合のみ）
   func (p *Pane) ToggleDisplayMode()

   // updateDisplayMode は端末幅に応じて表示モードを自動調整
   func (p *Pane) updateDisplayMode()

   // CanToggleMode は現在iキーが有効かどうかを返す
   func (p *Pane) CanToggleMode() bool
   ```
   - `ToggleDisplayMode()`: BasicとDetailを交互に切り替え、userSelectedModeを更新
   - `updateDisplayMode()`: 端末幅とuserSelectedModeから実際のdisplayModeを決定
     - 幅が狭い → 強制的にMinimal
     - 幅が十分 → userSelectedMode（BasicまたはDetail）を使用
   - `CanToggleMode()`: displayMode != Minimal の場合true

4. **SetSizeメソッドの更新** (`internal/ui/pane.go`)
   - 既存のSetSize()メソッド内で`updateDisplayMode()`を呼び出し
   - 端末リサイズ時に自動的にモードを再計算

**依存関係**:
- Phase 2完了（フォーマット関数、特にCalculateColumnWidths）

**テスト**:
- `internal/ui/pane_test.go`: 表示モード関連のテスト
  - 端末幅が狭い場合、自動的にMinimalモードになる
  - 端末幅が広い場合、iキーでBasic⇔Detailを切り替えられる
  - 端末リサイズ時にモードが適切に更新される
  - CanToggleMode()が正しくtrue/falseを返す
  - 各ペインが独立してモードを管理できる

**推定工数**: 中（2日）

---

### Phase 4: ヘッダー拡張とディスク容量表示

**目標**: ヘッダーを2行構成に拡張し、マーク情報とディスク空き容量を表示する

**作成・修正するファイル**:
- `internal/ui/pane.go` - View()メソッドとヘッダーレンダリングの拡張
- `internal/ui/model.go` - Model構造体とディスク容量管理の拡張
- `internal/ui/messages.go` - ディスク容量更新メッセージ（新規作成）

**実装ステップ**:

1. **Model構造体の拡張** (`internal/ui/model.go`)
   ```go
   type Model struct {
       leftPane           *Pane
       rightPane          *Pane
       leftPath           string
       rightPath          string
       activePane         PanePosition
       dialog             Dialog
       width              int
       height             int
       ready              bool
       lastDiskSpaceCheck time.Time      // 最後のディスク容量チェック時刻
       leftDiskSpace      uint64          // 左ペインのディスク空き容量
       rightDiskSpace     uint64          // 右ペインのディスク空き容量
   }
   ```

2. **ディスク容量更新メッセージの定義** (`internal/ui/messages.go` 新規作成)
   ```go
   // diskSpaceUpdateMsg はディスク容量の定期更新を通知
   type diskSpaceUpdateMsg struct{}

   // diskSpaceTickCmd は5秒後にdiskSpaceUpdateMsgを送信するコマンド
   func diskSpaceTickCmd() tea.Cmd {
       return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
           return diskSpaceUpdateMsg{}
       })
   }
   ```

3. **Model.Update()の拡張** (`internal/ui/model.go`)
   - `diskSpaceUpdateMsg`を受信した時:
     ```go
     case diskSpaceUpdateMsg:
         m.updateDiskSpace()
         return m, diskSpaceTickCmd()  // 次の5秒後に再度更新
     ```
   - `tea.WindowSizeMsg`の初期化時にディスク容量を初回取得し、タイマー開始
   - ディレクトリ移動時にもディスク容量を更新（パーティションが変わる可能性）

4. **updateDiskSpace()メソッドの実装** (`internal/ui/model.go`)
   ```go
   func (m *Model) updateDiskSpace() {
       if leftSpace, err := fs.GetDiskSpace(m.leftPane.Path()); err == nil {
           m.leftDiskSpace = leftSpace
       }
       if rightSpace, err := fs.GetDiskSpace(m.rightPane.Path()); err == nil {
           m.rightDiskSpace = rightSpace
       }
       m.lastDiskSpaceCheck = time.Now()
   }
   ```

5. **Pane.View()のヘッダー拡張** (`internal/ui/pane.go`)
   - ヘッダーを2行に拡張:
     ```
     行1: ディレクトリパス（既存）
     行2: マーク情報 + 空き容量（新規）または ローディングメッセージ
     ```
   - ヘッダー2行目のレンダリング関数を追加:
     ```go
     // renderHeaderLine2 はヘッダー2行目（マーク情報と空き容量）をレンダリング
     func (p *Pane) renderHeaderLine2(diskSpace uint64) string
     ```
   - ローディング中は`p.loadingProgress`を表示、通常時はマーク情報と空き容量を表示
   - レイアウト: `"Marked X/N SIZE"` + スペース + `"SIZE Free"` （右揃え）

6. **Pane.View()の更新** (`internal/ui/pane.go`)
   - 現在のView()メソッドを更新し、ヘッダー2行目を追加
   - ディスク容量はメソッド引数で受け取る（Modelから渡す）
   - visibleLinesの計算を調整（ヘッダーが2行 + ボーダー1行 = 3行）

7. **Model.View()の更新** (`internal/ui/model.go`)
   - Pane.View()呼び出し時にディスク容量を渡す:
     ```go
     leftPaneView := m.leftPane.ViewWithDiskSpace(m.leftDiskSpace)
     rightPaneView := m.rightPane.ViewWithDiskSpace(m.rightDiskSpace)
     ```

**依存関係**:
- Phase 1完了（ディスク容量取得関数）
- Phase 2完了（フォーマット関数）

**テスト**:
- `internal/ui/model_test.go`: ディスク容量更新のテスト
  - 初期化時にディスク容量が取得される
  - 5秒ごとに更新される（モックタイマーを使用）
  - ディレクトリ移動時に更新される
- `internal/ui/pane_test.go`: ヘッダー表示のテスト
  - ヘッダー2行目にマーク情報と空き容量が表示される
  - ローディング中はローディングメッセージが表示される
  - 端末幅に応じてレイアウトが調整される

**推定工数**: 中（2-3日）

---

### Phase 5: ファイルリスト詳細表示の実装

**目標**: 表示モードに応じてファイルエントリを適切にレンダリングする

**作成・修正するファイル**:
- `internal/ui/pane.go` - formatEntry()メソッドの大幅な拡張

**実装ステップ**:

1. **formatEntry()メソッドの再設計** (`internal/ui/pane.go`)
   - 現在のシンプルな実装を表示モード対応に拡張
   - 各表示モードで異なるカラム構成:
     ```go
     // formatEntry はエントリを表示モードに応じてフォーマット
     func (p *Pane) formatEntry(entry fs.FileEntry, isCursor bool) string
     ```

2. **Minimalモードのレンダリング**
   - ファイル名のみを表示
   - シンボリックリンクは `name -> target` 形式
   - リンク先が長い場合は `name -> /very/long/...` と省略
   - ディレクトリは末尾に`/`を付与

3. **Basicモードのレンダリング**
   - カラム構成: `Name` + `Size` + `Timestamp`
   - カラム間のスペース: 2文字
   - レイアウト例:
     ```
     README.md         1.2 KiB  2024-12-01 10:30
     notes.txt         450 B    2024-11-28 15:45
     link -> /target   512 B    2024-12-01 10:30
     ```
   - Size: 右揃え、ディレクトリは`-`、リンク切れは`?`
   - Timestamp: 固定幅16文字、`FormatTimestamp()`を使用

4. **Detailモードのレンダリング**
   - カラム構成: `Name` + `Permissions` + `Owner` + `Group`
   - カラム間のスペース: 2文字
   - レイアウト例:
     ```
     README.md         rw-r--r--  user  staff
     notes.txt         rw-r--r--  user  staff
     link -> /target   rwxrwxrwx  user  staff
     ```
   - Permissions: 固定幅10文字、`FormatPermissions()`を使用
   - Owner/Group: 左揃え、最大幅を動的に決定（長い名前は切り詰め）

5. **カラム幅の動的調整**
   - Phase 2で作成した`CalculateColumnWidths()`を使用
   - ファイル名カラムは可変幅（残りスペースを使用）
   - 長いファイル名は `...` で切り詰め

6. **色とスタイルの適用**
   - カーソル位置: 背景色を変更（既存の実装を維持）
   - ディレクトリ: 青色
   - シンボリックリンク: シアン色
   - リンク切れ: 赤色またはグレーアウト
   - アクティブペイン: 明るい色
   - 非アクティブペイン: 暗い色

**依存関係**:
- Phase 2完了（フォーマット関数）
- Phase 3完了（表示モード管理）

**テスト**:
- `internal/ui/pane_test.go`: ファイルリスト表示のテスト
  - Minimalモードで名前のみ表示
  - Basicモードでサイズとタイムスタンプ表示
  - Detailモードでパーミッション、所有者、グループ表示
  - シンボリックリンクが適切にフォーマットされる
  - リンク切れが視覚的に区別される
  - 長いファイル名とリンク先が適切に切り詰められる

**推定工数**: 中（2-3日）

---

### Phase 6: シンボリックリンク対応の改善

**目標**: シンボリックリンク先への移動とリンク切れの視覚的表現を実装する

**作成・修正するファイル**:
- `internal/ui/pane.go` - EnterDirectory()メソッドの拡張
- `internal/ui/styles.go` - リンク切れ用のスタイル追加

**実装ステップ**:

1. **EnterDirectory()メソッドの拡張** (`internal/ui/pane.go`)
   - 現在の実装（ディレクトリのみ処理）をシンボリックリンク対応に拡張:
     ```go
     func (p *Pane) EnterDirectory() error {
         entry := p.SelectedEntry()
         if entry == nil {
             return nil
         }

         // シンボリックリンクの場合
         if entry.IsSymlink {
             if entry.LinkBroken {
                 return nil  // リンク切れは何もしない
             }

             // リンク先がディレクトリかチェック
             if isDir, err := fs.IsDirectory(entry.LinkTarget); err == nil && isDir {
                 p.path = entry.LinkTarget
                 return p.LoadDirectory()
             }
             return nil  // リンク先がファイルまたはエラー → 何もしない
         }

         // 通常のディレクトリ処理（既存）
         if !entry.IsDir {
             return nil
         }
         // ... 既存のコード
     }
     ```

2. **IsDirectory()ヘルパー関数の追加** (`internal/fs/operations.go` または新規ファイル)
   ```go
   // IsDirectory は指定されたパスがディレクトリかどうかを判定
   func IsDirectory(path string) (bool, error) {
       info, err := os.Stat(path)
       if err != nil {
           return false, err
       }
       return info.IsDir(), nil
   }
   ```

3. **リンク切れのスタイル定義** (`internal/ui/styles.go`)
   ```go
   var (
       // 既存のスタイル...

       brokenLinkStyle = lipgloss.NewStyle().
           Foreground(lipgloss.Color("9")).  // 赤色
           Strikethrough(true)                // 打ち消し線（オプション）

       symlinkStyle = lipgloss.NewStyle().
           Foreground(lipgloss.Color("14"))  // シアン色
   )
   ```

4. **formatEntry()での色適用** (`internal/ui/pane.go`)
   - Phase 5で実装したformatEntry()に色適用ロジックを追加:
     ```go
     if entry.IsSymlink {
         if entry.LinkBroken {
             style = brokenLinkStyle
         } else {
             style = symlinkStyle
         }
     } else if entry.IsDir {
         style = directoryStyle
     }
     ```

5. **シンボリックリンクのサイズ表示** (`internal/ui/pane.go`)
   - Basicモードでのサイズカラム:
     - 通常のシンボリックリンク: リンク先のサイズを表示
     - リンク切れ: `?` を表示
   - Phase 5のformatEntry()内で処理

**依存関係**:
- Phase 1完了（シンボリックリンク情報取得）
- Phase 5完了（ファイルリスト表示）

**テスト**:
- `internal/ui/pane_test.go`: シンボリックリンク操作のテスト
  - ディレクトリへのシンボリックリンクでEnter → リンク先に移動
  - ファイルへのシンボリックリンクでEnter → 何もしない
  - リンク切れでEnter → 何もしない
  - リンク切れが赤色で表示される
  - 正常なシンボリックリンクがシアン色で表示される
- 手動テスト:
  - 実際のシンボリックリンクを作成してナビゲーションをテスト
  - 相対パスと絶対パスのシンボリックリンクをテスト

**推定工数**: 小（1日）

---

### Phase 7: ローディング表示の実装

**目標**: 大規模ディレクトリ読み込み時にローディングフィードバックを提供する

**作成・修正するファイル**:
- `internal/ui/pane.go` - 非同期ディレクトリ読み込みの実装
- `internal/ui/messages.go` - ローディング関連メッセージの追加
- `internal/fs/reader.go` - 進捗報告対応（オプション）

**実装ステップ**:

1. **ローディングメッセージの定義** (`internal/ui/messages.go`)
   ```go
   // directoryLoadStartMsg はディレクトリ読み込み開始を通知
   type directoryLoadStartMsg struct {
       pane *Pane
   }

   // directoryLoadCompleteMsg はディレクトリ読み込み完了を通知
   type directoryLoadCompleteMsg struct {
       pane    *Pane
       entries []fs.FileEntry
       err     error
   }

   // directoryLoadProgressMsg は読み込み進捗を通知（オプション）
   type directoryLoadProgressMsg struct {
       pane       *Pane
       fileCount  int
   }
   ```

2. **LoadDirectoryAsyncコマンドの実装** (`internal/ui/pane.go`)
   ```go
   // LoadDirectoryAsync は非同期でディレクトリを読み込む
   func (p *Pane) LoadDirectoryAsync() tea.Cmd {
       pane := p  // コピー
       return func() tea.Msg {
           entries, err := fs.ReadDirectory(pane.path)
           if err != nil {
               return directoryLoadCompleteMsg{pane: pane, err: err}
           }

           fs.SortEntries(entries)
           return directoryLoadCompleteMsg{
               pane:    pane,
               entries: entries,
               err:     nil,
           }
       }
   }
   ```

3. **EnterDirectory()とMoveToParent()の非同期化** (`internal/ui/pane.go`)
   - 既存の同期的な`LoadDirectory()`呼び出しを削除
   - 代わりに`loading`フラグをtrueに設定し、`LoadDirectoryAsync()`を返す
   - 実際のエントリ更新はModel.Update()で`directoryLoadCompleteMsg`を受信した時に行う

4. **Model.Update()でのローディングメッセージ処理** (`internal/ui/model.go`)
   ```go
   case directoryLoadCompleteMsg:
       // メッセージの対象ペインを特定
       targetPane := m.getTargetPane(msg.pane)

       if msg.err != nil {
           // エラーダイアログを表示
           m.dialog = NewErrorDialog(fmt.Sprintf("Failed to read directory: %v", msg.err))
           targetPane.loading = false
       } else {
           // エントリを更新
           targetPane.entries = msg.entries
           targetPane.cursor = 0
           targetPane.scrollOffset = 0
           targetPane.loading = false
           targetPane.loadingProgress = ""
       }
       return m, nil
   ```

5. **getTargetPane()ヘルパーメソッドの実装** (`internal/ui/model.go`)
   ```go
   // getTargetPane はペインポインタから対応するペインを取得
   // （メモリアドレスではなくパスで比較する方が安全）
   func (m *Model) getTargetPane(pane *Pane) *Pane {
       if pane.path == m.leftPane.path {
           return m.leftPane
       }
       return m.rightPane
   }
   ```

6. **ローディング中のUI表示** (`internal/ui/pane.go`)
   - Phase 4で実装したrenderHeaderLine2()でローディングメッセージを表示:
     ```go
     func (p *Pane) renderHeaderLine2(diskSpace uint64) string {
         if p.loading {
             return p.loadingProgress  // "Loading directory..."
         }
         // 通常のマーク情報と空き容量を表示
     }
     ```
   - ローディング中はファイルリストを空にするか、"Loading..."を表示

7. **進捗報告の実装（オプション）**
   - ファイル数が非常に多い場合、定期的に進捗を報告
   - ReadDirectory()内で100ファイルごとにチャネルで進捗送信
   - Bubble Teaのサブスクリプション機能を使用して進捗を受信
   - 複雑さが増すため、まずは基本的なローディング表示から実装

**依存関係**:
- Phase 4完了（ヘッダー2行目の実装）

**テスト**:
- `internal/ui/model_test.go`: 非同期ローディングのテスト
  - ディレクトリ読み込み開始時にloadingフラグがtrueになる
  - 読み込み完了時にエントリが更新される
  - エラー時にエラーダイアログが表示される
- 手動テスト:
  - 大規模ディレクトリ（数千ファイル）を開いてローディング表示を確認
  - ローディング中にキー入力が応答するか確認（UIブロックしない）

**推定工数**: 中（2-3日）

---

### Phase 8: キーバインディングの追加

**目標**: `i`キーで表示モードを切り替える機能を実装する

**作成・修正するファイル**:
- `internal/ui/keys.go` - キー定義の追加
- `internal/ui/model.go` - `i`キーのハンドリング追加

**実装ステップ**:

1. **キー定義の追加** (`internal/ui/keys.go`)
   ```go
   const (
       // 既存のキー定義...

       KeyToggleInfo = "i"  // 表示モード切り替え
   )
   ```

2. **Model.Update()でのキーハンドリング** (`internal/ui/model.go`)
   ```go
   case KeyToggleInfo:
       activePane := m.getActivePane()
       if activePane.CanToggleMode() {
           activePane.ToggleDisplayMode()
       }
       // CanToggleMode() == false の場合は何もしない（無効化）
       return m, nil
   ```

3. **ステータスバーのヒント更新** (`internal/ui/model.go`)
   - renderStatusBar()内のキーヒント文字列を更新:
     ```go
     hints := "i:info ?:help q:quit"
     ```
   - または、動的にヒントを変更（iキーが有効な場合のみ表示）:
     ```go
     hints := "?:help q:quit"
     if m.getActivePane().CanToggleMode() {
         hints = "i:info " + hints
     }
     ```

**依存関係**:
- Phase 3完了（表示モード管理）

**テスト**:
- `internal/ui/model_test.go`: iキーのテスト
  - 端末幅が広い時、iキーでBasic⇔Detailを切り替え
  - 端末幅が狭い時、iキーが無効（何も起こらない）
  - 左右のペインで独立して切り替え可能

**推定工数**: 小（0.5日）

---

### Phase 9: 統合テストとバグ修正

**目標**: 全機能を統合し、エッジケースのテストとバグ修正を行う

**作成・修正するファイル**:
- 各種テストファイルの追加・修正
- バグ修正のためのコード修正

**実装ステップ**:

1. **統合テストシナリオの実行**
   - SPEC.mdのTest Scenarios（Test 1〜Test 60）を手動で確認
   - 自動化可能なテストはテストコードとして追加

2. **エッジケースのテスト**
   - 空のディレクトリ
   - `.`のみのディレクトリ
   - 削除されたユーザー/グループのファイル
   - 特殊文字を含むファイル名
   - シンボリックリンクのチェーン
   - 循環シンボリックリンク
   - 10,000+ファイルのディレクトリ
   - 非常に狭い端末（20カラム）
   - 非常に広い端末（200カラム）

3. **パフォーマンステスト**
   - 1,000ファイルのディレクトリで応答性を確認
   - 10,000ファイルのディレクトリでローディング表示を確認
   - ディスク容量取得のオーバーヘッド測定

4. **クロスプラットフォームテスト**
   - Linux: 完全な機能テスト
   - macOS: 完全な機能テスト
   - Windows: 制限的な機能テスト（所有者情報など）

5. **バグ修正とリファクタリング**
   - 発見されたバグを修正
   - コードの重複を削減
   - パフォーマンスボトルネックを最適化

6. **ドキュメント更新**
   - README.mdの更新（新機能の説明）
   - CHANGELOG.mdの作成（変更履歴）

**依存関係**:
- Phase 1〜8全て完了

**テスト**:
- 全テストスイートの実行: `go test ./...`
- カバレッジ測定: `go test -cover ./...`
- レースコンディション検出: `go test -race ./...`

**推定工数**: 中（2-3日）

---

## ファイル構成

実装後のファイル構成:

```
duofm/
├── cmd/duofm/
│   └── main.go                 # エントリーポイント（既存）
├── internal/
│   ├── fs/
│   │   ├── types.go            # FileEntry構造体（拡張）
│   │   ├── reader.go           # ReadDirectory（拡張）
│   │   ├── reader_test.go      # ReadDirectoryテスト
│   │   ├── sort.go             # ソート処理（既存）
│   │   ├── sort_test.go        # ソートテスト（既存）
│   │   ├── operations.go       # ファイル操作（既存）
│   │   ├── operations_test.go  # ファイル操作テスト（既存）
│   │   ├── owner.go            # 所有者・グループ取得（新規）
│   │   ├── owner_test.go       # 所有者取得テスト（新規）
│   │   ├── symlink.go          # シンボリックリンク処理（新規）
│   │   ├── symlink_test.go     # シンボリックリンクテスト（新規）
│   │   ├── diskspace.go        # ディスク容量取得（新規）
│   │   └── diskspace_test.go   # ディスク容量テスト（新規）
│   └── ui/
│       ├── model.go            # Modelとメインロジック（拡張）
│       ├── model_test.go       # Modelテスト（拡張）
│       ├── pane.go             # Pane表示とロジック（大幅拡張）
│       ├── pane_test.go        # Paneテスト（拡張）
│       ├── display_mode.go     # 表示モード定義（新規）
│       ├── format.go           # フォーマット関数（新規）
│       ├── format_test.go      # フォーマットテスト（新規）
│       ├── messages.go         # Bubble Teaメッセージ（新規）
│       ├── keys.go             # キー定義（拡張）
│       ├── styles.go           # スタイル定義（拡張）
│       ├── dialog.go           # ダイアログ基底（既存）
│       ├── dialog_test.go      # ダイアログテスト（既存）
│       ├── error_dialog.go     # エラーダイアログ（既存）
│       ├── help_dialog.go      # ヘルプダイアログ（既存）
│       └── confirm_dialog.go   # 確認ダイアログ（既存）
├── go.mod                      # 依存関係定義（既存）
├── go.sum                      # 依存関係チェックサム（既存）
├── Makefile                    # ビルド自動化（既存）
└── doc/
    └── tasks/
        └── ui-enhancement/
            ├── SPEC.md                 # 技術仕様（既存）
            ├── 要件定義書.md           # 要件定義（既存）
            └── IMPLEMENTATION.md       # 実装計画（本ファイル）
```

### 新規作成ファイル（7ファイル + テスト7ファイル）
- `internal/fs/owner.go` + `owner_test.go`
- `internal/fs/symlink.go` + `symlink_test.go`
- `internal/fs/diskspace.go` + `diskspace_test.go`
- `internal/ui/display_mode.go`
- `internal/ui/format.go` + `format_test.go`
- `internal/ui/messages.go`

### 大幅修正ファイル（3ファイル）
- `internal/fs/types.go` - FileEntry拡張
- `internal/ui/pane.go` - View/formatEntry/ヘッダー実装
- `internal/ui/model.go` - ディスク容量管理、メッセージハンドリング

### 小規模修正ファイル（4ファイル）
- `internal/fs/reader.go` - 拡張情報取得
- `internal/ui/keys.go` - iキー追加
- `internal/ui/styles.go` - シンボリックリンク色追加
- 各テストファイル - テストケース追加

## テスト戦略

### ユニットテスト

#### ファイルシステム層（internal/fs）
- **owner.go**: 所有者・グループ名の取得
  - 正常系: 既存ファイルの所有者取得成功
  - 異常系: 存在しないファイル、権限エラー
  - プラットフォーム差異: Unix/Linux/macOS vs Windows

- **symlink.go**: シンボリックリンク処理
  - 正常系: 有効なシンボリックリンクの情報取得
  - リンク切れ検出
  - 相対パス→絶対パス変換
  - シンボリックリンクのチェーン（link -> link -> file）

- **diskspace.go**: ディスク容量取得
  - 正常系: ルートディレクトリ、ホームディレクトリの容量取得
  - 異常系: 存在しないパス、権限エラー
  - プラットフォーム差異: Unix vs Windows

#### UI層（internal/ui）
- **format.go**: フォーマット関数
  - FormatSize: 各単位の変換（0 B, 512 B, 1.0 KiB, 1.5 MiB, 2.3 GiB）
  - FormatTimestamp: ISO 8601フォーマット
  - FormatPermissions: Unix形式パーミッション
  - テーブル駆動テストで網羅的にテスト

- **display_mode.go**: 表示モード
  - MinRequiredWidth()の各モードでの戻り値
  - String()の文字列表現

### 統合テスト

#### Model/Paneの統合（internal/ui）
- ディスク容量の定期更新（5秒ごと）
- ディレクトリ移動時のディスク容量更新
- 表示モード切り替え（iキー）
- 端末リサイズ時のモード自動調整
- 非同期ディレクトリ読み込み
- ローディング表示の開始・完了

### 手動テストチェックリスト

#### ヘッダー表示
- [ ] ディレクトリパスが正しく表示される（ホームは`~`表示）
- [ ] ヘッダー2行目にマーク情報と空き容量が表示される
- [ ] 空き容量が適切な単位で表示される（GiB, TiBなど）
- [ ] 空き容量が5秒ごとに更新される
- [ ] ローディング中はヘッダー2行目に"Loading..."が表示される

#### 表示モード切り替え
- [ ] 端末幅60カラム以上でiキーが有効
- [ ] iキーでBasic⇔Detailを切り替え
- [ ] 端末幅が狭くなると自動的にMinimalモードになる
- [ ] 端末幅が広くなると最後のモードに戻る
- [ ] 左右のペインで独立して切り替え可能
- [ ] Minimalモード時にiキーが無効（何も起こらない）

#### ファイル情報表示
- [ ] Basicモード: サイズとタイムスタンプが表示される
- [ ] Detailモード: パーミッション、所有者、グループが表示される
- [ ] タイムスタンプが`2024-12-17 22:28`形式
- [ ] ディレクトリのサイズが`-`で表示
- [ ] ファイルサイズが適切な単位で表示
- [ ] パーミッションがUnix形式（rwxr-xr-x）

#### シンボリックリンク
- [ ] シンボリックリンクが`name -> target`形式で表示
- [ ] ディレクトリへのリンクでEnterキー → 移動成功
- [ ] ファイルへのリンクでEnterキー → 何も起こらない
- [ ] リンク切れが赤色で表示
- [ ] リンク切れのサイズ欄に`?`が表示
- [ ] リンク切れでEnterキー → 何も起こらない
- [ ] 長いリンク先パスが`...`で省略

#### ローディング表示
- [ ] 大規模ディレクトリでローディング表示が出る
- [ ] ローディング中もUIが応答する（キー入力可能）
- [ ] ローディング完了後、通常表示に戻る

#### レイアウトとレスポンシブ
- [ ] 端末リサイズに追従
- [ ] 狭い端末（30カラム）でも表示が崩れない
- [ ] 広い端末（150カラム）でスペースが有効活用される
- [ ] 長いファイル名が適切に切り詰められる

#### エッジケース
- [ ] 空のディレクトリを開ける
- [ ] 削除されたユーザーのファイルで"unknown"表示
- [ ] 特殊文字を含むファイル名が正しく表示
- [ ] 10,000+ファイルのディレクトリで応答性を維持

## 依存関係

### 外部ライブラリ

既存の依存関係をそのまま使用:

```go
require (
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    golang.org/x/sys v0.36.0  // ディスク容量取得に使用（既に依存関係にある）
)
```

追加の外部ライブラリは不要。

### 内部依存関係

実装順序の依存関係:

```
Phase 1 (データ構造拡張)
    ↓
Phase 2 (フォーマット関数) ← Phase 1に依存
    ↓
Phase 3 (表示モード管理) ← Phase 2に依存
    ↓
Phase 4 (ヘッダー拡張) ← Phase 1, 2に依存
    ↓
Phase 5 (ファイルリスト表示) ← Phase 2, 3に依存
    ↓
Phase 6 (シンボリックリンク対応) ← Phase 1, 5に依存
    ↓
Phase 7 (ローディング表示) ← Phase 4に依存
    ↓
Phase 8 (キーバインディング) ← Phase 3に依存
    ↓
Phase 9 (統合テスト) ← Phase 1〜8全てに依存
```

**並列実装可能な部分**:
- Phase 4（ヘッダー拡張）とPhase 5（ファイルリスト表示）は一部並列可能
- Phase 6（シンボリックリンク対応）とPhase 7（ローディング表示）は並列可能

## リスク評価

### 技術的リスク

#### リスク1: プラットフォーム依存のファイル情報取得
- **説明**: ファイル所有者やディスク容量取得がUnix/Windows/macOSで異なるAPIを使用
- **影響度**: 中（Windowsで一部機能が制限される可能性）
- **軽減策**:
  - ビルドタグを使用してプラットフォームごとに実装を分離
  - Windowsでは所有者情報を"N/A"で表示するなど、graceful degradation
  - 各プラットフォームでのテスト

#### リスク2: 大規模ディレクトリのパフォーマンス
- **説明**: 数万ファイルのディレクトリで所有者情報やシンボリックリンク解決が遅延
- **影響度**: 中（ユーザー体験の低下）
- **軽減策**:
  - 非同期読み込み（Phase 7）で対応
  - ローディング表示でフィードバック提供
  - 必要に応じてgoroutineで並列処理（将来の最適化）

#### リスク3: 端末幅計算の複雑さ
- **説明**: 各表示モードで必要な最小幅の計算が複雑でバグの温床になる可能性
- **影響度**: 低（UIが崩れる程度）
- **軽減策**:
  - CalculateColumnWidths()を徹底的にテスト
  - マジックナンバーを定数化してメンテナンス性向上
  - 端末幅が極端に狭い場合でも最低限の表示を保証

#### リスク4: Bubble Teaの非同期メッセージング
- **説明**: goroutineとBubble Teaのメッセージングの組み合わせでレースコンディションやメモリリーク
- **影響度**: 高（アプリケーションクラッシュ）
- **軽減策**:
  - `go test -race`でレースコンディション検出
  - Bubble Teaのベストプラクティスに従う（状態はコピーして渡す）
  - メッセージハンドラーでエラーハンドリングを徹底

#### リスク5: シンボリックリンクの循環参照
- **説明**: 循環シンボリックリンクで無限ループやスタックオーバーフロー
- **影響度**: 中（特定のディレクトリで動作不能）
- **軽減策**:
  - シンボリックリンク解決時にos.Stat()を使用（内部で循環検出）
  - エラーをキャッチしてリンク切れとして扱う
  - 深いチェーン（5階層以上）は警告（将来実装）

## パフォーマンス考慮事項

### ディスク容量の取得
- **キャッシュ**: 5秒間キャッシュして頻繁なディスクアクセスを回避
- **非同期**: ディスク容量取得をBubble Teaのコマンドで非同期実行（UIをブロックしない）

### ファイル情報の取得
- **ワンパス**: ReadDirectory()で一度にすべての情報を取得（表示モード切り替え時に再取得しない）
- **エラースキップ**: 個別ファイルの情報取得失敗は無視して続行（全体の読み込み失敗を防ぐ）

### 大規模ディレクトリ
- **非同期読み込み**: Phase 7でgoroutineを使用
- **ローディングフィードバック**: 0.5秒以上かかる場合にローディング表示
- **将来の最適化**: 並列化、ページネーション、仮想スクロール

### UI描画
- **Lip Gloss**: 高速なスタイリングライブラリを使用
- **スクロール**: 表示範囲のみレンダリング（既存の実装を維持）

### メモリ使用量
- **エントリのキャッシュ**: 現在のディレクトリのエントリのみメモリに保持
- **文字列の最適化**: 長い文字列は切り詰めてメモリを節約

## セキュリティ考慮事項

### パス操作
- **パストラバーサル**: `filepath.Clean()`と`filepath.Join()`を使用してパストラバーサル攻撃を防ぐ
- **シンボリックリンク**: `os.Stat()`（シンボリックリンクを解決）と`os.Lstat()`（解決しない）を適切に使い分け

### ファイルアクセス権限
- **権限チェック**: アクセス権限のないファイルやディレクトリはエラーハンドリング
- **エラー表示**: 権限エラーをユーザーに明確に伝える

### ディスク容量情報
- **機密情報**: ディスク容量は一般的に機密性は低いが、システム情報を露出しないように注意
- **エラーハンドリング**: 取得失敗時は0を表示するか非表示（クラッシュしない）

### 入力検証
- **ファイル名**: ファイルシステムから取得した名前をそのまま表示（ユーザー入力ではない）
- **特殊文字**: ターミナル制御文字を適切にエスケープ（Lip Glossが処理）

## 未解決の質問

### 技術的な質問
- [ ] **ディレクトリサイズの計算**: ディレクトリのサイズを計算すべきか（duコマンドのように）？
  - 現状: `-`で表示
  - 代替案: バックグラウンドで計算、またはオプションで有効化

- [ ] **所有者・グループの幅**: 所有者名が非常に長い場合の最大幅は？
  - 提案: 10文字で切り詰め、またはユーザー設定で変更可能

- [ ] **ローディングの閾値**: 何ファイル以上でローディング表示を出すか？
  - 現状: 0.5秒以上かかる場合（ファイル数ではなく時間ベース）
  - 代替案: 1000ファイル以上、または設定可能に

- [ ] **シンボリックリンクのチェーン深さ**: 何階層までシンボリックリンクを辿るか？
  - 現状: os.Stat()に任せる（OSのデフォルト制限）
  - 代替案: 独自に深さ制限（5階層など）

### UI/UX関連の質問
- [ ] **表示モードの永続化**: アプリケーション再起動後も表示モードを保持すべきか？
  - 現状: 毎回Basicモードで起動
  - 代替案: 設定ファイルに保存（Phase 3で実装）

- [ ] **モード表示**: ステータスバーに現在の表示モードを表示すべきか？
  - 現状: 表示しない
  - 代替案: "Mode: Basic"などと表示

- [ ] **リンク切れの操作**: リンク切れをEnterで押した時にエラーダイアログを出すべきか？
  - 現状: 何もしない（サイレント）
  - 代替案: "Link target not found"ダイアログ表示

- [ ] **カラーテーマ**: シンボリックリンクやリンク切れの色をカスタマイズ可能にすべきか？
  - 現状: 固定色（シアン、赤）
  - 代替案: 設定ファイルで変更可能（Phase 3で実装）

### パフォーマンス関連の質問
- [ ] **並列化**: ファイル情報取得を並列化すべきか？
  - 現状: 逐次処理
  - 代替案: workerプールで並列取得（複雑さとのトレードオフ）

- [ ] **ページネーション**: 数万ファイルのディレクトリでページネーションすべきか？
  - 現状: 全ファイルを一度に読み込み
  - 代替案: 100ファイルずつ読み込み、スクロール時に追加読み込み

## 将来の拡張

### Phase 2: マーク機能（実装計画外）
- Spaceキーでファイルをマーク/アンマーク
- マークされたファイル数と合計サイズをヘッダーに表示
- マークされたファイルの一括操作（コピー、移動、削除）

### Phase 3: 設定ファイル対応（実装計画外）
- 設定ファイルの読み込み（YAML、TOMLなど）
- 空き容量リフレッシュレートの設定
- デフォルト表示モードの設定
- カラーテーマの設定
- キーバインディングのカスタマイズ

### 短期的な拡張（Post-Implementation）
1. **ソート機能**: 名前、サイズ、日付、拡張子でソート
2. **検索機能**: ファイル名のインクリメンタルサーチ
3. **ブックマーク**: よく使うディレクトリをブックマーク
4. **履歴**: ディレクトリ移動履歴のナビゲーション

### 長期的な拡張
1. **ファイルプレビュー**: テキストファイルや画像のプレビュー表示
2. **タブ機能**: 複数のディレクトリをタブで管理
3. **プラグインシステム**: カスタムフォーマッターやアクションの追加
4. **拡張属性**: xattrやACLの表示
5. **ネットワークファイルシステム最適化**: NFS、SMBでの高速化

## 参考資料

### 外部資料
- [Total Commander](https://www.ghisler.com/) - UIリファレンス
- [Midnight Commander](https://midnight-commander.org/) - 機能リファレンス
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea) - TUIフレームワーク
- [Lip Gloss Documentation](https://github.com/charmbracelet/lipgloss) - スタイリング
- [Go os package](https://pkg.go.dev/os) - ファイル操作API
- [Go syscall package](https://pkg.go.dev/syscall) - ディスク容量、所有者情報
- [Go os/user package](https://pkg.go.dev/os/user) - ユーザー・グループ情報

### プロジェクト内資料
- `doc/tasks/ui-enhancement/SPEC.md` - 技術仕様書
- `doc/tasks/ui-enhancement/要件定義書.md` - 要件定義書
- `CLAUDE.md` - プロジェクト概要とアーキテクチャ
- `doc/CONTRIBUTING.md` - 貢献ガイドライン（未作成の場合は作成推奨）

### Goのベストプラクティス
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

---

## 実装開始前のチェックリスト

実装を開始する前に、以下を確認してください:

- [ ] SPEC.mdと要件定義書.mdを読み、理解した
- [ ] CLAUDE.mdを読み、プロジェクト構造を理解した
- [ ] 既存のコードベースを確認し、アーキテクチャを把握した
- [ ] Go 1.21以上がインストールされている
- [ ] Bubble TeaとLip Glossのドキュメントに目を通した
- [ ] 開発環境でduofmをビルド・実行できることを確認した
- [ ] `go test ./...`で既存のテストが通ることを確認した
- [ ] この実装計画を読み、フェーズごとの目標を理解した

## 実装完了の定義（Definition of Done）

以下をすべて満たした場合、本実装は完了とみなします:

- [ ] Phase 1〜9のすべての実装が完了している
- [ ] すべてのユニットテストが成功する（`go test ./...`）
- [ ] テストカバレッジが70%以上
- [ ] レースコンディションが検出されない（`go test -race ./...`）
- [ ] SPEC.mdの Success Criteria がすべて満たされている
- [ ] 手動テストチェックリストのすべての項目が確認されている
- [ ] Linux、macOS、Windows（可能であれば）でテスト済み
- [ ] 1,000ファイルのディレクトリで応答性が保たれている
- [ ] 10,000ファイルのディレクトリでローディング表示が機能している
- [ ] コードが`gofmt`でフォーマットされている
- [ ] `go vet ./...`でエラーが出ない
- [ ] README.mdが新機能を反映して更新されている
- [ ] CHANGELOGまたはリリースノートが作成されている

---

## まとめ

本実装計画は、duofmのUIブラッシュアップを9つのフェーズに分けて段階的に実装するアプローチを提供します。各フェーズは独立してテスト可能で、依存関係が明確になっています。

**総推定工数**: 15〜20日（1人のフルタイム開発者）

**重要なポイント**:
1. **段階的な実装**: 各フェーズを完了させてから次に進む
2. **テスト駆動**: 各フェーズで十分なテストを実装
3. **レビューポイント**: Phase 3、Phase 6、Phase 9で全体レビュー
4. **柔軟性**: 実装中に発見された問題に応じて計画を調整

実装中に疑問が生じた場合は、未解決の質問セクションを参照し、必要に応じてSPEC.mdの作成者やチームメンバーと相談してください。
