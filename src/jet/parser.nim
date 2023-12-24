import std/sugar
import std/tables
import std/strformat
import std/strutils
import std/sequtils
import std/enumutils
import std/importutils

import jet/ast
import jet/scanner
import jet/token
import jet/literal
import jet/parser/block_context
import jet/parser/precedence

import lib/stack
import lib/grammar_docs

import pkg/questionable

import utils
import utils/line_info
import utils/text_style


type
    Parser* = ref object
        scanner    : Scanner
        token      : Token
        prevToken  : Token
        blocks     : Stack[BlockContext]    ## Sequence of block contexts
        pragmaPool : seq[Node]

        prefix : OrderedTable[TokenKind, PrefixParseFn]
        infix  : OrderedTable[TokenKind, InfixParseFn]

    ParserError* = object of CatchableError

    PrefixParseFn = proc(self: Parser): Node
    InfixParseFn  = proc(self: Parser; left: Node): Node

const precedences = {
    LParen     : Call,
    LBracket   : Index,
    LBrace     : Highest,
    Dot        : Member,
    ColonColon : Member,
    Asterisk   : Product,
    Slash      : Product,
    Percent    : Product,
    Plus       : Sum,
    Minus      : Sum,
    PlusPlus   : Sum,
    EqOp       : Cmp,
    NeOp       : Cmp,
    LtOp       : Cmp,
    GtOp       : Cmp,
    LeOp       : Cmp,
    GeOp       : Cmp,
    KwAnd      : And,
    KwOr       : Or,
    Eq         : Assign,
}.toOrderedTable()

proc newParser*(scanner: Scanner): Parser
proc fillTables(self: Parser)
proc parseAll*(self: Parser): Node
proc parseExpr(self: Parser; precedence: Precedence = Lowest): Node
proc parseDef(self: Parser): Node
proc parseLet(self: Parser): Node
proc parseReturn(self: Parser): Node
proc parseIdOrDotExpr(self: Parser): Node
proc parseId(self: Parser): Node
proc parseLit(self: Parser): Node
proc parseTypeExpr(self: Parser): Node
proc parseSemicolon(self: Parser): Node
proc parseSemicolon(self: Parser; left: Node): Node
proc parseIfExpr(self: Parser): Node
proc parseDoExpr(self: Parser): Node
proc parseEqExpr(self: Parser): Node
proc parseBlockExpr(self: Parser; result: Node)
proc parseParen(self: Parser): Node
proc parsePragma(self: Parser): Node

proc parseInfixOp(self: Parser; left: Node): Node
proc parseExprParen(self: Parser; left: Node): Node
proc parseExprEqExpr(self: Parser; left: Node): Node
proc parseVarDecl(self: Parser; left: Node): Node

proc getIntLit(self: Parser): Literal
proc getUIntLit(self: Parser): Literal
proc getFloatLit(self: Parser): Literal
proc parseTestComment(self: Parser)

proc dbg(self: Parser; msg: string = "") =
    const dbgStyle = TextStyle(foreground: Cyan, underlined: true)
    let msg = if msg == "": "" else: (msg @ dbgStyle) & ": "

    privateAccess(Scanner)
    debug(
        fmt"{msg}Parser state:" &
        fmt("\n\tprev: {self.prevToken.human()}") &
        fmt("\n\tcurr: {self.token.human()}") &
        fmt("\n\tscanner:") &
            fmt("\n\t\tprev: {self.scanner.prevToken.human()}") &
            fmt("\n\t\tcurr: {self.scanner.token.human()}") &
        fmt("\n\tblocks: {$self.blocks}")
    )

proc checkIndent(token: Token; context: BlockContext): int
proc nextToken(self: Parser; checkIndent: bool = true)


# ----- ERRORS ----- #
proc err(self: Parser; msg: string) =
    error(msg, self.token.info)
    raise newException(ParserError, msg)

proc errSyntax(self: Parser; msg: string) =
    self.err(fmt"Syntax error: {msg}")

proc errExpectedId(self: Parser) =
    self.err(fmt"expected identifier, got {self.token.kind}")

