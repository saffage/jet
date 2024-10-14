package types

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/util"
)

type (
	errorIllFormedAst struct {
		node ast.Node
	}

	errorUnimplementedFeature struct {
		rng     token.Range
		node    ast.Node
		feature string
		reason  string
	}

	errorUndefinedIdent struct {
		name ast.Ident
	}

	errorDefinedAsType struct {
		node ast.Ident
		t    ast.Ident
	}

	errorAlreadyDefined struct {
		name ast.Ident
		prev ast.Ident
		hint string
	}

	errorParamAlreadyDefined struct {
		name ast.Ident
		prev ast.Ident
	}

	errorUnknownExtern struct {
		extern     *ast.Extern
		externName string
	}

	errorTypeMismatch struct {
		node  ast.Node
		dest  ast.Node
		tNode Type
		tDest Type
	}

	errorElemTypeMismatch struct {
		elem      ast.Node
		reason    ast.Node
		tElem     Type
		tExpected Type
	}

	errorArgTypeMismatch struct {
		node      ast.Node
		tArg      Type
		tExpected Type
		index     int
		variadic  bool
	}

	errorIncorrectArity struct {
		node     ast.Node
		expected int
		got      int
	}

	errorValueCannotBeStoredAsX struct {
		node  ast.Node
		tNode Type
		tDest Type
	}

	errorNotAssignable struct {
		node ast.Node
	}

	errorPositionalParamAfterNamed struct {
		node  ast.Node
		named ast.Node
	}
)

func (err *errorIllFormedAst) Error() string              { return err.Info().Error() }
func (err *errorUnimplementedFeature) Error() string      { return err.Info().Error() }
func (err *errorUndefinedIdent) Error() string            { return err.Info().Error() }
func (err *errorDefinedAsType) Error() string             { return err.Info().Error() }
func (err *errorAlreadyDefined) Error() string            { return err.Info().Error() }
func (err *errorParamAlreadyDefined) Error() string       { return err.Info().Error() }
func (err *errorUnknownExtern) Error() string             { return err.Info().Error() }
func (err *errorTypeMismatch) Error() string              { return err.Info().Error() }
func (err *errorElemTypeMismatch) Error() string          { return err.Info().Error() }
func (err *errorArgTypeMismatch) Error() string           { return err.Info().Error() }
func (err *errorIncorrectArity) Error() string            { return err.Info().Error() }
func (err *errorValueCannotBeStoredAsX) Error() string    { return err.Info().Error() }
func (err *errorNotAssignable) Error() string             { return err.Info().Error() }
func (err *errorPositionalParamAfterNamed) Error() string { return err.Info().Error() }

func (err *errorIllFormedAst) Info() *report.Info {
	return &report.Info{
		Title: "ill-formed AST",
		Range: err.node.Range(),
	}
}

func (err *errorUndefinedIdent) Info() *report.Info {
	return &report.Info{
		Title: "identifier is undefined",
		Range: err.name.Range(),
	}
}

func (err *errorUnimplementedFeature) Info() *report.Info {
	info := &report.Info{
		Title: err.feature + " is not implemented",
		Range: err.rng,
	}

	if err.node != nil && err.node.Range().IsValid() {
		info.Range = err.node.Range()
	}

	if err.reason != "" {
		info.Descriptions = []report.Description{
			{
				Message: err.reason,
			},
		}
	}

	return info
}

func (err *errorDefinedAsType) Info() *report.Info {
	info := &report.Info{
		Range: err.node.Range(),
		Title: "undeclared identifier",
		Hint:  "this is a type, not a value",
	}

	if err.t != nil && err.t.Range().IsValid() {
		info.Descriptions = append(info.Descriptions, report.Description{
			Message: "this type is defined here",
			Range:   err.t.Range(),
		})
	}

	return info
}

func (err *errorAlreadyDefined) Info() *report.Info {
	info := &report.Info{
		Range: err.name.Range(),
		Title: "name is already defined in the current environment",
		Hint:  err.hint,
	}

	if err.prev != nil && err.prev.Range().IsValid() {
		info.Descriptions = append(info.Descriptions, report.Description{
			Message: "this name is defined here",
			Range:   err.prev.Range(),
		})
	}

	return info
}

