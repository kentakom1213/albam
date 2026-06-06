# Codex 追加指示: 写真グリッドをより自然な Google Photos 風タイリングに改善する

## 背景

現在の写真グリッドは，右端の極端に細いタイルは解消されているが，全体としてまだ通常のグリッドに近く見える．

原因として，次が考えられる．

```txt
- 行ごとの高さや枚数の変化が弱い
- 写真カードの幅が似通っている
- 写真の多くが近い縦横比である
- rowHeight の制御が強く，行ごとのリズムが単調になっている
- border / shadow / gap により，写真ギャラリーというよりカード一覧に見えている
```

目標は，CSS Grid 風の整然とした配置ではなく，Google Photos のような自然な横方向タイリングに近づけることである．

ただし，真の packing や Masonry は実装しない．写真順は維持する．

## 変更方針

現在の `fixed row tiling` を，より自然な `balanced justified tiling` に寄せる．

基本方針は次の通りである．

```txt
- 各行の高さは完全固定にしない
- targetRowHeight を基準にしつつ，minRowHeight / maxRowHeight の範囲で変化させる
- 各行について，写真の縦横比の合計から自然な行高を計算する
- 極端に小さいタイルや大きすぎるタイルが出る行分割は避ける
- 同じ枚数の行が連続しすぎないようにする
- 最終行だけ不自然に大きくならないようにする
```

## レイアウトアルゴリズム

行に含まれる写真集合を `row` とする．
各写真の縦横比を次で定義する．

```txt
aspectRatio = width / height
```

行内の写真部分に使える幅は次である．

```txt
availableWidth = containerWidth - gap * (row.length - 1)
```

行内の縦横比の合計は次である．

```txt
aspectSum = sum(photo.aspectRatio)
```

この行を横幅いっぱいに表示するための自然な行高は次である．

```txt
rowHeight = availableWidth / aspectSum
```

各写真の表示幅は次である．

```txt
tileWidth = rowHeight * photo.aspectRatio
```

この方式では，特定の 1 枚だけを残り幅に無理やり押し込まない．
行全体の高さを決めてから，各写真の幅を縦横比に応じて決める．

## 行候補の選び方

現在位置 `startIndex` から，最大 `maxItemsPerRow` 枚まで候補行を試す．

例えば `maxItemsPerRow = 6` の場合，次を候補にする．

```txt
photos[startIndex : startIndex + 2]
photos[startIndex : startIndex + 3]
photos[startIndex : startIndex + 4]
photos[startIndex : startIndex + 5]
photos[startIndex : startIndex + 6]
```

各候補行について，`rowHeight` と各 `tileWidth` を計算し，最も自然な候補を採用する．

1 枚だけの行は原則避ける．ただし，残り写真数が 1 枚の場合は許容する．

## スコア関数

候補行のスコアは，次の観点で評価する．

```txt
- rowHeight が targetRowHeight に近いほどよい
- rowHeight が minRowHeight 未満なら大きなペナルティ
- rowHeight が maxRowHeight を超えるなら大きなペナルティ
- tileWidth が minTileWidth 未満なら大きなペナルティ
- tileWidth が maxTileWidth を超えるなら軽いペナルティ
- 1 枚だけの行にはペナルティ
- 同じ row.length が連続する場合はペナルティ
- 同じ row.length が 3 回以上続く場合は強いペナルティ
- 前行と rowHeight が近すぎる場合は軽いペナルティ
```

特に，通常のグリッドに見えることを避けるため，次を重視する．

```txt
- 同じ枚数の行を連続させすぎない
- 行ごとの高さ差を少しだけ作る
- 最終行を大きく引き伸ばさない
```

## 推奨パラメータ

現在の見た目では，行ごとの高さ差がやや大きく，最終行が目立ちやすい．
まずは次の値を使う．

```ts
const layout = buildBalancedJustifiedLayout(photos, {
  containerWidth: 1380,
  gap: 10,
  targetRowHeight: 205,
  minRowHeight: 175,
  maxRowHeight: 240,
  minTileWidth: 145,
  maxTileWidth: 560,
  minItemsPerRow: 2,
  maxItemsPerRow: 6,
  avoidSameRowLengthPenalty: 120,
  avoidSameHeightPenalty: 40,
  justifyLastRow: false,
  lastRowMaxHeight: 205,
});
```

