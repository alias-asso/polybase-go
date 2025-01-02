package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.sr.ht/~alias/polybase-go/internal"
)

func parseUrl(filter string, r *http.Request) (internal.CourseID, error) {
	path := strings.TrimPrefix(r.URL.Path, filter)
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return internal.CourseID{}, fmt.Errorf("error in url path, insufficient information")
	}

	code := strings.TrimSpace(parts[0])
	kind := strings.TrimSpace(parts[1])

	part, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return internal.CourseID{}, fmt.Errorf("invalid part format: must be a valid integer")
	}

	return internal.ValidateCourseID(internal.NewCourseID(code, kind, part))
}
