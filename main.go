package main

import (
	"os"

	"github.com/saffage/jet/cmd"
	"github.com/saffage/jet/report"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		report.Report(err)
		os.Exit(1)
	}
}
