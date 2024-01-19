import
  std/strformat,
  std/sequtils,

  jet/ast,
  jet/symbol,

  lib/utils

{.push, raises: [].}

#
# Module
#

type
  ModuleRef* = ref Module
  Module = object
    rootScope* : ScopeRef
    rootTree*  : AstNode
    isMain*    : bool

  ModuleError* = object of CatchableError

template raiseModuleError(message: string) =
  raise (ref ModuleError)(msg: message)

func registerSymbol*(self: ModuleRef; symbol: SymbolRef)
  {.raises: [ModuleError, ValueError].} =
  if self.rootScope.getSymbolRec(symbol.id) != nil:
    raiseModuleError(&"attempt to redefine identifier: '{symbol.id}'")

  self.rootScope.symbols &= symbol

func registerPrimitives(self: ModuleRef)
  {.raises: [ModuleError, ValueError].} =
  let i32Type = TypeRef(kind: tyI32)
  let i32Sym = SymbolRef(id: "i32", kind: skType, `type`: i32Type)
  self.registerSymbol(i32Sym)

proc newModule*(rootTree: AstNode): ModuleRef
  {.raises: [ModuleError, ValueError].} =
  result = ModuleRef(rootTree: rootTree, rootScope: newScope())
  result.registerPrimitives()
