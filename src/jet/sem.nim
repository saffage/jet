import
  jet/ast,
  jet/symbol,
  jet/module,

  lib/utils

{.push, raises: [].}

func isSymDecl(tree: AstNode): bool =
  result =
    tree.kind == Branch and
    tree.branchKind in {ValDecl, VarDecl}

proc typeFromExpr(module: ModuleRef; expr: AstNode): TypeRef =
  # что может быть лучше этого?
  result = module.rootScope.getSymbol(expr.id).`type`

func genSym(module: ModuleRef; tree: AstNode): SymbolRef =
  result = nil

  if not tree.isSymDecl():
    return

  case tree.branchKind
  of VarDecl:
    let id = tree.children[0].id
    let typeExpr = tree.children[1]

    result = SymbolRef(
      id: id,
      kind: skVar,
      `type`: module.typeFromExpr(typeExpr),
      # scope: module.rootScope, # recursive
    )
  else:
    unimplemented()

proc traverseSymbols*(module: ModuleRef)
  {.raises: [ModuleError, ValueError].} =
  for tree in module.rootTree.children:
    let sym = module.genSym(tree)
    module.registerSymbol(sym)
