//go:build report_panics

package main

import (
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
)

func handlePanic() {
	if err := recover(); err != nil {
		if err, _ := err.(error); err != nil {
			report.Error(err)
		}

		if config.Global.Flags.Debug {
			panic("internal error")
		}
	}
}
