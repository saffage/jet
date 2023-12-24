from std/sugar import collect

import std/os
import std/streams
import std/strutils
import std/strformat

import jet/token

import utils
import utils/line_info


type ScannerBase* = object of RootObj
    buffer*      : string       ## Content of the file to be scanned
    pos*         : int = 0      ## Position in the buffer
    lineStart*   : int = 0      ## Position of line start in the buffer
    linePos      : int = 1      ## Current line number
    lineIndices  : seq[int]     ## Indices of newline chars for every line

const
    EOF*      = '\0'                    ## End of File
    CR*       = '\r'                    ## Carriage Return
    LF*       = '\n'                    ## Line Feed
    TAB*      = '\t'                    ## Tabulation (horisontal)
    SPACE*    = ' '                     ## Just a whitespace char, nothing special
    NewLines* = {CR, LF}                ## New line char
    EOL*      = NewLines + {EOF}        ## End of Line
    Spaces*   = NewLines + {TAB, ' '}   ## Any char that can be treated as whitespace

func handleNewLine*(self: var ScannerBase) =
    ## Call this when **CR** or **LF** is reached.
    ## The `pos` field must be at a position of this char.
    ##
    ## **Returns:** position of the next symbol to be scanned.
    assert(self.buffer[self.pos] in NewLines)

    let wasCR      = (self.buffer[self.pos] == CR)
    self.pos      += (1 + wasCR.ord)
    self.linePos  += 1
    self.lineStart = self.pos

func skipLine*(self: var ScannerBase) =
    ## Skip all the line until a new line.
    while self.buffer[self.pos] notin EOL:
        inc(self.pos)

    if self.buffer[self.pos] != EOF:
        self.handleNewLine()

template line*(self: ScannerBase): int =
    ## **Returns:** line number.
    self.linePos

template column*(self: ScannerBase): int =
    ## **Returns:** column number in the current line.
    self.pos - self.lineStart

template ch*(self: ScannerBase): char =
    ## **Returns:** current character in the `buffer`.
    self.buffer[self.pos]

template at*(self: ScannerBase; pos: int): char =
    ## **Returns:** character in the `buffer` at specified `pos`.
    self.buffer[pos]

template `[]`*(self: ScannerBase; offset: int): char =
    ## **Returns:** current character in the `buffer` with specified offset.
    self.buffer[self.pos + offset]

template eatChar*(self: var ScannerBase): char =
    ## Get character and increment `pos`.
    inc(self.pos)
    self.buffer[self.pos.pred]

template eatChar*(self: var ScannerBase; replacement: char): char =
    ## Get a replacement character and increment `pos`.
    inc(self.pos)
    replacement

template eatChar*(self: var ScannerBase; str: var string) =
    ## Add character to the `str` and increment `pos` field.
    str.add(self.buffer[self.pos])
    inc(self.pos)

template eatChar*(self: var ScannerBase; str: var string; replacement: char) =
    ## Add replacement character to the `str` and increment `pos` field.
    str.add(replacement)
    inc(self.pos)

func getLineInfo*(self: ScannerBase): LineInfo =
    result = LineInfo(line: self.line().uint32, column: self.column().uint32)

func getLine*(self: ScannerBase; line: Positive): string =
    ## `line` must be in range `1 .. <last-buffer-line-num>`.
    ##
    ## **Returns:** line at `line` in `buffer` (new line character excluded).
    if line > self.lineIndices.len():
        raise newException(ValueError, fmt"line '{line}' does not exists (max: '{self.lineIndices.len()}')")

    let lineStart = self.lineIndices[line - 1] + 1
    var i = lineStart

    while self.buffer[i] notin EOL:
        inc(i)

    result = self.buffer[lineStart ..< i]

func getLines*(self: ScannerBase; lines: openArray[int] | Slice): seq[string] =
    ## **Returns:** specified lines from the `buffer`.
    result = @[]

    for line in lines:
        result.add(self.getLine(line))

func errorAt*(self: ScannerBase; info: LineInfo; inversed=false; fill=false; oneLine=true): string =
    result = ""

    let line            = info.line.int
    let firstLineNum    = max(line - 3, 1)
    let lineMaxNumChars = numLen(max(line + int(not oneLine), 0))
    let lineNums =
        if oneLine: line .. line
        else: firstLineNum .. (line + 1)

    for lineNum in lineNums:
        if lineNum != firstLineNum: result &= '\n'

        result &= fmt" {align($lineNum, lineMaxNumChars)} |"
        result &= self.getLine(lineNum)

        if lineNum == line:
            result &= fmt("\n {spaces(numLen(line))} |")
            result &= spaces(info.column)

            if fill:
                result &= repeat('^', max(info.length, 1))
            elif inversed:
                if info.length > 0: result &= repeat('~', info.length - 1)
                result &= '^'
            else:
                result &= '^'
                if info.length > 0: result &= repeat('~', info.length - 1)

template errorAt*(self: ScannerBase; token: Token; inversed=false; fill=false; oneLine=true): string =
    self.errorAt(token.info, inversed, fill, oneLine)

template scanInfo*(msg: string) =
    ## Must be called in context where `self` is of type **ScannerBase**.
    info(msg, self.getLineInfo())

template scanHint*(msg: string) =
    ## Must be called in context where `self` is of type **ScannerBase**.
    hint(msg, self.getLineInfo())

template scanWarn*(msg: string) =
    ## Must be called in context where `self` is of type **ScannerBase**.
    warn(msg, self.getLineInfo())

template scanError*(msg: string) =
    ## Must be called in context where `self` is of type **ScannerBase**.
    error(msg, self.getLineInfo())

template scanPanic*(msg: string) =
    ## Must be called in context where `self` is of type **ScannerBase**.
    panic(msg, self.getLineInfo())

func findAll(str: string; c: char): seq[int] =
    result = @[-1] # TODO: remove '-1'

    result.add collect do:
        for i, ch in str:
            if ch == c: i

func openScannerBase*(content: sink string): ScannerBase =
    ## Opens the **ScannerBase** with specified `content`.
    # manualy insert an additional '\0' to be able to index at it
    # WARNING: string can be reallocated
    let lineIndices = content.findAll(LF)
    content.setLen(content.len() + 1)
    content[content.high] = '\0'

    result = ScannerBase(buffer: ensureMove(content), lineIndices: lineIndices)

proc openScannerBaseFile*(file: File): ScannerBase =
    ## Read whole file into the `buffer` field in the resulting **ScannerBase**.
    assert(file != nil)

    var fstream = newFileStream(file)
    result = openScannerBase(fstream.readAll())

proc openScannerBaseFile*(fileName: string): ScannerBase =
    ## Open file and read all data from it.
    if not fileExists(fileName):
        panic(fmt"can't open file '{fileName}' for 'ScannerBase'")

    var file = open(fileName, fmRead)
    result = openScannerBaseFile(file)

func open*(self: var ScannerBase; content: string) =
    self = openScannerBase(content)

proc openFile*(self: var ScannerBase; file: File) =
    self = openScannerBaseFile(file)

proc openFile*(self: var ScannerBase; fileName: string) =
    self = openScannerBaseFile(fileName)
