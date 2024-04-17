package parser

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/saffage/jet/internal/log" // for log.NoColors
)

func (p *Parser) printTrace(args ...any) {
	pos := p.tok.Start

	if log.NoColors {
		fmt.Printf("%6d:%4d: ", pos.Line, pos.Char)
	} else {
		fmt.Print(color.HiCyanString("%6d:%4d: ", pos.Line, pos.Char))
	}

	for i := 0; i < p.indent; i++ {
		fmt.Print("  ")
	}

	if log.NoColors {
		fmt.Printf("- %s", fmt.Sprint(args...))
	} else {
		fmt.Print(color.HiGreenString("- "))
		fmt.Println(color.YellowString(fmt.Sprint(args...)))
	}
}

func (p *Parser) trace() {
	caller := "unknown caller"

	if pc, _, _, ok := runtime.Caller(1); ok {
		if details := runtime.FuncForPC(pc); details != nil {
			// Remove type argments.
			caller = strings.TrimSuffix(details.Name(), "[...]")

			i := strings.LastIndex(caller, "parse")
			dot := strings.LastIndex(caller, ".")

			if i != -1 && i == dot+1 {
				caller = caller[i+len("parse"):]
			} else {
				caller = caller[dot+1:]
			}

			if caller == "error" {
				caller = color.RedString(caller)
			}
		}
	}

	p.printTrace(caller)
	p.indent++
}

func (p *Parser) untrace() {
	p.indent--
}
