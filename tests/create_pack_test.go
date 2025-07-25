package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// A new pack can be created with valid courses
func TestCreatePackWithValidCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Cours",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "TME",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 30,
			Total:    60,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS103",
			Kind:     "TD",
			Part:     1,
			Parts:    1,
			Name:     "Programming Tutorial",
			Quantity: 20,
			Total:    40,
			Shown:    true,
			Semester: "S1",
		},
	}

	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	tests := []struct {
		name     string
		packName string
		courses  []libpolybase.CourseID
	}{
		{
			name:     "single course pack",
			packName: "Programming Basics",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
		{
			name:     "multiple course pack",
			packName: "Complete Programming",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS103", Kind: "TD", Part: 1},
			},
		},
		{
			name:     "pack with spaces in name",
			packName: "   Programming Pack   ",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := pb.CreatePack(ctx, "testuser", tt.packName, tt.courses)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}

			want := libpolybase.Pack{
				ID:      created.ID,
				Name:    strings.TrimSpace(tt.packName),
				Courses: tt.courses,
			}

			db.AssertPackExists(created.ID)
			db.AssertPackEqual(created.ID, want)

			for _, courseID := range tt.courses {
				db.AssertCourseInPack(created.ID, courseID)
			}

			if got := db.CountPackCourses(created.ID); got != len(tt.courses) {
				t.Errorf("pack course count = %d, want %d", got, len(tt.courses))
			}
		})
	}
}

// Creating a pack with no courses is properly rejected
func TestCreatePackWithNoCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	tests := []struct {
		name     string
		packName string
		courses  []libpolybase.CourseID
	}{
		{
			name:     "empty course list",
			packName: "Empty Pack",
			courses:  []libpolybase.CourseID{},
		},
		{
			name:     "nil course list",
			packName: "Nil Pack",
			courses:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pb.CreatePack(ctx, "testuser", tt.packName, tt.courses)
			if err == nil {
				t.Fatal("expected error when creating pack with no courses, got nil")
			}

			if !strings.Contains(err.Error(), "pack must contain at least one course") {
				t.Errorf("got error %q, want error containing 'pack must contain at least one course'", err.Error())
			}

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&count)
			if err != nil {
				t.Fatalf("failed to count packs: %v", err)
			}

			if count != 0 {
				t.Errorf("pack count = %d, want 0", count)
			}

			err = db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&count)
			if err != nil {
				t.Fatalf("failed to count pack courses: %v", err)
			}

			if count != 0 {
				t.Errorf("pack_courses count = %d, want 0", count)
			}
		})
	}
}

// Creating a pack with non-existent courses fails properly
func TestCreatePackWithNonExistentCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	existingCourse := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", existingCourse)
	if err != nil {
		t.Fatalf("failed to create test course: %v", err)
	}

	existingID := libpolybase.CourseID{
		Code: existingCourse.Code,
		Kind: existingCourse.Kind,
		Part: existingCourse.Part,
	}

	tests := []struct {
		name     string
		packName string
		courses  []libpolybase.CourseID
	}{
		{
			name:     "single non-existent course",
			packName: "Invalid Pack",
			courses: []libpolybase.CourseID{
				{Code: "FAKE101", Kind: "Missing", Part: 1},
			},
		},
		{
			name:     "multiple non-existent courses",
			packName: "Invalid Pack Multiple",
			courses: []libpolybase.CourseID{
				{Code: "FAKE101", Kind: "Missing", Part: 1},
				{Code: "FAKE102", Kind: "Missing", Part: 1},
			},
		},
		{
			name:     "mix of existing and non-existent courses",
			packName: "Mixed Pack",
			courses: []libpolybase.CourseID{
				existingID,
				{Code: "FAKE101", Kind: "Missing", Part: 1},
			},
		},
		{
			name:     "existing code with wrong kind",
			packName: "Wrong Kind Pack",
			courses: []libpolybase.CourseID{
				{Code: existingID.Code, Kind: "WrongKind", Part: existingID.Part},
			},
		},
		{
			name:     "existing code and kind with wrong part",
			packName: "Wrong Part Pack",
			courses: []libpolybase.CourseID{
				{Code: existingID.Code, Kind: existingID.Kind, Part: 999},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pb.CreatePack(ctx, "testuser", tt.packName, tt.courses)
			if err == nil {
				t.Fatal("expected error when creating pack with non-existent courses, got nil")
			}

			if !strings.Contains(err.Error(), "does not exist") {
				t.Errorf("got error %q, want error containing 'does not exist'", err.Error())
			}

			db.AssertCount(1)
			var packCount int
			err = db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&packCount)
			if err != nil {
				t.Fatalf("failed to count packs: %v", err)
			}

			if packCount != 0 {
				t.Errorf("pack count = %d, want 0", packCount)
			}

			var linkCount int
			err = db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&linkCount)
			if err != nil {
				t.Fatalf("failed to count pack courses: %v", err)
			}

			if linkCount != 0 {
				t.Errorf("pack_courses count = %d, want 0", linkCount)
			}
		})
	}
}

