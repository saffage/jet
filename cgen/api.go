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
	gen.initFunc()

	if mainFn != nil {
		gen.fn(mainFn)
	}

	_, _ = gen.out.WriteString("\n/* TYPES */\n")
	_, _ = gen.out.WriteString(gen.typeSect.String())
	_, _ = gen.out.WriteString("\n/* DECL */\n")
	_, _ = gen.out.WriteString(gen.declVarsSect.String())
	_, _ = gen.out.WriteString("\n")
	_, _ = gen.out.WriteString(gen.declFnsSect.String())
	_, _ = gen.out.WriteString("\n/* CODE */\n")
	_, _ = gen.out.WriteString(gen.codeSect.String())

	if err := gen.out.Flush(); err != nil {
		panic(err)
	}

	return gen.errors
}
