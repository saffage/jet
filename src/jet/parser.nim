import std/sugar
import std/tables
import std/strformat
import std/strutils
import std/sequtils
import std/options

import jet/ast
import jet/lexer
import jet/token
import jet/literal
import jet/line_info

import lib/stack
import lib/utils

import pkg/questionable


type
    BlockContext = tuple[line, column: int]

    Precedence = enum
        Lowest
        Assign
        Or
        And
        Cmp
        Postfix
        Sum
        Product
        Call
        Index
        Member
        Highest

    Parser* = ref object
        lexer      : Lexer
        curr       : Token
        prev       : Token
        blocks     : Stack[BlockContext]
        precedence : Option[Precedence]
        prefix     : OrderedTable[TokenKind, ParseFn]
        infix      : OrderedTable[TokenKind, InfixParseFn]

    ParserError* = object of CatchableError
        lineInfo : LineInfo

    ParseFn       = proc(self: Parser): Node {.nimcall.}
    InfixParseFn  = proc(self: Parser; left: Node): Node {.nimcall.}

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
    Assign     : Assign,
}.toTable()

proc newParser*(lexer: sink Lexer): Parser
proc fillTables(self: Parser)
proc parseAll*(self: Parser): Node
proc parseExpr(self: Parser): Node
proc parseId(self: Parser): Node
proc parseLit(self: Parser): Node
proc parseAnnotation(self: Parser): Node
proc parseNegation(self: Parser): Node

proc parseInfixOp(self: Parser; left: Node): Node
proc parseArgs(self: Parser; left: Node): Node

proc expectSameIndent(self: Parser)


when defined(jetDebugParserState):
    import std/importutils
    import lib/utils/text_style

    proc dbg(self: Parser; msg: string = "") =
        const dbgStyle = TextStyle(foreground: Cyan, underlined: true)
        let msg = if msg == "": "" else: (msg @ dbgStyle) & ": "

        privateAccess(Lexer)
        debug(
            fmt"{msg}Parser state:" &
            fmt("\n\tprev: {self.prev.human()}") &
            fmt("\n\tcurr: {self.curr.human()}") &
            fmt("\n\tscanner:") &
                fmt("\n\t\tprev: {self.lexer.prev.human()}") &
                fmt("\n\t\tcurr: {self.lexer.curr.human()}") &
            fmt("\n\tblocks: {$self.blocks}")
        )
else:
    template dbg(self: Parser; msg: string = "") = discard

proc checkIndent(token: Token; context: BlockContext): int
proc nextToken(self: Parser; checkIndent: bool = true)

proc isNewBlockContext(self: Parser; context: BlockContext): bool =
    result = context.column > self.blocks.peek().column


# ----- ERRORS ----- #
{.push, used, noreturn.}

proc err(self: Parser; msg: string; info: LineInfo) =
    error(msg, info)
    raise (ref ParserError)(msg: msg, lineInfo: info)

proc err(self: Parser; msg: string) =
    self.err(msg, self.curr.info)

proc errSyntax(self: Parser; msg: string) =
    self.err(fmt"syntax error; {msg}")

proc errExpectedId(self: Parser) =
    self.err(fmt"expected identifier, got {self.curr.kind}")

proc errExpectedLit(self: Parser) =
    self.err(fmt"expected literal, got {self.curr.kind}")

proc errExpectedExprStart(self: Parser) =
    self.errSyntax(fmt"token '{self.curr.kind}' is not an expression start")

proc errExpectedNodeOf(self: Parser; kind: NodeKind) =
    self.errSyntax(fmt"expected node of kind {kind}, got {self.curr.kind} instead")

proc errExpectedNodeOf(self: Parser; kinds: set[NodeKind]) =
    self.errSyntax(fmt"expected node of kinds {kinds}, got {self.curr.kind} instead")

proc errExpected(self: Parser; kind: TokenKind) =
    self.errSyntax(fmt"expected token {kind}, got {self.curr.kind} instead")

proc errExpected(self: Parser; kinds: set[TokenKind]) =
    let kinds = kinds.mapIt($it).join(" or ")
    self.errSyntax(fmt"expected token {kinds}, got {self.curr.kind} instead")

