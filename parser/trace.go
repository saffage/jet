package parser

import (
	"fmt"

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

func (p *Parser) trace(id string) {
	p.printTrace(id)
	p.indent++
}

func (p *Parser) untrace() {
	p.indent--
}
