export type ApiBreadcrumb = {
  id: string;
  title: string;
  path: string;
  links: {
    self: string;
  };
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
  breadcrumbs: ApiBreadcrumb[];
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

export type Breadcrumb = {
  id: string;
  title: string;
  path: string;
  href: string;
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
  breadcrumbs: Breadcrumb[];
  coverSrc?: string;
  tone?: "peach" | "linen" | "mint" | "sky" | "lilac" | "lemon";
};

export type Photo = {
  id: string;
  title: string;
  albumId: string;
  src?: string;
  previewSrc?: string;
  originalSrc?: string;
  width: number;
  height: number;
  aspectRatio: number;
  favorite?: boolean;
  tone?: Album["tone"];
};

export type AlbumPhotosResult = {
  photos: Photo[];
  pagination: Pagination;
};

export type ApiConfig = {
  enableOriginalDownload: boolean;
};

export type AlbumsResult = {
  albums: Album[];
  pagination: Pagination;
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

type PhotoResponse = {
  photo: ApiPhoto;
};

type ConfigResponse = {
  enable_original_download: boolean;
};

type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
  };
};

const apiBaseUrl =
  import.meta.env.ALBAM_API_BASE_URL ??
  import.meta.env.PUBLIC_ALBAM_API_BASE_URL ??
  "/api";

const tones: NonNullable<Album["tone"]>[] = ["peach", "linen", "mint", "sky", "lilac", "lemon"];

function runtimeOrigin() {
  return window.location.origin;
}

function apiBase() {
  return new URL(apiBaseUrl.endsWith("/") ? apiBaseUrl : `${apiBaseUrl}/`, runtimeOrigin());
}

async function request<T>(path: string, params?: Record<string, string | number | boolean>) {
  const url = new URL(path.replace(/^\//, ""), apiBase());

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

  if (path.startsWith("/")) {
    return path;
  }

  return new URL(path, runtimeOrigin()).toString();
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

function toBreadcrumb(breadcrumb: ApiBreadcrumb): Breadcrumb {
  return {
    id: breadcrumb.id,
    title: breadcrumb.title,
    path: breadcrumb.path,
    href: breadcrumb.links.self,
  };
}

export function toAlbum(album: ApiAlbum, index = 0): Album {
  return {
    id: album.id,
    title: album.title,
    kind: "PHOTO ALBUM",
    description: album.description,
    photoCount: album.photo_count,
    createdAt: formatDate(album.created_at),
    updatedAt: formatDate(album.updated_at),
    size: "-",
    visibility: album.visibility,
    breadcrumbs: album.breadcrumbs.map(toBreadcrumb),
    coverSrc: resolveAssetUrl(album.links.cover),
    tone: tones[index % tones.length],
  };
}

export function toPhoto(photo: ApiPhoto, index = 0): Photo {
  const width = photo.width && photo.width > 0 ? photo.width : 1000;
  const height = photo.height && photo.height > 0 ? photo.height : 1000;
  const aspectRatio =
    photo.aspect_ratio && photo.aspect_ratio > 0 ? photo.aspect_ratio : width / height;

  return {
    id: photo.id,
    title: photo.title ?? photo.filename,
    albumId: photo.album_id,
    src: resolveAssetUrl(photo.links.thumb),
    previewSrc: resolveAssetUrl(photo.links.preview),
    originalSrc: resolveAssetUrl(photo.links.original),
    width,
    height,
    aspectRatio,
    favorite: photo.favorite,
    tone: tones[index % tones.length],
  };
}

export async function getAlbumsWithPagination(params: { limit?: number; offset?: number } = {}): Promise<AlbumsResult> {
  const offset = params.offset ?? 0;
  const body = await request<AlbumsResponse>("albums", {
    limit: params.limit ?? 50,
    offset,
  });

  return {
    albums: body.albums.map((album, index) => toAlbum(album, offset + index)),
    pagination: body.pagination,
  };
}

export async function getAlbums(): Promise<Album[]> {
  const body = await getAlbumsWithPagination();
  return body.albums;
}

export async function getAlbum(albumId: string): Promise<Album> {
  const body = await request<AlbumResponse>(`albums/${albumId}`);
  return toAlbum(body.album);
}

export async function getAlbumPhotosWithPagination(
  albumId: string,
  params: { limit?: number; offset?: number } = {},
): Promise<AlbumPhotosResult> {
  const offset = params.offset ?? 0;
  const body = await request<PhotosResponse>(`albums/${albumId}/photos`, {
    limit: params.limit ?? 100,
    offset,
  });

  return {
    photos: body.photos.map((photo, index) => toPhoto(photo, offset + index)),
    pagination: body.pagination,
  };
}

export async function getAlbumPhotos(albumId: string): Promise<Photo[]> {
  const body = await getAlbumPhotosWithPagination(albumId);
  return body.photos;
}

export async function getPhoto(photoId: string): Promise<Photo> {
  const body = await request<PhotoResponse>(`photos/${photoId}`);
  return toPhoto(body.photo);
}

export async function getApiConfig(): Promise<ApiConfig> {
  const body = await request<ConfigResponse>("config");

  return {
    enableOriginalDownload: body.enable_original_download,
  };
}
