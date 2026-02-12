package routes

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) withContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("X-Auth-Token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			log.Printf("error: %v", err)
			return
		}

		// Parse and validate the token
		type Claims struct {
			Username string `json:"username"`
			jwt.RegisteredClaims
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (any, error) {
			return []byte(s.cfg.Auth.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get username from claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add username to request context
		next(w, r.WithContext(config.SetAuth(r.Context(), claims.Username)))
	}
}
