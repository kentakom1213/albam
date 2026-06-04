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
	s.registerDynamicRoutes(mux)
	return mux
}

func (s *Server) RoutesWithStatic(publicDir string) (http.Handler, error) {
	if err := EnsurePublicDir(publicDir); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	s.registerDynamicRoutes(mux)

	// ServeMux は最長一致なので，/api/* や /media/* が / より優先される．
	mux.Handle("/", NewStaticHandler(publicDir))

	return mux, nil
}

func (s *Server) registerDynamicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/albums", s.handleListAlbums)
	mux.HandleFunc("/api/albums/", s.handleAlbumSubroutes)
	mux.HandleFunc("/api/tags", s.handleListTags)

	mediaHandler := NewMediaHandler(s.store, s.cfg.Media.SourceDir, s.cfg.Media.CacheDir)
	mux.HandleFunc("/media/", mediaHandler.ServeHTTP)
	mux.HandleFunc("/media/photos/", mediaHandler.ServeHTTP)
}
