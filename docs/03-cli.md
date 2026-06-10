# albam CLI Document

## 概要

`albam` CLI は，ローカルの画像ディレクトリを読み取り，アルバム情報を生成し，Go 製サーバーとして公開するためのコマンドラインツールです．

Hugo のように，1 つの Go 製バイナリで次を担当します．

```txt
- プロジェクト初期化
- 画像ディレクトリのスキャン
- アルバム・写真メタデータのインデックス作成
- Astro テーマのビルド
- 静的ファイル，API，画像ファイルの配信
- 本番用サーバーの起動
```

本番では `albam serve` を systemd などで常駐させ，Caddy や nginx などのリバースプロキシを前段に置くことを想定します．

```txt
Browser
  -> Caddy / nginx
  -> albam serve
      -> static theme files
      -> /api/*
      -> /media/*
```

## 基本方針

CLI は次の思想で設計します．

```txt
albam init      プロジェクトを作る
albam index     アルバム・写真情報を保存する
albam build     Astro テーマを静的ファイルとしてビルドする
albam serve     Go サーバーとして公開する
albam version   バージョン情報を表示する
albam doctor    設定や依存関係を確認する
```

MVP では，まず次のコマンドを実装対象にします．

```txt
albam index
albam build
albam serve
albam version
```

`albam index` は，画像ディレクトリのスキャンとメタデータ保存をまとめて行います．
`albam doctor` は，運用時の確認用として後から追加します．

## コマンド一覧

```txt
albam init [--force] [--theme default] [--no-theme] [dir]
albam index [--config <path>] [dir]
albam build [--config <path>]
albam serve [--config <path>] [--host <host>] [--port <port>] [--public-dir <path>] [--api-only]
albam version
albam version --verbose
albam doctor
```

## 共通オプション

すべてのサブコマンドで共通して利用できるオプションです．

```txt
--config <path>     設定ファイルのパス
--verbose           詳細ログを表示する
--quiet             通常ログを抑制する
--help              ヘルプを表示する
```

オプションは `--option` 形式に統一します．`-option` 形式は使用しません．

例:

```sh
albam serve --config albam.toml
albam index --config albam.toml ./albums
```

## ディレクトリ構成

`albam init` 後の標準的な構成は次の通りです．

```txt
my-album/
├── albam.toml
├── albums/
│   ├── weekend-trip/
│   │   ├── IMG_001.jpg
│   │   ├── IMG_002.jpg
│   │   └── album.toml
│   └── coffee-time/
│       ├── IMG_101.jpg
│       └── album.toml
├── themes/
│   └── default/
└── .albam/
    ├── public/
    ├── cache/
    └── db.sqlite
```

役割は次の通りです．

| path                   | role                                   |
| ---------------------- | -------------------------------------- |
| `albam.toml`           | プロジェクト設定                       |
| `albums/`              | 元画像ディレクトリ                     |
| `albums/**/album.toml` | アルバム単位のメタデータ               |
| `themes/default/`      | Astro 製テーマ                         |
| `.albam/public/`       | ビルド済みテーマ                       |
| `.albam/cache/`        | サムネイル・プレビュー画像のキャッシュ |
| `.albam/db.sqlite`     | アルバム・写真メタデータ               |

## 設定ファイル

標準設定ファイルは `albam.toml` です．

```toml
title = "My Albums"

[server]
host = "127.0.0.1"
port = 8080

[media]
source_dir = "albums"
cache_dir = ".albam/cache"
allow_original_download = false

[database]
path = ".albam/db.sqlite"

[build]
out_dir = ".albam/public"

[theme]
name = "default"
dir = "themes/default"

[theme.params.appearance]
accent = "coral"

[theme.params.layout]
album_grid_columns = 4

[theme.params.features]
show_header = true
show_footer = true

[theme.params.content]
brand = "albam"
home_title = "My Albums"
home_eyebrow = "SELF-HOSTED PHOTO ALBUM"
home_description = "a simple folder-based album"
copyright = "© 2026"
```

