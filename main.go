package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/internal/jet"
	"github.com/saffage/jet/internal/report"
)

func main() {
	// for debug
	spew.Config.Indent = "    "
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true

	defer catchInternalErrors()

	jet.ProcessArgs(os.Args)
}

func catchInternalErrors() {
	if panicErr := recover(); panicErr != nil {
		if err, ok := panicErr.(error); ok {
			report.TaggedErrorf("internal", err.Error())
		} else {
			report.TaggedErrorf("internal", "%v", panicErr)
		}

		// repanic for stack trace
		panic(panicErr)
	}
}
