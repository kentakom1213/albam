package api

import (
	"net/http"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/storage"
)

type Server struct {
	store      *storage.Storage
	cfg        config.Config
	configPath string
}

func NewServer(store *storage.Storage, cfg config.Config) *Server {
	return &Server{
		store: store,
		cfg:   cfg,
	}
}

func NewServerWithConfigPath(store *storage.Storage, cfg config.Config, configPath string) *Server {
	return &Server{
		store:      store,
		cfg:        cfg,
		configPath: configPath,
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
	mux.HandleFunc("/api/config", s.handleGetConfig)
	mux.HandleFunc("/theme/runtime.css", s.handleRuntimeCSS)
	mux.HandleFunc("/api/albums", s.handleListAlbums)
	mux.HandleFunc("/api/albums/", s.handleAlbumSubroutes)
	mux.HandleFunc("/api/media/", s.handlePhotoSubroutes)

	mediaHandler := NewMediaHandler(
		s.store,
		s.cfg.Media.SourceDir,
		s.cfg.Media.CacheDir,
		s.cfg.Media.AllowOriginalDownload,
	)
	mux.HandleFunc("/media/", mediaHandler.ServeHTTP)
}
