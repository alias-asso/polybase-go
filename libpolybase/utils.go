package libpolybase

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PB struct {
	db        *sql.DB
	logPath   string
	logStdout bool
}

func New(db *sql.DB, logPath string, logStdout bool) *PB {
	return &PB{db: db, logPath: logPath, logStdout: logStdout}
}

func NewCourseID(code string, kind string, part int) CourseID {
	return CourseID{code, kind, part}
}

func (e *CourseNotFound) Error() string {
	return "course not found"
}

func ValidateCourseID(id CourseID) (CourseID, error) {
	// Validate code: only uppercase, numbers, dashes, and curly braces
	if !regexp.MustCompile(`^[A-Z0-9\-{},]+$`).MatchString(id.Code) {
		return CourseID{}, fmt.Errorf("invalid code format: must only contain uppercase letters, numbers, dashes, and curly braces")
	}

	// Validate kind: only letters (upper and lowercase)
	if !regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(id.Kind) {
		return CourseID{}, fmt.Errorf("invalid kind format: must only contain letters")
	}

	return id, nil
}

func (c Course) ID() string {
	return fmt.Sprintf("%s/%s/%d", c.Code, c.Kind, c.Part)
}

func (c Course) CID() CourseID {
	return CourseID{
		Code: c.Code,
		Kind: c.Kind,
		Part: c.Part,
	}
}

func (c Course) SID() string {
	sid := fmt.Sprintf("course-%s-%s-%d", c.Code, c.Kind, c.Part)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sid = reg.ReplaceAllString(sid, "-")
	return strings.ToLower(sid)
}

func (c CourseID) ID() string {
	return fmt.Sprintf("%s/%s/%d", c.Code, c.Kind, c.Part)
}


func (c CourseID) SID() string {
	sid := fmt.Sprintf("course-%s-%s-%d", c.Code, c.Kind, c.Part)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sid = reg.ReplaceAllString(sid, "-")
	return strings.ToLower(sid)
}

func (c CourseID) PID() string {
	return fmt.Sprintf("%s %s %d", c.Code, c.Kind, c.Part)
}

func validateSemester(semester string) error {
	if semester == "" {
		return fmt.Errorf("semester cannot be empty")
	}

	if !strings.HasPrefix(semester, "S") {
		return fmt.Errorf("semester must start with 'S'")
	}

	n, err := strconv.Atoi(semester[1:])
	if err != nil {
		return fmt.Errorf("invalid semester format: must be S followed by a number")
	}

	if n != 1 && n != 2 {
		return fmt.Errorf("invalid semester format: semester number must be either 1 or 2")
	}

	return nil
}

func validateQuantity(quantity int, total int) error {
	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	if total <= 0 {
		return fmt.Errorf("total cannot be negative or nil")
	}

	if quantity > total {
		return fmt.Errorf("quantity (%d) cannot exceed total (%d)", quantity, total)
	}

	return nil
}

func clampQuantity(quantity, total int) int {
	if quantity < 0 {
		return 0
	}
	if quantity > total {
		return total
	}
	return quantity
}

func (pb *PB) exists(ctx context.Context, id CourseID, querier interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}) (bool, error) {
	var exists int
	err := querier.QueryRowContext(ctx, `
    SELECT 1 FROM courses WHERE code = ? AND kind = ? AND part = ?`,
		id.Code, id.Kind, id.Part).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pb *PB) getCourse(ctx context.Context, id CourseID, querier interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}) (Course, error) {
	var course Course
	var shown int
	err := querier.QueryRowContext(ctx, `
    SELECT code, kind, part, parts, name, quantity, total, shown, semester
    FROM courses
    WHERE code = ? AND kind = ? AND part = ?`,
		id.Code, id.Kind, id.Part).Scan(
		&course.Code, &course.Kind, &course.Part, &course.Parts,
		&course.Name, &course.Quantity, &course.Total, &shown, &course.Semester)
	if err == sql.ErrNoRows {
		return Course{}, &CourseNotFound{}
	}
	if err != nil {
		return Course{}, fmt.Errorf("failed to retrieve course: %w", err)
	}
	course.Shown = shown == 1
	return course, nil
}

func (pb *PB) logAction(user string, action string, details string) error {
	if pb.logPath == "" {
		return nil
	}

	f, err := os.OpenFile(pb.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer f.Close()

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	logEntry := fmt.Sprintf("%s [%s] %s: %s\n", timestamp, user, action, details)

	if _, err := f.WriteString(logEntry); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	fmt.Printf("%s", logEntry)

	return nil
}
