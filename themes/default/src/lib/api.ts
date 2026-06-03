export type Tag = {
  id: string;
  name: string;
  photo_count?: number;
  album_count?: number;
};

export type ApiAlbum = {
  id: string;
  title: string;
  description: string;
  date: string | null;
  created_at: string;
  updated_at: string;
  photo_count: number;
  cover_photo_id: string | null;
  visibility: "public" | "private";
  tags: Tag[];
  links: {
    self: string;
    photos: string;
    cover: string | null;
  };
};

export type ApiPhoto = {
  id: string;
  album_id: string;
  filename: string;
  title: string | null;
  description: string | null;
  taken_at: string | null;
  width: number | null;
  height: number | null;
  aspect_ratio: number | null;
  favorite: boolean;
  tags: Tag[];
  links: {
    self: string;
    thumb: string;
    preview: string;
    original: string;
  };
};

export type Pagination = {
  limit: number;
  offset: number;
  total: number;
  has_next: boolean;
};

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
  coverSrc?: string;
  tone?: "peach" | "linen" | "mint" | "sky" | "lilac" | "lemon";
};

export type Photo = {
  id: string;
  title: string;
  albumId: string;
  src?: string;
  previewSrc?: string;
  tone?: Album["tone"];
};

type AlbumsResponse = {
  albums: ApiAlbum[];
  pagination: Pagination;
};

type AlbumResponse = {
  album: ApiAlbum;
};

type PhotosResponse = {
  photos: ApiPhoto[];
  pagination: Pagination;
};

type TagsResponse = {
  tags: Tag[];
};

type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
  };
};

export const apiBaseUrl =
  import.meta.env.ALBAM_API_BASE_URL ??
  import.meta.env.PUBLIC_ALBAM_API_BASE_URL ??
  "http://localhost:8080/api";

const strictApi = import.meta.env.ALBAM_API_STRICT === "1";
const apiBase = new URL(apiBaseUrl.endsWith("/") ? apiBaseUrl : `${apiBaseUrl}/`);
const apiOrigin = apiBase.origin;

const tones: NonNullable<Album["tone"]>[] = ["peach", "linen", "mint", "sky", "lilac", "lemon"];

async function request<T>(path: string, params?: Record<string, string | number | boolean>) {
  const url = new URL(path.replace(/^\//, ""), apiBase);

  for (const [key, value] of Object.entries(params ?? {})) {
    url.searchParams.set(key, String(value));
  }

  const response = await fetch(url);

  if (!response.ok) {
    let message = `API request failed: ${response.status} ${response.statusText}`;

    try {
      const body = (await response.json()) as ApiErrorResponse;
      if (body.error?.message) {
        message = body.error.message;
      }
    } catch {
      // Keep the HTTP status message when the body is not an API error JSON.
    }

    throw new Error(message);
  }

  return (await response.json()) as T;
}

function resolveAssetUrl(path: string | null | undefined) {
  if (!path) {
    return undefined;
  }

  return new URL(path, apiOrigin).toString();
}

function formatDate(value: string | null | undefined, fallback = "-") {
  if (!value) {
    return fallback;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
  }).format(date);
}

export function formatCompactDate(value: string | null | undefined) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  })
    .format(date)
    .replace(/\//g, ".");
}

export function formatCompactMonth(value: string | null | undefined) {
  const compact = formatCompactDate(value);
  return compact === "-" ? compact : compact.split(".").slice(0, 2).join(".");
}

function albumKind(album: ApiAlbum) {
  const primaryTag = album.tags[0]?.name;
  return primaryTag ? `${primaryTag.toUpperCase()} ALBUM` : "PHOTO ALBUM";
}

function toMockAlbum(album: (typeof mockAlbums)[number]): Album {
  return { ...album };
}

function toMockPhoto(photo: (typeof mockPhotos)[number]): Photo {
  return { ...photo };
}

export function toAlbum(album: ApiAlbum, index = 0): Album {
  return {
    id: album.id,
    title: album.title,
    kind: albumKind(album),
    description: album.description,
    photoCount: album.photo_count,
    createdAt: formatDate(album.created_at),
    updatedAt: formatDate(album.updated_at),
    size: "-",
    visibility: album.visibility,
    tags: album.tags.map((tag) => tag.name),
    coverSrc: resolveAssetUrl(album.links.cover),
    tone: tones[index % tones.length],
  };
}

export function toPhoto(photo: ApiPhoto, index = 0): Photo {
  return {
    id: photo.id,
    title: photo.title ?? photo.filename,
    albumId: photo.album_id,
    src: resolveAssetUrl(photo.links.thumb),
    previewSrc: resolveAssetUrl(photo.links.preview),
    tone: tones[index % tones.length],
  };
}

export async function getAlbums(): Promise<Album[]> {
  try {
    const body = await request<AlbumsResponse>("albums", { limit: 50, offset: 0 });
    return body.albums.map(toAlbum);
  } catch (error) {
    if (strictApi) {
      throw error;
    }

    return mockAlbums.map(toMockAlbum);
  }
}

export async function getAlbum(albumId: string): Promise<Album> {
  try {
    const body = await request<AlbumResponse>(`albums/${albumId}`);
    return toAlbum(body.album);
  } catch (error) {
    if (strictApi) {
      throw error;
    }

    const album = mockAlbums.find((mockAlbum) => mockAlbum.id === albumId);
    if (!album) {
      throw error;
    }

    return toMockAlbum(album);
  }
}

export async function getAlbumPhotos(albumId: string): Promise<Photo[]> {
  try {
    const body = await request<PhotosResponse>(`albums/${albumId}/photos`, {
      limit: 100,
      offset: 0,
    });
    return body.photos.map(toPhoto);
  } catch (error) {
    if (strictApi) {
      throw error;
    }

    return mockPhotos.filter((photo) => photo.albumId === albumId).map(toMockPhoto);
  }
}

export async function getTags() {
  try {
    const body = await request<TagsResponse>("tags");
    return body.tags;
  } catch (error) {
    if (strictApi) {
      throw error;
    }

    const tagNames = [...new Set(mockAlbums.flatMap((album) => album.tags))];
    return tagNames.map((name) => ({ id: name, name }));
  }
}
import { albums as mockAlbums, photos as mockPhotos } from "../data/mock";
