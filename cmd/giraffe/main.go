package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/folio-labs/giraffe/okapilog"
)

func usage() string {
	return "" +
		"Usage:  giraffe <command> [arguments]\n" +
		"where command is one of:\n" +
		"  call  Generate a call graph\n" +
		"  help  Show this list of commands\n" +
		"Use \"giraffe help <command>\" for more information about " +
		"that command.\n"
}

func runCall(inputFlag, outputFlag, formatFlag *string, debugFlag *bool,
	rsTimeFlag *int) error {
	if *inputFlag == "" {
		return fmt.Errorf("Input file not specified")
	}
	if *outputFlag == "" {
		return fmt.Errorf("Output file not specified")
	}
	if *formatFlag != "dot" &&
		*formatFlag != "pdf" &&
		*formatFlag != "png" &&
		*formatFlag != "jpeg" {
		return fmt.Errorf("Format \"%s\" not supported", *formatFlag)
	}
	ifile, err := os.Open(*inputFlag)
	if err != nil {
		return err
	}
	defer ifile.Close()
	olog, err := okapilog.NewLog(ifile)
	if err != nil {
		return err
	}
	cg, err := newCallGraph(olog)
	if err != nil {
		return err
	}
	out := &callOutput{
		graph: []callEdge{},
	}
	cg.prepareOutput(out)
	graph := out.graph
	sortByLineno(graph)
	out.graph = graph
	if *formatFlag == "dot" {
		ofile, err := os.Create(*outputFlag)
		if err != nil {
			return err
		}
		defer ofile.Close()
		write(out, ofile, rsTimeFlag)
	} else {
		cmd := exec.Command("dot", "-T", *formatFlag, "-o", *outputFlag)
		pipein, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer pipein.Close()
			write(out, pipein, rsTimeFlag)
		}()
		dotout, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", dotout)
	}
	return nil
}

func run() error {
	switch {
	case len(os.Args) < 2:
		fallthrough
	case os.Args[1] == "-h":
		fallthrough
	case os.Args[1] == "-help":
		fallthrough
	case os.Args[1] == "--help":
		fmt.Printf("%s", usage())
		return nil
	}
	// help
	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
	// call
	callCmd := flag.NewFlagSet("call", flag.ExitOnError)
	callInputFlag := callCmd.String("i", "", "input file name")
	callOutputFlag := callCmd.String("o", "", "output file name")
	callFormatFlag := callCmd.String("T", "pdf",
		"\"pdf\", \"png\", \"jpeg\", or \"dot\" output format")
	callDebugFlag := callCmd.Bool("debug", false, "enable debugging output")
	callRsTimeFlag := callCmd.Int("rstime", 0,
		"highlight response times >= specified threshold (in "+
			"milliseconds)")
	// Select command
	switch os.Args[1] {
	case "help":
		helpCmd.Parse(os.Args[2:])
	case "call":
		callCmd.Parse(os.Args[2:])
	default:
		return fmt.Errorf("'%s' is not a giraffe command.\n%s",
			os.Args[1], usage())
	}
	if helpCmd.Parsed() {
		if len(helpCmd.Args()) == 0 {
			fmt.Printf("%s", usage())
		} else {
			cmd := helpCmd.Args()[0]
			switch cmd {
			case "help":
				fmt.Printf("The help command has no " +
					"arguments.\n")
			case "call":
				fmt.Printf("Usage of call:\n")
				callCmd.SetOutput(os.Stdout)
				callCmd.PrintDefaults()
			default:
				return fmt.Errorf(
					"unknown command \"%s\".\n\n%s",
					cmd, usage())
			}
		}
	}
	if callCmd.Parsed() {
		err := runCall(callInputFlag, callOutputFlag, callFormatFlag,
			callDebugFlag, callRsTimeFlag)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		}
		os.Exit(1)
	}
}
