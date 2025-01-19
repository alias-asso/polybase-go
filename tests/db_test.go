package tests

import (
	"database/sql"
	"testing"

	"git.sr.ht/~alias/polybase-go/internal"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS courses (
    code TEXT,
    kind TEXT,
    part INTEGER DEFAULT 1,
    parts INTEGER DEFAULT 1,
    name TEXT,
    quantity INTEGER,
    total INTEGER,
    shown INTEGER DEFAULT 1,
    semester TEXT,
    PRIMARY KEY (code, kind, part)
);

CREATE TABLE IF NOT EXISTS packs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS pack_courses (
    pack_id INTEGER,
    course_code TEXT,
    course_kind TEXT,
    course_part INTEGER,
    FOREIGN KEY (pack_id) REFERENCES packs(id) ON DELETE CASCADE,
    FOREIGN KEY (course_code, course_kind, course_part) 
        REFERENCES courses(code, kind, part) ON UPDATE CASCADE,
    PRIMARY KEY (pack_id, course_code, course_kind, course_part)
);`

// DB encapsulates a test database connection and test helper functions
type DB struct {
	*sql.DB
	t *testing.T
}

// NewDB creates a new in-memory SQLite database for testing
func NewDB(t *testing.T) *DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return &DB{DB: db, t: t}
}

// Insert adds a course to the test database
func (db *DB) Insert(c internal.Course) {
	db.t.Helper()
	shown := 0
	if c.Shown {
		shown = 1
	}

	_, err := db.Exec(`
		INSERT INTO courses (code, kind, part, parts, name, quantity, total, shown, semester)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Code, c.Kind, c.Part, c.Parts, c.Name, c.Quantity, c.Total, shown, c.Semester)
	if err != nil {
		db.t.Fatalf("failed to insert test course: %v", err)
	}
}

// InsertMany adds multiple courses to the test database
func (db *DB) InsertMany(courses []internal.Course) {
	db.t.Helper()
	for _, c := range courses {
		db.Insert(c)
	}
}

// Clear removes all data from the test database
func (db *DB) Clear() {
	db.t.Helper()
	_, err := db.Exec("DELETE FROM pack_courses")
	if err != nil {
		db.t.Fatalf("failed to clear pack_courses: %v", err)
	}
	_, err = db.Exec("DELETE FROM packs")
	if err != nil {
		db.t.Fatalf("failed to clear packs: %v", err)
	}
	_, err = db.Exec("DELETE FROM courses")
	if err != nil {
		db.t.Fatalf("failed to clear courses: %v", err)
	}
}

// Count returns the number of courses in the test database
func (db *DB) Count() int {
	db.t.Helper()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&count)
	if err != nil {
		db.t.Fatalf("failed to count courses: %v", err)
	}
	return count
}

// Get retrieves a single course from the database
func (db *DB) Get(id internal.CourseID) internal.Course {
	db.t.Helper()
	var c internal.Course
	var shown int

	err := db.QueryRow(`
		SELECT code, kind, part, parts, name, quantity, total, shown, semester
		FROM courses
		WHERE code = ? AND kind = ? AND part = ?`,
		id.Code, id.Kind, id.Part).Scan(
		&c.Code, &c.Kind, &c.Part, &c.Parts,
		&c.Name, &c.Quantity, &c.Total, &shown, &c.Semester)

	if err != nil {
		db.t.Fatalf("failed to get course: %v", err)
	}

	c.Shown = shown == 1
	return c
}

// AssertCount checks if the number of courses matches the expected count
func (db *DB) AssertCount(want int) {
	db.t.Helper()
	got := db.Count()
	if got != want {
		db.t.Errorf("course count = %d, want %d", got, want)
	}
}

// AssertExists checks if a course exists in the database
func (db *DB) AssertExists(id internal.CourseID) {
	db.t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM courses 
			WHERE code = ? AND kind = ? AND part = ?
		)`,
		id.Code, id.Kind, id.Part).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check course existence: %v", err)
	}
	if !exists {
		db.t.Errorf("course %v does not exist", id)
	}
}

// AssertNotExists checks if a course does not exist in the database
func (db *DB) AssertNotExists(id internal.CourseID) {
	db.t.Helper()
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM courses 
			WHERE code = ? AND kind = ? AND part = ?
		)`,
		id.Code, id.Kind, id.Part).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check course existence: %v", err)
	}
	if exists {
		db.t.Errorf("course %v exists but should not", id)
	}
}

// AssertCourseEqual compares a course with what's in the database
func (db *DB) AssertCourseEqual(id internal.CourseID, want internal.Course) {
	db.t.Helper()
	got := db.Get(id)
	if got != want {
		db.t.Errorf("course mismatch\ngot: %+v\nwant: %+v", got, want)
	}
}

// InsertPack adds a pack to the test database
func (db *DB) InsertPack(p internal.Pack) {
	db.t.Helper()

	result, err := db.Exec(`
		INSERT INTO packs (id, name)
		VALUES (?, ?)`,
		p.ID, p.Name)
	if err != nil {
		db.t.Fatalf("failed to insert test pack: %v", err)
	}

	// If no ID was provided, get the auto-generated one
	if p.ID == 0 {
		id, err := result.LastInsertId()
		if err != nil {
			db.t.Fatalf("failed to get last insert id: %v", err)
		}
		p.ID = int(id)
	}

	// Insert pack courses if any
	for _, courseID := range p.Courses {
		db.InsertPackCourse(p.ID, courseID)
	}
}

