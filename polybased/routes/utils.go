package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"git.sr.ht/~alias/polybase/internal"
)

func parseUrl(filter string, r *http.Request) (internal.CourseID, error) {
	path := strings.TrimPrefix(r.URL.Path, filter)
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return internal.CourseID{}, fmt.Errorf("error in url path, insufficient information")
	}

	code := strings.TrimSpace(parts[0])
	kind := strings.TrimSpace(parts[1])
	partStr := strings.TrimSpace(parts[2])

	// Validate code: only uppercase, numbers, dashes, and curly braces
	if !regexp.MustCompile(`^[A-Z0-9\-{},]+$`).MatchString(code) {
		return internal.CourseID{}, fmt.Errorf("invalid code format: must only contain uppercase letters, numbers, dashes, and curly braces")
	}

	// Validate kind: only letters (upper and lowercase)
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(kind) {
		return internal.CourseID{}, fmt.Errorf("invalid kind format: must only contain letters")
	}

	// Convert and validate part as integer
	part, err := strconv.Atoi(partStr)
	if err != nil {
		return internal.CourseID{}, fmt.Errorf("invalid part format: must be a valid integer")
	}

	return internal.NewCourseID(code, kind, part), nil
}
