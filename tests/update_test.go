package tests

import (
	"context"
	"git.sr.ht/~alias/polybase-go/internal"
	"testing"
)

// All fields of a course can be updated simultaneously
func TestUpdateAllFields(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	// First verify we can create a course
	original := internal.Course{
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
	_, err = pb.GetCourse(ctx, internal.CourseID{
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

	partial := internal.PartialCourse{
		Code:     &newCode,
		Kind:     &newKind,
		Part:     &newPart,
		Name:     &newName,
		Quantity: &newQuantity,
		Total:    &newTotal,
		Shown:    &newShown,
		Semester: &newSemester,
	}
	// Note: We removed Parts from partial update since it's internally managed

	t.Log("Attempting to update course...")
	t.Logf("Update values: %+v", partial)

	// Perform update
	updated, err := pb.UpdateCourse(ctx, "testuser", internal.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}, partial)
	if err != nil {
		t.Fatalf("failed to update course: %v", err)
	}
	t.Logf("Update successful, received: %+v", updated)

	expected := internal.Course{
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
	db.AssertNotExists(internal.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	db.AssertExists(internal.CourseID{
		Code: newCode,
		Kind: newKind,
		Part: newPart,
	})
	db.AssertCourseEqual(internal.CourseID{
		Code: newCode,
		Kind: newKind,
		Part: newPart,
	}, expected)
	t.Log("Test completed successfully")
}

// Each individual field can be updated independently

// Updating non-existent course returns appropriate error

// Updates with invalid values are properly rejected

// Updating to create a duplicate is properly rejected

// Update with no actual changes is handled gracefully
func TestUpdateCourseWithSameInfo(t *testing.T) {
	db := NewDB(t)
	pb := internal.New(db.DB, "", false)
	ctx := context.Background()

	// Create initial course
	original := internal.Course{
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
	fetchedBefore, err := pb.GetCourse(ctx, internal.CourseID{
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

	partial := internal.PartialCourse{
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
	updated, err := pb.UpdateCourse(ctx, "testuser", internal.CourseID{
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
	db.AssertExists(internal.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	})
	db.AssertCourseEqual(internal.CourseID{
		Code: original.Code,
		Kind: original.Kind,
		Part: original.Part,
	}, original)
	t.Log("Test completed successfully")
}
