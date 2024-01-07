import std/tables
import std/hashes

import jet/ast/types


type SymKind* = enum
    skType
    skVar
    skVal
    skParam         ## A parameter variable
    skLoopVar       ## A 'for' loop variable
    skReturnVal     ## A value from 'return' statement
    skFunc
    skBuiltinFunc

type Sym* = ref object of RootObj
    `type`* : Type
    id*     : string
    kind*   : SymKind

type Scope* = ref object
    depth* : int
    outer* : Scope
    syms*  : OrderedTable[string, Sym]

let builtinTypeSyms* = [
    Sym(id: "unknown", `type`: unknownType),
    Sym(id: "nil", `type`: nilType),
    Sym(id: "unit", `type`: unitType),
    Sym(id: "isize", `type`: isizeType),
    Sym(id: "usize", `type`: usizeType),
    Sym(id: "i8", `type`: i8Type),
    Sym(id: "i16", `type`: i16Type),
    Sym(id: "i32", `type`: i32Type),
    Sym(id: "i64", `type`: i64Type),
    Sym(id: "u8", `type`: u8Type),
    Sym(id: "u16", `type`: u16Type),
    Sym(id: "u32", `type`: u32Type),
    Sym(id: "u64", `type`: u64Type),
    Sym(id: "f32", `type`: f32Type),
    Sym(id: "f64", `type`: f64Type),
    Sym(id: "char", `type`: charType),
    Sym(id: "bool", `type`: boolType),
    Sym(id: "string", `type`: stringType),
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
