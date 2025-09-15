package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// Pack can be deleted successfully
func TestDeletePackSuccessfully(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test courses
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
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Create pack
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
	}

	pack, err := pb.CreatePack(ctx, "testuser", "Test Pack", courseIDs)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// Verify pack exists
	db.AssertPackExists(pack.ID)
	if count := db.CountPackCourses(pack.ID); count != 2 {
		t.Errorf("pack course count = %d, want 2", count)
	}

	// Delete pack
	err = pb.DeletePack(ctx, "testuser", pack.ID)
	if err != nil {
		t.Fatalf("failed to delete pack: %v", err)
	}

	// Verify pack no longer exists
	db.AssertPackNotExists(pack.ID)

	// Verify pack courses removed
	if count := db.CountPackCourses(pack.ID); count != 0 {
		t.Errorf("pack course count after deletion = %d, want 0", count)
	}

	// Verify courses still exist
	for _, courseID := range courseIDs {
		_, err := pb.GetCourse(ctx, courseID)
		if err != nil {
			t.Errorf("course %s should still exist after pack deletion", courseID.ID())
		}
	}
}

// Deleting a non-existent pack returns appropriate error
func TestDeleteNonExistentPack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	tests := []struct {
		name   string
		packID int
	}{
		{
			name:   "non-existent positive ID",
			packID: 999,
		},
		{
			name:   "zero ID",
			packID: 0,
		},
		{
			name:   "negative ID",
			packID: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pb.DeletePack(ctx, "testuser", tt.packID)
			if err == nil {
				t.Error("expected error when deleting non-existent pack, got nil")
			}

			if !strings.Contains(err.Error(), "pack not found") {
				t.Errorf("got error %q, want error containing 'pack not found'", err.Error())
			}

			db.AssertPackNotExists(tt.packID)
		})
	}
}

// Deleting a pack removes associated course links
func TestDeletePackRemovesCourseLinks(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course
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

	// Create pack
	pack, err := pb.CreatePack(ctx, "testuser", "Test Pack", []libpolybase.CourseID{courseID})
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// Verify associations exist
	db.AssertCourseInPack(pack.ID, courseID)

	// Delete pack
	err = pb.DeletePack(ctx, "testuser", pack.ID)
	if err != nil {
		t.Fatalf("failed to delete pack: %v", err)
	}

	// Verify associations removed
	db.AssertCourseNotInPack(pack.ID, courseID)

	// Verify course still exists
	_, err = pb.GetCourse(ctx, courseID)
	if err != nil {
		t.Errorf("course should still exist after pack deletion: %v", err)
	}
}

// Deleted pack can be recreated with same name
func TestRecreateDeletedPack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course
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

	packName := "Test Pack"

	// Create pack
	pack1, err := pb.CreatePack(ctx, "testuser", packName, []libpolybase.CourseID{courseID})
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	firstID := pack1.ID

	// Delete pack
	err = pb.DeletePack(ctx, "testuser", pack1.ID)
	if err != nil {
		t.Fatalf("failed to delete pack: %v", err)
	}

	// Recreate pack with same name
	pack2, err := pb.CreatePack(ctx, "testuser", packName, []libpolybase.CourseID{courseID})
	if err != nil {
		t.Fatalf("failed to recreate pack: %v", err)
	}

	// Verify new pack has different ID
	if pack2.ID == firstID {
		t.Errorf("recreated pack has same ID %d as deleted pack", firstID)
	}

	// Verify pack exists and has correct data
	if pack2.Name != packName {
		t.Errorf("recreated pack name = %q, want %q", pack2.Name, packName)
	}

	if len(pack2.Courses) != 1 || pack2.Courses[0] != courseID {
		t.Errorf("recreated pack courses = %v, want %v", pack2.Courses, []libpolybase.CourseID{courseID})
	}

	db.AssertPackExists(pack2.ID)
	db.AssertCourseInPack(pack2.ID, courseID)
}

