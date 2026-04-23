package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/alias-asso/polybase-go/polybased/config"
	"github.com/alias-asso/polybase-go/polybased/routes"
)

const version = "0.1.0"

var (
	showHelp    bool   = false
	showVersion bool   = false
	devMode     bool   = false
	configPath  string = "/etc/polybase/config.cfg"
)

func init() {
	flag.BoolVar(&showHelp, "h", showHelp, "show the help")
	flag.BoolVar(&showHelp, "help", showHelp, "show the help")
	flag.BoolVar(&showVersion, "v", showVersion, "show the version")
	flag.BoolVar(&devMode, "dev", devMode, "enable dev mode")
	flag.StringVar(&configPath, "c", configPath, "set the config path")
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

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv, err := routes.NewServer(&cfg)
	if err != nil {
		log.Printf("Could not create server %s", err)
	}
	srv.Run(config.CreateContext(context.Background(), &cfg, devMode))
}

func printUsage() {
	fmt.Printf(`Usage: polybased [OPTIONS]

Manage polybase database from the web browser.

Options:
  -c <path>   Path to config file (default: /etc/polybase/config.cfg)
  -v          Print version information
  -h          Print this help message
  -dev        Enable dev mode

For bug reporting and more information, please see:
https://github.com/alias-asso/polybase-go
`)
}

func printVersion() {
	fmt.Printf("polybased version %s\n", version)
}
