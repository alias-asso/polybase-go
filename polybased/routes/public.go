package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/alias-asso/polybase-go/views"
)

func (s *Server) getHome(w http.ResponseWriter, r *http.Request) {
	if ok := config.IsLogged(r.Context()); ok {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), false, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	s.count += 1

	err = views.Public(courses, s.count).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	err := views.Login().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) postAuth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Échec de l'analyse du formulaire", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	cfg := config.GetConfig(r.Context())

	authorized, err := authenticate(username, password, cfg)
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

	token, err := generateToken(username, cfg)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	expiry, err := time.ParseDuration(cfg.Auth.JWTExpiry)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "X-Auth-Token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(expiry.Hours()) * 3600,
	})

	w.Header().Set("HX-Redirect", "/admin")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getNotFound(w http.ResponseWriter, r *http.Request) {
	err := views.NotFound().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}
