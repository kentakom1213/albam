# albam API Document

## 概要

`albam` の API は，Go 製 CLI / サーバーと Astro 製テーマの間で共有するインターフェースです．

Go 側は，アルバム，写真，画像配信，サムネイル配信を担当します．
Astro 側は，この API から取得したデータをもとに，トップページ，各アルバムページ，写真一覧 UI を描画します．

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

アルバム ID と写真 ID は，どちらも外部 API では slug を使います．DB 内部では整数 ID を持っていてもよいですが，URL や JSON には出しません．

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

| status | 用途                           |
| -----: | ------------------------------ |
|  `200` | 成功                           |
|  `400` | 不正なリクエスト               |
|  `403` | 許可されていない操作           |
|  `404` | リソースが存在しない           |
|  `405` | 許可されていない HTTP メソッド |
|  `500` | サーバー内部エラー             |

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
  breadcrumbs: Breadcrumb[];
  links: {
    self: string;
    photos: string;
    cover: string | null;
  };
};
```

`photo_count` は，そのアルバム直下だけでなく，子孫アルバムに含まれる写真も含めた枚数です．

`cover_photo_id` は，そのアルバム配下の代表写真 ID です．アルバム直下に写真がない場合でも，子孫アルバムに写真があれば，その中から代表写真を返します．

### Breadcrumb object

```ts
type Breadcrumb = {
  id: string;
  title: string;
  path: string;
  links: {
    self: string;
  };
};
```

`breadcrumbs` は，現在のアルバムまでの祖先アルバム列です．パンくずリストの表示に使います．

例:

```json
{
  "id": "kochi-2025",
  "title": "11_kochi",
  "description": "",
  "date": null,
  "created_at": "2026-06-03T16:16:26Z",
  "updated_at": "2026-06-03T16:16:26Z",
  "photo_count": 6,
  "cover_photo_id": "x7KpQ2mL9a",
  "visibility": "private",
  "breadcrumbs": [
    {
      "id": "year-2025",
      "title": "2025",
      "path": "2025",
      "links": {
        "self": "/albums/year-2025/"
      }
    },
    {
      "id": "kochi-2025",
      "title": "11_kochi",
      "path": "2025/11_kochi",
      "links": {
        "self": "/albums/kochi-2025/"
      }
    }
  ],
  "links": {
    "self": "/api/albums/kochi-2025",
    "photos": "/api/albums/kochi-2025/media",
    "cover": "/media/x7KpQ2mL9a/thumb"
  }
}
```

## GET /api/albums

アルバム一覧を取得します．
トップページで使用します．

MVP では，写真を 1 枚以上持つアルバムだけを返します．
親アルバムが子孫アルバムに写真を持つ場合も，写真を持つアルバムとして扱います．

### Query parameters

| name     | type   | default | description |
| -------- | ------ | ------: | ----------- |
| `limit`  | number |    `50` | 取得件数    |
| `offset` | number |     `0` | 開始位置    |
| `sort`   | string | `date_asc` | 並び順      |

`sort` は次を指定できます．

```txt
date_desc
date_asc
```

### Response

```json
{
  "albums": [
    {
      "id": "year-2025",
      "title": "2025",
      "description": "",
      "date": null,
      "created_at": "2026-06-03T16:16:26Z",
      "updated_at": "2026-06-03T16:16:26Z",
      "photo_count": 12,
      "latest_month": "2025/11",
      "oldest_taken_at": "2025-11-01T10:20:00Z",
      "newest_taken_at": "2025-11-14T07:56:39Z",
      "cover_photo_id": "x7KpQ2mL9a",
      "visibility": "private",
      "breadcrumbs": [],
      "links": {
        "self": "/api/albums/year-2025",
        "photos": "/api/albums/year-2025/media",
        "cover": "/media/x7KpQ2mL9a/thumb"
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

一覧レスポンスでは，`breadcrumbs` は空配列でもよいです．トップページでパンくずを使わないためです．実装を揃えたい場合は，ここでも各アルバムの breadcrumbs を返して構いません．

## GET /api/albums/{album_id}

指定したアルバムの詳細情報を取得します．
各アルバムページのヒーロー部分，パンくずリスト，メタ情報表示で使用します．

### Path parameters

| name       | type   | description |
| ---------- | ------ | ----------- |
| `album_id` | string | アルバム ID |

### Response

```json
{
  "album": {
    "id": "kochi-2025",
    "title": "11_kochi",
    "description": "",
    "date": null,
    "created_at": "2026-06-03T16:16:26Z",
    "updated_at": "2026-06-03T16:16:26Z",
    "photo_count": 6,
    "cover_photo_id": "x7KpQ2mL9a",
    "visibility": "private",
    "breadcrumbs": [
      {
        "id": "year-2025",
        "title": "2025",
        "path": "2025",
        "links": {
          "self": "/albums/year-2025/"
        }
      },
      {
        "id": "kochi-2025",
        "title": "11_kochi",
        "path": "2025/11_kochi",
        "links": {
          "self": "/albums/kochi-2025/"
        }
      }
    ],
    "links": {
      "self": "/api/albums/kochi-2025",
      "photos": "/api/albums/kochi-2025/media",
      "cover": "/media/x7KpQ2mL9a/thumb"
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
  "id": "x7KpQ2mL9a",
  "album_id": "kochi-2025",
  "filename": "PXL_20251114_075639698.jpg",
  "title": null,
  "description": null,
  "taken_at": null,
  "width": null,
  "height": null,
  "aspect_ratio": null,
  "favorite": false,
  "links": {
    "self": "/api/media/x7KpQ2mL9a",
    "thumb": "/media/x7KpQ2mL9a/thumb",
    "preview": "/media/x7KpQ2mL9a/preview",
    "original": "/media/x7KpQ2mL9a/original"
  }
}
```

## GET /api/albums/{album_id}/media

指定したアルバム配下の写真一覧を取得します．
各アルバムページの写真グリッドで使用します．

このエンドポイントは，指定したアルバム直下の写真だけでなく，子孫アルバムに含まれる写真も返します．

たとえば，次のようなディレクトリ構造があるとします．

```txt
albums/
└── 2025/
    ├── 11_kochi/
    │   ├── IMG_001.jpg
    │   └── IMG_002.jpg
    └── 12_tokyo/
        └── IMG_003.jpg
```

`2025` に対応する `album_id` でこの API を呼ぶと，`11_kochi` と `12_tokyo` 配下の写真も含めて返します．

### Query parameters

| name     | type   | default | description |
| -------- | ------ | ------: | ----------- |
| `limit`  | number |   `100` | 取得件数    |
| `offset` | number |     `0` | 開始位置    |
| `sort`   | string | `taken_at_asc` | 並び順 |

`sort` は次を指定できます．

```txt
taken_at_desc
taken_at_asc
```

### Response

```json
{
  "photos": [
    {
      "id": "x7KpQ2mL9a",
      "album_id": "kochi-2025",
      "filename": "PXL_20251114_075639698.jpg",
      "title": null,
      "description": null,
      "taken_at": null,
      "width": null,
      "height": null,
      "aspect_ratio": null,
      "favorite": false,
      "links": {
        "self": "/api/media/x7KpQ2mL9a",
        "thumb": "/media/x7KpQ2mL9a/thumb",
        "preview": "/media/x7KpQ2mL9a/preview",
        "original": "/media/x7KpQ2mL9a/original"
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

MVP では未実装でもよいです．写真一覧だけでグリッド表示できるためです．

### Response

```json
{
  "photo": {
    "id": "x7KpQ2mL9a",
    "album_id": "kochi-2025",
    "filename": "PXL_20251114_075639698.jpg",
    "title": null,
    "description": null,
    "taken_at": null,
    "width": null,
    "height": null,
    "aspect_ratio": null,
    "favorite": false,
    "links": {
      "self": "/api/media/x7KpQ2mL9a",
      "thumb": "/media/x7KpQ2mL9a/thumb",
      "preview": "/media/x7KpQ2mL9a/preview",
      "original": "/media/x7KpQ2mL9a/original"
    }
  }
}
```

### Error

```json
{
  "error": {
    "code": "photo_not_found",
    "message": "photo not found"
  }
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

## GET /media/{photo_id}/original

オリジナル画像を取得します．
ダウンロードや高解像度表示で使用します．

### Path parameters

| name       | type   | description |
| ---------- | ------ | ----------- |
| `photo_id` | string | 写真 ID     |

### Response

元画像をそのまま返します．
`Content-Type` は元画像の形式に応じます．

```txt
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
```

用途:

```txt
/api/albums -> アルバムグリッド
```

### 各アルバムページ

使用エンドポイント:

```txt
GET /api/albums/{album_id}
GET /api/albums/{album_id}/media
```

用途:

```txt
/api/albums/{album_id}       -> ヒーロー，説明，パンくず，サイドバー
/api/albums/{album_id}/media -> 写真グリッド
```

### 写真詳細モーダル

使用エンドポイント:

```txt
GET /api/media/{photo_id}
GET /media/{photo_id}/preview
GET /media/{photo_id}/original
```

用途:

```txt
/api/media/{photo_id}        -> メタデータ
/media/{photo_id}/preview    -> 表示画像
/media/{photo_id}/original   -> ダウンロード
```

MVP では，写真詳細 API を使わず，`GET /api/albums/{album_id}/media` のレスポンスだけで写真グリッドと簡易プレビューを作ってもよいです．

## TypeScript types

フロントエンドでは，`src/lib/api.ts` などに次の型を定義します．

```ts
export type Breadcrumb = {
  id: string;
  title: string;
  path: string;
  links: {
    self: string;
  };
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
  breadcrumbs: Breadcrumb[];
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
type Breadcrumb struct {
	ID    string          `json:"id"`
	Title string          `json:"title"`
	Path  string          `json:"path"`
	Links BreadcrumbLinks `json:"links"`
}

type BreadcrumbLinks struct {
	Self string `json:"self"`
}

type Album struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Date         *string      `json:"date"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	PhotoCount   int          `json:"photo_count"`
	LatestMonth  *string      `json:"latest_month"`
	OldestTakenAt *string     `json:"oldest_taken_at"`
	NewestTakenAt *string     `json:"newest_taken_at"`
	CoverPhotoID *string      `json:"cover_photo_id"`
	Visibility   string       `json:"visibility"`
	Breadcrumbs  []Breadcrumb `json:"breadcrumbs"`
	Links        AlbumLinks   `json:"links"`
}

type AlbumLinks struct {
	Self   string  `json:"self"`
	Photos string  `json:"photos"`
	Cover  *string `json:"cover"`
}

type Photo struct {
	ID          string     `json:"id"`
	AlbumID     string     `json:"album_id"`
	Filename    string     `json:"filename"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	TakenAt     *string    `json:"taken_at"`
	Width       *int       `json:"width"`
	Height      *int       `json:"height"`
	AspectRatio *float64   `json:"aspect_ratio"`
	GPSLatitude  *float64   `json:"gps_latitude"`
	GPSLongitude *float64   `json:"gps_longitude"`
	CameraMake   *string    `json:"camera_make"`
	CameraModel  *string    `json:"camera_model"`
	LensMake     *string    `json:"lens_make"`
	LensModel    *string    `json:"lens_model"`
	FocalLengthMM       *float64 `json:"focal_length_mm"`
	FocalLength35mm     *int     `json:"focal_length_35mm"`
	ApertureFNumber     *float64 `json:"aperture_f_number"`
	ExposureTimeSeconds *float64 `json:"exposure_time_seconds"`
	ISO                 *int     `json:"iso"`
	Orientation         *int     `json:"orientation"`
	Favorite    bool       `json:"favorite"`
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

最初に安定させるのは，次の 6 つです．

```txt
GET /api/albums
GET /api/albums/{album_id}
GET /api/albums/{album_id}/media
GET /media/{photo_id}/thumb
GET /media/{photo_id}/preview
GET /media/{photo_id}/original
```

`GET /api/media/{photo_id}` は，写真詳細モーダルを本格的に作る段階まで後回しでよいです．

## 廃止したもの

タグ機能を廃止したため，次は API 仕様から削除します．

```txt
Tag object
GET /api/tags
Album.tags
Photo.tags
GET /api/albums?tag=...
GET /api/albums/{album_id}/media?tag=...
TagsResponse
```

検索を入れる場合も，タグ検索ではなく，アルバムタイトル，アルバム説明，ファイル名を対象にします．
