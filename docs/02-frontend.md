# albam フロントエンド実装指示

## 目的

Figma で作成した `albam` のデザイン案を，Astro 製テーマとして実装する．

今回はサーバー側の実装は行わない．
Go API，画像処理，DB，ルーティングサーバーなどは触らず，フロントエンドの見た目とページ構成のみを実装する．

## 対象技術

* Astro
* TypeScript
* CSS
* 静的モックデータ

バックエンド API との接続はまだ行わない．
必要なデータは Astro 側の仮データとして定義する．

## 対象ディレクトリ

`themes/default` 以下のみを変更する．

想定構成は次の通り．

```txt
themes/default/
├── package.json
├── astro.config.mjs
├── tsconfig.json
└── src/
    ├── components/
    │   ├── Layout.astro
    │   ├── Header.astro
    │   ├── Button.astro
    │   ├── Chip.astro
    │   ├── AlbumCard.astro
    │   ├── PhotoCard.astro
    │   └── ImagePlaceholder.astro
    ├── data/
    │   └── mock.ts
    ├── pages/
    │   ├── index.astro
    │   └── albums/
    │       └── [albumId].astro
    └── styles/
        ├── global.css
        └── theme.css
```

既存構成がある場合は，なるべくそれに合わせる．
ただし，サーバー側・CLI 側・Go 側のコードは変更しない．

## Figma デザイン

Figma ファイル:

```txt
https://www.figma.com/design/735a3rFCVGq8cvBWh9JSy6
```

再現対象のフレーム:

```txt
Desktop / Album grid home
Mobile / Album grid home
Desktop / Album detail page
Mobile / Album detail page
Design tokens / flat cute monochrome
Variables guide / editable accent variables
Accent color candidates
```

実装では，特に次の 2 種類の画面を再現する．

```txt
/
```

トップページ．アルバム一覧を表示する．

```txt
/albums/[albumId]/
```

各アルバムページ．アルバム詳細と写真一覧を表示する．

## Figma の渡し方

Codex に対しては，Figma の URL だけでは不十分である．
デザインを忠実に再現させるには，次の形式で渡すのが望ましい．

### 必須

1. Figma の共有 URL
2. 対象フレーム名
3. 各フレームのスクリーンショット
4. design tokens の一覧
5. レイアウト仕様の文章化

### 可能なら追加

6. Figma Dev Mode から取得した CSS 値
7. 各コンポーネントの寸法・余白・角丸・影
8. 使用するフォント，font weight，font size
9. 色変数と CSS variables の対応表

特にスクリーンショットは重要である．
Codex が Figma を直接正確に読めない環境でも，スクリーンショットと設計メモがあれば再現精度が上がる．

## Codex に渡すべきデザイン資料

`docs/design/albam-figma-reference.md` のようなファイルを作り，次の内容を入れる．

```md
# albam Figma reference

## Figma URL

https://www.figma.com/design/735a3rFCVGq8cvBWh9JSy6

## Target frames

- Desktop / Album grid home
- Mobile / Album grid home
- Desktop / Album detail page
- Mobile / Album detail page

## Screenshots

Place screenshots here:

- docs/design/screenshots/home-desktop.png
- docs/design/screenshots/home-mobile.png
- docs/design/screenshots/album-detail-desktop.png
- docs/design/screenshots/album-detail-mobile.png

## Design direction

- Flat design
- No base gradient
- White and black foundation
- One accent color
- Simple and cute
- Instagram-like photo grid
- Soft rounded cards
- Thin borders
- Subtle shadows
- Photos should be the main visual element

## Color tokens

| Figma variable | CSS variable | Default |
|---|---|---|
| theme/bg | --theme-bg | #FAFAFA |
| theme/surface | --theme-surface | #FFFFFF |
| theme/text | --theme-text | #111111 |
| theme/muted | --theme-muted | #777777 |
| theme/border | --theme-border | #E8E8E8 |
| theme/current-accent | --theme-current-accent | #FF6FAE |
| theme/current-accent-soft | --theme-current-accent-soft | #FFE3EF |

## Accent candidates

| Name | accent | accent-soft |
|---|---|---|
| Sakura Pink | #FF6FAE | #FFE3EF |
| Coral Pop | #FF6B5F | #FFE6E2 |
| Mint Fresh | #35C99B | #DFF8EF |
| Sky Blue | #4DA3FF | #E3F1FF |
| Lavender | #9B7CFF | #EEE8FF |
| Lemon | #F4C430 | #FFF5C7 |
| Tomato Red | #F04438 | #FFE4E0 |
| Warm Beige | #C88A5A | #F4E8DD |
```