// Multiple packs can be deleted in sequence
func TestDeleteMultiplePacksSequentially(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test courses
	courses := []libpolybase.Course{
		{
			Code: "CS101", Kind: "Cours", Part: 1, Parts: 1,
			Name: "Programming I", Quantity: 50, Total: 100,
			Shown: true, Semester: "S1",
		},
		{
			Code: "CS102", Kind: "TME", Part: 1, Parts: 1,
			Name: "Programming Lab", Quantity: 30, Total: 60,
			Shown: true, Semester: "S1",
		},
		{
			Code: "CS103", Kind: "TD", Part: 1, Parts: 1,
			Name: "Programming Tutorial", Quantity: 20, Total: 40,
			Shown: true, Semester: "S1",
		},
	}

	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course: %v", err)
		}
	}

	// Create multiple packs
	packs := []struct {
		name    string
		courses []libpolybase.CourseID
	}{
		{
			name: "Pack 1",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
		{
			name: "Pack 2",
			courses: []libpolybase.CourseID{
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS103", Kind: "TD", Part: 1},
			},
		},
		{
			name: "Pack 3",
			courses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS103", Kind: "TD", Part: 1},
			},
		},
	}

	var createdPacks []libpolybase.Pack
	for _, packDef := range packs {
		pack, err := pb.CreatePack(ctx, "testuser", packDef.name, packDef.courses)
		if err != nil {
			t.Fatalf("failed to create pack %s: %v", packDef.name, err)
		}
		createdPacks = append(createdPacks, pack)
	}

	// Delete all packs
	for i, pack := range createdPacks {
		err := pb.DeletePack(ctx, "testuser", pack.ID)
		if err != nil {
			t.Fatalf("failed to delete pack %d: %v", pack.ID, err)
		}

		// Verify pack is deleted
		db.AssertPackNotExists(pack.ID)

		// Verify remaining packs still exist
		for j := i + 1; j < len(createdPacks); j++ {
			db.AssertPackExists(createdPacks[j].ID)
		}
	}

	// Verify all pack_courses links are cleaned up
	var totalLinks int
	err := db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&totalLinks)
	if err != nil {
		t.Fatalf("failed to count pack_courses: %v", err)
	}
	if totalLinks != 0 {
		t.Errorf("pack_courses count = %d, want 0", totalLinks)
	}

	// Verify courses still exist
	for _, course := range courses {
		courseID := libpolybase.CourseID{
			Code: course.Code,
			Kind: course.Kind,
			Part: course.Part,
		}
		_, err := pb.GetCourse(ctx, courseID)
		if err != nil {
			t.Errorf("course %s should still exist: %v", courseID.ID(), err)
		}
	}
}

// Deletion of pack with many courses works correctly
func TestDeletePackWithManyCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create 20 test courses
	var courseIDs []libpolybase.CourseID
	for i := 1; i <= 20; i++ {
		course := libpolybase.Course{
			Code: fmt.Sprintf("CS%03d", 100+i), Kind: "Cours", Part: 1, Parts: 1,
			Name: fmt.Sprintf("Course %d", i), Quantity: 10, Total: 50,
			Shown: true, Semester: "S1",
		}

		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %d: %v", i, err)
		}

		courseIDs = append(courseIDs, libpolybase.CourseID{
			Code: course.Code, Kind: course.Kind, Part: course.Part,
		})
	}

	// Create pack with all courses
	pack, err := pb.CreatePack(ctx, "testuser", "Large Pack", courseIDs)
	if err != nil {
		t.Fatalf("failed to create large pack: %v", err)
	}

	// Verify all course links exist
	if count := db.CountPackCourses(pack.ID); count != 20 {
		t.Errorf("pack course count = %d, want 20", count)
	}

	// Delete pack
	err = pb.DeletePack(ctx, "testuser", pack.ID)
	if err != nil {
		t.Fatalf("failed to delete large pack: %v", err)
	}

	// Verify pack deleted
	db.AssertPackNotExists(pack.ID)

	// Verify all course links removed
	if count := db.CountPackCourses(pack.ID); count != 0 {
		t.Errorf("pack course count after deletion = %d, want 0", count)
	}

	// Verify all courses still exist
	for _, courseID := range courseIDs {
		_, err := pb.GetCourse(ctx, courseID)
		if err != nil {
			t.Errorf("course %s should still exist: %v", courseID.ID(), err)
		}
	}
}

