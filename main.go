package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/internal/jet"
	"github.com/saffage/jet/internal/log"
)

func main() {
	// for debug
	spew.Config.Indent = "|   "
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true

	defer catchInternalErrors()

	jet.ProcessArgs(os.Args)
}

func catchInternalErrors() {
	if panicErr := recover(); panicErr != nil {
		if err, ok := panicErr.(error); ok {
			log.InternalError(err.Error())
		} else {
			log.InternalError("%v", panicErr)
		}

		// for stack trace
		panic(panicErr)
	}
}
