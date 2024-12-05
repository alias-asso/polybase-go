package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"git.sr.ht/~alias/polybase/internal"
)

const defaultDBPath = "/var/lib/polybase/polybase.db"
const version = "0.1.0"

func parseArgs() (string, []string, error) {
	flags := flag.NewFlagSet("polybase", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}

	for i, arg := range os.Args[1:] {
		if arg == "-h" || arg == "help" {
			if err := flags.Parse(os.Args[i+2:]); err != nil {
				printUsage()
				return "", nil, err
			}

			args := flags.Args()

			if err := runHelp(args); err != nil {
				return "", nil, err
			}
			os.Exit(0)
		}
	}

	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "version" {
			fmt.Printf("polybase version %s\n", version)
			os.Exit(0)
		}
	}

	dbPath := flags.String("db", defaultDBPath, "Database path")

	if err := flags.Parse(os.Args[1:]); err != nil {
		printUsage()
		return "", nil, err
	}

	args := flags.Args()
	if len(args) == 0 {
		printUsage()
		return "", nil, nil
	}

	return *dbPath, args, nil
}

func dispatch(pb internal.Polybase, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	ctx := context.Background()
	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "create":
		return runCreate(pb, ctx, cmdArgs)
	case "get":
		return runGet(pb, ctx, cmdArgs)
	case "update":
		return runUpdate(pb, ctx, cmdArgs)
	case "delete":
		return runDelete(pb, ctx, cmdArgs)
	case "list":
		return runList(pb, ctx, cmdArgs)
	case "quantity":
		return runQuantity(pb, ctx, cmdArgs)
	case "visibility":
		return runVisibility(pb, ctx, cmdArgs)
	case "help":
		return runHelp(cmdArgs)
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
