package types

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/saffage/jet/ast"
	. "github.com/saffage/jet/internal/debug"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
)

type checkerFlags byte

const (
	ForceEval checkerFlags = 1 << iota
	IgnoreBadNodes
)

type checker struct {
	module *Module
	env    *Env
	errors []error
	flags  checkerFlags
}

func (check *checker) symbolOf(ident ast.Ident) Symbol {
	// return check.module.SymbolOf(ident)
	symbol, _ := check.env.Lookup(ident.Name())
	return symbol
}

// Used for better readability like:
//
//	defer check.setEnv(check.env)
//	check.setEnv(someEnv)
func (check *checker) setEnv(scope *Env) {
	Assert(scope != nil)
	check.env = scope
}

func (check *checker) setValue(expr ast.Node, value *Value) {
	Assert(expr != nil)
	Assert(value != nil)
	Assert(value.T != nil)

	check.module.Values[expr] = value
}

func (check *checker) newDef(ident ast.Ident, sym Symbol) {
	Assert(ident != nil)
	Assert(sym != nil)

	if _, ok := ident.(*ast.Placeholder); ok {
		return
	}

	var symStr string

	if debugPrinter, _ := sym.(debugPrinter); debugPrinter != nil {
		symStr = debugPrinter.debug()
	} else {
		symStr = symbolTypeNoQualifier(sym)
	}

	report.Debug("def %s `%s`", color.HiBlueString(symStr), ident)
	check.module.Defs.Put(ident, sym)

	switch sym.(type) {
	case *TypeDef, *TypeAlias:
		report.Debug(
			"set symbol `%s`(%p) as `%[3]s`(%[3]p) type owner",
			sym.Name(),
			sym,
			SkipDescriptor(sym.Type()),
		)
		check.module.TypeOwners[SkipDescriptor(sym.Type())] = sym
	}
}

func (check *checker) newUse(ident ast.Ident, sym Symbol) {
	Assert(ident != nil)
	Assert(sym != nil)

	_, isDef := check.module.Defs.Get(ident)
	Assert(!isDef)

	// TODO CLI command to get all symbols usages
	check.module.Uses[ident] = sym
}

func (check *checker) error(err error) {
	if err != nil {
		if errs, _ := err.(interface{ Unwrap() []error }); errs != nil {
			check.errors = append(check.errors, errs.Unwrap()...)
		} else {
			check.errors = append(check.errors, err)
		}
	}
}

func (check *checker) resolveDeclType(
	expr ast.Node,
) (
	desc Descriptor,
	paramsEnv *Env,
	err error,
) {
	switch expr := expr.(type) {
	case nil:
		// Type expression is not exists.
		return

	case *ast.Signature:
		var t Type

		t, paramsEnv, err = check.resolveSignature(expr)

		if err != nil {
			return
		}

		desc = NewDescriptor(t)

	default:
		var value *Value

		value, err = check.eval(expr)

		if err != nil {
			return
		}

		if !Is[Descriptor](value.T) {
			err = errExprIsNotAType(expr, value.T)
			return
		}

		desc = NewDescriptor(value.T)
	}

	return
}

//
//
//

func (check *checker) resolveTypeDef(node *ast.TypeDef) {
	if node.Args != nil {
		check.error(errUnimplementedFeature(
			"type parameters",
			"",
			node.Args.Range(),
		))
	}

	t := NewCustom(node.Ident.Name(), nil, nil)
	env := NewNamedEnv(TypeEnv, check.env, node.Ident.Name())
	def := NewTypeDef(check.env, env, t, node)

	check.resolveTypeDefBody(def, node.Body)
	report.Debug(
		"resolved: %s\nfields: %+v\nvariants: %+v",
		def.custom.name,
		def.custom.fields,
		def.custom.variants,
	)

	if defined := check.env.Define(def); defined != nil {
		check.error(errAlreadyDefined(
			node.Ident.Range(),
			defined.Ident().Range(),
			"",
			"",
		))
	}

	check.newDef(node.Ident, def)
}

