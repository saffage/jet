package main

import (
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/internal/jet"
	"github.com/saffage/jet/internal/log"
)

func main() {
	// for debug
	spew.Config.Indent = "|   "
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true

	defer catchAssetionFail()

	jet.ProcessArgs(os.Args)
}

func catchAssetionFail() {
	if err := recover(); err != nil {
		if err, ok := err.(assert.Fail); ok {
			log.InternalError(err.Msg)
			os.Exit(1)
		}
		panic(err)
	}
}
