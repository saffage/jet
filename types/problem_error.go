package types

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
)

type (
	errorIllFormedAst struct {
		node ast.Node
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

	errorValueCannotBeStoredAsX struct {
		node  ast.Node
		tNode Type
		tDest Type
	}

	errorNotAssignable struct {
		node ast.Node
	}
)

func (err *errorIllFormedAst) Error() string           { return err.Info().Error() }
func (err *errorUndefinedIdent) Error() string         { return err.Info().Error() }
func (err *errorDefinedAsType) Error() string          { return err.Info().Error() }
func (err *errorAlreadyDefined) Error() string         { return err.Info().Error() }
func (err *errorParamAlreadyDefined) Error() string    { return err.Info().Error() }
func (err *errorUnknownExtern) Error() string          { return err.Info().Error() }
func (err *errorTypeMismatch) Error() string           { return err.Info().Error() }
func (err *errorValueCannotBeStoredAsX) Error() string { return err.Info().Error() }
func (err *errorNotAssignable) Error() string          { return err.Info().Error() }

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

func (err *errorDefinedAsType) Info() *report.Info {
	info := &report.Info{
		Range: err.node.Range(),
		Title: "undeclared identifier",
		Hint:  "this is a type, not a value",
	}

	if err.t != nil && err.t.Range().IsValid() {
		info.Descriptions = append(info.Descriptions, report.Description{
			Range:       err.t.Range(),
			Description: "this type is defined here",
		})
	}

	return info
}

func (err *errorAlreadyDefined) Info() *report.Info {
	info := &report.Info{
		Range: err.name.Range(),
		Title: "name is already defined in the current environment",
	}

	if err.prev != nil && err.prev.Range().IsValid() {
		info.Descriptions = append(info.Descriptions, report.Description{
			Range:       err.prev.Range(),
			Description: "this name is defined here",
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
				Range:       err.prev.Range(),
				Description: "previous parameter was defined here",
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
			Description: "expected because of this type constraint",
			Range:       err.dest.Range(),
		})
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
