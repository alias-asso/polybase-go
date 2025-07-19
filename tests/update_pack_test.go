package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// Pack name can be updated while preserving courses
func TestUpdatePackNameOnly(t *testing.T) {
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

	// Insert test courses
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Create initial pack
	initialName := "Programming Pack"
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", initialName, courseIDs)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// Update pack name
	newName := "Updated Programming Pack"
	updated, err := pb.UpdatePack(ctx, "testuser", created.ID, libpolybase.PartialPack{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("failed to update pack name: %v", err)
	}

	// Verify updated state
	if updated.Name != newName {
		t.Errorf("pack name = %q, want %q", updated.Name, newName)
	}

	if len(updated.Courses) != len(courseIDs) {
		t.Errorf("got %d courses, want %d", len(updated.Courses), len(courseIDs))
	}

	db.AssertPackExists(created.ID)
	db.AssertPackEqual(created.ID, libpolybase.Pack{
		ID:      created.ID,
		Name:    newName,
		Courses: courseIDs,
	})

	// Verify all courses are still in the pack
	for _, courseID := range courseIDs {
		db.AssertCourseInPack(created.ID, courseID)
	}

	// Verify course count hasn't changed
	if count := db.CountPackCourses(created.ID); count != len(courseIDs) {
		t.Errorf("pack course count = %d, want %d", count, len(courseIDs))
	}
}

// Pack courses can be updated while preserving name
func TestUpdatePackCoursesOnly(t *testing.T) {
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

	// Insert test courses
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Create initial pack with subset of courses
	packName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", packName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// Update pack courses
	newCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS103", Kind: "TD", Part: 1},
	}

	updated, err := pb.UpdatePack(ctx, "testuser", created.ID, libpolybase.PartialPack{
		Courses: &newCourses,
	})
	if err != nil {
		t.Fatalf("failed to update pack courses: %v", err)
	}

	// Verify updated state
	if updated.Name != packName {
		t.Errorf("pack name = %q, want %q", updated.Name, packName)
	}

	if len(updated.Courses) != len(newCourses) {
		t.Errorf("got %d courses, want %d", len(updated.Courses), len(newCourses))
	}

	db.AssertPackExists(created.ID)
	db.AssertPackEqual(created.ID, libpolybase.Pack{
		ID:      created.ID,
		Name:    packName,
		Courses: newCourses,
	})

	// Verify new course membership
	for _, courseID := range newCourses {
		db.AssertCourseInPack(created.ID, courseID)
	}

	// Verify removed courses are no longer in pack
	db.AssertCourseNotInPack(created.ID, libpolybase.CourseID{
		Code: "CS102",
		Kind: "TME",
		Part: 1,
	})

	// Verify course count matches new list
	if count := db.CountPackCourses(created.ID); count != len(newCourses) {
		t.Errorf("pack course count = %d, want %d", count, len(newCourses))
	}
}

// Pack courses and name can be updated simultaneously
func TestUpdatePackNameAndCourses(t *testing.T) {
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

	// Insert test courses
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Create initial pack
	initialName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", initialName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	// Update both name and courses
	newName := "Advanced Programming Pack"
	newCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS103", Kind: "TD", Part: 1},
	}

	updated, err := pb.UpdatePack(ctx, "testuser", created.ID, libpolybase.PartialPack{
		Name:    &newName,
		Courses: &newCourses,
	})
	if err != nil {
		t.Fatalf("failed to update pack: %v", err)
	}

	// Verify updated state
	if updated.Name != newName {
		t.Errorf("pack name = %q, want %q", updated.Name, newName)
	}

	if len(updated.Courses) != len(newCourses) {
		t.Errorf("got %d courses, want %d", len(updated.Courses), len(newCourses))
	}

	db.AssertPackExists(created.ID)
	db.AssertPackEqual(created.ID, libpolybase.Pack{
		ID:      created.ID,
		Name:    newName,
		Courses: newCourses,
	})

	// Verify new course membership
	for _, courseID := range newCourses {
		db.AssertCourseInPack(created.ID, courseID)
	}

	// Verify removed courses are no longer in pack
	db.AssertCourseNotInPack(created.ID, libpolybase.CourseID{
		Code: "CS102",
		Kind: "TME",
		Part: 1,
	})

	// Verify course count matches new list
	if count := db.CountPackCourses(created.ID); count != len(newCourses) {
		t.Errorf("pack course count = %d, want %d", count, len(newCourses))
	}

	// Verify we can still get the updated pack
	retrieved, err := pb.GetPack(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get updated pack: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("retrieved pack ID = %d, want %d", retrieved.ID, created.ID)
	}

	if retrieved.Name != newName {
		t.Errorf("retrieved pack name = %q, want %q", retrieved.Name, newName)
	}

	if len(retrieved.Courses) != len(newCourses) {
		t.Errorf("retrieved pack has %d courses, want %d", len(retrieved.Courses), len(newCourses))
	}

	// Convert courses to maps for comparison
	wantCourses := make(map[string]struct{})
	for _, c := range newCourses {
		wantCourses[c.ID()] = struct{}{}
	}

	for _, got := range retrieved.Courses {
		if _, exists := wantCourses[got.ID()]; !exists {
			t.Errorf("unexpected course in retrieved pack: %v", got)
		}
		delete(wantCourses, got.ID())
	}

	if len(wantCourses) > 0 {
		t.Errorf("missing courses in retrieved pack: %v", wantCourses)
	}
}

