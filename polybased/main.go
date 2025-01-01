package main

import (
	"log"
	"os"

	"git.sr.ht/~alias/polybase/polybased/config"
	"git.sr.ht/~alias/polybase/polybased/routes"
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
