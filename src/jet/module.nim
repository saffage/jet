import
  std/strformat,
  std/strutils,
  std/sequtils,

  jet/ast,
  jet/symbol,

  lib/utils

{.push, raises: [].}

#
# Primitives
#

let
  i8Type* = TypeRef(kind: tyI8)
  i16Type* = TypeRef(kind: tyI16)
  i32Type* = TypeRef(kind: tyI32)
  i64Type* = TypeRef(kind: tyI64)

let
  i8Sym* = SymbolRef(id: "i8", kind: skType, `type`: i8Type)
  i16Sym* = SymbolRef(id: "i16", kind: skType, `type`: i16Type)
  i32Sym* = SymbolRef(id: "i32", kind: skType, `type`: i32Type)
  i64Sym* = SymbolRef(id: "i64", kind: skType, `type`: i64Type)

#
# Module
#

type
  ModuleRef* = ref Module
  Module = object
    rootScope* : ScopeRef
    rootTree*  : AstNode
    builtins*  : seq[SymbolRef]
    isMain*    : bool

  ModuleError* = object of CatchableError

template raiseModuleError(message: string) =
  raise (ref ModuleError)(msg: message)

func registerSymbol*(self: ModuleRef; symbol: SymbolRef)
  {.raises: [ModuleError, ValueError].} =
  if self.rootScope.getSymbolRec(symbol.id) != nil:
    raiseModuleError(&"attempt to redefine identifier: '{symbol.id}'")

  self.rootScope.symbols &= symbol

proc registerPrimitives(self: ModuleRef) =
  self.builtins = @[
    i8Sym,
    i16Sym,
    i32Sym,
    i64Sym,
  ]

func getBuiltinSym*(self: ModuleRef; id: string): SymbolRef
  {.raises: [ModuleError].} =
  let idx = self.builtins.findIt(it.id == id)
  result =
    if idx < 0: raiseModuleError("unknown builtin symbol: '" & id & "'")
    else: self.builtins[idx]

func getBuiltinType*(self: ModuleRef; kind: TypeKind): TypeRef
  {.raises: [ModuleError].} =
  assert(kind in {tyI32})
  let idx = self.builtins.findIt(it.`type`.kind == kind)
  result =
    if idx < 0: raiseModuleError("unknown builtin type: '" & $kind & "'")
    else: self.builtins[idx].`type`

func getSym*(self: ModuleRef; id: string): SymbolRef =
  result =
    try:
      self.getBuiltinSym(id)
    except ModuleError:
      self.rootScope.getSymbolRec(id)

proc newModule*(rootTree: AstNode): ModuleRef =
  result = ModuleRef(rootTree: rootTree, rootScope: newScope())
  result.registerPrimitives()
