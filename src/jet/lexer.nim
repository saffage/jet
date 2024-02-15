import
  std/strutils,
  std/options,
  std/unicode,

  jet/lexerbase,
  jet/token,

  lib/lineinfo

export
  LexerError

{.push, raises: [].}

type
  Lexer* = object of LexerBase
    curr* : Token
    prev* : Token

const
  IdChars*      = IdentChars
  IdStartChars* = IdentStartChars
  Eol           = {'\0'} + Newlines
  Whitespace    = {' '} + Eol

func buildCharSet(): set[char]
  {.compileTime.} =
  result = {}

  for kind in OperatorKinds:
    for c in $kind:
      result.incl(c)

const
  operatorChars = buildCharSet()

func lexSpace(self: var Lexer): Token =
  let data = self.parseWhile(it == ' ')

  result = Token(kind: Space, data: data)

func lexEndl(self: var Lexer): Token =
  result = Token(kind: Endl)

  while self.handleNewline():
    result.data &= '\n'

func lexId(self: var Lexer): Token =
  let data = self.parseWhile(it in IdChars)
  let kind = data.toTokenKind().get(Id)

  result =
    if kind == Id:
      Token(kind: Id, data: data)
    else:
      Token(kind: kind)

func lexNumber(self: var Lexer): Token
  {.raises: [LexerError].} =
  template withSeparatorCheck(charSet: set[char]): bool =
    block:
      var wasSeparator = false
      if it == '_':
        if wasSeparator:
          raiseLexerError(
            "more than 1 separator in a row is not allowed",
            self.peekPos() .. self.peekPos() + 1)
        wasSeparator = true
        true
      else:
        wasSeparator = false
        it in charSet

  func parseNumberChecked(self: var Lexer; charSet: set[char]): string
    {.raises: [LexerError].} =
    if self.peek() notin charSet:
      raiseLexerError("expected number here", self.peekPos() .. self.peekPos() + 1)

    let startPos = self.peekPos()
    result = self.parseWhile(withSeparatorCheck(charSet))

    if result.startsWith('_'):
      raiseLexerError(
        "leading underscores in number literal is illegal",
        startPos .. startPos + 1)

    if result.endsWith('_'):
      raiseLexerError(
        "trailing underscores in number literal is illegal",
        self.peekPos() .. self.peekPos() + 1)

  let numPart =
    if self.popChar('0'):
      case self.peek()
      of 'x':
        self.pop()
        "0x" & self.parseNumberChecked(HexDigits)
      of 'b':
        self.pop()
        "0b" & self.parseNumberChecked({'0'..'1'})
      of 'o':
        self.pop()
        "0o" & self.parseNumberChecked({'0'..'7'})
      of '.', 'e', 'E':
        "0"
      else:
        raiseLexerError(
          "'0' as first character of a number literal are not allowed",
          self.peekPos() .. self.peekPos() + 1)
    else:
      self.parseNumberChecked(Digits)

  var fracPart = ""
  var expPart = ""

  if self.popChar('.'):
    fracPart &= self.parseNumberChecked(Digits)

  if self.popChar({'e', 'E'}):
    expPart &= 'e'

    if self.peek() in {'-', '+'}:
      expPart &= self.pop()

    expPart &= self.parseNumberChecked(Digits)

  result = Token(kind: IntLit, data: numPart)

  if fracPart.len() > 0:
    result.kind  = FloatLit
    result.data &= '.' & fracPart

  if expPart.len() > 0:
    result.kind  = FloatLit
    result.data &= expPart

func lexComment(self: var Lexer): Token =
  self.pop()

  result =
    if self.peek() == '#' and self.peekOffset(1) in Whitespace:
      self.pop()
      Token(kind: Comment)
    elif self.peek() == '!' and self.peekOffset(1) in Whitespace:
      self.pop()
      Token(kind: CommentModule)
    else:
      self.skipLine()
      return emptyToken

  if self.peek() == ' ':
    self.pop()

  result.data = self.parseUntil(it in Eol)

func lexPunctuation(self: var Lexer): Token =
  let kind = self.pop().toTokenKind().get()

  result = Token(kind: kind)

