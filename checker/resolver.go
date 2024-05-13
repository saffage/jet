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
	sym.isGlobal = sym.owner == check.module.Scope

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Binding.Name, sym)
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

	// Unit can be either value and type.
	if t.Equals(types.Unit) {
		typedesc = types.NewTypeDesc(types.Unit)
	}

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
			paramSym.isParam = true

			if defined := local.Define(paramSym); defined != nil {
				check.errorf(param, "paramter with the same name was already defined")
				return
			}

			check.newDef(param.Name, paramSym)
			fmt.Printf(">>> set `%s` type `%s`\n", paramSym.Name(), t)

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
	check.newDef(node.Name, sym)

	// Body.

	if sym.node.Body == nil {
		check.errorf(sym.Ident(), "functions without body is not allowed")
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

func (check *Checker) resolveStructDecl(node *ast.StructDecl) {
	fields := make(map[string]types.Type, len(node.Body.Nodes))
	local := NewScope(check.scope)

	if node.Body == nil {
		panic("struct body cannot be nil")
	}

	for _, bodyNode := range node.Body.Nodes {
		binding, _ := bodyNode.(*ast.Binding)
		if binding == nil {
			check.errorf(binding, "expected field declaration")
			return
		}

		tField := check.typeOf(binding.Type)
		if tField == nil {
			return
		}

		if !types.IsTypeDesc(tField) {
			check.errorf(binding.Type, "expected field type, got (%s) instead", tField)
			return
		}

		if types.IsUntyped(tField) {
			panic("typedesc cannot have an untyped base")
		}

		t := types.AsTypeDesc(tField).Base()
		fieldSym := NewVar(local, t, binding, binding.Name)
		fieldSym.isField = true
		fields[binding.Name.Name] = t

		if defined := local.Define(fieldSym); defined != nil {
			err := NewErrorf(fieldSym.Ident(), "duplicate field '%s'", fieldSym.Name())
			err.Notes = []*Error{NewError(defined.Ident(), "field was defined here")}
			check.addError(err)
			continue
		}

		check.newDef(binding.Name, fieldSym)
	}

	t := types.NewTypeDesc(types.NewStruct(fields))
	sym := NewStruct(check.scope, local, t, node)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Name, sym)
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
}
