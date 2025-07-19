package main

import (
	"log"
	"os"

	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/alias-asso/polybase-go/polybased/routes"
)

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	if args.ShowHelp {
		printUsage()
		os.Exit(0)
	}

	if args.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	config, err := config.LoadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv, err := routes.NewServer(&config)
	if err != nil {
		log.Printf("Could not create server %s", err)
	}
	srv.Run()
}
