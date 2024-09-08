package checker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/types"
)

type Struct struct {
	owner *Scope
	body  *Scope
	t     *types.TypeDesc
	node  *ast.LetDecl
}

func NewStruct(owner *Scope, body *Scope, t *types.TypeDesc, decl *ast.LetDecl) *Struct {
	if !types.IsStruct(t.Base()) {
		panic("expected struct type")
	}
	if body.Parent() != owner {
		panic("invalid local scope parent")
	}
	return &Struct{owner, body, t, decl}
}

func (sym *Struct) Owner() *Scope    { return sym.owner }
func (sym *Struct) Type() types.Type { return sym.t }
func (sym *Struct) Name() string     { return sym.Ident().String() }
func (sym *Struct) Ident() ast.Ident { return sym.node.Decl.Name }
func (sym *Struct) Node() ast.Node   { return sym.node }

func (check *checker) resolveStructDecl(decl *ast.LetDecl, value *ast.StructType) {
	fields := make([]types.StructField, len(value.Fields))
	local := NewScope(check.scope, "struct "+decl.Decl.Name.String())

	for i, fieldDecl := range value.Fields {
		tField := check.typeOf(fieldDecl.Decl.Type)
		if tField == nil {
			return
		}

		if !types.IsTypeDesc(tField) {
			check.errorf(fieldDecl.Decl.Type, "expected field type, got (%s) instead", tField)
			return
		}

		if types.IsUntyped(tField) {
			panic("typedesc cannot have an untyped base")
		}

		t := types.AsTypeDesc(tField).Base()
		fieldSym := NewBinding(local, t, fieldDecl.Decl, fieldDecl)
		fieldSym.isField = true
		fields[i] = types.StructField{
			Name: fieldDecl.Decl.Name.String(),
			Type: t,
		}

		if defined := local.Define(fieldSym); defined != nil {
			err := newErrorf(fieldSym.Ident(), "duplicate field '%s'", fieldSym.Name())
			err.Notes = append(err.Notes, &Error{
				Message: "field was defined here",
				Node:    defined.Ident(),
			})
			check.addError(err)
			continue
		}

		check.newDef(fieldDecl.Decl.Name, fieldSym)
	}

	t := types.NewTypeDesc(types.NewStruct(fields...))
	sym := NewStruct(check.scope, local, t, decl)

	if defined := check.scope.Define(sym); defined != nil {
		check.addError(errorAlreadyDefined(sym.Ident(), defined.Ident()))
		return
	}

	check.newDef(decl.Decl.Name, sym)
}

func (check *checker) structInit(initList *ast.Parens, ty *types.Struct) {
	// tTypeStruct := types.AsStruct(typedesc.Base())
	// if tTypeStruct == nil {
	// 	check.errorf(node.X, "type (%s) is not a struct", typedesc)
	// 	return nil
	// }

	// initList, _ := node.Y.(*ast.CurlyList)
	// if initList == nil {
	// 	check.errorf(node.Y, "expected struct initializer")
	// 	return nil
	// }

	initFields := map[string]types.Type{}
	initFieldValues := map[string]ast.Node{}
	initFieldNames := map[string]*ast.Name{}

	// Collect fields.
	for _, init := range initList.Nodes {
		switch init := init.(type) {
		case *ast.Name:
			panic("field initializer shortcut is not implemented")

		case *ast.Op:
			if init.Kind != ast.OperatorAssign {
				panic("expected assign expression in struct initializer")
			}

			fieldNameNode, _ := init.X.(*ast.Name)
			if fieldNameNode == nil {
				panic("expected identifier for a field name")
			}

			tFieldValue := check.typeOf(init.Y)
			if tFieldValue == nil {
				return
			}

			if _, hasField := initFields[fieldNameNode.Data]; hasField {
				// TODO point to the previous field assignment.
				err := newErrorf(fieldNameNode, "field '%s' is already specified", fieldNameNode.Data)
				check.addError(err)
			} else {
				initFields[fieldNameNode.Data] = tFieldValue
				initFieldValues[fieldNameNode.Data] = init.Y
				initFieldNames[fieldNameNode.Data] = fieldNameNode
			}

			// TODO this doesn't catch all cases, so temporarily remove this

			// if typeSym, ok := check.module.TypeSyms[ty]; ok && typeSym != nil {
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
	for _, field := range ty.Fields() {
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
			initList,
			"missing field '%s' in struct initializer",
			missingFieldNames[0],
		)
	} else if len(missingFieldNames) > 1 {
		check.errorf(
			initList,
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
}

func (check *checker) structMember(selector *ast.Dot, ty *types.Struct) types.Type {
	if ty == types.String {
		check.errorf(selector.X, "member access on string type is not implemented")
		return nil
	}

	fieldIndex := slices.IndexFunc(ty.Fields(), func(field types.StructField) bool {
		return field.Name == selector.Y.Data
	})

	if fieldIndex == -1 {
		check.errorf(selector, "unknown field '%s'", selector.Y.Data)
		return nil
	}

	// if typeSym, ok := check.module.TypeSyms[t]; ok && typeSym != nil {
	// 	structSym, ok := typeSym.(*Struct)
	// 	if !ok || structSym == nil {
	// 		panic("unreachable")
	// 	}
	// 	fieldSym := structSym.body.Member(selector.Name)
	// 	if fieldSym == nil {
	// 		panic("unreachable")
	// 	}
	// 	check.newUse(selector, fieldSym)
	// } else {
	// 	panic("unreachable")
	// }

	return ty.Fields()[fieldIndex].Type
}
