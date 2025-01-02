package tests

import (
	"database/sql"
	"testing"

	"git.sr.ht/~alias/polybase/internal"
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
)`

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

	// Ensure database is closed after test
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	})

	// Create schema
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
	_, err := db.Exec("DELETE FROM courses")
	if err != nil {
		db.t.Fatalf("failed to clear test database: %v", err)
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
