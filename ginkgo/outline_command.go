package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"github.com/onsi/ginkgo/ginkgo/outline"
)

func BuildOutlineCommand() *Command {
	const defaultFormat = "csv"
	var format string
	flagSet := flag.NewFlagSet("outline", flag.ExitOnError)
	flagSet.StringVar(&format, "format", defaultFormat, "Format of outline. Accepted: 'csv', 'json'")
	return &Command{
		Name:         "outline",
		FlagSet:      flagSet,
		UsageCommand: "ginkgo outline <filename>",
		Usage: []string{
			"Outline of Ginkgo symbols for the file",
		},
		Command: func(args []string, additionalArgs []string) {
			outlineFile(args, format)
		},
	}
}

func outlineFile(args []string, format string) {
	if len(args) != 1 {
		println("usage: ginkgo outline <filename>")
		os.Exit(1)
	}

	filename := args[0]
	fset := token.NewFileSet()

	src, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		println(fmt.Sprintf("error parsing source: %s", err))
		os.Exit(1)
	}

	o, err := outline.FromASTFile(src)
	if err != nil {
		println(fmt.Sprintf("error creating outline: %s", err))
		os.Exit(1)
	}

	var oerr error
	switch format {
	case "csv":
		_, oerr = fmt.Print(o)
	case "json":
		b, err := json.Marshal(o)
		if err != nil {
			println(fmt.Sprintf("error marshalling to json: %s", err))
		}
		_, oerr = fmt.Println(string(b))
	default:
		complainAndQuit(fmt.Sprintf("format %s not accepted", format))
	}
	if oerr != nil {
		println(fmt.Sprintf("error writing outline: %s", oerr))
		os.Exit(1)
	}
}
