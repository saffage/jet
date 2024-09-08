package checker

import (
	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/config"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type checker struct {
	module *Module
	scope  *Scope
	errs   []error

	cfg    *config.Config
	fileID config.FileID
}

// Type checks 'expr' and returns it's type.
//
// Also, the value of the expression will also be evaluated (if it's possible)
// and stored in the [TypeInfo.Types] field of the related module.
//
// If error was occured, result is undefined.
func (check *checker) typeOf(expr ast.Node) types.Type {
	if v := check.valueOf(expr); v != nil {
		return v.Type
	}
	if t := check.typeOfInternal(expr); t != nil {
		check.setType(expr, t)
		return t
	}
	return nil
}

func (check *checker) valueOf(expr ast.Node) *TypedValue {
	if t, ok := check.module.Types[expr]; ok {
		return t
	}
	if value := check.valueOfInternal(expr); value != nil {
		check.setValue(expr, value)
		return value
	}
	return nil
}

// Used for better readability like:
//
//	defer check.setScope(check.scope)
//	check.setScope(someScope)
func (check *checker) setScope(scope *Scope) {
	assert(scope != nil)
	check.scope = scope
}

func (check *checker) setType(expr ast.Node, t types.Type) {
	assert(expr != nil)
	assert(t != nil)

	if prev := check.module.Types[expr]; prev != nil {
		check.module.Types[expr] = &TypedValue{t, prev.Value}
	} else {
		check.module.Types[expr] = &TypedValue{t, nil}
	}
}

func (check *checker) setValue(expr ast.Node, value *TypedValue) {
	assert(expr != nil)
	assert(value != nil)
	assert(value.Type != nil)
	check.module.Types[expr] = value
}

func (check *checker) newDef(ident ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)

	var symStr string

	if debugPrinter, _ := sym.(debugPrinter); debugPrinter != nil {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}

	report.TaggedDebugf(
		"checker",
		"def %s `%s`",
		color.HiBlueString(symStr),
		ident,
	)

	if !check.module.Defs.Set(ident, sym) {
		report.TaggedWarningf("checker", "identifier '%s' was redefined", ident.String())
	}

	// check.setType(ident, sym.Type())
	switch sym.(type) {
	case *Struct, *Enum:
		check.module.TypeSyms[types.SkipTypeDesc(sym.Type())] = sym
	}
}

func (check *checker) newUse(ident ast.Ident, sym Symbol) {
	assert(ident != nil)
	assert(sym != nil)
	_, isDef := check.module.Defs.Get(ident)
	assert(!isDef)
	var symStr string
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
