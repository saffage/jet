package cmd

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/urfave/cli/v2"
)

func Parse(file *text.File) error {
	errs := text.Errors{}
	opts := parser.Options{}
	opts.ErrorHandler = errs.Collect

	stmts := parser.FromFile(file, opts).Parse()

	if len(errs) == 0 {
		fmt.Println(ast.Render(stmts))
	}

	return report.Join(errs...)
}

func actionParse(ctx *cli.Context) error {
	file, err := fileArgument(ctx)

	if err != nil {
		return err
	}

	return Parse(file)
}
