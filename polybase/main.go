package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"git.sr.ht/~alias/polybase/internal"
	_ "github.com/mattn/go-sqlite3"
)

const defaultDBPath = "/var/lib/polybase/polybase.db"

type command struct {
	Name        string
	Description string
	Usage       string
	Run         func(pb internal.Polybase, ctx context.Context, args []string) error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flags.SetOutput(nil)

	dbFlag := flags.String("db", defaultDBPath, "database path")

	if err := flags.Parse(os.Args[1:]); err != nil {
		printUsage()
		return nil
	}

	if flag.NArg() == 0 {
		printUsage()
		return nil
	}

	args := flag.Args()
	dbPath := *dbFlag

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		dbErr := fmt.Errorf("failed to open database at %s: %w", dbPath, err)

		if len(args) == 0 {
			return dbErr
		}

		cmd := args[0]
		if cmd == "-h" || cmd == "--help" || cmd == "help" {
			printUsage()
			return nil
		}

		// Check if the command exists
		if _, ok := getCommands()[cmd]; !ok {
			return fmt.Errorf("invalid command %q and %v", cmd, dbErr)
		}

		// If the command is valid but database failed, return just the database error
		return dbErr
	}
	defer db.Close()

	// Handle command
	cmd := args[0]
	if cmd == "-h" || cmd == "--help" {
		cmd = "help"
	}

	commands := getCommands()
	command, ok := commands[cmd]
	if !ok {
		return fmt.Errorf("unknown command %q", cmd)
	}

	return command.Run(internal.New(db), context.Background(), args[1:])
}

func getCommands() map[string]command {
	return map[string]command{
		"create": {
			Name:        "create",
			Description: "Create a new course",
			Usage:       "create <CODE> <KIND> [-n NAME] [-P PART] [-p PARTS] [-q QUANTITY] [-t TOTAL] [-S SEMESTER]",
			Run:         runCreate,
		},
		"get": {
			Name:        "get",
			Description: "Get course details",
			Usage:       "get <CODE> <KIND> <PART>",
			Run:         runGet,
		},
		"update": {
			Name:        "update",
			Description: "Update course details",
			Usage:       "update <CODE> <KIND> <PART> [-n NAME] [-p PARTS] [-q QUANTITY] [-t TOTAL] [-S SEMESTER]",
			Run:         runUpdate,
		},
		"delete": {
			Name:        "delete",
			Description: "Delete a course",
			Usage:       "delete <CODE> <KIND> <PART>",
			Run:         runDelete,
		},
		"list": {
			Name:        "list",
			Description: "List all courses",
			Usage:       "list [-a] (include hidden courses)",
			Run:         runList,
		},
		"quantity": {
			Name:        "quantity",
			Description: "Update course quantity",
			Usage:       "quantity <CODE> <KIND> <PART> <DELTA>",
			Run:         runUpdateQuantity,
		},
		"visibility": {
			Name:        "visibility",
			Description: "Toggle course visibility",
			Usage:       "visibility <CODE> <KIND> <PART> [-s STATE]",
			Run:         runUpdateVisibility,
		},
		"help": {
			Name:        "help",
			Description: "Show help message",
			Usage:       "help [command]",
			Run: func(pb internal.Polybase, ctx context.Context, args []string) error {
				return runHelp(args)
			},
		},
	}
}

