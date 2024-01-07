import std/sequtils

import ast/nodes

export nodes


# ----- Leaf ------ #
proc newProgram*(): Node =
    result = newNode(nkProgram)


# ----- Stmt ------ #
proc newFunc*(name, params, returnType, body: Node): Node =
    assert(name != nil and name.kind in {nkId, nkExprDotExpr})
    assert(params != nil and params.kind == nkParen)
    assert(returnType != nil)
    assert(body != nil and body.kind in {nkEmpty, nkEqExpr})

    result  = newNode(nkFunc)
    result &= name
    result &= params
    result &= returnType
    result &= body
    result &= newEmptyNode()

proc newType*(name, body: Node): Node =
    assert(name != nil and name.kind == nkId)
    assert(body != nil and body.kind == nkEqExpr)

    result  = newNode(nkType)
    result &= name
    result &= body
    result &= newEmptyNode()

proc newReturnStmt*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkReturn)
    result &= expr


# ----- Expr ------ #
proc newEmptyIfExpr*(): Node =
    result = newNode(nkIf)

proc newIfExpr*(branches: openArray[Node]; elseBranch: Node = nil): Node =
    assert(branches.len() > 0)
    assert(branches.allIt(it != nil and it.kind == nkIfBranch), $branches.mapIt($it.kind))

    result  = newNode(nkIf)
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

proc newEmptyMatchExpr*(): Node =
    result = newNode(nkMatch)

proc newMatchExpr*(expr: Node; cases: openArray[Node]; elseBranch: Node = nil): Node =
    assert(expr != nil)
    assert(cases.allIt(it != nil))

    result  = newNode(nkMatch)
    result &= expr
    result &= cases

    if elseBranch != nil:
        assert(elseBranch.kind == nkElseBranch, $elseBranch.kind)
        result &= elseBranch

proc newMatchExpr*(expr: Node; cases: Node; elseBranch: Node = nil): Node =
    result = newMatchExpr(expr, cases, elseBranch)

proc newMatchCase*(expr, bodyOrGuard: Node): Node =
    assert(expr != nil)
    assert(bodyOrGuard != nil and bodyOrGuard.kind in {nkDoExpr, nkIfBranch},
        if bodyOrGuard != nil: $bodyOrGuard.kind else: "null")

    result  = newNode(nkCase)
    result &= expr
    result &= bodyOrGuard

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

proc newEmptyExprParen*(expr: Node): Node =
    assert(expr != nil)

    result  = newNode(nkExprParen)
    result &= expr
    result &= newParen()

proc newExprParen*(expr: Node; elems: openArray[Node]): Node =
    assert(elems.allIt(it != nil))

    result = newEmptyExprParen(expr)
    result[1].children = @elems

proc newExprParen*(expr: Node; elem: Node): Node =
    result = newExprParen(expr, [elem])

proc newExprParenFromParen*(expr: Node; paren: Node): Node =
    assert(expr != nil)
    assert(paren != nil and paren.kind == nkParen)

    result  = newNode(nkExprParen)
    result &= expr
    result &= paren

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

proc newAnnotationList*(annotations: openArray[Node]): Node =
    result = newNode(nkAnnotationList)

    for annotation in annotations:
        assert(annotation != nil and annotation.kind in {nkId, nkExprParen, nkAnnotationList},
            if annotation == nil: "nil" else: $annotation.kind)

        if annotation.kind == nkAnnotationList:
            result &= annotation.children
        else:
            result &= annotation

proc newAnnotationList*(annotation: Node): Node =
    result = newAnnotationList([annotation])

proc newEmptyAnnotationList*(): Node =
    result = newAnnotationList([])

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

proc newVariant*(entry: Node): Node =
    assert(entry != nil and entry.kind in {nkId, nkInfix},
        if entry == nil: "nil" else: $entry.kind)

    result  = newNode(nkVariant)
    result &= entry
