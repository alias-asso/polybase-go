package tests

import (
	"context"
	"strings"
	"testing"

	"git.sr.ht/~alias/polybase-go/internal"
)

// A new course can be successfully created in an empty database
func TestCreateBasicCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	course := internal.Course{
		Code:     "LU3IN005",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Architecture des ordinateurs",
		Quantity: 30,
		Total:    50,
		Shown:    true,
		Semester: "S1",
	}

	created, err := pb.CreateCourse(context.Background(), "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	if created != course {
		t.Errorf("returned course mismatch:\ngot: %+v\nwant: %+v", created, course)
	}
	db.AssertCount(1)

	courseID := internal.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}

	db.AssertExists(courseID)
	db.AssertCourseEqual(courseID, course)
}

// Multiple distinct courses can be created sequentially
func TestCreateMultipleCourses(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	courses := []internal.Course{
		{
			Code:     "LU3IN009",
			Kind:     "TD",
			Part:     1,
			Parts:    1,
			Name:     "Database Systems",
			Quantity: 60,
			Total:    60,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "LU3IN009",
			Kind:     "TME",
			Part:     1,
			Parts:    1,
			Name:     "Database Systems Lab",
			Quantity: 30,
			Total:    30,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "LU2IN018",
			Kind:     "Cours",
			Part:     1,
			Parts:    1,
			Name:     "Operating Systems",
			Quantity: 100,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		},
	}

	for _, course := range courses {
		created, err := pb.CreateCourse(ctx, "", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}

		db.AssertExists(internal.CourseID{Code: course.Code, Kind: course.Kind, Part: course.Part})
		db.AssertCourseEqual(internal.CourseID{Code: course.Code, Kind: course.Kind, Part: course.Part}, created)
	}

	db.AssertCount(len(courses))
}

// Creating a duplicate course returns an appropriate error
func TestCreateDuplicateCourse(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	original := internal.Course{
		Code:     "LU3IN009",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Database Systems",
		Quantity: 60,
		Total:    60,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "", original)
	if err != nil {
		t.Fatalf("failed to create initial course: %v", err)
	}

	duplicate := internal.Course{
		Code:     "LU3IN009",
		Kind:     "TD",
		Part:     1,
		Parts:    2,
		Name:     "Different Name",
		Quantity: 30,
		Total:    30,
		Shown:    false,
		Semester: "S2",
	}

	_, err = pb.CreateCourse(ctx, "", duplicate)
	if err == nil {
		t.Fatal("expected error when creating duplicate course, got nil")
	}
	if !strings.Contains(err.Error(), "course already exists") {
		t.Errorf("expected 'course already exists' error, got: %v", err)
	}

	db.AssertCount(1)
	db.AssertCourseEqual(internal.CourseID{Code: original.Code, Kind: original.Kind, Part: original.Part}, original)
}

// All course fields are properly validated during creation
func TestCreateCourseFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		course  internal.Course
		wantErr string
	}{
		{
			name: "empty code",
			course: internal.Course{
				Code:     "",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "CODE cannot be empty",
		},
		{
			name: "invalid code characters",
			course: internal.Course{
				Code:     "lu3in009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "invalid course id",
		},
		{
			name: "empty kind",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "KIND cannot be empty",
		},
		{
			name: "invalid kind characters",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD2",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "KIND must be one of: TD, Cours, Memento, TME",
		},
		{
			name: "negative quantity",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: -1,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "quantity cannot be negative",
		},
		{
			name: "negative total",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    -1,
				Semester: "S1",
			},
			wantErr: "total cannot be negative",
		},
		{
			name: "quantity exceeds total",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 20,
				Total:    10,
				Semester: "S1",
			},
			wantErr: "quantity (20) cannot exceed total (10)",
		},
		{
			name: "empty semester",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "",
			},
			wantErr: "SEMESTER must be either S1 or S2",
		},
		{
			name: "invalid semester format",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "X1",
			},
			wantErr: "SEMESTER must be either S1 or S2",
		},
		{
			name: "invalid semester number",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S3",
			},
			wantErr: "SEMESTER must be either S1 or S2",
		},
		{
			name: "valid course with minimum values",
			course: internal.Course{
				Code:     "LU3IN009",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 0,
				Total:    1,
				Semester: "S1",
			},
			wantErr: "",
		},
		{
			name: "valid course with special code characters",
			course: internal.Course{
				Code:     "LU3IN009-{A},B",
				Kind:     "TD",
				Part:     1,
				Name:     "Test Course",
				Quantity: 10,
				Total:    10,
				Semester: "S2",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewDB(t)
			pb := internal.New(db.DB, "", false)

			_, err := pb.CreateCourse(context.Background(), "", tt.course)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("[%s] expected success, got error: %v", tt.name, err)
				}
				db.AssertCount(1)
				db.AssertExists(internal.CourseID{
					Code: tt.course.Code,
					Kind: tt.course.Kind,
					Part: tt.course.Part,
				})
			} else {
				if err == nil {
					t.Errorf("[%s] expected error containing %q, got nil", tt.name, tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("[%s] expected error containing %q, got: %v", tt.name, tt.wantErr, err)
				}
				db.AssertCount(0)
			}
		})
	}
}

// A course can be created with maximum valid values for all fields
func TestCreateWithMaxValues(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	course := internal.Course{
		Code:     "COURSE-123{A,B,C}-456",
		Kind:     "Cours",
		Part:     999,
		Parts:    999,
		Name:     "Advanced Topics in Theoretical Computer Science and Distributed Systems Engineering with Applications in Machine Learning",
		Quantity: 9999,
		Total:    10000,
		Shown:    true,
		Semester: "S1",
	}

	created, err := pb.CreateCourse(context.Background(), "testUser", course)
	if err != nil {
		t.Fatalf("failed to create course with maximum values: %v", err)
	}

	db.AssertCount(1)

	id := internal.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}
	db.AssertExists(id)

	db.AssertCourseEqual(id, course)

	if created != course {
		t.Errorf("created course does not match input\ngot: %+v\nwant: %+v", created, course)
	}
}

// A course can be created with minimum valid values for all fields
func TestCreateWithMinValues(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)

	course := internal.Course{
		Code:     "A",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "x",
		Quantity: 0,
		Total:    1,
		Shown:    true,
		Semester: "S1",
	}

	created, err := pb.CreateCourse(context.Background(), "testUser", course)
	if err != nil {
		t.Fatalf("failed to create course with minimum values: %v", err)
	}

	db.AssertCount(1)

	id := internal.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}
	db.AssertExists(id)
	db.AssertCourseEqual(id, course)

	if created != course {
		t.Errorf("created course does not match input\ngot: %+v\nwant: %+v", created, course)
	}
}