func runCreate(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("CODE and KIND are required")
	}

	code := args[0]
	kind := args[1]

	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	name := fs.String("n", "", "course name")
	part := fs.Int("P", 1, "course part number")
	parts := fs.Int("p", 1, "total number of parts")
	quantity := fs.Int("q", -1, "initial quantity (required)")
	total := fs.Int("t", 0, "total quantity")
	semester := fs.String("S", "", "semester")

	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	if *quantity == -1 {
		return fmt.Errorf("quantity (-q) is required")
	}

	if *total == 0 {
		*total = *quantity
	}

	course := internal.Course{
		Code:     code,
		Kind:     kind,
		Part:     *part,
		Parts:    *parts,
		Name:     *name,
		Quantity: *quantity,
		Total:    *total,
		Shown:    true, // Default to shown
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
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	name := fs.String("n", "", "update name")
	parts := fs.Int("p", 0, "total parts")
	quantity := fs.Int("q", 0, "quantity")
	total := fs.Int("t", 0, "total quantity")
	semester := fs.String("S", "", "semester")

	if err := fs.Parse(args[3:]); err != nil {
		return err
	}

	id := internal.CourseID{
		Code: args[0],
		Kind: args[1],
		Part: part,
	}

	course := internal.Course{
		Name:     *name,
		Parts:    *parts,
		Quantity: *quantity,
		Total:    *total,
		Semester: *semester,
	}

	updated, err := pb.Update(ctx, id, course)
	if err != nil {
		return err
	}

	printCourse(updated)
	return nil
}

func runDelete(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
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

func runUpdateQuantity(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 4 {
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

func runUpdateVisibility(pb internal.Polybase, ctx context.Context, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("CODE, KIND and PART are required")
	}

	part, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid part number: %s", args[2])
	}

	fs := flag.NewFlagSet("visibility", flag.ContinueOnError)
	shown := fs.Bool("s", true, "visibility state")

	if err := fs.Parse(args[3:]); err != nil {
		return err
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

func runList(pb internal.Polybase, ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	showHidden := fs.Bool("a", false, "show hidden courses")

	if err := fs.Parse(args); err != nil {
		return err
	}

	courses, err := pb.List(ctx, *showHidden)
	if err != nil {
		return err
	}

	for _, course := range courses {
		printCourse(course)
		fmt.Println()
	}
	return nil
}

func printCourse(c internal.Course) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Code:\t%s\n", c.Code)
	fmt.Fprintf(w, "Name:\t%s\n", c.Name)
	fmt.Fprintf(w, "Kind:\t%s\n", c.Kind)
	fmt.Fprintf(w, "Part:\t%d/%d\n", c.Part, c.Parts)
	fmt.Fprintf(w, "Quantity:\t%d/%d\n", c.Quantity, c.Total)
	fmt.Fprintf(w, "Semester:\t%s\n", c.Semester)
	fmt.Fprintf(w, "Visible:\t%v\n", c.Shown)
	w.Flush()
}

func runHelp(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "create":
		fmt.Println(`Usage: create <CODE> <KIND> [-n NAME] [-P PART] [-p PARTS] [-q QUANTITY] [-t TOTAL] [-S SEMESTER]
Flags:
  -n    course name
  -P    part number (default 1)
  -p    total parts (default 1)
  -q    initial quantity (required)
  -S    semester
  -t    total quantity (default same as quantity)`)

	case "get":
		fmt.Println(`Usage: get <CODE> <KIND> <PART>`)

	case "update":
		fmt.Println(`Usage: update <CODE> <KIND> <PART> [-n NAME] [-p PARTS] [-q QUANTITY] [-t TOTAL] [-S SEMESTER]
Flags:
  -n    update name
  -p    total parts
  -q    quantity
  -S    semester
  -t    total quantity`)

	case "delete":
		fmt.Println(`Usage: delete <CODE> <KIND> <PART>`)

	case "list":
		fmt.Println(`Usage: list [-a]
Flags:
  -a    show hidden courses`)

	case "quantity":
		fmt.Println(`Usage: quantity <CODE> <KIND> <PART> <DELTA>`)

	case "visibility":
		fmt.Println(`Usage: visibility <CODE> <KIND> <PART> [-s STATE]
Flags:
  -s    visibility state (default true)`)

	default:
		return fmt.Errorf("unknown command %q", args[0])
	}

	return nil
}

func printUsage() {
	fmt.Printf(`Usage: polybase [-db PATH] <command> [arguments]

Default database path: %s

Commands:
  create      Create a new course
  get         Get course details by CODE KIND PART
  update      Update course details
  delete      Delete a course
  list        List all courses
  quantity    Update course quantity
  visibility  Toggle course visibility
  help        Show help message

Use "polybase help <command>" for more information about a command.
`, defaultDBPath)
}
