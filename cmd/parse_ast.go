package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/scanner"
	"github.com/urfave/cli/v2"
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

	stmts, err := parser.Parse(tokens, parser.DefaultFlags)
	if err != nil {
		return err
	}

	out, _ := json.MarshalIndent(stmts, "", "    ")
	fmt.Fprintln(os.Stdout, string(out))
	return nil
}

func actionParseAst(ctx *cli.Context) error {
	err := readFileToConcig(ctx, config.Global, config.MainFileID)
	if err != nil {
		return err
	}
	return ParseAst(config.Global)
}
