package routes

import (
	"log"
	"net/http"
)

// Register sets up all routes for the application
func (s *Server) registerRoutes() {
	s.registerStatic()

	// Public routes
	s.mux.HandleFunc("GET /", s.getHome)
	s.mux.HandleFunc("GET /login", s.getLogin)
	s.mux.HandleFunc("POST /auth", s.postAuth)

	s.mux.HandleFunc("GET /admin", s.withAuth(s.getAdmin))

	s.mux.HandleFunc("GET /admin/individual", s.withAuth(s.getAdminIndividual))
	s.mux.HandleFunc("GET /admin/bulk", s.withAuth(s.getAdminBulk))

	s.mux.HandleFunc("GET /admin/courses/new", s.withAuth(s.getAdminCoursesNew))
	s.mux.HandleFunc("GET /admin/courses/edit/{code}/{kind}/{part}", s.withAuth(s.getAdminCoursesEdit))
	s.mux.HandleFunc("GET /admin/courses/delete/{code}/{kind}/{part}", s.withAuth(s.getAdminCoursesDelete))

	s.mux.HandleFunc("PUT /admin/courses/{code}/{kind}/{part}", s.withAuth(s.putAdminCourses))
	s.mux.HandleFunc("DELETE /admin/courses/{code}/{kind}/{part}", s.withAuth(s.deleteAdminCourses))

	s.mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/quantity", s.withAuth(s.patchAdminCoursesQuantity))
	s.mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/visibility", s.withAuth(s.patchAdminCoursesQuantity))
}

func (s *Server) registerStatic() {
	if s.cfg.Server.Static == "" {
		log.Fatalf("Warning: Static file path not configured, static assets will not be served")
		return
	}

	fs := http.FileServer(http.Dir(s.cfg.Server.Static))

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	})

	s.mux.Handle("GET /static/", staticHandler)
}

// func registerAdminRoutes(mux *http.ServeMux, ctx *ServerContext) {
// 	// Admin mode management
// 	mux.HandleFunc("GET /admin/individual", adminAuth(ctx, getAdminIndividual))
// 	mux.HandleFunc("GET /admin/bulk", adminAuth(ctx, getAdminBulk))
//
// 	// Course management display
// 	mux.HandleFunc("GET /admin/courses/new", adminAuth(ctx, getAdminCoursesNew))
// 	mux.HandleFunc("GET /admin/courses/edit/{code}/{kind}/{part}", adminAuth(ctx, getAdminCoursesEdit))
// 	mux.HandleFunc("GET /admin/courses/delete/{code}/{kind}/{part}", adminAuth(ctx, getAdminCoursesDelete))
//
// 	// Course management
// 	mux.HandleFunc("PUT /admin/courses/{code}/{kind}/{part}", adminAuth(ctx, putAdminCourses))
// 	mux.HandleFunc("DELETE /admin/courses/{code}/{kind}/{part}", adminAuth(ctx, deleteAdminCourses))
//
// 	// Course updating
// 	mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/quantity", adminAuth(ctx, patchAdminCoursesQuantity))
// 	mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/visibility", adminAuth(ctx, patchAdminCoursesVisibility))
//
// 	// Bulk updating
// 	mux.HandleFunc("PATCH /admin/courses/quantities", adminAuth(ctx, patchAdminCoursesQuantities))
//
// }
//
// type routeHandler func(ctx *ServerContext, w http.ResponseWriter, r *http.Request)
//
// func withContext(ctx *ServerContext, handler routeHandler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		handler(ctx, w, r)
// 	}
// }