## 実装方針

まずは Figma の見た目を Astro で静的に再現する．
API 接続は行わず，`src/data/mock.ts` に仮データを置く．

ページは次の 2 つを実装する．

```txt
src/pages/index.astro
src/pages/albums/[albumId].astro
```

`index.astro` はトップページで，アルバム一覧を表示する．
`albums/[albumId].astro` は各アルバムページで，アルバム情報と写真グリッドを表示する．

## 変更してよい範囲

変更してよい:

```txt
themes/default/
```

変更しない:

```txt
cmd/
internal/
pkg/
go.mod
go.sum
```

サーバー側の API 実装は不要．
Go コードは変更しない．

## デザイントークン

`src/styles/theme.css` を作成し，Figma Variables に対応する CSS variables を定義する．

```css
:root {
  --theme-bg: #fafafa;
  --theme-surface: #ffffff;
  --theme-text: #111111;
  --theme-muted: #777777;
  --theme-border: #e8e8e8;
  --theme-current-accent: #ff6fae;
  --theme-current-accent-soft: #ffe3ef;

  --radius-card: 18px;
  --radius-panel: 30px;
  --radius-pill: 999px;

  --shadow-card: 0 10px 28px rgb(0 0 0 / 7%);
  --shadow-panel: 0 12px 30px rgb(0 0 0 / 4.5%);
}
```

色は直書きせず，できるだけ CSS variables を使う．
写真プレースホルダーなどの装飾色だけは，必要に応じて個別色を使ってよい．

## コンポーネント

### `Layout.astro`

共通レイアウトを担当する．

* HTML skeleton
* global CSS の読み込み
* theme CSS の読み込み
* 共通背景
* `<slot />`

### `Header.astro`

トップページと各アルバムページで共通利用する．

要素:

* `albam` ロゴ
* ロゴ横の小さなアクセントドット
* `Albums`
* `Tags`
* `Settings`
* 検索ボタン風の丸いアイコン

モバイルではナビゲーションを隠してよい．

### `Chip.astro`

丸いラベル・フィルタ UI．

props:

```ts
type Props = {
  label: string;
  active?: boolean;
  small?: boolean;
};
```

active のとき:

* background: `--theme-current-accent-soft`
* border: `--theme-current-accent`
* text: `--theme-text`

通常時:

* background: `--theme-surface`
* border: `--theme-border`
* text: `--theme-muted`

### `Button.astro`

操作ボタン．

props:

```ts
type Props = {
  label: string;
  href?: string;
  primary?: boolean;
};
```

primary のとき黒背景，白文字にする．
通常時は白背景，黒文字，薄い枠線にする．

### `AlbumCard.astro`

トップページのアルバムカード．

表示内容:

* 正方形の画像またはプレースホルダー
* アルバムタイトル
* 写真枚数
* 小さなアクセントドット

トップページでは Instagram 風グリッドに並べる．

### `PhotoCard.astro`

各アルバムページの写真カード．

表示内容:

* 正方形の画像またはプレースホルダー
* 写真名
* 選択状態のアクセントドット

### `ImagePlaceholder.astro`

実画像がまだないため，Figma に近い抽象的なプレースホルダーを表示する．

* やわらかい背景色
* 白い円や角丸図形
* 小さなアクセント図形
* 写真が入ったときにも差し替えやすい構造

後で実画像に置き換えやすいように，`src` があれば `<img>`，なければプレースホルダーを表示する設計にする．

## モックデータ

`src/data/mock.ts` を作成する．

