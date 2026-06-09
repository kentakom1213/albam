CREATE TABLE IF NOT EXISTS albums (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	parent_id INTEGER,
	path TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL,
	title TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (parent_id) REFERENCES albums(id)
);

CREATE TABLE IF NOT EXISTS assets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT NOT NULL UNIQUE,
	album_id INTEGER NOT NULL,
	path TEXT NOT NULL UNIQUE,
	filename TEXT NOT NULL,
	ext TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	file_mtime DATETIME NOT NULL,

	width INTEGER,
	height INTEGER,
	taken_at TEXT,
	gps_latitude REAL,
	gps_longitude REAL,
	gps_hiden INTEGER NOT NULL DEFAULT 0,
	camera_make TEXT,
	camera_model TEXT,
	lens_make TEXT,
	lens_model TEXT,
	focal_length_mm REAL,
	focal_length_35mm INTEGER,
	aperture_f_number REAL,
	exposure_time_seconds REAL,
	iso INTEGER,
	orientation INTEGER,

	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (album_id) REFERENCES albums(id)
);

CREATE INDEX IF NOT EXISTS idx_albums_slug ON albums(slug);
CREATE INDEX IF NOT EXISTS idx_albums_path ON albums(path COLLATE BINARY);
CREATE INDEX IF NOT EXISTS idx_assets_album_id ON assets(album_id);
