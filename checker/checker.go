package checker

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

type Checker struct {
	*TypeInfo
	module         *Module
	scope          *Scope
	builtIns       []*BuiltIn
	errors         []error
	isErrorHandled bool
}

type TypeInfo struct {
	// Constant data.
	Data map[ast.Node]*TypedValue

	// Type of every AST node.
	Types map[ast.Node]*TypedValue

	// Type symbols for every type defined in the module.
	TypeSyms map[types.Type]Symbol

	// Every definition in the entire module.
	Defs map[*ast.Ident]Symbol

	// Every definition in the entire module.
	Uses map[*ast.Ident]Symbol

	// Main scope of the module.
	Scope *Scope
}

func Check(node *ast.ModuleDecl) (*TypeInfo, []error) {
	module := NewModule(Global, node)
	check := &Checker{
		TypeInfo: &TypeInfo{
			Data:     make(map[ast.Node]*TypedValue),
			Types:    make(map[ast.Node]*TypedValue),
			TypeSyms: make(map[types.Type]Symbol),
			Defs:     make(map[*ast.Ident]Symbol),
			Uses:     make(map[*ast.Ident]Symbol),
			Scope:    module.scope,
		},
		module:         module,
		scope:          module.scope,
		errors:         make([]error, 0),
		isErrorHandled: true,
	}

	check.defBuiltIns()
	check.defPrimitives()

	{
		nodes := []ast.Node(nil)

		switch body := node.Body.(type) {
		case *ast.List:
			nodes = body.Nodes

		case *ast.CurlyList:
			nodes = body.List.Nodes

		default:
			panic("ill-formed AST")
		}

		for _, node := range nodes {
			ast.WalkTopDown(check.visit, node)
		}

		module.completed = true
	}

	return check.TypeInfo, check.errors
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
	if t, ok := check.Types[expr]; ok {
		return t
	}

	if value := check.valueOfInternal(expr); value != nil {
		check.setValue(expr, *value)
		return value
	}

	return nil
}

func (check *Checker) setScope(scope *Scope) {
	check.scope = scope
}

func (check *Checker) setType(expr ast.Node, t types.Type) {
	assert.Ok(expr != nil)
	assert.Ok(t != nil)

	if check.Types != nil {
		check.Types[expr] = &TypedValue{t, nil}
	}
}

func (check *Checker) setValue(expr ast.Node, value TypedValue) {
	assert.Ok(expr != nil)
	assert.Ok(value.Type != nil)

	if check.Types != nil {
		check.Types[expr] = &value
	}
}

func (check *Checker) newDef(ident *ast.Ident, sym Symbol) {
	assert.Ok(ident != nil)

	if check.Defs != nil {
		symStr := ""
		if debugPrinter, _ := sym.(debugSymbolPrinter); debugPrinter != nil {
			symStr = debugPrinter.debug()
		} else {
			symStr = symbolTypeNoQualifier(sym)
		}
		fmt.Printf(
			">>> def %s `%s`\n",
			color.HiBlueString(symStr),
			ident,
		)
		check.Defs[ident] = sym
	}

	if check.TypeSyms != nil {
		switch sym.(type) {
		case *Struct:
			check.TypeSyms[types.SkipTypeDesc(sym.Type())] = sym
		}
	}
}

func (check *Checker) newUse(ident *ast.Ident, sym Symbol) {
	assert.Ok(ident != nil)
	assert.Ok(sym != nil)

	_, isDef := check.Defs[ident]
	assert.Ok(!isDef)

	if check.Uses != nil {
		symStr := ""
		if debugPrinter, _ := sym.(debugSymbolPrinter); debugPrinter != nil {
			symStr = debugPrinter.debug()
		} else {
			symStr = symbolTypeNoQualifier(sym)
		}
		fmt.Printf(
			">>> use %s `%s` of `%s`\n",
			color.HiBlueString(symStr),
			ident,
			sym.Ident(),
		)
		check.Uses[ident] = sym
	}
}
