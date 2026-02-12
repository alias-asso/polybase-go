package routes

import (
	"context"
	"net/http"

	"github.com/alias-asso/polybase-go/polybased/config"
)

func (s *Server) withContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = config.SetAuth(ctx, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connected := config.IsLogged(r.Context())
		if !connected {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
