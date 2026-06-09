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
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Date          *string      `json:"date"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
	PhotoCount    int          `json:"photo_count"`
	LatestMonth   *string      `json:"latest_month"`
	OldestTakenAt *string      `json:"oldest_taken_at"`
	NewestTakenAt *string      `json:"newest_taken_at"`
	CoverPhotoID  *string      `json:"cover_photo_id"`
	Visibility    string       `json:"visibility"`
	Breadcrumbs   []Breadcrumb `json:"breadcrumbs"`
	Links         AlbumLinks   `json:"links"`
}

type AlbumLinks struct {
	Self   string  `json:"self"`
	Photos string  `json:"photos"`
	Cover  *string `json:"cover"`
}

type Photo struct {
	ID                  string     `json:"id"`
	AlbumID             string     `json:"album_id"`
	Filename            string     `json:"filename"`
	Title               *string    `json:"title"`
	Description         *string    `json:"description"`
	TakenAt             *string    `json:"taken_at"`
	Width               *int       `json:"width"`
	Height              *int       `json:"height"`
	AspectRatio         *float64   `json:"aspect_ratio"`
	GPSLatitude         *float64   `json:"gps_latitude"`
	GPSLongitude        *float64   `json:"gps_longitude"`
	CameraMake          *string    `json:"camera_make"`
	CameraModel         *string    `json:"camera_model"`
	LensMake            *string    `json:"lens_make"`
	LensModel           *string    `json:"lens_model"`
	FocalLengthMM       *float64   `json:"focal_length_mm"`
	FocalLength35mm     *int       `json:"focal_length_35mm"`
	ApertureFNumber     *float64   `json:"aperture_f_number"`
	ExposureTimeSeconds *float64   `json:"exposure_time_seconds"`
	ISO                 *int       `json:"iso"`
	Orientation         *int       `json:"orientation"`
	Favorite            bool       `json:"favorite"`
	Links               PhotoLinks `json:"links"`
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

type PhotoResponse struct {
	Photo Photo `json:"photo"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
