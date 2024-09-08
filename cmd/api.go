package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
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
		&cli.PathFlag{
			Name:  "cc",
			Usage: "path to a C compiler executable",
			Value: "gcc",
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
				ArgsUsage:       " <file path>",
				HideHelpCommand: true,
				Flags:           buildFlags,
				Action:          actionBuild,
				Before:          beforeBuild,
			},
			{
				Name:            "check",
				Usage:           "perform check of the specified file",
				ArgsUsage:       " <file path>",
				Args:            true,
				HideHelpCommand: true,
				Action:          actionCheck,
			},
			{
				Name:            "parse-ast",
				Usage:           "parses program AST and prints it in YAML format",
				ArgsUsage:       " <file path>",
				Args:            true,
				HideHelpCommand: true,
				Action:          actionParseAst,
				Before:          beforeParseAst,
			},
		},
	}

	config.Global.MaxErrors = 3
	config.Global.Files = map[config.FileID]config.FileInfo{}
	return app.Run(args)
}

func beforeCommand(ctx *cli.Context) error {
	config.Global.Flags.Debug = ctx.Bool("debug")
	config.Global.Flags.NoHints = ctx.Bool("no-hints")
	config.Global.Flags.NoCoreLib = ctx.Bool("no-core-lib")
	config.Global.Options.CoreLibPath = ctx.Path("core-lib-path")
	config.Global.Options.CacheDir = ctx.String("cache-dir")

	switch {
	case config.Global.Flags.Debug:
		report.Level = report.KindDebug

	case config.Global.Flags.NoHints:
		report.Level = report.KindWarning

	default:
		report.Level = report.KindHint
	}

	return nil
}

func readFileToConfig(ctx *cli.Context, cfg *config.Config, fileID config.FileID) error {
	if !ctx.Args().Present() {
		return errors.New("expected path to a file")
	}

	if ctx.Args().Len() != 1 {
		return errors.New("invalid arguments count (expected 1)")
	}

	path := filepath.Clean(ctx.Args().Get(0))
	name, data, err := readFile(path)
	if err != nil {
		return err
	}

	cfg.Files[fileID] = config.FileInfo{
		Name: name,
		Path: path,
		Buf:  bytes.NewBuffer(data),
	}
	return nil
}

func readFile(path string) (name string, data []byte, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	if !stat.Mode().IsRegular() {
		err = fmt.Errorf("'%s' is not a file", path)
		return
	}

	fileExt := filepath.Ext(path)

	if fileExt != ".jet" {
		err = fmt.Errorf("expected file extension '.jet', got '%s' instead", fileExt)
		return
	}

	name = filepath.Base(path[:len(path)-len(fileExt)])
	if _, err = token.IsValidIdent(name); err != nil {
		err = errors.Join(fmt.Errorf("filename is not a valid identifier"), err)
		return
	}

	data, err = os.ReadFile(path)
	if err != nil {
		err = errors.Join(fmt.Errorf("while reading file '%s'", path), err)
		return
	}

	return
}
