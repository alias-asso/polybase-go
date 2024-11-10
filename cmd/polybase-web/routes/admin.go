package routes

import (
	"log"
	"net/http"
)

// getAdmin
func getAdmin(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get admin"))
}

// getAdminIndividual
func getAdminIndividual(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin individual - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get admin individual"))
}

// getAdminBulk
func getAdminBulk(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin bulk - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get admin bulk"))
}

// getAdminCoursesNew
func getAdminCoursesNew(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get admin courses new - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get admin courses new"))
}

// getAdminCoursesEdit
func getAdminCoursesEdit(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/edit/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses edit - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Get admin courses edit"))
}

// getAdminCoursesDelete
func getAdminCoursesDelete(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/delete/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Get admin courses delete - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Get admin courses delete"))
}

// putAdminCourses
func putAdminCourses(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Put admin courses - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Put admin courses"))
}

// deleteAdminCourses
func deleteAdminCourses(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Delete admin courses - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Delete admin courses"))
}

// patchAdminCoursesQuantity
func patchAdminCoursesQuantity(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Patch admin courses quantity - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Patch admin courses quantity"))
}

// patchAdminCoursesVisibility
func patchAdminCoursesVisibility(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	code, kind, part, err := parseUrl("/admin/courses/", r)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	log.Printf("Patch admin courses visibility - Config: %+v, DB: %+v, Code: %+v, Kind: %+v, Part: %+v", ctx.Config, ctx.DB, code, kind, part)
	w.Write([]byte("Patch admin courses visibility"))
}

// patchAdminCoursesQuantities
func patchAdminCoursesQuantities(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Patch admin courses quantities - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Patch admin courses quantities"))
}
