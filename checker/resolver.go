package checker

import (
	"fmt"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/types"
)

func (check *Checker) resolveVarDecl(node *ast.VarDecl) {
	// 'tValue' can be nil.
	tValue, ok := check.resolveVarValue(node.Value)
	if !ok {
		return
	}

	// 'tType' must be not nil.
	tType := check.resolveVarType(node.Binding.Type, tValue)
	if tType == nil {
		return
	}

	if tValue != nil {
		fmt.Printf(">>> var value '%s'\n", tValue)
	}

	fmt.Printf(">>> var type '%s'\n", tType)

	if tValue != nil && !tType.Equals(tValue) {
		check.errorf(
			node.Binding.Name,
			"type mismatch, expected '%s', got '%s'",
			tType,
			tValue)
		return
	}

	fmt.Printf(">>> var actual type '%s'\n", tType)
	sym := NewVar(check.scope, tType, node.Binding, node.Binding.Name)

	if defined := check.scope.Define(sym); defined != nil {
		err := errorAlreadyDefined(sym.Ident(), defined.Ident())
		check.errors = append(check.errors, err)
		return
	}

	check.newDef(node.Binding.Name, sym)
	fmt.Printf(">>> def var `%s`\n", node.Binding.Name)
}

func (check *Checker) resolveVarValue(value ast.Node) (types.Type, bool) {
	if value != nil {
		t := check.typeOf(value)
		if t == nil {
			return nil, false
		}

		if types.IsTypeDesc(t) {
			check.errorf(value, "expected value, got type '%s' instead", t)
			return nil, false
		}

		return types.SkipUntyped(t), true
	}

	return nil, true
}

func (check *Checker) resolveVarType(typeExpr ast.Node, value types.Type) types.Type {
	if typeExpr == nil {
		return value
	}

	t := check.typeOf(typeExpr)
	if t == nil {
		return value
	}

	typedesc := types.AsTypeDesc(t)
	if typedesc == nil {
		check.errorf(typeExpr, "expression is not a type")
		return nil
	}

	return typedesc.Base()
}

func (check *Checker) resolveFuncDecl(node *ast.FuncDecl) {
	sig := node.Signature
	tParams := []types.Type{}
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

			if defined := local.Define(paramSym); defined != nil {
				check.errorf(param, "paramter with the same name was already defined")
				return
			}

			check.newDef(param.Name, paramSym)
			fmt.Printf(">>> set `%s` type `%s`\n", paramSym.Name(), t)
			fmt.Printf(">>> def param `%s`\n", paramSym.Name())

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

	fmt.Printf(">>> set `%s` type `%s`\n", sym.Name(), t)

	if defined := check.scope.Define(sym); defined != nil {
		err := errorAlreadyDefined(sym.Ident(), defined.Ident())
		check.errors = append(check.errors, err)
		return
	}

	// Define function symbol inside their scope for recursion.
	local.Define(sym)

	// Body.

	if sym.node.Body == nil {
		check.errorf(sym.Ident(), "functions without body is not allowed")
		return
	}

	prevScope := check.scope
	check.scope = local

	defer func() {
		check.scope = prevScope
	}()

	tBody := check.typeOf(sym.node.Body)
	if tBody == nil {
		return
	}

	if !tResult.Equals(tBody) {
		check.errorf(
			sym.node.Body.Nodes[len(sym.node.Body.Nodes)-1],
			"expected expression of type '%s' for function result, got '%s' instead",
			tResult,
			tBody,
		)
		return
	}

	check.newDef(node.Name, sym)
	fmt.Printf(">>> def func `%s`\n", node.Name.Name)
}

func (check *Checker) resolveTypeAliasDecl(node *ast.TypeAliasDecl) {
	t := check.typeOf(node.Expr)
	if t == nil {
		return
	}

	typedesc := types.AsTypeDesc(t)

	if typedesc == nil {
		check.errorf(node.Expr, "expression is not a type (%s)", t)
		return
	}

	sym := NewTypeAlias(check.scope, typedesc, node)

	if defined := check.scope.Define(sym); defined != nil {
		err := errorAlreadyDefined(sym.Ident(), defined.Ident())
		check.errors = append(check.errors, err)
		return
	}

	check.newDef(node.Name, sym)
	check.setType(node, typedesc)
	fmt.Printf(">>> def alias `%s`\n", node.Name.Name)
}
