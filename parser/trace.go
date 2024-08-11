package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/report"
)

func printTrace(p *parser, args ...any) {
	pos := p.tok.Start

	if report.UseColors {
		fmt.Print(color.HiCyanString("%6d:%4d: ", pos.Line, pos.Char))
	} else {
		fmt.Printf("%6d:%4d: ", pos.Line, pos.Char)
	}

	for i := 0; i < p.indent; i++ {
		fmt.Print("  ")
	}

	if report.UseColors {
		fmt.Print(color.HiGreenString("- "))
		fmt.Println(color.YellowString(fmt.Sprint(args...)))
	} else {
		fmt.Printf("- %s", fmt.Sprint(args...))
	}
}

func trace(p *parser) *parser {
	caller := "unknown caller"

	if pc, _, _, ok := runtime.Caller(1); ok {
		if details := runtime.FuncForPC(pc); details != nil {
			// Remove type arguments.
			caller = strings.TrimSuffix(details.Name(), "[...]")

			i := strings.LastIndex(caller, "parse")
			dot := strings.LastIndex(caller, ".")

			if i != -1 && i == dot+1 {
				caller = caller[i+len("parse"):]
			} else {
				caller = caller[dot+1:]
			}

			if strings.HasPrefix(strings.ToLower(caller), "error") ||
				strings.HasSuffix(strings.ToLower(caller), "error") {
				caller = color.RedString("error")
			}
		}
	}

	printTrace(p, caller)
	p.indent++
	return p
}

func un(p *parser) {
	p.indent--
}
