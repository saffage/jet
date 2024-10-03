package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/report"
	"github.com/urfave/cli/v2"
)

func Run(cfg *config.Config, args []string) error {
	buildFlags := []cli.Flag{
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
		Before:  beforeCommand(cfg),
		Commands: []*cli.Command{
			{
				Name:            "build",
				Args:            true,
				ArgsUsage:       " <file>",
				HideHelpCommand: true,
				Flags:           buildFlags,
				Action:          actionBuild(cfg),
				Before:          beforeBuild(cfg),
			},
			{
				Name:            "check",
				Usage:           "perform check of the specified file",
				ArgsUsage:       " <files>",
				Args:            true,
				HideHelpCommand: true,
				Action:          actionCheck(cfg),
			},
			{
				Name:            "dump-ast",
				Usage:           "parse input file(s) and dump AST(s) into YAML file(s)",
				ArgsUsage:       " <files>",
				Args:            true,
				HideHelpCommand: true,
				Action:          actionParseAst(cfg),
				Before:          beforeParseAst(cfg),
			},
		},
	}

	cfg.MaxErrors = 3
	cfg.Files = map[config.FileID]config.FileInfo{}
	return app.Run(args)
}

func beforeCommand(cfg *config.Config) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		cfg.Flags.Debug = ctx.Bool("debug")
		cfg.Flags.NoHints = ctx.Bool("no-hints")
		cfg.Flags.NoCoreLib = ctx.Bool("no-core-lib")
		cfg.Options.CoreLibPath = ctx.Path("core-lib-path")
		cfg.Options.CacheDir = ctx.String("cache-dir")

		switch {
		case cfg.Flags.Debug:
			report.MinDisplayLevel = report.LevelDebug

		case cfg.Flags.NoHints:
			report.MinDisplayLevel = report.LevelWarning

		default:
			report.MinDisplayLevel = report.LevelHint
		}

		return nil
	}
}

func readFileToConfig(
	ctx *cli.Context,
	cfg *config.Config,
	fileID config.FileID,
) error {
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
