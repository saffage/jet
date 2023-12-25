import std/tables
import std/hashes

import jet/ast/types


type SymKind* = enum
    skInvalid
    skType
    skLit
    skVar
    skParam
    skVal           ## A value from anywhere
    skReturnVal     ## A value from 'return expr' statement
    skFunc
    skBuiltinFunc

type Sym* = ref object of RootObj
    id*  : string
    typ* : Type

    case kind* : SymKind
    of skInvalid:
        nil
    of skType, skFunc, skVar, skParam, skLit, skBuiltinFunc, skVal, skReturnVal:
        nil

type Scope* = ref object
    depth* : int
    outer* : Scope
    syms*  : OrderedTable[string, Sym]

let builtinTypeSyms* = [
    Sym(id: "unknown", typ: unknownType),
    Sym(id: "null", typ: nullType),
    Sym(id: "unit", typ: unitType),
    Sym(id: "isize", typ: isizeType),
    Sym(id: "usize", typ: usizeType),
    Sym(id: "i8", typ: i8Type),
    Sym(id: "i16", typ: i16Type),
    Sym(id: "i32", typ: i32Type),
    Sym(id: "i64", typ: i64Type),
    Sym(id: "u8", typ: u8Type),
    Sym(id: "u16", typ: u16Type),
    Sym(id: "u32", typ: u32Type),
    Sym(id: "u64", typ: u64Type),
    Sym(id: "f32", typ: f32Type),
    Sym(id: "f64", typ: f64Type),
    Sym(id: "char", typ: charType),
    Sym(id: "bool", typ: boolType),
    Sym(id: "string", typ: stringType),
]


# ----- ENV ----- #
func newScope*(): Scope =
    result = Scope(depth: 0, outer: nil, syms: initOrderedTable[string, Sym]())

func newScopeEnclosed*(outer: Scope): Scope =
    result = Scope(depth: outer.depth + 1, outer: outer, syms: initOrderedTable[string, Sym]())

func getSym*(self: Scope; name: string): Sym =
    assert(self != nil)

    result = self.syms.getOrDefault(name, nil)

    if result == nil and self.outer != nil:
        result = self.outer.getSym(name)

func setSym*(self: Scope; name: string, sym: Sym) =
    assert(self != nil)

    self.syms[name] = sym

func contains*(self: Scope; name: string): bool =
    assert(self != nil)

    result = self.syms.contains(name)

    if not result and self.outer != nil:
        result = self.outer.contains(name)

func clone*(self: Scope): Scope =
    ## **Note:** the `outer` field will points to the same object.
    if self == nil: return nil

    result       = newScope()
    result.depth = self.depth
    result.outer = self.outer
    result.syms  = self.syms

func cloneRecursive*(self: Scope): Scope =
    ## Same with `clone`, but the `outer` field is also clones recursively.
    if self == nil: return nil

    result       = newScope()
    result.depth = self.depth
    result.outer = self.outer.cloneRecursive()
    result.syms  = self.syms

template `[]`*(self: Scope; name: string): Sym = self.getSym(name)
template `[]=`*(self: Scope; name: string; sym: Sym) = self.setSym(name, sym)
