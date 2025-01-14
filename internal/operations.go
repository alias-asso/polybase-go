package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func (c CourseID) ID() string {
	return fmt.Sprintf("%s/%s/%d", c.Code, c.Kind, c.Part)
}

func (c Course) SID() string {
	sid := fmt.Sprintf("course-%s-%s-%d", c.Code, c.Kind, c.Part)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sid = reg.ReplaceAllString(sid, "-")
	return strings.ToLower(sid)
}

func (pb *PB) CreateCourse(ctx context.Context, user string, course Course) (Course, error) {
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

	if err := pb.setParts(ctx, CourseID{course.Code, course.Kind, course.Part}); err != nil {
		return Course{}, fmt.Errorf("set parts: %w", err)
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

func (pb *PB) UpdateCourse(ctx context.Context, user string, id CourseID, partial PartialCourse) (Course, error) {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	id, err = ValidateCourseID(id)
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

	if err := pb.setParts(ctx, CourseID{course.Code, course.Kind, course.Part}); err != nil {
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
	defer tx.Rollback()

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

func (pb *PB) UpdateQuantity(ctx context.Context, user string, id CourseID, delta int) (Course, error) {
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

func (pb *PB) UpdateShown(ctx context.Context, user string, id CourseID, shown bool) (Course, error) {
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

func (pb *PB) CreatePack(ctx context.Context, user string, name string, courses []CourseID) (Pack, error) {
	if err := validatePack(name, courses); err != nil {
		return Pack{}, err
	}

	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Pack{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, courseID := range courses {
		exists, err := pb.exists(ctx, courseID)
		if err != nil {
			return Pack{}, fmt.Errorf("check course existence: %w", err)
		}
		if !exists {
			return Pack{}, fmt.Errorf("course %s does not exist", courseID.ID())
		}
	}

	result, err := tx.ExecContext(ctx, `
    INSERT INTO packs (name) VALUES (?)`,
		strings.TrimSpace(name))
	if err != nil {
		return Pack{}, fmt.Errorf("create pack: %w", err)
	}

	packID, err := result.LastInsertId()
	if err != nil {
		return Pack{}, fmt.Errorf("get pack id: %w", err)
	}

	for _, courseID := range courses {
		_, err = tx.ExecContext(ctx, `
      INSERT INTO pack_courses (pack_id, course_code, course_kind, course_part)
      VALUE (?, ?, ?, ?)`,
			packID, courseID.Code, courseID.Kind, courseID.Part)
		if err != nil {
			return Pack{}, fmt.Errorf("add course to pack: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Pack{}, fmt.Errorf("commit transaction: %w", err)
	}
	details := fmt.Sprintf("created pack %d with %d courses", packID, len(courses))
	if err := pb.logAction(user, "CREATE PACK", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return pb.GetPack(ctx, int(packID))
}

func (pb *PB) UpdatePack(ctx context.Context, user string, id int, partial PartialPack) (Pack, error) {
	if partial.Name == nil && partial.Courses == nil {
		return Pack{}, fmt.Errorf("at least one field must be updated")
	}

	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Pack{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM PACKS WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return Pack{}, fmt.Errorf("check pack existence: %w", err)
	}

	if !exists {
		return Pack{}, fmt.Errorf("pack not found")
	}

	if partial.Name != nil {
		if strings.TrimSpace(*partial.Name) == "" {
			return Pack{}, fmt.Errorf("pack name cannot be empty")
		}
		_, err = tx.ExecContext(ctx, "UPDATE packs SET name = ? WHERE id = ?",
			strings.TrimSpace(*partial.Name), id)
		if err != nil {
			return Pack{}, fmt.Errorf("update pack name: %w", err)
		}
	}

	if partial.Courses != nil {
		if len(*partial.Courses) == 0 {
			return Pack{}, fmt.Errorf("pack must contain at least one course")
		}

		for _, courseID := range *partial.Courses {
			exists, err := pb.exists(ctx, courseID)
			if err != nil {
				return Pack{}, fmt.Errorf("check course existence: %w", err)
			}
			if !exists {
				return Pack{}, fmt.Errorf("course %s does not exist", courseID.ID())
			}
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM pack_courses WHERE pack_id = ?", id)
		if err != nil {
			return Pack{}, fmt.Errorf("remove existing courses: %w", err)
		}
		for _, courseID := range *partial.Courses {
			_, err = tx.ExecContext(ctx, `
          INSERT INTO pack_courses (pack_id, course_code, course_kind, course_part)
          VALUES (?, ?, ?, ?)`,
				id, courseID.Code, courseID.Kind, courseID.Part)
			if err != nil {
				return Pack{}, fmt.Errorf("add course to pack: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Pack{}, fmt.Errorf("commit transaction: %w", err)
	}

	details := fmt.Sprintf("updated pack %d", id)
	if err := pb.logAction(user, "UPDATE PACK", details); err != nil {
		log.Printf("Warning: failed to log action :%v", err)
	}

	return pb.GetPack(ctx, id)
}

func (pb *PB) DeletePack(ctx context.Context, user string, id int) error {
	var exists bool
	err := pb.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM packs WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check pack existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("pack not found")
	}

	_, err = pb.db.ExecContext(ctx, "DELETE FROM packs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete pack: %w", err)
	}

	details := fmt.Sprintf("deleted pack %d", id)
	if err := pb.logAction(user, "DELETE PACK", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return nil
}

func (pb *PB) GetPack(ctx context.Context, id int) (Pack, error) {
	var pack Pack
	err := pb.db.QueryRowContext(ctx, `
    SELECT id, name
    FROM packs
    WHERE id = ?`, id).Scan(&pack.ID, &pack.Name)
	if err == sql.ErrNoRows {
		return Pack{}, fmt.Errorf("pack not found")
	}
	if err != nil {
		return Pack{}, fmt.Errorf("get pack: %w", err)
	}

	rows, err := pb.db.QueryContext(ctx, `
    SELECT c.code, c.kind, c.part, c.parts, c.name, c.quantity, c.total, c.shown, c.semester
    FROM courses c
    JOIN pack_courses pc ON c.code = pc.course_code
      AND c.kind = pc.course_kind
      AND c.part = pc.course_part
    WHERE pc.pack_id = ?
    ORDER BY c.code, c.kind, c.part`, id)
	if err != nil {
		return Pack{}, fmt.Errorf("get pack courses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var course Course
		if err := rows.Scan(
			&course.Code, &course.Kind, &course.Part, &course.Parts,
			&course.Name, &course.Quantity, &course.Total, &course.Shown,
			&course.Semester); err != nil {
			return Pack{}, fmt.Errorf("scan course: %w", err)
		}
		pack.Courses = append(pack.Courses, CourseID{
			Code: course.Code,
			Kind: course.Kind,
			Part: course.Part,
		})
	}

	if err = rows.Err(); err != nil {
		return Pack{}, fmt.Errorf("iterate courses: %w", err)
	}

	return pack, nil
}

func (pb *PB) ListPacks(ctx context.Context) ([]Pack, error) {
	rows, err := pb.db.QueryContext(ctx, `
    SELECT id, name
    FROM packs
    ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list packs: %w", err)
	}
	defer rows.Close()

	var packs []Pack
	packMap := make(map[int]*Pack)

	var packIDs []int
	for rows.Next() {
		var pack Pack
		if err := rows.Scan(&pack.ID, &pack.Name); err != nil {
			return nil, fmt.Errorf("scan pack: %w", err)
		}
		packs = append(packs, pack)
		packMap[pack.ID] = &packs[len(packs)-1]
		packIDs = append(packIDs, pack.ID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate packs: %w", err)
	}

	if len(packs) == 0 {
		return packs, nil
	}

	rows, err = pb.db.QueryContext(ctx, `
      SELECT 
          pc.pack_id,
          c.code, c.kind, c.part, c.parts, c.name, 
          c.quantity, c.total, c.shown, c.semester
      FROM courses c
      JOIN pack_courses pc ON c.code = pc.course_code 
          AND c.kind = pc.course_kind 
          AND c.part = pc.course_part
      WHERE pc.pack_id IN (`+strings.Repeat("?,", len(packIDs)-1)+"?"+`)
      ORDER BY pc.pack_id, c.code, c.kind, c.part`,
		interface{}(packIDs))
	if err != nil {
		return nil, fmt.Errorf("get pack courses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var packID int
		var course Course
		if err := rows.Scan(
			&packID,
			&course.Code, &course.Kind, &course.Part, &course.Parts,
			&course.Name, &course.Quantity, &course.Total, &course.Shown,
			&course.Semester); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}

		pack := packMap[packID]
		pack.Courses = append(pack.Courses, CourseID{
			Code: course.Code,
			Kind: course.Kind,
			Part: course.Part,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate course: %w", err)
	}

	return packs, nil
}

func (pb *PB) UpdatePackQuantity(ctx context.Context, user string, id int, delta int) (Pack, error) {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Pack{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
    SELECT course_code, course_kind, course_part, quantity, total
    FROM pack_courses pc
    JOIN courses c ON pc.course_code = c.code
      AND pc.course_kind = c.kind
      AND pc.course_part = c.part
    WHERE pack_id = ?
    FOR UPDATE`, id)
	if err != nil {
		return Pack{}, fmt.Errorf("get pack courses: %w", err)
	}
	defer rows.Close()

	var coursesToUpdate []CourseID
	for rows.Next() {
		var code, kind string
		var part, quantity, total int
		if err := rows.Scan(&code, &kind, &part, &quantity, &total); err != nil {
			return Pack{}, fmt.Errorf("scan course: %w", err)
		}

		if delta < 0 && quantity+delta < 0 {
			continue
		}

		coursesToUpdate = append(coursesToUpdate, CourseID{
			Code: code,
			Kind: kind,
			Part: part,
		})
	}
	if err = rows.Err(); err != nil {
		return Pack{}, fmt.Errorf("iterate courses: %w", err)
	}

	if len(coursesToUpdate) == 0 {
		return pb.GetPack(ctx, id)
	}

	for _, courseID := range coursesToUpdate {
		_, err = tx.ExecContext(ctx, `
      UPDATE courses
      SET quantity = quantity + ?
      WHERE code = ? AND kind = ? AND part = ?`,
			delta, courseID.Code, courseID.Kind, courseID.Part)
		if err != nil {
			return Pack{}, fmt.Errorf("update course quantity: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return Pack{}, fmt.Errorf("commit transaction: %w", err)
	}

	details := fmt.Sprintf("updated quantities for pack %d by %d", id, delta)
	if err := pb.logAction(user, "UPDATE PACK QUANTITY", details); err != nil {
		log.Printf("Warning: failed to log action: %v", err)
	}

	return pb.GetPack(ctx, id)
}

func validatePack(name string, courses []CourseID) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("pack name cannot be empty")
	}

	if len(courses) == 0 {
		return fmt.Errorf("pack must contain at least one course")
	}

	seen := make(map[string]bool)
	for _, id := range courses {
		if seen[id.ID()] {
			return fmt.Errorf("duplicate course in pack: %s", id.ID())
		}
		seen[id.ID()] = true
	}

	return nil
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

	current, err := pb.GetCourse(ctx, id)
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

func (pb *PB) setParts(ctx context.Context, courseID CourseID) error {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	var maxPart int
	err = tx.QueryRowContext(ctx, `
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
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
