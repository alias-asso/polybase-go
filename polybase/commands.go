package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"

	"git.sr.ht/~alias/polybase/internal"
)

func runCreate(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printCreateUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}

	name := flags.String("n", "", "course name")
	quantity := flags.Int("q", -1, "initial quantity")
	total := flags.Int("t", 0, "total quantity")
	semester := flags.String("s", "", "semester")

	if err := flags.Parse(args[3:]); err != nil {
		return err
	}

	if *name == "" || *quantity == -1 || *semester == "" {
		printCreateUsage()
		return fmt.Errorf("name (-n), quantity (-q) and semester (-s) are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	if *total == 0 {
		*total = *quantity
	}

	course := internal.Course{
		Code:     args[0],
		Kind:     args[1],
		Part:     part,
		Parts:    1,
		Name:     *name,
		Quantity: *quantity,
		Total:    *total,
		Shown:    true,
		Semester: *semester,
	}

	created, err := pb.Create(ctx, course)
	if err != nil {
		return err
	}

	printCourse(created)
	return nil
}

func runGet(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printGetUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	course, err := pb.Get(ctx, id)
	if err != nil {
		return err
	}

	printCourse(course)
	return nil
}

func runUpdate(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printUpdateUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	code := args[0]
	kind := args[1]
	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	newCode := fs.String("c", "", "update code")
	newKind := fs.String("k", "", "update kind")
	newPart := fs.Int("p", 0, "update part")
	newName := fs.String("n", "", "update name")
	newQuantity := fs.Int("q", 0, "update quantity")
	newTotal := fs.Int("t", 0, "update total")
	newSemester := fs.String("s", "", "update semester")

	if err := fs.Parse(args[3:]); err != nil {
		return err
	}

	id := internal.CourseID{
		Code: code,
		Kind: kind,
		Part: part,
	}

	partial := internal.PartialCourse{
		Code:     newCode,
		Kind:     newKind,
		Part:     newPart,
		Name:     newName,
		Quantity: newQuantity,
		Total:    newTotal,
		Semester: newSemester,
	}

	updated, err := pb.Update(ctx, id, partial)
	if err != nil {
		return err
	}

	printCourse(updated)
	return nil
}

func runDelete(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printDeleteUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	return pb.Delete(ctx, id)
}

func runList(pb internal.Polybase, ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	showHidden := fs.Bool("a", false, "show hidden courses")
	semester := fs.String("s", "", "filter by semester")
	code := fs.String("c", "", "filter by course code")
	kind := fs.String("k", "", "filter by kind")
	part := fs.Int("p", 0, "filter by part number")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var filterSemester, filterCode, filterKind *string
	var filterPart *int

	if *semester != "" {
		filterSemester = semester
	}
	if *code != "" {
		filterCode = code
	}
	if *kind != "" {
		filterKind = kind
	}
	if *part != 0 {
		filterPart = part
	}

	courses, err := pb.List(ctx, *showHidden, filterSemester, filterCode, filterKind, filterPart)
	if err != nil {
		return err
	}

	for i, course := range courses {
		printCourse(course)

		if i != len(courses)-1 {
			fmt.Println()
		}
	}
	return nil
}

func runQuantity(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 4 {
		printQuantityUsage()
		return fmt.Errorf("CODE, KIND, PART and DELTA are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	delta, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("invalid delta value: %s", args[3])
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	updated, err := pb.UpdateQuantity(ctx, id, delta)
	if err != nil {
		return err
	}

	printCourse(updated)
	return nil
}

func runVisibility(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printVisibilityUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	fs := flag.NewFlagSet("visibility", flag.ContinueOnError)
	shown := fs.Bool("s", true, "visibility state")

	if err := fs.Parse(args[3:]); err != nil {
		return err
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	updated, err := pb.UpdateShown(ctx, id, *shown)
	if err != nil {
		return err
	}

	printCourse(updated)
	return nil
}

func runHelp(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "create":
		printCreateUsage()
	case "get":
		printGetUsage()
	case "update":
		printUpdateUsage()
	case "delete":
		printDeleteUsage()
	case "list":
		printListUsage()
	case "quantity":
		printQuantityUsage()
	case "visibility":
		printVisibilityUsage()
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}
