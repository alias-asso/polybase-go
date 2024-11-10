package main

import (
	"fmt"
	"log"
	"os"

	"git.sr.ht/~alias/polybase/cmd/polybase-web/config"
	"git.sr.ht/~alias/polybase/cmd/polybase-web/routes"
)

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	if args.ShowHelp {
		printHelp()
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

	fmt.Println(config)

	srv, err := routes.NewServer(&config)
	if err != nil {
		log.Printf("Could not create server %s", err)
	}
	srv.Run()
}
