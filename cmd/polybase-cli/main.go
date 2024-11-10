package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		showVersion bool
	)

	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.Parse()

	if showVersion {
		fmt.Println("polybase-cli version 0.1.0")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: polybase-cli [command] [args...]")
		fmt.Println("\nCommands:")
		fmt.Println("  course list     - List all courses")
		fmt.Println("  course add      - Add a new course")
		fmt.Println("  course modify   - Modify course details")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "course":
		if len(os.Args) < 3 {
			fmt.Println("Usage: polybase-cli course [list|add|modify]")
			os.Exit(1)
		}
		handleCourse(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleCourse(cmd string) {
	switch cmd {
	case "list":
		fmt.Println("Listing courses...")
	case "add":
		fmt.Println("Adding course...")
	case "modify":
		fmt.Println("Modifying course...")
	default:
		fmt.Printf("Unknown course command: %s\n", cmd)
		os.Exit(1)
	}
}
