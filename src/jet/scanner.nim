import std/strutils except Newlines
import std/sequtils
import std/tables

import pkg/questionable

import jet/scanner_base
import jet/token

import lib/utils
import lib/utils/line_info

export scanner_base


type Scanner* = object of ScannerBase
    token*     : Token
    prevToken* : Token

const
    BinChars*         = {'0'..'1'}
    OctChars*         = {'0'..'7'}
    HexChars*         = {'0'..'9', 'a'..'f', 'A'..'F'}
    IdChars*          = {'_'} + Letters + Digits
    IdStartChars*     = IdChars - Digits
    UnaryOpWhitelist* = {' ', ',', ';', '(', '[', '{'} + EOL # + '<' if it is a generic params

func atAnyComment(self: Scanner; pos = self.pos): bool =
    return self.at(pos) == '/' and self.at(pos + 1) == '/'

func atComment(self: Scanner; pos = self.pos): bool =
    return self.atAnyComment() and self.at(pos + 2) notin {'/', '!'}

func atCommentStmt(self: Scanner; pos = self.pos): bool =
    return self.atAnyComment() and self.at(pos + 2) in {'/', '!'}

proc skipComment(self: var Scanner) =
    ## Stoping at newline char
    assert(self.atComment())

    while self.ch() notin NewLines:
        inc(self.pos)

proc skipSpaces(self: var Scanner) =
    ## Skip spaces and set spacing\indentation for a token
    self.token.setSpacesBefore(0)

    while true:
        case self[0]
        of TAB:
            scanPanic("tabs are not allowed")
        of SPACE:
            inc(self.pos)
            self.token.setSpacesBefore(!self.token.spacesBefore() + 1)
        of Newlines:
            self.handleNewLine()
            self.token.setFirstInLine(true)
            let lastPos = self.pos

            while self[0] == SPACE:
                inc(self.pos)

            # space is the last special char in ASCII so
            # any char after it ord value is valid
            if self[0] > SPACE and not self.atComment():
                self.token.setIndent(self.pos - lastPos)
                break
            else:
                self.token.setSpacesBefore(self.pos - lastPos)
        elif self.atComment():
            self.skipComment()
        else:
            break

func getByIndex[T](self: set[T]; n: int): T
    {.raises: [ValueError].} =
    {.warning[ProveInit]: off.}
    if n notin 0 ..< self.card():
        raise newException(ValueError, "index not in set")

    var i = 0
    for item in self.items():
        if n == i: return item
        inc(i)
    unreachable()

proc scanComment(self: var Scanner) =
    ## Scans the doc comment statement (multiline)
    assert(self.atAnyComment())

    case self[2]
    of '/': self.token.kind = Comment
    of '!': self.token.kind = TopLevelComment
    else:
        # regular comment
        self.token.kind = Invalid
        self.skipComment()
        return

    inc(self.pos, 3)

    let spacesBefore   = self.token.spacesBefore() |? !self.token.indent()
    var firstLine      = true
    var skipFirstSpace = true
    var skipedSpaces   = 0

    while true:
        if skipFirstSpace:
            if self.ch() == ' ':
                inc(self.pos)
                inc(skipedSpaces)
            else:
                skipFirstSpace = false
        if firstLine:
            firstLine = false
        else:
            self.token.value.add('\n')

        while self.ch() notin NewLines:
            self.token.value.add(self.ch())
            inc(self.pos)
        self.handleNewLine()

        let lastPos = self.pos
        var indent = 0

        while self.ch() == ' ':
            inc(self.pos)
            inc(indent)

        if indent != spacesBefore:
            # another comment stmt
            self.pos = lastPos
            break

        if not self.atCommentStmt(self.pos): break

        case self[2]
        of '/': (if self.token.kind != Comment: break)
        of '!': (if self.token.kind != TopLevelComment: break)
        else: unreachable()
        inc(self.pos, 3)
    if not skipFirstSpace:
        # string reallocation is all you need to be happy
        var i = 0
        while skipedSpaces > 0:
            dec(skipedSpaces)
            self.token.value.insert(" ", i)
            i = self.token.value.find('\n', i) + 1