調整方針は次である．

```txt
グリッドっぽく見える場合:
  - avoidSameRowLengthPenalty を上げる
  - maxTileWidth を少し上げ，横長写真の差を潰しすぎない
  - targetRowHeight を少し上げる
  - feature row を数行に 1 回入れる

最終行が大きすぎる場合:
  - justifyLastRow を false にする
  - lastRowMaxHeight を targetRowHeight 以下にする

右端が細くなる場合:
  - minTileWidth を上げる
  - minRowHeight を上げる
  - rowHeight が minRowHeight 未満の候補に強いペナルティを入れる
```

## 実装ファイル

既存実装に合わせてよいが，基本的には次を更新する．

```txt
themes/default/src/
├── lib/
│   └── balancedJustifiedLayout.ts
├── components/
│   ├── PhotoGrid.astro
│   └── PhotoTile.astro
└── pages/
    └── albums/
        └── [albumId].astro
```

既存の `fixedRowLayout.ts` がある場合は，置き換えるか，`balancedJustifiedLayout.ts` を新設して `PhotoGrid.astro` 側の import を変更する．

## 型定義

```ts
export type JustifiedPhotoInput = {
  id: string;
  width: number;
  height: number;
  aspectRatio?: number;
};

export type BalancedJustifiedLayoutOptions = {
  containerWidth: number;
  gap: number;
  targetRowHeight: number;
  minRowHeight: number;
  maxRowHeight: number;
  minTileWidth: number;
  maxTileWidth: number;
  minItemsPerRow: number;
  maxItemsPerRow: number;
  avoidSameRowLengthPenalty: number;
  avoidSameHeightPenalty: number;
  justifyLastRow: boolean;
  lastRowMaxHeight: number;
};

export type BalancedJustifiedLayoutItem = {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
};

export type BalancedJustifiedLayoutResult = {
  width: number;
  height: number;
  items: BalancedJustifiedLayoutItem[];
};
```

## `buildBalancedJustifiedLayout`

```ts
export function buildBalancedJustifiedLayout(
  photos: JustifiedPhotoInput[],
  options: BalancedJustifiedLayoutOptions,
): BalancedJustifiedLayoutResult {
  const items: BalancedJustifiedLayoutItem[] = [];

  let index = 0;
  let y = 0;
  let previousRowLength: number | null = null;
  let previousRowHeight: number | null = null;
  let sameRowLengthRun = 0;

  while (index < photos.length) {
    const row = chooseNextBalancedRow({
      photos,
      startIndex: index,
      previousRowLength,
      previousRowHeight,
      sameRowLengthRun,
      options,
    });

    const isLastRow = index + row.length >= photos.length;

    const layouted = layoutBalancedRow({
      row,
      y,
      isLastRow,
      options,
    });

    items.push(...layouted.items);

    if (previousRowLength === row.length) {
      sameRowLengthRun += 1;
    } else {
      sameRowLengthRun = 1;
    }

    previousRowLength = row.length;
    previousRowHeight = layouted.height;

    y += layouted.height + options.gap;
    index += row.length;
  }

  return {
    width: options.containerWidth,
    height: Math.max(0, y - options.gap),
    items,
  };
}
```

## `chooseNextBalancedRow`

```ts
function chooseNextBalancedRow(args: {
  photos: JustifiedPhotoInput[];
  startIndex: number;
  previousRowLength: number | null;
  previousRowHeight: number | null;
  sameRowLengthRun: number;
  options: BalancedJustifiedLayoutOptions;
}): JustifiedPhotoInput[] {
  const {
    photos,
    startIndex,
    previousRowLength,
    previousRowHeight,
    sameRowLengthRun,
    options,
  } = args;

  const remaining = photos.length - startIndex;

  if (remaining <= options.minItemsPerRow) {
    return photos.slice(startIndex);
  }

  let bestRow = photos.slice(
    startIndex,
    Math.min(photos.length, startIndex + options.minItemsPerRow),
  );
  let bestScore = Number.POSITIVE_INFINITY;

  const maxEnd = Math.min(photos.length, startIndex + options.maxItemsPerRow);

  for (let end = startIndex + options.minItemsPerRow; end <= maxEnd; end += 1) {
    const row = photos.slice(startIndex, end);
    const metrics = getRowMetrics(row, options);

    const score = getBalancedRowScore({
      row,
      metrics,
      previousRowLength,
      previousRowHeight,
      sameRowLengthRun,
      options,
    });

    if (score < bestScore) {
      bestScore = score;
      bestRow = row;
    }
  }

  return bestRow;
}
```

