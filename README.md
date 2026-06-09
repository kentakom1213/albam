# albam

`albam` は，ローカルの画像ディレクトリからセルフホスト型の写真アルバムを作るためのツールです．

画像ファイルを `albums/` に置き，`albam index` でアルバム情報を作成し，`albam serve` でブラウザから見られるアルバムサイトとして公開できます．

Go 製のサーバーと Astro 製のテーマを組み合わせ，1 つの `albam` コマンドで次の処理を扱います．

```txt
- アルバムプロジェクトの作成
- 画像ディレクトリの読み取り
- アルバム・写真メタデータの保存
- テーマのビルド
- Web UI，JSON API，画像ファイルの配信
```

Hugo のように，ローカルのファイルを元にサイトを構築し，そのまま Go 製サーバーとして起動できる構成を目指しています．

## 特徴

`albam` は，次のような用途を想定しています．

- 手元の写真ディレクトリをそのままアルバム化する
- 写真本体はファイルシステムに置いたまま，メタデータだけを SQLite に保存する
- 自分のサーバーや Raspberry Pi などで写真アルバムを公開する
- Astro 製テーマを差し替えて，見た目をカスタマイズする
- Caddy，nginx，Cloudflare Tunnel などの背後で運用する

画像ファイルは DB に保存しません．`albam` は画像のパスやアルバム情報を SQLite に保存し，必要に応じて画像ファイルを配信します．

## インストール

Go が入っている環境では，次のコマンドでインストールできます．

```sh
go install github.com/kentakom1213/albam/cmd/albam@latest
```

特定バージョンを指定する場合は，次のようにします．

```sh
go install github.com/kentakom1213/albam/cmd/albam@v0.1.0
```

インストールできたか確認します．

```sh
albam version
```

詳細情報を表示する場合は，次を使います．

```sh
albam version --verbose
```

ソースコードから手元でビルドする場合は，次のようにします．

```sh
git clone https://github.com/kentakom1213/albam.git
cd albam
go build -o bin/albam ./cmd/albam
./bin/albam version
```

## はじめてのアルバムを作る

新しいアルバムプロジェクトを作成します．

```sh
albam init my-album
cd my-album
```

`albam init` は，次のようなプロジェクト構成を作成します．

```txt
my-album/
├── albam.toml
├── albums/
│   └── example/
│       └── sample.png
├── themes/
│   └── default/
│       └── theme.toml
└── .albam/
```

写真は `albums/` 以下に置きます．たとえば，次のような構成にできます．

```txt
albums/
├── weekend-trip/
│   ├── IMG_001.jpg
│   └── IMG_002.jpg
└── coffee-time/
    └── IMG_101.jpg
```

`albam` は，画像ファイルの親ディレクトリをアルバムとして扱います．

```txt
albums/weekend-trip/IMG_001.jpg
  -> album: weekend-trip
```

画像を追加したら，インデックスを作成します．

```sh
albam index
```

テーマをビルドします．

```sh
cd themes/default
pnpm install
cd ../..
albam build
```

サーバーを起動します．

```sh
albam serve
```

デフォルトでは，次のアドレスで開けます．

```txt
http://127.0.0.1:8080
```

## 基本的な流れ

通常は，次の順番で使います．

```sh
albam init my-album
cd my-album

# albums/ に画像を追加する
albam index

# Web UI をビルドする
cd themes/default
pnpm install
cd ../..
albam build

# アルバムサイトを起動する
albam serve
```

写真を追加・削除したときは，もう一度 `albam index` を実行します．

```sh
albam index
```

テーマを変更したときは，もう一度 `albam build` を実行します．

```sh
albam build
```

## ディレクトリ構成

標準的なプロジェクト構成は次の通りです．

```txt
my-album/
├── albam.toml
├── albums/
│   ├── weekend-trip/
│   │   ├── IMG_001.jpg
│   │   └── IMG_002.jpg
│   └── coffee-time/
│       └── IMG_101.jpg
├── themes/
│   └── default/
│       ├── theme.toml
│       ├── package.json
│       ├── astro.config.mjs
│       └── src/
└── .albam/
    ├── public/
    ├── cache/
    └── db.sqlite
```

| path                          | 説明                        |
| ----------------------------- | --------------------------- |
| `albam.toml`                  | プロジェクト設定            |
| `albums/`                     | 元画像ディレクトリ          |
| `themes/default/`             | Astro 製テーマ              |
| `themes/default/theme.toml`   | テーマ設定                  |
| `themes/default/package.json` | テーマの npm パッケージ設定 |
| `.albam/public/`              | ビルド済みテーマ            |
| `.albam/cache/`               | 画像キャッシュ              |
| `.albam/db.sqlite`            | アルバム・写真メタデータ    |

## 設定ファイル

プロジェクト設定ファイルは，デフォルトでは `albam.toml` です．

```toml
title = "My Albums"

[server]
host = "127.0.0.1"
port = 8080

[media]
source_dir = "albums"
cache_dir = ".albam/cache"
allow_original_download = false

[privacy]
map_enabled = false
expose_gps = false
location_precision = "hidden"

[database]
path = ".albam/db.sqlite"

[build]
out_dir = ".albam/public"

[theme]
name = "default"
dir = "themes/default"
```

`albam.toml` が存在しない場合は，デフォルト値で動作します．

別の設定ファイルを使う場合は，`--config` で指定します．

```sh
albam index --config albam.toml
albam build --config albam.toml
albam serve --config albam.toml
```

## テーマ設定

デフォルトテーマは `themes/default/` に配置されています．

