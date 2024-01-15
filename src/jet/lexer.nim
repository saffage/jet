import
  std/strutils,
  std/options,
  std/parseutils,

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
  EOL        = {'\0'} + Newlines  ## End of Line characters
  WHITESPACE = {' '} + EOL

func peek(self: Lexer): char =
  ## Returns character as the position `pos` in the `buffer`.
  result =
    if self.pos > self.buffer.high: '\0'
    else: self.buffer[self.pos]

func slice(self: Lexer): openArray[char] =
  ## Returns buffer content from `pos` to the end of the buffer.
  assert(self.pos <= self.buffer.high)
  result = self.buffer.toOpenArray(self.pos, self.buffer.high)

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
    if self.peek() == '\r':
      self.pos += 1

    if self.peek() == '\n':
      self.pos += 1

    self.lineNum += 1
    self.linePos  = self.pos

func skipLine*(self: var Lexer) =
  while self.peek() notin EOL:
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

const
  BinChars*         = {'0'..'1'}
  OctChars*         = {'0'..'7'}
  HexChars*         = {'0'..'9', 'a'..'f', 'A'..'F'}
  IdChars*          = IdentChars
  IdStartChars*     = IdentStartChars
  PrefixWhitelist*  = {' ', ',', ';', '(', '[', '{'} + EOL
  PostfixWhitelist* = {' ', ',', ';', ')', ']', '}', '#'} + EOL

func buildCharSet(): set[char]
    {.compileTime.} =
    result = {}
    for kind in OperatorKinds:
      for c in $kind:
        result.incl(c)

const
  operatorChars = buildCharSet()

func escape*(self: string): string =
  result = self.multiReplace(
    ("\'", "\\'"),
    ("\"", "\\\""),
    ("\\", "\\\\"),
    ("\0", "\\0"),
    ("\t", "\\t"),
    ("\n", "\\n"),
    ("\r", "\\r"),
  )

# Why STD does not provide this functions?
func parseWhile(s: openArray[char]; validChars: set[char]): string =
  result = ""
  discard s.parseWhile(result, validChars)

func parseWhile(s: openArray[char]; validChar: char): string =
  result = parseWhile(s, {validChar})

func parseUntil(s: openArray[char]; until: set[char]): string =
  result = ""
  discard s.parseUntil(result, until)

func parseUntil(s: openArray[char]; until: char): string =
  result = parseUntil(s, {until})

func parseWhileWithSeparator(
  s: openArray[char];
  charSet: set[char];
  separator: char;
  allowDoubledSeparator: bool = false
): string {.raises: [ValueError].} =
  result = ""
  var wasSeparator = false
  for c in s:
    if c == separator:
      result &= c
      if wasSeparator and not allowDoubledSeparator:
        raise (ref ValueError)(msg: "more than 1 separator in a row is not allowed")
      wasSeparator = true
    elif c in charSet:
      result &= c
      wasSeparator = false
    else:
      break

func lexSpace(buffer: openArray[char]; parsed: out uint): Token =
  result = emptyToken

  let kind = if buffer[0] == ' ': HSpace else: VSpace
  let data = buffer.parseWhile(buffer[0])

  parsed = data.len().uint

  result = initToken(kind, data)

func lexId(buffer: openArray[char]; parsed: out uint): Token =
  result = emptyToken

  let data = buffer.parseWhile(IdChars)
  let kind = data.toTokenKind().get(Id)

  parsed = data.len().uint

  result =
    if kind == Id:
      initToken(Id, data)
    else:
      initToken(kind)

func lexNumber(buffer: openArray[char]; parsed: out uint): Token
  {.raises: [LexerError].} =
  result = emptyToken

  let numPart = try:
    buffer.parseWhileWithSeparator(Digits, '_')
  except ValueError as e:
    raise (ref LexerError)(msg: e.msg)

  if numPart.endsWith('_'):
    raise (ref LexerError)(msg: "trailing underscore in number literal is illegal")

  # TODO: hex, oct, bin numbers
  # TODO: handle '1e10' notation
  # TODO: suffix

  if numPart.len() <= buffer.high and buffer[numPart.len()] == '.':
    let fracPart = buffer
      .toOpenArray(numPart.len() + 1, buffer.high)
      .parseWhile(Digits)

    if fracPart.len() == 0:
      raise (ref LexerError)(msg: "expected number after '.' in float literal")
    else:
      parsed = numPart.len().uint + fracPart.len().uint + 1
      result = initToken(FloatLit, numPart & "." & fracPart)
  else:
    parsed = numPart.len().uint
    result = initToken(IntLit, numPart)

