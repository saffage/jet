import std/strformat
import std/strutils
import std/sequtils

import jet/ast
import jet/literal

import lib/utils


proc ast2jet*(tree: Node; level: Natural = 0): string =
    const indent = "    "
    result = indent.repeat(level)

    result.add case tree.kind:
        of nkProgram:
            tree.children
                .mapIt(ast2jet(it))
                .join("\n")
        of nkEmpty:
            ""
        of nkId:
            tree.id
        of nkGenericId:
            fmt"<{tree.id}>"
        of nkLit:
            let lit = tree.lit.str()
            fmt"{lit}"
        of nkDefStmt:
            let
                name       = ast2jet(tree[0])
                params     = ast2jet(tree[1])
                body       = ast2jet(tree[3])
                returnType =
                    if (let tmp = ast2jet(tree[2]); tmp.len() > 0): tmp & " "
                    else: ""

            fmt"def {name}{params} {returnType}{body}"
        of nkTypedefStmt:
            let
                name = ast2jet(tree[0])
                body = ast2jet(tree[1])

            fmt"typedef {name} {body}"
        of nkExprDotExpr:
            let
                left  = ast2Jet(tree[0])
                right = ast2Jet(tree[1])

            fmt"{left}.{right}"
        of nkEqExpr:
            # if tree.len() == 1:
            #     let expr = ast2jet(tree[0])

            #     fmt("= {expr}")
            let exprs = tree.children
                .mapIt(ast2jet(it, level + 1))
                .join("\n")

            fmt("= \n{exprs}")
        of nkDoExpr:
            let exprs = tree.children
                .mapIt(ast2jet(it, level + 1))
                .join("\n")

            fmt("do\n{exprs}")
        of nkParen:
            let elems = tree.children
                .mapIt(ast2jet(it))
                .join(", ")

            fmt"({elems})"
        of nkBrace:
            let elems = tree.children
                .mapIt(ast2jet(it))
                .join(", ")

            fmt"{{ {elems} }}"
        of nkExprParen:
            let
                expr  = ast2jet(tree[0])
                paren = ast2jet(tree[1])

            fmt"{expr}{paren}"
        of nkExprBrace:
            let
                expr  = ast2jet(tree[0])
                brace = ast2jet(tree[1])

            fmt"{expr}{brace}"
        of nkVarDecl:
            let
                names = tree[0 ..^ 4]
                    .mapIt(ast2jet(it))
                    .join(", ")
                value = ast2jet(tree[^2])
                typeExpr =
                    if (let tmp = ast2jet(tree[^3]); tmp.len() > 0):
                        if value.len() > 0: tmp & " "
                        else: tmp
                    else: ""

            fmt"{names} {typeExpr}{value}"
        of nkReturnStmt:
            let expr = ast2jet(tree[0])
            fmt"return {expr}"
        of nkInfix:
            let
                infix = ast2jet(tree[0])
                left  = ast2jet(tree[1])
                right = ast2jet(tree[2])

            fmt"{left} {infix} {right}"
        of nkPrefix:
            let
                prefix  = ast2jet(tree[0])
                operand = ast2jet(tree[1])

            fmt"{prefix}{operand}"
        of nkPostfix:
            let
                postfix = ast2jet(tree[0])
                operand = ast2jet(tree[1])

            fmt"{operand}{postfix}"
        of nkIfExpr:
            let
                elseBranch =
                    if tree[^1].kind == nkElseBranch: ast2jet(tree[^1])
                    else: ""
                branchNodes =
                    if tree[^1].kind == nkElseBranch: tree[0 ..^ 2]
                    else: tree.children
                branches = branchNodes
                    .mapIt(ast2jet(it))
                    .join("\nel")

            fmt("{branches}\n{elseBranch}")
        of nkIfBranch:
            let
                cond = ast2Jet(tree[0])
                body = ast2Jet(tree[1])

            fmt"if {cond} {body}"
        of nkElseBranch:
            if tree.len() == 0:
                unreachable()

            let exprs = tree.children
                .mapIt(ast2jet(it, level + 1))
                .join("\n")

            fmt("else\n{exprs}")
        of nkExprEqExpr:
            let
                left  = ast2jet(tree[0])
                right = ast2jet(tree[1])

            fmt"{left} = {right}"
        of nkExprColonExpr:
            let
                left  = ast2jet(tree[0])
                right = ast2jet(tree[1])

            fmt"{left}: {right}"
        of nkPragma:
            let
                name = ast2jet(tree[0])
                args = ast2jet(tree[1])

            fmt"#{name}{args}"
        of nkPragmaList:
            let pragmas = tree.children
                .mapIt(ast2jet(it))
                .join(", ")

            fmt"#[{pragmas}]"
        of nkMatch:
            let
                expr  = ast2jet(tree[0])
                elseBranch =
                    if tree[^1].kind == nkElseBranch:
                        var tmp = ast2jet(tree[^1])
                        tmp.insert("| ", tmp.findIt(it != ' ')) # oh god...
                        tmp
                    else: ""
                caseNodes =
                    if tree[^1].kind == nkElseBranch: tree[1 ..^ 2]
                    else: tree[1 ..^ 1]
                cases = caseNodes
                    .mapIt(if it.kind == nkIfBranch: "| " & ast2jet(it) else: ast2jet(it))
                    .join("\n")

            if tree[^1].kind == nkElseBranch:
                fmt("match {expr}\n{cases}\n{elseBranch}")
            else:
                fmt("match {expr}\n{cases}")
        of nkVariant:
            let variant = ast2jet(tree[0])

            fmt"| {variant}"
        of nkCase:
            let
                expr        = ast2jet(tree[0])
                bodyOrGuard = ast2jet(tree[1])

            fmt"| {expr} {bodyOrGuard}"
