import
  std/strutils,
  std/parseutils,

  lib/lineinfo,
  lib/utils

{.push, raises: [].}

type
  LexerBase* = object of RootObj
    buffer*       : openArray[char] ## Content of the file
    idx*          : int = 0         ## Index of the current character in the buffer
    idxStartPrev* : int = 0         ## Index of the first character in the current line
    idxStart*     : int = 0         ## Index of the first character in the current line
    lineNum*      : int = 1         ## Current line number
    lineOffset*   : int             ## Used in `line` function
    columnOffset* : int             ## Used in `column` function

  LexerError* = object of CatchableError
    range* : FileRange

template raiseLexerError*(message: string; fileRange: FileRange) =
  raise (ref LexerError)(msg: message, range: fileRange)

template raiseLexerError*(message: string; filePos: FilePos) =
  raise (ref LexerError)(msg: message, range: filePos .. filePos)

func isEmpty*(self: LexerBase): bool =
  ## Returns *true* when `buffer` is fully processed
  result = self.idx > self.buffer.high

func peekOffset*(self: LexerBase; offset: int): char =
  ## Returns current character in the `buffer` with specified `offset`
  assert(self.idx + offset >= 0)

  result =
    if self.idx + offset > self.buffer.high:
      '\0'
    else:
      self.buffer[self.idx + offset]

func peek*(self: LexerBase): char =
  ## Returns current character in the `buffer`
  result = self.peekOffset(0)

func pop*(self: var LexerBase): char
  {.discardable.} =
  ## Returns current character in the `buffer` and increments an index
  result = self.peek()
  if result != '\0': self.idx += 1

func popChar*(self: var LexerBase; charSet: set[char]): bool
  {.discardable.} =
  ## Returns *true* if the current character in the `buffer` is `c`, otherwise *false* \
  ## Character will be popped from the buffer
  result =
    if self.peek() in charSet:
      self.pop()
      true
    else:
      false

func popChar*(self: var LexerBase; c: char): bool
  {.discardable.} =
  ## Returns *true* if the current character in the `buffer` is `c`, otherwise *false* \
  ## Character will be popped from the buffer
  result = self.popChar({c})

func linePrev*(self: LexerBase): int =
  ## Returns a current line
  result = self.lineNum - int(self.idx == self.idxStart) + self.lineOffset

func line*(self: LexerBase): int =
  ## Returns a current line
  result = self.lineNum + self.lineOffset

func columnPrev*(self: LexerBase): int =
  ## Returns a column in the current line
  let lineStart =
    if self.idx == self.idxStart:
      self.idxStartPrev
    else:
      self.idxStart

  result = self.idx - lineStart + self.columnOffset

func column*(self: LexerBase): int =
  ## Returns a column in the current line
  result = (self.idx + 1) - self.idxStart + self.columnOffset

func peekPos*(self: LexerBase): FilePos =
  ## Returns current character position
  result = FilePos(line: self.line().uint32, column: self.column().uint32)

func prevPos*(self: LexerBase): FilePos =
  ## Returns previous character position
  result = FilePos(line: self.linePrev().uint32, column: self.columnPrev().uint32)

func handleNewline*(self: var LexerBase): bool
  {.discardable.} =
  result =
    if self.peek() in Newlines:
      self.popChar('\r')
      self.popChar('\n')
      self.lineNum     += 1
      self.idxStartPrev = self.idxStart
      self.idxStart     = self.idx
      true
    else:
      false

func skipLine*(self: var LexerBase) =
  while self.peek() notin Newlines:
    self.idx += 1

func getLine*(self: LexerBase; lineNum: Positive): string
  {.raises: [ValueError].} =
  ## Returns line content at `lineNum` from the `buffer`
  ##
  ## Note: new line character are not included
  var lexer = LexerBase(buffer: self.buffer.toOpenArray(0, self.buffer.high))
  var i     = 1
  var start = 0
  result    = ""

  while true:
    lexer.skipLine()
    if i == lineNum:
      result = self.buffer.toString(start..<self.idx)
      break
    if not lexer.handleNewline():
      raise newException(ValueError, "line " & $lineNum & " does not exist in the buffer")
    start = lexer.idx
    i += 1

func getLines*(self: LexerBase; lineNums: openArray[int] | Slice[int]): seq[string]
  {.raises: [ValueError].} =
  ## Returns specified lines content from the `buffer`
  result = @[]
  for line in lineNums:
    result &= self.getLine(line)

template parseWhile*(self: var LexerBase; predicate: untyped; bufSize = 0.Natural): string =
  block:
    var result = newStringOfCap(bufSize)
    var it {.inject.} = self.peek()

    while self.idx <= self.buffer.high and predicate:
      result &= self.pop()
      it = self.peek()

    result

template parseUntil*(self: var LexerBase; predicate: untyped; bufSize = 0.Natural): string =
  self.parseWhile(not(predicate), bufSize)

{.pop.} # raises: []
