package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/alias-asso/polybase-go/libpolybase"
	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath, args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return dispatch(libpolybase.New(db, "/var/log/polybase/polybase.log", false), args)
}