func lexOperator(self: var Lexer): Token
  {.raises: [LexerError].} =
  let info = self.peekPos()
  let op   = self.parseWhile(it in operatorChars)
  let kind = toTokenKind(op)

  if kind.isNone():
    raiseLexerError("unknown operator: '" & op & "'", info .. info + op.len().uint32)

  result = Token(kind: kind.get())

func lexString(self: var Lexer; raw = false): Token
  {.raises: [LexerError].} =
  let quote =
    if raw: '\"'
    else: '\''
  let info = self.peekPos()
  self.pop()
  let data = self.parseUntil(it == quote and self.peekOffset(-1) != '\\')

  if self.peek() == '\0':
    raiseLexerError("missing closing " & quote, info)

  self.pop()

  let kind =
    if raw: RawStringLit
    else: StringLit

  result = Token(kind: kind, data: data)

func nextToken(self: var Lexer)
  {.raises: [LexerError].} =
  if self.isEmpty():
    self.curr = Token(kind: TokenKind.Eof, range: self.peekPos() .. self.peekPos())
    return

  let prevFilePos = self.peekPos()

  self.curr = case self.peek():
    of '#':
      self.lexComment()
    of IdStartChars:
      self.lexId()
    of Digits:
      self.lexNumber()
    of '&', '@', '$', '(', ')', '{', '}', '[', ']', ',', ';', ':':
      self.lexPunctuation()
    of '.':
      if self.peekOffset(1) in operatorChars:
        self.lexOperator()
      else:
        self.lexPunctuation()
    of operatorChars - {'.'}:
      self.lexOperator()
    of ' ':
      self.lexSpace()
    of Newlines:
      self.lexEndl()
    of '\"':
      self.lexString(raw = true)
    of '\'':
      self.lexString()
    of '\0':
      Token(kind: TokenKind.Eof)
    else:
      raiseLexerError("invalid character: " & strutils.escape($self.peek()), self.peekPos())

  self.curr.range = prevFilePos .. self.prevPos()

func nextTokenNotEmpty(self: var Lexer)
  {.raises: [LexerError].} =
  while true:
    self.nextToken()
    if self.curr.kind != Empty: break

#
# API
#

proc newLexer*(buffer: openArray[char]; offset = (line: 0, column: 0)): Lexer
  {.raises: [LexerError].} =
  result = Lexer(
    buffer: buffer.toOpenArray(0, buffer.high),
    curr: Token(kind: TokenKind.Eof),
    lineOffset: offset.line,
    columnOffset: offset.column,
  )
  result.nextToken()

func getToken*(self: var Lexer): Token
  {.raises: [LexerError].} =
  if self.curr.kind == Empty:
    self.nextTokenNotEmpty()

  self.prev = self.curr

  if self.curr.kind != TokenKind.Eof:
    self.nextToken()

  result = self.prev

func getAllTokens*(self: var Lexer): seq[Token]
  {.raises: [LexerError].} =
  result = @[]

  while true:
    let token = self.getToken()
    result &= token
    if token.kind == TokenKind.Eof: break

func getAllTokens*(input: string; offset = (line: 0, column: 0)): seq[Token]
  {.raises: [LexerError].} =
  var lexer = newLexer(input, offset)
  result = lexer.getAllTokens()

func normalizeTokens*(tokens: seq[Token]): seq[Token] =
  result = @[]
  var prevKind = TokenKind.Eof
  var spaces   = 0
  var wasEndl  = true

  for token in tokens:
    case token.kind
    of Empty:
      continue
    of Endl:
      wasEndl = true
      spaces  = 0
    of Space:
      spaces += token.data.len()
    else:
      var token = token
      token.spaces.wasEndl = wasEndl
      token.spaces.leading =
        if prevKind == Space: spaces
        else: 0

      if result.len() > 0:
        result[^1].spaces.trailing =
          if wasEndl: spacesLast
          else: token.spaces.leading

      result &= token
      wasEndl = false
      spaces  = 0

      if token.kind == Eof:
        result[^1].spaces.trailing = spacesLast
        break

    prevKind = token.kind

{.pop.} # raises: []
