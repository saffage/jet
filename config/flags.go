package config

import (
	"flag"
	"os"
)

// Display a program AST of the specified Jet module and exit.
var FlagParseAst = false

// Trace parser calls (for debugging).
var FlagTraceParser = false

// Generate C source file from the specified Jet module.
var FlagGenC = false

// Non-flag command line arguments.
var Args []string

func init() {
	flagSet := flag.NewFlagSet("jet", flag.ExitOnError)

	flagSet.BoolVar(
		&FlagParseAst,
		"parse_ast",
		false,
		"Display a module AST and exit.",
	)
	flagSet.BoolVar(
		&FlagTraceParser,
		"trace_parser",
		false,
		"Trace parser calls (for debugging).",
	)
	flagSet.BoolVar(
		&FlagGenC,
		"gen_c",
		false,
		"Generate C source file from the specified Jet module.",
	)

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		// Must be unreachable due to specified error handling.
		panic(err)
	}

	Args = flagSet.Args()
}
