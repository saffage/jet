import jet/ast
import jet/ast/sym
import jet/ast/types

import lib/utils


proc getTreeType(tree: Node): Type =
    result = case tree.kind:
        of nkDefStmt, nkVarDecl:
            types.unitType
        of nkExprParen, nkExprBrace:
            tree[0].getTreeType()
        of nkId, nkLit:
            tree.`type`
        else: unimplemented($tree.kind)

proc getType(tree: Node): Type =
    result = case tree.kind:
        of nkDefStmt:
            types.unitType
        of nkExprParen, nkExprBrace:
            tree[0].getTreeType()
        of nkId, nkLit:
            tree.`type`
        else: unimplemented($tree.kind)

proc genVarDeclSym(tree: Node) =
    let `type` = tree.getType()
    let sym = Sym()

proc semGenTreeTypes*(tree: Node) =
    discard

proc semGenTreeSyms*(tree: Node) =
    case tree.kind
    of nkEmpty, nkId, nkGenericId, nkLit:
        discard
    of nkVarDecl:
        tree.genVarDeclSym()
    else:
        for child in tree.children:
            semGenTreeSyms(child)