```txt
themes/default/
├── theme.toml
├── package.json
├── astro.config.mjs
└── src/
```

`theme.toml` では，テーマの表示名，アクセントカラー，トップページの文言，グリッド列数，ヘッダー，フッター，favicon などを設定できます．

```toml
name = "default"
title = "albam"

[params]
brand = "albam"
accent = "coral"
home_title = "My Albums"
home_eyebrow = "SELF-HOSTED PHOTO ALBUM"
home_description = "a simple folder-based album"
album_grid_columns = 4
photo_grid_columns = 5

[params.nav]
albums = "Albums"

[params.header]
enabled = true

[params.footer]
text = "© 2026 powell"
powered_by = true

[params.favicon]
href = "/favicon.svg"
type = "image/svg+xml"
```

主な設定項目は次の通りです．

| key                         | 説明                                  |
| --------------------------- | ------------------------------------- |
| `name`                      | テーマ名                              |
| `title`                     | サイトタイトル                        |
| `params.brand`              | ヘッダー左上に表示するブランド名      |
| `params.accent`             | アクセントカラー名                    |
| `params.home_title`         | トップページの大きなタイトル          |
| `params.home_eyebrow`       | トップページタイトル上の小さなラベル  |
| `params.home_description`   | トップページの説明文                  |
| `params.album_grid_columns` | トップページのアルバムグリッド列数    |
| `params.photo_grid_columns` | アルバム詳細ページの写真グリッド列数  |
| `params.nav.albums`         | ナビゲーションのアルバム一覧ラベル    |
| `params.header.enabled`     | ヘッダーを表示するかどうか            |
| `params.footer.text`        | フッターに表示するテキスト            |
| `params.footer.powered_by`  | `Powered by albam` を表示するかどうか |
| `params.favicon.href`       | favicon のパス                        |
| `params.favicon.type`       | favicon の MIME type                  |

`params.accent` には，次の値を指定できます．

```txt
pink
coral
mint
blue
lavender
lemon
red
beige
```

たとえば，アクセントカラーを青系にする場合は，次のようにします．

```toml
[params]
accent = "blue"
```

ブランド名を表示したくない場合は，`brand` を空文字にします．

```toml
[params]
brand = ""
```

ヘッダー自体を非表示にする場合は，次のようにします．

```toml
[params.header]
enabled = false
```

フッターの `Powered by albam` を非表示にする場合は，次のようにします．

```toml
[params.footer]
powered_by = false
```

favicon を変更する場合は，`href` と `type` を指定します．

```toml
[params.favicon]
href = "/favicon.svg"
type = "image/svg+xml"
```

## テーマの開発とビルド

`themes/default/package.json` には，Astro テーマ用の開発・ビルドコマンドが定義されています．

```json
{
  "scripts": {
    "dev": "astro dev --host 0.0.0.0",
    "build": "astro build",
    "preview": "astro preview --host 0.0.0.0"
  }
}
```

テーマを開発する場合は，次のようにします．

```sh
cd themes/default
pnpm install
pnpm dev
```

テーマをビルドする場合は，通常はプロジェクトルートから `albam build` を実行します．

```sh
albam build
```

手動で Astro のビルドだけを確認する場合は，テーマディレクトリで次を実行します．

```sh
cd themes/default
pnpm build
```

`astro.config.mjs` では `output: "static"` が指定されており，デフォルトテーマは静的ファイルとしてビルドされます．ビルド後の成果物は，`albam build` によって `.albam/public/` に配置され，`albam serve` から配信されます．

テーマを切り替える場合は，`albam.toml` の `[theme]` セクションを変更します．

```toml
[theme]
name = "my-theme"
dir = "themes/my-theme"
```

## 対応画像形式

現在，次の拡張子の画像を対象にします．

```txt
.jpg
.jpeg
.png
.webp
```

## コマンド概要

`albam` には，主に次のコマンドがあります．

| command         | 説明                                                       |
| --------------- | ---------------------------------------------------------- |
| `albam init`    | 新しいアルバムプロジェクトを作成します                     |
| `albam index`   | 画像ディレクトリを読み取り，アルバム・写真情報を保存します |
| `albam build`   | Astro 製テーマをビルドします                               |
| `albam serve`   | Web UI，API，画像ファイルを配信します                      |
| `albam version` | バージョン情報を表示します                                 |

## API だけを起動する

テーマ開発中は，Go サーバーを API と画像配信だけで起動できます．

```sh
albam index ./albums
albam serve --api-only --port 8080
```

別ターミナルで Astro の開発サーバーを起動します．

```sh
cd themes/default
pnpm install
pnpm dev
```

この使い方では，バックエンドとテーマを分けて開発できます．

## 本番に近い確認をする

ビルド済みテーマを Go サーバーから配信する場合は，次の順に実行します．

```sh
albam index ./albums
albam build
albam serve
```

本番運用では，`albam serve` を systemd などで常駐させ，Caddy や nginx などのリバースプロキシを前段に置く構成を想定しています．

```txt
Browser
  -> Caddy / nginx
  -> albam serve
      -> static theme files
      -> /api/*
      -> /media/*
```

## 必要なもの

`albam index` と `albam serve --api-only` を使うだけなら，Go 製バイナリだけで動作します．

`albam build` を使う場合は，テーマディレクトリで `pnpm build` を実行するため，`pnpm` が必要です．

```sh
cd themes/default
pnpm install
```

また，`go install` や手元ビルドで SQLite ドライバをビルドする場合，環境によっては C コンパイラが必要です．
