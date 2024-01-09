import std/strformat
import std/strutils
import std/options
import std/parseutils

import jet/token

import lib/line_info
import lib/utils


type Lexer* = object
    buffer  : openArray[char]
    pos     : int = 0           ## Position in the buffer
    linePos : int = 0           ## Position in the buffer of the line start character
    lineNum : int = 1
    curr*   : Token
    prev*   : Option[Token] = none(Token)

type LexResult = object
    token  : Option[Token] = none(Token)
    error  : Option[string] = none(string)
    parsed : uint = 1

const
    SPACE    = ' '  ## Just a whitespace char, nothing special
    EOF      = '\0' ## End of File
    CR       = '\r' ## Carriage Return
    LF       = '\n' ## Line Feed
    TAB      = '\t' ## Tabulation (horisontal)
    EOL      = {EOF} + Newlines  ## End of Line characters

func peek(self: Lexer): char =
    ## Peek current character from the `buffer`.
    result = self.buffer[self.pos]

func slice(self: Lexer): openArray[char] =
    result = self.buffer.toOpenArray(self.pos, self.buffer.high)

func line(self: Lexer): int =
    ## Returns line number.
    result = self.lineNum

func column(self: Lexer): int =
    ## Returns column number in the current line.
    result = (self.pos + 1) - self.linePos

func lineInfo(self: Lexer): LineInfo =
    result = LineInfo(line: self.line().uint32, column: self.column().uint32)

func getLine*(self: Lexer; lineNum: Positive): Option[string] =
    ## **Returns:** line at `line` in `buffer` (new line character excluded).
    result = none(string)
    var i = 1
    for line in ($self.buffer).splitLines():
        if i == lineNum:
            result = some(line)
            break
        i += 1

func getLines*(self: Lexer; lineNums: openArray[int] | Slice[int]): Option[seq[string]] =
    ## **Returns:** specified lines from the `buffer`.
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
    UnaryOpWhitelist* = {' ', ',', ';', '(', '[', '{'} + EOL

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

func parseWhile(s: openArray[char]; charSet: set[char]): string =
    result = ""
    discard s.parseWhile(result, charSet)

func parseWhile(s: openArray[char]; c: char): string =
    result = parseWhile(s, {c})

func lexSpace(buffer: openArray[char]): LexResult =
    if buffer[0] notin {SPACE, LF}:
        return

    let data = buffer.parseWhile(buffer[0])
    result.parsed = data.len().uint
    result.token = some(initToken(HSpace, data))

func lexId(buffer: openArray[char]): LexResult =
    if buffer[0] notin IdStartChars:
        return

    let data = buffer.parseWhile(IdChars)
    let kind = data.toTokenKind().get(Id)

    result.parsed = data.len().uint
    result.token =
        if kind == Id:
            some(initToken(Id, data))
        else:
            some(initToken(kind))

func lexNumber(buffer: openArray[char]): LexResult =
    if buffer[0] notin Digits:
        return

    let numPart = buffer.parseWhile(Digits)

    result.token =
        if numPart.len() <= buffer.high and buffer[numPart.len()] == '.':
            let fracPart = buffer
                .toOpenArray(numPart.len() + 1, buffer.high)
                .parseWhile(Digits)
            if fracPart.len() == 0:
                none(Token)
            else:
                result.parsed = numPart.len().uint + fracPart.len().uint + 1
                some(initToken(FloatLit, &"{numPart}.{fracPart}"))
        else:
            result.parsed = numPart.len().uint
            some(initToken(IntLit, numPart))

func lexComment(buffer: openArray[char]): LexResult =
    if buffer[0] != '#':
        return

    case buffer[1]
    of '#':
        unimplemented("documentation comment scan")
    of '!':
        unimplemented("module documentation comment scan")
    else:
        result.token = some(initToken(Empty))
        result.parsed = buffer.findIt(it == LF).uint

func nextToken(self: var Lexer) =
    if self.pos > self.buffer.high:
        self.curr = initToken(TokenKind.Eof, info = self.lineInfo())
        return

    let lexResult = case self.peek():
        of '#':
            self.slice().lexComment()
        of IdStartChars:
            self.slice().lexId()
        of Digits:
            self.slice().lexNumber()
        of SPACE, Newlines:
            self.slice().lexSpace()
        of EOF:
            LexResult(token: some(initToken(TokenKind.Eof, info = self.lineInfo())))
        else:
            LexResult(error: some(&"invalid character: '{strutils.escape($self.peek())}'"))

    if lexResult.error.isSome():
        panic(lexResult.error.get())

    if lexResult.token.isNone():
        todo()

    self.curr = lexResult.token.get()
    self.curr.info = self.lineInfo()
    self.curr.info.length = lexResult.parsed.uint32

    self.pos += lexResult.parsed.int

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
    result = Lexer(
        buffer: buffer.toOpenArray(0, buffer.high),
        curr: initToken(TokenKind.Eof),
        prev: none(Token),
    )
    result.nextToken()



when isMainModule:
    let file = """
foo 100 1.1 # comment
func 1234567890"""

    var lexer = newLexer(file)
    var tok = lexer.getToken()

    echo "---"
    echo file
    echo "---"

    while tok.isSome():
        echo tok.get().human()
        tok = lexer.getToken()
