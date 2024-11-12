package routes

import (
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/templates"
)

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) {
	courses, err := s.pb.List(r.Context(), false)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = templates.Index(courses).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	err := templates.Login().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) postAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	creds := struct {
		Username string
		Password string
	}{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	log.Printf("Post auth - Username: %s", creds.Username)

	switch creds.Username {
	case "success":
		token := "dummy-jwt-token-123"

		http.SetCookie(w, &http.Cookie{
			Name:     "X-Auth-Token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   3600 * 24,
		})

		w.Header().Set("HX-Redirect", "/admin")
		w.WriteHeader(http.StatusOK)

	case "error":
		http.Error(w, "Internal server error", http.StatusInternalServerError)

	default:
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
}
