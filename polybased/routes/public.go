package routes

import (
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase-go/views"
	"github.com/golang-jwt/jwt/v5"
)

// getHome
func (s *Server) getHome(w http.ResponseWriter, r *http.Request) {
	if ok := s.isLoggedIn(r); ok {
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

// getLogin
func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	err := views.Login().Render(r.Context(), w)
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
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600 * 24,
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

func (s *Server) isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("X-Auth-Token")
	if err != nil {
		return false
	}

	type Claims struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return false
	}

	_, ok := token.Claims.(*Claims)
	return ok
}
