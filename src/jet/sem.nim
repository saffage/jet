import
  std/strformat,

  jet/ast,
  jet/symbol,
  jet/module,

  lib/utils

{.push, raises: [].}

type
  SemanticError* = object of CatchableError

template raiseSemanticError*(message: string) =
  raise (ref SemanticError)(msg: message)

func isSymDecl(tree: AstNode): bool =
  result =
    tree.kind == Branch and
    tree.branchKind in {ValDecl, VarDecl}

proc checkOperandTypes(module: ModuleRef; op: OperatorKind; left, right: TypeRef): TypeRef
  {.raises: [SemanticError].} =
  assert(left != nil and right != nil)
  result = case op:
    of OpAdd, OpSub:
      if left.kind != right.kind:
        raiseSemanticError(
          "both expressions must be of the same type, got '" &
          $left.kind & "' and '" & $right.kind & "'")

      left
    else: todo()

proc typeOfExpr(module: ModuleRef; expr: AstNode; expectedType = nil.TypeRef): TypeRef
  {.raises: [SemanticError, ModuleError].} =
  result = case expr.kind:
    of Empty:
      nil
    of Id:
      var sym = module.getSym(expr.id)

      if sym == nil:
        raiseSemanticError("unbound identifier: '" & expr.id & "'")

      if sym.kind != skType and sym.`type` == nil:
        raiseSemanticError("expression '" & expr.id & "' has no type")

      sym.`type`
    of Lit:
      case expr.lit.kind
      of lkInt:
        if expectedType == nil:
          module.getBuiltinType(tyI32)
        else:
          # TODO: check int range
          expectedType
      else:
        todo($expr.lit.kind)
    of Operator:
      todo()
    of Branch:
      case expr.branchKind
      of Infix:
        let opKind = expr.children[0].op
        let left = module.typeOfExpr(expr.children[1])
        let right = module.typeOfExpr(expr.children[2])
        module.checkOperandTypes(opKind, left, right)
      else: todo()

func genSym(module: ModuleRef; tree: AstNode): SymbolRef
  {.raises: [SemanticError, ValueError, ModuleError].} =
  result = nil

  if not tree.isSymDecl():
    return

  case tree.branchKind
  of VarDecl:
    let id = tree.children[0].id
    let typeExpr = tree.children[1]
    let body = tree.children[2]

    var `type` = module.typeOfExpr(typeExpr)
    let bodyType = module.typeOfExpr(body, `type`)

    if `type` == nil:
      `type` = module.typeOfExpr(body)
    else:
      # check body type
      if `type`.kind != bodyType.kind:
        raiseSemanticError(&"invalid type for '{id}'; expected {`type`.kind}, got {bodyType.kind}")

    if `type` == nil:
      raiseSemanticError(&"unable to infer type for '{id}'")

    result = SymbolRef(
      id: id,
      kind: skVar,
      `type`: `type`,
      scope: module.rootScope, # recursive
    )
  else:
    unimplemented()

proc traverseSymbols*(module: ModuleRef)
  {.raises: [ModuleError, SemanticError, ValueError].} =
  for tree in module.rootTree.children:
    let sym = module.genSym(tree)
    module.registerSymbol(sym)
