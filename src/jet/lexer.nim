import
  std/strutils,
  std/options,
  std/unicode,

  jet/lexerbase,
  jet/token,

  lib/lineinfo,
  lib/utils

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

func lexHSpace(self: var Lexer): Token =
  let data = self.parseWhile(it == ' ')

  result = Token(kind: HSpace, data: data)

func lexVSpace(self: var Lexer): Token =
  result = Token(kind: VSpace)

  while self.handleNewline():
    result.data &= '\n'

func lexId(self: var Lexer): Token =
  let data = self.parseWhile(it in IdChars)
  let kind = data.toTokenKind().get(Id)

  result =
    if kind == Id: Token(kind: Id, data: data)
    else: Token(kind: kind)

func lexNumber(self: var Lexer): Token
  {.raises: [LexerError].} =
  template withSeparatorCheck(charSet: set[char]): bool =
    block:
      var wasSeparator = false
      if it == '_':
        if wasSeparator:
          raiseLexerError(
            "more than 1 separator in a row is not allowed",
            self.peekPos().withLength(1))
        wasSeparator = true
        true
      else:
        wasSeparator = false
        it in charSet

  func parseNumberChecked(self: var Lexer; charSet: set[char]): string
    {.raises: [LexerError].} =
    if self.peek() notin charSet:
      raiseLexerError("expected number here", self.peekPos().withLength(1))

    let startPos = self.peekPos()
    result = self.parseWhile(withSeparatorCheck(charSet))

    if result.startsWith('_'):
      raiseLexerError(
        "leading underscores in number literal is illegal",
        startPos.withLength(1))

    if result.endsWith('_'):
      raiseLexerError(
        "trailing underscores in number literal is illegal",
        self.peekPos().withLength(1))

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
          self.peekPos().withLength(1))
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
    raiseLexerError("unknown operator: '" & op & "'", info.withLength(op.len().uint32))

  result = Token(kind: kind.get())

func escapeString(s: string; startPos: FilePosition): string
  {.raises: [LexerError].} =
  result = ""
  var i = 0

  while i <= s.high:
    let info = startPos.withOffset(i)

    if s[i] == '\\':
      if i == s.high:
        raiseLexerError(
          "invalid character escape; expected character after `\\`, got end of string literal",
          info)

      i += 1
      result.add case s[i]:
        of 'n': "\n"
        of 'r': "\r"
        of 't': "\t"
        of '\\': "\\"
        of '\'': "\'"
        of '\"': "\""
        of 'x', 'u', 'U': todo()
        of Digits:
          if i+2 <= s.high and s[i+1] in Digits and s[i+2] in Digits:
            var num = 0
            num = (num * 10) + (ord(s[i+0]) - ord('0'))
            num = (num * 10) + (ord(s[i+1]) - ord('0'))
            num = (num * 10) + (ord(s[i+2]) - ord('0'))

            if num > 255:
              raiseLexerError(
                "invalid character escape; constant must be in range 0..255",
                info.withLength(4))

            i += 2
            $char(num)
          else:
            if s[1] != '0':
              raiseLexerError(
                "invalid character escape: '" & s[1] & "'",
                info.withLength(2))

            "\0"
        else:
          raiseLexerError("invalid character escape: '\\" & s[i] & "'", info.withOffset(-1).withLength(2))
    elif s[i] in PrintableChars:
      result &= s[i]
    else:
      raiseLexerError("invalid character: " & escape($s[i], "'\\", "'"), info)

    i += 1

func lexString(self: var Lexer; raw = false): Token
  {.raises: [LexerError].} =
  let quote =
    if raw: '\"'
    else: '\''
  let info = self.peekPos()
  self.pop()

  let infoInsideLit = self.peekPos()
  let data = self.parseUntil(it == quote and self.peekOffset(-1) != '\\')

  if self.peek() == '\0':
    raiseLexerError("missing closing " & quote, info)
  self.pop()

  result =
    if raw: Token(kind: StringLit, data: data)
    else: Token(kind: StringLit, data: data.escapeString(infoInsideLit))

func nextToken(self: var Lexer)
  {.raises: [LexerError].} =
  if self.idx > self.buffer.high:
    self.curr = Token(kind: TokenKind.Eof, rng: self.peekPos().withLength(0))
    return

  var prevFilePos = self.peekPos()
  let oldPos  = self.idx

  let token = case self.peek():
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
      self.lexHSpace()
    of Newlines:
      self.lexVSpace()
    of '\"':
      self.lexString(raw = true)
    of '\'':
      self.lexString()
    of '\0':
      Token(kind: TokenKind.Eof, rng: self.peekPos().withLength(0))
    else:
      raiseLexerError("invalid character: " & strutils.escape($self.peek()), self.peekPos())

  let rng = prevFilePos.withLength(self.idx - oldPos)
  self.curr     = token
  self.curr.rng = rng

func nextTokenNotEmpty(self: var Lexer)
  {.raises: [LexerError].} =
  self.nextToken()
  while self.curr.kind == Empty: self.nextToken()

#
# API
#

proc newLexer*(buffer: openArray[char]; posOffset = emptyFilePos): Lexer
  {.raises: [LexerError].} =
  result = Lexer(
    buffer: buffer.toOpenArray(0, buffer.high),
    curr: Token(kind: TokenKind.Eof),
    posOffset: posOffset,
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

func getAllTokens*(input: string; posOffset = emptyFilePos): seq[Token]
  {.raises: [LexerError].} =
  var lexer = newLexer(input, posOffset)
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
    of VSpace:
      wasEndl = true
      spaces  = 0
    of HSpace:
      spaces += token.data.len()
    else:
      var token = token
      token.spaces.wasEndl = wasEndl
      token.spaces.leading =
        if prevKind == HSpace: spaces
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
