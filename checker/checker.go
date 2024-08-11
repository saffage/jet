package checker

import (
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type Checker struct {
	module *Module
	scope  *Scope
	errors []error

	cfg    *config.Config
	fileID config.FileID
}

// Type checks 'expr' and returns its type.
// Also, the value of the expression will also be evaluated
// (if possible) and stored in the 'check.Types' field.
// If error was occured, result is undefined.
func (check *Checker) typeOf(expr ast.Node) types.Type {
	if v := check.valueOf(expr); v != nil {
		return v.Type
	}

	if t := check.typeOfInternal(expr); t != nil {
		check.setType(expr, t)
		return t
	}

	return nil
}

func (check *Checker) valueOf(expr ast.Node) *TypedValue {
	if t, ok := check.module.Types[expr]; ok {
		return t
	}

	if value := check.valueOfInternal(expr); value != nil {
		check.setValue(expr, value)
		return value
	}

	return nil
}

func (check *Checker) setScope(scope *Scope) {
	assert(scope != nil)

	check.scope = scope
}

func (check *Checker) setType(expr ast.Node, t types.Type) {
	assert(expr != nil)
	assert(t != nil)

	if prev := check.module.Types[expr]; prev != nil {
		check.module.Types[expr] = &TypedValue{t, prev.Value}
	} else {
		check.module.Types[expr] = &TypedValue{t, nil}
	}
}

func (check *Checker) setValue(expr ast.Node, value *TypedValue) {
	assert(expr != nil)
	assert(value != nil)
	assert(value.Type != nil)

	check.module.Types[expr] = value
}

func (check *Checker) newDef(ident *ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)

	symStr := ""
	if debugPrinter, _ := sym.(debugPrinter); debugPrinter != nil {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}
	report.TaggedDebugf(
		"checker", "def %s `%s`",
		color.HiBlueString(symStr),
		ident,
	)
	if !check.module.Defs.Set(ident, sym) {
		report.TaggedWarningf("checker", "identifier '%s' was redefined", ident.Name)
	}
	// check.setType(ident, sym.Type())

	switch sym.(type) {
	case *Struct, *Enum:
		check.module.TypeSyms[types.SkipTypeDesc(sym.Type())] = sym
	}
}

func (check *Checker) newUse(ident *ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)

	_, isDef := check.module.Defs.Get(ident)
	assert(!isDef)

	symStr := ""
	if debugPrinter, _ := sym.(debugPrinter); debugPrinter != nil {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}
	report.TaggedDebugf(
		"checker", "use %s `%s` of `%s`",
		color.HiBlueString(symStr),
		ident,
		sym.Ident(),
	)
	check.module.Uses[ident] = sym
}
