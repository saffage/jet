import
  std/strformat,
  std/strutils,
  std/sequtils,
  std/tables,
  std/options,
  std/enumutils,

  jet/token,
  jet/ast,

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
    Assign
    Or
    And
    Ord
    Sum
    Product
    Suffix
    Path
    Highest

  ParserError = object of CatchableError
    info* : LineInfo

  BlockContext = tuple[line, indent: int]

  ParsePrefixFunc = proc(self: var Parser): AstNode {.nimcall, noSideEffect.}
  ParseInfixFunc  = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect.}
  ParseSuffixFunc = proc(self: var Parser; left: AstNode): AstNode {.nimcall, noSideEffect.}

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
  Assign     : Assign,
}.toTable()

#
# Parse Functions
#

type
  ParseMode = enum
    Block
    List
    Adaptive

func parseExpr(self: var Parser): AstNode
func parseEmpty(self: var Parser): AstNode
func parseId(self: var Parser): AstNode
func parseFunc(self: var Parser): AstNode
func parseVarDecl(self: var Parser): AstNode
func parseDo(self: var Parser): AstNode
func parseDoOrBlock(self: var Parser): AstNode
func parseBlock(
  self: var Parser;
  body: var seq[AstNode];
  mode: ParseMode = Block;
  until: Option[TokenKind] = none(TokenKind);
  fn: ParsePrefixFunc = parseExpr;
): ParseMode {.discardable.}

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
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind in kinds:
    self.skipToken(kinds)
    result = true
  else:
    result = false

func skipTokenMaybe(self: var Parser; kind: TokenKind): bool
  {.raises: [ParserError, ValueError].} =
  self.skipTokenMaybe({kind})

func isNewBlockContext(self: Parser; context: BlockContext): bool =
  self.blockStack.isEmpty() or context.indent > self.blockStack.peek().indent

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
  result.prefixFuncs[Empty] = parseEmpty
  result.prefixFuncs[VSpace] = parseEmpty
  result.prefixFuncs[Id] = parseId
  result.prefixFuncs[KwFunc] = parseFunc

#
# Parse Functions Implementation
#

# TODO: fix errors
func parseExpr(self: var Parser): AstNode =
  debug("parseExpr()")
  var token = ?self.peekToken()
  let fn = self.prefixFuncs.getOrDefault(token.kind)

  if fn == nil:
    return err (&"expression is expected, got {token.kind}", token.info)

  var tree = ?fn(self)
  token = ?self.peekToken()

  if token.kind == Eof:
    return ok(tree)

  while self.precedence.isNone() or
        self.precedence.get() < precedences.getOrDefault(token.kind, Lowest):
      let fn = self.infixFuncs.getOrDefault(token.kind)

      if fn == nil:
        break

      var newTree = AstNode(kind: Branch, branchKind: Infix, children: newSeqOfCap[AstNode](2))
      newTree.children &= tree
      newTree.children &= ?fn(self, tree)
      tree = newTree

  self.precedence = none(Precedence)
  ok(tree)

func parseEmpty(self: var Parser): ParseResult =
  debug("parseEmpty()")
  self.skipAnyToken()
  ok(AstNode(kind: Empty))

func parseId(self: var Parser): ParseResult =
  debug("parseId()")
  let token = ?self.popToken()

  if token.kind == Id:
    ok(AstNode(kind: Id, id: token.data))
  else:
    err((&"expected identifier, got {token.kind}", token.info))

func parseFunc(self: var Parser): ParseResult =
  debug("parseFunc()")
  ?self.skipToken(KwFunc)
  ?self.skipToken(HSpace)

  var funcTree = AstNode(kind: Branch, branchKind: Func, children: newSeqOfCap[AstNode](4))
  var id = ?self.parseId()

  ?self.skipToken(LeRound)
  var params = AstNode(kind: Branch, branchKind: Tuple)
  discard ?self.parseBlock(
    params.children,
    mode = List,
    until = some(RiRound),
    fn = parseVarDecl)
  ?self.skipToken(RiRound)

  discard ?self.skipTokenMaybe(HSpace)
  var returnType = AstNode(kind: Empty)

  var body = ?self.parseDoOrBlock()

  funcTree.children &= id
  funcTree.children &= params
  funcTree.children &= returnType
  funcTree.children &= body

  ok(funcTree)

