package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/alias-asso/polybase-go/libpolybase"
)

// Empty database returns empty pack list
func TestListPacksEmptyDatabase(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	packs, err := pb.ListPacks(ctx)
	if err != nil {
		t.Fatalf("ListPacks failed: %v", err)
	}

	if len(packs) != 0 {
		t.Errorf("got %d packs, want 0", len(packs))
	}

	var packCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM packs").Scan(&packCount); err != nil {
		t.Fatalf("failed to count packs: %v", err)
	}

	if packCount != 0 {
		t.Errorf("database contains %d packs, want 0", packCount)
	}

	var linkCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM pack_courses").Scan(&linkCount); err != nil {
		t.Fatalf("failed to count pack_courses: %v", err)
	}

	if linkCount != 0 {
		t.Errorf("database contains %d pack_courses links, want 0", linkCount)
	}
}

// Single pack is listed correctly
func TestListSinglePack(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	courses := []libpolybase.Course{
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
			Kind:     "Cours",
			Part:     2,
			Parts:    2,
			Name:     "Programming II",
			Quantity: 45,
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
	}

	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	pack := libpolybase.Pack{
		Name: "Programming Bundle",
		Courses: []libpolybase.CourseID{
			{Code: "CS101", Kind: "Cours", Part: 1},
			{Code: "CS101", Kind: "Cours", Part: 2},
			{Code: "CS101", Kind: "TME", Part: 1},
		},
	}

	created, err := pb.CreatePack(ctx, "testuser", pack.Name, pack.Courses)
	if err != nil {
		t.Fatalf("failed to create pack: %v", err)
	}

	packs, err := pb.ListPacks(ctx)
	if err != nil {
		t.Fatalf("ListPacks failed: %v", err)
	}

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	got := packs[0]
	want := libpolybase.Pack{
		ID:      created.ID,
		Name:    pack.Name,
		Courses: pack.Courses,
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

	wantCourses := make(map[string]bool)
	for _, course := range want.Courses {
		wantCourses[course.ID()] = true
	}

	gotCourses := make(map[string]bool)
	for _, course := range got.Courses {
		gotCourses[course.ID()] = true
	}

	for _, course := range want.Courses {
		if !gotCourses[course.ID()] {
			t.Errorf("missing expected course: %+v", course)
		}
	}

	for _, course := range got.Courses {
		if !wantCourses[course.ID()] {
			t.Errorf("got unexpected course: %+v", course)
		}
	}

	db.AssertPackExists(created.ID)
	db.AssertPackEqual(created.ID, want)

	for _, courseID := range pack.Courses {
		db.AssertCourseInPack(created.ID, courseID)
	}

	if got := db.CountPackCourses(created.ID); got != len(pack.Courses) {
		t.Errorf("pack course count = %d, want %d", got, len(pack.Courses))
	}
}

// Multiple packs are listed in correct order
func TestListPacksOrder(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create three distinct courses
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

	// Create courses
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Prepare course IDs
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
		{Code: "CS103", Kind: "TD", Part: 1},
	}

	// Create packs
	packNames := []string{"Pack with 1 and 2", "Pack with 1 and 3", "Pack with 2 and 3"}
	packCourses := [][]libpolybase.CourseID{
		{courseIDs[0], courseIDs[1]},
		{courseIDs[0], courseIDs[2]},
		{courseIDs[1], courseIDs[2]},
	}

	var createdPacks []libpolybase.Pack
	for i, name := range packNames {
		pack, err := pb.CreatePack(ctx, "testuser", name, packCourses[i])
		if err != nil {
			t.Fatalf("failed to create pack %q: %v", name, err)
		}
		createdPacks = append(createdPacks, pack)
	}

	// List packs
	packs, err := pb.ListPacks(ctx)
	if err != nil {
		t.Fatalf("ListPacks failed: %v", err)
	}

	// Verify pack count
	if len(packs) != len(packNames) {
		t.Fatalf("got %d packs, want %d", len(packs), len(packNames))
	}

	// Verify packs are listed in order of their IDs
	for i, pack := range packs {
		want := createdPacks[i]

		// Use AssertPackEqual for robust comparison
		db.AssertPackEqual(pack.ID, want)
	}
}

