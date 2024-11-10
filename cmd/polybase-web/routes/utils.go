package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func parseUrl(filter string, r *http.Request) (string, string, int, error) {
	path := strings.TrimPrefix(r.URL.Path, filter)
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return "", "", 0, fmt.Errorf("error in url path, insufficient information")
	}

	code := strings.TrimSpace(parts[0])
	kind := strings.TrimSpace(parts[1])
	partStr := strings.TrimSpace(parts[2])

	// Validate code: only uppercase, numbers, dashes, and curly braces
	if !regexp.MustCompile(`^[A-Z0-9\-{}]+$`).MatchString(code) {
		return "", "", 0, fmt.Errorf("invalid code format: must only contain uppercase letters, numbers, dashes, and curly braces")
	}

	// Validate kind: only letters (upper and lowercase)
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(kind) {
		return "", "", 0, fmt.Errorf("invalid kind format: must only contain letters")
	}

	// Convert and validate part as integer
	part, err := strconv.Atoi(partStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid part format: must be a valid integer")
	}

	return code, kind, part, nil
}
