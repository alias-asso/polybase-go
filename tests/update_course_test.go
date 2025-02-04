package tests

import (
	"context"
	"strings"
	"testing"

	"git.sr.ht/~alias/polybase-go/libpolybase"
)

// All fields of a course can be updated simultaneously
func TestUpdateAllFields(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// First verify we can create a course
	original := libpolybase.Course{
		Code:     "LU3IN005",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Operating Systems",
		Quantity: 30,
		Total:    50,
		Shown:    true,
		Semester: "S1",
	}

	t.Log("Creating initial course...")
	created, err := pb.CreateCourse(ctx, "testuser", original)
	if err != nil {
		t.Fatalf("failed to create initial course: %v", err)
	}
	t.Logf("Course created successfully: %+v", created)

	// Verify course exists before update
	t.Log("Verifying course exists...")
	_, err = pb.GetCourse(ctx, libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	if err != nil {
		t.Fatalf("failed to get course before update: %v", err)
	}
	t.Log("Course exists and can be retrieved")

	// Prepare update
	newCode := "LU3IN006"
	newKind := "TME"
	newPart := 2
	newName := "Advanced Operating Systems"
	newQuantity := 40
	newTotal := 60
	newShown := false
	newSemester := "S2"

	partial := libpolybase.PartialCourse{
		Code:     &newCode,
		Kind:     &newKind,
		Part:     &newPart,
		Name:     &newName,
		Quantity: &newQuantity,
		Total:    &newTotal,
		Shown:    &newShown,
		Semester: &newSemester,
	}
	// Note: We removed Parts from partial update since it's libpolybasely managed

	t.Log("Attempting to update course...")
	t.Logf("Update values: %+v", partial)

	// Perform update
	updated, err := pb.UpdateCourse(ctx, "testuser", libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}, partial)
	if err != nil {
		t.Fatalf("failed to update course: %v", err)
	}
	t.Logf("Update successful, received: %+v", updated)

	expected := libpolybase.Course{
		Code:     newCode,
		Kind:     newKind,
		Part:     newPart,
		Parts:    newPart, // Parts should match the highest part number (2)
		Name:     newName,
		Quantity: newQuantity,
		Total:    newTotal,
		Shown:    newShown,
		Semester: newSemester,
	}

	if updated != expected {
		t.Errorf("updated course mismatch:\ngot: %+v\nwant: %+v", updated, expected)
	}

	t.Log("Verifying database state...")
	// Verify final state
	db.AssertNotExists(libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	db.AssertExists(libpolybase.CourseID{
		Code: newCode,
		Kind: newKind,
		Part: newPart,
	})
	db.AssertCourseEqual(libpolybase.CourseID{
		Code: newCode,
		Kind: newKind,
		Part: newPart,
	}, expected)
	t.Log("Test completed successfully")
}

// Each individual field can be updated independently
func TestUpdateSingleField(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	original := libpolybase.Course{
		Code:     "LU3IN005",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Operating Systems",
		Quantity: 30,
		Total:    50,
		Shown:    true,
		Semester: "S1",
	}

	tests := []struct {
		name    string
		partial libpolybase.PartialCourse
		want    libpolybase.Course
	}{
		{
			name: "update code",
			partial: libpolybase.PartialCourse{
				Code: stringPtr("LU3IN006"),
			},
			want: libpolybase.Course{
				Code:     "LU3IN006",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 30,
				Total:    50,
				Shown:    true,
				Semester: "S1",
			},
		},
		{
			name: "update kind",
			partial: libpolybase.PartialCourse{
				Kind: stringPtr("TME"),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TME",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 30,
				Total:    50,
				Shown:    true,
				Semester: "S1",
			},
		},
		{
			name: "update name",
			partial: libpolybase.PartialCourse{
				Name: stringPtr("Advanced OS"),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Advanced OS",
				Quantity: 30,
				Total:    50,
				Shown:    true,
				Semester: "S1",
			},
		},
		{
			name: "update quantity",
			partial: libpolybase.PartialCourse{
				Quantity: intPtr(40),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 40,
				Total:    50,
				Shown:    true,
				Semester: "S1",
			},
		},
		{
			name: "update total",
			partial: libpolybase.PartialCourse{
				Total: intPtr(60),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 30,
				Total:    60,
				Shown:    true,
				Semester: "S1",
			},
		},
		{
			name: "update shown",
			partial: libpolybase.PartialCourse{
				Shown: boolPtr(false),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 30,
				Total:    50,
				Shown:    false,
				Semester: "S1",
			},
		},
		{
			name: "update semester",
			partial: libpolybase.PartialCourse{
				Semester: stringPtr("S2"),
			},
			want: libpolybase.Course{
				Code:     "LU3IN005",
				Kind:     "TD",
				Part:     1,
				Parts:    1,
				Name:     "Operating Systems",
				Quantity: 30,
				Total:    50,
				Shown:    true,
				Semester: "S2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database and recreate original course for each test
			db.Clear()
			created, err := pb.CreateCourse(ctx, "testuser", original)
			if err != nil {
				t.Fatalf("failed to create initial course: %v", err)
			}

			// Perform single field update
			updated, err := pb.UpdateCourse(ctx, "testuser", libpolybase.CourseID{
				Code: created.Code,
				Kind: created.Kind,
				Part: created.Part,
			}, tt.partial)

			if err != nil {
				t.Fatalf("failed to update course: %v", err)
			}

			if updated != tt.want {
				t.Errorf("updated course mismatch:\ngot: %+v\nwant: %+v", updated, tt.want)
			}

			db.AssertCourseEqual(libpolybase.CourseID{
				Code: updated.Code,
				Kind: updated.Kind,
				Part: updated.Part,
			}, tt.want)
		})
	}
}

// Updating non-existent course returns appropriate error
func TestUpdateNonExistentCourse(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	newName := "Updated Name"
	partial := libpolybase.PartialCourse{
		Name: &newName,
	}

	id := libpolybase.CourseID{
		Code: "NOTFOUND",
		Kind: "MISSING",
		Part: 1,
	}

	_, err := pb.UpdateCourse(ctx, "testuser", id, partial)
	if err == nil {
		t.Fatal("expected error when updating non-existent course, got nil")
	}
	if !strings.Contains(err.Error(), "get current course: course not found") {
		t.Errorf("expected 'course does not exists' error, got: %v", err)
	}

	db.AssertCount(0)
	db.AssertNotExists(id)
}

// Updates with invalid values are properly rejected
func TestUpdateWithInvalidValues(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	original := libpolybase.Course{
		Code:     "LU3IN005",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Operating Systems",
		Quantity: 30,
		Total:    50,
		Shown:    true,
		Semester: "S1",
	}

	tests := []struct {
		name    string
		partial libpolybase.PartialCourse
	}{
		{
			name: "invalid code format",
			partial: libpolybase.PartialCourse{
				Code: stringPtr(""),
			},
		},
		{
			name: "invalid kind with numbers",
			partial: libpolybase.PartialCourse{
				Kind: stringPtr(""),
			},
		},
		{
			name: "invalid kind with numbers",
			partial: libpolybase.PartialCourse{
				Part: intPtr(0),
			},
		},
		{
			name: "quantity exceeds total",
			partial: libpolybase.PartialCourse{
				Quantity: intPtr(100),
			},
		},
		{
			name: "negative quantity",
			partial: libpolybase.PartialCourse{
				Quantity: intPtr(-10),
			},
		},
		{
			name: "negative total",
			partial: libpolybase.PartialCourse{
				Total: intPtr(-50),
			},
		},
		{
			name: "invalid semester format",
			partial: libpolybase.PartialCourse{
				Semester: stringPtr("X1"),
			},
		},
		{
			name: "invalid semester number",
			partial: libpolybase.PartialCourse{
				Semester: stringPtr("S3"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database and recreate original course for each test
			db.Clear()
			_, err := pb.CreateCourse(ctx, "testuser", original)
			if err != nil {
				t.Fatalf("failed to create initial course: %v", err)
			}

			id := libpolybase.CourseID{
				Code: original.Code,
				Kind: original.Kind,
				Part: original.Part,
			}

			_, err = pb.UpdateCourse(ctx, "testuser", id, tt.partial)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Verify course remains unchanged
			db.AssertCourseEqual(id, original)
		})
	}
}

// Updating to create a duplicate is properly rejected
func TestUpdateToDuplicate(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create first course
	first := libpolybase.Course{
		Code:     "LU3IN005",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Operating Systems",
		Quantity: 30,
		Total:    50,
		Shown:    true,
		Semester: "S1",
	}

	// Create second course
	second := libpolybase.Course{
		Code:     "LU3IN006",
		Kind:     "TD",
		Part:     1,
		Parts:    1,
		Name:     "Algorithms",
		Quantity: 40,
		Total:    60,
		Shown:    true,
		Semester: "S1",
	}

	// Insert both courses
	_, err := pb.CreateCourse(ctx, "testuser", first)
	if err != nil {
		t.Fatalf("failed to create first course: %v", err)
	}

	_, err = pb.CreateCourse(ctx, "testuser", second)
	if err != nil {
		t.Fatalf("failed to create second course: %v", err)
	}

	// Try to update second course to have same ID as first
	partial := libpolybase.PartialCourse{
		Code: &first.Code,
		Kind: &first.Kind,
		Part: &first.Part,
	}

	_, err = pb.UpdateCourse(ctx, "testuser", libpolybase.CourseID{
		Code: second.Code,
		Kind: second.Kind,
		Part: second.Part,
	}, partial)

	if err == nil {
		t.Fatal("expected error when updating to create duplicate, got nil")
	}

	// Verify both courses remain unchanged
	db.AssertCourseEqual(libpolybase.CourseID{
		Code: first.Code,
		Kind: first.Kind,
		Part: first.Part,
	}, first)

	db.AssertCourseEqual(libpolybase.CourseID{
		Code: second.Code,
		Kind: second.Kind,
		Part: second.Part,
	}, second)
}

// Update with no actual changes is handled gracefully
func TestUpdateCourseWithSameInfo(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create initial course
	original := libpolybase.Course{
		Code:     "LU3IN009",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Systèmes de Gestion de Bases de Données",
		Quantity: 8,
		Total:    60,
		Shown:    true,
		Semester: "S1",
	}

	t.Log("Creating initial course...")
	created, err := pb.CreateCourse(ctx, "testuser", original)
	if err != nil {
		t.Fatalf("failed to create initial course: %v", err)
	}
	t.Logf("Course created successfully: %+v", created)

	// Verify initial state
	t.Log("Verifying course exists...")
	fetchedBefore, err := pb.GetCourse(ctx, libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	if err != nil {
		t.Fatalf("failed to get course before update: %v", err)
	}
	if fetchedBefore != original {
		t.Errorf("initial course mismatch:\ngot: %+v\nwant: %+v", fetchedBefore, original)
	}
	t.Log("Course exists and matches expected state")

	// Prepare update with same values
	sameCode := original.Code
	sameKind := original.Kind
	samePart := original.Part
	sameName := original.Name
	sameQuantity := original.Quantity
	sameTotal := original.Total
	sameShown := original.Shown
	sameSemester := original.Semester

	partial := libpolybase.PartialCourse{
		Code:     &sameCode,
		Kind:     &sameKind,
		Part:     &samePart,
		Name:     &sameName,
		Quantity: &sameQuantity,
		Total:    &sameTotal,
		Shown:    &sameShown,
		Semester: &sameSemester,
	}

	t.Log("Attempting update with same values...")
	t.Logf("Update values: %+v", partial)

	// Perform update
	updated, err := pb.UpdateCourse(ctx, "testuser", libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}, partial)
	if err != nil {
		t.Fatalf("failed to update course with same values: %v", err)
	}
	t.Logf("Update successful, received: %+v", updated)

	// Everything should match the original
	if updated != original {
		t.Errorf("updated course mismatch:\ngot: %+v\nwant: %+v", updated, original)
	}

	t.Log("Verifying final database state...")
	// Verify the course still exists with same values
	db.AssertExists(libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	db.AssertCourseEqual(libpolybase.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}, original)
	t.Log("Test completed successfully")
}