proc errExpectedSameLine(self: Parser) =
    self.errSyntax(fmt"expected expression on the same line")

proc errInvalidIndent(self: Parser; explanation: string) =
    self.errSyntax(fmt"invalid indentation; {explanation}")

proc errInvalidIndent(self: Parser) =
    self.errSyntax(fmt"invalid indentation")

proc errInvalidBlockContext(self: Parser; context: BlockContext) =
    self.errInvalidIndent(
        fmt"this token is offside the context started at [{context[0]}:{context[1]}], " &
        fmt"token position is [{self.curr.info.dupNoLength()}]")

proc errInvalidBlockContext(self: Parser) =
    self.errInvalidBlockContext(self.blocks.peek())

proc errInvalidNotation(self: Parser; explanation: string) =
    self.errSyntax(fmt"invalid notation; {explanation}")

proc errInvalidNotation(self: Parser) =
    self.errSyntax(fmt"invalid notation")

proc errExpectedFirstInLine(self: Parser; explanation: string) =
    self.errSyntax(fmt"token {self.curr} must be first in line. {explanation}")

proc errExpectedFirstInLine(self: Parser) =
    self.errSyntax(fmt"token {self.curr} must be first in line")

proc errExpectedLastInLine(self: Parser; explanation: string) =
    self.errSyntax(fmt"token {self.curr} must be last in line. {explanation}")

proc errExpectedLastInLine(self: Parser) =
    self.errSyntax(fmt"token {self.curr} must be last in line")

proc errUnknownOp(self: Parser; op, explanation: string) =
    self.err(fmt"Unknown operator: '{op}'. {explanation}")

proc errUnknownOp(self: Parser; op: string) =
    self.err(fmt"Unknown operator: '{op}'")

{.pop.} # used, noreturn


# ----- PRIVATE ----- #
template isKind(self: Parser; tokenKind: TokenKind): bool = self.curr.kind == tokenKind
template isKind(self: Parser; tokenKinds: set[TokenKind]): bool = self.curr.kind in tokenKinds

proc isSameLine(self: Parser): bool =
    return self.prev.info.line == self.curr.info.line

proc expected(self: Parser; kind: TokenKind) =
    if not self.isKind(kind):
        self.errSyntax(fmt"expected '{kind}', got '{self.curr.kind}'")

proc expected(self: Parser; kinds: set[TokenKind]) =
    if not self.isKind(kinds):
        let kindsStr = kinds.mapIt(fmt"'{it}'").join(", ")
        self.errSyntax(fmt"expected one of {kindsStr}, got '{self.curr.kind}'")

proc tokenNotation(self: Parser): Notation =
    return self.curr.notation(self.prev.kind, self.lexer.curr.kind)

proc skipToken(self: Parser) =
    debug fmt"token {self.curr.kind} at {self.curr.info} was skipped"
    self.curr = self.lexer.getToken().get()

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

    while self.curr.kind != Last:
        let token = self.lexer.getToken().get()

        if token.info.line > line:
            self.curr = token
            break

        debug fmt"token {token.kind} at {token.info} was skipped"

    dbg self, fmt"skipLine {line} end"

proc skipLine(self: Parser) =
    self.skipLine(self.curr.info.line)

proc checkIndent(token: Token; context: BlockContext): int =
    ## **Returns:**
    ## - -1 if `token.indent < context` - drop block
    ## - 1 if `token.indent > context` - error
    ## - 0 if `token.indent == context` - ok
    result = 0

    if indent =? token.indent():
        if indent > context.column:
            return 1
        elif indent < context.column:
            return -1
        else:
            return 0

proc nextToken(self: Parser; checkIndent: bool) =
    let token = self.lexer.getToken().get()

    self.prev = self.curr
    self.curr     = token

proc tokenSameLine(self: Parser) =
    if not self.isSameLine():
        self.errExpectedSameLine()

proc tokenFirstInLine(self: Parser) =
    if not self.curr.isFirstInLine(): self.errExpectedFirstInLine()

proc tokenLastInLine(self: Parser) =
    if not self.curr.isLastInLine(): self.errExpectedLastInLine()

