import std/strformat

import jet/literal


type
    NodeKinds* = set[NodeKind]

    NodeKind* = enum
        # Leaf
        nkEmpty
            ## Empty node
        nkProgram
            ## File
        nkId
            ## Unckeched symbol
        nkSym
            ## Checked symbol
        nkLit
            ## Some typed literal

        # Statements
        nkLetStmt
            ## Id(name)
            ## expr(type)
            ## EqExpr(value)
            ## PragmaList
        nkDefStmt
            ## Id(name) | ExprDotExpr(instance-and-name)
            ## ParamList(params)
            ## expr(return-type)
            ## EqExpr(body)
            ## PragmaList
        nkTypedefStmt
            ## Id(name)
            ## EqExpr(body)
            ## PragmaList
        nkReturnStmt
            ## expr

        nkParam
            ## Id(name)
            ## expr(type)
            ## expr(default-value)
            ## Pragma(pragma)
        nkGenericParam
            ## Id
        nkParamList
            ## nkParam

        nkIfExpr
            ## IfBranch+
            ## ElseBranch?
        nkIfBranch
            ## expr(condition)
            ## DoExpr(body)
        nkElseBranch
            ## stmt+
        nkDoExpr
            ## stmt+
        nkEqExpr
            ## stmt+
        nkBarExpr
            ## expr
        nkExprEqExpr
            ## expr(left)
            ## expr(right)
        nkExprColonExpr
            ## expr(left)
            ## expr(right)
        nkExprDotExpr
            ## expr(left)
            ## expr(right)

        nkParen
            ## expr(elements)*
        nkBrace
            ## expr(elements)*
        nkExprParen
            ## expr(prefix)
            ## Paren(elements)
        nkExprBrace
            ## expr(prefix)
            ## Brace(elements)
        nkAssign
            ## Id(variable)
            ## EqExpr(value)
        nkPrefix
            ## Id(op)
            ## expr(operand)
        nkPostfix
            ## Id(op)
            ## expr(operand)
        nkInfix
            ## Id(op)
            ## expr(left-operand)
            ## expr(right-operand)
        nkPragma
            ## Id(name)
            ## Paren(args)
        nkPragmaList
            ## Pragma*
        nkVarDecl
            ## Id(name)+
            ## expr(type)
            ## EqExpr(value)
            ## PragmaList(pragmas)

    Node* = ref object
        case kind* : NodeKind
        of nkEmpty : nil
        of nkId    : id*       : string
        of nkLit   : lit*      : TypedLiteral
        else       : children* : seq[Node]

func isLeaf*(self: Node): bool =
    result = self.kind in {nkEmpty, nkId, nkLit}

func `$`*(self: Node): string =
    result = ($self.kind)[2..^1]

    case self.kind
    of nkLit:
        result &= $self.lit
    of nkId:
        result &= fmt" '{self.id}'"
    else:
        discard