func scanMatchChars(
    self: var Scanner;
    buffer: var string;
    chars: set[char];
    minCharsCount: Natural = 0,
    handleUnderscore: static[bool] = false,
): bool
    {.inline, discardable.} =
    let oldPos = self.pos
    result = false

    while self.ch() in chars:
        self.eatChar(buffer)
        result = true

        when handleUnderscore:
            if self.ch() == '_':
                if self[1] in chars:
                    self.eatChar(buffer)
                else:
                    result = false
                    scanError("trailing or doubled '_' in number literals are invalid")

    if minCharsCount > 0 and (not result or (self.pos - oldPos) < minCharsCount):
        scanError("expected $1 or more characters, got $2" % [
            $minCharsCount,
            $(self.pos - oldPos)
        ])

func scanMatchChars(self: var Scanner; chars: set[char]; handleUnderscore: static[bool] = false): string =
    self.scanMatchChars(result, chars, handleUnderscore)

func scanIdOrKeyword(self: var Scanner) =
    assert(self.ch() in IdStartChars)

    self.scanMatchChars(self.token.value, IdChars)
    self.token.kind = keywordToTokenKind(self.token.value)
    # self.tok.id   = self.ids[symbol]

    if self.token.kind == Invalid:
        self.token.kind = Id

func scanEscapedChar(self: var Scanner) =
    assert(self.ch() == '\\')
    inc(self.pos) # skip \

    case self.ch()
    of '\'': self.eatChar(self.token.value, '\'')
    of '\"': self.eatChar(self.token.value, '\"')
    of '\\': self.eatChar(self.token.value, '\\')
    of '0': self.eatChar(self.token.value, '\0')
    of 'b': self.eatChar(self.token.value, '\b')
    of 'f': self.eatChar(self.token.value, '\f')
    of 'v': self.eatChar(self.token.value, '\v')
    of 't': self.eatChar(self.token.value, TAB)
    of 'n': self.eatChar(self.token.value, LF)
    of 'r': self.eatChar(self.token.value, CR)
    of 'x':
        unimplemented("\\x character escape")
    of 'u':
        if self.token.kind == CharLit:
            scanError("\\u is not allowed in character literals")

        unimplemented("\\u character escape")
    else: scanError("invalid character escape")

func scanChar(self: var Scanner) =
    assert(self.ch() == '\'')
    inc(self.pos) # skip '

    self.token.kind = CharLit

    case self.ch()
    of '\0'..pred(' '): scanError("invalid character literal")
    of '\'': scanError("empty character literal")
    of '\\': self.scanEscapedChar()
    else: self.eatChar(self.token.value)

    if self.ch() == '\'':
        inc(self.pos)
    else:
        scanError("missing closing \'")

func performMultilineStringLitIndentation(self: var Scanner; lineIndices: openArray[int]) =
    assert(self[0] == '\"' and self[1] == '\"' and self[2] == '\"')
    assert(lineIndices.len() > 0)

    var indentStr = self.buffer[self.lineStart ..< self.pos]

    if not indentStr.isEmptyOrWhitespace():
        #|  foo""" or something like that
        return

    unimplemented()

    # var newlinePos = 0
    # var lines      = self.token.value.splitLines()

    # while newlinePos != -1:
    #     block inner:
    #         for i in 0 ..< indentStr.len():
    #             if self.token.value[newlinePos + i] != indentStr[i]:
    #                 break inner

    #         self.token.value[newlinePos .. indentStr.len()] = ""
    #         newlinePos = self.token.value.find('\n', newlinePos) + 1

func scanString(self: var Scanner; raw = false) =
    assert(self.ch() == '\"')
    inc(self.pos)

    let multiline   = self[0] == '\"' and self[1] == '\"'
    var lineIndices = newSeq[int]()

    self.token.kind =
        if multiline:
            inc(self.pos, 2) # skip ""

            # skip trailing spaces\tabs
            var pos = self.pos
            while self.ch() in {SPACE, TAB}:
                inc(pos)
            if self.at(pos) in NewLines:
                self.pos = pos
                self.handleNewLine()

            if raw: LongRawStringLit
            else: LongStringLit
        else:
            if raw: RawStringLit
            else: StringLit

    while true:
        case self.ch()
        of '\"':
            if multiline:
                if self[1] == '\"' and self[2] == '\"' and self[3] != '\"':
                    # compute indentation and remove it from string value
                    if lineIndices.len() > 0:
                        self.performMultilineStringLitIndentation(lineIndices)
                    inc(self.pos, 3)
                    break
                self.eatChar(self.token.value)
            else:
                inc(self.pos)
                break
        of '\\':
            if raw:
                self.eatChar(self.token.value)
            else:
                self.scanEscapedChar()
        of EOF:
            if multiline:
                scanError("expected \"\"\" but EOF reached")
            else:
                scanError("expected \" but EOF reached")
            break
        of NewLines:
            if multiline:
                lineIndices.add(self.line())
                self.handleNewLine()
                self.token.value.add('\n')
            else:
                scanError("missing closing \"")
                break
        else:
            self.eatChar(self.token.value)

