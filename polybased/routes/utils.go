package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.sr.ht/~alias/polybase-go/libpolybase"
)

func parseCourseUrl(filter string, r *http.Request) (libpolybase.CourseID, error) {
	path := strings.TrimPrefix(r.URL.Path, filter)
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return libpolybase.CourseID{}, fmt.Errorf("error in url path, insufficient information")
	}

	code := strings.TrimSpace(parts[0])
	kind := strings.TrimSpace(parts[1])

	part, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return libpolybase.CourseID{}, fmt.Errorf("invalid part format: must be a valid integer")
	}

	return libpolybase.ValidateCourseID(libpolybase.NewCourseID(code, kind, part))
}

func parsePackUrl(filter string, r *http.Request) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, filter)
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		return 0, fmt.Errorf("error in url path, insufficient information")
	}

	id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, fmt.Errorf("invalid part format: must be a valid integer")
	}

	return id, nil
}
