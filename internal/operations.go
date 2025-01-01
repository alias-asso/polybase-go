package internal

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type PB struct {
	db *sql.DB
}

func New(db *sql.DB) *PB {
	return &PB{db: db}
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

func (c Course) SID() string {
	sid := fmt.Sprintf("course-%s-%s-%d", c.Code, c.Kind, c.Part)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sid = reg.ReplaceAllString(sid, "-")
	return strings.ToLower(sid)
}

func (pb *PB) Create(ctx context.Context, course Course) (Course, error) {
	course, err := validateCourse(course)
	if err != nil {
		return Course{}, err
	}

	exists, err := pb.exists(ctx, CourseID{course.Code, course.Kind, course.Part})
	if err != nil {
		return Course{}, fmt.Errorf("failed to check course existence: %w", err)
	}
	if exists {
		return Course{}, fmt.Errorf("course already exists")
	}

	_, err = ValidateCourseID(NewCourseID(course.Code, course.Kind, course.Part))
	if err != nil {
		return Course{}, fmt.Errorf("invalid course id")
	}

	course.Shown = true

	if _, err := pb.db.ExecContext(ctx, `
    INSERT INTO courses (code, kind, part, parts, name, quantity, total, shown, semester)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		course.Code, course.Kind, course.Part, course.Parts, course.Name, course.Quantity, course.Total, course.Shown, course.Semester); err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}

	return course, nil
}

func (pb *PB) Get(ctx context.Context, id CourseID) (Course, error) {
	id, err := ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	var course Course
	var shown int

	err = pb.db.QueryRowContext(ctx, `
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

func (pb *PB) Update(ctx context.Context, id CourseID, partial PartialCourse) (Course, error) {
	id, err := ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	course, err := pb.mergeCourse(ctx, id, partial)
	if err != nil {
		return Course{}, err
	}

	exists, err := pb.exists(ctx, id)
	if err != nil {
		return Course{}, fmt.Errorf("failed to check course existence: %w", err)
	}
	if !exists {
		return Course{}, fmt.Errorf("course does not exists")
	}

	if _, err := pb.db.ExecContext(ctx, `
        UPDATE courses 
        SET code = ?, kind = ?, part = ?, parts = ?, name = ?, quantity = ?, total = ?, shown = ?, semester = ?
        WHERE code = ? AND kind = ? AND part = ?`,
		course.Code, course.Kind, course.Part, course.Parts,
		course.Name, course.Quantity, course.Total, course.Shown, course.Semester,
		id.Code, id.Kind, id.Part,
	); err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}

	return course, nil
}

func (pb *PB) Delete(ctx context.Context, id CourseID) error {
	id, err := ValidateCourseID(id)
	if err != nil {
		return err
	}

	exists, err := pb.exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check course existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("course does not exists")
	}

	if _, err := pb.db.ExecContext(ctx, `
    DELETE FROM courses
    WHERE code = ? AND KIND = ? AND part = ?`,
		id.Code, id.Kind, id.Part); err != nil {
		return fmt.Errorf("delete course: %w", err)
	}

	return nil
}

func (pb *PB) List(ctx context.Context, showHidden bool, filterSemester *string, filterCode *string, filterKind *string, filterPart *int) ([]Course, error) {
	var courses []Course
	var conditions []string
	var args []interface{}

	if !showHidden {
		conditions = append(conditions, "shown = 1")
	}

	if filterSemester != nil {
		conditions = append(conditions, "semester = ?")
		args = append(args, *filterSemester)
	}

	if filterCode != nil {
		conditions = append(conditions, "code = ?")
		args = append(args, *filterCode)
	}

	if filterKind != nil {
		conditions = append(conditions, "kind = ?")
		args = append(args, *filterKind)
	}

	if filterPart != nil {
		conditions = append(conditions, "part = ?")
		args = append(args, *filterPart)
	}

	query := `SELECT code, kind, part, parts, name, quantity, total, shown, semester FROM courses`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY semester DESC, code ASC, kind ASC, part ASC"

	rows, err := pb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Course

		if err := rows.Scan(&c.Code, &c.Kind, &c.Part, &c.Parts, &c.Name,
			&c.Quantity, &c.Total, &c.Shown, &c.Semester); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}

		courses = append(courses, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate courses: %w", err)
	}

	return courses, nil
}

