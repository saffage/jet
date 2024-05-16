package checker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/internal/report"
	"github.com/saffage/jet/types"
)

type Struct struct {
	owner *Scope
	body  *Scope
	t     *types.TypeDesc
	node  *ast.StructDecl
}

func NewStruct(owner *Scope, body *Scope, t *types.TypeDesc, node *ast.StructDecl) *Struct {
	if !types.IsStruct(t.Base()) {
		panic("expected struct type")
	}
	if body.Parent() != owner {
		panic("invalid local scope parent")
	}
	return &Struct{owner, body, t, node}
}

func (sym *Struct) Owner() *Scope     { return sym.owner }
func (sym *Struct) Type() types.Type  { return sym.t }
func (sym *Struct) Name() string      { return sym.node.Name.Name }
func (sym *Struct) Ident() *ast.Ident { return sym.node.Name }
func (sym *Struct) Node() ast.Node    { return sym.node }

func (check *Checker) resolveStructDecl(node *ast.StructDecl) {
	fields := make([]types.StructField, len(node.Body.Nodes))
	local := NewScope(check.scope, "struct "+node.Name.Name)

	if node.Body == nil {
		panic("struct body cannot be nil")
	}

	for i, bodyNode := range node.Body.Nodes {
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
		fields[i] = types.StructField{binding.Name.Name, t}

		if defined := local.Define(fieldSym); defined != nil {
			err := NewErrorf(fieldSym.Ident(), "duplicate field '%s'", fieldSym.Name())
			err.Notes = []*Error{NewError(defined.Ident(), "field was defined here")}
			check.addError(err)
			continue
		}

		check.newDef(binding.Name, fieldSym)
	}

	t := types.NewTypeDesc(types.NewStruct(fields...))
	sym := NewStruct(check.scope, local, t, node)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(node.Name, sym)
}

func (check *Checker) structInit(node *ast.MemberAccess, typedesc *types.TypeDesc) types.Type {
	tTypeStruct := types.AsStruct(typedesc.Base())
	if tTypeStruct == nil {
		check.errorf(node.X, "type (%s) is not a struct", typedesc)
		return nil
	}

	initList, _ := node.Selector.(*ast.CurlyList)
	if initList == nil {
		check.errorf(node.Selector, "expected struct initializer")
		return nil
	}

	initFields := map[string]types.Type{}
	initFieldValues := map[string]ast.Node{}
	initFieldNames := map[string]*ast.Ident{}

	// Collect fields.
	for _, init := range initList.Nodes {
		switch init := init.(type) {
		case *ast.Ident:
			panic("field initializer shortcut is not implemented")

		case *ast.InfixOp:
			if init.Opr.Kind != ast.OperatorAssign {
				panic(fmt.Sprintf(
					"unexpected infix expression '%s' in struct initializer",
					init.Opr,
				))
			}

			fieldNameNode, _ := init.X.(*ast.Ident)
			if fieldNameNode == nil {
				panic("expected identifier for field name")
			}

			tFieldValue := check.typeOf(init.Y)
			if tFieldValue == nil {
				return nil
			}

			if _, hasField := initFields[fieldNameNode.Name]; hasField {
				// TODO point to the previous field assignment.
				err := NewErrorf(fieldNameNode, "field '%s' is already specified", fieldNameNode.Name)
				check.addError(err)
			} else {
				initFields[fieldNameNode.Name] = tFieldValue
				initFieldValues[fieldNameNode.Name] = init.Y
				initFieldNames[fieldNameNode.Name] = fieldNameNode
			}

			// TODO this doesn't catch all cases, so temporarily remove this

			// if typeSym, ok := check.module.TypeSyms[tTypeStruct]; ok && typeSym != nil {
			// 	structSym, ok := typeSym.(*Struct)
			// 	if !ok || structSym == nil {
			// 		panic("unreachable")
			// 	}
			// 	fieldSym := structSym.body.Member(fieldNameNode.Name)
			// 	if fieldSym == nil {
			// 		panic("unreachable")
			// 	}
			// 	check.newUse(fieldNameNode, fieldSym)
			// } else {
			// 	panic("unreachable")
			// }

		default:
			panic(fmt.Sprintf("unexpected node of type '%T' in struct initializer", init))
		}
	}

	missingFieldNames := []string{}

	// Check fields.
	for _, field := range tTypeStruct.Fields() {
		tInit, initialized := initFields[field.Name]

		if !initialized {
			missingFieldNames = append(missingFieldNames, field.Name)
			continue
		}

		if !tInit.Equals(field.Type) {
			check.errorf(
				initFieldValues[field.Name],
				"type mismatch, expected (%s) for field '%s', got (%s) instead",
				field.Type,
				field.Name,
				tInit,
			)
		}

		// Set a correct type to the value.
		if tValue := types.AsArray(tInit); tValue != nil && types.IsUntyped(tValue.ElemType()) {
			// TODO this causes codegen to generate two similar typedefs.
			check.setType(initFieldValues[field.Name], field.Type)
			report.TaggedDebugf("checker", "struct init set value type: %s", field.Type)
		}

		// Delete this field so we can find extra fields later.
		delete(initFields, field.Name)
		delete(initFieldValues, field.Name)
	}

	if len(missingFieldNames) == 1 {
		check.errorf(
			node.Selector,
			"missing field '%s' in struct initializer",
			missingFieldNames[0],
		)
	} else if len(missingFieldNames) > 1 {
		check.errorf(
			node.Selector,
			"missing fields '%s' in struct initializer",
			strings.Join(missingFieldNames, "', '"),
		)
	}

	if len(initFields) > 0 {
		for name := range initFields {
			check.errorf(
				initFieldNames[name],
				"extra field '%s' in struct initializer",
				name,
			)
		}
	}

	// Base can be alias.
	return typedesc.Base()
}

func (check *Checker) structMember(node *ast.MemberAccess, t *types.Struct) types.Type {
	selector, _ := node.Selector.(*ast.Ident)
	if selector == nil {
		check.errorf(node.Selector, "expected field identifier")
		return nil
	}

	fieldIndex := slices.IndexFunc(t.Fields(), func(field types.StructField) bool {
		return field.Name == selector.Name
	})

	if fieldIndex == -1 {
		check.errorf(selector, "unknown field '%s'", selector.Name)
		return nil
	}

	if typeSym, ok := check.module.TypeSyms[t]; ok && typeSym != nil {
		structSym, ok := typeSym.(*Struct)
		if !ok || structSym == nil {
			panic("unreachable")
		}
		fieldSym := structSym.body.Member(selector.Name)
		if fieldSym == nil {
			panic("unreachable")
		}
		check.newUse(selector, fieldSym)
	} else {
		panic("unreachable")
	}

	return t.Fields()[fieldIndex].Type
}