// InsertPackCourse links a course to a pack
func (db *DB) InsertPackCourse(packID int, courseID internal.CourseID) {
	db.t.Helper()

	_, err := db.Exec(`
		INSERT INTO pack_courses (pack_id, course_code, course_kind, course_part)
		VALUES (?, ?, ?, ?)`,
		packID, courseID.Code, courseID.Kind, courseID.Part)
	if err != nil {
		db.t.Fatalf("failed to insert pack course: %v", err)
	}
}

// GetPack retrieves a pack from the database
func (db *DB) GetPack(id int) internal.Pack {
	db.t.Helper()

	var pack internal.Pack
	err := db.QueryRow(`
		SELECT id, name
		FROM packs
		WHERE id = ?`, id).Scan(&pack.ID, &pack.Name)
	if err != nil {
		db.t.Fatalf("failed to get pack: %v", err)
	}

	// Get pack courses
	rows, err := db.Query(`
		SELECT course_code, course_kind, course_part
		FROM pack_courses
		WHERE pack_id = ?
		ORDER BY course_code, course_kind, course_part`, id)
	if err != nil {
		db.t.Fatalf("failed to get pack courses: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var courseID internal.CourseID
		if err := rows.Scan(&courseID.Code, &courseID.Kind, &courseID.Part); err != nil {
			db.t.Fatalf("failed to scan pack course: %v", err)
		}
		pack.Courses = append(pack.Courses, courseID)
	}

	if err = rows.Err(); err != nil {
		db.t.Fatalf("error iterating pack courses: %v", err)
	}

	return pack
}

// AssertPackExists checks if a pack exists
func (db *DB) AssertPackExists(id int) {
	db.t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM packs 
			WHERE id = ?
		)`, id).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check pack existence: %v", err)
	}
	if !exists {
		db.t.Errorf("pack %d does not exist", id)
	}
}

// AssertPackNotExists checks if a pack doesn't exist
func (db *DB) AssertPackNotExists(id int) {
	db.t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM packs 
			WHERE id = ?
		)`, id).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check pack existence: %v", err)
	}
	if exists {
		db.t.Errorf("pack %d exists but should not", id)
	}
}

// AssertPackEqual compares a pack with database content
func (db *DB) AssertPackEqual(id int, want internal.Pack) {
	db.t.Helper()

	got := db.GetPack(id)

	// Compare non-course fields
	if got.ID != want.ID || got.Name != want.Name {
		db.t.Errorf("pack basic fields mismatch\ngot: ID=%d, Name=%q\nwant: ID=%d, Name=%q",
			got.ID, got.Name, want.ID, want.Name)
		return
	}

	// Check if course counts match
	if len(got.Courses) != len(want.Courses) {
		db.t.Errorf("course count mismatch\ngot %d courses, want %d courses",
			len(got.Courses), len(want.Courses))
		return
	}

	// Convert courses to maps for set comparison
	wantCourses := make(map[string]internal.CourseID)
	for _, course := range want.Courses {
		wantCourses[course.ID()] = course
	}

	gotCourses := make(map[string]internal.CourseID)
	for _, course := range got.Courses {
		gotCourses[course.ID()] = course
	}

	// Check for missing courses
	for id, wantCourse := range wantCourses {
		if _, exists := gotCourses[id]; !exists {
			db.t.Errorf("missing course: %+v", wantCourse)
		}
	}

	// Check for unexpected extra courses
	for id, gotCourse := range gotCourses {
		if _, exists := wantCourses[id]; !exists {
			db.t.Errorf("unexpected course: %+v", gotCourse)
		}
	}
}

// AssertCourseInPack checks if a course is in a pack
func (db *DB) AssertCourseInPack(packID int, courseID internal.CourseID) {
	db.t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pack_courses 
			WHERE pack_id = ? 
			AND course_code = ? 
			AND course_kind = ? 
			AND course_part = ?
		)`,
		packID, courseID.Code, courseID.Kind, courseID.Part).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check course in pack: %v", err)
	}
	if !exists {
		db.t.Errorf("course %+v not found in pack %d", courseID, packID)
	}
}

// AssertCourseNotInPack checks if a course is not in a pack
func (db *DB) AssertCourseNotInPack(packID int, courseID internal.CourseID) {
	db.t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pack_courses 
			WHERE pack_id = ? 
			AND course_code = ? 
			AND course_kind = ? 
			AND course_part = ?
		)`,
		packID, courseID.Code, courseID.Kind, courseID.Part).Scan(&exists)
	if err != nil {
		db.t.Fatalf("failed to check course not in pack: %v", err)
	}
	if exists {
		db.t.Errorf("course %+v found in pack %d but should not be", courseID, packID)
	}
}

// CountPackCourses counts courses in a pack
func (db *DB) CountPackCourses(packID int) int {
	db.t.Helper()

	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM pack_courses 
		WHERE pack_id = ?`,
		packID).Scan(&count)
	if err != nil {
		db.t.Fatalf("failed to count pack courses: %v", err)
	}
	return count
}
