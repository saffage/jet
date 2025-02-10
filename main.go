package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/cmd"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

func main() {
	// for debug
	spew.Config.Indent = "    "
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true

	if err := cmd.Run(os.Args); err != nil {
		report.Report(config.Global, err)
		os.Exit(1)
	}
}
