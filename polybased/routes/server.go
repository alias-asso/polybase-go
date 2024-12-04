package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/internal"
	"git.sr.ht/~alias/polybase/polybased/config"
	_ "github.com/mattn/go-sqlite3"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	mux  *http.ServeMux
	addr string
	cfg  *config.Config
	pb   internal.Polybase
}

// New creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pb := internal.New(db)

	srv := &Server{
		mux:  http.NewServeMux(),
		addr: fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		cfg:  cfg,
		pb:   pb,
	}

	// Register all routes
	srv.registerRoutes()

	return srv, nil
}

func (s *Server) Run() {
	log.Printf("Starting server on %s", s.addr)
	if http.ListenAndServe(s.addr, s.mux) != nil {
		log.Fatalf("Error when listening and serving")
	}
}