// Updating pack to have no courses is rejected
func TestUpdatePackWithNoCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create initial course
	course := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming I",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	// Create initial pack
	initialName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", initialName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	tests := []struct {
		name   string
		update libpolybase.PartialPack
	}{
		{
			name: "empty course slice",
			update: libpolybase.PartialPack{
				Courses: &[]libpolybase.CourseID{},
			},
		},
		{
			name: "nil course slice",
			update: libpolybase.PartialPack{
				Courses: new([]libpolybase.CourseID),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt update with no courses
			_, err := pb.UpdatePack(ctx, "testuser", created.ID, tt.update)

			if err == nil {
				t.Fatal("expected error when updating pack with no courses, got nil")
			}

			if !strings.Contains(err.Error(), "pack must contain at least one course") {
				t.Errorf("got error %q, want error containing 'pack must contain at least one course'", err.Error())
			}

			// Verify pack remains unchanged
			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    initialName,
				Courses: initialCourses,
			})

			// Verify course relationships remain intact
			for _, courseID := range initialCourses {
				db.AssertCourseInPack(created.ID, courseID)
			}

			if count := db.CountPackCourses(created.ID); count != len(initialCourses) {
				t.Errorf("pack course count = %d, want %d", count, len(initialCourses))
			}
		})
	}
}

// Updating pack with non-existent courses fails properly
func TestUpdatePackWithNonExistentCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create initial course
	course := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming I",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	existingID := libpolybase.CourseID{
		Code: course.Code,
		Kind: course.Kind,
		Part: course.Part,
	}

	// Create initial pack
	initialName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{existingID}

	created, err := pb.CreatePack(ctx, "testuser", initialName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	tests := []struct {
		name          string
		updateCourses []libpolybase.CourseID
	}{
		{
			name: "single non-existent course",
			updateCourses: []libpolybase.CourseID{
				{Code: "FAKE101", Kind: "Missing", Part: 1},
			},
		},
		{
			name: "multiple non-existent courses",
			updateCourses: []libpolybase.CourseID{
				{Code: "FAKE101", Kind: "Missing", Part: 1},
				{Code: "FAKE102", Kind: "Missing", Part: 1},
			},
		},
		{
			name: "mix of existing and non-existent courses",
			updateCourses: []libpolybase.CourseID{
				existingID,
				{Code: "FAKE101", Kind: "Missing", Part: 1},
			},
		},
		{
			name: "existing code with wrong kind",
			updateCourses: []libpolybase.CourseID{
				{Code: existingID.Code, Kind: "WrongKind", Part: existingID.Part},
			},
		},
		{
			name: "existing code and kind with wrong part",
			updateCourses: []libpolybase.CourseID{
				{Code: existingID.Code, Kind: existingID.Kind, Part: 999},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt update with non-existent courses
			_, err := pb.UpdatePack(ctx, "testuser", created.ID, libpolybase.PartialPack{
				Courses: &tt.updateCourses,
			})

			if err == nil {
				t.Fatal("expected error when updating pack with non-existent courses, got nil")
			}

			if !strings.Contains(err.Error(), "does not exist") {
				t.Errorf("got error %q, want error containing 'does not exist'", err.Error())
			}

			// Verify pack remains unchanged
			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    initialName,
				Courses: initialCourses,
			})

			// Verify original course relationship remains intact
			db.AssertCourseInPack(created.ID, existingID)

			if count := db.CountPackCourses(created.ID); count != len(initialCourses) {
				t.Errorf("pack course count = %d, want %d", count, len(initialCourses))
			}
		})
	}
}