## `getRowMetrics`

```ts
function getRowMetrics(
  row: JustifiedPhotoInput[],
  options: BalancedJustifiedLayoutOptions,
): {
  rowHeight: number;
  tileWidths: number[];
  minTileWidth: number;
  maxTileWidth: number;
} {
  const aspectRatios = row.map(getAspectRatio);
  const aspectSum = aspectRatios.reduce((sum, value) => sum + value, 0);

  const gapWidth = options.gap * Math.max(0, row.length - 1);
  const availableWidth = options.containerWidth - gapWidth;

  const rowHeight =
    aspectSum > 0 ? availableWidth / aspectSum : options.targetRowHeight;

  const tileWidths = aspectRatios.map((aspectRatio) => rowHeight * aspectRatio);

  return {
    rowHeight,
    tileWidths,
    minTileWidth: Math.min(...tileWidths),
    maxTileWidth: Math.max(...tileWidths),
  };
}
```

## `getBalancedRowScore`

```ts
function getBalancedRowScore(args: {
  row: JustifiedPhotoInput[];
  metrics: {
    rowHeight: number;
    tileWidths: number[];
    minTileWidth: number;
    maxTileWidth: number;
  };
  previousRowLength: number | null;
  previousRowHeight: number | null;
  sameRowLengthRun: number;
  options: BalancedJustifiedLayoutOptions;
}): number {
  const {
    row,
    metrics,
    previousRowLength,
    previousRowHeight,
    sameRowLengthRun,
    options,
  } = args;

  let score = Math.abs(metrics.rowHeight - options.targetRowHeight);

  if (metrics.rowHeight < options.minRowHeight) {
    score += (options.minRowHeight - metrics.rowHeight) * 10;
  }

  if (metrics.rowHeight > options.maxRowHeight) {
    score += (metrics.rowHeight - options.maxRowHeight) * 8;
  }

  for (const width of metrics.tileWidths) {
    if (width < options.minTileWidth) {
      score += (options.minTileWidth - width) * 12;
    }

    if (width > options.maxTileWidth) {
      score += (width - options.maxTileWidth) * 2;
    }
  }

  if (row.length === 1) {
    score += options.containerWidth * 0.5;
  }

  if (previousRowLength === row.length) {
    score += options.avoidSameRowLengthPenalty;

    if (sameRowLengthRun >= 2) {
      score += options.avoidSameRowLengthPenalty * 2;
    }
  }

  if (previousRowHeight !== null) {
    const heightDiff = Math.abs(metrics.rowHeight - previousRowHeight);

    if (heightDiff < 10) {
      score += options.avoidSameHeightPenalty;
    }
  }

  return score;
}
```

## `layoutBalancedRow`

```ts
function layoutBalancedRow(args: {
  row: JustifiedPhotoInput[];
  y: number;
  isLastRow: boolean;
  options: BalancedJustifiedLayoutOptions;
}): {
  height: number;
  items: BalancedJustifiedLayoutItem[];
} {
  const { row, y, isLastRow, options } = args;
  const metrics = getRowMetrics(row, options);

  let rowHeight = metrics.rowHeight;

  if (isLastRow && !options.justifyLastRow) {
    rowHeight = Math.min(options.targetRowHeight, options.lastRowMaxHeight);
  } else {
    rowHeight = clamp(rowHeight, options.minRowHeight, options.maxRowHeight);
  }

  const roundedHeight = Math.round(rowHeight);

  let x = 0;

  const items = row.map((photo) => {
    const width = Math.round(getAspectRatio(photo) * rowHeight);

    const item: BalancedJustifiedLayoutItem = {
      id: photo.id,
      x,
      y,
      width,
      height: roundedHeight,
    };

    x += width + options.gap;

    return item;
  });

  return {
    height: roundedHeight,
    items,
  };
}
```

