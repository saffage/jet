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
  jet/lexer,
  jet/lexerbase,

  lib/utils,
  lib/lineinfo

{.push, raises: [].}

type
  Parser* {.byref.} = object
    tokens   : openArray[Token]
    isModule : bool
    filename : string
    curr     : int = 0
    ast      : AstNode = initAstNodeEmpty()
    priority : Option[Priority]

  Priority = enum
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
    range* : Option[FileRange]

  FmtLexer = object of LexerBase

  FmtLexerError* = object of CatchableError

  PrefixFunc = proc(self: var Parser): AstNode
    {.nimcall, noSideEffect, raises: [LexerError, FmtLexerError, ParserError, ValueError].}

  InfixFunc = proc(self: var Parser; left: AstNode): AstNode
    {.nimcall, noSideEffect, raises: [LexerError, FmtLexerError, ParserError, ValueError].}

  SuffixFunc = proc(self: var Parser; left: AstNode): AstNode
    {.nimcall, noSideEffect, raises: [LexerError, FmtLexerError, ParserError, ValueError].}

func str(kind: TokenKind): string =
  $kind

func `$`(kind: TokenKind): string =
  kind.symbolName()

func `$`(node: AstNode): string =
  node.kindStr()

const priorities = {
  LeRound:  Highest,
  LeSquare: Highest,
  LeCurly:  Highest,
  Dot:      Path,
  Asterisk: Product,
  Slash:    Product,
  Percent:  Product,
  Plus:     Sum,
  Minus:    Sum,
  PlusPlus: Sum,
  EqOp:     Ord,
  NeOp:     Ord,
  LtOp:     Ord,
  GtOp:     Ord,
  LeOp:     Ord,
  GeOp:     Ord,
  KwAnd:    And,
  KwOr:     Or,
  Eq:       Eq,
}.toTable()

#
# Parse functions
#

{.pop.} # raises: []
{.push, raises: [LexerError, FmtLexerError, ParserError, ValueError].}

func parseExpr(self: var Parser): AstNode
func parseIfExpr(self: var Parser): AstNode
func parseWhenExpr(self: var Parser): AstNode
func parseReturnExpr(self: var Parser): AstNode
func parseFuncExpr(self: var Parser): AstNode

func parseModule(self: var Parser; isBodyRequired = true): AstNode
func parseStruct(self: var Parser): AstNode

func parseId(self: var Parser): AstNode
func parseLiteral(self: var Parser): AstNode
func parsePrimary(self: var Parser): AstNode
func parseSuffix(self: var Parser; left: AstNode): AstNode
func parseDotId(self: var Parser; left: AstNode): AstNode
func parseExprColonExprList(self: var Parser; left: AstNode): AstNode
func parseExprEqExprList(self: var Parser; left: AstNode): AstNode

func parseBlock(self: var Parser): AstNode

func parseAll*(input: string; offset = (line: 0, column: 0); isModule = false): AstNode
func parseExpr*(input: string; offset = (line: 0, column: 0); isModule = false): AstNode

{.pop.} # raises: [LexerError, FmtLexerError, ParserError, ValueError]
{.push, raises: [].}

#
# Util Functions
#

template raiseParserError*(message: string) =
  raise (ref ParserError)(msg: message)

template raiseParserError*(message: string; node: AstNode) =
  raise (ref ParserError)(msg: message, range: some(node.range))

template raiseParserError*(message: string; fileRange: FileRange) =
  raise (ref ParserError)(msg: message, range: some(fileRange))

template raiseParserError*(message: string; filePos: FilePos) =
  raise (ref ParserError)(msg: message, range: some(filePos .. filePos + 1))

func isEmpty(self: Parser): bool =
  result = self.curr > self.tokens.high

func skipToken(self: var Parser) =
  if self.curr < self.tokens.high:
    self.curr += 1

func peekToken(self: Parser): Token
  {.raises: [ValueError].} =
  if self.isEmpty():
    raise newException(ValueError, "no token to peek")

  result = self.tokens[self.curr]

func peekToken(self: Parser; kinds: set[TokenKind]): Token
  {.raises: [ParserError, ValueError].} =
  result = self.peekToken()

  if result.kind notin kinds:
    let kindsStr = kinds.toSeq().join(" or ")
    raiseParserError(&"expected token of kind {kindsStr}, got {result.kind} instead", result.range)

func peekToken(self: Parser; kind: TokenKind): Token
  {.raises: [ParserError, ValueError].} =
  result = self.peekToken({kind})

func peekKind(self: Parser): TokenKind
  {.raises: [ValueError].} =
  result = self.peekToken().kind

