package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"git.sr.ht/~alias/polybase-go/internal"
	"git.sr.ht/~alias/polybase-go/views"
)

// getAdmin
func (s *Server) getAdmin(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Admin(courses, packs, username).Render(r.Context(), w)
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
	id, err := parseCourseUrl("/admin/courses/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	course, err := s.pb.GetCourse(r.Context(), id)
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
	id, err := parseCourseUrl("/admin/courses/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	course, err := s.pb.GetCourse(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.CourseDeleteConfirm(course).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminPacksNew
func (s *Server) getAdminPacksNew(w http.ResponseWriter, r *http.Request) {
	courses, err := s.pb.ListCourse(r.Context(), false, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.NewPackForm(courses).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminPacksEdit
func (s *Server) getAdminPacksEdit(w http.ResponseWriter, r *http.Request) {
	id, err := parsePackUrl("/admin/packs/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), false, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to get courses", http.StatusInternalServerError)
		log.Printf("Failed to get courses: %v", err)
	}

	pack, err := s.pb.GetPack(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get pack", http.StatusInternalServerError)
		log.Printf("Failed to get pack: %v", err)
	}

	err = views.EditPackForm(pack, courses).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminPacksDelete
func (s *Server) getAdminPacksDelete(w http.ResponseWriter, r *http.Request) {
	id, err := parsePackUrl("/admin/packs/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	pack, err := s.pb.GetPack(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.PackDeleteConfirm(pack).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) getAdminPack(w http.ResponseWriter, r *http.Request) {
	id, err := parsePackUrl("/admin/packs/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	expandedStr := r.Form.Get("expanded")
	if expandedStr == "" {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	expanded, err := strconv.ParseBool(expandedStr)
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	pack, err := s.pb.GetPack(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get course", http.StatusInternalServerError)
		log.Printf("Failed to get course: %v", err)
	}

	err = views.PackCard(pack, expanded).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// getAdminPacksNew
func (s *Server) getAdminStatistics(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin statistics - Config: %+v, Polybase: %+v", s.cfg, s.pb)
	w.Write([]byte("Get admin statistics"))
}

// postAdminCourses
func (s *Server) postAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseCourseUrl("/admin/courses/", r)
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

	_, err = s.pb.GetCourse(r.Context(), id)
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

	_, err = s.pb.CreateCourse(r.Context(), username, course)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), packs, true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// putAdminCourses
func (s *Server) putAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseCourseUrl("/admin/courses/", r)
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

	_, err = s.pb.GetCourse(r.Context(), id)
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

	_, err = s.pb.UpdateCourse(r.Context(), username, id, course)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), packs, true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// deleteAdminCourses
func (s *Server) deleteAdminCourses(w http.ResponseWriter, r *http.Request) {
	id, err := parseCourseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	username := getUsernameFromContext(r.Context())

	_, err = s.pb.GetCourse(r.Context(), id)
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

	err = s.pb.DeleteCourse(r.Context(), username, id)
	if err != nil {
		http.Error(w, "Failed to add course", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), packs, true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

// patchAdminCoursesQuantity
func (s *Server) patchAdminCoursesQuantity(w http.ResponseWriter, r *http.Request) {
	id, err := parseCourseUrl("/admin/courses/", r)
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
	id, err := parseCourseUrl("/admin/courses/", r)
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

func (s *Server) patchAdminPacksQuantity(w http.ResponseWriter, r *http.Request) {
	id, err := parsePackUrl("/admin/packs/", r)
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

	_, err = s.pb.UpdatePackQuantity(r.Context(), username, id, delta)
	if err != nil {
		log.Println("Patch admin course quantity - error:", err)
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), packs, true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) postAdminPacks(w http.ResponseWriter, r *http.Request) {
	username := getUsernameFromContext(r.Context())

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")
	var coursesId []internal.CourseID
	for _, idStr := range r.Form["courses"] {
		parts := strings.Split(idStr, "/")

		code := parts[0]
		kind := parts[1]
		part, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Failed to get part", http.StatusInternalServerError)
			log.Printf("%s", err)
			return
		}

		id := internal.CourseID{
			Code: code,
			Kind: kind,
			Part: part,
		}

		_, err = s.pb.GetCourse(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to get course", http.StatusInternalServerError)
			log.Printf("%s", err)
			return
		}

		coursesId = append(coursesId, id)
	}

	fmt.Println(username)
	fmt.Println(name)
	for _, course := range coursesId {
		fmt.Println(course)
	}

	_, err = s.pb.CreatePack(r.Context(), username, name, coursesId)
	if err != nil {
		http.Error(w, "Failed to add pack", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	courses, err := s.pb.ListCourse(r.Context(), true, nil, nil, nil, nil)
	if err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	packs, err := s.pb.ListPacks(r.Context())
	if err != nil {
		http.Error(w, "Failed to list packs", http.StatusInternalServerError)
		log.Printf("%s", err)
		return
	}

	err = views.Grid(views.GroupCoursesBySemesterAndKind(courses), packs, true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Failed to render template: %v", err)
	}
}

func (s *Server) putAdminPacks(w http.ResponseWriter, r *http.Request) {
	// TODO:
}

func (s *Server) deleteAdminPacks(w http.ResponseWriter, r *http.Request) {
	// TODO:
}

func getUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return ""
	}
	return username
}
