package cgen

import (
	"bufio"
	"io"
	"strings"

	"github.com/saffage/jet/checker"
)

type Generator struct {
	*checker.Module

	dataSect     strings.Builder
	typeSect     strings.Builder
	declVarsSect strings.Builder
	declFnsSect  strings.Builder
	codeSect     strings.Builder
	out          *bufio.Writer
	errors       []error
	numIndent    int
}

func Generate(w io.Writer, m *checker.Module) []error {
	gen := &Generator{
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

	var mainFn *checker.Func

	for _, def := range gen.Defs {
		if def.Owner() != gen.Scope {
			continue
		}

		switch sym := def.(type) {
		case *checker.Var:
			gen.Var(sym)

		case *checker.Const:
			gen.Const(sym)

		case *checker.Struct:
			gen.structure(sym)

		case *checker.Func:
			if sym.Name() == "main" {
				// Generate it later.
				mainFn = sym
				break
			}

			gen.Func(sym)
			// gen.codeSect.WriteString(fmt.Sprintf("/* %s: %s */\n\n", sym.Name(), sym.Type()))

		default:
			panic("not implemented")
		}
	}

	if mainFn != nil {
		gen.Func(mainFn)
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

func (gen *Generator) indent(w io.StringWriter) {
	if gen.numIndent > 0 {
		_, err := w.WriteString(strings.Repeat("\t", gen.numIndent))
		if err != nil {
			panic(err)
		}
	}
}

const prelude = `/* GENERATED BY JET COMPILER */

#undef NDEBUG
#include <assert.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <stdbool.h>

typedef int8_t   Ti8;
typedef int16_t  Ti16;
typedef int32_t  Ti32;
typedef int64_t  Ti64;
typedef uint8_t  Tu8;
typedef uint16_t Tu16;
typedef uint32_t Tu32;
typedef uint64_t Tu64;
typedef float    Tf32;
typedef double   Tf64;
typedef uint8_t  Tbool;
`

const fnMainHead = "\nint main(const int argc, const char *const *const argv)"