func scanOperator(self: var Scanner) =
    assert(self.ch() in OperatorCharSet)

    var foundKind = Invalid
    var pos       = self.pos

    block loop:
        for (kind, op) in Operators.pairs():
            var i = 0 # 'op' index
            pos   = self.pos

            while self.at(pos) == op[i]:
                inc(pos)

                # check it was the last char in this operator
                if i == op.high:
                    # check if the next char in 'OperatorChars'
                    # if not, this is a needed operator
                    if self.at(pos) notin OperatorCharSet:
                        foundKind = OperatorKinds.getByIndex(kind)
                        break loop
                    else:
                        # longer than this operator
                        break
                inc(i)
                # check char can be operator
                if self.at(pos) notin OperatorCharSet:
                    break
            # it is another operator, skip

    if foundKind == Invalid:
        scanError("unknown operator '$1'" % self.buffer[self.pos ..< pos])

    self.token.kind = foundKind
    self.pos = pos

func scanNumber(self: var Scanner) =
    assert(self.ch() in Digits + {'-'})

    self.token.kind  = IntLit
    self.token.value = newStringOfCap(64) # IDK why 64

    block beforeSuffix:
        if self.ch() == '-':
            self.eatChar(self.token.value)
        if self.ch() == '0':
            self.eatChar(self.token.value)
            var chars: set[char] = {}

            case self.ch()
            of 'b'           : chars = BinChars
            of 'x'           : chars = HexChars
            of 'o'           : chars = OctChars
            of 'B', 'X', 'O' : scanError("using uppercase letters for number literals are not allowed. Use lowercase letters instead")
            of Digits        : scanError("'0' can't be a first digit in literals if there is no '.' after it")
            else             : break beforeSuffix

            if chars != {}:
                self.eatChar(self.token.value)
                self.scanMatchChars(self.token.value, chars, minCharsCount=1, handleUnderscore=true)
        else:
            self.scanMatchChars(self.token.value, Digits, minCharsCount=1, handleUnderscore=true)

            if self.ch() == '.':
                self.token.kind = FloatLit
                self.eatChar(self.token.value)
                self.scanMatchChars(self.token.value, Digits, minCharsCount=1, handleUnderscore=true)
            if self.ch() in {'e', 'E'}:
                self.token.kind = FloatLit
                self.eatChar(self.token.value, 'e')

                if self.ch() in {'+', '-'}:
                    self.eatChar(self.token.value)

                self.scanMatchChars(self.token.value, Digits, minCharsCount=1)

    # scan literal suffix
    if self.ch() in Letters:
        var suffix = newStringOfCap(3)
        self.scanMatchChars(suffix, IdChars - {'_'}, minCharsCount=1)

        case self.token.kind
        of FloatLit:
            case suffix.toLowerAscii()
            of "f32" : self.token.kind = F32Lit
            of "f64" : self.token.kind = F64Lit
            of "f"   : scanError("redundant float literal suffix")
            else     : scanError("invalid suffix '$1' for float literal" % suffix)
        of IntLit:
            case suffix.toLowerAscii()
            of "f32" : self.token.kind = F32Lit
            of "f64" : self.token.kind = F64Lit
            of "f"   : self.token.kind = FloatLit
            of "i8"  : self.token.kind = I8Lit
            of "i16" : self.token.kind = I16Lit
            of "i32" : self.token.kind = I32Lit
            of "i64" : self.token.kind = I64Lit
            of "u8"  : self.token.kind = U8Lit
            of "u16" : self.token.kind = U16Lit
            of "u32" : self.token.kind = U32Lit
            of "u64" : self.token.kind = U64Lit
            of "u"   : self.token.kind = USizeLit
            of "i"   : self.token.kind = ISizeLit
            else: scanError("invalid suffix '$1' for integer literal" % suffix)
        else:
            unreachable()

        if self.token.kind in {UIntLit, USizeLit, U8Lit..U64Lit} and
           self.token.value.startsWith('-'):
            scanError("unsigned integer can't be negative" % suffix)

