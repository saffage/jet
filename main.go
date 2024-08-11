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

	defer func() {
		if err := recover(); err != nil {
			report.TaggedErrorf("internal", "%s", err)
			// repanic for a stack trace

			if config.Global.Flags.Debug {
				panic("internal error")
			}
		}
	}()

	if err := cmd.Run(os.Args); err != nil {
		report.Errors(err)
	}
}