func (check *checker) resolveTypeDefBody(def *TypeDef, body *ast.Block) {
	fields := make([]Field, 0, len(body.Stmts.Items))
	variants := make([]Variant, 0, len(body.Stmts.Items))

	for _, item := range body.Stmts.Items {
		// NOTE: variant index will be set later
		sym := check.resolveTypeDefBodyItem(def, item)

		if sym == nil {
			continue
		}

		if sym.IsField() {
			fields = append(fields, Field{
				Name: sym.Name(),
				T:    sym.Type(),
			})
		} else {
			variants = append(variants, Variant{
				Name:   sym.Name(),
				Params: sym.ParamTypes(),
			})
		}

		if defined := def.local.Define(sym); defined != nil {
			what := "variant"

			if sym.IsField() {
				what = "field"
			}

			check.error(errAlreadyDefined(
				sym.Ident().Range(),
				defined.Ident().Range(),
				what,
				"",
			))
		}
	}

	def.custom = NewCustom(
		def.Name(),
		fields[0:len(fields):len(fields)],
		variants[0:len(variants):len(variants)],
	)
}

func (check *checker) resolveTypeDefBodyItem(
	def *TypeDef,
	item ast.Node,
) (sym *Binding) {
	switch item := item.(type) {
	case *ast.Label:
		decl, isDecl := item.X.(*ast.Decl)

		if !isDecl {
			panic(errIllFormedAst(item))
		}

		return check.resolveTypeField(def, decl, item.Label())

	case *ast.Decl:
		return check.resolveTypeField(def, item, nil)

	case *ast.Variant:
		return check.resolveTypeVariant(def, item)

	default:
		panic(errIllFormedAst(item))
	}
}

func (check *checker) resolveTypeField(
	def *TypeDef,
	node *ast.Decl,
	label *ast.Lower,
) *Binding {
	value, err := check.eval(node.Type)

	if err != nil {
		check.error(err)
		return nil
	}

	if !Is[Descriptor](value.T) {
		check.error(errExprIsNotAType(node.Type, value.T))
	}

	name := node.Ident.Name()

	if _, ok := node.Ident.(*ast.Placeholder); ok {
		name = ""
	}

	field := &Field{Name: name, T: value.T}
	return NewField(def.local, field, node, label)
}

func (check *checker) resolveTypeVariant(def *TypeDef, node *ast.Variant) *Binding {
	if node.Name.Name() == def.Name() {
		check.error(errAlreadyDefined(
			node.Name.Range(),
			def.Ident().Range(),
			"",
			"The variant type cannot be named the same as the type in which it's defined",
		))
		return nil
	}

	variant := &Variant{Name: node.Name.Name(), parent: def.custom}

	var params []*Binding
	var env *Env
	var err error

	if isVariantHasParams(node) {
		variant.Params = make(TypeList, len(node.Params.Nodes))
		params = make([]*Binding, len(node.Params.Nodes))
		env = NewNamedEnv(TypeEnv, def.local, def.Name()+"."+node.Name.Name())

		for i, param := range node.Params.Nodes {
			decl := &ast.Decl{Type: param}

			var paramName *ast.Lower

			if label, _ := param.(*ast.Label); label != nil {
				paramName = label.Label()

				if paramName == nil {
					panic(errIllFormedAst(param))
				}

				decl.Ident = paramName
				param = label.X
			} else {
				if err == nil && i > 0 && params[i-1].labelNode != nil {
					err = errPositionalParamAfterNamed(
						param.Range(),
						params[i-1].Node().Range(),
					)
				}

				decl.Ident = &ast.Placeholder{
					Data: "_" + strconv.Itoa(i),
					Span: node.Params.Nodes[i].Range(),
				}
			}

			value, err := check.eval(param)
			check.error(err)

			sym := NewBinding(env, nil, value, decl, nil)
			sym.labelNode = paramName
			sym.isParam = true

			if defined := env.Define(sym); defined != nil {
				check.error(errAlreadyDefined(
					node.Name.Range(),
					defined.Ident().Range(),
					"",
					"",
				))
			}

			if err == nil {
				params[i] = sym
				variant.Params[i] = value.T
			}
		}
	}

	check.error(err)
	return NewVariant(def.local, env, variant, params, node)
}

func isVariantHasParams(node *ast.Variant) bool {
	return node.Params != nil && len(node.Params.Nodes) > 0
}

//
//
//

