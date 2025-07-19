package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/alias-asso/polybase-go/libpolybase"
	_ "github.com/mattn/go-sqlite3"
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
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("invalid database file: %w", err)
	}

	return dispatch(libpolybase.New(db, "/var/log/polybase/polybase.log", false), args)
}
