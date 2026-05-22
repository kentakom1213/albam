package model

import "time"

type Album struct {
	ID       int64
	ParentID *int64
	Path     string
	Slug     string
	Title    string
}

type Asset struct {
	ID       int64
	AlbumID  int64
	Path     string
	Filename string
	Ext      string
	Size     int64
	ModTime  time.Time
}
