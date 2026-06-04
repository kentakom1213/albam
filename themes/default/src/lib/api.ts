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
  tone?: Album["tone"];
};

export type AlbumPhotosResult = {
  photos: Photo[];
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
  return {
    id: photo.id,
    title: photo.title ?? photo.filename,
    albumId: photo.album_id,
    src: resolveAssetUrl(photo.links.thumb),
    previewSrc: resolveAssetUrl(photo.links.preview),
    originalSrc: resolveAssetUrl(photo.links.original),
    tone: tones[index % tones.length],
  };
}

export async function getAlbums(): Promise<Album[]> {
  const body = await request<AlbumsResponse>("albums", { limit: 50, offset: 0 });
  return body.albums.map(toAlbum);
}

export async function getAlbum(albumId: string): Promise<Album> {
  const body = await request<AlbumResponse>(`albums/${albumId}`);
  return toAlbum(body.album);
}

export async function getAlbumPhotosWithPagination(albumId: string): Promise<AlbumPhotosResult> {
  const body = await request<PhotosResponse>(`albums/${albumId}/photos`, {
    limit: 100,
    offset: 0,
  });
  return {
    photos: body.photos.map(toPhoto),
    pagination: body.pagination,
  };
}

export async function getAlbumPhotos(albumId: string): Promise<Photo[]> {
  const body = await getAlbumPhotosWithPagination(albumId);
  return body.photos;
}