proc errExpectedExprStart(self: Parser) =
    self.errSyntax(fmt"token '{self.token.kind}' is not an expression start")

proc errExpectedNodeOf(self: Parser; kind: NodeKind) =
    self.errSyntax(fmt"expected node of kind {kind}, got {self.token.kind} instead")

proc errExpectedNodeOf(self: Parser; kinds: NodeKinds) =
    self.errSyntax(fmt"expected node of kinds {kinds}, got {self.token.kind} instead")

proc errExpected(self: Parser; kind: TokenKind) =
    self.errSyntax(fmt"expected token {kind}, got {self.token.kind} instead")

proc errExpected(self: Parser; kinds: set[TokenKind]) =
    let kinds = kinds.mapIt($it).join(" or ")
    self.errSyntax(fmt"expected token {kinds}, got {self.token.kind} instead")

proc errExpectedSameLine(self: Parser) =
    self.errSyntax(fmt"expected expression on one line")

proc errInvalidIndent(self: Parser; explanation: string) =
    self.errSyntax(fmt"invalid indentation. {explanation}")

proc errInvalidIndent(self: Parser) =
    self.errSyntax(fmt"invalid indentation")

proc errInvalidBlockContext(self: Parser) =
    let blockContext = self.blocks.peek()
    self.errInvalidIndent(
        fmt"This token is offside the context started at position [{blockContext.line}:{blockContext.getColumn()}], " &
        fmt"token position is [{self.token.info.dupNoLength()}]. This line will be skipped")

proc errInvalidNotation(self: Parser; explanation: string) =
    self.errSyntax(fmt"invalid notation. {explanation}")

proc errInvalidNotation(self: Parser) =
    self.errSyntax(fmt"invalid notation")

proc errExpectedFirstInLine(self: Parser; explanation: string) =
    self.errSyntax(fmt"token {self.token} must be first in line. {explanation}")

proc errExpectedFirstInLine(self: Parser) =
    self.errSyntax(fmt"token {self.token} must be first in line")

proc errExpectedLastInLine(self: Parser; explanation: string) =
    self.errSyntax(fmt"token {self.token} must be last in line. {explanation}")

proc errExpectedLastInLine(self: Parser) =
    self.errSyntax(fmt"token {self.token} must be last in line")

proc errUnknownOp(self: Parser; op: string) =
    self.err(fmt"Unknown operator: '{op}'")

proc errUnknownOp(self: Parser; op, explanation: string) =
    self.err(fmt"Unknown operator: '{op}'. {explanation}")

# ----- PRIVATE ----- #
template isKind(self: Parser; tokenKind: TokenKind): bool = self.token.kind == tokenKind
template isKind(self: Parser; tokenKinds: set[TokenKind]): bool = self.token.kind in tokenKinds

proc isSameLine(self: Parser): bool =
    return self.prevToken.info.line == self.token.info.line

proc isSameIndent(self: Parser): bool =
    if not self.blocks.isEmpty():
        return self.blocks.peek().getColumn() == self.token.indent() |? -2 # '-2' is unreachable
    else:
        return false

proc expected(self: Parser; kind: TokenKind) =
    if not self.isKind(kind):
        self.errSyntax(fmt"Syntax error: expected '{kind}', got '{self.token.kind}'")

proc expected(self: Parser; kinds: set[TokenKind]) =
    if not self.isKind(kinds):
        let kindsStr = kinds.mapIt(fmt"'{it}'").join(", ")
        self.errSyntax(fmt"expected one of {kindsStr}, got '{self.token.kind}'")

proc tokenNotation(self: Parser): Notation =
    return self.token.notation(self.prevToken.kind, self.scanner.token.kind)

proc skipToken(self: Parser) =
    debug fmt"token {self.token.kind} at {self.token.info} was skipped"
    self.token = !self.scanner.getToken()

proc skip(self: Parser; kind: TokenKind) =
    self.expected(kind)
    self.nextToken()

proc skip(self: Parser; kinds: set[TokenKind]) =
    self.expected(kinds)
    self.nextToken()

proc skipMaybe(self: Parser; kind: TokenKind): bool =
    result = self.isKind(kind)
    if result: self.nextToken() # skip expected

