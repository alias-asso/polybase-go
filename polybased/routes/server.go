package routes

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/alias-asso/polybase-go/libpolybase"
	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

// Server represents the HTTP server and its dependencies
type Server struct {
	mux             *http.ServeMux
	addr            string
	pb              libpolybase.Polybase
	oauth2Config    *oauth2.Config
	oidcAuthOptions []oauth2.AuthCodeOption
	oidcVerifier    *oidc.IDTokenVerifier
	count           int
}

// New creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pb := libpolybase.New(db, cfg.Server.Log, true)
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.OIDC.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("create OIDC provider: %w", err)
	}

	oauth2Config := newOAuth2Config(cfg, provider)
	oidcAuthOptions, err := buildOIDCAuthCodeOptions(cfg.OIDC.ExtraParams)
	if err != nil {
		return nil, fmt.Errorf("parse oidc.extra_params: %w", err)
	}

	srv := &Server{
		mux:             http.NewServeMux(),
		addr:            fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		pb:              pb,
		oauth2Config:    oauth2Config,
		oidcAuthOptions: oidcAuthOptions,
		oidcVerifier:    provider.Verifier(&oidc.Config{ClientID: cfg.OIDC.ClientID}),
		count:           0,
	}

	// Register all routes
	srv.registerRoutes()

	return srv, nil
}

func (s *Server) Run(ctx context.Context) {
	log.Printf("Starting server on %s", s.addr)
	if err := http.ListenAndServe(s.addr, s.withContext(ctx, s.mux)); err != nil {
		log.Fatalf("Error when listening and serving %s", err)
	}
}
