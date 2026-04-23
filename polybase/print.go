package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alias-asso/polybase-go/libpolybase"
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
`, defaultDBPath)
}

func printVersion() {
	fmt.Printf("polybase version %s\n", version)
}

func usage(usage, target string, flags *flag.FlagSet) func() {
	return func() {
		fmt.Printf("Usage: %s\n\t%s\n\n", usage, target)
		if flags != nil {
			fmt.Println("Options:")
			flags.VisitAll(func(f *flag.Flag) {
				def := f.DefValue
				if len(def) == 0 {
					def = `""`
				}
				fmt.Printf("-%s\t%s (default: %v)\n", f.Name, f.Usage, def)
			})
			fmt.Println()
		}
	}
}

func createUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase create <CODE> <KIND> <PART> [OPTIONS]`,
		`Create a new course entry.`,
		flags,
	)
}

func getUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase get <CODE> <KIND> <PART> [OPTIONS]`,
		`Display details for a specific course`,
		flags,
	)
}

func updateUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase update <CODE> <KIND> <PART> [OPTIONS]`,
		`Update course information`,
		flags,
	)
}

func deleteUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase delete <CODE> <KIND> <PART>`,
		`Remove a course from the database`,
		flags,
	)
}

func listUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase list [OPTIONS]`,
		`List all courses`,
		flags,
	)
}

func quantityUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase quantity <CODE> <KIND> <PART> <DELTA> [OPTIONS]`,
		`Update course quantity by adding DELTA (can be negative)`,
		flags,
	)
}

func visibilityUsage(flags *flag.FlagSet) func() {
	return usage(
		`polybase visibility <CODE> <KIND> <PART> [-s STATE] [OPTIONS]`,
		`Set course visibility`,
		flags,
	)
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

func newCourseJSON(c *libpolybase.Course) CourseJSON {
	return CourseJSON{
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
}

func printCourses(courses []libpolybase.Course, jsonOutput bool) error {
	if jsonOutput {
		var coursesJSON []CourseJSON
		for _, c := range courses {
			coursesJSON = append(coursesJSON, newCourseJSON(&c))
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

func printCourse(c libpolybase.Course, jsonOutput bool) error {
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(newCourseJSON(&c))
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
