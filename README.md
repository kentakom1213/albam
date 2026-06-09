# albam

`albam` は，ローカルの画像ディレクトリを読み取り，アルバム・写真情報を SQLite に保存する Go 製 CLI です．

現在の CLI では，主に次の操作ができます．

```txt
albam init
albam index
albam build
albam serve
albam version
```

## インストール

Go が入っている環境では，次のコマンドでインストールできます．

```sh
go install github.com/kentakom1213/albam/cmd/albam@latest
```

特定バージョンを指定する場合は，次のようにします．

```sh
go install github.com/kentakom1213/albam/cmd/albam@v0.1.0
```

インストール後，次で確認します．

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

## 必要なもの

`albam index` と `albam serve --api-only` を使うだけなら，Go 製バイナリだけで動作します．

`albam build` を使う場合は，テーマディレクトリで `pnpm build` を実行するため，`pnpm` が必要です．

```sh
cd themes/default
pnpm install
```

また，`go install` や手元ビルドで SQLite ドライバをビルドする場合，環境によっては C コンパイラが必要です．

## クイックスタート

新しいプロジェクトを作成します．

```sh
albam init my-album
cd my-album
```

`init` は，設定ファイル，画像ディレクトリ，内部データ用ディレクトリ，サンプル画像を作成します．デフォルトでは `themes/default` も GitHub Releases から取得します．

テーマを取得せずに初期化する場合は，次のようにします．

```sh
albam init my-album --no-theme
```

既存ファイルを上書きして初期化する場合は，`--force` を使います．

```sh
albam init my-album --force
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

デフォルトでは，次のアドレスで起動します．

```txt
http://127.0.0.1:8080
```

## ディレクトリ構成

`albam init my-album` を実行すると，おおむね次の構成が作成されます．

```txt
my-album/
├── albam.toml
├── albums/
│   └── example/
│       └── sample.png
├── themes/
│   └── default/
└── .albam/
```

インデックス作成やビルド後は，次のような構成になります．

```txt
my-album/
├── albam.toml
├── albums/
│   ├── example/
│   │   └── sample.png
│   ├── weekend-trip/
│   │   ├── IMG_001.jpg
│   │   └── IMG_002.jpg
│   └── coffee-time/
│       └── IMG_101.jpg
├── themes/
│   └── default/
└── .albam/
    ├── public/
    ├── cache/
    └── db.sqlite
```

各パスの役割は次の通りです．

| path                        | 説明                          |
| --------------------------- | ----------------------------- |
| `albam.toml`                | プロジェクト設定              |
| `albums/`                   | 元画像ディレクトリ            |
| `albums/example/sample.png` | `init` が作成するサンプル画像 |
| `themes/default/`           | テーマディレクトリ            |
| `.albam/public/`            | ビルド済みテーマ              |
| `.albam/cache/`             | 画像キャッシュ                |
| `.albam/db.sqlite`          | アルバム・写真メタデータ      |

## 設定ファイル

設定ファイルは，デフォルトでは `albam.toml` です．

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

## 画像ディレクトリ

画像は `albums/` 以下に配置します．

```txt
albums/
├── weekend-trip/
│   ├── IMG_001.jpg
│   └── IMG_002.jpg
└── coffee-time/
    └── IMG_101.jpg
```

`albam` は画像ファイルの親ディレクトリをアルバムとして扱います．

```txt
albums/weekend-trip/IMG_001.jpg
  -> album: weekend-trip
```

現時点では，アルバムごとの `album.toml` は読み込まれません．
アルバムタイトルはディレクトリ名から生成されます．

対応している画像拡張子は次の通りです．

```txt
.jpg
.jpeg
.png
.webp
```

## 基本的な使い方

### 1. プロジェクトを作成する

```sh
albam init my-album
cd my-album
```

カレントディレクトリを初期化する場合は，ディレクトリ引数を省略します．

```sh
albam init
```

テーマを取得しない場合は，`--no-theme` を使います．

```sh
albam init --no-theme
```

`init` で作成済みのファイルがある場合はエラーになります．上書きする場合は `--force` を使います．

```sh
albam init --force
```

### 2. インデックスを作成する

画像ディレクトリを読み取り，アルバム・写真情報を SQLite に保存します．

```sh
albam index ./albums
```

`albam.toml` の `[media].source_dir` を使う場合は，ディレクトリ引数を省略できます．

```sh
albam index
```

実行例です．

```txt
indexed 2 albums and 3 assets, removed 0 assets
```

### 3. テーマをビルドする

テーマをビルドし，`[build].out_dir` に出力します．

```sh
albam build
```

デフォルトでは，次のように出力されます．

```txt
themes/default -> .albam/public
```

テーマの依存関係が未インストールの場合は，先に次を実行してください．

```sh
cd themes/default
pnpm install
```

### 4. サーバーを起動する

API，画像，静的ファイルを配信します．

```sh
albam serve
```

デフォルトでは，次のアドレスで起動します．

```txt
http://127.0.0.1:8080
```

ホストやポートを指定する場合は，次のようにします．

```sh
albam serve --host 127.0.0.1 --port 8080
```

静的ファイルを配信せず，API と画像配信だけを起動する場合は，次を使います．

```sh
albam serve --api-only --port 8080
```

静的ファイルのディレクトリを指定する場合は，`--public-dir` を使います．

```sh
albam serve --public-dir .albam/public
```

## コマンド

### `albam init`

新しい `albam` プロジェクトを作成します．

```sh
albam init [--force] [--theme default] [--no-theme] [dir]
```

例です．

```sh
albam init my-album
albam init
albam init my-album --no-theme
albam init my-album --force
```

現在サポートされているテーマ名は `default` のみです．

```sh
albam init --theme default
```

### `albam index`

画像ディレクトリを読み取り，アルバム・写真情報を保存します．

```sh
albam index [--config path] [dir]
```

例です．

```sh
albam index ./albums
albam index --config albam.toml
```

### `albam build`

テーマをビルドします．

```sh
albam build [--config path]
```

例です．

```sh
albam build
albam build --config albam.toml
```

### `albam serve`

API，画像，静的ファイルを配信します．

```sh
albam serve [--config path] [--api-only] [--host host] [--port port] [--public-dir dir]
```

例です．

```sh
albam serve
albam serve --api-only
albam serve --host 127.0.0.1 --port 8080
albam serve --public-dir .albam/public
```

### `albam version`

バージョン情報を表示します．

```sh
albam version
```

詳細表示です．

```sh
albam version --verbose
```

`--version` でもバージョンを表示できます．

```sh
albam --version
```

## 開発時の使い方

API だけを起動します．

```sh
albam index ./albums
albam serve --api-only --port 8080
```

別ターミナルでテーマを開発します．

```sh
cd themes/default
pnpm install
pnpm dev
```

ビルド済みテーマを使って確認する場合は，次の順に実行します．

```sh
albam index ./albums
albam build
albam serve
```
