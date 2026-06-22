# albam API Document

## 概要

`albam` の API は，Go 製 CLI / サーバーと Astro 製テーマの間で共有するインターフェースです．

Go 側は，アルバム，写真，タグ，画像配信，サムネイル配信を担当します．
Astro 側は，この API から取得したデータをもとに，トップページ，各アルバムページ，写真詳細 UI を描画します．

本番運用では，`albam serve` が API と静的ファイル配信をまとめて担当します．

```txt
Browser
  -> Caddy / nginx
  -> albam serve
      -> static theme files
      -> /api/*
      -> /media/*
```

## 基本方針

API は `/api` 以下に配置します．
画像ファイルそのものは `/media` 以下に配置します．

```txt
/api/*     JSON API
/media/*   image delivery
```

レスポンス形式は JSON です．
日時は ISO 8601 形式の文字列で返します．
ID は URL に使いやすい slug 形式を基本とします．

例:

```json
{
  "id": "weekend-trip",
  "title": "Weekend trip",
  "created_at": "2026-05-18T10:30:00+09:00"
}
```

## 共通レスポンス

### 成功レスポンス

単一リソースの場合:

```json
{
  "album": {
    "id": "weekend-trip",
    "title": "Weekend trip"
  }
}
```

一覧リソースの場合:

```json
{
  "albums": [],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 120,
    "has_next": true
  }
}
```

### エラーレスポンス

```json
{
  "error": {
    "code": "album_not_found",
    "message": "album not found"
  }
}
```

代表的なステータスコードは次の通りです．

| status | 用途                 |
| -----: | -------------------- |
|  `200` | 成功                 |
|  `400` | 不正なリクエスト     |
|  `404` | リソースが存在しない |
|  `500` | サーバー内部エラー   |

## Album

### Album object

```ts
type Album = {
  id: string;
  title: string;
  description: string;
  date: string | null;
  created_at: string;
  updated_at: string;
  photo_count: number;
  latest_month: string | null;
  oldest_taken_at: string | null;
  newest_taken_at: string | null;
  cover_photo_id: string | null;
  visibility: "public" | "private";
  tags: Tag[];
  links: {
    self: string;
    photos: string;
    cover: string | null;
  };
};
```

例:

```json
{
  "id": "weekend-trip",
  "title": "Weekend trip",
  "description": "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。",
  "date": "2026-05-18",
  "created_at": "2026-05-18T10:30:00+09:00",
  "updated_at": "2026-05-23T18:10:00+09:00",
  "photo_count": 48,
  "latest_month": "2026/05",
  "oldest_taken_at": "2026-05-18T15:20:00+09:00",
  "newest_taken_at": "2026-05-23T18:10:00+09:00",
  "cover_photo_id": "img-001",
  "visibility": "public",
  "tags": [
    {
      "id": "travel",
      "name": "travel"
    },
    {
      "id": "sea",
      "name": "sea"
    }
  ],
  "links": {
    "self": "/api/albums/weekend-trip",
    "photos": "/api/albums/weekend-trip/media",
    "cover": "/media/img-001/thumb"
  }
}
```

## GET /api/albums

アルバム一覧を取得します．
トップページで使用します．

### Query parameters

| name     | type   |     default | description          |
| -------- | ------ | ----------: | -------------------- |
| `limit`  | number |        `50` | 取得件数             |
| `offset` | number |         `0` | 開始位置             |
| `tag`    | string |        なし | 指定タグで絞り込み   |
| `q`      | string |        なし | タイトル・説明文検索 |
| `sort`   | string | `date_asc` | 並び順               |

`sort` は次を想定します．

```txt
date_desc
date_asc
title_asc
title_desc
updated_desc
```

### Response

```json
{
  "albums": [
    {
      "id": "weekend-trip",
      "title": "Weekend trip",
      "description": "友人との週末旅行。",
      "date": "2026-05-18",
      "created_at": "2026-05-18T10:30:00+09:00",
      "updated_at": "2026-05-23T18:10:00+09:00",
      "photo_count": 48,
      "cover_photo_id": "img-001",
      "visibility": "public",
      "tags": [
        {
          "id": "travel",
          "name": "travel"
        }
      ],
      "links": {
        "self": "/api/albums/weekend-trip",
        "photos": "/api/albums/weekend-trip/media",
        "cover": "/media/img-001/thumb"
      }
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 1,
    "has_next": false
  }
}
```