func (err *errorParamAlreadyDefined) Info() *report.Info {
	return &report.Info{
		Range: err.name.Range(),
		Title: "parameter with the same name is already defined",
		Descriptions: []report.Description{
			{
				Message: "previous parameter was defined here",
				Range:   err.prev.Range(),
			},
		},
	}
}

func (err *errorUnknownExtern) Info() *report.Info {
	return &report.Info{
		Range: err.extern.Range(),
		Title: "unknown external name",
		Hint:  fmt.Sprintf("external name `%s` is unknown", err.externName),
	}
}

func (err *errorTypeMismatch) Info() *report.Info {
	info := &report.Info{
		Range: err.node.Range(),
		Title: "type mismatch",
		Hint:  fmt.Sprintf("expected `%s` here, not `%s`", err.tDest, err.tNode),
	}

	if err.dest != nil && err.dest.Range().IsValid() {
		info.Descriptions = append(info.Descriptions, report.Description{
			Message: "expected because of this type constraint",
			Range:   err.dest.Range(),
		})
	}

	return info
}

func (err *errorElemTypeMismatch) Info() *report.Info {
	info := &report.Info{
		Range: err.elem.Range(),
		Title: "element type mismatch",
		Hint: fmt.Sprintf(
			"expected `%s` here, got `%s`",
			err.tExpected,
			err.tElem,
		),
	}

	if err.reason != nil && err.reason.Range().IsValid() {
		info.Descriptions = []report.Description{
			{
				Message: "because of this",
				Range:   err.reason.Range(),
			},
		}
	}

	return info
}

func (err *errorArgTypeMismatch) Info() *report.Info {
	info := &report.Info{Title: "argument type mismatch"}

	if err.variadic {
		info.Hint = fmt.Sprintf(
			"expected `%s` for variadic argument, got `%s`",
			err.tExpected,
			err.tArg,
		)
	} else {
		info.Hint = fmt.Sprintf(
			"expected `%s` for %d-%s argument, got `%s`",
			err.tExpected,
			err.index+1,
			util.OrdinalSuffix(err.index+1),
			err.tArg,
		)
	}

	if err.node != nil && err.node.Range().IsValid() {
		info.Range = err.node.Range()
	}

	return info
}

func (err *errorIncorrectArity) Info() *report.Info {
	info := &report.Info{
		Title: "incorrect arity",
		Hint: fmt.Sprintf(
			"expected %d arguments, got %d",
			err.expected,
			err.got,
		),
	}

	if err.node != nil && err.node.Range().IsValid() {
		info.Range = err.node.Range()
	}

	if err.expected > err.got {
		info.Title += ", not enough arguments"
	} else {
		info.Title += ", too many arguments"
	}

	return info
}

func (err *errorValueCannotBeStoredAsX) Info() *report.Info {
	info := &report.Info{
		Range: err.node.Range(),
		Title: "invalid value type",
		Hint:  fmt.Sprintf("expected `%s` here, not `%s`", err.tDest, err.tNode),
	}

	// if err.dest != nil && err.dest.Range().IsValid() {
	// 	info.Descriptions = append(info.Descriptions, report.Description{
	// 		Description: "Expected because of this constraint",
	// 		Range:       err.dest.Range(),
	// 	})
	// }

	return info
}

func (err *errorNotAssignable) Info() *report.Info {
	return &report.Info{
		Range: err.node.Range(),
		Title: "expression cannot be assigned to",
	}
}

func (err *errorPositionalParamAfterNamed) Info() *report.Info {
	info := &report.Info{
		Range: err.node.Range(),
		Title: "positional parameter after named one",
		Hint:  "this parameter must follow before any named parameter",
	}

	if err.named != nil && err.named.Range().IsValid() {
		info.Descriptions = []report.Description{
			{
				Message: "named parameter was here",
				Range:   err.named.Range(),
			},
		}
	}

	return info
}
