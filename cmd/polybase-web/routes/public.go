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
	log.Printf("Get login - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get login"))
}

func (s *Server) postAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("Post auth - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Post auth"))
}
