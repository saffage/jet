import
  std/strutils,
  std/options,
  std/parseutils,
  std/unicode,

  jet/token,

  lib/lineinfo,
  lib/utils

{.push, raises: [].}

type
  Lexer* {.byref.} = object
    buffer  : openArray[char]
    pos     : int = 0           ## Position in the buffer
    linePos : int = 0           ## Position in the buffer of the line start character
    lineNum : int = 1
    curr*   : Token
    prev*   : Token

  LexerError* = object of CatchableError
    rng* : FileRange

const
  Eol        = {'\0'} + Newlines
  Whitespace = {' '} + Eol

template raiseLexerError(message: string; fileRange: FileRange): untyped =
  raise (ref LexerError)(msg: message, rng: fileRange)

template raiseLexerError(message: string; filePos: FilePosition): untyped =
  raise (ref LexerError)(msg: message, rng: filePos.withLength(0))

func peek(self: Lexer; offset: int = 0): char =
  ## Returns character at the position `pos + offset` in the `buffer`.
  assert(self.pos + offset >= 0)
  result =
    if self.pos + offset > self.buffer.high:
      '\0'
    else:
      self.buffer[self.pos + offset]

func pop(self: var Lexer): char
  {.discardable.} =
  ## Returns character at the position `pos` in the `buffer`
  ## and increments `pos`.
  result = self.peek()
  if result != '\0': self.pos += 1

func line(self: Lexer): int =
  ## Returns line number.
  result = self.lineNum

func column(self: Lexer): int =
  ## Returns column number in the current line.
  result = (self.pos + 1) - self.linePos

func peekPos(self: Lexer): FilePosition =
  ## Returns current character position.
  result = FilePosition(line: self.line().uint32, column: self.column().uint32)

func handleNewline(self: var Lexer): bool
  {.discardable.} =
  result = self.peek() in Newlines

  if result:
    debug("lexer: new line at pos " & $self.pos)

    if self.peek() == '\r':
      self.pos += 1

    if self.peek() == '\n':
      self.pos += 1

    self.lineNum += 1
    self.linePos  = self.pos

func skipLine*(self: var Lexer) =
  while self.peek() notin Newlines:
    self.pos += 1

func getLine*(self: Lexer; lineNum: Positive): string
  {.raises: [ValueError], warning[ProveInit]: off.} =
  ## Returns line at `line` in `buffer` (new line character excluded).
  var i = 1
  # TODO: use `find`
  for line in ($self.buffer).splitLines():
    if i == lineNum: return line
    i += 1
  raise (ref ValueError)(msg: "line " & $lineNum & "does not exist")

func getLines*(self: Lexer; lineNums: openArray[int] | Slice[int]): seq[string]
  {.raises: [ValueError].} =
  ## Returns specified lines from the `buffer`.
  result = @[]
  for line in lineNums:
    result &= self.getLine(line)

template parseWhile(self: var Lexer; fn: untyped; startOffset = 0; lastIdx = -1): string =
  block:
    var result = newStringOfCap(16)
    let until  = if lastIdx < 0: self.buffer.high else: min(self.buffer.high, lastIdx)

    self.pos += startOffset
    var it {.inject.} = self.peek()

    while self.pos <= until and fn:
      result &= self.pop()
      it = self.peek()

    result

template parseUntil(self: var Lexer; fn: untyped; startOffset = 0; lastIdx = -1): string =
  self.parseWhile(not(fn), startOffset, lastIdx)

const
  IdChars*          = IdentChars
  IdStartChars*     = IdentStartChars
  # PrefixWhitelist*  = {' ', ',', ';', '(', '[', '{'} + Newlines
  # PostfixWhitelist* = {' ', ',', ';', ')', ']', '}', '#'} + Newlines

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
  var wasSeparator = false

  template withSeparatorCheck(): bool =
    debug($wasSeparator)
    if it == '_':
      if wasSeparator:
        raiseLexerError(
          "more than 1 separator in a row is not allowed",
          self.peekPos().withLength(1))
      wasSeparator = true
      true
    else:
      wasSeparator = false
      it in Digits

  let numPart = self.parseWhile(withSeparatorCheck)

  if numPart.endsWith('_'):
    raiseLexerError(
      "trailing underscores in number literal is illegal",
      self.peekPos().withLength(1))

  # TODO: hex, oct, bin numbers
  # TODO: handle '1e10' notation

  result =
    if self.peek() == '.':
      self.pop()
      let dotInfo = self.peekPos()
      let fracPart = self.parseWhile(withSeparatorCheck)

      if fracPart.len() == 0 or Digits notin fracPart:
        raiseLexerError("expected number after '.' in float literal", dotInfo.withLength(1))

      if fracPart.startsWith('_'):
        raiseLexerError(
          "leading underscores in number literal is illegal",
          dotInfo.withOffset(1).withLength(1))

      if fracPart.endsWith('_') or fracPart.startsWith('_'):
        raiseLexerError(
          "trailing underscores in number literal is illegal",
          self.peekPos().withLength(1))

      Token(kind: FloatLit, data: numPart & '.' & fracPart)
    else:
      Token(kind: IntLit, data: numPart)

func lexComment(self: var Lexer): Token =
  self.pop()

  result =
    if self.peek() == '#' and self.peek(1) in Whitespace:
      self.pop()
      Token(kind: Comment)
    elif self.peek() == '!' and self.peek(1) in Whitespace:
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

  # TODO: parse string literals inside `${}`
  let infoInsideLit = self.peekPos()
  let data = self.parseUntil(it == quote and self.peek(-1) != '\\')

  if self.peek() == '\0':
    raiseLexerError("missing closing " & quote, info)
  self.pop()

  result =
    if raw: Token(kind: StringLit, data: data)
    else: Token(kind: StringLit, data: data.escapeString(infoInsideLit))

func nextToken(self: var Lexer)
  {.raises: [LexerError].} =
  if self.pos > self.buffer.high:
    self.curr = Token(kind: TokenKind.Eof, rng: self.peekPos().withLength(0))
    return

  var prevFilePos = self.peekPos()
  let oldPos  = self.pos

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
      if self.peek(1) in operatorChars:
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

  let rng = prevFilePos.withLength(self.pos - oldPos)
  self.curr     = token
  self.curr.rng = rng

func nextTokenNotEmpty(self: var Lexer)
  {.raises: [LexerError].} =
  self.nextToken()
  while self.curr.kind == Empty: self.nextToken()

#
# API
#

proc newLexer*(buffer: openArray[char]): Lexer
  {.raises: [LexerError].} =
  result = Lexer(buffer: buffer.toOpenArray(0, buffer.high), curr: Token(kind: TokenKind.Eof))
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

func normalizeTokens*(tokens: seq[Token]): seq[Token] =
  result = @[]
  var prevKind = TokenKind.Eof
  var spaces   = 0
  var wasLF    = true

  for token in tokens:
    case token.kind
    of Empty:
      continue
    of VSpace:
      wasLF = true
      spaces = 0
    of HSpace:
      spaces += token.data.len()
    else:
      var token = token
      token.spaces.wasLF = wasLF
      token.spaces.leading =
        if prevKind == HSpace: spaces
        else: 0

      if result.len() > 0:
        result[^1].spaces.trailing =
          if wasLF: spacesLast
          else: token.spaces.leading

      result &= token
      wasLF = false
      spaces = 0

      if token.kind == Eof:
        result[^1].spaces.trailing = spacesLast
        break
    prevKind = token.kind

{.pop.} # raises: []
