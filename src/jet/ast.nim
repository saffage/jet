import std/sequtils

import ast/nodes

export nodes


# ----- Leaf ------ #
proc newProgram*(): Node =
    result = newNode(nkProgram)


# ----- Stmt ------ #
proc newLetStmt*(id, typeExpr, value: Node): Node =
    assert(id != nil and id.kind == nkId)
    assert(value != nil and value.kind == nkEqExpr)

    result  = newNode(nkLetStmt)
    result &= id
    result &= (if typeExpr == nil: newEmptyNode() else: typeExpr)
    result &= value
    result &= newEmptyNode()

proc newDefStmt*(name, params, returnType, body: Node): Node =
    assert(name != nil and name.kind in {nkId, nkExprDotExpr})
    assert(params != nil and params.kind == nkParamList)
    assert(returnType != nil)
    assert(body != nil and body.kind in {nkEmpty, nkEqExpr})

    result  = newNode(nkDefStmt)
    result &= name
    result &= params
    result &= returnType
    result &= body
    result &= newEmptyNode()

proc newTypedefStmt*(name, body: Node): Node =
    assert(name != nil and name.kind == nkId)
    assert(body != nil and body.kind == nkEqExpr)

    result  = newNode(nkTypedefStmt)
    result &= name
    result &= body
    result &= newEmptyNode()

proc newReturnStmt*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkReturnStmt)
    result &= expr


# ----- Expr ------ #
proc newEmptyIfExpr*(): Node =
    result = newNode(nkIfExpr)

proc newIfExpr*(branches: openArray[Node]; elseBranch: Node = nil): Node =
    assert(branches.len() > 0)
    assert(branches.allIt(it.kind == nkIfBranch), $branches.mapIt($it.kind))

    result  = newNode(nkIfExpr)
    result &= branches

    if elseBranch != nil:
        assert(elseBranch.kind == nkElseBranch, $elseBranch.kind)
        result &= elseBranch

proc newIfExpr*(branch: Node; elseBranch: Node = nil): Node =
    result = newIfExpr([branch], elseBranch)

proc newIfBranch*(expr: Node; body: Node): Node =
    assert(expr != nil)
    assert(body != nil and body.kind == nkDoExpr, if body == nil: "null" else: $body.kind)

    result  = newNode(nkIfBranch)
    result &= expr
    result &= body

proc newElseBranch*(body: Node): Node =
    assert(body != nil and body.kind == nkDoExpr, if body == nil: "null" else: $body.kind)

    result  = newNode(nkElseBranch)
    result &= body

proc newEmptyElseBranch*(): Node =
    result = newNode(nkElseBranch)

proc newDoExpr*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkDoExpr)
    result &= expr

proc newDoExpr*(): Node =
    result = newNode(nkDoExpr)

proc newEqExpr*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkEqExpr)
    result &= expr

proc newEqExpr*(): Node =
    result = newNode(nkEqExpr)

proc newBarExpr*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkBarExpr)
    result &= expr

proc newExprEqExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprEqExpr)
    result &= left
    result &= right

proc newExprColonExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprColonExpr)
    result &= left
    result &= right

proc newExprDotExpr*(left, right: Node): Node =
    assert(left != nil)
    assert(right != nil)

    result  = newNode(nkExprDotExpr)
    result &= left
    result &= right


# ----- Other ------ #
proc newBrace*(): Node =
    result = newNode(nkBrace)

proc newParen*(): Node =
    result = newNode(nkParen)

proc newExprBrace*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprBrace)
    result &= expr
    result &= newBrace()

proc newExprParen*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprParen)
    result &= expr
    result &= newParen()

proc newAssign*(id, value: Node): Node =
    assert(id != nil and id.kind == nkId)
    assert(value != nil and value.kind == nkEqExpr)

    result  = newNode(nkAssign)
    result &= id
    result &= value

proc newPrefix*(op: Node; operand: Node): Node =
    assert(op != nil)
    assert(operand != nil)

    result = Node(kind: nkPrefix)
    result &= op
    result &= operand

proc newPostfix*(op: Node; operand: Node): Node =
    assert(op != nil)
    assert(operand != nil)

    result = Node(kind: nkPostfix)
    result &= op
    result &= operand

proc newInfix*(op: Node; leftOperand, rightOperand: Node): Node =
    assert(op != nil)
    assert(leftOperand != nil)
    assert(rightOperand != nil)

    result = Node(kind: nkInfix)
    result &= op
    result &= leftOperand
    result &= rightOperand

proc newPragma*(name, args: Node): Node =
    assert(name != nil and name.kind == nkId)

    result  = newNode(nkPragma)
    result &= name

    if args != nil:
        assert(args.kind == nkParen)
        result &= args
    else:
        result &= newEmptyNode()

proc newParam*(name, typeExpr, defaultValue, pragmas: Node): Node =
    assert(name != nil and name.kind == nkId)

    result  = newNode(nkParam)
    result &= name

    if typeExpr != nil:
        result &= typeExpr
    else:
        result &= newEmptyNode()

    if defaultValue != nil:
        result &= defaultValue
    else:
        assert(typeExpr != nil)
        result &= newEmptyNode()

    if pragmas != nil:
        assert(pragmas.kind in {nkEmpty, nkPragmaList}, $pragmas.kind)
        result &= pragmas
    else:
        result &= newEmptyNode()

proc newParamList*(params: openArray[Node]): Node =
    result = newNode(nkParamList)

    for param in params:
        assert(param != nil and param.kind == nkParam)
        result &= param

proc newParamList*(param: Node): Node =
    result = newParamList([param])

proc newEmptyParamList*(): Node =
    result = newParamList([])

proc newPragmaList*(pragmas: openArray[Node]): Node =
    result = newNode(nkPragmaList)

    for pragma in pragmas:
        assert(pragma != nil and pragma.kind == nkPragma)
        result &= pragma

proc newPragmaList*(pragma: Node): Node =
    result = newPragmaList([pragma])

proc newEmptyPragmaList*(): Node =
    result = newPragmaList([])

proc newGenericParam*(id: Node): Node =
    assert(id != nil and id.kind == nkId)

    result  = newNode(nkGenericParam)
    result &= id

proc newVarDecl*(names: openArray[Node]; typeExpr: Node; eqExpr: Node = nil): Node =
    assert(names.len() > 0)
    assert(names.allIt(it != nil and it.kind == nkId))

    result  = newNode(nkVarDecl)
    result &= names
    result &= (if typeExpr != nil: typeExpr else: newEmptyNode())
    result &= (if eqExpr != nil: (assert(eqExpr.kind == nkEqExpr); eqExpr) else: newEmptyNode())
    result &= newEmptyNode()

proc newVarDecl*(name: Node; typeExpr: Node; eqExpr: Node = nil): Node =
    result = newVarDecl([name], typeExpr, eqExpr)
