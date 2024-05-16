package cgen

import (
	"bufio"
	"io"

	"github.com/saffage/jet/checker"
)

func Generate(w io.Writer, m *checker.Module) []error {
	gen := &generator{
		Module: m,
		out:    bufio.NewWriter(w),
	}

	gen.out.WriteString(prelude)

	if err := gen.out.Flush(); err != nil {
		panic(err)
	}

	// for node, data := range gen.Data {
	// 	switch data.Type {
	// 	case types.UntypedString:
	// 		if s := constant.AsString(data.Value); s != nil {
	// 			gen.dataSect.WriteString(fmt.Sprintf(
	// 				"static const char str_lit_%d[%d] = %s;\n",
	// 				reflect.ValueOf(node).Pointer(),
	// 				len(*s),
	// 				strconv.Quote(*s),
	// 			))
	// 		}
	// 	}
	// }

	mainFn := gen.defs(gen.Defs, gen.Scope, false)
	gen.codeSect.WriteString(gen.initFunc())

	if mainFn != nil {
		gen.funcDecl(mainFn)
	}

	gen.out.WriteString("\n/* TYPES */\n")
	gen.out.WriteString(gen.typeSect.String())
	// gen.out.WriteString("\n/* DATA */\n")
	// gen.out.WriteString(gen.dataSect.String())
	gen.out.WriteString("\n/* DECL */\n")
	gen.out.WriteString(gen.declVarsSect.String())
	gen.out.WriteString("\n")
	gen.out.WriteString(gen.declFnsSect.String())
	gen.out.WriteString("\n/* CODE */\n")
	gen.out.WriteString(gen.codeSect.String())

	if err := gen.out.Flush(); err != nil {
		panic(err)
	}

	return gen.errors
}
