import
  std/strformat,
  std/tables,

  jet/ast,
  jet/symbol,
  jet/magics

{.push, raises: [].}

#
# Module
#

type
  ModuleRef* = ref Module
  Module = object of Symbol
    tree*   : AstNode
    magics* : Table[MagicKind, SymbolRef] ## Keep in sync with `magics.MagicKind`
    isMain* : bool

    importedModules* : seq[ModuleRef]
    importedSymbols* : seq[SymbolRef]

  ModuleError* = object of CatchableError

template raiseModuleError(message: string) =
  raise (ref ModuleError)(msg: message)

func registerSymbol*(self: ModuleRef; symbol: SymbolRef)
  {.raises: [ModuleError, ValueError].} =
  if self.scope.getSymbolRec(symbol.id) != nil:
    raiseModuleError(&"attempt to redefine identifier: '{symbol.id}'")

  self.scope.symbols &= symbol

proc registerMagicSyms(self: ModuleRef) =
  let
    i8Type  = TypeRef(kind: tyI8)
    i16Type = TypeRef(kind: tyI16)
    i32Type = TypeRef(kind: tyI32)
    i64Type = TypeRef(kind: tyI64)

    i8Sym  = SymbolRef(kind: skType, typ: i8Type)
    i16Sym = SymbolRef(kind: skType, typ: i16Type)
    i32Sym = SymbolRef(kind: skType, typ: i32Type)
    i64Sym = SymbolRef(kind: skType, typ: i64Type)

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
  result = self.scope.getSymbolRec(id)

proc newModule*(tree: AstNode): ModuleRef =
  result = ModuleRef(tree: tree, scope: newScope())
  result.registerMagicSyms()

{.pop.} # raises: []
