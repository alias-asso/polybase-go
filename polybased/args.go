package main

import (
	"fmt"
	"os"
	"strings"
)

type Args struct {
	ConfigPath  string
	ShowHelp    bool
	ShowVersion bool
}

func parseArgs() (*Args, error) {
	args := &Args{
		ConfigPath: "/etc/polybase/polybase.cfg",
	}

	// If no arguments provided, return defaults
	if len(os.Args) <= 1 {
		return args, nil
	}

	// First pass: check for help flag
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			printUsage()
			os.Exit(0)
		}
	}

	// Second pass: process other flags
	osArgs := os.Args[1:]
	for i := 0; i < len(osArgs); i++ {
		arg := osArgs[i]

		switch arg {
		case "-v":
			args.ShowVersion = true
			return args, nil

		case "-c":
			if i+1 >= len(osArgs) {
				return nil, fmt.Errorf("error: -c requires a path argument")
			}
			nextArg := osArgs[i+1]
			if strings.HasPrefix(nextArg, "-") {
				return nil, fmt.Errorf("error: -c requires a path argument")
			}
			args.ConfigPath = nextArg
			i++

		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("error: unknown flag: %s", arg)
			}
		}
	}

	return args, nil
}

func printUsage() {
	fmt.Printf(`Usage: polybased [OPTIONS]

Manage polybase database from the web browser.

Options:
  -c <path>   Path to config file (default: /etc/polybase/config.cfg)
  -v          Print version information
  -h          Print this help message

For bug reporting and more information, please see:
https://git.sr.ht/~alias/polybase
`)
}
