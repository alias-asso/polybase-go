package internal

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
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

func (c Course) ID() string {
	return fmt.Sprintf("%s/%s/%d", c.Code, c.Kind, c.Part)
}

func (c Course) SID() string {
	fullID := fmt.Sprintf("%s-%s-%d", c.Code, c.Kind, c.Part)
	hasher := sha256.New()
	hasher.Write([]byte(fullID))
	hash := hasher.Sum(nil)
	encoded := strings.ToLower(base32.StdEncoding.EncodeToString(hash))

	// If the first character is a number (2-7 in base32), prepend 'a'
	if encoded[0] >= '2' && encoded[0] <= '7' {
		return "a" + encoded[:7]
	}

	return encoded[:8]
}

func (pb *PB) Create(ctx context.Context, course Course) (Course, error) {
	shown := 0
	if course.Shown {
		shown = 1
	}
	if _, err := pb.db.ExecContext(ctx, `
    INSERT INTO courses (code, kind, part, parts, name, quantity, total, shown, semester)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		course.Code, course.Kind, course.Part, course.Parts, course.Name, course.Quantity, course.Total, shown, course.Semester); err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}
	return course, nil
}

func (pb *PB) Get(ctx context.Context, id CourseID) (Course, error) {
	var course Course
	var shown int
	err := pb.db.QueryRowContext(ctx, `
    SELECT code, kind, part, parts, name, quantity, total, shown, semester
    FROM courses
    WHERE code = ? AND kind = ? AND part = ?`,
		id.Code, id.Kind, id.Part).Scan(
		&course.Code, &course.Kind, &course.Part, &course.Parts,
		&course.Name, &course.Quantity, &course.Total, &shown, &course.Semester)
	if err != nil {
		return Course{}, fmt.Errorf("get course: %w", err)
	}
	course.Shown = shown == 1
	return course, nil
}

func (pb *PB) Update(ctx context.Context, id CourseID, course Course) (Course, error) {
	shown := 0
	if course.Shown {
		shown = 1
	}
	if _, err := pb.db.ExecContext(ctx, `
        UPDATE courses 
        SET code = ?, kind = ?, part = ?, parts = ?, name = ?, quantity = ?, total = ?, shown = ?, semester = ?
        WHERE code = ? AND kind = ? AND part = ?`,
		course.Code, course.Kind, course.Part, course.Parts,
		course.Name, course.Quantity, course.Total, shown, course.Semester,
		id.Code, id.Kind, id.Part,
	); err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}
	return course, nil
}

func (pb *PB) Delete(ctx context.Context, id CourseID) error {
	if _, err := pb.db.ExecContext(ctx, `
    DELETE FROM courses
    WHERE code = ? AND KIND = ? AND part = ?`,
		id.Code, id.Kind, id.Part); err != nil {
		return fmt.Errorf("delete course: %w", err)
	}

	return nil
}

func (pb *PB) List(ctx context.Context, showHidden bool) ([]Course, error) {
	var courses []Course
	query := `SELECT code, kind, part, parts, name, quantity, total, shown, semester FROM courses`
	if !showHidden {
		query += ` WHERE shown = 1`
	}

	rows, err := pb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Course
		err := rows.Scan(&c.Code, &c.Kind, &c.Part, &c.Parts, &c.Name,
			&c.Quantity, &c.Total, &c.Shown, &c.Semester)
		if err != nil {
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
	_, err := pb.db.ExecContext(ctx, `
        UPDATE courses 
        SET quantity = quantity + ?
        WHERE code = ? AND kind = ? AND part = ?`,
		delta, id.Code, id.Kind, id.Part,
	)
	if err != nil {
		return Course{}, fmt.Errorf("update quantity: %w", err)
	}
	return pb.Get(ctx, id)
}

func (pb *PB) UpdateShown(ctx context.Context, id CourseID, shown bool) (Course, error) {
	shownInt := 0
	if shown {
		shownInt = 1
	}
	_, err := pb.db.ExecContext(ctx, `
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
