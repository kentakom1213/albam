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

type RowMetrics = {
  rowHeight: number;
  tileWidths: number[];
  minTileWidth: number;
  maxTileWidth: number;
};

function getAspectRatio(photo: JustifiedPhotoInput): number {
  if (photo.aspectRatio && Number.isFinite(photo.aspectRatio) && photo.aspectRatio > 0) {
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

function getRowMetrics(row: JustifiedPhotoInput[], options: BalancedJustifiedLayoutOptions): RowMetrics {
  const aspectRatios = row.map(getAspectRatio);
  const aspectSum = aspectRatios.reduce((sum, value) => sum + value, 0);
  const gapWidth = options.gap * Math.max(0, row.length - 1);
  const availableWidth = Math.max(0, options.containerWidth - gapWidth);
  const rowHeight = aspectSum > 0 ? availableWidth / aspectSum : options.targetRowHeight;
  const tileWidths = aspectRatios.map((aspectRatio) => rowHeight * aspectRatio);

  return {
    rowHeight,
    tileWidths,
    minTileWidth: Math.min(...tileWidths),
    maxTileWidth: Math.max(...tileWidths),
  };
}

function getBalancedRowScore(args: {
  row: JustifiedPhotoInput[];
  metrics: RowMetrics;
  previousRowLength: number | null;
  previousRowHeight: number | null;
  sameRowLengthRun: number;
  remainingAfterRow: number;
  options: BalancedJustifiedLayoutOptions;
}): number {
  const {
    row,
    metrics,
    previousRowLength,
    previousRowHeight,
    sameRowLengthRun,
    remainingAfterRow,
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

  if (remainingAfterRow > 0 && remainingAfterRow < options.minItemsPerRow) {
    score += options.containerWidth * 0.35;
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

  const minItems = Math.max(1, options.minItemsPerRow);
  const maxItems = Math.max(minItems, Math.min(options.maxItemsPerRow, remaining));
  let bestRow = photos.slice(startIndex, startIndex + minItems);
  let bestScore = Number.POSITIVE_INFINITY;

  for (let rowLength = minItems; rowLength <= maxItems; rowLength++) {
    const row = photos.slice(startIndex, startIndex + rowLength);
    const metrics = getRowMetrics(row, options);
    const score = getBalancedRowScore({
      row,
      metrics,
      previousRowLength,
      previousRowHeight,
      sameRowLengthRun,
      remainingAfterRow: remaining - row.length,
      options,
    });

    if (score < bestScore) {
      bestScore = score;
      bestRow = row;
    }
  }

  return bestRow;
}

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

  let x = 0;

  const items = row.map((photo) => {
    const width = getAspectRatio(photo) * rowHeight;
    const item = {
      id: photo.id,
      x,
      y,
      width,
      height: rowHeight,
    };

    x += width + options.gap;

    return item;
  });

  return {
    height: rowHeight,
    items,
  };
}

export function buildBalancedJustifiedLayout(
  photos: JustifiedPhotoInput[],
  options: BalancedJustifiedLayoutOptions,
): BalancedJustifiedLayoutResult {
  const containerWidth = Math.max(0, Math.floor(options.containerWidth));
  const normalizedOptions = { ...options, containerWidth };
  const items: BalancedJustifiedLayoutItem[] = [];
  let index = 0;
  let y = 0;
  let previousRowLength: number | null = null;
  let previousRowHeight: number | null = null;
  let sameRowLengthRun = 0;

  if (containerWidth <= 0 || photos.length === 0) {
    return { width: containerWidth, height: 0, items };
  }

  while (index < photos.length) {
    const row = chooseNextBalancedRow({
      photos,
      startIndex: index,
      previousRowLength,
      previousRowHeight,
      sameRowLengthRun,
      options: normalizedOptions,
    });
    const isLastRow = index + row.length >= photos.length;
    const layouted = layoutBalancedRow({
      row,
      y,
      isLastRow,
      options: normalizedOptions,
    });

    items.push(...layouted.items);

    if (previousRowLength === row.length) {
      sameRowLengthRun += 1;
    } else {
      sameRowLengthRun = 1;
    }

    previousRowLength = row.length;
    previousRowHeight = layouted.height;
    y += layouted.height + normalizedOptions.gap;
    index += row.length;
  }

  return {
    width: containerWidth,
    height: Math.max(0, y - normalizedOptions.gap),
    items,
  };
}
