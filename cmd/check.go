package cmd

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
	"github.com/saffage/jet/types"
	"github.com/urfave/cli/v2"
)

var (
	usages      []string
	definitions []string
)

func Check(file *text.File) error {
	_, err := types.CheckFile(file)

	if err != nil {
		return err
	}

	return nil
}

func actionCheck(ctx *cli.Context) error {
	file, err := fileArgument(ctx)

	if err != nil {
		return err
	}

	return Check(file)
}

func actionUsages(ctx *cli.Context, s string) error {
	usages = strings.Split(s, ",")

	for _, usage := range usages {
		_, err := token.CheckIdent(usage)
		if err != nil {
			return fmt.Errorf("can't use '%s' as identifier, %w", s, err)
		}
	}

	return nil
}

func actionDefinition(ctx *cli.Context, s string) error {
	definitions = strings.Split(s, ",")

	for _, usage := range usages {
		_, err := token.CheckIdent(usage)
		if err != nil {
			return fmt.Errorf("can't use '%s' as identifier, %w", s, err)
		}
	}

	return nil
}