func (pb *PB) UpdateQuantity(ctx context.Context, id CourseID, delta int) (Course, error) {
	id, err := ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	current, err := pb.Get(ctx, id)
	if err != nil {
		return Course{}, fmt.Errorf("failed to get current course: %w", err)
	}

	newQuantity := clampQuantity(current.Quantity+delta, current.Total)

	if _, err = pb.db.ExecContext(ctx, ` UPDATE courses 
        SET quantity = ?
        WHERE code = ? AND kind = ? AND part = ?`,
		newQuantity, id.Code, id.Kind, id.Part); err != nil {
		return Course{}, fmt.Errorf("update quantity: %w", err)
	}

	return pb.Get(ctx, id)
}

func (pb *PB) UpdateShown(ctx context.Context, id CourseID, shown bool) (Course, error) {
	id, err := ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	shownInt := 0
	if shown {
		shownInt = 1
	}
	_, err = pb.db.ExecContext(ctx, `
        UPDATE courses 
        SET shown = ?
        WHERE code = ? AND kind = ? AND part = ?`,
		shownInt, id.Code, id.Kind, id.Part,
	)
	if err != nil {
		return Course{}, fmt.Errorf("update shown: %w", err)
	}
	return pb.Get(ctx, id)
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

func (pb *PB) exists(ctx context.Context, id CourseID) (bool, error) {
	var exists int
	err := pb.db.QueryRowContext(ctx, `
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

func validateCourse(course Course) (Course, error) {
	course.Code = strings.TrimSpace(course.Code)
	if course.Code == "" {
		return Course{}, fmt.Errorf("CODE cannot be empty")
	}

	course.Kind = strings.TrimSpace(course.Kind)
	if course.Kind == "" {
		return Course{}, fmt.Errorf("KIND cannot be empty")
	}

	course.Name = strings.TrimSpace(course.Name)

	if err := validateQuantity(course.Quantity, course.Total); err != nil {
		return Course{}, fmt.Errorf("invalid quantities: %w", err)
	}

	course.Semester = strings.TrimSpace(course.Semester)
	if err := validateSemester(course.Semester); err != nil {
		return Course{}, fmt.Errorf("invalid semester: %w", err)
	}

	return course, nil
}

func (pb *PB) mergeCourse(ctx context.Context, id CourseID, partial PartialCourse) (Course, error) {
	if partial.Code == nil &&
		partial.Kind == nil &&
		partial.Part == nil &&
		partial.Parts == nil &&
		partial.Name == nil &&
		partial.Quantity == nil &&
		partial.Total == nil &&
		partial.Shown == nil &&
		partial.Semester == nil {
		return Course{}, fmt.Errorf("at least one field must be updated")
	}

	current, err := pb.Get(ctx, id)
	if err != nil {
		return Course{}, fmt.Errorf("get current course: %w", err)
	}

	course := Course{
		Code:     current.Code,
		Kind:     current.Kind,
		Part:     current.Part,
		Parts:    current.Parts,
		Name:     current.Name,
		Quantity: current.Quantity,
		Total:    current.Total,
		Shown:    current.Shown,
		Semester: current.Semester,
	}

	if partial.Code != nil {
		course.Code = *partial.Code
	}
	if partial.Kind != nil {
		course.Kind = *partial.Kind
	}
	if partial.Part != nil {
		course.Part = *partial.Part
	}
	if partial.Parts != nil {
		course.Parts = *partial.Parts
	}
	if partial.Name != nil {
		course.Name = *partial.Name
	}
	if partial.Quantity != nil {
		course.Quantity = *partial.Quantity
	}
	if partial.Total != nil {
		course.Total = *partial.Total
	}
	if partial.Shown != nil {
		course.Shown = *partial.Shown
	}
	if partial.Semester != nil {
		course.Semester = *partial.Semester
	}

	return validateCourse(course)
}
