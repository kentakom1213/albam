# Codex 追加修正指示: グリッドに見えすぎる写真レイアウトを自然な justified tiling に改善する

## 背景

現在の写真グリッドは，右端の極端に細いタイル問題は改善されたが，今度は全体がほぼ均一な 5 カラムグリッドのように見えている．

原因は，次の可能性が高い．

```txt
- 全行で rowHeight が固定されている
- 写真の多くが近い縦横比である
- minTileWidth / maxTileWidth の clamp により，幅の差が潰れている
- 行分割が毎回ほぼ同じ枚数になる
- CSS 側で aspect-ratio や grid-template-columns が効いている
```

目標は，通常の CSS グリッドではなく，Google Photos に近い自然な justified tiling にすることである．

## 重要な方針変更

以前の `fixed row tiling` をそのまま使わない．
完全固定行高ではなく，次の方式に変更する．

```txt
targetRowHeight を基準にする
minRowHeight / maxRowHeight の範囲で行高を変える
各行について，写真の縦横比の合計から行高を計算する
極端な行高・極端なタイル幅になる行分割は避ける
同じ行パターンが続くことにペナルティを入れる
```

つまり，名前は `balancedJustifiedLayout` とする．

## 採用するアルゴリズム

### 基本式

行に含める写真集合を `row` とする．
各写真の縦横比を `aspectRatio = width / height` とする．

行の写真部分に使える幅は次である．

```txt
availableWidth = containerWidth - gap * (row.length - 1)
```

行の縦横比合計は次である．

```txt
aspectSum = sum(photo.aspectRatio)
```

この行を横幅いっぱいに表示するための自然な行高は次である．

```txt
rowHeight = availableWidth / aspectSum
```

各写真の幅は次である．

```txt
tileWidth = rowHeight * photo.aspectRatio
```

この方式では，ある 1 枚だけを残り幅に押し込まない．
行全体で同じ高さになり，各写真幅は縦横比に応じて自然に決まる．

## 行候補の選び方

写真を先頭から順に処理する．
現在位置 `startIndex` から，最大 `maxItemsPerRow` 枚まで候補行を試す．

例えば，`maxItemsPerRow = 6` の場合，次を候補にする．

```txt
photos[startIndex : startIndex + 1]
photos[startIndex : startIndex + 2]
photos[startIndex : startIndex + 3]
photos[startIndex : startIndex + 4]
photos[startIndex : startIndex + 5]
photos[startIndex : startIndex + 6]
```

各候補行について `rowHeight` と各 `tileWidth` を計算し，スコアが最もよい候補を採用する．

## スコア関数

行候補のスコアは，次の観点で決める．

```txt
- rowHeight が targetRowHeight に近いほどよい
- rowHeight が minRowHeight / maxRowHeight を外れる場合は大きなペナルティ
- tileWidth が minTileWidth より小さい場合は大きなペナルティ
- tileWidth が maxTileWidth より大きい場合はペナルティ
- 1 枚だけの行は避ける
- 同じ row.length が連続する場合は少しペナルティ
- 同じような rowHeight が続く場合も少しペナルティ
```

特に，現在の問題を避けるために，次を入れること．

```txt
同じ枚数の行が 3 回以上続かないようにする
```

これは厳密制約でなく，スコア上のペナルティでよい．

## 推奨パラメータ

デスクトップでは次を基準にする．

```ts
const layoutOptions = {
  containerWidth,
  gap: 14,
  targetRowHeight: 220,
  minRowHeight: 170,
  maxRowHeight: 290,
  minTileWidth: 150,
  maxTileWidth: 520,
  minItemsPerRow: 2,
  maxItemsPerRow: 6,
  avoidSameRowLengthPenalty: 90,
  avoidSameHeightPenalty: 40,
};
```

今のようにグリッドっぽくなる場合は，次のように調整する．

```txt
- targetRowHeight を少し上げる
- maxRowHeight を上げる
- maxItemsPerRow を 6 のままにする
- 同じ row.length へのペナルティを強くする
- maxTileWidth の clamp を弱める
```

逆に，右端が細くなる場合は次で調整する．

```txt
- minTileWidth を上げる
- minRowHeight を上げる
- rowHeight が minRowHeight 未満になる候補を採用しない
```

## 実装方針

`src/lib/fixedRowLayout.ts` を置き換えるか，新しく `src/lib/balancedJustifiedLayout.ts` を作る．
名前は `balancedJustifiedLayout.ts` を推奨する．

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

