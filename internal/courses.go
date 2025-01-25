package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
)

func (pb *PB) CreateCourse(ctx context.Context, user string, course Course) (Course, error) {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	course, err = validateCourse(course)
	if err != nil {
		return Course{}, err
	}

	exists, err := pb.exists(ctx, CourseID{course.Code, course.Kind, course.Part}, tx)
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
	if _, err := tx.ExecContext(ctx, `
    INSERT INTO courses (code, kind, part, parts, name, quantity, total, shown, semester)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		course.Code, course.Kind, course.Part, course.Parts, course.Name,
		course.Quantity, course.Total, course.Shown, course.Semester); err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}

	if err := pb.setParts(ctx, CourseID{course.Code, course.Kind, course.Part}, tx); err != nil {
		return Course{}, fmt.Errorf("set parts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Course{}, fmt.Errorf("commit transaction: %w", err)
	}

	updatedCourse, err := pb.GetCourse(ctx, CourseID{course.Code, course.Kind, course.Part})
	if err != nil {
		return Course{}, fmt.Errorf("get updated course: %w", err)
	}

	details := fmt.Sprintf("created course %s", course.ID())
	if err := pb.logAction(user, "CREATE", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return updatedCourse, nil
}

func (pb *PB) UpdateCourse(ctx context.Context, user string, id CourseID, partial PartialCourse) (Course, error) {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	id, err = ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	course, err := pb.mergeCourse(ctx, id, partial, tx)
	if err != nil {
		return Course{}, err
	}

	exists, err := pb.exists(ctx, id, tx)
	if err != nil {
		return Course{}, fmt.Errorf("failed to check course existence: %w", err)
	}

	if !exists {
		return Course{}, fmt.Errorf("course does not exists")
	}

	if _, err := tx.ExecContext(ctx, `
    UPDATE courses 
    SET code = ?, kind = ?, part = ?, parts = ?, name = ?, quantity = ?, total = ?, shown = ?, semester = ?
    WHERE code = ? AND kind = ? AND part = ?`,
		course.Code, course.Kind, course.Part, course.Parts,
		course.Name, course.Quantity, course.Total, course.Shown, course.Semester,
		id.Code, id.Kind, id.Part,
	); err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}

	if err := pb.setParts(ctx, CourseID{course.Code, course.Kind, course.Part}, tx); err != nil {
		return Course{}, fmt.Errorf("set parts: %w", err)
	}

	if partial.Code != nil || partial.Kind != nil || partial.Part != nil {
		newID := CourseID{
			Code: id.Code,
			Kind: id.Kind,
			Part: id.Part,
		}
		if partial.Code != nil {
			newID.Code = *partial.Code
		}
		if partial.Kind != nil {
			newID.Kind = *partial.Kind
		}
		if partial.Part != nil {
			newID.Part = *partial.Part
		}

		_, err = tx.ExecContext(ctx, `
            UPDATE pack_courses 
            SET course_code = ?, course_kind = ?, course_part = ?
            WHERE course_code = ? AND course_kind = ? AND course_part = ?`,
			newID.Code, newID.Kind, newID.Part,
			id.Code, id.Kind, id.Part)
		if err != nil {
			return Course{}, fmt.Errorf("update pack course references: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Course{}, fmt.Errorf("commit transaction: %w", err)
	}

	updatedCourse, err := pb.GetCourse(ctx, CourseID{course.Code, course.Kind, course.Part})
	if err != nil {
		return Course{}, fmt.Errorf("get updated course: %w", err)
	}

	details := fmt.Sprintf("updated course %s", course.ID())
	if err := pb.logAction(user, "UPDATE", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return updatedCourse, nil
}

func (pb *PB) DeleteCourse(ctx context.Context, user string, id CourseID) error {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	exists, err := pb.exists(ctx, id, tx)
	if err != nil {
		return fmt.Errorf("failed to check course existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("course does not exists")
	}

	var maxPart int
	err = tx.QueryRowContext(ctx, `
        SELECT COALESCE(MAX(part), 0)
        FROM courses 
        WHERE code = ? AND kind = ? AND part != ?`,
		id.Code, id.Kind, id.Part).Scan(&maxPart)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("get max part: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM pack_courses 
        WHERE course_code = ? AND course_kind = ? AND course_part = ?`,
		id.Code, id.Kind, id.Part)
	if err != nil {
		return fmt.Errorf("remove course from packs: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM courses
        WHERE code = ? AND KIND = ? AND part = ?`,
		id.Code, id.Kind, id.Part)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE courses
        SET parts = ?
        WHERE code = ? AND kind = ?`,
		maxPart, id.Code, id.Kind)
	if err != nil {
		return fmt.Errorf("update parts: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	details := fmt.Sprintf("deleted course %s", id.ID())
	if err := pb.logAction(user, "DELETE", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return nil
}

func (pb *PB) GetCourse(ctx context.Context, id CourseID) (Course, error) {
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

func (pb *PB) ListCourse(ctx context.Context, showHidden bool, filterSemester *string, filterCode *string, filterKind *string, filterPart *int) ([]Course, error) {
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

func (pb *PB) UpdateCourseQuantity(ctx context.Context, user string, id CourseID, delta int) (Course, error) {
	id, err := ValidateCourseID(id)
	if err != nil {
		return Course{}, err
	}

	current, err := pb.GetCourse(ctx, id)
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

	details := fmt.Sprintf("updated quantity of course %s", id.ID())
	if err := pb.logAction(user, "UPDATE QUANTITY", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return pb.GetCourse(ctx, id)
}

func (pb *PB) UpdateCourseShown(ctx context.Context, user string, id CourseID, shown bool) (Course, error) {
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

	details := fmt.Sprintf("updated visibility of course %s", id.ID())
	if err := pb.logAction(user, "UPDATE VISIBILITY", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return pb.GetCourse(ctx, id)
}

func (pb *PB) setParts(ctx context.Context, courseID CourseID, tx *sql.Tx) error {
	var maxPart int
	err := tx.QueryRowContext(ctx, `
    SELECT COALESCE(MAX(part), 0)
    FROM courses 
    WHERE code = ? AND kind = ?`,
		courseID.Code, courseID.Kind).Scan(&maxPart)
	if err != nil {
		return fmt.Errorf("get max part: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
    UPDATE courses
    SET parts = ?
    WHERE code = ? AND kind = ?`,
		maxPart, courseID.Code, courseID.Kind)
	if err != nil {
		return fmt.Errorf("update parts: %w", err)
	}
	return nil
}

func validateCourse(course Course) (Course, error) {
	// Validate Code
	course.Code = strings.TrimSpace(course.Code)
	if course.Code == "" {
		return Course{}, fmt.Errorf("CODE cannot be empty")
	}
	if ok, _ := regexp.MatchString(`^[A-Za-z0-9{},._-]+$`, course.Code); !ok {
		return Course{}, fmt.Errorf("CODE can only contain letters, numbers, and the characters {},._")
	}

	// Validate Kind
	course.Kind = strings.TrimSpace(course.Kind)
	if course.Kind == "" {
		return Course{}, fmt.Errorf("KIND cannot be empty")
	}
	switch course.Kind {
	case "TD", "Cours", "Memento", "TME":
		// valid
	default:
		return Course{}, fmt.Errorf("KIND must be one of: TD, Cours, Memento, TME")
	}

	// Validate Part
	if course.Part <= 0 || course.Part >= 1000 {
		return Course{}, fmt.Errorf("PART must be in 1-1000")
	}

	// Validate Name
	course.Name = strings.TrimSpace(course.Name)

	// Validate Quantities
	if err := validateQuantity(course.Quantity, course.Total); err != nil {
		return Course{}, fmt.Errorf("invalid quantities: %w", err)
	}

	// Validate Semester
	course.Semester = strings.TrimSpace(course.Semester)
	switch course.Semester {
	case "S1", "S2":
		// valid
	default:
		return Course{}, fmt.Errorf("SEMESTER must be either S1 or S2")
	}

	return course, nil
}

func (pb *PB) mergeCourse(ctx context.Context, id CourseID, partial PartialCourse, tx *sql.Tx) (Course, error) {
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

	current, err := pb.getCourse(ctx, id, tx)
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