## GET /api/albums/{album_id}

指定したアルバムの詳細情報を取得します．
各アルバムページのヒーロー部分，サイドバー，メタ情報表示で使用します．

### Path parameters

| name       | type   | description |
| ---------- | ------ | ----------- |
| `album_id` | string | アルバム ID |

### Response

```json
{
  "album": {
    "id": "weekend-trip",
    "title": "Weekend trip",
    "description": "友人との週末旅行。海沿いの散歩，カフェ，夕方の空をまとめたアルバムです。",
    "date": "2026-05-18",
    "created_at": "2026-05-18T10:30:00+09:00",
    "updated_at": "2026-05-23T18:10:00+09:00",
    "photo_count": 48,
    "cover_photo_id": "img-001",
    "visibility": "public",
    "tags": [
      {
        "id": "travel",
        "name": "travel"
      },
      {
        "id": "sea",
        "name": "sea"
      }
    ],
    "links": {
      "self": "/api/albums/weekend-trip",
      "photos": "/api/albums/weekend-trip/media",
      "cover": "/media/img-001/thumb"
    }
  }
}
```

### Error

```json
{
  "error": {
    "code": "album_not_found",
    "message": "album not found"
  }
}
```

## Photo

### Photo object

```ts
type Photo = {
  id: string;
  album_id: string;
  filename: string;
  title: string | null;
  description: string | null;
  taken_at: string | null;
  width: number | null;
  height: number | null;
  aspect_ratio: number | null;
  gps_latitude: number | null;
  gps_longitude: number | null;
  camera_make: string | null;
  camera_model: string | null;
  lens_make: string | null;
  lens_model: string | null;
  focal_length_mm: number | null;
  focal_length_35mm: number | null;
  aperture_f_number: number | null;
  exposure_time_seconds: number | null;
  iso: number | null;
  orientation: number | null;
  favorite: boolean;
  tags: Tag[];
  links: {
    self: string;
    thumb: string;
    preview: string;
    original: string;
  };
};
```

例:

```json
{
  "id": "img-001",
  "album_id": "weekend-trip",
  "filename": "IMG_001.jpg",
  "title": "Sea side",
  "description": null,
  "taken_at": "2026-05-18T15:20:00+09:00",
  "width": 4032,
  "height": 3024,
  "aspect_ratio": 1.3333,
  "favorite": true,
  "tags": [
    {
      "id": "sea",
      "name": "sea"
    }
  ],
  "links": {
    "self": "/api/media/img-001",
    "thumb": "/media/img-001/thumb",
    "preview": "/media/img-001/preview",
    "original": "/media/img-001/original"
  }
}
```

## GET /api/albums/{album_id}/media

指定したアルバムに含まれる写真一覧を取得します．
各アルバムページの写真グリッドで使用します．

### Query parameters

| name       | type    |        default | description        |
| ---------- | ------- | -------------: | ------------------ |
| `limit`    | number  |          `100` | 取得件数           |
| `offset`   | number  |            `0` | 開始位置           |
| `tag`      | string  |           なし | 指定タグで絞り込み |
| `favorite` | boolean |           なし | お気に入りのみ取得 |
| `sort`     | string  | `taken_at_asc` | 並び順             |

`sort` は次を想定します．

```txt
taken_at_asc
taken_at_desc
filename_asc
filename_desc
```

### Response

