package routes

import (
	"net/http"

	"github.com/alias-asso/polybase-go/static"
)

// Register sets up all routes for the application
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.getNotFound)

	s.registerStatic()

	s.mux.HandleFunc("GET /{$}", s.getHome)
	s.mux.HandleFunc("GET /login", s.getLogin)
	s.mux.HandleFunc("POST /auth", s.postAuth)

	s.mux.HandleFunc("GET /admin", s.withAuth(s.getAdmin))

	s.mux.HandleFunc("GET /admin/courses/new", s.withAuth(s.getAdminCoursesNew))
	s.mux.HandleFunc("GET /admin/courses/edit/{code}/{kind}/{part}", s.withAuth(s.getAdminCoursesEdit))
	s.mux.HandleFunc("GET /admin/courses/delete/{code}/{kind}/{part}", s.withAuth(s.getAdminCoursesDelete))

	s.mux.HandleFunc("GET /admin/packs/new", s.withAuth(s.getAdminPacksNew))
	s.mux.HandleFunc("GET /admin/packs/edit/{id}", s.withAuth(s.getAdminPacksEdit))
	s.mux.HandleFunc("GET /admin/packs/delete/{id}", s.withAuth(s.getAdminPacksDelete))

	s.mux.HandleFunc("GET /admin/packs/{id}", s.withAuth(s.getAdminPack))

	// s.mux.HandleFunc("GET /admin/statistics", s.withAuth(s.getAdminStatistics))

	s.mux.HandleFunc("POST /admin/courses/{code}/{kind}/{part}", s.withAuth(s.postAdminCourses))
	s.mux.HandleFunc("PUT /admin/courses/{code}/{kind}/{part}", s.withAuth(s.putAdminCourses))
	s.mux.HandleFunc("DELETE /admin/courses/{code}/{kind}/{part}", s.withAuth(s.deleteAdminCourses))

	s.mux.HandleFunc("POST /admin/packs", s.withAuth(s.postAdminPacks))
	s.mux.HandleFunc("PUT /admin/packs/{id}", s.withAuth(s.putAdminPacks))
	s.mux.HandleFunc("DELETE /admin/packs/{id}", s.withAuth(s.deleteAdminPacks))

	s.mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/quantity", s.withAuth(s.patchAdminCoursesQuantity))
	s.mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/visibility", s.withAuth(s.patchAdminCoursesVisibility))

	s.mux.HandleFunc("PATCH /admin/packs/{id}/quantity", s.withAuth(s.patchAdminPacksQuantity))
}

func (s *Server) registerStatic() {
	fs := http.FileServer(static.FileSystem())
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Server.Mode == "dev" {
			w.Header().Set("Cache-Control", "public, max-age=0")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=63072000")
		}
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})
	s.mux.Handle("GET /static/", staticHandler)
}
