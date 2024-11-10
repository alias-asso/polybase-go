package routes

import (
	"net/http"

	"git.sr.ht/~alias/polybase/cmd/polybase-web/config"
)

// Register sets up all routes for the application
func register(mux *http.ServeMux, cfg *config.Config) error {
	// Register public routes
	registerPublicRoutes(mux)

	// Register admin routes
	registerAdminRoutes(mux, cfg)

	return nil
}

func registerPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/auth", handleAuth)
	mux.HandleFunc("/courses", handleCoursesList)
}

func registerAdminRoutes(mux *http.ServeMux, cfg *config.Config) {
	// Create admin handler with authentication
	admin := http.HandlerFunc(handleAdmin)
	mux.Handle("/admin", RequireAuth(admin, cfg))

	// Admin course management routes
	mux.Handle("/admin/mode", RequireAuth(http.HandlerFunc(handleAdminMode), cfg))
	mux.Handle("/admin/courses/new", RequireAuth(http.HandlerFunc(handleNewCourse), cfg))
	mux.Handle("/admin/courses/bulk", RequireAuth(http.HandlerFunc(handleBulkUpdate), cfg))
	mux.Handle("/admin/courses/", RequireAuth(http.HandlerFunc(handleCourseOperations), cfg))
}
