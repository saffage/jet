package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type Func struct {
	owner    *Scope
	local    *Scope
	params   []*Var
	t        *types.Func
	node     *ast.FuncDecl
	isExtern bool
}

func NewFunc(owner *Scope, local *Scope, t *types.Func, node *ast.FuncDecl) *Func {
	return &Func{owner, local, nil, t, node, false}
}

func (sym *Func) Owner() *Scope     { return sym.owner }
func (sym *Func) Type() types.Type  { return sym.t }
func (sym *Func) Name() string      { return sym.node.Name.Name }
func (sym *Func) Ident() *ast.Ident { return sym.node.Name }
func (sym *Func) Node() ast.Node    { return sym.node }
func (sym *Func) Local() *Scope     { return sym.local }
func (sym *Func) Params() []*Var    { return sym.params }
func (sym *Func) IsExtern() bool    { return sym.isExtern }

func (check *Checker) resolveFuncDecl(node *ast.FuncDecl) {
	sig := node.Signature
	tParams := []types.Type{}
	params := []*Var{}
	local := NewScope(check.scope)

	for _, param := range sig.Params.Exprs {
		switch param := param.(type) {
		case *ast.Binding:
			t := check.typeOf(param.Type)
			if t == nil {
				return
			}

			t = types.SkipTypeDesc(t)
			tParams = append(tParams, t)

			paramSym := NewVar(local, t, param, param.Name)
			paramSym.isParam = true

			if defined := local.Define(paramSym); defined != nil {
				check.errorf(param, "paramter with the same name was already defined")
				return
			}

			params = append(params, paramSym)
			check.newDef(param.Name, paramSym)
			report.TaggedDebugf("checker", "func: def param: %s", paramSym.Name())
			report.TaggedDebugf("checker", "func: set param type: %s", t)

		case *ast.BindingWithValue:
			check.errorf(param, "parameters can't have a default value")
			return

		default:
			panic(fmt.Sprintf("ill-formed AST: unexpected node type '%T'", param))
		}
	}

	// Result.

	tResult := types.Unit

	if sig.Result != nil {
		t := check.typeOf(sig.Result)
		if t == nil {
			return
		}

		tResult = types.NewTuple(types.SkipTypeDesc(t))
	}

	// Produce function type.

	t := types.NewFunc(tResult, types.NewTuple(tParams...))
	sym := NewFunc(check.scope, local, t, node)
	sym.params = params
	report.TaggedDebugf("checker", "func: set type: %s", t)

	if defined := check.scope.Define(sym); defined != nil {
		err := errorAlreadyDefined(sym.Ident(), defined.Ident())
		check.errors = append(check.errors, err)
		return
	}

	// Define function symbol inside their scope for recursion.
	local.Define(sym)
	check.newDef(node.Name, sym)

	// Body.

	attrExternC := getAttribute(sym, "ExternC")

	if sym.node.Body == nil {
		if attrExternC == nil {
			check.errorf(sym.Ident(), "functions without body is not allowed")
			return
		}

		sym.isExtern = true
		return
	}

	if attrExternC != nil {
		check.errorf(sym.Ident(), "functions with 'ExternC' attribute must have no body")
		return
	}

	defer check.setScope(check.scope)
	check.scope = local

	tBody := check.typeOf(sym.node.Body)
	if tBody == nil {
		return
	}

	if !tResult.Equals(tBody) {
		if len(sym.node.Body.Nodes) == 0 {
			check.errorf(
				sym.node.Body,
				"expected expression of type '%s' for function result, got '%s' instead",
				tResult,
				tBody,
			)
		} else {
			check.errorf(
				sym.node.Body.Nodes[len(sym.node.Body.Nodes)-1],
				"expected expression of type '%s' for function result, got '%s' instead",
				tResult,
				tBody,
			)
		}
		return
	}
}
