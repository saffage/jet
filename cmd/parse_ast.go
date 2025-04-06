package cmd

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
	"github.com/urfave/cli/v2"
)

func ParseAst(file *text.File) error {
	errs := []error{}
	stmts := parser.FromFile(
		file,
		token.DefaultFlags,
		parser.DefaultFlags,
		func(err error) { errs = append(errs, err) },
	).Parse()

	if len(errs) == 0 {
		fmt.Println(ast.Render(stmts))
	}

	return report.Join(errs...)
}

func actionParseAst(ctx *cli.Context) error {
	file, err := fileArgument(ctx)

	if err != nil {
		return err
	}

	return ParseAst(file)
}