func (check *checker) resolveTypeAlias(node *ast.TypeAlias) {
	if node.Args != nil {
		check.error(errUnimplementedFeature(
			"type parameters",
			"",
			node.Args.Range(),
		))
	}

	var sym *TypeAlias

	if extern, _ := node.Expr.(*ast.External); extern != nil {
		var err error

		sym, err = check.resolveExternTypeAlias(extern, node)
		check.error(err)
	} else {
		v, err := check.eval(node.Expr)
		check.error(err)

		if v != nil {
			desc, ok := As[Descriptor](v.T)

			if !ok {
				check.error(errExprIsNotAType(node.Expr, v.T))
			}

			sym = NewTypeAlias(
				check.env,
				NewAlias(desc, node.Ident.Name()),
				node,
			)
		}
	}

	if sym == nil {
		report.Debug("failed to define type alias symbol")
		return
	}

	if defined := check.env.Define(sym); defined != nil {
		check.error(errAlreadyDefined(
			node.Ident.Range(),
			defined.Ident().Range(),
			"",
			"",
		))
	}

	check.newDef(node.Ident, sym)
}

//
//
//

func (check *checker) resolveSignature(sig *ast.Signature) (*Fn, *Env, error) {
	tParams := make([]Type, len(sig.Params.Nodes))
	tResult := Type(nil) // &Parameter{}

	defer check.setEnv(check.env)
	check.env = NewEnv(FnParamsEnv, check.env)

	for i, param := range sig.Params.Nodes {
		label, _ := param.(*ast.Label)

		if label != nil {
			check.error(errUnimplementedFeature(
				"labeled parameters",
				"",
				text.Span{
					From: label.Name.Span.From,
					To:   label.ColonTok,
				},
			))
		}

		decl, ok := param.(*ast.Decl)
		if !ok {
			panic(errIllFormedAst(param))
		}

		sym, err := check.resolveParam(decl, nil)
		check.error(err)
		tParams[i] = sym.Type()

		if prev := check.env.Define(sym); prev != nil {
			check.error(errParamAlreadyDefined(
				decl.Ident.Range(),
				prev.Ident().Range(),
			))
			continue
		}

		check.newDef(decl.Ident, sym)
	}

	var err error

	// Set generic type if no result type is provided
	if sig.Result != nil {
		var value *Value
		value, err = check.eval(sig.Result)

		if err == nil {
			Assert(value.T != nil)
			Assert(Is[Descriptor](value.T), fmt.Sprintf("%[1]T - %[1]s", value.T))

			tResult = SkipDescriptor(value.T)

			Assert(tResult != nil)
		}
	}

	// check.env here is parameter scope
	return NewFn(tParams, tResult, nil), check.env, err
}

func (check *checker) resolveParam(
	param *ast.Decl,
	label *ast.Lower,
) (sym *Binding, err error) {
	Assert(param != nil)

	sym = NewBinding(check.env, nil, &Value{T: nil, V: nil}, param, nil)
	sym.labelNode = label
	sym.isParam = true

	if param.TypeTok.IsValid() {
		err = errInternal(
			param.Range(),
			"type parameters is not implemented",
		)
		return
	}

	if param.Type == nil {
		err = errInternal(
			param.Range(),
			"parameter type inference is not implemented",
		)
		return
	}

	var value *Value
	value, err = check.eval(param.Type)

	if err != nil {
		return
	}

	sym.value.T = SkipDescriptor(value.T)
	sym.value.V = value.V
	return
}

//
//
//

func (check *checker) resolveSelector(node *ast.Dot) (*Value, error) {
	operand, err := check.eval(node.X)

	if err != nil {
		return nil, err
	}

	switch t := SkipAlias(operand.T).(type) {
	case module:
		// TODO implement module symbol lookup
		panic("unimplemented: module symbol lookup")

	case *Custom:
		// TODO implement custom type selector
		panic("unimplemented: custom type selector")

	case Descriptor:
		if custom, ok := SkipAlias(t.base).(*Custom); ok {
			y, ok := node.Y.(*ast.Upper)

			if !ok {
				return nil, errInvalidSelectorNode(
					node,
					Render(t),
					"a variant identifier",
				)
			}

			var selected *Variant

			// Search for y
			for i, variant := range custom.variants {
				if variant.Name == y.Name() {
					selected = custom.Variant(i)
					break
				}
			}

			if selected == nil {
				owner, ok := check.module.TypeOwners[custom].(*TypeDef)

				if !ok || owner == nil {
					panic("unreachable")
				}

				b := errUndefinedIdent(y, owner)
				b = suggestionGuessTypeVariant(b, custom.name, custom.variants)
				b = suggestionDefinedAt(b, "type `"+custom.name+"`", owner.Ident().Range())
				return nil, b
			}

			panic("unimplemented: type construction from variant")
		}

		panic("idk")
		// return nil, errInvalidSelectorType{node: node, t: t}

	default:
		return nil, errInvalidSelectorExprType(node, Render(t))
	}
}