proc tokenIndent(self: Parser; expectedIndent: Natural) =
    without indent =? self.curr.indent(): self.errInvalidIndent()
    if indent != expectedIndent: self.errInvalidIndent()

proc blockContextFromCurrentToken(self: Parser): BlockContext =
    result =
        if self.curr.isFirstInLine():
            (self.curr.info.line.int, !self.curr.indent())
        else:
            (self.curr.info.line.int, self.curr.info.column.int)

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
        let expected = notation.mapIt($it).join(" or ")
        self.errInvalidNotation(fmt"expected {expected}, got {self.tokenNotation()}")
    if expectedIndent =? indent:
        self.tokenIndent(expectedIndent)
        check()


proc getIntLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseBiggestInt(self.curr.value))
    except ValueError:
        panic(
            fmt"invalid value '{self.curr.value}' for integer literal, " &
            fmt"range is {BiggestInt.low}..{BiggestInt.high}",
            self.curr.info)

proc getUIntLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseBiggestUInt(self.curr.value))
    except ValueError:
        panic(
            fmt"invalid value '{self.curr.value}' for unsigned integer literal, " &
            fmt"range is {BiggestUInt.low}..{BiggestUInt.high}",
            self.curr.info)

proc getFloatLit(self: Parser): Literal =
    {.warning[ProveInit]: off.}

    try:
        result = newLit(parseFloat(self.curr.value))
    except ValueError:
        panic(fmt"try again (idk float is dumb)", self.curr.info)




type ParseMode = enum
    Block
    List
    Auto

# NEW PROCS
proc parseBlock(
    self: Parser;
    blockTree: var Node;
    mode: ParseMode = Block;
    until: Option[TokenKind] = none(TokenKind);
    context: Option[BlockContext] = none(BlockContext);
    fn: ParseFn = parseExpr;
): ParseMode {.discardable.} =
    ## Parses a block of code with the same indentation until it encounters
    ## a token with an indentation different from the current context of the block.
    ##
    ## If a token with an indentation smaller than the indentation of the current block
    ## context is found, parsing of this block will be completed.
    ##
    ## Expressions can be separated with `;`, but the next expression must be immediately after it.
    ##
    ## If no block context is given, creates its own context when the token with indentation is found.
    ## Does nothing with the `until` token, only parses the block until it.
    dbg self, fmt"parseBlock mode = {mode}"

    if mode == Block and not self.curr.isFirstInLine():
        self.err("expected block of code")

    var contextPushed = false
    var wasSemicolon  = false

    result = mode

    if context.isSome():
        self.blocks.push(context.unsafeGet())
        contextPushed = true

    let lastKinds =
        if untilKind =? until: {Last, untilKind}
        else: {Last}

    while self.curr.kind notin lastKinds:
        dbg self, "parseBlock loop"

        if contextPushed:
            hint fmt"check indentation for block at {self.blocks.peek()}"

            case self.curr.checkIndent(self.blocks.peek())
            of 1: self.errInvalidBlockContext()
            of -1: break
            else: hint fmt"indentation is ok: {self.curr.indent() |? -1}"

        if self.curr.isFirstInLine():
            if result == Block and wasSemicolon:
                self.err("expected expression after ';'")
            if not contextPushed:
                let newContext = self.blockContextFromCurrentToken()
                if not self.isNewBlockContext(newContext):
                    self.err(
                        fmt"a new block context was expected, got {newContext}, " &
                        fmt"which is the same or lower than the previous {self.blocks.peek()}")
                self.blocks.push(newContext)
                contextPushed = true
        else:
            if mode == Block and not wasSemicolon:
                self.err("the other expression must be on a new line or separated by ';'")
            wasSemicolon = false

        let node = fn(self)
        if node.kind != nkEmpty:
            blockTree &= node

        if result == Auto:
            result = if self.isKind(Comma): List else: Block
            hint fmt"determine mode of block parsing: {result}"

        case result
        of Block:
            if self.skipMaybe(Semicolon):
                wasSemicolon = true
        of List:
            if not self.skipMaybe(Comma):
                if self.curr.kind notin lastKinds:
                    self.err(fmt"expected ',' after expression", self.prev.info)
                break
        else: discard

    if result == Auto:
        # something like `()`
        result = List

    if result == Block and wasSemicolon:
        self.err("expected expression after ';'")

    if contextPushed:
        self.blocks.drop()

    dbg self, "parseBlock end"