func lexComment(buffer: openArray[char]; parsed: out uint): Token =
  result = emptyToken

  if buffer[1] == '#' and buffer[2] in WHITESPACE:
    result = initToken(Comment)
  elif buffer[1] == '!' and buffer[2] in WHITESPACE:
    result = initToken(CommentModule)
  else:
    parsed = buffer.findIt(it == '\n').uint
    return

  let skipPrefix = if buffer[2] in EOL: 2 else: 3

  result.data = buffer
    .toOpenArray(skipPrefix, buffer.high)
    .parseUntil(Newlines)

  parsed = uint(skipPrefix + result.data.len())

func lexPunctuation(buffer: openArray[char]; parsed: out uint): Token =
  let kind = case buffer[0]:
    of '(': LeRound
    of ')': RiRound
    of '{': LeCurly
    of '}': RiCurly
    of '[': LeSquare
    of ']': RiSquare
    of ',': Comma
    of ';': Semicolon
    else: return

  parsed = 1
  result = initToken(kind)

func lexOperator(buffer: openArray[char]; parsed: out uint): Token
  {.raises: [LexerError].} =
  let op = buffer.parseWhile(operatorChars)
  parsed = op.len().uint

  let kind = toTokenKind(op)

  if kind.isNone():
    raise (ref LexerError)(msg: "unknown operator: '" & op & "'")

  result = initToken(kind.get())

func lexOperatorSpecial(buffer: openArray[char]; parsed: out uint): Token =
  let kind = case buffer[0]:
    of '@': At
    of '$': Dollar
    else: return

  parsed = 1
  result = initToken(kind)

func nextToken(self: var Lexer)
  {.raises: [LexerError].} =
  if self.pos > self.buffer.high:
    self.curr = initToken(TokenKind.Eof, info = self.lineInfo())
    return

  var prevLineInfo = self.lineInfo()
  var parsed = 1'u
  let token = case self.peek():
    of '#':
      self.slice().lexComment(parsed)
    of IdStartChars:
      self.slice().lexId(parsed)
    of Digits:
      self.slice().lexNumber(parsed)
    of '(', ')', '{', '}', '[', ']', ',', ';':
      self.slice().lexPunctuation(parsed)
    of '@', '$':
      self.slice().lexOperatorSpecial(parsed)
    of operatorChars:
      self.slice().lexOperator(parsed)
    of ' ', Newlines:
      self.slice().lexSpace(parsed)
    of '\0':
      initToken(TokenKind.Eof, info = self.lineInfo())
    of '\t':
      raise (ref LexerError)(msg: "tabs are not allowed")
    else:
      raise (ref LexerError)(msg: "invalid character: " & strutils.escape($self.peek()), info: self.lineInfo())

  if self.handleNewline():
    while self.handleNewline(): discard
  else:
    self.pos += parsed.int

  prevLineInfo.length = parsed.uint32
  self.curr           = token
  self.curr.info      = prevLineInfo

func nextTokenNotEmpty(self: var Lexer)
  {.raises: [LexerError].} =
  self.nextToken()
  while self.curr.kind == Empty: self.nextToken()

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

  for token in tokens:
    if token.kind == prevKind and prevKind in {VSpace, HSpace}:
      result[^1].data &= token.data
    elif token.kind != Empty:
      result &= token
      prevKind = token.kind

#
# API
#

proc newLexer*(buffer: openArray[char]): Lexer
  {.raises: [LexerError].} =
  result = Lexer(buffer: buffer.toOpenArray(0, buffer.high), curr: initToken(TokenKind.Eof))
  result.nextToken()


when isMainModule:
  let file = open("tests_local/lexer/test.jet", fmRead).readAll()

  echo "---"
  echo file
  echo "---"

  maxErrors = 3

  var lexer = newLexer(file)
  var tok = lexer.getToken()

  while tok.kind != Eof:
    echo(tok.human())
    tok = lexer.getToken()