//
//
//

func (check *checker) resolveExternTypeAlias(
	extern *ast.External,
	node *ast.TypeAlias,
) (*TypeAlias, error) {
	if extern.Args != nil {
		panic("unimplemented")
	}

	// for _, arg := range extern.Args.Nodes {
	// 	lit, ok := arg.(*ast.Literal)

	// 	if !ok {
	// 		check.error(nil)
	// 		continue
	// 	}

	// 	if lit.Kind != ast.StringLiteral {
	// 		check.error(nil)
	// 	}
	// }

	externName := node.Ident.Data

	var desc Descriptor

	switch externName {
	case "Int":
		desc = Descriptor{base: IntType}

	case "Float":
		desc = Descriptor{base: FloatType}

	case "String":
		desc = Descriptor{base: StringType}

	default:
		return nil, errUnknownExtern(extern, externName)
	}

	return NewTypeAlias(
		check.env,
		NewAlias(desc, node.Ident.Name()),
		node,
	), nil
}

//
//
//

func (check *checker) resolveCall(node *ast.Call, fn *Fn) (Type, error) {
	Assert(fn != nil)
	Assert(node != nil)
	Assert(node.X != nil)
	Assert(node.Args != nil)

	// tParens, err := check.typeOfParens(node.Args)
	//
	// if err != nil || tParens == nil {
	// 	return nil, err
	// }
	//
	// for i := range tParens {
	// 	if fn.params[i] == nil {
	// 		// unresolved parameter
	// 		continue
	// 	}
	// 	tParens[i] = IntoTyped(tParens[i], fn.params[i])
	// }
	//
	// return fn.Result(), fn.CheckArgs(tParens, node.Args)

	Assert(fn.Result() != nil, fmt.Sprintf("%+v %+v", fn.params, fn.result))

	_, err := check.resolveArgs(node.Args, fn)
	return fn.Result(), err
}

func (check *checker) resolveArgs(args *ast.Parens, fn *Fn) (TypeList, error) {
	if err := check.resolveArgsArity(args, fn); err != nil {
		return nil, err
	}

	tArgs := make(TypeList, len(args.Nodes))
	errs := []error(nil)

	for i := range len(fn.params) {
		expected := fn.params[i]
		actual, err := check.eval(args.Nodes[i], expected)

		if err != nil {
			errs = append(errs, err)
		}

		if actual != nil && !actual.T.Equal(expected) {
			errs = append(errs, errArgTypeMismatch(
				args.Nodes[i],
				actual.T,
				expected,
				i,
				false,
			))
		}
	}

	// Check variadic.
	for i, arg := range args.Nodes[len(fn.params):] {
		i += len(fn.params)
		value, err := check.eval(arg, fn.variadic)

		if err != nil {
			errs = append(errs, err)
		}

		if value.T != nil && !value.T.Equal(fn.variadic) {
			errs = append(errs, errArgTypeMismatch(
				args.Nodes[i],
				value.T,
				fn.variadic,
				i,
				true,
			))
		}
	}

	return tArgs, report.Join(errs...)
}

func (check *checker) resolveArgsArity(args *ast.Parens, fn *Fn) error {
	// Example:
	//
	//	params  args    diff    idx
	//	1       2       -1      1
	//	2       1        1      1
	//	0       3       -3      0
	//	3       0        3      0
	var diff = len(fn.params) - len(args.Nodes)

	if diff < 0 && fn.variadic == nil {
		return errIncorrectArity(args, len(fn.params), len(args.Nodes))
	}

	if diff > 0 {
		return errIncorrectArity(args, len(fn.params), len(args.Nodes))
	}

	return nil
}

