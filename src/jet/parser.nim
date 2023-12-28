import std/sugar
import std/tables
import std/strformat
import std/strutils
import std/sequtils
import std/enumutils

import jet/ast
import jet/scanner
import jet/token
import jet/literal
import jet/parser/block_context
import jet/parser/precedence

import lib/stack
import lib/grammar_docs
import lib/utils
import lib/utils/line_info

import pkg/questionable



type
    Parser* = ref object
        scanner    : Scanner
        token      : Token
        prevToken  : Token
        blocks     : Stack[BlockContext]    ## Sequence of block contexts
        pragmaPool : seq[Node]
        precedence : ?Precedence

        prefix : OrderedTable[TokenKind, ParseFn]
        infix  : OrderedTable[TokenKind, InfixParseFn]

    ParserError* = object of CatchableError

    # TODO: .nimcall?
    ParseFn       = proc(self: Parser): Node
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
    KwOf       : Member, # IDK
    Eq         : Assign,
}.toTable()

proc newParser*(scanner: Scanner): Parser
proc fillTables(self: Parser)
proc parseAll*(self: Parser): Node
proc parseExpr(self: Parser): Node
proc parseDef(self: Parser): Node
proc parseTypedef(self: Parser): Node
proc parseVarDeclStmt(self: Parser): Node
proc parseReturn(self: Parser): Node
proc parseIdOrExprDotExpr(self: Parser): Node
proc parseId(self: Parser): Node
proc parseLit(self: Parser): Node
proc parseTypeExpr(self: Parser): Node
proc parseMatchExpr(self: Parser): Node
proc parseIfExpr(self: Parser): Node
proc parseIfBranch(self: Parser): Node
proc parseElseBranch(self: Parser): Node
proc parseDoExpr(self: Parser): Node
proc parseEqExpr(self: Parser): Node
proc parseParen(self: Parser): Node
proc parseBrace(self: Parser): Node
proc parsePragma(self: Parser): Node
proc parseExprParen(self: Parser; left: Node): Node
proc parseExprBrace(self: Parser; left: Node): Node
proc parseBar(self: Parser): Node

proc isTypeExprStart(self: Parser): bool
proc parseBlock(
    self: Parser;
    result: Node;
    context: BlockContext;
    until: TokenKind = Semicolon;
    fn: ParseFn = parseExpr)
proc parseList(
    self: Parser;
    result: Node;
    until: TokenKind;
    separator: TokenKind;
    fn: ParseFn = parseExpr)

proc parseInfixOp(self: Parser; left: Node): Node
proc parseExprDotExpr(self: Parser; left: Node): Node
proc parseExprEqExpr(self: Parser; left: Node): Node
proc parseVarDecl(self: Parser): Node
proc parseVarDeclNoHead(self: Parser; left: Node): Node
proc parseCommaInfix(self: Parser; left: Node): Node

proc getIntLit(self: Parser): Literal
proc getUIntLit(self: Parser): Literal
proc getFloatLit(self: Parser): Literal
proc parseTestComment(self: Parser)


when defined(jetDebugParserState):
    import std/importutils
    import lib/utils/text_style

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
else:
    template dbg(self: Parser; msg: string = "") = discard

proc checkIndent(token: Token; context: BlockContext): int
proc nextToken(self: Parser; checkIndent: bool = true)


# ----- ERRORS ----- #
{.push, used, noreturn.}

proc err(self: Parser; msg: string) =
    error(msg, self.token.info)
    raise newException(ParserError, msg)

proc errSyntax(self: Parser; msg: string) =
    self.err(fmt"Syntax error: {msg}")

proc errExpectedId(self: Parser) =
    self.err(fmt"expected identifier, got {self.token.kind}")

proc errExpectedLit(self: Parser) =
    self.err(fmt"expected literal, got {self.token.kind}")

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

proc errInvalidBlockContext(self: Parser; context: BlockContext) =
    self.errInvalidIndent(
        fmt"This token is offside the context started at position [{context.line}:{context.getColumn()}], " &
        fmt"token position is [{self.token.info.dupNoLength()}]. This line will be skipped")

