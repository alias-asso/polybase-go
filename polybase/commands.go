package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"

	"github.com/alias-asso/polybase-go/libpolybase"
)

var (
	ErrInvalidUsage = errors.New("invalid usage")
)

func scope(args []string, usage func()) ([]string, string, string, uint8, error) {
	if len(args) < 3 {
		usage()
		return nil, "", "", 0, errors.Join(ErrInvalidUsage, errors.New("CODE, KIND and PART are required"))
	}
	part, err := strconv.Atoi(args[2])
	if err != nil || part < 0 || part > 255 {
		return nil, "", "", 0, errors.Join(ErrInvalidUsage, fmt.Errorf("invalid part number: %s", args[2]))
	}
	return args[3:], args[0], args[1], uint8(part), nil
}

func runCreate(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("create", flag.ExitOnError)
	flags.Usage = createUsage(flags)

	name := flags.String("n", "", "course name")
	quantity := flags.Int("q", -1, "initial quantity")
	total := flags.Int("t", 0, "total quantity")
	semester := flags.String("s", "", "semester")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	args, code, kind, part, err := scope(args, flags.Usage)
	if err != nil {
		return err
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *name == "" || *quantity == -1 || *semester == "" {
		createUsage(flags)
		return errors.Join(ErrInvalidUsage, fmt.Errorf("name (-n), quantity (-q) and semester (-s) are required"))
	}

	if *total == 0 {
		*total = *quantity
	}

	created, err := pb.CreateCourse(ctx, getCurrentUser(), libpolybase.Course{
		Code:     code,
		Kind:     kind,
		Part:     int(part),
		Name:     *name,
		Quantity: *quantity,
		Total:    *total,
		Shown:    true,
		Semester: *semester,
	})
	if err != nil {
		return err
	}

	return printCourse(created, *jsonOutput)
}

func runGet(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	flags.Usage = getUsage(flags)

	jsonOutput := flags.Bool("json", false, "output in JSON format")

	args, code, kind, part, err := scope(args, flags.Usage)
	if err != nil {
		return err
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	id := libpolybase.CourseID{
		Code: code,
		Kind: kind,
		Part: int(part),
	}

	course, err := pb.GetCourse(ctx, id)
	if err != nil {
		return err
	}

	return printCourse(course, *jsonOutput)
}

func runUpdate(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("update", flag.ExitOnError)
	flags.Usage = updateUsage(flags)

	newCode := flags.String("c", "", "update code")
	newKind := flags.String("k", "", "update kind")
	newPart := flags.Int("p", 0, "update part")
	newName := flags.String("n", "", "update name")
	newQuantity := flags.Int("q", 0, "update quantity")
	newTotal := flags.Int("t", 0, "update total")
	newSemester := flags.String("s", "", "update semester")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	args, code, kind, part, err := scope(args, flags.Usage)
	if err != nil {
		return err
	}

	id := libpolybase.CourseID{
		Code: code,
		Kind: kind,
		Part: int(part),
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	partial := libpolybase.PartialCourse{}
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c":
			partial.Code = newCode
		case "k":
			partial.Kind = newKind
		case "p":
			partial.Part = newPart
		case "n":
			partial.Name = newName
		case "q":
			partial.Quantity = newQuantity
		case "t":
			partial.Total = newTotal
		case "s":
			partial.Semester = newSemester
		default:
			panic(errors.Join(ErrInvalidUsage, fmt.Errorf("unknown flag %s", f.Name)))
		}
	})

	username := getCurrentUser()
	updated, err := pb.UpdateCourse(ctx, username, id, partial)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
}

func runDelete(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	args, code, kind, part, err := scope(args, deleteUsage(nil))
	if err != nil {
		return err
	}

	id := libpolybase.CourseID{
		Code: code,
		Kind: kind,
		Part: int(part),
	}

	course, err := pb.GetCourse(ctx, id)
	if err != nil {
		return err
	}

	fmt.Println("Are you sure you want to delete this course?")
	fmt.Printf("  %s %s %d [y/N]: ", course.Code, course.Kind, course.Part)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	buffer := make([]byte, 1)
	_, err = os.Stdin.Read(buffer)
	if err != nil {
		return err
	}

	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		return fmt.Errorf("failed to restore terminal: %w", err)
	}

	fmt.Printf("\n")

	if buffer[0] != 'y' && buffer[0] != 'Y' {
		return nil
	}

	username := getCurrentUser()
	return pb.DeleteCourse(ctx, username, id)
}

func runList(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("list", flag.ExitOnError)
	flags.Usage = listUsage(flags)

	showHidden := flags.Bool("a", false, "show hidden courses")
	semester := flags.String("s", "", "filter by semester")
	code := flags.String("c", "", "filter by course code")
	kind := flags.String("k", "", "filter by kind")
	part := flags.Int("p", 0, "filter by part number")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if err := flags.Parse(args); err != nil {
		return err
	}

	var filterSemester, filterCode, filterKind *string
	var filterPart *int
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "s":
			filterSemester = semester
		case "c":
			filterCode = code
		case "k":
			filterKind = kind
		case "p":
			filterPart = part
		}
	})

	courses, err := pb.ListCourse(ctx, *showHidden, filterSemester, filterCode, filterKind, filterPart)
	if err != nil {
		return err
	}

	return printCourses(courses, *jsonOutput)
}

func runQuantity(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	flags.Usage = quantityUsage(flags)

	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if len(args) < 4 {
		flags.Usage()
		return fmt.Errorf("CODE, KIND, PART and DELTA are required")
	}

	args, code, kind, part, err := scope(args, flags.Usage)
	if err != nil {
		return err
	}

	delta, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid delta value: %s", args[0])
	}

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	id := libpolybase.CourseID{
		Code: code,
		Kind: kind,
		Part: int(part),
	}

	username := getCurrentUser()
	updated, err := pb.UpdateCourseQuantity(ctx, username, id, delta)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
}

func runVisibility(pb libpolybase.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("visibility", flag.ExitOnError)
	flags.Usage = visibilityUsage(flags)

	shown := flags.Bool("s", true, "visibility state")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	args, code, kind, part, err := scope(args, flags.Usage)
	if err != nil {
		return err
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	id := libpolybase.CourseID{
		Code: code,
		Kind: kind,
		Part: int(part),
	}

	username := getCurrentUser()
	updated, err := pb.UpdateCourseShown(ctx, username, id, *shown)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
}
