package tests

import (
	"context"
	"testing"

	"git.sr.ht/~alias/polybase-go/internal"
)

// create single course
// create duplicate course
// create with invalid course data
// create course with max values
// create with minimum valid values

func TestCreateCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB)

	course := internal.Course{
		Code:     "CS101",
		Kind:     "Notes",
		Part:     1,
		Parts:    1,
		Name:     "Introduction to CS",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.Create(context.Background(), course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	db.AssertCount(1)
	db.AssertExists(internal.CourseID{Code: "CS101", Kind: "Notes", Part: 1})
	db.AssertCourseEqual(internal.CourseID{Code: "CS101", Kind: "Notes", Part: 1}, course)
}
