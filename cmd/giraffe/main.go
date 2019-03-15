package main

import (
	"flag"
	"fmt"
	"os"
)

func printUsage() {
	fmt.Printf("" +
		"Usage:  giraffe <command> [<args>]\n" +
		"where command is one of:\n" +
		"  call  Generate a call graph\n" +
		"  help  Show this list of commands\n")
}

func run() error {
	if len(os.Args) == 1 {
		printUsage()
		return nil
	}
	// help
	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
	// call
	callCmd := flag.NewFlagSet("call", flag.ExitOnError)
	callFormatFlag := callCmd.String("format", "png", "output format")
	callDebugFlag := callCmd.Bool("debug", false, "enable debugging output")
	// Select command
	switch os.Args[1] {
	case "help":
		helpCmd.Parse(os.Args[2:])
	case "call":
		callCmd.Parse(os.Args[2:])
	default:
		return fmt.Errorf(
			"'%s' is not a giraffe command.  See 'giraffe help'.",
			os.Args[1])
	}
	if helpCmd.Parsed() {
		printUsage()
	}
	if callCmd.Parsed() {
		err := runCall(callFormatFlag, callDebugFlag)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