// Already deleted pack cannot be deleted again
func TestDeleteAlreadyDeletedPack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course and pack
	course := libpolybase.Course{
		Code: "CS101", Kind: "Cours", Part: 1, Parts: 1,
		Name: "Programming", Quantity: 50, Total: 100,
		Shown: true, Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	courseID := libpolybase.CourseID{
		Code: course.Code, Kind: course.Kind, Part: course.Part,
	}

	pack, err := pb.CreatePack(ctx, "testuser", "Test Pack", []libpolybase.CourseID{courseID})
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// First deletion
	err = pb.DeletePack(ctx, "testuser", pack.ID)
	if err != nil {
		t.Fatalf("failed to delete pack: %v", err)
	}

	db.AssertPackNotExists(pack.ID)

	// Second deletion attempt
	err = pb.DeletePack(ctx, "testuser", pack.ID)
	if err == nil {
		t.Error("expected error when deleting already deleted pack, got nil")
	}

	if !strings.Contains(err.Error(), "pack not found") {
		t.Errorf("got error %q, want error containing 'pack not found'", err.Error())
	}
}

// Empty database deletion attempts fail properly
func TestDeleteFromEmptyDatabase(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Attempt to delete from empty database
	err := pb.DeletePack(ctx, "testuser", 1)
	if err == nil {
		t.Error("expected error when deleting from empty database, got nil")
	}

	if !strings.Contains(err.Error(), "pack not found") {
		t.Errorf("got error %q, want error containing 'pack not found'", err.Error())
	}

	// Verify database remains empty
	var packCount int
	err = db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&packCount)
	if err != nil {
		t.Fatalf("failed to count packs: %v", err)
	}
	if packCount != 0 {
		t.Errorf("pack count = %d, want 0", packCount)
	}
}

// Stress test: Create and delete many packs rapidly
func TestDeletePackStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course
	course := libpolybase.Course{
		Code: "CS101", Kind: "Cours", Part: 1, Parts: 1,
		Name: "Programming", Quantity: 50, Total: 100,
		Shown: true, Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	courseID := libpolybase.CourseID{
		Code: course.Code, Kind: course.Kind, Part: course.Part,
	}

	// Create and delete 100 packs
	for i := 0; i < 100; i++ {
		// Create pack
		pack, err := pb.CreatePack(ctx, "testuser",
			fmt.Sprintf("Pack %d", i), []libpolybase.CourseID{courseID})
		if err != nil {
			t.Fatalf("failed to create pack %d: %v", i, err)
		}

		// Immediately delete pack
		err = pb.DeletePack(ctx, "testuser", pack.ID)
		if err != nil {
			t.Fatalf("failed to delete pack %d: %v", pack.ID, err)
		}

		db.AssertPackNotExists(pack.ID)
	}

	// Verify database is clean
	var packCount, linkCount int
	err = db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&packCount)
	if err != nil {
		t.Fatalf("failed to count packs: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&linkCount)
	if err != nil {
		t.Fatalf("failed to count pack_courses: %v", err)
	}

	if packCount != 0 {
		t.Errorf("final pack count = %d, want 0", packCount)
	}
	if linkCount != 0 {
		t.Errorf("final pack_courses count = %d, want 0", linkCount)
	}

	// Verify course still exists
	_, err = pb.GetCourse(ctx, courseID)
	if err != nil {
		t.Errorf("course should still exist after stress test: %v", err)
	}
}
