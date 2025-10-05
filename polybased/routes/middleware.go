package routes

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("X-Auth-Token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Parse and validate the token
		type Claims struct {
			Username string `json:"username"`
			jwt.RegisteredClaims
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next(w, r.WithContext(ctx))
	}
}
