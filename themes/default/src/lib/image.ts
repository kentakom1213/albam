import type { Asset, AssetSummary } from "./types";

export function getThumbnailUrl(asset: Asset | AssetSummary): string {
  return asset.variants.thumb;
}

export function getPreviewUrl(asset: Asset | AssetSummary): string {
  return asset.variants.large ?? asset.variants.medium ?? asset.variants.thumb;
}

export function getAspectRatio(asset: Asset): string {
  if (asset.width <= 0 || asset.height <= 0) {
    return "1 / 1";
  }

  return `${asset.width} / ${asset.height}`;
}
