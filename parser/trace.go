package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/report"
)

const indentation = ": "

type tracer struct {
	stack      []traceEntry
	enabled    bool
	tokenIndex int
}

type traceEntry struct {
	caller      string
	err         error
	indentation int
}

func (parse *parser) pushTrace(args ...string) bool {
	if parse.tracer.enabled {
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

		pos, _ := parse.scanner.GetPosition(parse.Span.From)
		fmt.Fprintf(
			report.Output,
			"%s%s%s[%s]\n",
			color.HiBlackString("%s", strings.Repeat(indentation, len(parse.tracer.stack))),
			color.HiGreenString("- "),
			color.YellowString("%s %s ", caller, strings.Join(args, ", ")),
			color.HiBlackString(
				"%d: %s (%s)",
				parse.tracer.tokenIndex,
				parse.Kind,
				fmt.Sprintf(
					"%s:%d:%d",
					pos.Path,
					pos.Line,
					pos.Char,
				),
			),
		)

		parse.tracer.stack = append(parse.tracer.stack, traceEntry{
			caller:      caller,
			indentation: len(parse.tracer.stack),
		})
	}

	return parse.tracer.enabled
}

func (parse *parser) popTrace() {
	entry := &parse.tracer.stack[len(parse.tracer.stack)-1]
	parse.tracer.stack = parse.tracer.stack[:len(parse.tracer.stack)-1]

	if entry.err != nil {
		pos, _ := parse.scanner.GetPosition(parse.Span.From)

		fmt.Fprintf(
			report.Output,
			"%s%s%s [%s]\n",
			color.HiBlackString("%s", strings.Repeat(indentation, entry.indentation)),
			color.HiGreenString("- "),
			color.RedString("%s", entry.err),
			color.HiBlackString(
				"%d: %s (%s)",
				parse.tracer.tokenIndex,
				parse.Kind,
				fmt.Sprintf(
					"%s:%d:%d",
					pos.Path,
					pos.Line,
					pos.Char,
				),
			),
		)
	}
}

func (parse *parser) error(err error) {
	if parse.tracer.enabled {
		entry := &parse.tracer.stack[len(parse.tracer.stack)-1]
		entry.err = err
	}

	panic(err)
}