func prevToken*(self: Parser): Token
  {.raises: [ValueError].} =
  let idx = self.curr - 1

  if idx < 0 or idx >= self.tokens.high:
    raise newException(ValueError, "no previous token to peek")

  result = self.tokens[idx]

func popToken(self: var Parser): Token
  {.raises: [ValueError].} =
  result = self.peekToken()
  self.skipToken()

func popToken(self: var Parser; kinds: set[TokenKind]): Token
  {.raises: [ParserError, ValueError].} =
  result = self.peekToken(kinds)
  self.skipToken()

func popToken(self: var Parser; kind: TokenKind): Token
  {.raises: [ParserError, ValueError].} =
  result = self.popToken({kind})

func skipToken(self: var Parser; kinds: set[TokenKind])
  {.raises: [ParserError, ValueError].} =
  let token = self.peekToken()

  if token.kind notin kinds:
    let kindsStr = kinds.toSeq().join(" or ")
    raiseParserError(&"expected token of kind {kindsStr}, got {token.kind}", token.range)

  self.skipToken()

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

#
# Parse Functions Implementation
#

func parseExpr(self: var Parser): AstNode =
  debug("parseExpr")

  let token = self.peekToken()

  result = case token.kind
    of LeCurly:
      self.parseBlock()
    of KwIf:
      self.parseIfExpr()
    of KwWhen:
      self.parseWhenExpr()
    of KwFunc:
      self.parseFuncExpr()
    of KwStruct:
      self.parseStruct()
    of KwReturn:
      self.parseReturnExpr()
    of KwModule:
      self.parseModule()
    else:
      self.parsePrimary()

func parseIfExpr(self: var Parser): AstNode =
  func parseIfBranch(self: var Parser): AstNode
    {.nimcall, raises: [ParserError, ValueError, LexerError, FmtLexerError].} =
    let token = self.popToken({KwIf, KwElif})
    let cond  = self.parseExpr()
    let body  = self.parseBlock()

    result = initAstNodeBranch(IfBranch, @[cond, body], token.range)

  func parseElseBranch(self: var Parser): AstNode
    {.nimcall, raises: [ParserError, ValueError, LexerError, FmtLexerError].} =
    let token = self.popToken(KwElse)
    let body  = self.parseBlock()

    result = initAstNodeBranch(ElseBranch, @[body], token.range)

  let token    = self.popToken(KwIf)
  var branches = newSeq[AstNode]()

  while true:
    branches &= self.parseIfBranch()
    if self.peekKind() != KwElif: break

  if self.peekKind() == KwElse:
    branches &= self.parseElseBranch()

  result = initAstNodeBranch(If, branches, token.range)

func parseWhenExpr(self: var Parser): AstNode =
  todo()

func parseReturnExpr(self: var Parser): AstNode =
  todo()

func parseFuncExpr(self: var Parser): AstNode =
  todo()

func parseModule(self: var Parser; isBodyRequired: bool): AstNode =
  debug("parseModule")

  let token = self.popToken(KwModule)
  let name  = self.parseExpr()

  if name.kind != Id:
    raiseParserError(&"expected identifier, got {name} instead", name)

  let body =
    if isBodyRequired:
      self.parseBlock()
    else:
      initAstNodeBranch(Block)

  result = initAstNodeBranch(Module, @[name, body], token.range)

func parseStruct(self: var Parser): AstNode =
  debug("parseStruct")

  let token = self.popToken(KwStruct)
  let name  = self.parseId()

  if name.kind != Id:
    raiseParserError(&"expected identifier, got {name} instead", name)

  template predicate(): bool =
    it.kind != Id or (it.kind == Branch and it.branchKind notin {VarDecl, ValDecl})

  let body = self.parseBlock()
  let invalidNodeIdx =
    body.children.findIt(predicate)

  if invalidNodeIdx >= 0:
    let invalidNode = body.children[invalidNodeIdx]
    raiseParserError(&"unexpected language construction, got {invalidNode}", invalidNode)

  result = initAstNodeBranch(Struct, @[name, body], token.range)

func parseId(self: var Parser): AstNode =
  debug("parseId")

  let token = self.popToken(Id)

  result = initAstNodeId(token.data, token.range)

func parseLiteral(self: var Parser): AstNode =
  let token = self.popToken({IntLit, FloatLit, StringLit, KwTrue, KwFalse, KwNil})
  let lit = case token.kind:
    of IntLit:
      newLit(token.data.parseBiggestInt())
    of FloatLit:
      newLit(token.data.parseFloat())
    of StringLit:
      newLit(token.data)
    of KwTrue:
      newLit(true)
    of KwFalse:
      newLit(false)
    of KwNil:
      newLit(nil)
    else:
      unreachable()

  result = initAstNodeLit(lit, token.range)

