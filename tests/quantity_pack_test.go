package tests

import (
	"context"
	"math"
	"testing"

	"git.sr.ht/~alias/polybase-go/libpolybase"
)

// Pack quantities can be increased within bounds
func TestUpdatePackQuantityIncrease(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()

	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 20,
			Total:    50,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "Lab",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 15,
			Total:    30,
			Shown:    true,
			Semester: "S1",
		},
	}

	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Lecture", Part: 1},
		{Code: "CS102", Kind: "Lab", Part: 1},
	}

	tests := []struct {
		name       string
		delta      int
		wantError  bool
		quantities map[string]int
	}{
		{
			name:      "increase by 10",
			delta:     10,
			wantError: false,
			quantities: map[string]int{
				"CS101": 30,
				"CS102": 25,
			},
		},
		{
			name:      "increase to maximum",
			delta:     15,
			wantError: false,
			quantities: map[string]int{
				"CS101": 35,
				"CS102": 30,
			},
		},
		{
			name:      "exceed maximum",
			delta:     30,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database to initial state
			db.Clear()
			for _, course := range courses {
				db.Insert(course)
			}

			// Create fresh pack
			pack, err := pb.CreatePack(ctx, "testuser", "Programming Pack", courseIDs)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}

			// Attempt quantity update
			_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, tt.delta)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify course quantities
			for code, wantQuantity := range tt.quantities {
				var courseID libpolybase.CourseID
				for _, id := range courseIDs {
					if id.Code == code {
						courseID = id
						break
					}
				}

				course := db.Get(courseID)
				if course.Quantity != wantQuantity {
					t.Errorf("course %s quantity = %d, want %d",
						code, course.Quantity, wantQuantity)
				}
			}
		})
	}
}

// Pack quantities can be decreased within bounds
func TestUpdatePackQuantityDecrease(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()
	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 20,
			Total:    50,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "Lab",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 15,
			Total:    30,
			Shown:    true,
			Semester: "S1",
		},
	}
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Lecture", Part: 1},
		{Code: "CS102", Kind: "Lab", Part: 1},
	}
	tests := []struct {
		name       string
		delta      int
		quantities map[string]int
	}{
		{
			name:  "decrease by 10",
			delta: -10,
			quantities: map[string]int{
				"CS101": 10,
				"CS102": 5,
			},
		},
		{
			name:  "decrease to zero",
			delta: -20,
			quantities: map[string]int{
				"CS101": 0,
				"CS102": 0,
			},
		},
		{
			name:  "attempt decrease below zero",
			delta: -30,
			quantities: map[string]int{
				"CS101": 0,
				"CS102": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database to initial state
			db.Clear()
			for _, course := range courses {
				db.Insert(course)
			}
			// Create fresh pack
			pack, err := pb.CreatePack(ctx, "testuser", "Programming Pack", courseIDs)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}
			// Attempt quantity update
			_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, tt.delta)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Verify course quantities
			for code, wantQuantity := range tt.quantities {
				var courseID libpolybase.CourseID
				for _, id := range courseIDs {
					if id.Code == code {
						courseID = id
						break
					}
				}
				course := db.Get(courseID)
				if course.Quantity != wantQuantity {
					t.Errorf("course %s quantity = %d, want %d",
						code, course.Quantity, wantQuantity)
				}
			}
		})
	}
}

// Updating pack quantity respects individual course limits
func TestUpdatePackQuantityCourseLimits(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()
	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 20,
			Total:    25, // Small headroom
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "Lab",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 15,
			Total:    50, // Large headroom
			Shown:    true,
			Semester: "S1",
		},
	}
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Lecture", Part: 1},
		{Code: "CS102", Kind: "Lab", Part: 1},
	}
	tests := []struct {
		name       string
		delta      int
		wantError  bool
		quantities map[string]int
	}{
		{
			name:      "small increase within all limits",
			delta:     3,
			wantError: false,
			quantities: map[string]int{
				"CS101": 23,
				"CS102": 18,
			},
		},
		{
			name:      "fails if any course exceeds limit",
			delta:     7,
			wantError: true,
			quantities: map[string]int{
				"CS101": 20, // Should remain unchanged
				"CS102": 15, // Should remain unchanged
			},
		},
		{
			name:      "at exact limit of first course",
			delta:     5,
			wantError: false,
			quantities: map[string]int{
				"CS101": 25,
				"CS102": 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database to initial state
			db.Clear()
			for _, course := range courses {
				db.Insert(course)
			}
			// Create fresh pack
			pack, err := pb.CreatePack(ctx, "testuser", "Programming Pack", courseIDs)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}
			// Attempt quantity update
			_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, tt.delta)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				// Verify no changes were made
				for code, wantQuantity := range tt.quantities {
					var courseID libpolybase.CourseID
					for _, id := range courseIDs {
						if id.Code == code {
							courseID = id
							break
						}
					}
					course := db.Get(courseID)
					if course.Quantity != wantQuantity {
						t.Errorf("course %s quantity = %d, want %d",
							code, course.Quantity, wantQuantity)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Verify course quantities
			for code, wantQuantity := range tt.quantities {
				var courseID libpolybase.CourseID
				for _, id := range courseIDs {
					if id.Code == code {
						courseID = id
						break
					}
				}
				course := db.Get(courseID)
				if course.Quantity != wantQuantity {
					t.Errorf("course %s quantity = %d, want %d",
						code, course.Quantity, wantQuantity)
				}
			}
		})
	}
}