func parseVarDecl(self: var Parser): ParseResult =
  debug("parseVarDecl()")
  var varDecl = AstNode(kind: Branch, branchKind: VarDecl, children: newSeqOfCap[AstNode](3))

  let id = ?self.parseId()
  ?self.skipToken(HSpace)
  let typeExpr = ?self.parseId()

  varDecl.children &= id
  varDecl.children &= typeExpr

  ok(varDecl)

func parseDo(self: var Parser): ParseResult =
  ?self.skipToken(KwDo)

  var body = AstNode(kind: Branch, branchKind: Block)
  let wasSpace = ?self.skipTokenMaybe(HSpace)

  if (?self.peekToken()).kind == VSpace:
    discard ?self.parseBlock(body.children)
  else:
    if wasSpace:
      body.children &= ?self.parseExpr()
    else:
      todo() # lambda?

  ok(body)

func parseDoOrBlock(self: var Parser): ParseResult =
  if (?self.peekToken()).kind == KwDo:
    ok(?self.parseDo())
  else:
    var body = AstNode(kind: Branch, branchKind: Block)
    discard ?self.parseBlock(body.children)
    ok(body)

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
    if ?self.skipTokenMaybe(VSpace) or self.curr == 0:
      isNewLine = true

    if (?self.peekToken()).kind in untilKinds:
      break

    if isNewLine:
      if mode == Block and wasSemicolon:
        return err("expected expression after semicolon")

      # compute indentation
      let token = ?self.peekToken()
      indent = block:
        if token.kind != HSpace: 0
        else:
          assert(token.data.len() > 0)
          token.data.len()

      if contextPushed:
        # check indentation of token
        if indent > self.blockStack.peek().indent:
          return err (
            &"token {token.human()} is offside the context started at {self.blockStack.peek()}",
            token.info
          )
        elif indent < self.blockStack.peek().indent:
          # end of block
          break
        else:
          discard
      else:
        # create a new context
        let newContext = (line: token.info.line.int, indent: indent)

        # validate new context
        if not self.isNewBlockContext(newContext) and self.blockStack.len() != 0 and indent != 0:
          return err(
            &"a new block context expected, but got {newContext}, " &
            &"which is the same or lower with previous context {self.blockStack.peek()}"
          )

        # push it
        self.blockStack.push(newContext)
        contextPushed = true

      if indent > 0:
        ?self.skipToken(HSpace)
    else:
      let token = ?self.peekToken()
      if mode == Block and not wasSemicolon:
        return err ("the other expression must be on a new line or separated by semicolon", token.info)
      wasSemicolon = false

    assert((?self.peekToken()).kind != HSpace, $isNewLine)

    let tree = ?fn(self)

    if tree.kind != Empty:
      body &= tree

    isNewLine = false
    let token = ?self.peekToken()

    # TODO: validate expression end

    if mode == Adaptive:
      mode = if token.kind == Comma: List else: Block
      hint fmt"determine mode of block parsing: {result}"

    if mode == Block:
      if ?self.skipTokenMaybe(Semicolon):
        wasSemicolon = true
    if mode == List:
      if not ?self.skipTokenMaybe(Comma):
        let token = ?self.peekToken()

        if token.kind notin untilKinds:
          return err(&"expected comma after expression")

        break

    while ?self.skipTokenMaybe(HSpace): discard

  if mode == Adaptive:
    # something like `()` or `[]`
    mode = List

  if mode == Block and wasSemicolon:
    return err("expected expression after ';'")

  if contextPushed:
    self.blockStack.drop()

  ok(mode)
