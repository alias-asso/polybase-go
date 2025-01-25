package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

func (pb *PB) CreatePack(ctx context.Context, user string, name string, courses []CourseID) (Pack, error) {
	if err := validatePack(name, courses); err != nil {
		return Pack{}, err
	}

	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Pack{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	for _, courseID := range courses {
		exists, err := pb.exists(ctx, courseID, tx)
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
      VALUES (?, ?, ?, ?)`,
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
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

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
			exists, err := pb.exists(ctx, courseID, tx)
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
    ORDER BY c.code,
    CASE c.kind 
        WHEN 'Memento' THEN 1
        WHEN 'TME' THEN 2
        WHEN 'Cours' THEN 3
        WHEN 'TD' THEN 4
        ELSE 5
    END,
    c.part`, id)
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
	// Get packs ordered by ID
	rows, err := pb.db.QueryContext(ctx, `
        SELECT id, name, course_code, course_kind, course_part
        FROM packs 
        LEFT JOIN pack_courses ON packs.id = pack_courses.pack_id
        ORDER BY packs.id, course_code, course_kind, course_part`)
	if err != nil {
		return nil, fmt.Errorf("list packs: %w", err)
	}
	defer rows.Close()

	var packs []Pack
	var currentPack *Pack

	for rows.Next() {
		var id int
		var name string
		var code, kind sql.NullString
		var part sql.NullInt64

		if err := rows.Scan(&id, &name, &code, &kind, &part); err != nil {
			return nil, fmt.Errorf("scan pack: %w", err)
		}

		// Start new pack if ID changes
		if currentPack == nil || currentPack.ID != id {
			packs = append(packs, Pack{
				ID:   id,
				Name: name,
			})
			currentPack = &packs[len(packs)-1]
		}

		// Add course if one exists for this row
		if code.Valid && kind.Valid && part.Valid {
			currentPack.Courses = append(currentPack.Courses, CourseID{
				Code: code.String,
				Kind: kind.String,
				Part: int(part.Int64),
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate packs: %w", err)
	}

	return packs, nil
}

func (pb *PB) UpdatePackQuantity(ctx context.Context, user string, id int, delta int) (Pack, error) {
	tx, err := pb.db.BeginTx(ctx, nil)
	if err != nil {
		return Pack{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()
	rows, err := tx.QueryContext(ctx, `
    SELECT c.code, c.kind, c.part, c.quantity, c.total
    FROM courses c
    JOIN pack_courses pc ON c.code = pc.course_code
      AND c.kind = pc.course_kind
      AND c.part = pc.course_part
    WHERE pc.pack_id = ?`, id)
	if err != nil {
		return Pack{}, fmt.Errorf("get pack courses: %w", err)
	}
	defer rows.Close()

	type courseUpdate struct {
		id       CourseID
		quantity int
		total    int
		delta    int // Store per-course delta
	}
	var coursesToUpdate []courseUpdate

	// First pass: validate all updates and compute per-course deltas
	for rows.Next() {
		var code, kind string
		var part, quantity, total int
		if err := rows.Scan(&code, &kind, &part, &quantity, &total); err != nil {
			return Pack{}, fmt.Errorf("scan course: %w", err)
		}

		// Calculate adjusted delta for this course
		courseDelta := delta
		if delta < 0 {
			// If reducing would go below 0, adjust delta to hit exactly 0
			if quantity+delta < 0 {
				courseDelta = -quantity
			}
		} else if quantity+delta > total {
			// Check upper bound
			return Pack{}, fmt.Errorf("quantity would exceed total for course %s/%s/%d", code, kind, part)
		}

		coursesToUpdate = append(coursesToUpdate, courseUpdate{
			id:       CourseID{Code: code, Kind: kind, Part: part},
			quantity: quantity,
			total:    total,
			delta:    courseDelta,
		})
	}
	if err = rows.Err(); err != nil {
		return Pack{}, fmt.Errorf("iterate courses: %w", err)
	}
	if len(coursesToUpdate) == 0 {
		return pb.GetPack(ctx, id)
	}

	// Second pass: apply the updates
	for _, course := range coursesToUpdate {
		_, err = tx.ExecContext(ctx, `
      UPDATE courses
      SET quantity = quantity + ?
      WHERE code = ? AND kind = ? AND part = ?`,
			course.delta, course.id.Code, course.id.Kind, course.id.Part)
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