func parsePrimary(self: var Parser): AstNode =
#[
ExprList <- (Expr (Comma Expr)* Comma?)?
ExprColonExprList <- (ExprColonExpr (Comma ExprColonExprList)* Comma?)?
ExprEqExprList <- (ExprColonExpr (Comma ExprColonExpr)* Comma?)?
ExprColonExpr <- Expr (':' Expr)?
ExprEqExpr <- Expr ('=' Expr)?

TupleLit <- '(' ExprList ')'
ArrayLit <- '[' ExprColonExprList ']'

Literal <- IntLit | FloatLit | StringLit | BoolLit | Nil
IntLit <- ...
FloatLit <- ...
StringLit <- ...
BoolLit <- 'true' | 'false'
Nil <- 'nil'

Primary <- (Id | Literal) PrimarySuffix*

GroupedExpr <- '(' Expr ')'
]#
  debug("parsePrimary")

  let token = self.peekToken()

  result = case token.kind:
    of OperatorKinds, WordLikeOperatorKinds:
      let op       = token.data.toOperatorKind().get()
      let operator = initAstNodeOperator(op)
      let operand  = self.parsePrimary()
      initAstNodeBranch(Prefix, @[operator, operand])
    of Id:
      self.parseId()
    of LiteralKinds:
      self.parseLiteral()
    of KeywordKinds - WordLikeOperatorKinds:
      raiseParserError("todo", token.range)
    else:
      raiseParserError(&"unexpected token: {token.kind}", token.range)

  result = self.parseSuffix(result)

func parseSuffix(self: var Parser; left: AstNode): AstNode =
#[
PrimarySuffix
  <- '[' ExprColonExprList ']'
   | '{' ExprEqExprList '}'
   | '.' Id
]#
  result = left

  while true:
    debug("parseSuffix: " & $self.peekKind())

    result = case self.peekKind():
      of LeSquare:
        self.parseExprColonExprList(result)
      of LeCurly:
        self.parseExprEqExprList(result)
      of Dot:
        self.parseDotId(result)
      else:
        break

func parseDotId(self: var Parser; left: AstNode): AstNode =
  debug("parseDotId")

  let token = self.popToken(Dot)
  let right = self.parseId()

  result = initAstNodeBranch(ExprDotExpr, @[left, right], token.range)

func parseBlock(self: var Parser): AstNode =
  if self.peekKind() != LeCurly:
    raiseParserError(&"expected block or code, got {self.peekToken()} instead", self.peekToken().range)

  result = initAstNodeBranch(Block)
  let token = self.popToken(LeCurly)

  while self.peekKind() != RiCurly:
    if self.peekKind() == Eof:
      raiseParserError("brace is never closed", token.range)

    if (let expr = self.parseExpr(); expr.kind != Empty):
      result.children &= expr

  self.skipToken(RiCurly)

func parseExprColonExprList(self: var Parser; left: AstNode): AstNode =
  discard

func parseExprEqExprList(self: var Parser; left: AstNode): AstNode =
  discard

#
# API
#

func newParser*(tokens: openArray[Token]; isModule = true; filename = ""): Parser =
  ## The `isModule` parameter specifies that `tokens` should
  ## contain a top-level declaration of the module name
  result = Parser(
    tokens: tokens.toOpenArray(0, tokens.high),
    isModule: isModule,
    filename: filename,
  )

func getAst*(self: Parser): AstNode =
  self.ast

func parseAll*(self: var Parser)
  {.raises: [LexerError, FmtLexerError, ParserError, ValueError].} =
  if self.tokens.len() == 0:
    if self.isModule:
      raiseParserError("this file should have the module name declaration at the beginning", FilePos())
    return

  self.ast =
    if self.isModule:
      self.parseModule(isBodyRequired=true)
    else:
      initAstNodeBranch(Block)

  while self.peekKind() != Eof:
    debug("parseAll: " & $self.peekKind())

    try:
      if (let expr = self.parseExpr(); expr.kind != Empty):
        self.ast.children &= expr
    except ParserError as err:
      discard self.popToken() # maybe skip whole line?
      {.cast(noSideEffect).}:
        try: stderr.write(self.filename & ":" & $err.range.a & ": ")
        except IOError: unreachable()
      error(err.msg)

func parseAll(input: string; offset: tuple[line, column: int]; isModule: bool): AstNode =
  let tokens = input.getAllTokens(offset).normalizeTokens()
  var parser = newParser(tokens, isModule)
  parser.parseAll()
  result = parser.getAst()

func parseExpr(input: string; offset: tuple[line, column: int]; isModule: bool): AstNode =
  let tokens = input.getAllTokens(offset).normalizeTokens()
  var parser = newParser(tokens, isModule)
  result = parser.parseExpr()

{.pop.} # raises: []
