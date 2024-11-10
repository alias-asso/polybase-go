package routes

import (
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/cmd/polybase-web/config"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	config *config.Config
	mux    *http.ServeMux
	addr   string
}

// New creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	srv := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
		addr:   fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
	}

	// Register all routes
	if err := register(srv.mux, cfg); err != nil {
		return nil, err
	}

	return srv, nil
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) Run() {
	log.Printf("Starting server on %s", s.addr)
	http.ListenAndServe(s.addr, s.mux)
}
