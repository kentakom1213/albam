import type { Album, Asset, Tag } from "./types";

export function albumHref(album: Pick<Album, "slug">): string {
  return `/albums/${encodeURIComponent(album.slug)}`;
}

export function assetHref(asset: Pick<Asset, "id">): string {
  return `/assets/${encodeURIComponent(asset.id)}`;
}

export function tagHref(tag: Pick<Tag, "slug">): string {
  return `/tags/${encodeURIComponent(tag.slug)}`;
}