```ts
export type Album = {
  id: string;
  title: string;
  kind: string;
  description: string;
  photoCount: number;
  createdAt: string;
  updatedAt: string;
  size: string;
  visibility: "public" | "private";
  tags: string[];
};

export type Photo = {
  id: string;
  title: string;
  albumId: string;
  src?: string;
};

export const albums: Album[] = [
  {
    id: "weekend-trip",
    title: "Weekend trip",
    kind: "TRAVEL ALBUM",
    description:
      "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。",
    photoCount: 48,
    createdAt: "May 18, 2026",
    updatedAt: "May 23, 2026",
    size: "1.2 GB",
    visibility: "public",
    tags: ["travel", "sea", "cafe", "friends", "film"],
  },
];

export const photos: Photo[] = Array.from({ length: 12 }, (_, i) => ({
  id: `photo-${String(i + 1).padStart(3, "0")}`,
  title: `photo-${String(i + 1).padStart(3, "0")}`,
  albumId: "weekend-trip",
}));
```

## トップページ

`src/pages/index.astro` を実装する．

要件:

* Figma の `Desktop / Album grid home` を再現する
* 白背景ではなく `--theme-bg` を使う
* 上部に共通ヘッダー
* 大きな `Your Albums`
* 説明文
* `All`，`Favorites`，`Travel`，`Daily` などの chips
* アルバムカードのグリッド
* デスクトップでは 4 カラム程度
* モバイルでは 2 カラム

## 各アルバムページ

`src/pages/albums/[albumId].astro` を実装する．

要件:

* Figma の `Desktop / Album detail page` と `Mobile / Album detail page` を再現する
* `getStaticPaths` で `albums` からページを生成する
* 上部に共通ヘッダー
* パンくず
* アルバムヒーローカード
* 表紙コラージュ
* アルバム種別
* アルバムタイトル
* 説明文
* 写真数，作成日，公開状態の chips
* `Share`，`Download`，`Edit album` のボタン
* 写真グリッド
* デスクトップでは右側に `Album info` サイドバー
* モバイルではサイドバーを非表示にして，本文を縦積みにする

## レスポンシブ方針

Figma の位置を完全に固定座標で再現しない．
CSS Grid と Flexbox を使い，見た目が近く，レスポンシブに崩れにくい実装にする．

ブレークポイントの目安:

```css
@media (max-width: 900px) {
  /* tablet / mobile */
}

@media (max-width: 640px) {
  /* mobile */
}
```

## デザイン再現の優先順位

優先度高:

1. 色
2. 余白
3. 角丸
4. タイポグラフィ
5. カードの階層
6. グリッドの雰囲気
7. モバイルでの縦積み

優先度中:

1. 影の完全一致
2. プレースホルダー図形の完全一致
3. 細かい座標の完全一致

優先度低:

1. Figma 上の装飾図形のピクセル単位の完全一致
2. 実画像がない状態での写真表現の完全一致

## 注意点

* サーバー側コードは変更しない
* API fetch はまだ書かない
* まず静的モックで見た目を完成させる
* 色は CSS variables を使う
* コンポーネントを過剰に抽象化しない
* ただし `Header`，`Chip`，`Button`，`AlbumCard`，`PhotoCard` は切り出す
* Figma の雰囲気を壊すような既存 UI ライブラリは使わない
* 追加依存はなるべく避ける
* 画像は後で差し替えられるようにする

## 完了条件

次を満たしたら完了とする．

* `pnpm install` が通る
* `pnpm dev` で Astro の開発サーバーが起動する
* `/` でトップページが表示される
* `/albums/weekend-trip/` で各アルバムページが表示される
* デスクトップ幅で Figma の雰囲気に近い
* モバイル幅で Figma の `Mobile` フレームに近い
* 色が `theme.css` の CSS variables から変更できる
* Go サーバー側のコードを変更していない

## Codex への追加指示

実装後に，変更ファイル一覧と，実装したコンポーネントの役割を簡潔に報告すること．
また，Figma と完全一致していない可能性がある箇所があれば，その理由を明記すること．
