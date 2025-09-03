package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/urfave/cli/v2"
)

func Run(args []string) error {
	buildFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:               "parse-ast",
			Usage:              "display program AST of the specified module and exit",
			DisableDefaultText: true,
		},
		&cli.PathFlag{
			Name:  "cc",
			Usage: "C compiler `COMMAND`",
			Value: "cc",
		},
		&cli.StringFlag{
			Name:        "target",
			Usage:       "specify build `TARGET`",
			DefaultText: "c",
			Action: func(ctx *cli.Context, value string) error {
				target, isValidTarget := config.BuildTargetFromString(
					strings.ToLower(value),
				)

				if !isValidTarget {
					return errInvalidEnumValueFor("build target", value, "c")
				}

				config.Target = target
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "cc-flags",
			Usage:       "pass `FLAGS` to a C compiler",
			DefaultText: "",
		},
		&cli.StringFlag{
			Name:        "ld-flags",
			Usage:       "pass `FLAGS` to a linker",
			DefaultText: "",
		},
	}
	checkFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     "usages",
			Category: "actions",
			Usage:    "display `SYMBOL` usages",
			Action:   actionUsages,
		},
		&cli.StringFlag{
			Name:     "def",
			Category: "actions",
			Usage:    "display `SYMBOL` definitions",
			Action:   actionDefinition,
		},
	}
	appFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:               "debug",
			Usage:              "enable compiler debug output",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "trace-parser",
			Usage:              "trace parser calls (used for debugging)",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "no-hints",
			Usage:              "disable compiler hints",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:  "no-core-lib",
			Usage: "disable core library",
		},
		&cli.BoolFlag{
			Name:  "no-builtin-lib",
			Usage: "disable builtin library",
		},
		&cli.PathFlag{
			Name:  "core-lib",
			Usage: "`PATH` to the directory of the language core library",
			Value: filepath.Join(args[0], "lib/core/"),
		},
		&cli.PathFlag{
			Name:  "builtin-lib",
			Usage: "`PATH` to the directory of the language core library",
			Value: filepath.Join(args[0], "lib/builtin/"),
		},
		&cli.BoolFlag{
			Name:  "show-timings",
			Usage: "show time of every compiler stage",
		},
		&cli.BoolFlag{
			Name:  "show-memory-usage",
			Usage: "show memory usage of every compiler stage",
		},
		&cli.StringFlag{
			Name:    "cache-dir",
			Usage:   "specify compiler cache `DIRECTORY`",
			Value:   ".jet-cache",
			EnvVars: []string{"JETCACHE"},
		},
	}
	app := &cli.App{
		Name:    "jet",
		Version: "0.0.1",
		Flags:   appFlags,
		Before:  beforeCommand,
		Commands: []*cli.Command{
			{
				Name:            "build",
				Args:            true,
				ArgsUsage:       " <FILEPATH>",
				HideHelpCommand: true,
				Flags:           buildFlags,
				Action:          actionBuild,
				Before:          beforeBuild,
			},
			{
				Name:      "check",
				Args:      true,
				ArgsUsage: " <FILEPATH>",
				Flags:     checkFlags,
				Action:    actionCheck,
			},
			{
				Name:      "parse",
				Args:      true,
				ArgsUsage: " <FILEPATH>",
				Action:    actionParse,
			},
			{
				Name:   "language-server",
				Action: actionLanguageServer,
			},
		},
	}

	return app.Run(args)
}

func beforeCommand(ctx *cli.Context) error {
	config.Debug = ctx.Bool("debug")
	config.TraceParser = ctx.Bool("trace-parser")
	config.NoHints = ctx.Bool("no-hints")
	config.NoCoreLib = ctx.Bool("no-core")
	config.NoBuiltinLib = ctx.Bool("no-builtin")
	config.ShowTimings = ctx.Bool("show-timings")
	config.ShowMemoryUsage = ctx.Bool("show-memory-usage")
	config.CacheDirName = ctx.String("cache-dir")

	switch {
	case config.Debug:
		report.MinDisplayLevel = report.LevelDebug

	case config.NoHints:
		report.MinDisplayLevel = report.LevelWarning

	default:
		report.MinDisplayLevel = report.LevelHint
	}

	return nil
}

func fileArgument(ctx *cli.Context) (*text.File, error) {
	if !ctx.Args().Present() {
		return nil, errors.New("expected path to a file")
	}

	if ctx.Args().Len() != 1 {
		return nil, errors.New("invalid arguments count (expected 1)")
	}

	argument := ctx.Args().Get(0)
	file, err := config.ReadFile(argument)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func errInvalidEnumValueFor(what, value string, available ...string) error {
	return fmt.Errorf(
		"unknown %s: '%s', available options is: %s",
		what,
		value,
		strings.Join(available, ", "),
	)
}