テーマに依存する設定のデフォルト値は，テーマ作者が管理する `themes/default/theme.toml` の `[defaults]` 以下に書きます．
サイト所有者は，`albam.toml` の `[theme.params]` 以下に同じ構造で上書きを書きます．
`albam build` は両者をマージし，`.albam/generated/theme.json` を生成して Astro テーマへ渡します．
`albam serve` では，`appearance` のような見た目設定を `/theme/runtime.css` として動的に反映します．
`content` の表示文言は `/api/config` からブラウザ側で読み込み，ページ読み込み時に動的に反映します．
文字列設定に空文字列を指定した場合は，デフォルト値へフォールバックせず，空の表示として扱われます．

```toml
[theme.params.appearance]
accent = "coral"

[theme.params.layout]
photo_grid = "justified"
album_grid_columns = 4

[theme.params.features]
show_header = true
show_footer = true
show_tags = true
show_album_count = true

[theme.params.content]
brand = "albam"
home_title = "My Albums"
home_eyebrow = "SELF-HOSTED PHOTO ALBUM"
home_description = "a simple folder-based album"
copyright = "© 2026"
footer_text = ""
```

`accent` は，`pink`，`coral`，`mint`，`blue`，`lavender`，`lemon`，`red`，`sakura` から選べます．

## albam init

新しい `albam` プロジェクトを作成します．

```txt
albam init [--force] [--theme default] [--no-theme] [dir]
```

### 用途

指定したディレクトリに，設定ファイル，画像ディレクトリ，キャッシュディレクトリ，テーマディレクトリを作成します．
`dir` を省略した場合は，カレントディレクトリを初期化します．

### 例

```sh
albam init my-album
cd my-album
```

カレントディレクトリを初期化する場合:

```sh
albam init
```

### 作成されるファイル

```txt
albam.toml
albums/
albums/sample.png
themes/
themes/default/
.albam/
```

### オプション

```txt
--force               既存ファイルがあっても初期化する
--theme default       インストールするテーマ
--no-theme            テーマを作成しない
```

### 例

```sh
albam init
albam init my-album
albam init --no-theme my-album
```

### 注意

`--force` は既存ファイルを上書きする可能性があります．
`themes/default` は，GitHub Release asset の `albam-theme-default_<version>.tar.gz` から取得します．
リリース版の CLI では同じ version tag の Release を使い，`dev` ビルドでは latest Release を使います．
`checksums.txt` が Release に含まれる場合は，テーマ tarball の SHA256 を検証します．
ネットワークに接続できない場合や，リリースが存在しない場合は失敗します．
MVP では，既存の `albam.toml` がある場合はエラーにするのが安全です．

## albam index

画像ディレクトリをスキャンし，アルバム・写真情報を保存します．

```txt
albam index [--config <path>] [dir]
```

### 用途

ディレクトリ構造からアルバム構造を作り，SQLite などの保存先にメタデータを保存します．
Go バックエンドでは，画像本体を DB に保存せず，ファイルパスとメタデータのみを保存する方針です

### 例

```sh
albam index ./albums
```

または，`albam.toml` の `[media].source_dir` を使う場合:

```sh
albam index
```

### 挙動

```txt
1. 画像ディレクトリを再帰的にスキャンする
2. 親ディレクトリをアルバムとして解釈する
3. album.toml があれば読み込む
4. アルバム情報を保存する
5. 写真情報を保存する
6. 不要になった古いインデックスを更新する
```

### オプション

```txt
--config <path>        設定ファイルのパス
--dry-run             保存せず，変更予定だけ表示する
--reset               既存インデックスを削除して作り直す
--json                結果を JSON で出力する
--no-thumbnails       サムネイル生成を行わない
```

### 出力例

```txt
Indexed albums: 2
Indexed photos: 128
Updated photos: 4
Removed photos: 1
```

### JSON 出力例

```json
{
  "albums": {
    "created": 2,
    "updated": 0,
    "removed": 0
  },
  "photos": {
    "created": 124,
    "updated": 4,
    "removed": 1
  }
}
```

## albam build

Astro 製テーマを静的ファイルとしてビルドします．

```txt
albam build [--config <path>]
```

### 用途

`themes/default` などのテーマをビルドし，`.albam/public` に出力します．
フロントエンドは Astro 製テーマとして実装する方針で，対象ディレクトリは `themes/default` 以下です

本番では，Go サーバーが `.albam/public` を静的ファイルとして配信します．

### 例

```sh
albam build
```

### 挙動