proc parseIf(self: Parser): Node
proc parseIfBranch(self: Parser): Node
proc parseElse(self: Parser): Node
proc parseList(self: Parser): Node
proc parseType(self: Parser): Node
proc parseTypeExpr(self: Parser): Node

proc parseVarDecl(self: Parser; left: Node): Node

proc parseDo(self: Parser): Node
proc parseDoOrBlock(self: Parser): Node
proc parseExprOrBlock(self: Parser): Node



# ----- API IMPL ----- #
proc newParser(lexer: sink Lexer): Parser =
    result = Parser(
        lexer: ensureMove(lexer),
        blocks: newStack[BlockContext]())

    result.fillTables()
    result.nextToken()

proc fillTables(self: Parser) =
    self.prefix[Id]       = parseId
    self.prefix[KwIf]     = parseIf
    self.prefix[KwDo]     = parseDo
    self.prefix[KwType]   = parseType
    self.prefix[Minus]    = parseNegation
    self.prefix[LBracket] = parseList
    self.prefix[LParen]   = parseList

    self.prefix[IntLit]           = parseLit
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

    self.infix[LParen]       = parseArgs
    self.infix[Id]           = parseVarDecl
    self.infix[LBracket]     = parseVarDecl
    self.infix[QuestionMark] = parseVarDecl
    self.infix[Ampersand]    = parseVarDecl

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

proc parseAll(self: Parser): Node =
    dbg self, "parseAll"

    result = newProgram()
    let fileBlockContext = (0, 0)

    try:
        discard self.parseBlock(result, context = some(fileBlockContext), until = some(Last))
    except ParserError as e:
        let
            line       = e.lineInfo.line.int
            column     = e.lineInfo.column.int
            length     = max(1, e.lineInfo.length.int)
            linePrefix = $line
            lineStr    = block:
                let returnedLine = self.lexer.getLine(line)
                if returnedLine.isNone(): "<end-of-file>"
                else: returnedLine.get()

        echo &" {linePrefix} |{lineStr}\n {spaces(linePrefix.len())} |{spaces(column)}{repeat('^', length)} {e.msg}"
        echo "# --- partialy-generated AST"
        echo result.treeRepr
        raise e

    dbg self, "parseAll end"

proc parseExpr(self: Parser): Node =
    dbg self, "parseExpr"

    let indentErrorCode = self.curr.checkIndent(self.blocks.peek())

    if indentErrorCode == 1:
        if self.blocks.len() == 1:
            self.errInvalidIndent(fmt"This token must have 0 indentation, but has {!self.curr.indent()}")
        else:
            self.errInvalidIndent(
                fmt"This token is offside the context started at position " &
                fmt"[{self.blocks.peek().line}:{self.blocks.peek().column}], " &
                fmt"token position is [{self.curr.info.dupNoLength()}]. This line will be skipped")

        self.skipLine()
        return newEmptyNode()
    elif indentErrorCode == -1:
        self.blocks.drop()

    let prefixFn = self.prefix.getOrDefault(self.curr.kind)

    if prefixFn == nil:
        self.errExpectedExprStart()

    result = prefixFn(self)

    if self.isKind(Last):
        return

    dbg self, "parseExpr infix"

    while self.precedence.isNone() or
          !self.precedence < precedences.getOrDefault(self.curr.kind, Lowest):
        dbg self, "parseExpr infix loop"

        # let notation = self.currNotation()
        # hint fmt"notation is {notation}"

        # if notation notin {Infix, Postfix}:
        #     # WARNING: 'Unknown' is ignored
        #     hint "not an Infix or Postfix"
        #     break

        if self.curr.isFirstInLine() and not self.curr.kind.isOperator():
            hint fmt"not an operator"
            break

        hint fmt"current is {self.precedence |? Lowest}, got {precedences.getOrDefault(self.curr.kind, Lowest)}"

        let infixFn = self.infix.getOrDefault(self.curr.kind)

        if infixFn == nil:
            hint fmt"{self.curr.kind} is not infix"
            break

        result = infixFn(self, result)

    self.precedence = none(Precedence)

