import
  std/strutils,
  std/parseutils,

  lib/lineinfo,
  lib/utils

{.push, raises: [].}

const
  initialFilePos* = FilePosition(line: 1, column: 1)

type
  LexerBase* = object of RootObj
    buffer*    : openArray[char]              ## Content of the file
    idx*       : int = 0                      ## Index of the current character in the buffer
    idxStart*  : int = 0                      ## Index of the first character in the buffer
    lineNum*   : int = 1                      ## Position of current character in the buffer
    posOffset* : FilePosition = emptyFilePos  ## Used in `line` & `column` functions

  LexerError* = object of CatchableError
    rng* : FileRange

template raiseLexerError*(message: string; fileRange: FileRange) =
  raise (ref LexerError)(msg: message, rng: fileRange)

template raiseLexerError*(message: string; filePos: FilePosition) =
  raise (ref LexerError)(msg: message, rng: filePos.withLength(0))

func isEmpty*(self: LexerBase): bool =
  ## Returns *true* when `buffer` is fully processed
  result = self.idx > self.buffer.high

func peekOffset*(self: LexerBase; offset: int): char =
  ## Returns current character in the `buffer` with specified `offset`.
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

func popChar*(self: var LexerBase; c: set[char]): bool
  {.discardable.} =
  ## Returns *true* if the current character in the `buffer` is `c`, otherwise *false* \
  ## Character will be popped from the buffer
  result =
    if self.peek() in c:
      self.pop()
      true
    else:
      false

func popChar*(self: var LexerBase; c: char): bool
  {.discardable.} =
  ## Returns *true* if the current character in the `buffer` is `c`, otherwise *false* \
  ## Character will be popped from the buffer
  result = self.popChar({c})

func line(self: LexerBase): int =
  ## Returns a current line
  result = self.lineNum.int + self.posOffset.line.int

func column(self: LexerBase): int =
  ## Returns a column in the current line
  result = (self.idx + 1) - self.idxStart + self.posOffset.column.int

func peekPos*(self: LexerBase): FilePosition =
  ## Returns current character position.
  result = FilePosition(line: self.line().uint32, column: self.column().uint32)

func handleNewline*(self: var LexerBase): bool
  {.discardable.} =
  result =
    if self.peek() in Newlines:
      self.popChar('\r')
      self.popChar('\n')
      self.lineNum += 1
      self.idxStart = self.idx
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

template parseWhile*(self: var LexerBase; predicate: untyped): string =
  block:
    var result = newStringOfCap(16)
    var it {.inject.} = self.peek()

    while self.idx <= self.buffer.high and predicate:
      result &= self.pop()
      it = self.peek()

    result

template parseUntil*(self: var LexerBase; predicate: untyped): string =
  self.parseWhile(not(predicate))

{.pop.} # raises: []
