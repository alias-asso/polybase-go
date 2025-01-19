package tests

import (
	"context"
	"testing"

	"git.sr.ht/~alias/polybase-go/internal"
)

// An existing course can be retrieved accurately
func TestGetExistingCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	course := internal.Course{
		Code:     "CS101",
		Kind:     "Lecture",
		Part:     1,
		Parts:    2,
		Name:     "Introduction to Programming",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	db.Insert(course)

	id := internal.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}

	got, err := pb.GetCourse(context.Background(), id)
	if err != nil {
		t.Fatalf("failed to get course: %v", err)
	}

	if got != course {
		t.Errorf("got course %+v, want %+v", got, course)
	}
}

// Retrieving a non-existent course returns CourseNotFound error
func TestGetNonexistentCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	id := internal.CourseID{
		Code: "NOTFOUND",
		Kind: "Missing",
		Part: 1,
	}

	_, err := pb.GetCourse(context.Background(), id)
	if _, ok := err.(*internal.CourseNotFound); !ok {
		t.Errorf("got error %T, want *CourseNotFound", err)
	}

	db.AssertCount(0)
	db.AssertNotExists(id)
}

// Retrieving with invalid course ID returns appropriate error
func TestGetInvalidCourseID(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	cases := []struct {
		name string
		id   internal.CourseID
	}{
		{
			name: "invalid code with lowercase",
			id:   internal.CourseID{Code: "lowercase", Kind: "Lecture", Part: 1},
		},
		{
			name: "invalid code with special chars",
			id:   internal.CourseID{Code: "CS101!", Kind: "Lecture", Part: 1},
		},
		{
			name: "invalid kind with numbers",
			id:   internal.CourseID{Code: "CS101", Kind: "Lecture1", Part: 1},
		},
		{
			name: "invalid kind with special chars",
			id:   internal.CourseID{Code: "CS101", Kind: "Lecture!", Part: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := pb.GetCourse(context.Background(), tc.id)
			if err == nil {
				t.Error("expected error for invalid course ID, got nil")
			}

			db.AssertNotExists(tc.id)
		})
	}
}

// Retrieving a deleted course returns CourseNotFound error
func TestGetAfterDeletion(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

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

	created, err := pb.CreateCourse(context.Background(), "testUser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	id := internal.CourseID{
		Code: created.Code,
		Kind: created.Kind,
		Part: created.Part,
	}

	err = pb.DeleteCourse(context.Background(), "testUser", id)
	if err != nil {
		t.Fatalf("failed to delete course: %v", err)
	}

	_, err = pb.GetCourse(context.Background(), id)
	if _, ok := err.(*internal.CourseNotFound); !ok {
		t.Errorf("got error %T, want *CourseNotFound", err)
	}

	db.AssertCount(0)
	db.AssertNotExists(id)
}

// Courses with multiple parts can be retrieved correctly
func TestGetMultiPartCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

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
			Quantity: 40,
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
			Quantity: 30,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
	}

	db.InsertMany(courses)

	for _, want := range courses {
		id := internal.CourseID{
			Code: want.Code,
			Kind: want.Kind,
			Part: want.Part,
		}

		got, err := pb.GetCourse(context.Background(), id)
		if err != nil {
			t.Fatalf("failed to get course part %d: %v", want.Part, err)
		}

		if got != want {
			t.Errorf("part %d: got %+v, want %+v", want.Part, got, want)
		}
	}

	db.AssertCount(3)
}