proc parseId(self: Parser): Node =
    dbg self, "parseId"

    self.skip({Id, Underscore})
    result = id(self.prev.value)

proc parseLit(self: Parser): Node =
    dbg self, "parseLit"

    result = case self.curr.kind:
        of KwTrue:
            newLitNode(newLit(true))
        of KwFalse:
            newLitNode(newLit(false))
        of StringLiteralKinds:
            newLitNode(newLit(self.curr.value))
        of TypedLiteralKinds, UntypedLiteralKinds, CharLit:
            let lit = case self.curr.kind:
                of IntLit, I8Lit, I16Lit, I32Lit, I64Lit : self.getIntLit()
                of U8Lit, U16Lit, U32Lit, U64Lit         : self.getUIntLit()
                of FloatLit, F32Lit, F64Lit              : self.getFloatLit()
                of CharLit: newLit(self.curr.value[0]) # TOOBAD: why only 1
                else: unreachable()
            let typedLit = case self.curr.kind:
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
                of IntLit, FloatLit : lit.toTypedLit()
                else: unreachable()
            newLitNode(typedLit)
        else:
            self.errExpectedLit()
    self.nextToken()

proc parseNegation(self: Parser): Node =
    dbg self, "parseNegation"

    self.skip(Minus)

    result = newPrefix(id"-", self.parseExpr())

proc parseAnnotation(self: Parser): Node =
    dbg self, "parseAnnotation"

    self.checkToken(notation={Prefix})
    self.skip(At)

    let name = self.parseId()

    result =
        if self.skipMaybe(LParen):
            var args = newEmptyParen()
            self.parseBlock(args, mode = List, until = some(RParen))
            self.skip(RParen)

            newAnnotationList(newExprParenFromParen(name, args))
        else:
            newAnnotationList(name)


proc parseInfixOp(self: Parser; left: Node): Node =
    dbg self, "parseInfixOp"

    if not self.curr.kind.isOperator() and not self.curr.kind.isWordLikeOperator():
        self.errUnknownOp(if self.curr.kind == Id: "id " & self.curr.value else: $self.curr.kind)

    let op = id($self.curr.kind)

    self.precedence = some(precedences[self.curr.kind])
    self.skipToken()

    dbg self, "parseInfixOp right"

    result = newInfix(op, left, self.parseExpr())

    dbg self, "parseInfixOp end"

proc parseArgs(self: Parser; left: Node): Node =
    dbg self, "parseArgs"

    if left.kind != nkId:
        self.errExpectedNodeOf(nkId)

    self.skip(LParen)
    result = newEmptyExprParen(left)
    self.parseBlock(result, mode = List, until = some(RParen))
    self.expectSameIndent()
    self.skip(RParen)




proc parseIf(self: Parser): Node =
    dbg self, "parseIf"

    let mainBranch = self.parseIfBranch()
    var branches = @[mainBranch]

    dbg self, "parseIf before loop"

    while self.isKind(KwElif):
        dbg self, "parseIf loop"

        self.expectSameIndent()
        branches &= self.parseIfBranch()

    dbg self, "parseIf after loop"

    let elseBranch =
        if self.isKind(KwElse): self.parseElse()
        else: nil.Node

    dbg self, "parseIf end"

    result = newIfExpr(branches, elseBranch)

proc parseIfBranch(self: Parser): Node =
    ## Parses `if` or `elif` branch
    self.skip({KwIf, KwElif})

    dbg self, "parseIfBranch"

    let condition = self.parseExprOrBlock()

    dbg self, "parseIfBranch body"

    let body = self.parseDoOrBlock()

    dbg self, "parseIfBranch end"

    result = newIfBranch(condition, body)