proc skipMaybe(self: Parser; kinds: set[TokenKind]): bool =
    result = self.isKind(kinds)
    if result: self.nextToken() # skip expected

proc skipLine(self: Parser; line: uint32) =
    dbg self, fmt"skipLine {line}"

    while self.token.kind != Last:
        let token = !self.scanner.getToken()

        if token.info.line > line:
            self.token = token
            break

        debug fmt"token {token.kind} at {token.info} was skipped"

    dbg self, fmt"skipLine {line} after"

proc skipBlock(self: Parser; indent: int) =
    while self.token.kind != Last:
        debug fmt"token {self.token.kind} at {self.token.info} was skipped"

        if tokenIndent =? self.token.indent():
            if tokenIndent >= indent: self.skipLine(self.token.info.line)
            else: break
        else: unreachable()

        self.token = !self.scanner.getToken()

proc skipLine(self: Parser) =
    self.skipLine(self.token.info.line)

proc checkIndent(token: Token; context: BlockContext): int =
    ## **Returns:**
    ## - -1 if `token.indent < context` - drop block
    ## - 1 if `token.indent > context` - error
    ## - 0 if `token.indent == context` - ok
    result = 0

    if indent =? token.indent():
        if indent > context:
            return 1
        elif indent < context:
            return -1
        else:
            return 0

proc nextToken(self: Parser; checkIndent: bool) =
    # dbg self, "nextToken"

    let token = !self.scanner.getToken()

    # block IndentCheck:
    #     if indent =? token.indent() and checkIndent:
    #         if self.blocks.isEmpty():
    #             if indent == 0:
    #                 debug fmt"push indentation {indent} on stack"
    #                 self.blocks.push(initBlockContext(Indent, token.info.line.int, indent))
    #             else:
    #                 self.errInvalidIndent(fmt"Expected token with 0 indentation, got {indent}")
    #                 self.skipLine(token.info.line)
    #                 break IndentCheck
    #         else:
    #             let currentIndent = self.blocks.peek()

    #             if indent < currentIndent:
    #                 var toDrop = 0
    #                 for prevIndent in self.blocks:
    #                     if prevIndent < indent:
    #                         self.errInvalidIndent()
    #                         self.skipLine(token.info.line)
    #                         break IndentCheck
    #                     elif prevIndent > indent:
    #                         inc(toDrop)
    #                     else: break
    #                 debug fmt"drop indentation {self.blocks.peek()} from stack"
    #                 self.blocks.drop(toDrop)
    #             elif indent > currentIndent:
    #                 if not self.blocks.isEmpty():
    #                     let blockContext = self.blocks.peek()
    #                     let column       = blockContext.getColumn()

    #                     if column != token.info.column.int:
    #                         self.errInvalidIndent(fmt"This token is offside the context started at position [{blockContext.line}:{column}], [{token.info}]")
    #                         self.skipLine(token.info.line)
    #                         break IndentCheck
    #                 debug fmt"push indentation {indent} on stack"
    #                 self.blocks.push(initBlockContext(Indent, token.info.line.int, indent))
    #             else: discard

    self.prevToken = self.token
    self.token     = token

    # dbg self, "nextToken after"

proc tokenSameLine(self: Parser) =
    if not self.isSameLine():
        self.errExpectedSameLine()
        raise newException(ParserError, "expected ")

proc tokenFirstInLine(self: Parser) =
    if not self.token.isFirstInLine(): self.errExpectedFirstInLine()

proc tokenLastInLine(self: Parser) =
    if not self.token.isLastInLine(): self.errExpectedLastInLine()

proc tokenIndent(self: Parser; expectedIndent: Natural) =
    without indent =? self.token.indent(): self.errInvalidIndent()
    if indent != expectedIndent: self.errInvalidIndent()

proc blockContextFromCurrentToken(self: Parser): BlockContext =
    let column: int

    if self.token.isFirstInLine():
        column = !self.token.indent()

        if column <= self.blocks.peek():
            self.errInvalidIndent()
    else:
        column = self.token.info.column.int

    result = initBlockContext(Indent, self.token.info.line.int, column)