proc errInvalidBlockContext(self: Parser) =
    self.errInvalidBlockContext(self.blocks.peek())

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

proc errUnknownOp(self: Parser; op, explanation: string) =
    self.err(fmt"Unknown operator: '{op}'. {explanation}")

proc errUnknownOp(self: Parser; op: string) =
    self.err(fmt"Unknown operator: '{op}'")

{.pop.} # used, noreturn


# ----- PRIVATE ----- #
template isKind(self: Parser; tokenKind: TokenKind): bool = self.token.kind == tokenKind
template isKind(self: Parser; tokenKinds: set[TokenKind]): bool = self.token.kind in tokenKinds

proc isSameLine(self: Parser): bool =
    return self.prevToken.info.line == self.token.info.line

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
    let token = !self.scanner.getToken()

    self.prevToken = self.token
    self.token     = token

proc tokenSameLine(self: Parser) =
    if not self.isSameLine():
        self.errExpectedSameLine()

proc tokenFirstInLine(self: Parser) =
    if not self.token.isFirstInLine(): self.errExpectedFirstInLine()

proc tokenLastInLine(self: Parser) =
    if not self.token.isLastInLine(): self.errExpectedLastInLine()

proc tokenIndent(self: Parser; expectedIndent: Natural) =
    without indent =? self.token.indent(): self.errInvalidIndent()
    if indent != expectedIndent: self.errInvalidIndent()

proc blockContextFromCurrentToken(self: Parser; allowSmallerIndent: bool = false): BlockContext =
    let column: int
    let kind: BlockContextKind

    if self.token.isFirstInLine():
        column = !self.token.indent()
        kind   = Indent

        if column <= self.blocks.peek() and not allowSmallerIndent:
            self.errInvalidBlockContext()
    else:
        column = self.token.info.column.int
        kind   = Column

    result = initBlockContext(kind, self.token.info.line.int, column)

    hint fmt"new block context created: {result}"

