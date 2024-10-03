package cmd

import (
	"github.com/saffage/jet/types"
	"github.com/saffage/jet/config"
	"github.com/urfave/cli/v2"
)

func Check(cfg *config.Config) error {
	_, err := types.CheckFile(cfg, config.MainFileID)
	return err
}

func actionCheck(cfg *config.Config) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		err := readFileToConfig(ctx, cfg, config.MainFileID)
		if err != nil {
			return err
		}
		return Check(cfg)
	}
}
