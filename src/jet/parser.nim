import
  std/strformat,
  std/strutils,
  std/sequtils,
  std/tables,
  std/options,
  std/enumutils,

  jet/token,
  jet/ast,
  jet/literal,

  lib/utils,
  lib/stacks,
  lib/line_info,

  pkg/questionable

{.push, raises: [].}

type
  Parser* {.byref.} = object
    tokens      : seq[Token]
    curr        : int = 0
    ast         : Option[AstNode] = none(AstNode)
    blockStack  : Stack[BlockContext]
    prefixFuncs : OrderedTable[TokenKind, ParsePrefixFunc]
    infixFuncs  : OrderedTable[TokenKind, ParseInfixFunc]
    suffixFuncs : OrderedTable[TokenKind, ParseSuffixFunc]
    precedence  : Option[Precedence]

  Precedence = enum
    Lowest
    Eq
    Or
    And
    Ord
    Sum
    Product
    Suffix
    Path
    Prefix
    Highest

  ParserError* = object of CatchableError
    info* : LineInfo

  BlockContext = tuple[line, indent: int]

  ParsePrefixFunc = proc(self: var Parser): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}
  ParseInfixFunc  = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}
  ParseSuffixFunc = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}

func str(kind: TokenKind): string =
  $kind

func `$`(kind: TokenKind): string =
  kind.symbolName()

const precedences = {
  LeRound    : Highest,
  LeSquare   : Highest,
  LeCurly    : Highest,
  Dot        : Path,
  Asterisk   : Product,
  Slash      : Product,
  Percent    : Product,
  Plus       : Sum,
  Minus      : Sum,
  PlusPlus   : Sum,
  EqOp       : Ord,
  NeOp       : Ord,
  LtOp       : Ord,
  GtOp       : Ord,
  LeOp       : Ord,
  GeOp       : Ord,
  KwAnd      : And,
  KwOr       : Or,
  KwNot      : Prefix,
  Eq         : Eq,
}.toTable()

#
# Parse Functions
#

type
  ParseMode = enum
    Block
    List
    Adaptive

