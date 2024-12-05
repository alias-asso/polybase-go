package internal

import (
	"context"
)

type CourseID struct {
	Code string
	Kind string
	Part int
}

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

type PartialCourse struct {
	Code     *string
	Kind     *string
	Part     *int
	Parts    *int
	Name     *string
	Quantity *int
	Total    *int
	Shown    *bool
	Semester *string
}

type Polybase interface {
	Create(ctx context.Context, cours Course) (Course, error)
	Get(ctx context.Context, id CourseID) (Course, error)
	Update(ctx context.Context, id CourseID, partial PartialCourse) (Course, error)
	Delete(ctx context.Context, id CourseID) error
	List(ctx context.Context, showHidden bool, filterSemester *string, filterCode *string, filterKind *string, filterPart *int) ([]Course, error)

	UpdateQuantity(ctx context.Context, id CourseID, delta int) (Course, error)
	UpdateShown(ctx context.Context, id CourseID, shown bool) (Course, error)
}
