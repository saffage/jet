package cgen

import (
	"fmt"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/internal/assert"
	"github.com/saffage/jet/types"
)

const (
	_ErrorNilType  = "ERROR_CGEN__NIL_TYPE"
	_ErrorMetaType = "ERROR_CGEN__META_TYPE"
)

var arrayTypes = map[types.Type]string{}

func (gen *generator) TypeString(ty types.Type) string {
	assert.Ok(!types.IsTypeDesc(ty))

	switch ty := ty.Underlying().(type) {
	case nil:
		return _ErrorNilType

	case *types.TypeDesc:
		return _ErrorMetaType

	case *types.Alias:
		panic("unreachable")

	case *types.Primitive:
		switch ty.Kind() {
		case types.KindUntypedInt, types.KindUntypedFloat, types.KindUntypedString, types.KindAny, types.KindAnyTypeDesc:
			return _ErrorMetaType

		case types.KindUntypedBool, types.KindBool:
			return "Tbool"

		case types.KindI8:
			return "Ti8"

		case types.KindI16:
			return "Ti16"

		case types.KindI32:
			return "Ti32"

		case types.KindI64:
			return "Ti64"

		case types.KindU8:
			return "Tu8"

		case types.KindU16:
			return "Tu16"

		case types.KindU32:
			return "Tu32"

		case types.KindU64:
			return "Tu64"

		case types.KindF32:
			return "Tf32"

		case types.KindF64:
			return "Tf64"

		case types.KindChar:
			return "char"

		case types.KindPointer:
			return "void*"

		default:
			panic("unreachable")
		}

	case *types.Func:
		panic("not implemented")

	case *types.Tuple:
		if ty.Equals(types.Unit) {
			return "void"
		}

		panic("not implemented")

	case *types.Array:
		return gen.arrayType(ty)

	case *types.Ref:
		return gen.TypeString(ty.Base()) + "*"

	case *types.Struct:
		if ty == types.String {
			return "char*"
		}
		return gen.findTypeSym(gen.Defs, ty)

	case *types.Enum:
		return gen.findTypeSym(gen.Defs, ty)

	default:
		panic(fmt.Sprintf("unknown type '%T'", ty))
	}
}

func (gen *generator) arrayType(ty *types.Array) string {
	if s, ok := arrayTypes[ty]; ok {
		return s
	}
	elemTypeName := gen.TypeString(ty.ElemType())
	typeName := fmt.Sprintf("%s_array%d", elemTypeName, ty.Size())
	alreadyDefined := false
	for _, typeName0 := range arrayTypes {
		if typeName0 == typeName {
			// Prevent similar typedefs.
			alreadyDefined = true
		}
	}
	if !alreadyDefined {
		gen.typeSect.WriteString(
			fmt.Sprintf("typedef %s %s[%d];\n", elemTypeName, typeName, ty.Size()),
		)
	}
	arrayTypes[ty] = typeName
	return typeName
}

func (gen *generator) findTypeSym(
	defs *orderedmap.OrderedMap[*ast.Ident, checker.Symbol],
	t types.Type,
	// prefix string,
) string {
	otherModulesDefs := []*orderedmap.OrderedMap[*ast.Ident, checker.Symbol]{}

	for def := defs.Front(); def != nil; def = def.Next() {
		def := def.Value

		switch sym := def.(type) {
		case *checker.Module:
			otherModulesDefs = append(otherModulesDefs, sym.Defs)

		case *checker.Struct, *checker.Enum, *checker.TypeAlias:
			if types.SkipTypeDesc(sym.Type()) == t {
				// return prefix + "Ty" + sym.Name()
				return gen.name(sym)
			}
		}
	}

	for _, otherMod := range otherModulesDefs {
		typeSymStr := gen.findTypeSym(otherMod, t)

		if typeSymStr != "ERROR_CGEN__CANNOT_FIND_TYPE_NAME" {
			return typeSymStr
		}
	}

	return "ERROR_CGEN__CANNOT_FIND_TYPE_NAME"
}
