package routes

import (
	"net/http"
)

// Register sets up all routes for the application
func register(mux *http.ServeMux, ctx *ServerContext) {
	// Register public routes
	registerPublicRoutes(mux, ctx)

	// Register admin routes
	registerAdminRoutes(mux, ctx)
}

func registerPublicRoutes(mux *http.ServeMux, ctx *ServerContext) {
	mux.HandleFunc("GET /", withContext(ctx, getHome))
	mux.HandleFunc("GET /login", withContext(ctx, getLogin))
	mux.HandleFunc("POST /auth", withContext(ctx, postAuth))
}

func registerAdminRoutes(mux *http.ServeMux, ctx *ServerContext) {
	// Admin dashboard
	mux.HandleFunc("GET /admin", adminAuth(ctx, getAdmin))

	// Admin mode management
	mux.HandleFunc("GET /admin/individual", adminAuth(ctx, getAdminIndividual))
	mux.HandleFunc("GET /admin/bulk", adminAuth(ctx, getAdminBulk))

	// Course management display
	mux.HandleFunc("GET /admin/courses/new", adminAuth(ctx, getAdminCoursesNew))
	mux.HandleFunc("GET /admin/courses/edit/{code}/{kind}/{part}", adminAuth(ctx, getAdminCoursesEdit))
	mux.HandleFunc("GET /admin/courses/delete/{code}/{kind}/{part}", adminAuth(ctx, getAdminCoursesDelete))

	// Course management
	mux.HandleFunc("PUT /admin/courses/{code}/{kind}/{part}", adminAuth(ctx, putAdminCourses))
	mux.HandleFunc("DELETE /admin/courses/{code}/{kind}/{part}", adminAuth(ctx, deleteAdminCourses))

	// Course updating
	mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/quantity", adminAuth(ctx, patchAdminCoursesQuantity))
	mux.HandleFunc("PATCH /admin/courses/{code}/{kind}/{part}/visibility", adminAuth(ctx, patchAdminCoursesVisibility))

	// Bulk updating
	mux.HandleFunc("PATCH /admin/courses/quantities", adminAuth(ctx, patchAdminCoursesQuantities))

}

type routeHandler func(ctx *ServerContext, w http.ResponseWriter, r *http.Request)

func withContext(ctx *ServerContext, handler routeHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(ctx, w, r)
	}
}
