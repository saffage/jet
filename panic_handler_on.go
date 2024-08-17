//go:build report_panics

package main

import (
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

func handlePanic() {
	if err := recover(); err != nil {
		report.TaggedErrorf("internal", "%s", err)
		// repanic for a stack trace

		if config.Global.Flags.Debug {
			panic("internal error")
		}
	}
}
