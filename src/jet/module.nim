import
  std/strformat,
  std/tables,

  jet/ast,
  jet/symbol,
  jet/magics

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
  i8Sym* = SymbolRef(kind: skType, `type`: i8Type)
  i16Sym* = SymbolRef(kind: skType, `type`: i16Type)
  i32Sym* = SymbolRef(kind: skType, `type`: i32Type)
  i64Sym* = SymbolRef(kind: skType, `type`: i64Type)

#
# Module
#

type
  ModuleRef* = ref Module
  Module = object
    rootScope* : ScopeRef
    rootTree*  : AstNode
    magics*    : Table[MagicKind, SymbolRef] ## Keep in sync with `magics.MagicKind`
    isMain*    : bool

  ModuleError* = object of CatchableError

template raiseModuleError(message: string) =
  raise (ref ModuleError)(msg: message)

func registerSymbol*(self: ModuleRef; symbol: SymbolRef)
  {.raises: [ModuleError, ValueError].} =
  if self.rootScope.getSymbolRec(symbol.id) != nil:
    raiseModuleError(&"attempt to redefine identifier: '{symbol.id}'")

  self.rootScope.symbols &= symbol

proc registerMagicSyms(self: ModuleRef) =
  self.magics = {
    mTypeI8: i8Sym,
    mTypeI16: i16Sym,
    mTypeI32: i32Sym,
    mTypeI64: i64Sym,
  }.toTable()

func getMagicSym*(self: ModuleRef; magic: MagicKind): SymbolRef =
  result = try:
    self.magics[magic]
  except KeyError:
    raise newException(Defect, "unimplemented magic: '" & $magic & "'")

func getSym*(self: ModuleRef; id: string): SymbolRef =
  result = self.rootScope.getSymbolRec(id)

proc newModule*(rootTree: AstNode): ModuleRef =
  result = ModuleRef(rootTree: rootTree, rootScope: newScope())
  result.registerMagicSyms()

{.pop.} # raises: []