proc checkToken(
    self        : Parser;
    notation    : Notation = Unknown;
    sameLine    : bool = false;
    firstInLine : bool = false;
    lastInLine  : bool = false;
    indent      : ?int = none(int);
    failureFn   : (Parser) -> void = skipLine
) =
    debug fmt"check {notation = }, {sameLine = }, {firstInLine = }, {lastInLine = }, {indent = }"
    let wasErrors = logger.errors

    template check() =
        if wasErrors != logger.errors and failureFn != nil:
            failureFn(self)

    if sameLine:
        self.tokenSameLine()
        check()
    elif firstInLine:
        self.tokenFirstInLine()
        check()
    elif lastInLine:
        self.tokenLastInLine()
        check()
    elif notation != Unknown and self.tokenNotation() != notation:
        self.errInvalidNotation()
        check()
    elif expectedIndent =? indent:
        self.tokenIndent(expectedIndent)
        check()


# ----- API IMPL ----- #
proc newParser(scanner: Scanner): Parser =
    result = Parser(
        scanner: scanner,
        blocks: initBlockContext(Indent, 0, 0).toStack())
    result.fillTables()
    result.nextToken()

proc fillTables(self: Parser) =
    self.prefix[KwDef]    = parseDef
    self.prefix[KwLet]    = parseLet
    self.prefix[KwIf]     = parseIfExpr
    self.prefix[KwDo]     = parseDoExpr
    self.prefix[KwReturn] = parseReturn
    self.prefix[Eq]       = parseEqExpr
    self.prefix[LParen]   = parseParen
    self.prefix[Hashtag]  = parsePragma

    self.prefix[Semicolon]        = parseSemicolon
    self.prefix[Id]               = parseId
    self.prefix[IntLit]           = parseLit
    self.prefix[UIntLit]          = parseLit
    self.prefix[FloatLit]         = parseLit
    self.prefix[CharLit]          = parseLit
    self.prefix[StringLit]        = parseLit
    self.prefix[RawStringLit]     = parseLit
    self.prefix[LongStringLit]    = parseLit
    self.prefix[LongRawStringLit] = parseLit

    self.prefix[ISizeLit] = parseLit
    self.prefix[USizeLit] = parseLit
    self.prefix[I8Lit]    = parseLit
    self.prefix[I16Lit]   = parseLit
    self.prefix[I32Lit]   = parseLit
    self.prefix[I64Lit]   = parseLit
    self.prefix[U8Lit]    = parseLit
    self.prefix[U16Lit]   = parseLit
    self.prefix[U32Lit]   = parseLit
    self.prefix[U64Lit]   = parseLit
    self.prefix[F32Lit]   = parseLit
    self.prefix[F64Lit]   = parseLit
    self.prefix[KwTrue]   = parseLit
    self.prefix[KwFalse]  = parseLit

    self.infix[LParen]     = parseExprParen
    self.infix[Eq]         = parseExprEqExpr
    self.infix[Id]         = parseVarDecl
    self.infix[DotDotDot]  = parseVarDecl
    self.infix[DotDot]     = parseInfixOp
    self.infix[DotDotLess] = parseInfixOp
    self.infix[KwAnd]      = parseInfixOp
    self.infix[KwOr]       = parseInfixOp
    self.infix[EqOp]       = parseInfixOp
    self.infix[NeOp]       = parseInfixOp
    self.infix[LtOp]       = parseInfixOp
    self.infix[GtOp]       = parseInfixOp
    self.infix[LeOp]       = parseInfixOp
    self.infix[GeOp]       = parseInfixOp
    self.infix[Plus]       = parseInfixOp
    self.infix[Minus]      = parseInfixOp
    self.infix[Asterisk]   = parseInfixOp
    self.infix[Slash]      = parseInfixOp
    self.infix[Percent]    = parseInfixOp
    self.infix[PlusPlus]   = parseInfixOp

proc isExprStart(self: Parser): bool =
    return self.token.kind in self.prefix