```txt
1. 設定ファイルを読み込む
2. theme.dir を確認する
3. Astro テーマの依存関係を確認する
4. テーマをビルドする
5. build.out_dir に出力する
```

### オプション

```txt
--config <path>       設定ファイルのパス
```

### 実行例

```sh
albam build
albam build --config albam.toml
```

### 内部で実行する処理

MVP では，テーマディレクトリで次を実行する想定です．

```sh
pnpm build
```

`pnpm install` は毎回自動実行しない方がよいです．
依存関係が存在しない場合は，エラーメッセージで `pnpm install` を促します．

例:

```txt
theme dependencies are not installed.
Run:

  cd themes/default
  pnpm install
```

### 注意

`albam build` は Node.js / pnpm / Astro に依存します．
ただし，ビルド後の本番運用では Node.js を常駐させません．

## albam serve

Go 製 HTTP サーバーを起動します．

```txt
albam serve [--config <path>] [--host <host>] [--port <port>] [--public-dir <path>] [--api-only]
```

### 用途

`albam serve` は，静的テーマ，JSON API，画像ファイルをまとめて配信します．
API ドキュメントでも，本番運用では `albam serve` が静的ファイル，`/api/*`，`/media/*` を担当する前提です

### 例

```sh
albam serve
```

ホストとポートを指定する場合:

```sh
albam serve --host 127.0.0.1 --port 8080
```

設定ファイルを指定する場合:

```sh
albam serve --config albam.toml
```

### 配信するパス

```txt
/                         静的テーマ
/albums/{album_id}/       静的テーマ
/assets/*                 Astro のビルド済み assets
/api/*                    JSON API
/media/*                  画像配信
```

API は `/api` 以下，画像ファイルは `/media` 以下に配置します

### オプション

```txt
--host <host>             listen host
--port <port>             listen port
--public-dir <path>       静的ファイルの配信元
--watch                   ファイル変更を監視して再インデックスする
--api-only                静的ファイルを配信せず API のみ起動する
--no-open                 ブラウザを自動で開かない
```

### 出力例

```txt
albam server started

  Local:   http://127.0.0.1:8080
  API:     http://127.0.0.1:8080/api
  Media:   http://127.0.0.1:8080/media
```

### 本番運用例

systemd で `albam serve` を常駐させます．

```ini
[Unit]
Description=albam server
After=network.target

[Service]
Type=simple
WorkingDirectory=/srv/albam
ExecStart=/usr/local/bin/albam serve --host 127.0.0.1 --port 8080
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Caddy でリバースプロキシします．

```caddyfile
album.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

## albam doctor

設定，依存関係，ディレクトリ構成を確認します．

```txt
albam doctor
```

### 用途

本番公開前やトラブルシュート時に，環境が正しく整っているか確認します．

### 確認項目

```txt
- albam.toml が存在するか
- albums/ が存在するか
- .albam/ が存在するか
- database.path にアクセスできるか
- media.source_dir が存在するか
- build.out_dir が存在するか
- theme.dir が存在するか
- pnpm が利用可能か
- theme の package.json が存在するか
- ポートが利用可能か
```

### 出力例

```txt
[ok] config: albam.toml
[ok] media source: albums
[ok] database: .albam/db.sqlite
[ok] theme: themes/default
[warn] build output does not exist: .albam/public
[warn] pnpm is not installed
```

### オプション

```txt
--json                JSON 形式で出力する
```

## albam version

バージョン情報を表示します．

```txt
albam version
```

### 出力例

```txt
albam 0.1.0
```

詳細表示:

```sh
albam version --verbose
```

```txt
albam 0.1.0
commit: abcdef0
built: 2026-06-03T12:00:00+09:00
go: go1.23
```

## アルバムメタデータ

各アルバムディレクトリには，任意で `album.toml` を置けます．

```txt
albums/weekend-trip/
├── album.toml
├── IMG_001.jpg
└── IMG_002.jpg
```

例:

```toml
title = "Weekend trip"
description = "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。"
date = "2026-05-18"
tags = ["travel", "sea", "cafe", "friends"]
visibility = "public"
cover = "IMG_001.jpg"
```

`album.toml` がない場合は，ディレクトリ名からタイトルと ID を生成します．

