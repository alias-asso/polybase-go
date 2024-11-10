package routes

import (
	"net/http"

	"git.sr.ht/~alias/polybase/cmd/polybase-web/config"
)

// RequireAuth is middleware that checks for authentication
func RequireAuth(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement proper JWT validation
		// For now, just check for a placeholder header
		if r.Header.Get("X-Auth-Token") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
