package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"git.sr.ht/~alias/polybase/internal"
)

func printUsage() {
	fmt.Printf(`Usage: polybase [-db PATH] command [arguments]

OPTIONS
    -db PATH    Path to database file (default: %s)
    -h          Print help information
    -v          Print version information

COMMANDS
    create      Create a new course entry
    get         Display details for a specific course
    update      Update course information
    delete      Remove a course from the database
    list        List all courses
    quantity    Update course quantity
    visibility  Set course visibility
    help        Show help message for a specific command

Use "polybase help command" for more information about a command.
`, defaultDBPath)
}

func printCreateUsage() {
	fmt.Print(`Usage: polybase create <CODE> <KIND> <PART> [OPTIONS]
    Create a new course entry.

    Options:
    -n NAME      Course name (required)
    -q QUANTITY  Initial quantity (required)
    -t TOTAL     Total quantity (default: same as quantity)
    -s SEMESTER  Semester (required)
`)
}

func printGetUsage() {
	fmt.Print(`Usage: polybase get <CODE> <KIND> <PART>
    Display details for a specific course
`)
}

func printUpdateUsage() {
	fmt.Print(`Usage: polybase update <CODE> <KIND> <PART> [OPTIONS]
    Update course information

    Options:
    -c CODE      Update course code
    -k KEY       Update course key
    -p PART      Update course part
    -n NAME      Update course name
    -q QUANTITY  Update quantity
    -t TOTAL     Update total quantity
    -s SEMESTER  Update semester
`)
}

func printDeleteUsage() {
	fmt.Print(`Usage: polybase delete <CODE> <KIND> <PART>
    Remove a course from the database
`)
}

func printListUsage() {
	fmt.Print(`Usage: polybase list [OPTIONS]
    List all courses

    Options:
    -a          Show hidden courses
`)
}

func printQuantityUsage() {
	fmt.Print(`Usage: polybase quantity <CODE> <KIND> <PART> <DELTA>
    Update course quantity by adding DELTA (can be negative)
`)
}

func printVisibilityUsage() {
	fmt.Print(`Usage: polybase visibility <CODE> <KIND> <PART> [-s STATE]
    Set course visibility

    Options:
    -s          Set visibility state (default: true)
`)
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
