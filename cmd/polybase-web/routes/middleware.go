package routes

import (
	"net/http"
)

// adminAuth is middleware that checks for authentication
func adminAuth(ctx *ServerContext, next routeHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement proper JWT validation
		if r.Header.Get("X-Auth-Token") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(ctx, w, r)
	}
}