// Updating pack with duplicate courses fails properly
func TestUpdatePackWithDuplicateCourses(t *testing.T) {
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
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	initialName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", initialName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	tests := []struct {
		name          string
		updateCourses []libpolybase.CourseID
	}{
		{
			name: "same course repeated",
			updateCourses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
			},
		},
		{
			name: "multiple duplicate courses",
			updateCourses: []libpolybase.CourseID{
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
			},
		},
		{
			name: "duplicates with different order",
			updateCourses: []libpolybase.CourseID{
				{Code: "CS102", Kind: "TME", Part: 1},
				{Code: "CS101", Kind: "Cours", Part: 1},
				{Code: "CS102", Kind: "TME", Part: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pb.UpdatePack(ctx, "testuser", created.ID, libpolybase.PartialPack{
				Courses: &tt.updateCourses,
			})
			if err == nil {
				t.Fatal("expected error when updating pack with duplicate courses, got nil")
			}

			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    initialName,
				Courses: initialCourses,
			})

			for _, courseID := range initialCourses {
				db.AssertCourseInPack(created.ID, courseID)
			}

			if count := db.CountPackCourses(created.ID); count != len(initialCourses) {
				t.Errorf("pack course count = %d, want %d", count, len(initialCourses))
			}
		})
	}
}

// Updating non-existent pack returns appropriate error
func TestUpdateNonExistentPack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	course := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming I",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	tests := []struct {
		name   string
		packID int
		update libpolybase.PartialPack
	}{
		{
			name:   "non-existent positive ID",
			packID: 999,
			update: libpolybase.PartialPack{
				Name: stringPtr("New Name"),
			},
		},
		{
			name:   "zero ID",
			packID: 0,
			update: libpolybase.PartialPack{
				Courses: &[]libpolybase.CourseID{
					{Code: "CS101", Kind: "Cours", Part: 1},
				},
			},
		},
		{
			name:   "negative ID",
			packID: -1,
			update: libpolybase.PartialPack{
				Name:    stringPtr("New Name"),
				Courses: &[]libpolybase.CourseID{{Code: "CS101", Kind: "Cours", Part: 1}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pb.UpdatePack(ctx, "testuser", tt.packID, tt.update)
			if err == nil {
				t.Fatal("expected error when updating non-existent pack, got nil")
			}

			db.AssertPackNotExists(tt.packID)
		})
	}
}

// Update with no changes succeeds without modifications
func TestUpdatePackWithNoChanges(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create test course
	course := libpolybase.Course{
		Code:     "CS101",
		Kind:     "Cours",
		Part:     1,
		Parts:    1,
		Name:     "Programming I",
		Quantity: 50,
		Total:    100,
		Shown:    true,
		Semester: "S1",
	}

	_, err := pb.CreateCourse(ctx, "testuser", course)
	if err != nil {
		t.Fatalf("failed to create course: %v", err)
	}

	// Create initial pack
	initialName := "Programming Pack"
	initialCourses := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
	}

	created, err := pb.CreatePack(ctx, "testuser", initialName, initialCourses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	tests := []struct {
		name   string
		update libpolybase.PartialPack
	}{
		{
			name: "same name",
			update: libpolybase.PartialPack{
				Name: &initialName,
			},
		},
		{
			name: "same courses",
			update: libpolybase.PartialPack{
				Courses: &initialCourses,
			},
		},
		{
			name: "same name and courses",
			update: libpolybase.PartialPack{
				Name:    &initialName,
				Courses: &initialCourses,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := pb.UpdatePack(ctx, "testuser", created.ID, tt.update)
			if err != nil {
				t.Fatalf("failed to update pack with no changes: %v", err)
			}

			if updated.Name != initialName {
				t.Errorf("pack name = %q, want %q", updated.Name, initialName)
			}

			db.AssertPackEqual(created.ID, libpolybase.Pack{
				ID:      created.ID,
				Name:    initialName,
				Courses: initialCourses,
			})

			for _, courseID := range initialCourses {
				db.AssertCourseInPack(created.ID, courseID)
			}

			if count := db.CountPackCourses(created.ID); count != len(initialCourses) {
				t.Errorf("pack course count = %d, want %d", count, len(initialCourses))
			}
		})
	}
}
