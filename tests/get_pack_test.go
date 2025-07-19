package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// An existing pack can be retrieved accurately
func TestGetExistingPack(t *testing.T) {
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
			packName: "Basic Pack",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
		{
			name:     "multiple course pack",
			packName: "Complete Pack",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS103", Kind: "TD", Part: 1},
			},
		},
		{
			name:     "pack with spaces in name",
			packName: "Programming   Course   Pack",
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

			got, err := pb.GetPack(ctx, created.ID)
			if err != nil {
				t.Fatalf("failed to get pack: %v", err)
			}

			want := libpolybase.Pack{
				ID:      created.ID,
				Name:    tt.packName,
				Courses: tt.courses,
			}

			if got.ID != want.ID {
				t.Errorf("pack ID = %d, want %d", got.ID, want.ID)
			}

			if got.Name != want.Name {
				t.Errorf("pack name = %q, want %q", got.Name, want.Name)
			}

			if len(got.Courses) != len(want.Courses) {
				t.Fatalf("got %d courses, want %d", len(got.Courses), len(want.Courses))
			}

			for i, course := range got.Courses {
				if course != want.Courses[i] {
					t.Errorf("course[%d] = %+v, want %+v", i, course, want.Courses[i])
				}
			}

			db.AssertPackEqual(created.ID, want)

			for _, courseID := range tt.courses {
				db.AssertCourseInPack(created.ID, courseID)
			}
		})
	}
}

// Retrieving a non-existent pack returns appropriate error
func TestGetNonExistentPack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	tests := []struct {
		name    string
		packID  int
		setup   func(t *testing.T, pb *libpolybase.PB)
		cleanup func(t *testing.T, pb *libpolybase.PB, id int)
	}{
		{
			name:   "never existed pack",
			packID: 999,
		},
		{
			name:   "zero id",
			packID: 0,
		},
		{
			name:   "negative id",
			packID: -1,
		},
		{
			name:   "deleted pack",
			packID: 1,
			setup: func(t *testing.T, pb *libpolybase.PB) {
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
					t.Fatalf("failed to create course: %v", err)
				}

				courseID := libpolybase.CourseID{
					Code: course.Code,
					Kind: course.Kind,
					Part: course.Part,
				}

				_, err = pb.CreatePack(ctx, "testuser", "Test Pack", []libpolybase.CourseID{courseID})
				if err != nil {
					t.Fatalf("failed to create pack: %v", err)
				}
			},
			cleanup: func(t *testing.T, pb *libpolybase.PB, id int) {
				err := pb.DeletePack(ctx, "testuser", id)
				if err != nil {
					t.Fatalf("failed to delete pack: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, pb)
			}

			if tt.cleanup != nil {
				tt.cleanup(t, pb, tt.packID)
			}

			_, err := pb.GetPack(ctx, tt.packID)
			if err == nil {
				t.Fatal("expected error when getting non-existent pack, got nil")
			}

			if !strings.Contains(err.Error(), "pack not found") {
				t.Errorf("got error %q, want error containing 'pack not found'", err.Error())
			}

			db.AssertPackNotExists(tt.packID)

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM pack_courses WHERE pack_id = ?", tt.packID).Scan(&count)
			if err != nil {
				t.Fatalf("failed to count pack courses: %v", err)
			}

			if count != 0 {
				t.Errorf("pack_courses count = %d, want 0", count)
			}
		})
	}
}

// Pack courses are returned in correct order
func TestPackCourseOrder(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Cours",
			Part:     2,
			Parts:    2,
			Name:     "Programming II",
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS101",
			Kind:     "Cours",
			Part:     1,
			Parts:    2,
			Name:     "Programming I",
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS101",
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
			Code:     "CS102",
			Kind:     "TD",
			Part:     1,
			Parts:    1,
			Name:     "Advanced Programming",
			Quantity: 20,
			Total:    40,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS100",
			Kind:     "Memento",
			Part:     1,
			Parts:    1,
			Name:     "Programming Notes",
			Quantity: 25,
			Total:    50,
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
		name       string
		inputOrder []libpolybase.CourseID
		wantOrder  []libpolybase.CourseID
	}{
		{
			name: "already sorted order",
			inputOrder: []libpolybase.CourseID{
				{Code: "CS100", Kind: "Memento", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS102", Kind: "TD", Part: 1},
			},
			wantOrder: []libpolybase.CourseID{
				{Code: "CS100", Kind: "Memento", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS102", Kind: "TD", Part: 1},
			},
		},
		{
			name: "reversed input order",
			inputOrder: []libpolybase.CourseID{
				{Code: "CS102", Kind: "TD", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS100", Kind: "Memento", Part: 1},
			},
			wantOrder: []libpolybase.CourseID{
				{Code: "CS100", Kind: "Memento", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS102", Kind: "TD", Part: 1},
			},
		},
		{
			name: "mixed input order",
			inputOrder: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS100", Kind: "Memento", Part: 1},
				{Code: "CS102", Kind: "TD", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
			wantOrder: []libpolybase.CourseID{
				{Code: "CS100", Kind: "Memento", Part: 1},
				{Code: "CS101", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 2},
				{Code: "CS102", Kind: "TD", Part: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := pb.CreatePack(ctx, "testuser", "Test Pack", tt.inputOrder)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}

			got, err := pb.GetPack(ctx, created.ID)
			if err != nil {
				t.Fatalf("failed to get pack: %v", err)
			}

			if len(got.Courses) != len(tt.wantOrder) {
				t.Fatalf("got %d courses, want %d", len(got.Courses), len(tt.wantOrder))
			}

			for i, course := range got.Courses {
				if course != tt.wantOrder[i] {
					t.Errorf("course[%d] = %+v, want %+v", i, course, tt.wantOrder[i])
				}
			}

			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    "Test Pack",
				Courses: tt.wantOrder,
			})
		})
	}
}
