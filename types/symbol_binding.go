package types

import (
	"github.com/saffage/jet/ast"
)

type Binding struct {
	owner      *Env
	local      *Env
	value      *Value
	decl       *ast.Decl
	node       *ast.LetDecl // May be nil.
	label      *ast.Lower   // May be nil.
	externName string
	params     []*Binding
	isExtern   bool
	isParam    bool
	isField    bool
	isGlobal   bool
}

func NewBinding(owner, local *Env, value *Value, decl *ast.Decl, letNode *ast.LetDecl) *Binding {
	assert(!IsUntyped(value.T), "untyped binding is illegal")
	// assert(decl != nil)

	return &Binding{
		owner: owner,
		local: local,
		value: value,
		decl:  decl,
		node:  letNode,
	}
}

func (sym *Binding) Type() Type { return sym.value.T }
func (sym *Binding) Name() string {
	if sym.decl.Name == nil {
		return "_"
	}
	return sym.Ident().String()
}
func (sym *Binding) Node() ast.Node     { return sym.decl }
func (sym *Binding) Ident() ast.Ident   { return sym.decl.Name }
func (sym *Binding) Owner() *Env        { return sym.owner }
func (sym *Binding) IsParam() bool      { return sym.isParam }
func (sym *Binding) IsField() bool      { return sym.isField }
func (sym *Binding) IsGlobal() bool     { return sym.isGlobal }
func (sym *Binding) IsLocal() bool      { return !sym.isParam && !sym.isField && !sym.isGlobal }
func (sym *Binding) IsExtern() bool     { return sym.isExtern }
func (sym *Binding) ExternName() string { return sym.externName }
func (sym *Binding) Params() []*Binding { return sym.params }
func (sym *Binding) Local() *Env        { return sym.local }

func (sym *Binding) Variadic() Type {
	if fn := As[*Func](sym.value.T); fn != nil {
		return fn.Variadic()
	}
	return nil
}

func (v *Binding) ValueNode() ast.Node {
	if v.node != nil {
		return v.node.Value
	}
	return nil
}

// func (check *checker) resolveFuncDecl(decl *ast.LetDecl, value *ast.Function) {
// 	var (
// 		isDefined                = false
// 		local                    = NewNamedEnv(check.env, "func "+decl.Decl.Name.String())
// 		ty, params, hasResult, _ = check.resolveFuncSignature(value.Signature, local)
// 		sym                      = NewFuncType(check.env, local, ty, decl)
// 	)
// 	sym.params = params
// 	report.TaggedDebugf("checker", "func: set type: %s", ty)

// 	if ty.Result() != nil {
// 		isDefined = true
// 		if defined := check.env.Define(sym); defined != nil {
// 			err := errorAlreadyDefined(sym.Ident(), defined.Ident())
// 			check.problems = append(check.problems, err)
// 		}
// 	}

// 	sym.ty = check.resolveFuncBody(decl, value.Body, ty, local, hasResult)
// 	if sym.ty == nil {
// 		// TODO error message?
// 		sym.ty = ty
// 		report.TaggedDebugf("checker", "func: set type: %s", sym.ty)
// 	}

// 	assert(sym.ty != nil)
// 	assert(sym.ty.Result() != nil)

// 	if !isDefined {
// 		if defined := check.env.Define(sym); defined != nil {
// 			err := errorAlreadyDefined(sym.Ident(), defined.Ident())
// 			check.problems = append(check.problems, err)
// 		}
// 	}

// 	check.resolveFuncAttrs(sym)
// 	check.newDef(decl.Decl.Name, sym)
// }

// func (check *checker) resolveFuncSignature(
// 	sig *ast.Signature,
// 	scope *Env,
// ) (ty *Func, params []*Binding, hasResult, wasError bool) {
// 	params = make([]*Binding, 0, len(sig.Params.Nodes))

// 	tyParams := make([]Type, 0, len(sig.Params.Nodes))
// 	tyResult := Unit
// 	variadic := Type(nil)

// 	for i, node := range sig.Params.Nodes {
// 		param, _ := node.(*ast.LetDecl)
// 		if param == nil {
// 			panic(
// 				fmt.Sprintf("ill-formed AST: unexpected node type '%T'", node),
// 			)
// 		}
// 		if param.Value != nil {
// 			check.errorf(
// 				param,
// 				"parameters with a default value is not supported",
// 			)
// 			wasError = true
// 			continue
// 		}

// 		var tyParam Type

// 		if paramType, _ := param.Decl.Type.(*ast.Op); paramType != nil &&
// 			paramType.Kind == ast.OperatorEllipsis {
// 			if paramType.Y == nil {
// 				variadic = Any
// 			} else if variadic = check.typeOf(paramType.Y); variadic == nil {
// 				wasError = true
// 				continue
// 			}
// 			variadic = SkipTypeDesc(variadic)
// 		} else if tyParam = check.typeOf(param.Decl.Type); tyParam == nil {
// 			wasError = true
// 			continue
// 		}

// 		if variadic != nil {
// 			if i != len(sig.Params.Nodes)-1 {
// 				check.errorf(
// 					param.Decl.Name,
// 					"parameter with ... can only be the last in the list",
// 				)
// 				wasError = true
// 				variadic = nil
// 				// Handle as non-variadic
// 			} else {
// 				// TODO create a symbol for the variadic parameter.
// 				break
// 			}
// 		}

// 		tyParam = SkipTypeDesc(tyParam)
// 		tyParams = append(tyParams, tyParam)

// 		paramSym := NewBinding(scope, tyParam, param.Decl, param)
// 		paramSym.isParam = true

// 		if defined := scope.Define(paramSym); defined != nil {
// 			check.errorf(
// 				param,
// 				"parameter with the same name was already defined",
// 			)
// 			wasError = true
// 			return
// 		}

