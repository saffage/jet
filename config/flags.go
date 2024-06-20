package config

import (
	"errors"
	"flag"
	"strings"
)

var (
	// Enable debug information.
	FlagDebug bool

	// Enable debug information.
	FlagNoHints bool

	// Dump the checker state after checking a module.
	FlagDumpCheckerState bool

	// Display a program AST of the specified Jet module and exit.
	FlagParseAst bool

	// Trace parser calls (for debugging).
	FlagTraceParser bool

	// Generate C source file from the specified Jet module.
	FlagGenC bool

	// Specifies the path to the core library.
	FlagCoreLibPath string

	// Disable core library.
	FlagNoCoreLib bool
)

var Exe string

type Action byte

const (
	ActionShowHelp Action = iota
	ActionError
	ActionCompileToC
	ActionNonCmd
)

func ParseArgs(args []string) ([]string, Action) {
	if len(args) == 0 {
		panic("unreachable")
	}

	Exe = args[0]
	args = args[1:]

	if len(args) == 0 {
		cmdJet.PrintDefaults()
		return nil, ActionShowHelp
	}

	switch strings.ToLower(args[0]) {
	case "c":
		FlagGenC = true
		args = args[1:]
		err := cmdC.Parse(args)
		if len(args) == 0 || errors.Is(err, flag.ErrHelp) {
			cmdC.PrintDefaults()
			return nil, ActionShowHelp
		}
		return cmdC.Args(), ActionCompileToC

	default:
		err := cmdJet.Parse(args)
		if errors.Is(err, flag.ErrHelp) {
			return nil, ActionShowHelp
		}
		return cmdJet.Args(), ActionNonCmd
	}
}

var cmdJet, cmdC *flag.FlagSet

func init() {
	cmdJet = flag.NewFlagSet("Jet Compiler", flag.ExitOnError)
	cmdJet.BoolVar(
		&FlagDebug,
		"debug",
		false,
		"Enable debug information",
	)
	cmdJet.BoolVar(
		&FlagNoHints,
		"no-hints",
		false,
		"Disable the compiler hints. Hints will still be enabled if the -debug flag is set",
	)
	cmdJet.BoolVar(
		&FlagDumpCheckerState,
		"dump-checker-state",
		false,
		"",
	)
	cmdJet.BoolVar(
		&FlagParseAst,
		"parse-ast",
		false,
		"Display a module AST and exit",
	)
	cmdJet.BoolVar(
		&FlagTraceParser,
		"trace-parser",
		false,
		"Trace parser calls (for debugging)",
	)
	cmdJet.BoolVar(
		&FlagGenC,
		"gen-c",
		false,
		"Generate C source file from the specified Jet module",
	)
	cmdJet.StringVar(
		&FlagCoreLibPath,
		"lib-path",
		"",
		"Specifies the path to the core library",
	)
	cmdJet.BoolVar(
		&FlagNoCoreLib,
		"no-core-lib",
		false,
		"",
	)

	cmdC = flag.NewFlagSet("c", flag.ExitOnError)
	cmdC.BoolVar(
		&FlagDebug,
		"debug",
		false,
		"Enable compiler's debug output",
	)
	cmdC.BoolVar(
		&FlagNoHints,
		"no-hints",
		false,
		"Disable the compiler hints. Hints will still be enabled if the -debug flag is set",
	)
	cmdC.BoolVar(
		&CFlagBuild,
		"build",
		false,
		"build the generated C source file using the specified compiler executable (gcc by default)",
	)
	cmdC.BoolVar(
		&CFlagRun,
		"run",
		false,
		"run the generated C source file (implies the --build command)",
	)
	cmdC.StringVar(
		&CFlagCC,
		"cc",
		"gcc",
		"specify C Compiler (gcc by default)",
	)
	cmdC.StringVar(
		&CLDflags,
		"ldflags",
		"",
		"specify linker flags",
	)
	cmdC.StringVar(
		&FlagCoreLibPath,
		"lib-path",
		"",
		"Specifies the path to the core library",
	)
}

// Experimental flags
var (
	CFlagBuild bool
	CFlagRun   bool
	CFlagCC    string
	CLDflags   string
)
