# duofm Build Scripts

このディレクトリには、duofmプロジェクトのビルドとパッケージング関連のスクリプトが含まれています。

## スクリプト一覧

### build-dpkg.sh

Debian/Ubuntuシステム用の`.deb`パッケージを作成するスクリプトです。

#### 使用方法

```bash
# Makefileターゲット経由（推奨）
make dpkg

# 直接実行
bash scripts/build-dpkg.sh
```

#### 機能

- Goバイナリの自動ビルド
- 適切なDebianパッケージ構造の作成
- メタデータ（control, changelog）の自動生成
- インストール/削除スクリプトの作成
- パッケージ情報の表示

#### 出力

- パッケージファイル: `duofm_<version>_<arch>.deb`
- 一時ビルドディレクトリ: `build/dpkg/` (自動削除)

#### 必要な環境

- `dpkg-deb`: Debianパッケージ作成ツール
  ```bash
  sudo apt-get install dpkg
  ```
- `make`: ビルドツール
- `go`: Goコンパイラ（1.21以降）

#### バージョン番号の決定

バージョン番号は以下の優先順位で決定されます:

1. gitタグ（`git describe --tags --always`）
2. デフォルト値: `0.1.0`

リリース時は適切なgitタグを付けることを推奨します:

```bash
git tag v0.2.0
make dpkg
```

#### アーキテクチャ対応

以下のアーキテクチャに対応しています:

- `x86_64` → `amd64` (Debian形式)
- `aarch64` → `arm64`
- `armv7l` → `armhf`
- `i686` → `i386`

#### パッケージ構造

作成されるパッケージには以下が含まれます:

```
duofm_<version>_<arch>/
├── DEBIAN/
│   ├── control          # パッケージメタデータ
│   ├── postinst         # インストール後スクリプト
│   ├── prerm            # 削除前スクリプト
│   └── postrm           # 削除後スクリプト
└── usr/
    ├── bin/
    │   └── duofm        # 実行ファイル
    └── share/
        └── doc/
            └── duofm/
                ├── README.md
                ├── copyright (LICENSE)
                └── changelog.gz
```

#### インストール

```bash
# パッケージのインストール
sudo dpkg -i duofm_0.1.0_amd64.deb

# インストール内容の確認
dpkg -L duofm

# アンインストール
sudo dpkg -r duofm

# 完全削除（設定ファイルも削除）
sudo dpkg -P duofm
```

#### トラブルシューティング

##### dpkg-deb: command not found

```bash
sudo apt-get install dpkg
```

##### ビルドエラー

```bash
# 依存関係の更新
make deps

# クリーンビルド
make clean && make build
```

##### パーミッションエラー

スクリプトに実行権限を付与:

```bash
chmod +x scripts/build-dpkg.sh
```

## Makefileターゲット

### dpkg

dpkgパッケージをビルドします。

```bash
make dpkg
```

### clean-dpkg

dpkg関連のビルド成果物を削除します。

```bash
make clean-dpkg
```

## ベストプラクティス

### リリース前のチェックリスト

- [ ] gitタグでバージョンを設定
- [ ] `make test`でテストが通ることを確認
- [ ] `make build`でビルドが成功することを確認
- [ ] `make dpkg`でパッケージを作成
- [ ] 作成したパッケージをテスト環境でインストール
- [ ] 動作確認
- [ ] パッケージをリポジトリまたはリリースページに公開

### バージョン管理

セマンティックバージョニングを使用:

- `v1.0.0`: メジャーバージョン（互換性のない変更）
- `v1.1.0`: マイナーバージョン（後方互換性のある機能追加）
- `v1.1.1`: パッチバージョン（バグフィックス）

### カスタマイズ

メンテナー情報やパッケージ説明をカスタマイズする場合は、
`scripts/build-dpkg.sh`内の以下の変数を編集してください:

```bash
PROJECT_NAME="duofm"
MAINTAINER="Your Name <your.email@example.com>"
```

また、DEBIANディレクトリ内の`control`ファイル生成部分で
説明文をカスタマイズできます。

## 参考資料

- [Debian Package Management](https://www.debian.org/doc/manuals/debian-faq/pkg-basics.en.html)
- [dpkg-deb Manual](https://man7.org/linux/man-pages/man1/dpkg-deb.1.html)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Semantic Versioning](https://semver.org/)
