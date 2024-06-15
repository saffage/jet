package cgen

import (
	"bufio"
	"errors"
	"io"

	"github.com/saffage/jet/checker"
)

func Generate(w io.Writer, m *checker.Module) error {
	gen := &generator{
		Module: m,
		out:    bufio.NewWriter(w),
		scope:  m.Scope,
	}

	gen.out.WriteString(prelude)

	if err := gen.out.Flush(); err != nil {
		panic(err)
	}

	mainFn := gen.defs(gen.Defs, gen.Scope)
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

	return errors.Join(gen.errors...)
}
