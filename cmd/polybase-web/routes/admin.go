package routes

import (
	"log"
	"net/http"

	"git.sr.ht/~alias/polybase/templates"
)

// getAdmin
func (s *Server) getAdmin(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	courses, err := s.pb.List(r.Context(), false)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = templates.Admin(courses, username, false).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminIndividual
func (s *Server) getAdminIndividual(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin individual - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin individual"))
}

// getAdminBulk
func (s *Server) getAdminBulk(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin bulk - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin bulk"))
}

// getAdminCoursesNew
func (s *Server) getAdminCoursesNew(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin courses new - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin courses new"))
}

// getAdminCoursesEdit
func (s *Server) getAdminCoursesEdit(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses edit - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Get admin courses edit"))
}

// getAdminCoursesDelete
func (s *Server) getAdminCoursesDelete(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses delete - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Get admin courses delete"))
}

// putAdminCourses
func (s *Server) putAdminCourses(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Put admin courses - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Put admin courses"))
}

// deleteAdminCourses
func (s *Server) deleteAdminCourses(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Delete admin courses - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Delete admin courses"))
}

// patchAdminCoursesQuantity
func (s *Server) patchAdminCoursesQuantity(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Patch admin courses quantity - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Patch admin courses quantity"))
}

// patchAdminCoursesVisibility
func (s *Server) patchAdminCoursesVisibility(w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Patch admin courses visibility - Config: %+v, Polybase: %+v, Code: %+v, Kind: %+v, Part: %+v", s.cfg, s.pb, code, kind, part)
	w.Write([]byte("Patch admin courses visibility"))
}

// patchAdminCoursesQuantities
func (s *Server) patchAdminCoursesQuantities(w http.ResponseWriter, r *http.Request) {
	log.Printf("Patch admin courses quantities - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Patch admin courses quantities"))
}