func (check *checker) elideDefaultArgs(
	fn *Fn,
	unlabeledArgs []*Value,
	labeledArgs map[*ast.Lower]*Value,
	unlabeledArgNodes []ast.Node,
	labeledArgNodes []*ast.Label,
) error {
	// Signature			asd
	//
	// ()					a
	// (Int)				a
	// (Int, Int)			a
	// (Int, a: Int)		a
	// (a: Int, b: Int)		a

	return nil
}

//
//
//

func (check *checker) resolveLetDecl(node *ast.LetDecl) {
	Assert(node.Decl.Ident != nil, "declaration must have a name")

	sym := NewBinding(check.env, nil, &Value{}, node.Decl, node)
	desc, paramsEnv, err := check.resolveDeclType(node.Decl.Type)
	check.error(err)

	if desc.Base() != nil {
		sym.value.T = desc.Base()
	}

	if paramsEnv != nil {
		// For recursion.
		paramsEnv.Define(sym)
		report.Debug("env %s: %+v", paramsEnv.name, paramsEnv.symbols)
	}

	t, err := check.resolveBindingValue(
		node.Value,
		node.Decl.Type,
		desc.base,
		paramsEnv,
	)
	check.error(err)

	// if tDecl, _ := desc.base.(*Parameter); tDecl != nil {
	// 	t = desc.base
	// }

	// TODO: move it somewhere else
	_, discarded := node.Decl.Ident.(*ast.Placeholder)
	if discarded && Is[*Fn](t) {
		check.error(warnDiscardedFuncDef(node.Decl.Ident.Range()))
	}

	if sym.value.T == nil {
		sym.value.T = t
	}

	check.env.Define(sym)
	check.newDef(node.Decl.Ident, sym)
}

func (check *checker) resolveBindingValue(
	node ast.Node,
	tDeclNode ast.Node,
	tDecl Type,
	local *Env,
) (Type, error) {
	if local != nil {
		defer check.setEnv(check.env)
		check.env = NewEnv(FnEnv, local)

		report.Debug("env path: %s", check.env.Path())

		tDecl = tDecl.(*Fn).result
		tDeclNode = tDeclNode.(*ast.Signature).Result
	}

	if !IsResolved(tDecl) {
		value, err := check.eval(node)

		if err != nil {
			return nil, err
		}

		tValue := IntoTyped(value)

		if !IsResolved(tValue) {
			return nil, errCannotInferTypeOfExpr(node)
		}

		return tValue, nil
	}

	value, err := check.evalExpected(
		node,
		tDeclNode,
		tDecl,
		"Expected because of this type constraint",
	)

	if err != nil {
		return nil, err
	}

	report.Debug("expected: %s, actual: %s", tDecl, value.T)
	return value.T, err
}

//
// Operators
//

func (check *checker) prefix(node *ast.Op, operand *Value) (*Value, error) {
	switch node.Kind {
	case ast.OperatorNot:
		if operand.T.Equal(BoolType) {
			return operand, nil
		}

	case ast.OperatorNeg:
		if operand.T.Equal(IntType) {
			return operand, nil
		}

	default:
		panic(fmt.Sprintf("unknown prefix operator: '%s'", node.Kind))
	}

	return nil, errInternal(
		node.Range(),
		"operator %s is not defined for the type `%s`",
		node.Kind,
		operand,
	)
}

func (check *checker) infix(node *ast.Op, x, y *Value) (*Value, error) {
	// if node.Kind == ast.OperatorAs {
	// 	return check.infixAs(node, tOperandX, tOperandY)
	// }

	// Assignment operation doesn't have a value.
	if node.Kind == ast.OperatorAssign {
		errs := []error{}

		if !check.assignable(node.X) {
			errs = append(
				errs,
				errInternal(node.X.Range(), "expression cannot be assigned to"),
			)
		}

		// TODO invalid type will be inferred if one of them is untyped
		if !y.T.Equal(x.T) && !IntoTyped(y).Equal(IntoTyped(x)) {
			errs = append(errs, errTypeMismatch(
				node.Y,
				node.X,
				"",
				Render(y.T),
				Render(x.T),
			))
			// errorf(node, "type mismatch (%s and %s)", x, y)
		}

		// check.setType(node.Y, x)
		return &Value{T: NoneType}, report.Join(errs...)
	}

	for _, opTypes := range operatorTypes[node.Kind] {
		if x.T.Equal(opTypes.x) && y.T.Equal(opTypes.y) {
			_, isAssignOp := operatorTypesAssign[node.Kind]
			if isAssignOp && !check.assignable(node.X) {
				check.error(errNotAssignable(node.X))
			}

			return &Value{T: opTypes.result}, nil
		}
	}

	// TODO: add help message for possible operator types
	return nil, errInternal(
		node.Range(),
		"type mismatch for operator `%s`, got `%s` and `%s`",
		node.Kind,
		x,
		y,
	)
}

