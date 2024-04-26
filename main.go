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
	if err := recover(); err != nil {
		if err, ok := err.(error); ok {
			log.InternalError(err.Error())
		}
		log.InternalError("%v", err)

		// for stack trace
		panic(err)
	}
}
