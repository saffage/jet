package config

import "flag"

// Enable debug information.
var FlagDebug = false

// Display a program AST of the specified Jet module and exit.
var FlagParseAst = false

// Trace parser calls (for debugging).
var FlagTraceParser = false

// Generate C source file from the specified Jet module.
var FlagGenC = false

// Specifies the path to the core library.
var FlagCoreLibPath = ""

// Non-flag command line arguments.
var Args []string

var Exe string

func ParseArgs(args []string) {
	flagSet := flag.NewFlagSet("jet", flag.ExitOnError)

	flagSet.BoolVar(
		&FlagDebug,
		"debug",
		false,
		"Enable debug information",
	)
	flagSet.BoolVar(
		&FlagParseAst,
		"parse_ast",
		false,
		"Display a module AST and exit",
	)
	flagSet.BoolVar(
		&FlagTraceParser,
		"trace_parser",
		false,
		"Trace parser calls (for debugging)",
	)
	flagSet.BoolVar(
		&FlagGenC,
		"gen_c",
		false,
		"Generate C source file from the specified Jet module",
	)
	flagSet.StringVar(
		&FlagCoreLibPath,
		"lib_path",
		"",
		"Specifies the path to the core library",
	)

	if err := flagSet.Parse(args[1:]); err != nil {
		// Must be unreachable due to specified error handling.
		panic(err)
	}

	Args = flagSet.Args()
	Exe = args[0]
}