メイン関数は次のようにする．

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

    const layouted = layoutBalancedRow({
      row,
      y,
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

  let bestRow: JustifiedPhotoInput[] = [photos[startIndex]];
  let bestScore = Number.POSITIVE_INFINITY;

  const maxEnd = Math.min(photos.length, startIndex + options.maxItemsPerRow);

  for (let end = startIndex + options.minItemsPerRow; end <= maxEnd; end++) {
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

  // 残り枚数が少ない場合は，候補がないことがあるため最低 1 枚は返す
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
    score += (options.minRowHeight - metrics.rowHeight) * 8;
  }

  if (metrics.rowHeight > options.maxRowHeight) {
    score += (metrics.rowHeight - options.maxRowHeight) * 8;
  }

  for (const width of metrics.tileWidths) {
    if (width < options.minTileWidth) {
      score += (options.minTileWidth - width) * 10;
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

    if (heightDiff < 12) {
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
  options: BalancedJustifiedLayoutOptions;
}): {
  height: number;
  items: BalancedJustifiedLayoutItem[];
} {
  const { row, y, options } = args;
  const metrics = getRowMetrics(row, options);

  const rowHeight = clamp(
    metrics.rowHeight,
    options.minRowHeight,
    options.maxRowHeight,
  );

  let x = 0;

  const items = row.map((photo) => {
    const width = Math.round(getAspectRatio(photo) * rowHeight);

    const item: BalancedJustifiedLayoutItem = {
      id: photo.id,
      x,
      y,
      width,
      height: Math.round(rowHeight),
    };

    x += width + options.gap;

    return item;
  });

  return {
    height: Math.round(rowHeight),
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

## 重要: CSS でグリッド化しないこと

現在の見た目がグリッド化している場合，CSS 側で次のような指定が残っている可能性がある．

```css
grid-template-columns: repeat(...);
aspect-ratio: 1 / 1;
width: 100%;
height: fixed;
```

デスクトップでは，これらで写真タイルを並べない．
デスクトップの配置は `buildBalancedJustifiedLayout` の `x`, `y`, `width`, `height` を使う．

```css
.photo-grid__canvas {
  position: relative;
}

.photo-tile {
  position: absolute;
}
```

`PhotoTile` の inline style で必ず次を指定する．

```astro
style={`
  transform: translate(${item.x}px, ${item.y}px);
  width: ${item.width}px;
  height: ${item.height}px;
`}
```

## 重要: maxTileWidth で幅を潰しすぎないこと

`maxTileWidth` を強く clamp すると，横長写真の違いが潰れてグリッドっぽくなる．

そのため，`tileWidth = clamp(rowHeight * aspectRatio, minTileWidth, maxTileWidth)` のように，最初から各タイル幅を clamp しすぎない．

今回の推奨は次である．

```txt
行候補の評価時には minTileWidth / maxTileWidth をペナルティとして使う
実際の配置では rowHeight * aspectRatio を基本にする
```

つまり，`maxTileWidth` は強制切り詰めではなく，候補選択のペナルティとして扱う．

## さらに自然にしたい場合

写真の多くが同じ縦横比の場合，どんなレイアウトでもグリッドに近くなりやすい．
これはアルゴリズムの問題というより，入力の縦横比が揃っているためである．

その場合，より自然にするには，次のどちらかが必要になる．

### 案 A: 行ごとの高さにリズムをつける

`targetRowHeight` を常に同じにせず，行ごとに少しだけ変える．

```ts
const targetRowHeights = [210, 240, 225, 270, 220];
```

行番号に応じて，基準行高を変える．

```ts
const rowTargetHeight = targetRowHeights[rowIndex % targetRowHeights.length];
```

これにより，同じ 4:3 写真が多い場合でも，行ごとの高さと枚数に変化が出る．

### 案 B: feature row を入れる

数行に 1 回，少し大きめの行を作る．

```txt
通常行: targetRowHeight = 210
feature row: targetRowHeight = 280
```

ただし，写真順は維持する．
feature row は「その行の写真が大きめに見える」だけで，真の packing や Masonry にはしない．

### 案 C: bounded crop を許す

写真の元の縦横比を厳密に守ると，同じ縦横比の写真は同じ形になりやすい．
よりデザイン寄りにするなら，表示枠の縦横比を元画像から少しだけずらしてよい．

```txt
cropTolerance = 0.12
```

例えば，4:3 の写真を少し横長または少し正方形寄りに表示する．
画像は `object-fit: cover` でクロップする．
ただし，人物写真などでは不自然な切れ方が起きる可能性があるため，今回は必須ではない．

## `PhotoGrid.astro` 側の指示

`PhotoGrid.astro` では，`buildBalancedJustifiedLayout` を使う．
固定の `rowHeight` ではなく，`targetRowHeight` / `minRowHeight` / `maxRowHeight` を渡す．

```ts
const layout = buildBalancedJustifiedLayout(photos, {
  containerWidth: 1380,
  gap: 14,
  targetRowHeight: 220,
  minRowHeight: 170,
  maxRowHeight: 290,
  minTileWidth: 150,
  maxTileWidth: 520,
  minItemsPerRow: 2,
  maxItemsPerRow: 6,
  avoidSameRowLengthPenalty: 90,
  avoidSameHeightPenalty: 40,
});
```

## 可能なら ResizeObserver を入れる

Astro 側で `containerWidth: 1380` を固定すると，実際の表示幅とずれて不自然になる場合がある．
可能なら，依存を追加せずに小さなクライアントスクリプトで `ResizeObserver` を使い，実際の `.photo-grid` 幅に合わせて再計算する．

ただし，実装が大きくなる場合は，今回は固定幅でもよい．
その場合でも，`containerWidth` は `.photo-grid` の `max-width` と一致させる．

## 完了条件

```txt
- デスクトップで 5 カラム固定グリッドのように見えない
- 行ごとに高さまたは枚数に自然な変化がある
- 写真の順序は維持されている
- 右端に極端に細い写真が出ない
- 写真の縦横比に応じて幅が変わる
- デスクトップでは CSS Grid ではなく absolute layout で配置している
- モバイルでは CSS Grid fallback でよい
- 追加依存は入れない
- Go 側は変更しない
```

## 実装後の報告内容

実装後，次を報告する．

```txt
- 変更したファイル一覧
- グリッドっぽく見えていた原因
- 採用したレイアウトアルゴリズム
- 同じ枚数の行が続きすぎないようにした方法
- 右端が細くならないようにした条件
- モバイルでの fallback
```

補足として，入力写真がほぼ全部 4:3 や 16:9 に揃っている場合，縦横比を尊重する限り，完全にはグリッド感を消せません．その場合は，「行高にリズムをつける」か，「feature row」を入れる指示を Codex に明示するとかなり改善します．
