# albam 設定ファイル仕様

## 概要

`albam` では，利用者が編集する主設定ファイルと，テーマ作者が管理するテーママニフェストを分離する．

```txt
albam.toml
themes/<name>/theme.toml
.albam/generated/theme.json
/theme/runtime.css
```

それぞれの役割は次の通りとする．

```txt
albam.toml
  サイト所有者が編集する設定

themes/<name>/theme.toml
  テーマ作者が管理するデフォルト値マニフェスト

.albam/generated/theme.json
  albam が生成するマージ済みテーマ設定

/theme/runtime.css
  serve 時に生成する動的 CSS variables
```

通常利用者は `theme.toml` を変更せず，テーマ設定を変える場合は `albam.toml` の `[theme.params]` を変更する．

## albam.toml

`albam.toml` は，サイト所有者が編集する主設定ファイルである．

サイト名，サーバー設定，画像ディレクトリ，DB，ビルド出力先，使用テーマ，テーマ設定の上書きを記述する．

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
accent = "sakura"

[theme.params.layout]
photo_grid = "justified"

[theme.params.features]
show_tags = true
show_album_count = true
show_header = true
show_footer = true
```

`theme.params` は，動的か静的かではなく，意味で名前空間を分ける．

標準テーマでは，MVP として次の名前空間を使う．

```toml
[theme.params.appearance]
accent = "sakura"

[theme.params.layout]
photo_grid = "justified"
album_grid_columns = 4

[theme.params.features]
show_tags = true
show_album_count = true
show_header = true
show_footer = true

[theme.params.content]
brand = "albam"
home_title = "My Albums"
home_eyebrow = "SELF-HOSTED PHOTO ALBUM"
home_description = "a simple folder-based album"
copyright = "© 2026"
footer_text = ""
```

## themes/<name>/theme.toml

`themes/<name>/theme.toml` は，テーマ作者が管理するテーママニフェストである．

通常利用者が編集することは想定しない．MVP では，テーマのデフォルト値のみを管理する．設定スキーマや厳密な検証は後回しにする．

テーママニフェストでは，トップレベルにテーマの基本情報を書き，設定値は `[defaults]` 以下にまとめる．

```toml
name = "default"
display_name = "Default"
version = "0.1.0"
author = "Example Author"
description = "Default theme for Albam"

[defaults.appearance]
accent = "sakura"

[defaults.layout]
photo_grid = "justified"

[defaults.features]
show_tags = true
show_album_count = true
show_header = true
show_footer = true
```

MVP 時点では，次の項目を想定する．

| key            | required | description                  |
| -------------- | -------- | ---------------------------- |
| `name`         | yes      | テーマ内部識別子             |
| `display_name` | yes      | UI 表示用のテーマ名          |
| `version`      | yes      | テーマバージョン             |
| `author`       | no       | 作者情報                     |
| `description`  | no       | 説明文                       |
| `[defaults]`   | no       | テーマ設定のデフォルト値     |

将来，テーマ設定の検証や設定 UI が必要になった場合は，`theme.toml` に `[options]` や `theme.schema.json` を追加する．ただし，MVP では `[defaults]` のみを対象とする．

## マージ規則

テーマ設定の最終値は，`theme.toml` の `[defaults]` を `albam.toml` の `[theme.params]` で上書きして決定する．

`albam` 本体は，`theme.params` の意味を原則として解釈しない．`theme.toml` の `[defaults]` と `albam.toml` の `[theme.params]` をマージし，`.albam/generated/theme.json` として出力する．

同じキーが両方にある場合は，`albam.toml` の値を優先する．ネストしたテーブルは，再帰的にマージする．

## .albam/generated/theme.json

`.albam/generated/theme.json` は，`albam` が生成して Astro テーマに渡す中間設定ファイルである．形式は JSON とする．

```json
{
  "site": {
    "title": "My Albums"
  },
  "theme": {
    "name": "default",
    "params": {
      "appearance": {
        "accent": "sakura"
      },
      "layout": {
        "photo_grid": "justified"
      },
      "features": {
        "show_tags": true,
        "show_album_count": true
      }
    }
  }
}
```

`albam build` は，`albam.toml` と `theme.toml` を読み，マージ済み設定を `.albam/generated/theme.json` として生成する．

Astro テーマには，環境変数でこのファイルのパスを渡す．

```txt
ALBAM_THEME_CONFIG_FILE=/absolute/path/to/.albam/generated/theme.json
```

## /theme/runtime.css

`albam serve` でも同じ設定を読めるようにする．

MVP では，`theme.params.appearance` のような見た目設定を `/theme/runtime.css` として動的に反映する．

また，`theme.params.content` のような表示文言は `/api/config` からブラウザ側で読み込み，ページ読み込み時に動的に反映する．

`layout` や `features` は原則としてビルド時反映とし，変更時は再ビルドが必要とする．

つまり，動的かどうかは TOML の名前空間ではなく，実装上の反映ルールとして扱う．

## albam build の挙動

`albam build` は，次の順序で動作する．

```txt
1. albam.toml を読み込む
2. theme.dir を解決する
3. theme.dir/package.json を確認する
4. theme.dir/theme.toml を読み込む
5. theme.toml の defaults と albam.toml の theme.params をマージする
6. .albam/generated/theme.json を生成する
7. ALBAM_THEME_CONFIG_FILE を設定してテーマをビルドする
8. build.out_dir に出力する
```

MVP では，テーマディレクトリで次を実行する．

```sh
pnpm build
```
