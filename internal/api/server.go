package api

import (
	"net/http"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/storage"
)

type Server struct {
	store *storage.Storage
	cfg   config.Config
}

func NewServer(store *storage.Storage, cfg config.Config) *Server {
	return &Server{
		store: store,
		cfg:   cfg,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/albums", s.handleListAlbums)
	mux.HandleFunc("/api/albums/", s.handleAlbumSubroutes)

	return mux
}
