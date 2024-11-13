package internal

import (
	"context"
)

type Course struct {
	Code     string
	Kind     string
	Part     int
	Parts    int
	Name     string
	Quantity int
	Total    int
	Shown    bool
	Semester string
}

type CourseID struct {
	Code string
	Kind string
	Part int
}

type Polybase interface {
	Create(ctx context.Context, course Course) (Course, error)
	Get(ctx context.Context, id CourseID) (Course, error)
	Update(ctx context.Context, id CourseID, course Course) (Course, error)
	Delete(ctx context.Context, id CourseID) error
	List(ctx context.Context, showHidden bool) ([]Course, error)

	UpdateQuantity(ctx context.Context, id CourseID, delta int) (Course, error)
	UpdateShown(ctx context.Context, id CourseID, shown bool) (Course, error)
}
