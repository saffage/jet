import
  std/strutils,
  std/options,
  std/parseutils,
  std/unicode,

  jet/token,

  lib/line_info,
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
    info* : LineInfo

const
  Eol        = {'\0'} + Newlines
  Whitespace = {' '} + Eol

template raiseLexerError(message: string; lineInfo: LineInfo = LineInfo()): untyped =
  raise (ref LexerError)(msg: message, info: lineInfo)

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

func lineInfo(self: Lexer): LineInfo =
  ## Returns current line info.
  result = LineInfo(line: self.line().uint32, column: self.column().uint32)

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

    warn("parseWhile: until = " & $until)
    warn("parseWhile: pos = " & $self.pos)
    warn("parseWhile: it = " & escape($it, "'", "'"))

    while self.pos <= until and fn:
      result &= self.pop()
      it = self.peek()
      warn("parseWhile: pos = " & $self.pos)
      warn("parseWhile: it = " & escape($it, "'", "'"))
    
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
          self.lineInfo().withLength(1))
      wasSeparator = true
      true
    else:
      wasSeparator = false
      it in Digits

  let numPart = self.parseWhile(withSeparatorCheck)

  if numPart.endsWith('_'):
    raiseLexerError(
      "trailing underscores in number literal is illegal",
      self.lineInfo().withLength(1))

  # TODO: hex, oct, bin numbers
  # TODO: handle '1e10' notation
  # TODO: suffix

  result = 
    if self.peek() == '.':
      self.pop()
      let dotInfo = self.lineInfo()
      let fracPart = self.parseWhile(withSeparatorCheck)

      if fracPart.len() == 0 or Digits notin fracPart:
        raiseLexerError("expected number after '.' in float literal", dotInfo + 1)
      
      if fracPart.startsWith('_'):
        raiseLexerError(
          "leading underscores in number literal is illegal",
          withLength(dotInfo + 1, 1))
      
      if fracPart.endsWith('_') or fracPart.startsWith('_'):
        raiseLexerError(
          "trailing underscores in number literal is illegal",
          self.lineInfo().withLength(1))
      
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
      return
  
  if self.peek() == ' ':
    self.pop()
  
  result.data = self.parseUntil(it in Eol)

func lexPunctuation(self: var Lexer): Token =
  let kind = self.pop().toTokenKind().get()

  result = Token(kind: kind)

func lexOperator(self: var Lexer): Token
  {.raises: [LexerError].} =
  let info = self.lineInfo()
  let op   = self.parseWhile(it in operatorChars)
  let kind = toTokenKind(op)

  if kind.isNone():
    raiseLexerError("unknown operator: '" & op & "'", info.withLength(op.len().uint32))

  result = Token(kind: kind.get())

func escapeString(s: string; startLineInfo: LineInfo): string
  {.raises: [LexerError].} =
  result = ""
  var i = 0

  while i <= s.high:
    if s[i] == '\\':
      if i == s.high:
        raiseLexerError("invalid character escape; expected character after \\")
      
      let info = startLineInfo + i
      
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
          raiseLexerError("invalid character escape: '\\" & s[i] & "'")
    elif s[i] in PrintableChars:
      result &= s[i]
    else:
      raiseLexerError("invalid character: " & escape($s[i], "'\\", "'"))

    i += 1

func lexString(self: var Lexer): Token
  {.raises: [LexerError].} =
  self.pop()
  let infoInsideLit = self.lineInfo()
  let data = self.parseUntil(it == '\"' and self.peek(-1) != '\\' or it in Eol)
  
  if self.peek() in Eol:
    raiseLexerError("missing closing \"", self.lineInfo())

  self.pop()

  result = Token(kind: StringLit, data: data.escapeString(infoInsideLit))

func lexChar(self: var Lexer): Token
  {.raises: [LexerError].} =
  self.pop()
  let infoInsideLit = self.lineInfo()
  var data = self.parseUntil(it == '\'' and self.peek(-1) != '\\' or it in Eol)
  let origLen = data.len()

  if self.peek() in Eol:
    raiseLexerError("missing closing \'", self.lineInfo())

  self.pop()
  
  if data.len() == 0:
    raiseLexerError("empty character literals are not allowed", infoInsideLit)
  
  if data[0] == '\\':
    data = data.escapeString(infoInsideLit)

  if data.len() > 1:
    raiseLexerError(
      "character is too long; use string literals for UTF-8 characters",
      infoInsideLit.withLength(origLen.uint32))

  result = Token(kind: CharLit, data: data)

func nextToken(self: var Lexer)
  {.raises: [LexerError].} =
  if self.pos > self.buffer.high:
    self.curr = Token(kind: TokenKind.Eof, info: self.lineInfo())
    return

  warn("pos = " & $self.pos & "; " & escape($self.peek(), "next: '", "'"))
  var prevLineInfo = self.lineInfo()
  let oldPos = self.pos
  
  let token = case self.peek():
    of '#':
      self.lexComment()
    of IdStartChars:
      self.lexId()
    of Digits:
      self.lexNumber()
    of '@', '$', '(', ')', '{', '}', '[', ']', ',', ';', ':':
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
      self.lexString()
    of '\'':
      self.lexChar()
    of '\0':
      Token(kind: TokenKind.Eof, info: self.lineInfo())
    else:
      raiseLexerError("invalid character: " & strutils.escape($self.peek()), self.lineInfo())

  prevLineInfo.length = uint32(self.pos - oldPos)
  self.curr           = token
  self.curr.info      = prevLineInfo

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
  var wasLF    = false

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

      case prevKind
      of VSpace, Eof:
        token.spaces.leading = 0
      of HSpace:
        token.spaces.leading = spaces
      else:
        discard

      if result.len() > 0:
        result[^1].spaces.trailing =
          if wasLF: spacingLast
          else: token.spaces.leading

      result &= token
      wasLF = false
      spaces = 0

      if token.kind == Eof:
        result[^1].spaces.trailing = spacingLast
        break
    prevKind = token.kind
