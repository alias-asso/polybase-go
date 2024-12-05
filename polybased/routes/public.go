package routes

import (
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/templates"
)

// getHome
func (s *Server) getHome(w http.ResponseWriter, r *http.Request) {
	courses, err := s.pb.List(r.Context(), false, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = templates.Public(courses).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getLogin
func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	err := templates.Login().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// postAuth
func (s *Server) postAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Échec de l'analyse du formulaire", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	authorized, err := authenticate(username, password, s.cfg)
	if err != nil {
		log.Print(err)
		http.Error(w, "Service LDAP temporairement indisponible", http.StatusInternalServerError)
		return
	}

	if !authorized {
		log.Print("Ldap wrong username or password")
		http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(username, s.cfg)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

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
}
