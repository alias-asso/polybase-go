package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/cmd/polybase-web/config"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	ctx  *ServerContext
	mux  *http.ServeMux
	addr string
}

type ServerContext struct {
	Config *config.Config
	DB     *sql.DB
}

// New creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	ctx := &ServerContext{
		Config: cfg,
	}

	srv := &Server{
		ctx:  ctx,
		mux:  http.NewServeMux(),
		addr: fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
	}

	// Register all routes
	register(srv.mux, srv.ctx)

	return srv, nil
}

func (s *Server) Run() {
	log.Printf("Starting server on %s", s.addr)
	if http.ListenAndServe(s.addr, s.mux) != nil {
		log.Fatalf("Error when listening and serving")
	}
}