// Negative and zero quantity updates are handled properly
func TestUpdatePackQuantityEdgeCases(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()
	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 20,
			Total:    50,
			Shown:    true,
			Semester: "S1",
		},
		{
			Code:     "CS102",
			Kind:     "Lab",
			Part:     1,
			Parts:    1,
			Name:     "Programming Lab",
			Quantity: 15,
			Total:    30,
			Shown:    true,
			Semester: "S1",
		},
	}
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Lecture", Part: 1},
		{Code: "CS102", Kind: "Lab", Part: 1},
	}
	tests := []struct {
		name       string
		delta      int
		wantError  bool
		quantities map[string]int
	}{
		{
			name:      "zero update does nothing",
			delta:     0,
			wantError: false,
			quantities: map[string]int{
				"CS101": 20,
				"CS102": 15,
			},
		},
		{
			name:      "negative zero equivalent update (-0) does nothing",
			delta:     -0,
			wantError: false,
			quantities: map[string]int{
				"CS101": 20,
				"CS102": 15,
			},
		},
		{
			name:      "minimum negative value is handled",
			delta:     math.MinInt,
			wantError: false,
			quantities: map[string]int{
				"CS101": 0,
				"CS102": 0,
			},
		},
		{
			name:      "very large negative number is handled",
			delta:     -999999,
			wantError: false,
			quantities: map[string]int{
				"CS101": 0,
				"CS102": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database to initial state
			db.Clear()
			for _, course := range courses {
				db.Insert(course)
			}
			// Create fresh pack
			pack, err := pb.CreatePack(ctx, "testuser", "Programming Pack", courseIDs)
			if err != nil {
				t.Fatalf("failed to create pack: %v", err)
			}

			// For sequential update test, do multiple updates
			if tt.name == "sequential negative updates work properly" {
				// First update
				_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, -5)
				if err != nil {
					t.Fatalf("unexpected error on first update: %v", err)
				}
				// Second update
				_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, -5)
				if err != nil {
					t.Fatalf("unexpected error on second update: %v", err)
				}
				// Third update
				_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, -5)
				if err != nil {
					t.Fatalf("unexpected error on third update: %v", err)
				}
			} else {
				// Single update for other tests
				_, err = pb.UpdatePackQuantity(ctx, "testuser", pack.ID, tt.delta)
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify course quantities
			for code, wantQuantity := range tt.quantities {
				var courseID libpolybase.CourseID
				for _, id := range courseIDs {
					if id.Code == code {
						courseID = id
						break
					}
				}
				course := db.Get(courseID)
				if course.Quantity != wantQuantity {
					t.Errorf("course %s quantity = %d, want %d",
						code, course.Quantity, wantQuantity)
				}
			}
		})
	}
}

// Quantity updates on non-existent pack fail properly
func TestUpdatePackQuantityNonExistent(t *testing.T) {
	db := NewDB(t)
	pb := libpolybase.New(db.DB, "", false)
	ctx := context.Background()
	courses := []libpolybase.Course{
		{
			Code:     "CS101",
			Kind:     "Lecture",
			Part:     1,
			Parts:    1,
			Name:     "Programming I",
			Quantity: 20,
			Total:    50,
			Shown:    true,
			Semester: "S1",
		},
	}
	courseIDs := []libpolybase.CourseID{
		{Code: "CS101", Kind: "Lecture", Part: 1},
	}

	tests := []struct {
		name      string
		setupFunc func(t *testing.T) int // Returns pack ID to test
		delta     int
	}{
		{
			name: "non-existent pack ID",
			setupFunc: func(t *testing.T) int {
				return 99999 // Use an ID that doesn't exist
			},
			delta: 10,
		},
		{
			name: "deleted pack",
			setupFunc: func(t *testing.T) int {
				// Create and then delete a pack
				pack, err := pb.CreatePack(ctx, "testuser", "Pack to Delete", courseIDs)
				if err != nil {
					t.Fatalf("failed to create pack: %v", err)
				}
				err = pb.DeletePack(ctx, "testuser", pack.ID)
				if err != nil {
					t.Fatalf("failed to delete pack: %v", err)
				}
				return pack.ID
			},
			delta: 10,
		},
		{
			name: "zero ID",
			setupFunc: func(t *testing.T) int {
				return 0
			},
			delta: 10,
		},
		{
			name: "negative ID",
			setupFunc: func(t *testing.T) int {
				return -1
			},
			delta: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset database to initial state
			db.Clear()
			for _, course := range courses {
				db.Insert(course)
			}

			// Get pack ID to test using the setup function
			packID := tt.setupFunc(t)

			// Initial quantity for verification
			var initialQuantity int
			if err := db.DB.QueryRowContext(ctx,
				"SELECT quantity FROM courses WHERE code = ? AND kind = ? AND part = ?",
				"CS101", "Lecture", 1).Scan(&initialQuantity); err != nil {
				t.Fatalf("failed to get initial quantity: %v", err)
			}

			// Attempt quantity update on non-existent pack
			_, err := pb.UpdatePackQuantity(ctx, "testuser", packID, tt.delta)
			if err == nil {
				t.Error("expected error updating non-existent pack, got nil")
			}

			// Verify course quantity remained unchanged
			var finalQuantity int
			if err := db.DB.QueryRowContext(ctx,
				"SELECT quantity FROM courses WHERE code = ? AND kind = ? AND part = ?",
				"CS101", "Lecture", 1).Scan(&finalQuantity); err != nil {
				t.Fatalf("failed to get final quantity: %v", err)
			}

			if finalQuantity != initialQuantity {
				t.Errorf("course quantity changed from %d to %d, expected no change",
					initialQuantity, finalQuantity)
			}
		})
	}
}
