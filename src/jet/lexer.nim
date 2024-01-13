import
  std/strformat,
  std/strutils,
  std/options,
  std/parseutils,

  jet/token,

  lib/line_info,
  lib/utils,

  pkg/results

type Lexer* = object
  buffer  : openArray[char]
  pos     : int = 0           ## Position in the buffer
  linePos : int = 0           ## Position in the buffer of the line start character
  lineNum : int = 1
  curr*   : Token
  prev*   : Option[Token] = none(Token)

type LexResult = object
  token  : Result[Token, string]
  parsed : uint = 1

const
  lexResultDefault = LexResult(token: Result[Token, string].ok(initToken(Empty)))

const
  SPACE      = ' '  ## Just a whitespace char, nothing special
  EOF        = '\0' ## End of File
  CR         = '\r' ## Carriage Return
  LF         = '\n' ## Line Feed
  TAB        = '\t' ## Tabulation (horisontal)
  EOL        = {EOF} + Newlines  ## End of Line characters
  WHITESPACE = {SPACE} + EOL

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
    if self.peek() == CR:
      self.pos += 1

    if self.peek() == LF:
      self.pos += 1

    self.lineNum += 1
    self.linePos  = self.pos

func skipLine(self: var Lexer) =
  while self.peek() notin EOL:
    self.pos += 1

func getLine*(self: Lexer; lineNum: Positive): Option[string] =
  ## Returns line at `line` in `buffer` (new line character excluded).
  result = none(string)
  var i = 1
  for line in ($self.buffer).splitLines():
    if i == lineNum:
      result = some(line)
      break
    i += 1

func getLines*(self: Lexer; lineNums: openArray[int] | Slice[int]): Option[seq[string]] =
  ## Returns specified lines from the `buffer`.
  result = some(newSeq[string]())
  for line in lineNums:
    let returnedLine = self.getLine(line)
    if returnedLine.isNone():
      result = none(seq[string])
      break
    result.get() &= returnedLine.get()

const
  BinChars*         = {'0'..'1'}
  OctChars*         = {'0'..'7'}
  HexChars*         = {'0'..'9', 'a'..'f', 'A'..'F'}
  IdChars*          = IdentChars
  IdStartChars*     = IdentStartChars
  PrefixWhitelist*  = {' ', ',', ';', '(', '[', '{'} + EOL
  PostfixWhitelist* = {' ', ',', ';', ')', ']', '}', '#'} + EOL

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

func parseWhileWithSeparator(s: openArray[char]; charSet: set[char]; separator: char): (string, bool) =
  result = ("", true)
  var wasSeparator = false
  for c in s:
    if c == separator:
      result[0] &= c
      if wasSeparator:
        result[1] = false
        return
      wasSeparator = true
    elif c in charSet:
      result[0] &= c
      wasSeparator = false
    else:
      break

func lexSpace(buffer: openArray[char]): LexResult =
  result = lexResultDefault

  let kind = if buffer[0] == SPACE: HSpace else: VSpace
  let data = buffer.parseWhile(buffer[0])
  result.parsed = data.len().uint
  result.token.ok(initToken(kind, data))

func lexId(buffer: openArray[char]): LexResult =
  result = lexResultDefault

  let data = buffer.parseWhile(IdChars)
  let kind = data.toTokenKind().get(Id)

  result.parsed = data.len().uint
  result.token.ok do:
    if kind == Id:
      initToken(Id, data)
    else:
      initToken(kind)

func lexNumber(buffer: openArray[char]): LexResult =
  result = lexResultDefault

  let (numPart, wasUnderscoreDoubled) = buffer.parseWhileWithSeparator(Digits, '_')

  if not wasUnderscoreDoubled:
    result.token.err("double underscore in number literal is illegal")
    return

  if numPart.endsWith('_'):
    result.token.err("trailing underscore in number literal is illegal")
    return

  # TODO: hex, oct, bin numbers
  # TODO: handle '1e10' notation
  # TODO: suffix

  result.token =
    if numPart.len() <= buffer.high and buffer[numPart.len()] == '.':
      let fracPart = buffer
        .toOpenArray(numPart.len() + 1, buffer.high)
        .parseWhile(Digits)
      if fracPart.len() == 0:
        Result[Token, string].err("expected number after '.' in float literal")
      else:
        result.parsed = numPart.len().uint + fracPart.len().uint + 1
        Result[Token, string].ok(initToken(FloatLit, &"{numPart}.{fracPart}"))
    else:
      result.parsed = numPart.len().uint
      Result[Token, string].ok(initToken(IntLit, numPart))

func lexComment(buffer: openArray[char]): LexResult =
  result = lexResultDefault

  if buffer[1] == '#' and buffer[2] in WHITESPACE:
    result.token.ok(initToken(Comment))
  elif buffer[1] == '!' and buffer[2] in WHITESPACE:
    result.token.ok(initToken(CommentModule))
  else:
    result.parsed = buffer.findIt(it == LF).uint
    return

  let skipPrefix = if buffer[2] in EOL: 2 else: 3
  result.token.unsafeValue().data = buffer
    .toOpenArray(skipPrefix, buffer.high)
    .parseUntil(Newlines)
  result.parsed = uint(skipPrefix + result.token.unsafeValue().data.len())

func lexPunctuation(buffer: openArray[char]): LexResult =
  result = lexResultDefault

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

  result.token.ok(initToken(kind))

func lexOperatorSpecial(buffer: openArray[char]): LexResult =
  result = lexResultDefault

  let kind = case buffer[0]:
    of '@': At
    of '$': Dollar
    else: return

  result.token.ok(initToken(kind))

func nextToken(self: var Lexer) =
  if self.pos > self.buffer.high:
    self.curr = initToken(TokenKind.Eof, info = self.lineInfo())
    return

  var prevLineInfo = self.lineInfo()
  let lexResult = case self.peek():
    of '#':
      self.slice().lexComment()
    of IdStartChars:
      self.slice().lexId()
    of Digits:
      self.slice().lexNumber()
    of '(', ')', '{', '}', '[', ']', ',', ';':
      self.slice().lexPunctuation()
    of '@', '$':
      self.slice().lexOperatorSpecial()
    of SPACE, Newlines:
      self.slice().lexSpace()
    of EOF:
      LexResult(token: Result[Token, string].ok(initToken(TokenKind.Eof, info = self.lineInfo())))
    of TAB:
      LexResult(token: Result[Token, string].err("tabs are not allowed"))
    else:
      LexResult(token: Result[Token, string].err(&"invalid character: {strutils.escape($self.peek())}"))

  if self.handleNewline():
    while self.handleNewline(): discard
  else:
    self.pos += lexResult.parsed.int

  if lexResult.token.isErr():
    error(lexResult.token.error(), self.lineInfo())
    self.skipLine()
    self.nextToken()
  else:
    prevLineInfo.length = lexResult.parsed.uint32
    self.curr           = lexResult.token.value()
    self.curr.info      = prevLineInfo

func getToken*(self: var Lexer): Option[Token] =
  if self.prev.isSome() and self.prev.get().kind == TokenKind.Eof:
    return none(Token)

  self.prev = some(self.curr)

  if self.curr.kind != TokenKind.Eof:
    self.nextToken()
    while self.curr.kind == Empty: self.nextToken()

  result = self.prev

func getAllTokens*(self: var Lexer): seq[Token] =
  result = @[]


# ----- [] ----- #
proc newLexer*(buffer: openArray[char]): Lexer =
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

  while tok.isSome():
    echo tok.get().human()
    tok = lexer.getToken()
