package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"

	"golang.org/x/term"

	"git.sr.ht/~alias/polybase-go/internal"
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
	jsonOutput := flags.Bool("json", false, "output in JSON format")

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
		Parts:    0,
		Name:     *name,
		Quantity: *quantity,
		Total:    *total,
		Shown:    true,
		Semester: *semester,
	}

	username := getCurrentUser()
	created, err := pb.CreateCourse(ctx, username, course)
	if err != nil {
		return err
	}

	return printCourse(created, *jsonOutput)
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

	flags := flag.NewFlagSet("get", flag.ContinueOnError)
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if err := flags.Parse(args[3:]); err != nil {
		return err
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	course, err := pb.GetCourse(ctx, id)
	if err != nil {
		return err
	}

	return printCourse(course, *jsonOutput)
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

	id := internal.CourseID{
		Code: code,
		Kind: kind,
		Part: part,
	}

	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	newCode := flags.String("c", "", "update code")
	newKind := flags.String("k", "", "update kind")
	newPart := flags.Int("p", 0, "update part")
	newName := flags.String("n", "", "update name")
	newQuantity := flags.Int("q", 0, "update quantity")
	newTotal := flags.Int("t", 0, "update total")
	newSemester := flags.String("s", "", "update semester")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if err := flags.Parse(args[3:]); err != nil {
		return err
	}

	partial := internal.PartialCourse{}
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
		}
	})

	username := getCurrentUser()
	updated, err := pb.UpdateCourse(ctx, username, id, partial)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
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

func runList(pb internal.Polybase, ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
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

	flags := flag.NewFlagSet("get", flag.ContinueOnError)
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if err := flags.Parse(args[4:]); err != nil {
		return err
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	username := getCurrentUser()
	updated, err := pb.UpdateCourseQuantity(ctx, username, id, delta)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
}

func runVisibility(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		printVisibilityUsage()
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	flags := flag.NewFlagSet("visibility", flag.ContinueOnError)
	shown := flags.Bool("s", true, "visibility state")
	jsonOutput := flags.Bool("json", false, "output in JSON format")

	if err := flags.Parse(args[3:]); err != nil {
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

	username := getCurrentUser()
	updated, err := pb.UpdateCourseShown(ctx, username, id, *shown)
	if err != nil {
		return err
	}

	return printCourse(updated, *jsonOutput)
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

func getCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		return "unknown-user"
	}

	// Extract just the username part, removing domain if present
	username := currentUser.Username
	if i := strings.LastIndex(username, "\\"); i >= 0 {
		username = username[i+1:]
	}
	return username
}