```json
{
  "photos": [
    {
      "id": "img-001",
      "album_id": "weekend-trip",
      "filename": "IMG_001.jpg",
      "title": "Sea side",
      "description": null,
      "taken_at": "2026-05-18T15:20:00+09:00",
      "width": 4032,
      "height": 3024,
      "aspect_ratio": 1.3333,
      "favorite": true,
      "tags": [
        {
          "id": "sea",
          "name": "sea"
        }
      ],
      "links": {
        "self": "/api/media/img-001",
        "thumb": "/media/img-001/thumb",
        "preview": "/media/img-001/preview",
        "original": "/media/img-001/original"
      }
    }
  ],
  "pagination": {
    "limit": 100,
    "offset": 0,
    "total": 48,
    "has_next": false
  }
}
```

## GET /api/media/{photo_id}

指定した写真の詳細情報を取得します．
写真詳細モーダルやライトボックスで使用します．

### Response

```json
{
  "photo": {
    "id": "img-001",
    "album_id": "weekend-trip",
    "filename": "IMG_001.jpg",
    "title": "Sea side",
    "description": null,
    "taken_at": "2026-05-18T15:20:00+09:00",
    "width": 4032,
    "height": 3024,
    "aspect_ratio": 1.3333,
    "favorite": true,
    "tags": [
      {
        "id": "sea",
        "name": "sea"
      }
    ],
    "links": {
      "self": "/api/media/img-001",
      "thumb": "/media/img-001/thumb",
      "preview": "/media/img-001/preview",
      "original": "/media/img-001/original"
    }
  }
}
```

## Tag

### Tag object

```ts
type Tag = {
  id: string;
  name: string;
  photo_count?: number;
  album_count?: number;
};
```

## GET /api/tags

タグ一覧を取得します．
トップページやフィルタ UI で使用します．

### Response

```json
{
  "tags": [
    {
      "id": "travel",
      "name": "travel",
      "album_count": 3,
      "photo_count": 128
    },
    {
      "id": "sea",
      "name": "sea",
      "album_count": 1,
      "photo_count": 24
    }
  ]
}
```

## GET /media/{photo_id}/thumb

写真のサムネイル画像を取得します．
写真グリッドで使用します．

### Path parameters

| name       | type   | description |
| ---------- | ------ | ----------- |
| `photo_id` | string | 写真 ID     |

### Response

WebP画像を返します．

```txt
Content-Type: image/webp
Cache-Control: public, max-age=31536000, immutable
```

## GET /media/{photo_id}/preview

写真のプレビュー画像を取得します．
ライトボックスや詳細ページで使用します．

```txt
Content-Type: image/webp
Cache-Control: public, max-age=31536000, immutable
```

## GET /media/{photo_id}/original

オリジナル画像を取得します．
ダウンロードや高解像度表示で使用します．

```txt
Content-Type: 元画像のMIMEタイプ
Cache-Control: private, max-age=3600
```

オリジナル画像の公開可否は設定で制御します．

```toml
[media]
allow_original_download = false
```

`allow_original_download = false` の場合，このエンドポイントは `403` を返します．

```json
{
  "error": {
    "code": "original_download_disabled",
    "message": "original download is disabled"
  }
}
```

## Frontend usage

Astro テーマ側では，以下のように API を使います．

### トップページ

使用エンドポイント:

```txt
GET /api/albums
GET /api/tags
```

用途:

```txt
/api/albums -> アルバムグリッド
/api/tags   -> フィルタチップ
```

### 各アルバムページ

使用エンドポイント:

```txt
GET /api/albums/{album_id}
GET /api/albums/{album_id}/media
```

用途:

```txt
/api/albums/{album_id}       -> ヒーロー，説明，タグ，サイドバー
/api/albums/{album_id}/media -> 写真グリッド
```

### 写真詳細モーダル

使用エンドポイント:

```txt
GET /api/media/{photo_id}
GET /media/{photo_id}/preview
```

用途:

```txt
/api/media/{photo_id}       -> メタデータ
/media/{photo_id}/preview   -> 表示画像
/media/{photo_id}/original  -> ダウンロード
```

## TypeScript types

フロントエンドでは，`src/lib/api.ts` などに次の型を定義します．