// 		params = append(params, paramSym)
// 		check.newDef(param.Decl.Name, paramSym)
// 		report.TaggedDebugf("checker", "func: def param: %s", paramSym.Name())
// 		report.TaggedDebugf("checker", "func: set param type: %s", tyParam)
// 	}

// 	if sig.Result != nil {
// 		if tyResultActual := check.typeOf(sig.Result); tyResultActual != nil {
// 			assert(!IsUntyped(SkipTypeDesc(tyResultActual)))
// 			tyResult = WrapInTuple(SkipTypeDesc(tyResultActual))
// 			hasResult = true
// 		} else {
// 			wasError = true
// 		}
// 	}

// 	ty = NewFunc(NewTuple(tyParams...), tyResult, variadic)
// 	return
// }

// func (check *checker) resolveFuncBody(
// 	decl *ast.LetDecl,
// 	body ast.Node,
// 	tyFunc *Func,
// 	scope *Env,
// 	hasResult bool,
// ) *Func {
// 	if body == nil {
// 		if !hasResult {
// 			check.errorf(decl.Decl.Name, "cannot infer a type of the function result")
// 		}
// 		return nil
// 	}

// 	var tyBody *Tuple

// 	// Check the body.
// 	defer check.setEnv(check.env)
// 	check.env = scope

// 	// For a note about recursion
// 	errorsLenBefore := len(check.problems)

// 	if tyBodyActual := check.typeOf(body); tyBodyActual != nil {
// 		tyBody = AsTuple(tyBodyActual)

// 		if tyBody == nil {
// 			tyBody = NewTuple(SkipUntyped(tyBodyActual))
// 		}
// 	} else {
// 		if len(check.problems) > errorsLenBefore {
// 			for i := range len(check.problems) - errorsLenBefore {
// 				err, _ := check.problems[i+errorsLenBefore].(*Problem)
// 				if err != nil {
// 					ident, _ := err.Node.(*ast.Name)
// 					if ident != nil && ident.Data == decl.Decl.Name.String() {
// 						err.AddNote(
// 							decl.Decl.Name,
// 							"cannot infer a type of the recursive definition",
// 						)
// 					}
// 				}
// 			}
// 		}
// 		return nil
// 	}

// 	if !hasResult {
// 		return NewFunc(tyFunc.Params(), tyBody, tyFunc.Variadic())
// 	}

// 	if !tyBody.Equal(tyFunc.Result()) {
// 		var resultNode ast.Node

// 		if list, _ := body.(*ast.Stmts); list != nil {
// 			if len(list.Nodes) == 0 {
// 				resultNode = list
// 			} else {
// 				resultNode = list.Nodes[len(list.Nodes)-1]
// 			}
// 		} else {
// 			resultNode = body
// 		}

// 		check.errorf(
// 			resultNode,
// 			"expected expression of type '%s' for function result, got '%s' instead",
// 			tyFunc.Result(),
// 			tyBody,
// 		)
// 	}

// 	return tyFunc
// }

// func (check *checker) resolveVarDecl(node *ast.LetDecl) {
// 	// 'tValue' can be nil.
// 	tValue, ok := check.resolveVarValue(node.Value)
// 	if !ok {
// 		return
// 	}

// 	// 'tType' cannot be nil.
// 	tType := check.resolveVarType(node.Decl.Type, tValue)
// 	if tType == nil {
// 		return
// 	}

// 	if tValue != nil {
// 		report.TaggedDebugf("checker", "var value type: %s", tValue)
// 	}

// 	report.TaggedDebugf("checker", "var specified type: %s", tType)

// 	if tValue != nil && !tValue.Equal(tType) {
// 		check.errorf(
// 			node.Value,
// 			"type mismatch, expected '%s', got '%s'",
// 			tType,
// 			tValue,
// 		)
// 		return
// 	}

// 	tType = SkipUntyped(tType)

// 	// Set a correct type to the value.
// 	if tValue := AsArray(tValue); tValue != nil && IsUntyped(tValue.ElemType()) {
// 		// TODO this causes codegen to generate two similar typedefs.
// 		check.setType(node.Value, tType)
// 		report.TaggedDebugf("checker", "var set value type: %s", tType)
// 	}

// 	report.TaggedDebugf("checker", "var type: %s", tType)
// 	sym := NewBinding(check.env, tType, node.Decl, node)
// 	sym.isGlobal = sym.owner == check.module.Env

// 	if defined := check.env.Define(sym); defined != nil {
// 		check.problem(errorAlreadyDefined(sym.Ident(), defined.Ident()))
// 		return
// 	}

// 	check.newDef(node.Decl.Name, sym)
// }

// func (check *checker) resolveVarValue(value ast.Node) (Type, bool) {
// 	if value != nil {
// 		t := check.typeOf(value)
// 		if t == nil {
// 			return nil, false
// 		}

// 		if IsTypeDesc(t) {
// 			check.errorf(value, "expected value, got type '%s' instead", t)
// 			return nil, false
// 		}

// 		return t, true
// 	}

// 	return nil, true
// }

// func (check *checker) resolveVarType(typeExpr ast.Node, value Type) Type {
// 	if typeExpr == nil {
// 		return value
// 	}

// 	t := check.typeOf(typeExpr)
// 	if t == nil {
// 		return value
// 	}

// 	typedesc := AsTypeDesc(t)

// 	// Unit can be either value and type.
// 	if t.Equal(Unit) {
// 		typedesc = NewTypeDesc(Unit)
// 	}

// 	if typedesc == nil {
// 		check.errorf(typeExpr, "expression is not a type")
// 		return nil
// 	}

// 	return typedesc.Base()
// }
