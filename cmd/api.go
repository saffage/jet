package cmd

import (
	"fmt"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/urfave/cli/v2"
)

func Run(args []string) error {
	buildFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:               "run",
			Usage:              "run a compiled executable",
			Aliases:            []string{"r"},
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "parse-ast",
			Usage:              "display program AST of the specified module and exit",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "trace-parser",
			Usage:              "trace parser calls (used for debugging)",
			DisableDefaultText: true,
		},
		&cli.PathFlag{
			Name:  "cc",
			Usage: "path to a C compiler executable `COMMAND`",
			Value: "gcc",
		},
		&cli.StringFlag{
			Name:        "target",
			Usage:       "build `TARGET`",
			DefaultText: "c",
			Action: func(ctx *cli.Context, value string) error {
				switch value {
				case "c":
					config.Target = config.TargetC
					return nil

				default:
					return fmt.Errorf(
						"unknown build target: '%s', available options is: c",
						value,
					)
				}
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
	appFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:               "debug",
			Usage:              "enable compiler debug output",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:               "no-hints",
			Usage:              "disable compiler hints",
			DisableDefaultText: true,
		},
		&cli.BoolFlag{
			Name:  "no-core-lib",
			Usage: "disable the language core library",
		},
		&cli.PathFlag{
			Name:    "core-lib-path",
			Usage:   "path to the directory of the language core library",
			Value:   "./lib",
			EnvVars: []string{"JETLIB"},
		},
		&cli.StringFlag{
			Name:    "cache-dir",
			Usage:   "compiler cache directory",
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
				Name:   "parse-ast",
				Action: actionParseAst,
			},
		},
	}

	return app.Run(args)
}

func beforeCommand(ctx *cli.Context) error {
	config.Debug = ctx.Bool("debug")
	config.NoHints = ctx.Bool("no-hints")
	config.NoBuiltinPackage = ctx.Bool("no-core-lib")
	config.BuiltinPackagePath = ctx.Path("core-lib-path")
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
