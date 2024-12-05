package routes

import (
	"log"
	"net/http"
	"strconv"

	"git.sr.ht/~alias/polybase/templates"
)

// getAdmin
func (s *Server) getAdmin(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	courses, err := s.pb.List(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = templates.Admin(courses, username).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminCoursesNew
func (s *Server) getAdminCoursesNew(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin courses new - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin courses new"))
}

// getAdminCoursesEdit
func (s *Server) getAdminCoursesEdit(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses edit - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Get admin courses edit"))
}

// getAdminCoursesDelete
func (s *Server) getAdminCoursesDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses delete - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Get admin courses delete"))
}

// putAdminCourses
func (s *Server) putAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Put admin courses - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Put admin courses"))
}

// deleteAdminCourses
func (s *Server) deleteAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Delete admin courses - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Delete admin courses"))
}

// patchAdminCoursesQuantity
func (s *Server) patchAdminCoursesQuantity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	delta, err := strconv.Atoi(r.URL.Query().Get("delta"))
	if err != nil {
		log.Printf("Invalid delta parameter: %v", err)
		http.Error(w, "Invalid delta parameter", http.StatusBadRequest)
		return
	}

	course, err := s.pb.UpdateQuantity(r.Context(), id, delta)
	if err != nil {
		log.Println("Patch admin course quantity - error:", err)
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	err = templates.AdminCard(course).Render(r.Context(), w)
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// patchAdminCoursesVisibility
func (s *Server) patchAdminCoursesVisibility(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	visibility, err := strconv.ParseBool(r.URL.Query().Get("visibility"))
	if err != nil {
		log.Printf("Invalid delta parameter: %v", err)
		http.Error(w, "Invalid delta parameter", http.StatusBadRequest)
		return
	}

	course, err := s.pb.UpdateShown(r.Context(), id, visibility)
	if err != nil {
		log.Println("Patch admin course quantity - error:", err)
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	log.Println(course)

	err = templates.AdminCard(course).Render(r.Context(), w)
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