## 補助関数

```ts
function getAspectRatio(photo: JustifiedPhotoInput): number {
  if (
    photo.aspectRatio &&
    Number.isFinite(photo.aspectRatio) &&
    photo.aspectRatio > 0
  ) {
    return photo.aspectRatio;
  }

  if (photo.width > 0 && photo.height > 0) {
    return photo.width / photo.height;
  }

  return 1;
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}
```

## CSS の調整

現在は写真カードの装飾が強く，カードグリッドに見えやすい．
写真ギャラリーらしくするため，通常時の影と枠線を弱める．

```css
.photo-grid {
  width: 100%;
  max-width: 1380px;
}

.photo-grid__canvas {
  position: relative;
  width: 100%;
}

.photo-tile {
  position: absolute;
  display: block;
  overflow: hidden;
  border-radius: 12px;
  border: 1px solid rgb(0 0 0 / 6%);
  background: var(--theme-surface);
  box-shadow: 0 4px 12px rgb(0 0 0 / 3%);
  transition:
    box-shadow 160ms ease,
    border-color 160ms ease,
    opacity 160ms ease;
}

.photo-tile:hover {
  border-color: color-mix(
    in srgb,
    var(--theme-current-accent) 30%,
    var(--theme-border)
  );
  box-shadow: 0 10px 24px rgb(0 0 0 / 7%);
}

.photo-tile img,
.photo-tile__placeholder {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
```

## 重要: デスクトップでは CSS Grid を使わない

デスクトップでは，写真配置に CSS Grid を使わない．
`buildBalancedJustifiedLayout` の `x`，`y`，`width`，`height` を使って絶対配置する．

次のような指定がデスクトップで効いていないか確認する．

```css
grid-template-columns: repeat(...);
aspect-ratio: 1 / 1;
grid-auto-flow: dense;
```

これらはモバイル fallback に限定する．

## モバイル fallback

モバイルでは，絶対配置を維持せず，2 カラムの CSS Grid にフォールバックしてよい．

```css
@media (max-width: 760px) {
  .photo-grid__canvas {
    height: auto !important;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .photo-tile {
    position: relative;
    width: auto !important;
    height: auto !important;
    transform: none !important;
    aspect-ratio: 1 / 1;
  }
}
```

横長写真を少し活かしたい場合は，横長写真だけ `grid-column: span 2` にしてよい．
ただし，`grid-auto-flow: dense` は使わない．

## さらに自然にする任意改善

写真の多くが同じ縦横比の場合，縦横比を忠実に守るだけではグリッド感が残る．
より自然にしたい場合は，次の改善を検討する．

### feature row を入れる

数行に 1 回だけ，少し大きめの行を入れる．

```txt
通常行: targetRowHeight = 205
feature row: targetRowHeight = 245
```

ただし，写真順は維持する．
feature row は，その行の写真を少し大きく見せるだけであり，Masonry や packing にはしない．

### 行ごとの target height にリズムをつける

行番号に応じて，基準行高を少し変える．

```ts
const targetRowHeights = [205, 220, 195, 235, 210];
```

候補行のスコア計算時に，現在行の `targetRowHeight` を使う．
これにより，同じ縦横比の写真が続いても，完全なグリッドには見えにくくなる．

## 完了条件

```txt
- デスクトップで固定カラムグリッドのように見えない
- 行ごとに枚数または高さに自然な変化がある
- 最終行だけ極端に大きくならない
- 右端に極端に細い写真が出ない
- 写真順が維持されている
- デスクトップでは absolute layout で配置している
- モバイルでは CSS Grid fallback で破綻しない
- 追加依存を入れていない
- Go 側のコードを変更していない
```

## 実装後の報告内容

実装後，次を報告する．

```txt
- 変更したファイル一覧
- 採用したレイアウトアルゴリズム
- グリッドっぽさを弱めるために入れた工夫
- 最終行が大きくならないようにした条件
- 右端が細くならないようにした条件
- モバイル fallback の方針
```
