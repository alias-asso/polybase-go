package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/", handleHome)

	// Admin routes
	mux.HandleFunc("/admin", handleAdmin)
	mux.HandleFunc("/admin/courses/", handleCourses)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Polybase Public View"))
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Admin Dashboard - LDAP Auth Required"))
}

func handleCourses(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Course Management"))
}
