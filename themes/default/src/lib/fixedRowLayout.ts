export type FixedRowPhotoInput = {
  id: string;
  width: number;
  height: number;
  aspectRatio?: number;
};

export type FixedRowLayoutOptions = {
  containerWidth: number;
  gap: number;
  rowHeight: number;
  minTileWidth: number;
  maxTileWidth: number;
  minScale: number;
  maxScale: number;
  maxItemsPerRow: number;
  justifyLastRow: boolean;
};

export type FixedRowLayoutItem = {
  id: string;
  x: number;
  y: number;
  width: number;
  height: number;
};

export type FixedRowLayoutResult = {
  width: number;
  height: number;
  items: FixedRowLayoutItem[];
};

function getAspectRatio(photo: FixedRowPhotoInput): number {
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

function getBaseTileWidth(photo: FixedRowPhotoInput, options: FixedRowLayoutOptions): number {
  return clamp(
    options.rowHeight * getAspectRatio(photo),
    options.minTileWidth,
    options.maxTileWidth,
  );
}

function getRowScore(args: {
  rowLength: number;
  rawRowWidth: number;
  scale: number;
  options: FixedRowLayoutOptions;
}): number {
  const { rowLength, rawRowWidth, scale, options } = args;
  let score = Math.abs(rawRowWidth - options.containerWidth);

  if (scale < options.minScale) {
    score += (options.minScale - scale) * options.containerWidth * 3;
  }

  if (scale > options.maxScale) {
    score += (scale - options.maxScale) * options.containerWidth * 3;
  }

  if (rowLength === 1) {
    score += options.containerWidth * 0.2;
  }

  return score;
}

function chooseNextRow(
  photos: FixedRowPhotoInput[],
  startIndex: number,
  options: FixedRowLayoutOptions,
): FixedRowPhotoInput[] {
  let bestRow: FixedRowPhotoInput[] = [photos[startIndex]];
  let bestScore = Number.POSITIVE_INFINITY;
  const maxEnd = Math.min(photos.length, startIndex + Math.max(1, options.maxItemsPerRow));

  for (let end = startIndex + 1; end <= maxEnd; end++) {
    const row = photos.slice(startIndex, end);
    const widths = row.map((photo) => getBaseTileWidth(photo, options));
    const gapWidth = options.gap * Math.max(0, row.length - 1);
    const rawImageWidth = widths.reduce((sum, width) => sum + width, 0);
    const rawRowWidth = rawImageWidth + gapWidth;
    const availableImageWidth = options.containerWidth - gapWidth;
    const scale = rawImageWidth > 0 ? availableImageWidth / rawImageWidth : 1;
    const score = getRowScore({
      rowLength: row.length,
      rawRowWidth,
      scale,
      options,
    });

    if (score < bestScore) {
      bestScore = score;
      bestRow = row;
    }

    if (rawRowWidth > options.containerWidth * 1.35) {
      break;
    }
  }

  return bestRow;
}

function layoutRow(args: {
  row: FixedRowPhotoInput[];
  y: number;
  isLastRow: boolean;
  options: FixedRowLayoutOptions;
  items: FixedRowLayoutItem[];
}): void {
  const { row, y, isLastRow, options, items } = args;
  const baseWidths = row.map((photo) => getBaseTileWidth(photo, options));
  const gapWidth = options.gap * Math.max(0, row.length - 1);
  const availableImageWidth = Math.max(0, options.containerWidth - gapWidth);
  const rawImageWidth = baseWidths.reduce((sum, width) => sum + width, 0);
  let scale = rawImageWidth > 0 ? availableImageWidth / rawImageWidth : 1;

  const shouldFillRow = !isLastRow || options.justifyLastRow;
  if (!shouldFillRow) {
    scale = 1;
  } else {
    scale = clamp(scale, options.minScale, options.maxScale);
  }

  const scaledWidths = baseWidths.map((width) => width * scale);
  const tileWidths = fitRowWidths({
    widths: scaledWidths,
    availableImageWidth,
    fillRow: shouldFillRow,
  });
  let x = 0;

  row.forEach((photo, index) => {
    const width = tileWidths[index];

    items.push({
      id: photo.id,
      x,
      y,
      width,
      height: options.rowHeight,
    });

    x += width + options.gap;
  });
}

function fitRowWidths(args: {
  widths: number[];
  availableImageWidth: number;
  fillRow: boolean;
}): number[] {
  const { widths, availableImageWidth, fillRow } = args;
  const rowWidth = widths.reduce((sum, width) => sum + width, 0);

  if (rowWidth <= 0 || availableImageWidth <= 0) {
    return widths.map(() => 0);
  }

  if (!fillRow && rowWidth <= availableImageWidth) {
    const roundedWidths = widths.map((width) => Math.round(width));
    const roundedRowWidth = roundedWidths.reduce((sum, width) => sum + width, 0);

    if (roundedRowWidth <= availableImageWidth) {
      return roundedWidths;
    }
  }

  const targetWidth = fillRow ? availableImageWidth : Math.min(rowWidth, availableImageWidth);
  const adjustedWidths = widths.map((width) => width * (targetWidth / rowWidth));
  return roundWidthsToTarget(adjustedWidths, Math.floor(targetWidth));
}

function roundWidthsToTarget(widths: number[], targetWidth: number): number[] {
  const floors = widths.map((width) => Math.floor(width));
  let remaining = targetWidth - floors.reduce((sum, width) => sum + width, 0);

  const indexes = widths
    .map((width, index) => ({ index, fraction: width - Math.floor(width) }))
    .sort((a, b) => b.fraction - a.fraction);

  for (let i = 0; i < indexes.length && remaining > 0; i++) {
    floors[indexes[i].index] += 1;
    remaining -= 1;
  }

  return floors;
}

export function buildFixedRowLayout(
  photos: FixedRowPhotoInput[],
  options: FixedRowLayoutOptions,
): FixedRowLayoutResult {
  const containerWidth = Math.max(0, Math.floor(options.containerWidth));
  const normalizedOptions = { ...options, containerWidth };
  const items: FixedRowLayoutItem[] = [];
  let index = 0;
  let y = 0;

  if (containerWidth <= 0 || photos.length === 0) {
    return { width: containerWidth, height: 0, items };
  }

  while (index < photos.length) {
    const row = chooseNextRow(photos, index, normalizedOptions);
    const isLastRow = index + row.length >= photos.length;

    layoutRow({
      row,
      y,
      isLastRow,
      options: normalizedOptions,
      items,
    });

    y += normalizedOptions.rowHeight + normalizedOptions.gap;
    index += row.length;
  }

  return {
    width: containerWidth,
    height: Math.max(0, y - normalizedOptions.gap),
    items,
  };
}