proc parseElse(self: Parser): Node =
    dbg self, "parseElse"

    self.skip(KwElse)

    result =
        if self.curr.isFirstInLine():
            var body = newEmptyExprList()
            self.parseBlock(body)
            newElseBranch(body)
        else:
            let expr = self.parseExpr()
            newElseBranch(expr)

proc parseList(self: Parser): Node =
    dbg self, "parseList"

    self.skip({LBracket, LParen, LBrace})
    let untilKind: TokenKind
    result = case self.prev.kind:
        of LBracket:
            untilKind = RBracket
            newEmptyBracket()
        of LParen:
            untilKind = RParen
            newEmptyParen()
        of LBrace:
            untilKind = RBrace
            newEmptyBrace()
        else: unreachable()
    let mode = self.parseBlock(result, mode = Auto, until = some(untilKind))
    self.expectSameIndent()
    self.skip(untilKind)

    case mode
    of List:
        result.flags.incl(nfList)
    of Block:
        if result.kind == nkParen:
            result = newExprList(result.children)
    else: discard

proc parseType(self: Parser): Node =
    dbg self, "parseType"

    self.skip(KwType)

    var id = self.parseId()
    let genericParams =
        if self.curr.kind == LBracket: self.parseList()
        else: nil.Node

    if genericParams != nil:
        id = newExprBracketFromBracket(id, genericParams)

    dbg self, "parseType typedesc"

    let expr =
        if self.skipMaybe(Assign):
            self.parseTypeExpr()
        else:
            self.skip({KwStruct, KwEnum})
            let body = self.parseExprOrBlock()
            body.expectKind(nkExprList)
            body

    result = newType(id, expr)

proc parseTypeExpr(self: Parser): Node =
    unimplemented()



proc parseDo(self: Parser): Node =
    self.skip(KwDo)

    result = newEmptyExprList()

    if self.curr.isFirstInLine():
        self.parseBlock(result)
    # elif self.curr.kind == LParen and self.curr.spacesAfter().get() == 0:
    #     # do(...expressions)
    #     self.skip(LParen)
    #     self.parseBlock(result, none(BlockContext), RParen)
    #     self.skip(RParen)
    else:
        result &= self.parseExpr()

proc parseDoOrBlock(self: Parser): Node =
    dbg self, "parseDoOrBlock"

    if self.curr.kind == KwDo:
        dbg self, "parseDoOrBlock do"

        result = self.parseDo()
    else:
        dbg self, "parseDoOrBlock block"

        result = newEmptyExprList()
        self.parseBlock(result)

proc parseExprOrBlock(self: Parser): Node =
    dbg self, "parseExprOrBlock"

    if self.curr.isFirstInLine():
        dbg self, "parseExprOrBlock block"

        result = newEmptyExprList()
        self.parseBlock(result)
    else:
        dbg self, "parseExprOrBlock expr"

        result = self.parseExpr()



proc parseVarDecl(self: Parser; left: Node): Node =
    dbg self, "parseVarDecl"

    left.expectKind({nkId, nkVar, nkVal})

    var prefix = nil.Node
    var id     = nil.Node

    while true:
        dbg self, "parseVarDecl loop"

        case self.curr.kind
        of Ampersand:
            self.skip(Ampersand)
            prefix =
                if prefix == nil: id"&"
                else: newPrefix(prefix, id"&")
        of QuestionMark:
            self.skip(QuestionMark)
            prefix =
                if prefix == nil: id"?"
                else: newPrefix(prefix, id"?")
        of LBracket:
            prefix =
                if prefix == nil: self.parseList()
                else: newPrefix(prefix, self.parseList())
        of Id:
            id = self.parseId()
            break
        else:
            break

    if id == nil:
        self.err("expected identifier for type expression")

    dbg self, "parseVarDecl genericInst"

    let genericInst =
        if self.curr.kind == LBracket: self.parseList()
        else: nil.Node

    if genericInst != nil:
        id = newExprBracketFromBracket(id, genericInst)

    dbg self, "parseVarDecl end"

    result =
        if prefix != nil: newPrefix(prefix, id)
        else: id


proc expectSameIndent(self: Parser) =
    if self.curr.checkIndent(self.blocks.peek()) != 0:
        self.errInvalidBlockContext()