```ts
export type Tag = {
  id: string;
  name: string;
  photo_count?: number;
  album_count?: number;
};

export type Album = {
  id: string;
  title: string;
  description: string;
  date: string | null;
  created_at: string;
  updated_at: string;
  photo_count: number;
  latest_month: string | null;
  oldest_taken_at: string | null;
  newest_taken_at: string | null;
  cover_photo_id: string | null;
  visibility: "public" | "private";
  tags: Tag[];
  links: {
    self: string;
    photos: string;
    cover: string | null;
  };
};

export type Photo = {
  id: string;
  album_id: string;
  filename: string;
  title: string | null;
  description: string | null;
  taken_at: string | null;
  width: number | null;
  height: number | null;
  aspect_ratio: number | null;
  gps_latitude: number | null;
  gps_longitude: number | null;
  camera_make: string | null;
  camera_model: string | null;
  lens_make: string | null;
  lens_model: string | null;
  focal_length_mm: number | null;
  focal_length_35mm: number | null;
  aperture_f_number: number | null;
  exposure_time_seconds: number | null;
  iso: number | null;
  orientation: number | null;
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

export type AlbumsResponse = {
  albums: Album[];
  pagination: Pagination;
};

export type AlbumResponse = {
  album: Album;
};

export type PhotosResponse = {
  photos: Photo[];
  pagination: Pagination;
};

export type PhotoResponse = {
  photo: Photo;
};

export type TagsResponse = {
  tags: Tag[];
};

export type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
  };
};
```

## Go structs

バックエンドでは，対応する構造体を次のように定義します．

```go
type Tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PhotoCount *int   `json:"photo_count,omitempty"`
	AlbumCount *int   `json:"album_count,omitempty"`
}

type Album struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Date         *string `json:"date"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	PhotoCount   int    `json:"photo_count"`
	LatestMonth  *string `json:"latest_month"`
	OldestTakenAt *string `json:"oldest_taken_at"`
	NewestTakenAt *string `json:"newest_taken_at"`
	CoverPhotoID *string `json:"cover_photo_id"`
	Visibility   string `json:"visibility"`
	Tags         []Tag  `json:"tags"`
	Links        AlbumLinks `json:"links"`
}

type AlbumLinks struct {
	Self   string  `json:"self"`
	Photos string  `json:"photos"`
	Cover  *string `json:"cover"`
}

type Photo struct {
	ID          string `json:"id"`
	AlbumID     string `json:"album_id"`
	Filename    string `json:"filename"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	TakenAt     *string `json:"taken_at"`
	Width       *int    `json:"width"`
	Height      *int    `json:"height"`
	AspectRatio *float64 `json:"aspect_ratio"`
	GPSLatitude  *float64 `json:"gps_latitude"`
	GPSLongitude *float64 `json:"gps_longitude"`
	CameraMake   *string `json:"camera_make"`
	CameraModel  *string `json:"camera_model"`
	LensMake     *string `json:"lens_make"`
	LensModel    *string `json:"lens_model"`
	FocalLengthMM       *float64 `json:"focal_length_mm"`
	FocalLength35mm     *int     `json:"focal_length_35mm"`
	ApertureFNumber     *float64 `json:"aperture_f_number"`
	ExposureTimeSeconds *float64 `json:"exposure_time_seconds"`
	ISO                 *int     `json:"iso"`
	Orientation         *int     `json:"orientation"`
	Favorite    bool    `json:"favorite"`
	Tags        []Tag   `json:"tags"`
	Links       PhotoLinks `json:"links"`
}

type PhotoLinks struct {
	Self     string `json:"self"`
	Thumb    string `json:"thumb"`
	Preview  string `json:"preview"`
	Original string `json:"original"`
}

type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}
```

## MVP で実装するエンドポイント

最初に実装するのは，次の 5 つで十分です．

```txt
GET /api/albums
GET /api/albums/{album_id}
GET /api/albums/{album_id}/media
GET /media/{photo_id}/thumb
GET /media/{photo_id}/preview
```

後回しでよいもの:

```txt
GET /api/media/{photo_id}
GET /api/tags
GET /media/{photo_id}/original
```

この順番なら，Figma で作ったトップページと各アルバムページを先に再現できます．検索，編集，アップロード，認証は後からでよいです．
