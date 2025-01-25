package tests

import (
	"context"
	"fmt"
	"testing"

	"git.sr.ht/~alias/polybase-go/internal"
)

// delete existing course
func TestDeleteExistingCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course
	course := internal.Course{
		Code:     "CS101",
		Kind:     "Lecture",
		Part:     1,
		Parts:    1,
		Name:     "Programming I",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	created, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create test course: %v", err)
	}

	// Create a pack containing the course
	courseID := internal.CourseID{
		Code: created.Code,
		Kind: created.Kind,
		Part: created.Part,
	}

	pack, err := pb.CreatePack(ctx, "testuser", "Test Pack", []internal.CourseID{courseID})
	if err != nil {
		t.Fatalf("failed to create test pack: %v", err)
	}

	// Delete the course
	err = pb.DeleteCourse(ctx, "testuser", courseID)
	if err != nil {
		t.Fatalf("failed to delete course: %v", err)
	}

	// Verify course is deleted
	db.AssertNotExists(courseID)

	// Verify course is removed from pack
	db.AssertCourseNotInPack(pack.ID, courseID)

	// Verify get returns CourseNotFound
	_, err = pb.GetCourse(ctx, courseID)
	if _, ok := err.(*internal.CourseNotFound); !ok {
		t.Errorf("got error %T, want *CourseNotFound", err)
	}

	// Verify pack still exists but is empty
	packAfter, err := pb.GetPack(ctx, pack.ID)
	if err != nil {
		t.Fatalf("failed to get pack after course deletion: %v", err)
	}

	if len(packAfter.Courses) != 0 {
		t.Errorf("pack contains %d courses after deletion, want 0", len(packAfter.Courses))
	}

	// Verify total course count
	db.AssertCount(0)
}

// delete non-existent course
func TestDeleteNonExistentCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	tests := []struct {
		name  string
		id    internal.CourseID
		setup func(t *testing.T, pb *internal.PB)
	}{
		{
			name: "never existed",
			id: internal.CourseID{
				Code: "NOTFOUND",
				Kind: "MISSING",
				Part: 1,
			},
		},
		{
			name: "already deleted",
			id: internal.CourseID{
				Code: "CS101",
				Kind: "Lecture",
				Part: 1,
			},
			setup: func(t *testing.T, pb *internal.PB) {
				course := internal.Course{
					Code:     "CS101",
					Kind:     "Lecture",
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

				err = pb.DeleteCourse(ctx, "testuser", internal.CourseID{
					Code: course.Code,
					Kind: course.Kind,
					Part: course.Part,
				})
				if err != nil {
					t.Fatalf("failed to delete test course: %v", err)
				}
			},
		},
		{
			name: "invalid course ID format",
			id: internal.CourseID{
				Code: "invalid!code",
				Kind: "invalid!kind",
				Part: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Clear()

			if tt.setup != nil {
				tt.setup(t, pb)
			}

			err := pb.DeleteCourse(ctx, "testuser", tt.id)
			if err == nil {
				t.Fatal("expected error when deleting non-existent course, got nil")
			}

			db.AssertNotExists(tt.id)
		})
	}
}

// delete and recreate
func TestDeleteAndRecreateCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	original := internal.Course{
		Code:     "CS101",
		Kind:     "Lecture",
		Part:     1,
		Parts:    1,
		Name:     "Programming",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	// Create initial course
	_, err := pb.CreateCourse(ctx, "testuser", original)
	if err != nil {
		t.Fatalf("failed to create initial course: %v", err)
	}

	id := internal.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}

	// Verify initial state
	db.AssertCount(1)
	db.AssertExists(id)
	db.AssertCourseEqual(id, original)

	// Delete the course
	err = pb.DeleteCourse(ctx, "testuser", id)
	if err != nil {
		t.Fatalf("failed to delete course: %v", err)
	}

	// Verify deletion
	db.AssertCount(0)
	db.AssertNotExists(id)

	// Create new course with same ID but different details
	recreated := internal.Course{
		Code:     original.Code,
		Kind:     original.Kind,
		Part:     original.Part,
		Parts:    1,
		Name:     "Programming Fundamentals", // Different name
		Quantity: 30,                         // Different quantity
		Total:    75,                         // Different total
		Semester: "S2",                       // Different semester
		Shown:    true,
	}

	_, err = pb.CreateCourse(ctx, "testuser", recreated)
	if err != nil {
		t.Fatalf("failed to recreate course: %v", err)
	}

	// Verify recreation
	db.AssertCount(1)
	db.AssertExists(id)
	db.AssertCourseEqual(id, recreated)
}

// delete with invalid id
func TestDeleteCourseWithInvalidID(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	tests := []struct {
		name string
		id   internal.CourseID
	}{
		{
			name: "empty code",
			id: internal.CourseID{
				Code: "",
				Kind: "Lecture",
				Part: 1,
			},
		},
		{
			name: "empty kind",
			id: internal.CourseID{
				Code: "CS101",
				Kind: "",
				Part: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database state
			db.Clear()

			// Attempt to delete with invalid ID
			err := pb.DeleteCourse(ctx, "testuser", tt.id)
			if err == nil {
				t.Fatal("expected error deleting course with invalid ID, got nil")
			}

			// Verify no changes to database
			db.AssertCount(0)
			db.AssertNotExists(tt.id)
		})
	}
}

// delete the last part of a multi-part course
func TestDeleteLastPartOfMultiPartCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	// Create a three-part course
	courses := []internal.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    3,
			Name:     "Programming I",
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     2,
			Parts:    3,
			Name:     "Programming II",
			Quantity: 45,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     3,
			Parts:    3,
			Name:     "Programming III",
			Quantity: 40,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
	}

	// Insert all parts
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course part %d: %v", course.Part, err)
		}
	}

	// Verify initial state
	db.AssertCount(3)

	// Delete parts one by one and verify parts count updates
	deletions := []struct {
		part        int
		remainParts int
	}{
		{part: 3, remainParts: 2},
		{part: 2, remainParts: 1},
		{part: 1, remainParts: 0},
	}

	for _, d := range deletions {
		t.Run(fmt.Sprintf("delete_part_%d", d.part), func(t *testing.T) {
			err := pb.DeleteCourse(ctx, "testuser", internal.CourseID{
				Code: "CS101",
				Kind: "Lecture",
				Part: d.part,
			})
			if err != nil {
				t.Fatalf("failed to delete part %d: %v", d.part, err)
			}

			// Check remaining parts have correct parts count
			remaining, err := pb.ListCourse(ctx, true, nil, nil, nil, nil)
			if err != nil {
				t.Fatalf("failed to list remaining courses: %v", err)
			}

			if len(remaining) != d.remainParts {
				t.Errorf("got %d remaining courses, want %d", len(remaining), d.remainParts)
			}

			for _, course := range remaining {
				if course.Parts != d.remainParts {
					t.Errorf("course part %d has Parts=%d, want %d",
						course.Part, course.Parts, d.remainParts)
				}
			}

			// Verify deleted part doesn't exist
			db.AssertNotExists(internal.CourseID{
				Code: "CS101",
				Kind: "Lecture",
				Part: d.part,
			})
		})
	}

	// Verify final state
	db.AssertCount(0)
}
