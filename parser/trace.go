package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/report"
)

type tracer struct {
	enabled bool
	stack   []traceEntry
}

type traceEntry struct {
	caller      string
	error       *error
	indentation int
}

const indentation = "⁝ "

func (t *tracer) trace(error *error) *traceEntry {
	if !t.enabled {
		return nil
	}

	caller := "untracked caller"

	if pc, _, _, ok := runtime.Caller(1); ok {
		if details := runtime.FuncForPC(pc); details != nil {
			const parserPrefix = "(*parser)."

			caller = details.Name()
			i := strings.LastIndex(caller, parserPrefix)

			if i >= 0 {
				caller = caller[len(parserPrefix)+i:]
			}
		}
	}

	fmt.Fprintf(
		report.Output,
		"%s%s%s\n",
		strings.Repeat(indentation, len(t.stack)),
		color.HiGreenString("- "),
		color.YellowString(caller),
	)

	t.stack = append(t.stack, traceEntry{
		caller:      caller,
		error:       error,
		indentation: len(t.stack),
	})
	return &t.stack[len(t.stack)-1]
}

func (t *tracer) un(entry *traceEntry) {
	if !t.enabled {
		return
	}

	if entry.error != nil && *entry.error != nil {
		fmt.Fprintf(
			report.Output,
			"%s%s%s\n",
			strings.Repeat(indentation, entry.indentation),
			color.HiGreenString("- "),
			color.RedString("%s", *entry.error),
		)
	}

	t.stack = t.stack[:len(t.stack)-1]
}
