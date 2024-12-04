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

type command struct {
	usage string
	run   func(pb internal.Polybase, ctx context.Context, args []string) error
}

func parseArgs() (string, []string, error) {
	// Quick check for help and version
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "help":
			printUsage()
			os.Exit(0)
		case "-v", "version":
			fmt.Printf("polybase version %s\n", version)
			os.Exit(0)
		}
	}

	// Parse database path
	flags := flag.NewFlagSet("polybase", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}

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