```txt
albums/weekend-trip/
  -> id: weekend-trip
  -> title: Weekend trip
```

## 終了コード

CLI は，処理結果に応じて次の終了コードを返します．

| code | meaning                          |
| ---: | -------------------------------- |
|  `0` | 成功                             |
|  `1` | 一般的なエラー                   |
|  `2` | 設定ファイルエラー               |
|  `3` | 入力ディレクトリ・ファイルエラー |
|  `4` | インデックス・DB エラー          |
|  `5` | テーマビルドエラー               |
|  `6` | サーバー起動エラー               |

## ログ方針

通常時は，人間が読みやすい簡潔なログを出します．

```txt
Indexed 2 albums and 128 photos.
```

`--verbose` のときは，ファイル単位の詳細ログを出します．

```txt
scan: found albums/weekend-trip/IMG_001.jpg
index: created photo img-001
thumbnail: generated .albam/cache/thumbs/img-001.webp
```

`--quiet` のときは，エラー以外を出力しません．

## MVP で実装するコマンド

最初の MVP では，次を実装します．

```txt
albam index [--config <path>] [dir]
albam build [--config <path>]
albam serve
albam version
```

MVP では後回しでよいもの:

```txt
albam build
albam doctor
albam serve --watch
```

理由は，まず Go 側で画像を読み，アルバム情報を返し，API と画像配信が動くことを優先するためです．
`index` が画像ディレクトリのスキャンと DB 保存をまとめて担当し，その後 `serve` で API を公開します．

## 実装順

実装順は次がよいです．

```txt
1. albam index [--config <path>] [dir]
2. albam serve --api-only
3. albam serve
4. albam build
5. albam version
6. albam init [--force] [--theme default] [--no-theme] [dir]
7. albam doctor
8. albam serve --watch
```

`albam build` より前に `albam serve --api-only` を作ると，バックエンド API と Astro テーマを別々に開発できます．
この段階では，Astro 側は静的モックで実装してよく，API 接続は後回しにできます

## 開発時の使い方

フロントエンドを静的モックで作る段階:

```sh
cd themes/default
pnpm install
pnpm dev
```

バックエンド API を開発する段階:

```sh
albam index ./albums
albam serve --api-only --port 8080
```

フロントエンドと API を接続する段階:

```sh
albam serve --api-only --port 8080
cd themes/default
pnpm dev
```

本番に近い動作確認:

```sh
albam index ./albums
albam build
albam serve --host 127.0.0.1 --port 8080
```

## 本番デプロイ手順

標準的な本番手順は次の通りです．

```sh
cd /srv/albam
albam index ./albums
albam build
albam serve --host 127.0.0.1 --port 8080
```

常駐化する場合は，`albam serve` のみを systemd で実行します．
画像を追加した場合は，手動で次を実行します．

```sh
albam index
albam build
sudo systemctl restart albam
```

将来的に `albam serve --watch` を実装すれば，画像追加時に自動で再インデックスできます．

## セキュリティ上の注意

画像ファイル配信では，パストラバーサル対策を必ず行います．
DB に保存された相対パスを使う場合でも，最終的な実パスが必ず `media.source_dir` 以下に収まることを確認します．

```txt
OK:
  albums/weekend-trip/IMG_001.jpg

NG:
  ../../etc/passwd
```

`albam serve` は，直接インターネットに公開するより，Caddy や nginx の背後で動かすことを推奨します．
認証が必要な場合は，Cloudflare Access やリバースプロキシ側の認証を使う想定です．

## 仕様上の未決定事項

次は，実装しながら決めてよいです．

```txt
- 設定ファイルを TOML に固定するか
- DB を最初から SQLite にするか，MVP だけ catalog.json にするか
- album.toml のファイル名
- photo_id の生成規則
- cover 未指定時の選び方
- サムネイル生成を index 時に行うか，初回アクセス時に行うか
- albam build で pnpm install まで自動実行するか
```

現時点では，MVP は次のようにするのが素直です．

```txt
- 設定ファイルは albam.toml
- DB は SQLite
- アルバムメタデータは album.toml
- サムネイルは初回アクセス時に生成
- pnpm install は自動実行しない
```

この方針なら，Go 製 CLI らしい使い勝手を保ちつつ，Astro テーマとの境界も明確にできます．
