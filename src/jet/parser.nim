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
    Highest

  ParserError* = object of CatchableError
    info* : LineInfo

  BlockContext = tuple[line, indent: int]

  ParsePrefixFunc = proc(self: var Parser): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}
  ParseInfixFunc  = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}
  ParseSuffixFunc = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect, raises: [ParserError, ValueError].}

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
  Eq     : Eq,
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
func parseFunc(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseIf(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseWhile(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseReturn(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseVar(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseVal(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseValDecl(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDo(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDoOrBlock(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
func parseDoOrExpr(self: var Parser): AstNode {.raises: [ParserError, ValueError].}
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

func peekToken(self: Parser): Token
  {.raises: [ParserError].} =
  if self.curr > self.tokens.high:
    raise (ref ParserError)(msg: "no token to peek")

  result = self.tokens[self.curr]

func peekToken(self: Parser; kind: TokenKind): Token
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind != kind:
    raise (ref ParserError)(msg: &"expected token of kind {kind}, got {token.kind}")

  result = token

func popToken(self: var Parser): Token
  {.raises: [ParserError].} =
  result = self.peekToken()
  self.curr += 1

func skipToken(self: var Parser; kinds: set[TokenKind])
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind notin kinds:
    let message = kinds.toSeq().join(" or ")
    raise (ref ParserError)(msg: &"expected token of kind {message}, got {token.kind}")

  self.curr += 1

func skipToken(self: var Parser; kind: TokenKind)
  {.raises: [ParserError, ValueError].} =
  self.skipToken({kind})

func skipAnyToken(self: var Parser) =
  self.curr += 1

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
  self.skipToken({KwIf, KwElif})
  self.skipToken(HSpace)

  let cond = self.parseExpr()
  self.skipTokenMaybe(HSpace)
  let body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: IfBranch, children: newSeqOfCap[AstNode](2))
  result.children &= cond
  result.children &= body

func parseElseBranch(self: var Parser): AstNode
  {.raises: [ParserError, ValueError].} =
  self.skipToken(KwElse)
  self.skipTokenMaybe(HSpace)

  let body =
    if self.peekToken().kind == VSpace:
      var body = newSeq[AstNode]()
      self.parseBlock(body)
      AstNode(kind: Branch, branchKind: Block, children: body)
    else:
      self.parseExpr()

  result = AstNode(kind: Branch, branchKind: ElseBranch, children: @[body])

#
# Parse Functions Implementation
#

func parseLit(self: var Parser): AstNode =
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
      raise (ref ParserError)(msg: &"expected literal, got {token.kind}", info: token.info)

func parseExpr(self: var Parser): AstNode =
  debug("parseExpr()")
  var token = self.peekToken()
  let fn = self.prefixFuncs.getOrDefault(token.kind)

  if fn == nil:
    raise (ref ParserError)(msg: &"expression is expected, got {token.kind}", info: token.info)

  result = fn(self)
  token = self.peekToken()

  if token.kind == Eof:
    return

  while self.precedence.isNone() or
        self.precedence.get() < precedences.getOrDefault(token.kind, Lowest):
      let fn = self.infixFuncs.getOrDefault(token.kind)

      if fn == nil:
        break

      var newTree = AstNode(kind: Branch, branchKind: Infix, children: newSeqOfCap[AstNode](2))
      newTree.children &= result
      newTree.children &= fn(self, result)
      result = newTree

  self.precedence = none(Precedence)

func parseId(self: var Parser): AstNode =
  debug("parseId()")
  let token = self.popToken()

  if token.kind != Id:
    raise (ref ParserError)(msg: &"expected identifier, got {token.kind}", info: token.info)

  result = AstNode(kind: Id, id: token.data)

func parseFunc(self: var Parser): AstNode =
  debug("parseFunc()")
  self.skipToken(KwFunc)
  self.skipToken(HSpace)

  var id = self.parseId()
  var params = AstNode(kind: Branch, branchKind: Tuple)

  self.skipToken(LeRound)
  self.parseBlock(
    params.children,
    mode = List,
    until = some(RiRound),
    fn = parseValDecl)
  self.skipToken(RiRound)

  discard self.skipTokenMaybe(HSpace)
  var returnType = AstNode(kind: Empty)
  var body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: Func, children: newSeqOfCap[AstNode](4))
  result.children &= id
  result.children &= params
  result.children &= returnType
  result.children &= body

func parseIf(self: var Parser): AstNode =
  var branches = newSeq[AstNode]()

  while true:
    branches &= self.parseIfBranch()
    self.skipTokenMaybe({HSpace, VSpace})
    if self.peekToken().kind != KwElif: break

  let elseBranch =
    if self.peekToken().kind == KwElse:
      self.parseElseBranch()
    else:
      emptyNode

  result = AstNode(kind: Branch, branchKind: If, children: branches)

  if elseBranch.kind != Empty:
    result.children &= elseBranch

func parseWhile(self: var Parser): AstNode =
  self.skipToken(KwWhile)
  self.skipToken(HSpace)

  let cond = self.parseExpr()
  self.skipTokenMaybe(HSpace)
  let body = self.parseDoOrBlock()

  result = AstNode(kind: Branch, branchKind: While, children: newSeqOfCap[AstNode](2))
  result.children &= cond
  result.children &= body

func parseReturn(self: var Parser): AstNode =
  self.skipToken(KwReturn)
  self.skipToken(HSpace)

  let expr = self.parseExpr()

  result = AstNode(kind: Branch, branchKind: Return, children: newSeqOfCap[AstNode](1))
  result.children &= expr

func parseVar(self: var Parser): AstNode =
  debug("parseVar()")

  self.skipToken(KwVar)
  self.skipToken(HSpace)
  result = self.parseValDecl()
  result = AstNode(kind: Branch, branchKind: VarDecl, children: result.children)

func parseVal(self: var Parser): AstNode =
  debug("parseVal()")

  self.skipToken(KwVal)
  self.skipToken(HSpace)
  result = self.parseValDecl()

func parseValDecl(self: var Parser): AstNode =
  debug("parseValDecl()")

  let id = self.parseId()
  self.skipToken(HSpace)
  let typeExpr = self.parseId()
  self.skipTokenMaybe(HSpace)
  let body =
    if self.skipTokenMaybe(Eq):
      self.skipTokenMaybe(HSpace)
      self.parseDoOrExpr()
    else:
      emptyNode

  result = AstNode(kind: Branch, branchKind: ValDecl, children: newSeqOfCap[AstNode](3))
  result.children &= id
  result.children &= typeExpr
  result.children &= body

func parseDo(self: var Parser): AstNode =
  debug("parseDo()")

  self.skipToken(KwDo)

  let wasSpace = self.skipTokenMaybe(HSpace)
  result = AstNode(kind: Branch, branchKind: Block)

  if (self.peekToken()).kind == VSpace:
    self.parseBlock(result.children)
  else:
    if wasSpace:
      result.children &= self.parseExpr()
    else:
      todo() # lambda?

func parseDoOrBlock(self: var Parser): AstNode =
  debug("parseDoOrBlock()")

  if (self.peekToken()).kind == KwDo:
    result = self.parseDo()
  else:
    result = AstNode(kind: Branch, branchKind: Block)
    self.parseBlock(result.children)

func parseDoOrExpr(self: var Parser): AstNode =
  debug("parseDoOrExpr()")

  if (self.peekToken()).kind == KwDo:
    result = self.parseDo()
  else:
    result = self.parseExpr()

func parseBlock(
  self: var Parser;
  body: var seq[AstNode];
  mode: ParseMode = Block;
  until: Option[TokenKind];
  fn: ParsePrefixFunc;
): ParseMode =
  debug("parseBlock()")
  var contextPushed = false
  var wasSemicolon  = false
  var mode          = mode

  let untilKinds =
    if untilKind =? until: {Eof, untilKind}
    else: {Eof}

  var isNewLine = false
  var indent    = -1

  while true:
    if self.skipTokenMaybe(VSpace) or self.curr == 0:
      isNewLine = true

    if (self.peekToken()).kind in untilKinds:
      break

    debug(&"parseBlock: token {self.peekToken().human()}")
    debug(&"parseBlock: isNewLine = {isNewLine}")

    if isNewLine:
      if mode == Block and wasSemicolon:
        raise (ref ParserError)(msg: "expected expression after semicolon")

      # compute indentation
      let token = self.peekToken()
      indent = block:
        if token.kind != HSpace: 0
        else:
          assert(token.data.len() > 0)
          token.data.len()

      if contextPushed:
        # check indentation of token
        if indent > self.blockStack.peek().indent:
          raise (ref ParserError)(
            msg: &"token {token.human()} is offside the context started at {self.blockStack.peek()}",
            info: token.info)
        elif indent < self.blockStack.peek().indent:
          # end of block
          if isNewLine: self.curr -= 1
          break
        else:
          discard
      else:
        # create a new context
        let newContext = (line: token.info.line.int, indent: indent)

        # validate new context
        if not self.isNewBlockContext(newContext) and self.blockStack.len() != 0 and indent != 0:
          raise (ref ParserError)(
            msg: &"a new block context expected, but got {newContext}, " &
                 &"which is the same or lower with previous context {self.blockStack.peek()}")

        # push it
        self.blockStack.push(newContext)
        contextPushed = true

      if indent > 0:
        self.skipToken(HSpace)
    else:
      let token = self.peekToken()
      if mode == Block and not wasSemicolon:
        raise (ref ParserError)(
          msg: "the other expression must be on a new line or separated by semicolon",
          info: token.info)
      wasSemicolon = false

    assert((self.peekToken()).kind != HSpace, $isNewLine)

    let tree = fn(self)

    if tree.kind != Empty:
      body &= tree

    isNewLine = false
    let token = self.peekToken()

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
          raise (ref ParserError)(msg: &"expected comma after expression")

        break

    while self.skipTokenMaybe(HSpace): discard

  if mode == Adaptive:
    # something like `()` or `[]`
    mode = List

  if mode == Block and wasSemicolon:
    raise (ref ParserError)(msg: "expected expression after ';'")

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
  self.blockStack.push((1, 0))
  self.parseBlock(ast.children, until = some(Eof))
  self.ast = some(ast)

func getAst*(self: Parser): Option[AstNode] =
  self.ast

func newParser*(tokens: openArray[Token]): Parser =
  result = Parser(tokens: @tokens)
  result.prefixFuncs[Id] = parseId
  result.prefixFuncs[KwDo] = parseDo
  result.prefixFuncs[KwFunc] = parseFunc
  result.prefixFuncs[KwVal] = parseVal
  result.prefixFuncs[KwVar] = parseVar
  result.prefixFuncs[KwIf] = parseIf
  result.prefixFuncs[KwWhile] = parseWhile
  result.prefixFuncs[KwReturn] = parseReturn

  result.prefixFuncs[KwNil] = parseLit
  result.prefixFuncs[KwTrue] = parseLit
  result.prefixFuncs[KwFalse] = parseLit
  result.prefixFuncs[IntLit] = parseLit
  result.prefixFuncs[FloatLit] = parseLit
  result.prefixFuncs[StringLit] = parseLit
  result.prefixFuncs[CharLit] = parseLit