func parseLit(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseExpr(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseId(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseNot(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseStruct(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseType(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseFunc(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseIf(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseWhile(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseReturn(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseVar(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseVal(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseValDecl(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseParamOrField(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDo(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDoOrBlock(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDoOrExpr(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseExprOrBlock(self: var Parser; fn: ParsePrefixFunc = parseExpr): AstNode {.raises: [ParserError, ValueError].}
func parseInfix(self: var Parser; left: AstNode): AstNode {.raises: [ParserError, ValueError].}
func parseList(self: var Parser; fn: ParsePrefixFunc): AstNode {.raises: [ParserError, ValueError].}
func parseList(self: var Parser): AstNode {.raises: [ParserError, ValueError].} = self.parseList(parseExpr)
func parseBlock(
  self: var Parser;
  body: var seq[AstNode];
  mode: ParseMode = Block;
  until: Option[TokenKind] = none(TokenKind);
  fn: ParsePrefixFunc = parseExpr;
): ParseMode {.discardable, raises: [ParserError, ValueError].}

#
# Util Functions
#

template raiseParserError(message: string; lineInfo: LineInfo) =
  raise (ref ParserError)(msg: message, info: lineInfo)

func peekToken(self: Parser): Token
  {.raises: [ParserError].} =
  if self.curr > self.tokens.high:
    raiseParserError("no token to peek", self.tokens[^1].info)

  result = self.tokens[self.curr]

func peekToken(self: Parser; kind: TokenKind): Token
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind != kind:
    raiseParserError(&"expected token of kind {kind}, got {token.kind}", token.info)

  result = token

func prevToken*(self: Parser): Token
  {.raises: [ParserError].} =
  let idx = self.curr - 1

  if idx < 0 or idx >= self.tokens.high:
    raiseParserError("no previous token to peek", self.tokens[0].info)

  result = self.tokens[idx]

func popToken(self: var Parser): Token
  {.raises: [ParserError].} =
  result = self.peekToken()
  self.curr += 1

func skipToken(self: var Parser; kinds: set[TokenKind])
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind notin kinds:
    let message = kinds.toSeq().join(" or ")
    raiseParserError(&"expected token of kind {message}, got {token.kind}", token.info)

  self.curr += 1

func skipToken(self: var Parser; kind: TokenKind)
  {.raises: [ParserError, ValueError].} =
  self.skipToken({kind})

func skipTokenMaybe(self: var Parser; kinds: set[TokenKind]): bool
  {.discardable, raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind in kinds:
    self.skipToken(kinds)
    result = true
  else:
    result = false

func skipTokenMaybe(self: var Parser; kind: TokenKind): bool
  {.discardable, raises: [ParserError, ValueError].} =
  self.skipTokenMaybe({kind})

func isNewBlockContext(self: Parser; context: BlockContext): bool =
  self.blockStack.isEmpty() or context.indent > self.blockStack.peek().indent

#
# Parse Functions AUX
#

func parseIfBranch(self: var Parser): AstNode
  {.raises: [ParserError, ValueError].} =
  debug("parseIfBranch")

  self.skipToken({KwIf, KwElif})

  debug(&"parseIfBranch: {self.prevToken().kind}")

  let cond = self.parseExpr()
  let body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: IfBranch, children: @[cond, body])

func parseElseBranch(self: var Parser): AstNode
  {.raises: [ParserError, ValueError].} =
  debug("parseElseBranch")

  self.skipToken(KwElse)

  let body = self.parseExprOrBlock()

  result = AstNode(kind: Branch, branchKind: ElseBranch, children: @[body])

#
# Parse Functions Implementation
#

func parseLit(self: var Parser): AstNode =
  debug("parseLit")

  let token = self.popToken()

  result = case token.kind:
    of KwNil:
      AstNode(kind: Lit, lit: newLit(nil))
    of KwTrue:
      AstNode(kind: Lit, lit: newLit(true))
    of KwFalse:
      AstNode(kind: Lit, lit: newLit(false))
    of StringLit:
      AstNode(kind: Lit, lit: newLit(token.data))
    of CharLit:
      if token.data.len() != 1:
        raise (ref ValueError)(msg: &"invalid character: '{token.data}'")
      AstNode(kind: Lit, lit: newLit(token.data[0]))
    of IntLit:
      AstNode(kind: Lit, lit: newLit(token.data.parseBiggestInt()))
    of FloatLit:
      AstNode(kind: Lit, lit: newLit(token.data.parseFloat()))
    else:
      raiseParserError(&"expected literal, got {token.kind}", token.info)

func parseExpr(self: var Parser): AstNode =
  debug("parseExpr")

  var token      = self.peekToken()
  let precedence = self.precedence.get(Lowest)
  let fn         = self.prefixFuncs.getOrDefault(token.kind)

  if fn == nil:
    raiseParserError(&"expression is expected, got {token.kind}", token.info)

  result = fn(self)

  if self.peekToken().kind == Eof:
    return

  while true:
    token = self.peekToken()

    debug(&"parseExpr: precedence = {precedence}")

    if precedence >= precedences.getOrDefault(token.kind, Lowest):
      break

    debug("parseExpr: infix")
    debug(&"parseExpr: token {token.human()}")

    let fn = self.infixFuncs.getOrDefault(token.kind)
    if fn == nil: break
    result = fn(self, result)

  self.precedence = none(Precedence)

func parseId(self: var Parser): AstNode =
  debug("parseId")

  let token = self.popToken()

  if token.kind != Id:
    raiseParserError(&"expected identifier, got {token.kind}", token.info)

  result = AstNode(kind: Id, id: token.data)

func parseNot(self: var Parser): AstNode =
  debug("parseNot")

  self.skipToken(KwNot)
  self.precedence = some(Precedence.Prefix)

  let expr = self.parseExpr()
  let notOp = AstNode(kind: Operator, op: OpNot)

  result = AstNode(kind: Branch, branchKind: Prefix, children: @[notOp, expr])

func parseStruct(self: var Parser): AstNode =
  debug("parseStruct")

  self.skipToken(KwStruct)

  let body = self.parseExprOrBlock(fn = parseParamOrField)
  result = AstNode(kind: Branch, branchKind: Struct, children: @[body])

func parseType(self: var Parser): AstNode =
  debug("parseType")

  self.skipToken(KwType)

  let id = self.parseId()
  let typeExpr = self.parseExpr()

  result = AstNode(kind: Branch, branchKind: Type, children: @[id, typeExpr])

func parseFunc(self: var Parser): AstNode =
  debug("parseFunc")

  self.skipToken(KwFunc)

  let id = self.parseId()
  let params = self.parseList(fn = parseParamOrField)
  let returnType =
    if self.prevToken().spaces.trailing != spacingLast and
       self.peekToken().kind != KwDo: self.parseExpr()
    else: emptyNode
  let body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: Func, children: @[id, params, returnType, body])

func parseIf(self: var Parser): AstNode =
  debug("parseIf")

  var branches = newSeq[AstNode]()

  while true:
    branches &= self.parseIfBranch()
    if self.peekToken().kind != KwElif: break

  let elseBranch =
    if self.peekToken().kind == KwElse: self.parseElseBranch()
    else: emptyNode

  result = AstNode(kind: Branch, branchKind: If, children: branches)

  if elseBranch.kind != Empty:
    result.children &= elseBranch

func parseWhile(self: var Parser): AstNode =
  debug("parseWhile")

  self.skipToken(KwWhile)

  let cond = self.parseExpr()
  let body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: While, children: @[cond, body])

func parseReturn(self: var Parser): AstNode =
  debug("parseReturn")

  self.skipToken(KwReturn)

  let expr = self.parseExpr()

  result = AstNode(kind: Branch, branchKind: Return, children: @[expr])

func parseVar(self: var Parser): AstNode =
  debug("parseVar()")

  self.skipToken(KwVar)

  result = self.parseValDecl()
  result = AstNode(kind: Branch, branchKind: VarDecl, children: result.children)

func parseVal(self: var Parser): AstNode =
  debug("parseVal()")

  self.skipToken(KwVal)

  result = self.parseValDecl()

func parseValDecl(self: var Parser): AstNode =
  debug("parseValDecl")

  let id = self.parseId()
  let typeExpr = self.parseExpr()
  let body =
    if self.skipTokenMaybe(Eq): self.parseDoOrExpr()
    else: emptyNode

  result = AstNode(kind: Branch, branchKind: ValDecl, children: @[id, typeExpr, body])

func parseParamOrField(self: var Parser): AstNode =
  debug("parseParamOrField")

  result = case self.peekToken().kind:
    of KwVar: self.parseVar()
    of KwVal: self.parseVal()
    else: self.parseValDecl()

func parseDo(self: var Parser): AstNode =
  debug("parseDo")

  self.skipToken(KwDo)

  let expr = self.parseExprOrBlock()

  result =
    if expr.kind == Branch and expr.branchKind == Block: expr
    else: AstNode(kind: Branch, branchKind: Block, children: @[expr])

func parseDoOrBlock(self: var Parser): AstNode =
  debug("parseDoOrBlock")

  if self.peekToken().kind == KwDo:
    result = self.parseDo()
  else:
    result = AstNode(kind: Branch, branchKind: Block)
    self.parseBlock(result.children)

func parseDoOrExpr(self: var Parser): AstNode =
  debug("parseDoOrExpr")

  result =
    if self.peekToken().kind == KwDo: self.parseDo()
    else: self.parseExpr()

func parseExprOrBlock(self: var Parser; fn: ParsePrefixFunc): AstNode =
  debug("parseExprOrBlock")

  if self.peekToken().spaces.wasLF:
    result = AstNode(kind: Branch, branchKind: Block, children: @[])
    self.parseBlock(result.children, fn = fn)
  else:
    result = fn(self)

func parseInfix(self: var Parser; left: AstNode): AstNode =
  debug("parseInfix")

  let token = self.popToken()

  if token.kind notin OperatorKinds + WordLikeOperatorKinds:
    raiseParserError(&"expected operator, got '{token.kind}'", token.info)

  let op = token.kind.str()
  let opKind = op.toOperatorKind()

  if opKind.isNone():
    raiseParserError(&"operator '{op}' not yet supported", token.info)

  self.precedence = some do:
    try:
      precedences[token.kind]
    except KeyError:
      unreachable()

  let opNode = AstNode(kind: Operator, op: opKind.get())
  let right = self.parseExpr()

  result = AstNode(kind: Branch, branchKind: Infix, children: @[opNode, left, right])

func parseList(self: var Parser; fn: ParsePrefixFunc): AstNode =
  debug("parseList")

  let until = case self.peekToken().kind:
    of LeRound: RiRound
    of LeCurly: RiCurly
    of LeSquare: RiSquare
    else: raiseParserError(&"expected ( or [ of {{, got {self.peekToken().kind}", self.peekToken().info)

  self.skipToken(self.peekToken().kind)
  var elems = newSeq[AstNode]()
  let mode = self.parseBlock(elems, mode = Adaptive, until = some(until), fn = fn)
  # TODO: check indentation
  self.skipToken(until)

  result = case mode
    of Block:
      if elems.len() == 1: elems[0]
      else: AstNode(kind: Branch, branchKind: Block, children: elems)
    of List: AstNode(kind: Branch, branchKind: List, children: elems)
    else: unreachable()

func parseBlock(
  self: var Parser;
  body: var seq[AstNode];
  mode: ParseMode = Block;
  until: Option[TokenKind];
  fn: ParsePrefixFunc;
): ParseMode =
  debug("parseBlock")

  var contextPushed = false
  var wasSemicolon  = false
  var mode          = mode

  let untilKinds =
    if untilKind =? until: {Eof, untilKind}
    else: {Eof}

  if self.blockStack.isEmpty():
    self.blockStack.push((1, 0))
    contextPushed = true

  while true:
    let token = self.peekToken()

    if token.kind in untilKinds:
      break

    debug(&"parseBlock: token {self.peekToken().human()}")

    if token.spaces.wasLF:
      let token = self.peekToken()
      let indent = token.spaces.leading

      if mode == Block and wasSemicolon:
        raiseParserError("expected expression after semicolon", token.info)

      if contextPushed:
        # check indentation of token
        if indent > self.blockStack.peek().indent:
          raiseParserError(
            &"invalid indentation, expected {self.blockStack.peek().indent}, got {indent}",
            token.info)
        elif indent < self.blockStack.peek().indent:
          # end of block
          break
      else:
        # create a new context
        let newContext = (line: token.info.line.int, indent: indent)

        # validate new context
        if not self.isNewBlockContext(newContext):
          raiseParserError(
            &"a new block context expected, but got {newContext}, " &
            &"which is the same or lower with previous context {self.blockStack.peek()}",
            token.info)

        # push a new context
        self.blockStack.push(newContext)
        contextPushed = true
    else:
      let token = self.peekToken()
      if mode == Block and not wasSemicolon:
        raiseParserError(
          "the other expression must be on a new line or separated by semicolon",
          token.info)
      wasSemicolon = false

    let tree = fn(self)

    if tree.kind != Empty:
      body &= tree

    # TODO: validate expression end

    if mode == Adaptive:
      mode = if token.kind == Comma: List else: Block
      hint fmt"determine mode of block parsing: {mode}"

    if mode == Block:
      if self.skipTokenMaybe(Semicolon):
        wasSemicolon = true
    if mode == List:
      if not self.skipTokenMaybe(Comma):
        let token = self.peekToken()

        if token.kind notin untilKinds:
          raiseParserError(&"expected comma after expression", self.prevToken().info)

        break

  if mode == Adaptive:
    # something like `()` or `[]`
    mode = List

  if mode == Block and wasSemicolon:
    raiseParserError("expected expression after semicolon", self.prevToken().info)

  if contextPushed:
    self.blockStack.drop()

  result = mode

#
# API
#

func parseAll*(self: var Parser)
  {.raises: [ParserError, ValueError].} =
  debug("parseAll()")
  if self.tokens.len() == 0:
    self.ast = some(AstNode(kind: Empty))
    return

  var ast = AstNode(kind: Branch, branchKind: Block, children: @[])
  self.parseBlock(ast.children, until = some(Eof))
  self.ast = some(ast)

func getAst*(self: Parser): Option[AstNode] =
  self.ast

func newParser*(tokens: openArray[Token]): Parser =
  result = Parser(tokens: @tokens)
  result.prefixFuncs[Id]       = parseId
  result.prefixFuncs[LeRound]  = parseList
  result.prefixFuncs[LeCurly]  = parseList
  result.prefixFuncs[LeSquare] = parseList
  result.prefixFuncs[KwNot]    = parseNot
  result.prefixFuncs[KwDo]     = parseDo
  result.prefixFuncs[KwStruct] = parseStruct
  result.prefixFuncs[KwType]   = parseType
  result.prefixFuncs[KwFunc]   = parseFunc
  result.prefixFuncs[KwVal]    = parseVal
  result.prefixFuncs[KwVar]    = parseVar
  result.prefixFuncs[KwIf]     = parseIf
  result.prefixFuncs[KwWhile]  = parseWhile
  result.prefixFuncs[KwReturn] = parseReturn

  result.prefixFuncs[KwNil]     = parseLit
  result.prefixFuncs[KwTrue]    = parseLit
  result.prefixFuncs[KwFalse]   = parseLit
  result.prefixFuncs[IntLit]    = parseLit
  result.prefixFuncs[FloatLit]  = parseLit
  result.prefixFuncs[StringLit] = parseLit
  result.prefixFuncs[CharLit]   = parseLit

  result.infixFuncs[KwAnd]    = parseInfix
  result.infixFuncs[KwOr]     = parseInfix
  result.infixFuncs[EqOp]     = parseInfix
  result.infixFuncs[NeOp]     = parseInfix
  result.infixFuncs[LtOp]     = parseInfix
  result.infixFuncs[GtOp]     = parseInfix
  result.infixFuncs[LeOp]     = parseInfix
  result.infixFuncs[GeOp]     = parseInfix
  result.infixFuncs[Plus]     = parseInfix
  result.infixFuncs[Minus]    = parseInfix
  result.infixFuncs[Asterisk] = parseInfix
  result.infixFuncs[Slash]    = parseInfix
  result.infixFuncs[Percent]  = parseInfix
  result.infixFuncs[Shl]      = parseInfix
  result.infixFuncs[Shr]      = parseInfix
