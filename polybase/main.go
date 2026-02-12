package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os/user"
	"strings"

	"github.com/alias-asso/polybase-go/libpolybase"
	_ "modernc.org/sqlite"
)

const (
	version       = "0.1.0"
	defaultDBPath = "/var/lib/polybase/polybase.db"
)

// Global args
var (
	showHelp    = false
	showVersion = false
	dbPath      = defaultDBPath
)

func init() {
	flag.BoolVar(&showHelp, "h", showHelp, "display the help")
	flag.BoolVar(&showHelp, "help", showHelp, "display the help")
	flag.BoolVar(&showVersion, "v", showVersion, "display the version of polybase")
	flag.StringVar(&dbPath, "db", dbPath, "path of the database")
}

func main() {
	flag.Parse()
	if showHelp {
		printUsage()
		return
	}
	if showVersion {
		printVersion()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		return
	}
	if args[0] == "version" {
		printVersion()
		return
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}

	err = dispatch(libpolybase.New(db, "/var/log/polybase/polybase.log", false), flag.Args())
	if err != nil {
		panic(err)
	}
}

var (
	ErrNoCommand      = errors.New("no command specified")
	ErrUnknownCommand = errors.New("unknown command")
)

func dispatch(pb libpolybase.Polybase, args []string) error {
	if len(args) == 0 {
		return ErrNoCommand
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
	default:
		printUsage()
		return errors.Join(ErrUnknownCommand, fmt.Errorf("command %s not supported", cmd))
	}
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
