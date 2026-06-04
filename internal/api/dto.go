package api

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

type AlbumsResponse struct {
	Albums     []Album    `json:"albums"`
	Pagination Pagination `json:"pagination"`
}

type AlbumResponse struct {
	Album Album `json:"album"`
}

type PhotosResponse struct {
	Photos     []Photo    `json:"photos"`
	Pagination Pagination `json:"pagination"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
