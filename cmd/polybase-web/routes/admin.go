package routes

import (
	"net/http"
	"strings"
)

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Admin Dashboard"))
}

func handleAdminMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Mode toggled"))
}

func handleNewCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("New course form"))
}

func handleBulkUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("Bulk update processed"))
}

func handleCourseOperations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/courses/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 2 && parts[1] == "edit" {
			w.Write([]byte("Edit course form"))
		}
	case http.MethodPut:
		if len(parts) == 3 {
			switch parts[1] {
			case "quantity":
				w.Write([]byte("Update quantity"))
			case "visibility":
				w.Write([]byte("Toggle visibility"))
			default:
				http.NotFound(w, r)
			}
		}
	case http.MethodDelete:
		w.Write([]byte("Delete course"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
