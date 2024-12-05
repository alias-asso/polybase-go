package main

import (
	"encoding/json"
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
    -json        Output in JSON format
`)
}

func printGetUsage() {
	fmt.Print(`Usage: polybase get <CODE> <KIND> <PART> [OPTIONS]
    Display details for a specific course

    Options:
    -json        Output in JSON format
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
    -json        Output in JSON format
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
    -a              Show hidden courses
    -s SEMESTER     Filter by semester
    -c CODE         Filter by code prefix
    -k KIND         Filter by kind
    -p PART         Filter by part number
    -json           Output in JSON format
`)
}

func printQuantityUsage() {
	fmt.Print(`Usage: polybase quantity <CODE> <KIND> <PART> <DELTA> [OPTIONS]
    Update course quantity by adding DELTA (can be negative)

    Options:
    -json           Output in JSON format
`)
}

func printVisibilityUsage() {
	fmt.Print(`Usage: polybase visibility <CODE> <KIND> <PART> [-s STATE] [OPTIONS]
    Set course visibility

    Options:
    -s              Set visibility state (default: true)
    -json           Output in JSON format
`)
}

type CourseJSON struct {
	Code     string `json:"code"`
	Kind     string `json:"kind"`
	Part     int    `json:"part"`
	Parts    int    `json:"parts"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Total    int    `json:"total"`
	Shown    bool   `json:"visible"`
	Semester string `json:"semester"`
}

func printCourses(courses []internal.Course, jsonOutput bool) error {
	if jsonOutput {
		var coursesJSON []CourseJSON
		for _, c := range courses {
			coursesJSON = append(coursesJSON, CourseJSON{
				Code:     c.Code,
				Kind:     c.Kind,
				Part:     c.Part,
				Parts:    c.Parts,
				Name:     c.Name,
				Quantity: c.Quantity,
				Total:    c.Total,
				Shown:    c.Shown,
				Semester: c.Semester,
			})
		}

		return json.NewEncoder(os.Stdout).Encode(courses)
	}

	for i, course := range courses {
		if err := printCourse(course, false); err != nil {
			return err
		}
		if i != len(courses)-1 {
			fmt.Println()
		}
	}
	return nil
}

func printCourse(c internal.Course, jsonOutput bool) error {
	if jsonOutput {
		courseJSON := CourseJSON{
			Code:     c.Code,
			Kind:     c.Kind,
			Part:     c.Part,
			Parts:    c.Parts,
			Name:     c.Name,
			Quantity: c.Quantity,
			Total:    c.Total,
			Shown:    c.Shown,
			Semester: c.Semester,
		}
		return json.NewEncoder(os.Stdout).Encode(courseJSON)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Code:\t%s\n", c.Code)
	fmt.Fprintf(w, "Kind:\t%s\n", c.Kind)
	fmt.Fprintf(w, "Part:\t%d/%d\n", c.Part, c.Parts)
	fmt.Fprintf(w, "Name:\t%s\n", c.Name)
	fmt.Fprintf(w, "Quantity:\t%d/%d\n", c.Quantity, c.Total)
	fmt.Fprintf(w, "Semester:\t%s\n", c.Semester)
	fmt.Fprintf(w, "Visible:\t%v\n", c.Shown)
	return w.Flush()
}
