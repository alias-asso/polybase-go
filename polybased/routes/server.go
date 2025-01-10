package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase-go/internal"
	"git.sr.ht/~alias/polybase-go/polybased/config"
	_ "github.com/mattn/go-sqlite3"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	mux   *http.ServeMux
	addr  string
	cfg   *config.Config
	pb    internal.Polybase
	count int
}

// New creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pb := internal.New(db, cfg.Server.Log, true)

	srv := &Server{
		mux:   http.NewServeMux(),
		addr:  fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		cfg:   cfg,
		pb:    pb,
		count: 0,
	}

	// Register all routes
	srv.registerRoutes()

	return srv, nil
}

func (s *Server) Run() {
	log.Printf("Starting server on %s", s.addr)
	if err := http.ListenAndServe(s.addr, s.mux); err != nil {
		log.Fatalf("Error when listening and serving %s", err)
	}
}
