package cgen

import (
	"fmt"

	"github.com/saffage/jet/checker"
	"github.com/saffage/jet/types"
)

const (
	_ErrorNilType  = "ERROR_CGEN__NIL_TYPE"
	_ErrorMetaType = "ERROR_CGEN__META_TYPE"
)

func (gen *Generator) TypeString(t types.Type) string {
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

		case types.KindI32:
			return "Ti32"

		case types.KindU8:
			return "Tu8"

		default:
			panic("unreachable")
		}

	case *types.Array:
		panic("not implemented")

	case *types.Func:
		panic("not implemented")

	case *types.Tuple:
		if types.Unit.Equals(t) {
			return "void"
		}

		panic("not implemented")

	case *types.Ref:
		return gen.TypeString(t.Base()) + "*"

	case *types.Struct:
		if t.Underlying() == types.String {
			return "char*"
		}

		for _, sym := range gen.Defs {
			switch sym.(type) {
			case *checker.Struct, *checker.TypeAlias:
				if types.SkipTypeDesc(sym.Type()) == t {
					return "Ty" + sym.Name()
				}
			}
		}

		return "ERROR_CGEN__CANNOT_FIND_TYPE_NAME"

	default:
		panic(fmt.Sprintf("unknown type '%T'", t))
	}
}
