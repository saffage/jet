package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type Func struct {
	owner      *Scope
	local      *Scope
	params     []*Var
	ty         *types.Func
	decl       *ast.Decl
	body       ast.Node
	isExtern   bool
	externName string
}

func NewFunc(owner *Scope, local *Scope, t *types.Func, decl *ast.Decl) *Func {
	return &Func{owner, local, nil, t, decl, nil, false, ""}
}

func (sym *Func) Owner() *Scope        { return sym.owner }
func (sym *Func) Type() types.Type     { return sym.ty }
func (sym *Func) Name() string         { return sym.decl.Ident.Name }
func (sym *Func) Ident() *ast.Ident    { return sym.decl.Ident }
func (sym *Func) Node() ast.Node       { return sym.decl }
func (sym *Func) Local() *Scope        { return sym.local }
func (sym *Func) Params() []*Var       { return sym.params }
func (sym *Func) IsExtern() bool       { return sym.isExtern }
func (sym *Func) ExternName() string   { return sym.externName }
func (sym *Func) Variadic() types.Type { return sym.ty.Variadic() }

func (check *Checker) resolveFuncDecl(decl *ast.Decl, value *ast.Function) {
	var (
		isDefined                = false
		local                    = NewScope(check.scope, "func "+decl.Ident.Name)
		ty, params, hasResult, _ = check.resolveFuncSignature(value.Signature, local)
		sym                      = NewFunc(check.scope, local, ty, decl)
	)
	sym.params = params
	sym.body = value.Body
	report.TaggedDebugf("checker", "func: set type: %s", ty)

	if ty.Result() != nil {
		isDefined = true
		if defined := check.scope.Define(sym); defined != nil {
			err := errorAlreadyDefined(sym.Ident(), defined.Ident())
			check.errors = append(check.errors, err)
		}
	}

	sym.ty = check.resolveFuncBody(decl, value.Body, ty, local, hasResult)
	if sym.ty == nil {
		// TODO error message?
		sym.ty = ty
		report.TaggedDebugf("checker", "func: set type: %s", sym.ty)
	}

	assert(sym.ty != nil)
	assert(sym.ty.Result() != nil)

	if !isDefined {
		if defined := check.scope.Define(sym); defined != nil {
			err := errorAlreadyDefined(sym.Ident(), defined.Ident())
			check.errors = append(check.errors, err)
		}
	}

	check.resolveFuncAttrs(sym)
	check.newDef(decl.Ident, sym)
}

func (check *Checker) resolveFuncSignature(
	sig *ast.Signature,
	scope *Scope,
) (ty *types.Func, params []*Var, hasResult, wasError bool) {
	params = make([]*Var, 0, len(sig.Params.Nodes))

	tyParams := make([]types.Type, 0, len(sig.Params.Nodes))
	tyResult := types.Unit
	variadic := types.Type(nil)

	for i, node := range sig.Params.Nodes {
		param, _ := node.(*ast.Decl)
		if param == nil {
			panic(
				fmt.Sprintf("ill-formed AST: unexpected node type '%T'", node),
			)
		}
		if param.Value != nil {
			check.errorf(
				param,
				"parameters with a default value is not supported",
			)
			wasError = true
			continue
		}

		var tyParam types.Type

		if paramType, _ := param.Type.(*ast.Op); paramType != nil &&
			paramType.Kind == ast.OperatorEllipsis {
			if paramType.Y == nil {
				variadic = types.Any
			} else if variadic = check.typeOf(paramType.Y); variadic == nil {
				wasError = true
				continue
			}
			variadic = types.SkipTypeDesc(variadic)
		} else if tyParam = check.typeOf(param.Type); tyParam == nil {
			wasError = true
			continue
		}

		if variadic != nil {
			if i != len(sig.Params.Nodes)-1 {
				check.errorf(
					param.Ident,
					"parameter with ... can only be the last in the list",
				)
				wasError = true
				variadic = nil
				// Handle as non-variadic
			} else {
				// TODO create a symbol for the variadic parameter.
				break
			}
		}

		tyParam = types.SkipTypeDesc(tyParam)
		tyParams = append(tyParams, tyParam)

		paramSym := NewVar(scope, tyParam, param)
		paramSym.isParam = true

		if defined := scope.Define(paramSym); defined != nil {
			check.errorf(
				param,
				"parameter with the same name was already defined",
			)
			wasError = true
			return
		}

		params = append(params, paramSym)
		check.newDef(param.Ident, paramSym)
		report.TaggedDebugf("checker", "func: def param: %s", paramSym.Name())
		report.TaggedDebugf("checker", "func: set param type: %s", tyParam)
	}

	if sig.Result != nil {
		if tyResultActual := check.typeOf(sig.Result); tyResultActual != nil {
			assert(!types.IsUntyped(types.SkipTypeDesc(tyResultActual)))
			tyResult = types.WrapInTuple(types.SkipTypeDesc(tyResultActual))
			hasResult = true
		} else {
			wasError = true
		}
	}

	ty = types.NewFunc(types.NewTuple(tyParams...), tyResult, variadic)
	return
}

func (check *Checker) resolveFuncBody(
	decl *ast.Decl,
	body ast.Node,
	tyFunc *types.Func,
	scope *Scope,
	hasResult bool,
) *types.Func {
	if body == nil {
		if !hasResult {
			check.errorf(decl.Ident, "cannot infer a type of the function result")
		}
		return nil
	}

	var tyBody *types.Tuple

	// Check the body.
	defer check.setScope(check.scope)
	check.scope = scope

	// For a note about recursion
	errorsLenBefore := len(check.errors)

	if tyBodyActual := check.typeOf(body); tyBodyActual != nil {
		tyBody = types.AsTuple(tyBodyActual)

		if tyBody == nil {
			tyBody = types.NewTuple(types.SkipUntyped(tyBodyActual))
		}
	} else {
		if len(check.errors) > errorsLenBefore {
			for i := range len(check.errors) - errorsLenBefore {
				err, _ := check.errors[i+errorsLenBefore].(*Error)
				if err != nil {
					ident, _ := err.Node.(*ast.Ident)
					if ident != nil && ident.Name == decl.Ident.Name {
						err.Notes = append(err.Notes, &Error{
							Message: "cannot infer a type of the recursive definition",
							Node:    decl.Ident,
						})
					}
				}
			}
		}
		return nil
	}

	if !hasResult {
		return types.NewFunc(tyFunc.Params(), tyBody, tyFunc.Variadic())
	}

	if !tyBody.Equals(tyFunc.Result()) {
		var resultNode ast.Node

		if list, _ := body.(*ast.CurlyList); list != nil {
			if len(list.Nodes) == 0 {
				resultNode = list
			} else {
				resultNode = list.Nodes[len(list.Nodes)-1]
			}
		} else {
			resultNode = body
		}

		check.errorf(
			resultNode,
			"expected expression of type '%s' for function result, got '%s' instead",
			tyFunc.Result(),
			tyBody,
		)
	}

	return tyFunc
}
