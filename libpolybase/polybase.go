package libpolybase

import (
	"context"
)

type CourseNotFound struct{}

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

type Pack struct {
	ID      int
	Name    string
	Courses []CourseID
}

type PartialPack struct {
	Name    *string
	Courses *[]CourseID
}

type Polybase interface {
	CreateCourse(ctx context.Context, user string, cours Course) (Course, error)
	GetCourse(ctx context.Context, id CourseID) (Course, error)
	UpdateCourse(ctx context.Context, user string, id CourseID, partial PartialCourse) (Course, error)
	DeleteCourse(ctx context.Context, user string, id CourseID) error
	ListCourse(ctx context.Context, showHidden bool, filterSemester *string, filterCode *string, filterKind *string, filterPart *int) ([]Course, error)

	UpdateCourseQuantity(ctx context.Context, user string, id CourseID, delta int) (Course, error)
	UpdateCourseShown(ctx context.Context, user string, id CourseID, shown bool) (Course, error)

	CreatePack(ctx context.Context, user string, name string, courses []CourseID) (Pack, error)
	GetPack(ctx context.Context, id int) (Pack, error)
	UpdatePack(ctx context.Context, user string, id int, partial PartialPack) (Pack, error)
	DeletePack(ctx context.Context, user string, id int) error
	ListPacks(ctx context.Context) ([]Pack, error)

	UpdatePackQuantity(ctx context.Context, user string, id int, delta int) (Pack, error)
}