proc checkToken(
    self        : Parser;
    notation    : set[Notation] = {};
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
    if firstInLine:
        self.tokenFirstInLine()
        check()
    if lastInLine:
        self.tokenLastInLine()
        check()
    if notation != {} and self.tokenNotation() notin notation:
        self.errInvalidNotation()
    if expectedIndent =? indent:
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
    self.prefix[KwDef]     = parseDef
    self.prefix[KwTypeDef] = parseTypedef
    self.prefix[KwLet]     = parseVarDeclStmt
    self.prefix[KwMut]     = parseVarDeclStmt
    self.prefix[KwVal]     = parseVarDeclStmt
    self.prefix[KwMatch]   = parseMatchExpr
    self.prefix[KwIf]      = parseIfExpr
    self.prefix[KwDo]      = parseDoExpr
    self.prefix[KwReturn]  = parseReturn
    self.prefix[Id]        = parseId
    self.prefix[Eq]        = parseEqExpr
    self.prefix[Hashtag]   = parsePragma
    self.prefix[LParen]    = parseParen
    self.prefix[LBrace]    = parseBrace
    self.prefix[Bar]       = parseBar

    self.prefix[IntLit]           = parseLit
    self.prefix[UIntLit]          = parseLit
    self.prefix[FloatLit]         = parseLit
    self.prefix[CharLit]          = parseLit
    self.prefix[StringLit]        = parseLit
    self.prefix[RawStringLit]     = parseLit
    self.prefix[LongStringLit]    = parseLit
    self.prefix[LongRawStringLit] = parseLit
    self.prefix[ISizeLit]         = parseLit
    self.prefix[USizeLit]         = parseLit
    self.prefix[I8Lit]            = parseLit
    self.prefix[I16Lit]           = parseLit
    self.prefix[I32Lit]           = parseLit
    self.prefix[I64Lit]           = parseLit
    self.prefix[U8Lit]            = parseLit
    self.prefix[U16Lit]           = parseLit
    self.prefix[U32Lit]           = parseLit
    self.prefix[U64Lit]           = parseLit
    self.prefix[F32Lit]           = parseLit
    self.prefix[F64Lit]           = parseLit
    self.prefix[KwTrue]           = parseLit
    self.prefix[KwFalse]          = parseLit

    # self.infix[Comma]      = parseCommaInfix
    self.infix[LParen]     = parseExprParen
    self.infix[LBrace]     = parseExprBrace
    self.infix[Dot]        = parseExprDotExpr
    self.infix[Eq]         = parseExprEqExpr
    self.infix[Id]         = parseVarDeclNoHead
    self.infix[DotDotDot]  = parseVarDeclNoHead
    self.infix[DotDot]     = parseInfixOp
    self.infix[DotDotLess] = parseInfixOp
    self.infix[KwAnd]      = parseInfixOp
    self.infix[KwOr]       = parseInfixOp
    self.infix[KwOf]       = parseInfixOp
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

    dbg self, "parseAll end"

proc parseExpr(self: Parser): Node =
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

    result = prefixFn(self)

    if self.isKind(Last):
        return

    dbg self, "parseExpr infix"

    while self.precedence.isNone() or !self.precedence < precedences.getOrDefault(self.token.kind, Lowest):
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

        hint fmt"current is {self.precedence |? Lowest}, got {precedences.getOrDefault(self.token.kind, Lowest)}"

        let infixFn = self.infix.getOrDefault(self.token.kind)

        if infixFn == nil:
            hint fmt"no func to parse infix {self.token.kind}"
            break

        result = infixFn(self, result)

    self.precedence = none(Precedence)

proc parseDef(self: Parser): Node =
    dbg self, "parseDef"

    self.skip(KwDef)
    self.checkToken(sameLine=true)

    # parse id or dot expr
    let head = self.parseIdOrExprDotExpr()
    head.expectKind({nkId, nkExprDotExpr})

    if head.kind == nkId:
        discard
    elif head.kind == nkExprDotExpr:
        discard

    let params = newEmptyParamList()

    self.checkToken(sameLine=true)
    self.skip(LParen)
    self.parseList(params, RParen, Semicolon, parseVarDecl)
    self.skip(RParen)

    # for item in self.parseParen().children:
    #     item.expectKind({nkExprEqExpr, nkVarDecl})

    #     params.add block:
    #         if item.kind == nkVarDecl:
    #             newParam(item[0], item[1], nil, item[2])
    #         else:
    #             item[0].expectKind(nkId)
    #             newParam(item[0], nil, item[1], nil)

    let returnTypeExpr =
        if self.isKind({Eq, Last}):
            newEmptyNode()
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
            newEmptyNode()

    result = newDefStmt(head, params, returnTypeExpr, body)

proc parseTypedef(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## TypedefStmt = KW_TYPEDEF Id EqExpr
    ## @end
    self.skip(KwTypedef)

    self.checkToken(sameLine=true)
    let name = self.parseId()

    self.checkToken(sameLine=true)
    let body = self.parseEqExpr()

    result = newTypedefStmt(name, body)

proc parseVarDeclStmt(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## VarDeclStmt <- (KW_LET / KW_MUT / KW_VAL) VarDecl
    ## @end
    result = nil

    dbg self, "parseLet"

    # TODO: flags for 'mut' and 'val'
    self.skip({KwLet, KwMut, KwVal})
    self.checkToken(sameLine=true)

    result = self.parseVarDecl()

proc parseReturn(self: Parser): Node =
    dbg self, "parseReturn"

    self.skip(KwReturn)
    self.checkToken(sameLine=true)

    dbg self, "parseReturn end"

    result = newReturnStmt(self.parseExpr())

proc parseIdOrExprDotExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## IdOrExprDotExpr <- Id (DOT Id)*
    ## @end
    dbg self, "parseIdOrExprDotExpr"
    self.expected(Id)

    result = id(self.token)
    self.nextToken()

    while self.skipMaybe(Dot):
        dbg self, "parseIdOrExprDotExpr loop"
        result = newExprDotExpr(result, id(self.token))
        self.nextToken()

    dbg self, "parseIdOrExprDotExpr end"

proc parseId(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## Id <- !Keyword [_a-zA-Z] [_a-zA-Z0-9]* Skip
    ## @end
    if self.token.kind != Id:
        self.errExpectedId()

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
        of TypedLiteralKinds, IntLit, UIntLit, FloatLit, CharLit:
            let lit = case self.token.kind:
                of IntLit, I8Lit, I16Lit, I32Lit, I64Lit  : self.getIntLit()
                of UIntLit, U8Lit, U16Lit, U32Lit, U64Lit : self.getUIntLit()
                of FloatLit, F32Lit, F64Lit               : self.getFloatLit()
                of CharLit: newLit(self.token.value[0]) # TOOBAD: why only 1
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
                of CharLit  : lit.tryIntoTyped(tlkChar)
                of IntLit, UIntLit, FloatLit : lit.toTypedLit()
                else: unreachable()
            newLitNode(typedLit)
        else:
            self.errExpectedLit()
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
    of DotDotDot:
        self.checkToken(sameLine=true, notation={Prefix})
        self.skipToken()
        self.checkToken(sameLine=true)
        result = newPrefix(id"...", self.parseTypeExpr())
    else:
        self.errExpected({Id, LtOp, DotDotDot})

    dbg self, "parseTypeExpr after"
    result = nil

proc parseMatchExpr(self: Parser): Node =
    self.skip(KwMatch)

    let expr  = self.parseExpr()
    var cases = newSeq[Node]()
    var elseCase = nil.Node

    # TODO: block context
    while self.isKind(Bar):
        if elseCase != nil:
            self.err(fmt"'else' case must be last case in 'match' expression")

        let branch = self.parseBar()

        if branch.kind == nkElseBranch:
            elseCase = branch
        else:
            cases &= branch

    result = newMatchExpr(expr, cases, elseCase)

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
        if branches.len() > 0 and not self.isKind(KwElif):
            break

        branches &= self.parseIfBranch()

        # if branches.len() == 0:
        #     self.skip(KwIf)
        # elif not self.skipMaybe(KwElif):
        #     break

        # self.blocks.push(self.blockContextFromCurrentToken())

        # case self.token.checkIndent(self.blocks.peek())
        # of -1 : self.blocks.drop()
        # of 1  : self.errInvalidBlockContext()
        # else: discard

        # let expr = self.parseExpr()
        # self.blocks.drop()
        # let body = self.parseDoExpr()

        # branches &= newIfBranch(expr, body)

    dbg self, "parseIfExpr else"

    let elseBranch =
        if self.isKind(KwElse): self.parseElseBranch()
        else: nil

    dbg self, "parseIfExpr end"

    result = newIfExpr(branches, elseBranch)

proc parseIfBranch(self: Parser): Node =
    self.skip({KwIf, KwElif})

    self.blocks.push(self.blockContextFromCurrentToken())

    case self.token.checkIndent(self.blocks.peek())
    of -1 : self.blocks.drop()
    of 1  : self.errInvalidBlockContext()
    else: discard

    let expr = self.parseExpr()
    self.blocks.drop()

    result = newIfBranch(expr, self.parseDoExpr())

proc parseElseBranch(self: Parser): Node =
    dbg self, "parseElseBranch"

    self.skip(KwElse)

    result = newEmptyElseBranch()
    self.parseBlock(result, self.blockContextFromCurrentToken())

    dbg self, "parseElseBranch end"

proc parseDoExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## DoBlock <- 'do' Skip ExprList
    ## @end
    dbg self, "parseDoExpr"
    self.skip(KwDo)

    result = newDoExpr()
    self.parseBlock(result, self.blockContextFromCurrentToken())

    # if self.isKind(Semicolon):
    #     self.checkToken(sameLine=true)
    #     self.skip(Semicolon)

proc parseEqExpr(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## EqBlock <- '=' Skip ExprList
    ## @end
    dbg self, "parseEqExpr"
    self.skip(Eq)

    result = newEqExpr()
    let context = self.blockContextFromCurrentToken()

    self.parseBlock(result, context)

proc parseParen(self: Parser): Node =
    dbg self, "parseParen"

    self.skip(LParen)
    result = self.parseExpr()
    self.skip(RParen)

    dbg self, "parseParen end"

proc parseBrace(self: Parser): Node =
    dbg self, "parseBrace"

    result = newBrace()

    self.skip(LBrace)
    self.parseList(result, RBrace, Semicolon, parseExpr)
    self.skip(RBrace)

    dbg self, "parseBrace end"

proc parseInfixOp(self: Parser; left: Node): Node =
    if self.token.kind notin OperatorKinds + {KwAnd, KwOr, KwOf}:
        self.errUnknownOp(if self.token.kind == Id: "id " & self.token.value else: $self.token.kind)

    let op = id($self.token.kind)

    self.precedence = some(precedences[self.token.kind])
    self.skipToken()
    result = newInfix(op, left, self.parseExpr())

proc parseExprParen(self: Parser; left: Node): Node =
    dbg self, "parseExprParen"

    result = newExprParen(left)

    self.skip(LParen)
    self.parseList(result[^1], RParen, Comma)
    self.skip(RParen)

proc parseExprBrace(self: Parser; left: Node): Node =
    dbg self, "parseExprParen"

    result = newExprBrace(left)
    result[^1] = self.parseBrace()

proc parseExprEqExpr(self: Parser; left: Node): Node =
    dbg self, "parseExprEqExpr"

    self.checkToken(notation={Infix})
    self.skip(Eq)
    self.checkToken(sameLine=true)
    self.precedence = some(Assign)

    dbg self, "parseExprEqExpr end"

    result = newExprEqExpr(left, self.parseExpr())

    # if result[1].len() == 1:
    #     # drop redundant 'EqExpr' node
    #     result[1] = result[1][0]

proc parseExprDotExpr(self: Parser; left: Node): Node =
    dbg self, "parseExprDotExpr"

    self.checkToken(notation={Infix, Prefix})
    self.skip(Dot)
    self.checkToken(sameLine=true)
    self.precedence = some(Member)

    dbg self, "parseExprDotExpr end"

    result = newExprDotExpr(left, self.parseExpr())

proc parseBar(self: Parser): Node =
    self.skip(Bar)

    case self.token.kind
    of KwElse:
        result = self.parseElseBranch()
        return
    of KwElif:
        result = self.parseIfBranch()
        return
    else:
        result = self.parseExpr()

    if result.kind == nkInfix:
        if result[0].id != "of":
            self.errExpected(KwOf)

        result = newVariant(result)

    case self.token.kind
    of Last, Bar:
        discard
    of KwIf:
        result = newMatchCase(result, self.parseIfBranch())
    of KwDo:
        result = newMatchCase(result, self.parseDoExpr())
    else:
        unimplemented(fmt"parseBar for {self.token.kind}")

    if result.kind notin {nkVariant, nkCase, nkElseBranch}:
        result = newVariant(result)

proc parseVarDecl(self: Parser): Node
    {.grammarDocs.} =
    ## @grammar
    ## VarDecl <- Id (COMMA Id)* TypeExpr? EqExpr?
    ## @end
    dbg self, "parseVarDecl"

    var names = @[self.parseId()]

    while self.skipMaybe(Comma):
        self.checkToken(sameLine=true)
        names &= self.parseId()

    let typeExpr =
        if self.isTypeExprStart():
            self.checkToken(sameLine=true)
            self.parseTypeExpr()
        else: nil

    let eqExpr =
        if self.isKind(Eq):
            self.checkToken(sameLine=true)
            self.parseEqExpr()
        else: nil

    result = newVarDecl(names, typeExpr, eqExpr)

    dbg self, "parseVarDecl end"

proc parseVarDeclNoHead(self: Parser; left: Node): Node =
    dbg self, "parseVarDeclNoHead"

    var names = @[left]

    while self.skipMaybe(Comma):
        self.checkToken(sameLine=true)
        names &= self.parseId()

    var typeExpr  = self.parseTypeExpr()
    var isVarArgs = false

    if typeExpr.kind == nkPrefix:
        if typeExpr[0].id != "...":
            self.errUnknownOp(typeExpr[0].id)

        typeExpr  = typeExpr[1]
        isVarArgs = true

    let eqExpr =
        if self.isKind(Eq):
            self.checkToken(sameLine=true)
            self.parseEqExpr()
        else: nil

    result = newVarDecl(names, typeExpr, eqExpr)

    if isVarArgs:
        result[2] = newPragmaList(newPragma(id"VarArgParam", nil))

    dbg self, "parseVarDeclNoHead end"

proc parseCommaInfix(self: Parser; left: Node): Node =
    result = case left.kind:
        of nkId:
            self.parseVarDeclNoHead(left)
        else:
            # 'Comma' token can be used only in parenthesis
            self.skip(Comma)
            left

proc parsePragmaAux(self: Parser): Node =
    assert(self.isKind(Id))

    dbg self, "parsePragmaAux"

    let name = self.parseId()
    let args = newParen()

    if self.skipMaybe(LParen):
        self.parseList(args, RParen, Comma)
        self.skip(RParen)

    dbg self, "parsePragmaAux after"

    result = newPragma(name, args)

proc parsePragma(self: Parser): Node =
    dbg self, "parsePragma"

    self.skip(Hashtag)
    self.expected({Id, LBracket})

    if spacesBefore =? self.token.spacesBefore():
        if spacesBefore != 0:
            self.errExpected({Id, Hashtag})

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

            let pragma = self.parsePragmaAux()
            tmp.add(pragma)

            if self.skipMaybe(Comma):
                discard
        self.skip(RBracket)

        dbg self, "parsePragma after loop"
        pragmas = tmp

    if pragmas.len() == 0:
        self.errSyntax(fmt"empty pragma blocks are invalid")

    dbg self, "parsePragma after"

    result = newPragmaList(pragmas)

proc isTypeExprStart(self: Parser): bool =
    result = self.isKind({Id, LtOp, DotDotDot})

proc parseBlock(self: Parser; result: Node; context: BlockContext; until: TokenKind; fn: ParseFn) =
    dbg self, "parseBlock"

    self.blocks.push(context)

    while self.token.kind notin {Last, until}:
        dbg self, "parseBlock loop"

        case self.token.checkIndent(self.blocks.peek())
        of 1:
            self.errInvalidBlockContext()
        of -1:
            break
        else:
            let node = fn(self)

            if node.kind != nkEmpty:
                result &= node

    # discard self.skipMaybe(until)
    self.blocks.drop()

    dbg self, "parseBlock end"

proc parseList(self: Parser; result: Node; until, separator: TokenKind; fn: ParseFn) =
    dbg self, "parseList"

    let allowSmallerIndent = self.blocks.peek().kind == Column
    let context            = self.blockContextFromCurrentToken(allowSmallerIndent)
    let isSmallerIndent    = context.getColumn() < self.blocks.peek().getColumn()

    self.parseBlock(result, context, until) do(this: Parser) -> Node:
        result =
            if not this.skipMaybe(separator):
                fn(this)
            else:
                newEmptyNode()

    if allowSmallerIndent and isSmallerIndent:
        #| ... = {
        #|     ...
        #| }
        #  ^ closing token must be located here
        let context = self.blocks.at(1)
        case self.token.checkIndent(context)
        of 0: discard
        else: self.errInvalidBlockContext(context)

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

# TODO: generate it from 'TokenKind' enum
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
