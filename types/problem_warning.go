package types

import (
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

type (
	warnDiscardedFuncDef struct {
		name ast.Ident
	}
)

func (warn *warnDiscardedFuncDef) Error() string { return warn.Info().Error() }

func (warn *warnDiscardedFuncDef) Info() *report.Info {
	return &report.Info{
		Kind:  report.KindWarning,
		Title: "unused function",
		Hint:  "this function will never be executed",
		Range: warn.name.Range(),
	}
}
