package routes

import (
	"net/http"
)

// Public route handlers
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Polybase Public View"))
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login Page"))
}

func handleCoursesList(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Public Course Listing"))
}