proc parseAll(self: Parser): Node =
    result = newProgram()

    dbg self, "parseAll"
    while self.token.kind != Last:
        dbg self, "parseAll loop"
        let tree = self.parseExpr()

        if tree == nil:
            panic("got null tree for program")

        if tree.kind == nkPragmaList:
            self.pragmaPool.add(tree)
        elif tree.kind != nkEmpty:
            if self.pragmaPool.len() > 0 and tree.canHavePragma():
                if tree.pragma.kind != nkPragmaList:
                    tree.pragma = newEmptyPragmaList()

                for pragma in self.pragmaPool:
                    tree.pragma.add(pragma.children)

            result.add(tree)

proc parseExpr(self: Parser; precedence: Precedence): Node =
    dbg self, "parseExpr"

    let indentErrorCode = self.token.checkIndent(self.blocks.peek())

    if indentErrorCode == 1:
        if self.blocks.len() == 1:
            self.errInvalidIndent(fmt"This token must have 0 indentation, but has {!self.token.indent()}")
        else:
            self.errInvalidIndent(fmt"This token is offside the context started at position [{self.blocks.peek().line}:{self.blocks.peek().getColumn()}], token position is [{self.token.info.dupNoLength()}]. This line will be skipped")

        self.skipLine()
        return newEmptyNode()
    elif indentErrorCode == -1:
        self.blocks.drop()

    if self.isKind(TopLevelComment):
        self.parseTestComment()
        return newEmptyNode()

    let prefixFn = self.prefix.getOrDefault(self.token.kind)

    if prefixFn == nil:
        self.errExpectedExprStart()
        return newEmptyNode()

    result = prefixFn(self)

    if self.isKind(Last):
        return

    dbg self, "parseExpr infix"

    while precedence <= precedences.getOrDefault(self.token.kind, Lowest):
        dbg self, "parseExpr infix loop"

        let notation = self.tokenNotation()
        echo "notation is ", $notation

        if notation notin {Infix, Postfix}:
            # WARNING: 'Unknown' is ignored
            hint "not an Infix or Postfix"
            break

        if self.token.isFirstInLine() and self.token.kind notin OperatorKinds:
            hint fmt"not an operator"
            break

        hint fmt"current is {precedence}, got {precedences.getOrDefault(self.token.kind, Lowest)}"

        let infixFn = self.infix.getOrDefault(self.token.kind)

        if infixFn == nil:
            hint fmt"no func to parse infix {self.token.kind}"
            break

        result = infixFn(self, result)

proc parseDef(self: Parser): Node =
    dbg self, "parseDef"

    self.skip(KwDef)
    self.checkToken(sameLine=true)

    # parse id or dot expr
    let head = self.parseIdOrDotExpr()
    head.expectKind({nkId, nkExprDotExpr})

    if head.kind == nkId:
        discard
    elif head.kind == nkExprDotExpr:
        discard

    self.checkToken(sameLine=true)
    let params = newEmptyParamList()

    for item in self.parseParen().children:
        item.expectKind({nkExprEqExpr, nkVarDecl})

        params.add block:
            if item.kind == nkVarDecl:
                newParam(item[0], item[1], nil, item[2])
            else:
                item[0].expectKind(nkId)
                newParam(item[0], nil, item[1], nil)

    let returnTypeExpr =
        if self.isKind({Eq, Last}):
            newIdNode("unit")
        else:
            self.checkToken(sameLine=true)
            self.parseTypeExpr()

    let body =
        if self.isKind(Eq):
            self.checkToken(sameLine=true)
            self.parseEqExpr()
        else:
            if not self.token.isFirstInLine() and not self.isKind(Last):
                self.errExpected(Eq)
                self.skipLine()
            newEmptyNode()

    result = newDefStmt(head, params, returnTypeExpr, body)