proc nextToken*(self: var Scanner) =
    ## Get next token and store it in the `tok` field.
    self.token = newToken(Invalid)

    if self.pos >= self.buffer.len():
        scanpanic("EOF reached, no tokens to scan (bufferPos: '$1', max: '$2')" % [
            $self.pos,
            $self.buffer.len()
        ])

    let posBeforeSkip = self.pos
    self.skipSpaces()

    if posBeforeSkip == 0 and spacesBefore =? self.token.spacesBefore():
        # first token
        self.token.setIndent(spacesBefore)

    self.token.scannerPos = self.pos
    self.token.info       = self.getLineInfo()

    if self.token.kind != Invalid:
        unreachable()

    case self.ch()
    of EOF:
        self.token.kind = Last
    of Newlines, ' ':
        unreachable()
    of '`':
        unimplemented("scanAccentId")
    of '#':
        self.token.kind = Hashtag
        inc(self.pos)
    of ',', ';', '(', ')', '{', '}', '[', ']':
        self.token.kind = punctuationToTokenKind(self.ch())
        inc(self.pos)
    of ':':
        if self[1] == ':':
            self.token.kind = ColonColon
            inc(self.pos, 2)
        else:
            self.token.kind = Colon
            inc(self.pos)
    of '-', '+':
        if self[1] in Digits and
          (self[-1] in UnaryOpWhitelist + {'<'}):
            if self[0] == '+':
                scanError("unary + is illegal")
                inc(self.pos)
            self.scanNumber()
        else:
            self.scanOperator()
    of '|':
        self.token.kind = Bar
        inc(self.pos)
    of '.':
        if self[1] == '.':
            if self[2] == '<':
                self.token.kind = DotDotLess
                inc(self.pos, 3)
            elif self[2] == '.':
                self.token.kind = DotDotDot
                inc(self.pos, 3)
            else:
                self.token.kind = DotDot
                inc(self.pos, 2)
        else:
            self.token.kind = Dot
            inc(self.pos)
    of '/':
        if self[1] == '/':
            if self[2] in {'/', '!'}:
                self.scanComment()
            else:
                unreachable("comment are not skiped by 'skipSpaces', fix pls")
        else:
            self.scanOperator()
    of '_':
        if self[1] in IdChars:
            self.scanIdOrKeyword()
        else:
            self.token.kind = Underscore
            inc(self.pos)
    of '\"':
        self.scanString()
    of '\'':
        self.scanChar()
    of OperatorStartCharSet - {'/', '_', '.', '-', '+', ':'}:
        self.scanOperator()
    of Digits:
        self.scanNumber()
    of Letters:
        if self[0] in {'r', 'R'} and
           self[1] == '\"':
            inc(self.pos)
            self.scanString(raw=true)
        else:
            self.scanIdOrKeyword()
    else:
        self.token.value  = $self.ch()
        self.token.kind = Invalid

        scanError("invalid token '$1'" % self.token.value)
        inc(self.pos)

    self.token.info.length = uint32(self.pos - self.token.scannerPos)

proc getToken*(self: var Scanner): ?Token
    {.discardable.} =
    if self.token.kind == Invalid:
        if self.token.value.len() > 0:
            scanError("got invalid token, data: '$1'" % self.token.value)
        else:
            scanError("got invalid token")
        return none(Token)

    self.prevToken = self.token

    if self.token.kind != Last:
        self.nextToken()

    if self.token.isFirstInLine():
        self.prevToken.setLastInLine(true)
    else:
        self.prevToken.setSpacesAfter(!self.token.spacesBefore())

    return some(self.prevToken)

proc getAllTokens*(self: var Scanner): seq[Token] =
    result = @[]

    while token =? self.getToken():
        result.add(token)

proc openScanner*(content: sink string): Scanner =
    {.warning[Uninit]: off.}
    result = Scanner(token: newToken(Invalid), prevToken: newToken(Invalid))

    result.open(content)
    result.nextToken()

proc openScannerFile*(fileName: string): Scanner =
    result = Scanner(token: newToken(Invalid), prevToken: newToken(Invalid))

    result.openFile(fileName)
    result.nextToken()