// Creating a pack with duplicate courses fails properly
func TestCreatePackWithDuplicateCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Cours",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "TME",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 30,
			Total:    60,
			Shown:    true,
			Semester: "S1",
		},
	}

	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create test course: %v", err)
		}
	}

	tests := []struct {
		name     string
		packName string
		courses  []libpolybase.CourseID
	}{
		{
			name:     "exact duplicate course",
			packName: "Duplicate Pack",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
		{
			name:     "multiple duplicates",
			packName: "Multiple Duplicates Pack",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
			},
		},
		{
			name:     "duplicate with valid courses",
			packName: "Mixed Pack",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pb.CreatePack(ctx, "testuser", tt.packName, tt.courses)
			if err == nil {
				t.Fatal("expected error when creating pack with duplicate courses, got nil")
			}

			if !strings.Contains(err.Error(), "duplicate course in pack") {
				t.Errorf("got error %q, want error containing 'duplicate course in pack'", err.Error())
			}

			var packCount int
			err = db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&packCount)
			if err != nil {
				t.Fatalf("failed to count packs: %v", err)
			}

			if packCount != 0 {
				t.Errorf("pack count = %d, want 0", packCount)
			}

			var linkCount int
			err = db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&linkCount)
			if err != nil {
				t.Fatalf("failed to count pack courses: %v", err)
			}

			if linkCount != 0 {
				t.Errorf("pack_courses count = %d, want 0", linkCount)
			}
		})
	}
}

// Pack name is properly trimmed during creation
func TestCreatePackNameTrimming(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	course := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create test course: %v", err)
	}

	courseID := libpolybase.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}

	tests := []struct {
		name      string
		inputName string
		wantName  string
		courses   []libpolybase.CourseID
	}{
		{
			name:      "leading spaces",
			inputName: "   Programming Pack",
			wantName:  "Programming Pack",
			courses:   []libpolybase.CourseID{courseID},
		},
		{
			name:      "trailing spaces",
			inputName: "Programming Pack   ",
			wantName:  "Programming Pack",
			courses:   []libpolybase.CourseID{courseID},
		},
		{
			name:      "leading and trailing spaces",
			inputName: "   Programming Pack   ",
			wantName:  "Programming Pack",
			courses:   []libpolybase.CourseID{courseID},
		},
		{
			name:      "multiple libpolybase spaces preserved",
			inputName: "   Programming    Pack   ",
			wantName:  "Programming    Pack",
			courses:   []libpolybase.CourseID{courseID},
		},
		{
			name:      "tabs and newlines",
			inputName: "\tProgramming\nPack\t",
			wantName:  "Programming\nPack",
			courses:   []libpolybase.CourseID{courseID},
		},
		{
			name:      "only whitespace",
			inputName: "     ",
			wantName:  "",
			courses:   []libpolybase.CourseID{courseID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := pb.CreatePack(ctx, "testuser", tt.inputName, tt.courses)
			if tt.wantName == "" {
				if err == nil {
					t.Fatal("expected error for empty pack name, got nil")
				}
				if !strings.Contains(err.Error(), "pack name cannot be empty") {
					t.Errorf("got error %q, want error containing 'pack name cannot be empty'", err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}

			if created.Name != tt.wantName {
				t.Errorf("pack name = %q, want %q", created.Name, tt.wantName)
			}

			db.AssertPackExists(created.ID)
			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    tt.wantName,
				Courses: tt.courses,
			})
		})
	}
}
