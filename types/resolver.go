package types

import (
	"errors"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/debug"
	"github.com/saffage/jet/report"
)

var (
	ErrTypeIsNotResolved = errors.New("type is not resolved")
)

var (
	HintTypeOfExprMustBeKnown = "type of this expression must be known at this point"
	HintTypeOfVarMustBeKnown  = "type of this variable must be known at this point"
)

// Resolves a type of the node passed in.
type resolver struct {
	currentModule *Module     // The nodule where expression comes from.
	currentEnv    *Env        // The env where expression comes from.
	currentValue  *Value      // Value of the expression.
	errorHandler  func(error) // All errors, occurred while resolving expression.
	forceEval     bool        // Forces constant evaluation for expression.
}

// Implements [ast.Visitor].
func (resolve *resolver) IsVisitor() {}

func (resolve *resolver) VisitLower(node *ast.Lower) {
	sym, _ := resolve.currentEnv.Lookup(node.Name())

	if sym == nil {
		resolve.error(errUndefinedIdent(node, nil))
		return
	}

	resolve.currentValue = sym.Value()
}

func (resolve *resolver) VisitUpper(node *ast.Upper) {
	sym, _ := resolve.currentEnv.Lookup(node.Name())

	if sym == nil {
		resolve.error(errUndefinedIdent(node, nil))
		return
	}

	resolve.currentValue = sym.Value()
}

func (resolve *resolver) VisitPlaceholder(node *ast.Placeholder) {
	// NOTE must not be in a pattern.
	panic("todo")

	// sym, _ := resolve.currentEnv.Lookup(node.Name())
	//
	// if sym == nil {
	// 	resolve.error(
	// 		suggestionLocalSymbols(
	// 			errUndefinedIdent(node, nil),
	// 			resolve.currentEnv,
	// 		),
	// 	)
	// 	return
	// }
	//
	// resolve.currentValue = sym.Value()
}

func (resolve *resolver) error(err error) {
	if err != nil && resolve.errorHandler != nil {
		resolve.errorHandler(err)
	}
}

func (resolve *resolver) node(node ast.Node, target Type) (*Value, error) {
	debug.Assert(node != nil)
	debug.Assert(target == nil || IsResolved(target), "target type must be resolved or unknown")

	// TODO: resolve node

	if resolve.currentValue != nil && resolve.currentValue.T == nil {
		resolve.error(
			report.
				Build(ErrTypeIsNotResolved).
				Selection(node.Range(), ""),
		)
	} else if target != nil {
		// TODO: check node's type
	}

	return &Value{T: target}, nil
}

func (resolve *resolver) decl(node *ast.Decl) {
	// value, err := resolve.node(node.Type)

	// if err != nil {
	// 	return
	// }

	// if !Is[Descriptor](value.T) {
	// 	err = errExprIsNotAType(expr, value.T)
	// 	return
	// }

	// desc = NewDescriptor(value.T)
}
func (resolve *resolver) typeDecl()  {}
func (resolve *resolver) variable()  {}
func (resolve *resolver) signature() {}
func (resolve *resolver) params()    {}
func (resolve *resolver) param()     {}
func (resolve *resolver) result()    {}
