package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"git.sr.ht/~alias/polybase-go/internal"
	"git.sr.ht/~alias/polybase-go/views"
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

	err = views.Admin(courses, username).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminCoursesNew
func (s *Server) getAdminCoursesNew(w http.ResponseWriter, r *http.Request) {
	err := views.NewCourseForm().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminCoursesEdit
func (s *Server) getAdminCoursesEdit(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	course, err := s.pb.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.EditCourseForm(course).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminCoursesDelete
func (s *Server) getAdminCoursesDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	course, err := s.pb.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.DeleteConfirm(course).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminPacksNew
func (s *Server) getAdminPacksNew(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin packs new - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin packs new"))
}

// getAdminPacksEdit
func (s *Server) getAdminPacksEdit(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/packs/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin packs edit - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Get admin packs edit"))
}

// getAdminPacksDelete
func (s *Server) getAdminPacksDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/packs/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin packs delete - Config: %+v, Polybase: %+v, Id: %+v", s.cfg, s.pb, id)
	w.Write([]byte("Get admin packs delete"))
}

// getAdminPacksNew
func (s *Server) getAdminStatistics(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin statistics - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin statistics"))
}

// postAdminCourses
func (s *Server) postAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	username := getUsernameFromContext(r.Context())

	err = r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	_, err = s.pb.Get(r.Context(), id)
	exists := true
	if err != nil {
		if _, ok := err.(*internal.CourseNotFound); ok {
			exists = false
		} else {
			http.Error(w, "Failed to get course", http.StatusInternalServerError)
			log.Printf("%s", err)
			return
		}
	}

	if exists {
		http.Error(w, "Course already exists", http.StatusBadRequest)
		log.Printf("Failed to add course: course already exists")
		return
	}

	code := id.Code
	kind := id.Kind
	part := id.Part

	parts := 0

	name := r.Form.Get("name")

	quantity, err := strconv.Atoi(r.Form.Get("quantity"))
	if err != nil {
		http.Error(w, "Failed to get quantity", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	totalStr := r.Form.Get("total")
	total := quantity
	if totalStr != "" {
		total, err = strconv.Atoi(totalStr)
		if err != nil {
			http.Error(w, "Invalid total parameter", http.StatusBadRequest)
			log.Printf("Failed to parse total: %s", err)
			return
		}
	}

	shown := true
	semester := r.Form.Get("semester")

	course := internal.Course{
		Code:     code,
		Kind:     kind,
		Part:     part,
		Parts:    parts,
		Name:     name,
		Quantity: quantity,
		Total:    total,
		Shown:    shown,
		Semester: semester,
	}

	_, err = s.pb.Create(r.Context(), username, course)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.List(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// putAdminCourses
func (s *Server) putAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	username := getUsernameFromContext(r.Context())

	err = r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	_, err = s.pb.Get(r.Context(), id)
	exists := true
	if err != nil {
		if err == fmt.Errorf("course not found") {
			exists = false
		} else {
			http.Error(w, "Failed to get course", http.StatusInternalServerError)
			log.Printf("%s", err)
			return
		}
	}

	if !exists {
		http.Error(w, "Course does not exists", http.StatusBadRequest)
		log.Printf("Failed to edit course: course does not exists")
		return
	}

	code := r.Form.Get("code")
	kind := r.Form.Get("kind")
	part, err := strconv.Atoi(r.Form.Get("part"))
	if err != nil {
		http.Error(w, "Failed to get part", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	parts := 0

	name := r.Form.Get("name")

	quantity, err := strconv.Atoi(r.Form.Get("quantity"))
	if err != nil {
		http.Error(w, "Failed to get quantity", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	totalStr := r.Form.Get("total")
	total := quantity
	if totalStr != "" {
		total, err = strconv.Atoi(totalStr)
		if err != nil {
			http.Error(w, "Invalid total parameter", http.StatusBadRequest)
			log.Printf("Failed to parse total: %s", err)
			return
		}
	}

	shown := true

	semester := r.Form.Get("semester")

	course := internal.PartialCourse{
		Code:     &code,
		Kind:     &kind,
		Part:     &part,
		Parts:    &parts,
		Name:     &name,
		Quantity: &quantity,
		Total:    &total,
		Shown:    &shown,
		Semester: &semester,
	}

	_, err = s.pb.Update(r.Context(), username, id, course)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.List(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// deleteAdminCourses
func (s *Server) deleteAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	username := getUsernameFromContext(r.Context())

	_, err = s.pb.Get(r.Context(), id)
	exists := true
	if err != nil {
		if err == fmt.Errorf("course not found") {
			exists = false
		} else {
			http.Error(w, "Failed to get course", http.StatusInternalServerError)
			log.Printf("%s", err)
			return
		}
	}

	if !exists {
		http.Error(w, "Course does not exists", http.StatusBadRequest)
		log.Printf("Failed to edit course: course does not exists")
		return
	}

	err = s.pb.Delete(r.Context(), username, id)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.List(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// patchAdminCoursesQuantity
func (s *Server) patchAdminCoursesQuantity(w http.ResponseWriter, r *http.Request) {
	id, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	username := getUsernameFromContext(r.Context())

	delta, err := strconv.Atoi(r.URL.Query().Get("delta"))
	if err != nil {
		log.Printf("Invalid delta parameter: %v", err)
		http.Error(w, "Invalid delta parameter", http.StatusBadRequest)
		return
	}

	course, err := s.pb.UpdateQuantity(r.Context(), username, id, delta)
	if err != nil {
		log.Println("Patch admin course quantity - error:", err)
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	err = views.CardQuantity(course.Quantity).Render(r.Context(), w)
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

	username := getUsernameFromContext(r.Context())

	visibility, err := strconv.ParseBool(r.URL.Query().Get("visibility"))
	if err != nil {
		log.Printf("Invalid delta parameter: %v", err)
		http.Error(w, "Invalid delta parameter", http.StatusBadRequest)
		return
	}

	course, err := s.pb.UpdateShown(r.Context(), username, id, visibility)
	if err != nil {
		log.Println("Patch admin course quantity - error:", err)
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	err = views.CourseCard(course, true).Render(r.Context(), w)
	if err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func getUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return ""
	}
	return username
}
