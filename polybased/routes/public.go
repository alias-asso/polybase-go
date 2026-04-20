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
	for i, c := range courses {
		c.Semester = "Semestre " + string([]rune(c.Semester)[1:])
		courses[i] = c
	}

	s.count += 1

	err = views.Public(courses, s.count).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	if config.IsLogged(r.Context()) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	cfg := config.GetConfig(r.Context())
	isDev := config.IsDev(r.Context())

	state, err := generateState()
	if err != nil {
		http.Error(w, "Erreur lors de la génération de l'état", http.StatusInternalServerError)
		return
	}

	if err := setOIDCStateCookie(w, state, cfg, isDev); err != nil {
		http.Error(w, "Failed to prepare OIDC state", http.StatusInternalServerError)
		return
	}

	authURL, err := s.getOIDCURL(state)
	if err != nil {
		http.Error(w, "Failed to generate OIDC login URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

func (s *Server) getAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	cfg := config.GetConfig(r.Context())
	isDev := config.IsDev(r.Context())
	defer clearOIDCStateCookie(w, isDev)

	if code == "" || state == "" {
		http.Error(w, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	if !validOIDCState(r, state, cfg) {
		http.Error(w, "CSRF validation failed, try again", http.StatusForbidden)
		return
	}

	givenName, err := s.verifyOIDCCode(code)
	if err != nil {
		log.Printf("OIDC verification failed: %v", err)
		http.Error(w, "Erreur d'authentification", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(givenName, cfg)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	expiry, err := time.ParseDuration(cfg.Auth.JWTExpiry)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	cookieSameSite := http.SameSiteStrictMode
	if isDev {
		cookieSameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "X-Auth-Token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDev,
		SameSite: cookieSameSite,
		MaxAge:   int(expiry.Hours()) * 3600,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) getNotFound(w http.ResponseWriter, r *http.Request) {
	err := views.NotFound().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}