proc parseLet(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## LetStmt <- PragmaList? KW_LET KW_MUT? Id TypeExpr? EqExpr
    ## @end
    result = nil

    dbg self, "parseLet"

    self.skip(KwLet)
    self.checkToken(sameLine=true)

    let isMut = self.skipMaybe(KwMut)

    if isMut:
        self.checkToken(sameLine=true)

    let name = self.parseId()
    let typeExpr =
        if not self.isKind(Eq):
            self.checkToken(sameLine=true)
            self.parseTypeExpr()
        else: nil

    self.checkToken(sameLine=true)
    let valueExpr = self.parseEqExpr()

    dbg self, "parseLet after"

    result = newLetStmt(name, typeExpr, valueExpr)

proc parseReturn(self: Parser): Node =
    self.skip(KwReturn)
    result = newReturnStmt(self.parseExpr())

proc parseIdOrDotExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## IdOrDotExpr <- Id ('.' Id)*
    ## @end
    dbg self, "parseIdOrDotExpr"
    self.expected(Id)

    result = id(self.token)
    self.nextToken()

    while self.skipMaybe(Dot):
        dbg self, "parseIdOrDotExpr loop"
        result = newExprDotExpr(result, id(self.token))
        self.nextToken()

    dbg self, "parseIdOrDotExpr end"

proc parseId(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## Id <- !Keyword [_a-zA-Z] [_a-zA-Z0-9]* Skip
    ## @end
    assert(self.token.kind == Id, $self.token.kind)

    result = id(self.token)
    self.nextToken()

proc parseLit(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## Lit <-
    ## IntLit
    ## CharLit <- '\'' Char '\''
    ##
    ## CharEscape
    ##     <- '\\x' hex{2}
    ##      / '\\u' LBrace hex{1,4} ('\\' hex{1,4})* RBrace
    ## CharAscii <- [...]
    ## Char <- CharEscape / CharAscii
    ## @end
    dbg self, "parseLit"

    result = case self.token.kind:
        of KwTrue:
            newLitNode(newLit(true))
        of KwFalse:
            newLitNode(newLit(false))
        of StringLit, RawStringLit, LongStringLit, LongRawStringLit:
            newLitNode(newLit(self.token.value))
        of TypedLiteralKinds, IntLit, UIntLit, FloatLit:
            let lit = case self.token.kind:
                of IntLit, I8Lit, I16Lit, I32Lit, I64Lit  : self.getIntLit()
                of UIntLit, U8Lit, U16Lit, U32Lit, U64Lit : self.getUIntLit()
                of FloatLit, F32Lit, F64Lit               : self.getFloatLit()
                else: unreachable()
            let typedLit = case self.token.kind:
                of ISizeLit : lit.tryIntoTyped(tlkISize)
                of USizeLit : lit.tryIntoTyped(tlkISize)
                of I8Lit    : lit.tryIntoTyped(tlkI8)
                of I16Lit   : lit.tryIntoTyped(tlkI16)
                of I32Lit   : lit.tryIntoTyped(tlkI32)
                of I64Lit   : lit.tryIntoTyped(tlkI64)
                of U8Lit    : lit.tryIntoTyped(tlkU8)
                of U16Lit   : lit.tryIntoTyped(tlkU16)
                of U32Lit   : lit.tryIntoTyped(tlkU32)
                of U64Lit   : lit.tryIntoTyped(tlkU64)
                of F32Lit   : lit.tryIntoTyped(tlkF32)
                of F64Lit   : lit.tryIntoTyped(tlkF64)
                of IntLit, UIntLit, FloatLit : lit.toTypedLit()
                else: unreachable()
            newLitNode(typedLit)
        else: unimplemented()
    self.nextToken()

proc parseTypeExpr(self: Parser): Node =
    # Identifier: i32
    # Generic parameter: <T>
    # Type with generic parameters: table[string, i32]
    dbg self, "parseTypeExpr"

    case self.token.kind
    of Id:
        return self.parseId()
    of LtOp:
        let errMsg = fmt"generic parameter form is <T> without spaces around identifier"

        if self.tokenNotation() != Prefix: self.errSyntax(errMsg)
        self.skip(LtOp)

        let id = self.parseId()

        if self.tokenNotation() != Postfix: self.errSyntax(errMsg)
        self.skip(GtOp)

        return newGenericParam(id)
    else:
        self.errExpected({Id, LtOp})

    dbg self, "parseTypeExpr after"
    result = nil

proc parseSemicolon(self: Parser): Node =
    dbg self, "parseSemicolon"
    result = newParen()

proc parseSemicolon(self: Parser; left: Node): Node =
    dbg self, "parseSemicolon left"
    result = left

proc parseIfExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## IfExpr <- (KW_IF Expr DoExpr) ElifBranch* ElseBranch?
    ## ElifBranch <- KW_ELIF Expr DoExpr
    ## ElseBranch <- KW_ELSE Expr+
    ## @end
    dbg self, "parseIfExpr"

    var branches = newSeqOfCap[Node](1)

    while true:
        if branches.len() == 0:
            self.skip(KwIf)
        elif not self.skipMaybe(KwElif):
            break

        self.blocks.push(self.blockContextFromCurrentToken())

        case self.token.checkIndent(self.blocks.peek())
        of -1 : self.blocks.drop()
        of 1  : self.errInvalidBlockContext()
        else: discard

        let expr = self.parseExpr()
        self.blocks.drop()
        let body = self.parseDoExpr()

        branches &= newIfBranch(expr, body)

    dbg self, "parseIfExpr else"

    var elseBranch = nil.Node

    if self.skipMaybe(KwElse):
        elseBranch = newEmptyElseBranch()
        self.parseBlockExpr(elseBranch)

    dbg self, "parseIfExpr end"

    result = newIfExpr(branches, elseBranch)

proc parseDoExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## DoBlock <- 'do' Skip ExprList
    ## @end
    dbg self, "parseDoExpr"
    self.skip(KwDo)

    result = newDoExpr()
    self.parseBlockExpr(result)

proc parseEqExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## EqBlock <- '=' Skip ExprList
    ## @end
    dbg self, "parseEqExpr"
    self.skip(Eq)

    result = newEqExpr()
    self.parseBlockExpr(result)

proc parseBlockExpr(self: Parser; result: Node) =
    dbg self, "parseBlockExpr"

    self.blocks.push(self.blockContextFromCurrentToken())

    while self.token.kind notin {Semicolon, Last}:
        dbg self, "parseBlockExpr loop"

        # first is always true
        let checkResult = self.token.checkIndent(self.blocks.peek())

        if checkResult > 0:
            self.errInvalidBlockContext()
            self.skipLine()
            continue
        elif checkResult < 0:
            self.blocks.drop()
            break

        if not self.isExprStart() and result.len() > 0:
            break

        let expr = self.parseExpr()
        result.children.add(expr)

    dbg self, "parseBlockExpr end"

proc parseParen(self: Parser): Node =
    dbg self, "parseParen"

    self.skip(LParen)

    result = newParen()

    while self.token.kind notin {Last, RParen}:
        dbg self, "parseParen loop"

        let expr = self.parseExpr()
        result.add(expr)

        if self.skipMaybe(Semicolon):
            discard

    dbg self, "parseParen after"

    self.skip(RParen)

proc parseInfixOp(self: Parser; left: Node): Node =
    if self.token.kind notin OperatorKinds:
        hint fmt"{self.token.kind}"
        self.errUnknownOp(if self.token.kind == Id: "id " & self.token.value else: $self.token.kind)

    let precedence = precedences[self.token.kind]
    let op         = id($self.token.kind)

    self.skipToken()
    result = newInfix(op, left, self.parseExpr(precedence))

proc parseExprParen(self: Parser; left: Node): Node =
    assert(self.isKind(LParen))

    dbg self, "parseExprParen"

    result = newExprParen(left)
    result[^1] = self.parseParen()

proc parseExprEqExpr(self: Parser; left: Node): Node =
    dbg self, "parseExprEqExpr"

    result = newExprEqExpr(left, self.parseEqExpr())

    if result[1].len() == 1:
        # drop redundant 'EqExpr' node
        result[1] = result[1][0]

proc parseVarDecl(self: Parser; left: Node): Node =
    dbg self, "parseVarDecl"

    let isVarArgs = self.skipMaybe(DotDotDot)
    let typeExpr  = self.parseTypeExpr()

    result = newVarDecl(left, typeExpr)

    if isVarArgs:
        result[2] = newPragmaList(newPragma(newIdNode("VarArgParam"), nil))

    dbg self, "parseVarDecl after"

proc parsePragmaAux(self: Parser): Node =
    assert(self.isKind(Id))

    dbg self, "parsePragmaAux"

    let name = self.parseId()
    let args =
        if self.isKind(LParen): self.parseParen()
        else: nil

    dbg self, "parsePragmaAux after"

    result = newPragma(name, args)

proc parsePragma(self: Parser): Node =
    dbg self, "parsePragma"

    self.skip(Hashtag)
    self.expected({Id, LBracket})

    if spacesBefore =? self.token.spacesBefore():
        if spacesBefore != 0:
            self.errExpected({Id, Hashtag})
            return

    let pragmas: seq[Node]

    if self.isKind(Id):
        pragmas = @[self.parsePragmaAux()]
    else:
        dbg self, "parsePragma before loop"
        self.skip(LBracket)

        var tmp = newSeq[Node]()
        while self.token.kind notin {Last, RBracket}:
            dbg self, "parsePragma loop"

            if not self.isKind(Id):
                self.errExpectedId()
                self.skipLine()

            let pragma = self.parsePragmaAux()
            tmp.add(pragma)

            if self.skipMaybe(Comma):
                discard
        self.skip(RBracket)

        dbg self, "parsePragma after loop"
        pragmas = tmp

    if pragmas.len() == 0:
        self.errSyntax(fmt"empty pragma blocks are invalid")
        return

    dbg self, "parsePragma after"

    result = newPragmaList(pragmas)

proc getIntLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseBiggestInt(self.token.value))
    except ValueError:
        panic(
            fmt"invalid value '{self.token.value}' for integer literal, " &
            fmt"range is {BiggestInt.low}..{BiggestInt.high}",
            self.token.info)

proc getUIntLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseBiggestUInt(self.token.value))
    except ValueError:
        panic(
            fmt"invalid value '{self.token.value}' for unsigned integer literal, " &
            fmt"range is {BiggestUInt.low}..{BiggestUInt.high}",
            self.token.info)

proc getFloatLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseFloat(self.token.value))
    except ValueError:
        panic(fmt"try again (idk float is dumb)", self.token.info)

