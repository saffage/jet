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
    of OpAdd, OpSub, OpMul:
      if left.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator")

      if right.kind notin {tyI8, tyI16, tyI32, tyI64}:
        raiseSemanticError("invalid type for " & $op & " operator")

      if left.kind != right.kind:
        raiseSemanticError(
          "both expressions must be of the same type, got '" &
          $left & "' and '" & $right & "'")

      left
    else:
      todo()

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
        if expectedType != nil and
           expectedType.kind in {tyI8, tyI16, tyI32, tyI64}:
          # TODO: check int range
          return expectedType
        module.getBuiltinType(tyI32)
      of lkNil:
        # if expectedType != nil and
        #    expectedType.kind == tyRef:
        #     return expectedType
        module.getBuiltinType(tyNil)
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
      of Prefix:
        let opKind = expr.children[0].op
        let operand = module.typeOfExpr(expr.children[1])

        if opKind == OpRef:
          TypeRef(kind: tyRef, parent: operand)
        else:
          todo()
      else:
        todo()

func genSym(module: ModuleRef; tree: AstNode): SymbolRef
  {.raises: [SemanticError, ValueError, ModuleError].} =
  result = nil

  if not tree.isSymDecl():
    return

  case tree.branchKind
  of VarDecl, ValDecl:
    let id = tree.children[0].id
    let typeExpr = tree.children[1]
    let body = tree.children[2]

    var `type` = module.typeOfExpr(typeExpr)
    let bodyType = module.typeOfExpr(body, `type`)

    debugEcho(`type`)
    debugEcho(bodyType)

    if `type` == nil:
      `type` = module.typeOfExpr(body)
    else:
      # check body type
      if not isCompatibleTypes(`type`, bodyType):
        raiseSemanticError(&"invalid type for '{id}'; expected {`type`}, got {bodyType}")

    if `type` == nil:
      raiseSemanticError(&"unable to infer type for '{id}'")

    let symKind = case tree.branchKind:
      of VarDecl: skVar
      of ValDecl: skVal
      else: unreachable()

    if `type`.kind == tyNil:
      raiseSemanticError("variable cannot have type nil")

    result = SymbolRef(
      id: id,
      kind: symKind,
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