func (check *checker) postfix(node *ast.Op, operand *Value) (*Value, error) {
	panic("unreachable")
}

func (check *checker) assignable(node ast.Node) bool {
	switch operand := node.(type) {
	case *ast.Lower:
		if operand != nil {
			varSym, ok := check.symbolOf(operand).(*Binding)
			if !ok || varSym == nil {
				check.error(errInternal(
					operand.Range(),
					"identifier is not a variable",
				))
				return false
			}

			// report.DebugX("checker", "assign '%s' at '%s'", varSym.Name(), operand)
			check.newUse(operand, varSym)
			return true
		}

	case *ast.Dot:
		if operand != nil {
			// fieldIdent, _ := operand.Selector.(*ast.Ident)
			// if fieldIdent == nil {
			// 	break
			// }

			// fieldSym, ok := check.symbolOf(fieldIdent).(*Var)
			// if !ok || fieldSym == nil {
			// 	check.errorf(fieldIdent, "identifier is not a variable")
			// 	return false
			// }

			// check.newUse(fieldIdent, fieldSym)
			return check.assignable(operand.X)
		}
	}

	return false
}

type operandTypes struct{ x, y, result Type }

// NOTE assignment operator is checked before.
//
// NOTE additional checks for in-place assignment are located in [checker.matchOpTypes].
var operatorTypes = map[ast.OperatorKind][]operandTypes{
	ast.OperatorNot:       {{nil, BoolType, BoolType}},
	ast.OperatorNeg:       {{nil, IntType, IntType}, {nil, FloatType, FloatType}},
	ast.OperatorAdd:       {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorAddAssign: {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorSub:       {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorSubAssign: {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorMul:       {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorMulAssign: {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorDiv:       {{IntType, IntType, IntType}, {FloatType, FloatType, FloatType}},
	ast.OperatorDivAssign: {{IntType, IntType, NoneType}, {FloatType, FloatType, NoneType}},
	ast.OperatorMod:       {{IntType, IntType, IntType}},
	ast.OperatorModAssign: {{IntType, IntType, NoneType}},
	ast.OperatorBitAnd:    {{IntType, IntType, IntType}},
	ast.OperatorBitOr:     {{IntType, IntType, IntType}},
	// ast.OperatorBitXor:       {{IntType, IntType, IntType}},
	// ast.OperatorBitShl:       {{IntType, IntType, IntType}},
	// ast.OperatorBitShr:       {{IntType, IntType, IntType}},
	// ast.OperatorBitAndAssign: {{IntType, IntType, NoneType}},
	// ast.OperatorBitOrAssign:  {{IntType, IntType, NoneType}},
	// ast.OperatorBitXorAssign: {{IntType, IntType, NoneType}},
	// ast.OperatorBitShlAssign: {{IntType, IntType, NoneType}},
	// ast.OperatorBitShrAssign: {{IntType, IntType, NoneType}},
	ast.OperatorEq: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorNe: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorLt: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorLe: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorGt: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	ast.OperatorGe: {{IntType, IntType, BoolType}, {FloatType, FloatType, BoolType}, {BoolType, BoolType, BoolType}},
	// ast.OperatorAnd: {{BoolType, BoolType, BoolType}},
	// ast.OperatorOr:  {{BoolType, BoolType, BoolType}},
}

var operatorTypesAssign = map[ast.OperatorKind]struct{}{
	ast.OperatorModAssign: {},
	// ast.OperatorBitAndAssign: {},
	// ast.OperatorBitOrAssign:  {},
	// ast.OperatorBitXorAssign: {},
	// ast.OperatorBitShlAssign: {},
	// ast.OperatorBitShrAssign: {},
}