// Pack list includes all associated courses
func TestListPackIncludesAllCourses(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create courses
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

	// Create courses
	for _, course := range courses {
		_, err := pb.CreateCourse(ctx, "testuser", course)
		if err != nil {
			t.Fatalf("failed to create course %s: %v", course.ID(), err)
		}
	}

	// Prepare course IDs
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Cours", Part: 1},
		{Code: "CS102", Kind: "TME", Part: 1},
		{Code: "CS103", Kind: "TD", Part: 1},
	}

	// Create packs
	packDefinitions := []struct {
		name    string
		courses []libpolybase.CourseID
	}{
		{
			name:    "Pack 1",
			courses: []libpolybase.CourseID{courseIDs[0], courseIDs[1]},
		},
		{
			name:    "Pack 2",
			courses: []libpolybase.CourseID{courseIDs[1], courseIDs[2]},
		},
	}

	var createdPacks []libpolybase.Pack
	for _, packDef := range packDefinitions {
		pack, err := pb.CreatePack(ctx, "testuser", packDef.name, packDef.courses)
		if err != nil {
			t.Fatalf("failed to create pack %q: %v", packDef.name, err)
		}
		createdPacks = append(createdPacks, pack)
	}

	// List packs
	packs, err := pb.ListPacks(ctx)
	if err != nil {
		t.Fatalf("ListPacks failed: %v", err)
	}

	// Verify pack count
	if len(packs) != len(packDefinitions) {
		t.Fatalf("got %d packs, want %d", len(packs), len(packDefinitions))
	}

	// Verify courses in packs
	for i, pack := range packs {
		t.Logf("Checking pack %s", pack.Name)
		t.Logf("Pack courses: %+v", pack.Courses)

		expectedCourses := packDefinitions[i].courses

		if len(pack.Courses) != len(expectedCourses) {
			t.Fatalf("pack %s: got %d courses, want %d",
				pack.Name, len(pack.Courses), len(expectedCourses))
		}

		// Use AssertPackEqual for robust comparison
		db.AssertPackEqual(pack.ID, createdPacks[i])
	}
}

// Pack list handles large number of packs efficiently
func TestListPacksLargeNumberOfPacks(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	// Create 10 courses
	courses := make([]libpolybase.Course, 10)
	for i := 0; i < 10; i++ {
		courses[i] = libpolybase.Course{
			Code:     fmt.Sprintf("CS%03d", 100+i),
			Kind:     "Cours",
			Part:     1,
			Parts:    1,
			Name:     fmt.Sprintf("Course %d", i+1),
			Quantity: 50,
			Total:    100,
			Shown:    true,
			Semester: "S1",
		}

		// Create the course
		_, err := pb.CreateCourse(ctx, "testuser", courses[i])
		if err != nil {
			t.Fatalf("failed to create course %s: %v", courses[i].ID(), err)
		}
	}

	// Prepare course IDs
	courseIDs := make([]libpolybase.CourseID, 10)
	for i, course := range courses {
		courseIDs[i] = libpolybase.CourseID{
			Code: course.Code,
			Kind: course.Kind,
			Part: course.Part,
		}
	}

	// Create packs with 2 unique courses
	var createdPackIDs []int
	packCount := 0

	for i := 0; i < len(courseIDs); i++ {
		for j := i + 1; j < len(courseIDs); j++ {
			// Create a pack with these two courses
			packName := fmt.Sprintf("Pack %d", packCount)
			packCourses := []libpolybase.CourseID{courseIDs[i], courseIDs[j]}

			pack, err := pb.CreatePack(ctx, "testuser", packName, packCourses)
			if err != nil {
				t.Fatalf("failed to create pack %q: %v", packName, err)
			}

			createdPackIDs = append(createdPackIDs, pack.ID)
			packCount++
		}
	}

	// List packs
	packs, err := pb.ListPacks(ctx)
	if err != nil {
		t.Fatalf("ListPacks failed: %v", err)
	}

	// Verify pack count
	if len(packs) != packCount {
		t.Fatalf("got %d packs, want %d", len(packs), packCount)
	}

	// Create a map of pack IDs to verify
	createdPackIDSet := make(map[int]bool)
	for _, id := range createdPackIDs {
		createdPackIDSet[id] = true
	}

	// Verify each pack
	for _, pack := range packs {
		// Check if the pack ID exists in our created packs
		if !createdPackIDSet[pack.ID] {
			t.Errorf("unexpected pack ID %d", pack.ID)
		}

		// Retrieve the full pack details to verify
		fullPack, err := pb.GetPack(ctx, pack.ID)
		if err != nil {
			t.Fatalf("failed to get pack %d: %v", pack.ID, err)
		}

		// Use AssertPackEqual for robust comparison
		db.AssertPackEqual(pack.ID, fullPack)

		// Remove the pack ID from the set
		delete(createdPackIDSet, pack.ID)
	}

	// Ensure all created packs were found
	if len(createdPackIDSet) > 0 {
		t.Errorf("missed packs: %v", createdPackIDSet)
	}

	t.Logf("Created %d packs", packCount)
}