proc parseTestComment(self: Parser) =
    self.skip(TopLevelComment)

    for line in self.prevToken.value.splitLines():
        let cmd = line
            .split(':')
            .mapIt(it[it.findIt(it != ' ') .. it.rfindIt(it != ' ')])

        if cmd.len() == 0:
            return

        case cmd[0]
        of "SKIP":
            var i = 1
            let required =
                if cmd.len() > i and cmd[i] == "REQUIRED":
                    inc(i)
                    true
                else: false
            let count =
                if cmd.len() > i:
                    inc(i)
                    cmd[i.pred].parseInt()
                else: 1
            let kind =
                if cmd.len() > i:
                    inc(i)
                    let table {.global.} = TokenKind
                        .items()
                        .toSeq()
                        .indexBy((it: TokenKind) => it.symbolName)
                    table.getOrDefault(cmd[i.pred], Invalid)
                else: Invalid

            debug fmt"command: SKIP, {required=}, {count=}, kind={kind.symbolName}"
            for i in 0 ..< count:
                if kind == Invalid:
                    self.skipToken()

                    if self.token.kind == Last and required and i != count.pred:
                        panic(fmt"token is needed for SKIP command, but EOF is reached")
                else:
                    if not self.skipMaybe(kind) and required:
                        panic(fmt"token of kind {kind} is needed for SKIP command")
        else:
            panic(fmt"invalid command for test: '{cmd[0]}'")


grammarDocs do:
    ## ContainerDocComment <- ('///' [^\n]* Skip)+
    ## DocComment <- ('///' [^\n]* Skip)+
    ## Comment <- '//' ![!/] [^\n]* / '////' [^\n]*

grammarDocs do:
    ## KW_IF  <- 'if'
    ## KW_LET <- 'let'
    ## KW_MUT <- 'mut'
    ##
    ## Keyword
    ##     <- KW_IF
    ##      / KW_LET
    ##      / KW_MUT

when defined(jetBuildGrammar):
    static:
        writeFile("grammar.peg", getGrammar())
