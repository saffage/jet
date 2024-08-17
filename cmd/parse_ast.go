package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func ParseAst(cfg *config.Config) error {
	fi := cfg.Files[config.MainFileID]
	if fi.Buf == nil {
		return errors.New("empty file buffer")
	}

	tokens, err := scanner.Scan(fi.Buf.Bytes(), config.MainFileID, scanner.DefaultFlags)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "# %+v\n\n", tokens)

	stmts, err := parser.Parse(tokens, parser.DefaultFlags)
	if err != nil {
		return err
	}

	out, err := yaml.Marshal(stmts)

	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(out))
	fmt.Fprintln(os.Stderr, stmts.Repr())

	// var newStmts ast.StmtList
	// if err := yaml.Unmarshal(out, &newStmts); err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%+v\n", newStmts)

	return nil
}

func actionParseAst(ctx *cli.Context) error {
	err := readFileToConcig(ctx, config.Global, config.MainFileID)
	if err != nil {
		return err
	}
	return ParseAst(config.Global)
}

func beforeParseAst(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return errors.New("expected file path")
	}
	path := ctx.Args().Get(0)
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("'%s' is not a file", path)
	}
	return nil
}
