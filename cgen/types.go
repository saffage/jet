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

func (gen *generator) TypeString(t types.Type) string {
	assert.Ok(!types.IsTypeDesc(t))

	switch t := t.Underlying().(type) {
	case nil:
		return _ErrorNilType

	case *types.TypeDesc:
		return _ErrorMetaType

	case *types.Alias:
		panic("unreachable")

	case *types.Primitive:
		switch t.Kind() {
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

		default:
			panic("unreachable")
		}

	case *types.Func:
		panic("not implemented")

	case *types.Tuple:
		if t.Equals(types.Unit) {
			return "void"
		}

		panic("not implemented")

	case *types.Array:
		if s, ok := arrayTypes[t]; ok {
			return s
		}
		elemTypeStr := gen.TypeString(t.ElemType())
		typeStr := fmt.Sprintf("%s_array%d", elemTypeStr, t.Size())
		gen.typeSect.WriteString(
			fmt.Sprintf("typedef %s %s[%d];\n\n", elemTypeStr, typeStr, t.Size()),
		)
		arrayTypes[t] = typeStr
		return typeStr

	case *types.Ref:
		return gen.TypeString(t.Base()) + "*"

	case *types.Struct:
		if t == types.String {
			return "char*"
		}
		return gen.findTypeSym(gen.Defs, t)

	case *types.Enum:
		return gen.findTypeSym(gen.Defs, t)

	default:
		panic(fmt.Sprintf("unknown type '%T'", t))
	}
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
