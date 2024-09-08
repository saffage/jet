package cmd

import (
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/config"
	"github.com/urfave/cli/v2"
)

func Check(cfg *config.Config) error {
	if err := checker.CheckBuiltInPkgs(cfg); err != nil {
		return err
	}

	m, err := checker.CheckFile(cfg, config.MainFileID)
	if err != nil {
		return err
	}

	_ = m
	return nil
}

func actionCheck(ctx *cli.Context) error {
	err := readFileToConfig(ctx, config.Global, config.MainFileID)
	if err != nil {
		return err
	}
	return Check(config.Global)
}
