import std/strformat

import pkg/questionable

import jet/token_kind

import utils
import utils/line_info

export token_kind


type Token* = object
    value* : string     ## String representation
    kind*  : TokenKind  ## Kind of the token
    info*  : LineInfo   ## Line information about token

    scannerPos*     : int = -1      ## A position in the **Scanner**'s buffer
    spacesBefore    : int = -1      ## Used with spacing (used as indentation if `leadingSpacing`)
    spacesAfter     : int = -1      ## Used with spacing (not used if `trailingSpacing`)
    leadingSpacing  : bool = false  ## false -> `<other> this`, true  -> ` this`
    trailingSpacing : bool = false  ## false -> `this <other>`, true  -> `this `

type Notation* = enum
    Unknown     ## When `token.firstAndLast()`
    Infix       ## `a ! ...` or `a!...` or `a !` or `! a`
    Prefix      ## ` !a` or `(!a` or `,!a` or `;!a`
    Postfix     ## `a! ` or `a!)` or `a!,` or `a!;`

func isFirstInLine*(self: Token): bool =
    ## **Returns:** `true` if it's the first token in the line
    return self.leadingSpacing

func isLastInLine*(self: Token): bool =
    ## **Returns:** `true` if it's the last token in the line
    return self.trailingSpacing

func isFirstAndLast*(self: Token): bool =
    ## **Returns:** `true` if it's both first and last token in the line
    return self.leadingSpacing and self.trailingSpacing

func isFirstOrLast*(self: Token): bool =
    ## **Returns:** `true` if have any spacing around this token
    return self.leadingSpacing or self.trailingSpacing

func spacesBefore*(self: Token): ?int =
    ## **Returns:** spaces before token, if this token not first in line
    if not self.isFirstInLine():
        assert(self.spacesBefore != -1, "spacesBefore can't be -1")
        return some(self.spacesBefore)
    else:
        return none(int)

func spacesAfter*(self: Token): ?int =
    ## **Returns:** spaces after token, if this token not last in line
    if not self.isLastInLine():
        assert(self.spacesAfter != -1, "spacesAfter can't be -1")
        return some(self.spacesAfter)
    else:
        return none(int)

func indent*(self: Token): ?int =
    ## **Returns:** indentation, if this token first in line
    if self.isFirstInLine():
        assert(self.spacesBefore != -1, "spacesBefore can't be -1")
        return some(self.spacesBefore)
    else:
        return none(int)

func setFirstInLine*(self: var Token; value: bool) =
    ## Sets that it's the first token in the line
    ##
    ## Also sets `spacesBefore` to `-1` to prevent using an invalid data
    if value and not self.leadingSpacing:
        self.leadingSpacing = true
        self.spacesBefore   = -1
    elif self.leadingSpacing:
        self.leadingSpacing = false
        self.spacesBefore   = -1

func setLastInLine*(self: var Token; value: bool) =
    ## Sets that it's the last token in the line
    ##
    ## Also sets `spacesAfter` to `-1` to prevent using an invalid data
    if value and not self.trailingSpacing:
        self.trailingSpacing = true
        self.spacesAfter     = -1
    elif self.trailingSpacing:
        self.trailingSpacing = false
        self.spacesAfter     = -1

func setFirstAndLast*(self: var Token; value: bool) =
    ## Sets that it's the last token in the line
    if value:
        self.leadingSpacing  = true
        self.trailingSpacing = true
    else:
        self.leadingSpacing  = false
        self.trailingSpacing = false

func setSpacesBefore*(self: var Token; value: int) =
    ## Sets spaces before token and off **Leading** spacing
    if self.isFirstInLine(): self.setFirstInLine(false)
    self.spacesBefore = value

func setSpacesAfter*(self: var Token; value: int) =
    ## Sets spaces after token and off **Trailing** spacing
    if self.isLastInLine(): self.setLastInLine(false)
    self.spacesAfter = value

func setIndent*(self: var Token; value: Natural) =
    ## Sets indentation value and **Leading** spacing
    if not self.isFirstInLine(): self.setFirstInLine(true)
    self.spacesBefore = value

template `firstInLine=`*(self: var Token; value: bool) =
    self.setFirstInLine(value)

template `lastInLine=`*(self: var Token; value: bool) =
    self.setLastInLine(value)

template `firstAndLast=`*(self: var Token; value: bool) =
    self.setFirstAndLast(value)

template `spacesBefore=`*(self: var Token; value: int) =
    self.setSpacesBefore(value)

template `spacesAfter=`*(self: var Token; value: int) =
    self.setSpacesAfter(value)

template `indent=`*(self: var Token; value: Natural) =
    self.setIndent(value)

func notation*(self: Token; prev = Invalid; next = Invalid): Notation =
    let spacesBefore = self.spacesBefore() |? -2
    let spacesAfter  = self.spacesAfter() |? -2
    let firstInLine  = self.isFirstInLine()
    let lastInLine   = self.isLastInLine()
    var prefix       = false
    var postfix      = false

    const PrefixWhitelist  = {Invalid, Comma, Semicolon, LParen, LBracket, LBrace}
    const PostfixWhitelist = {Invalid, Comma, Semicolon, RParen, RBracket, RBrace, Last} # idk are 'Last' is needed

    if firstInLine and lastInLine:
        return Unknown

    if not lastInLine and spacesAfter == 0 and (spacesBefore != 0 or firstInLine):
        prefix = prev notin PrefixWhitelist

    if not firstInLine and spacesBefore == 0 and (spacesAfter != 0 or lastInLine):
        postfix = next notin PostfixWhitelist

    if prefix and postfix:
        # maybe somethis like `(!)`
        return Unknown

    if prefix:
        return Prefix

    if postfix:
        return Postfix

    return Infix

func isPrefix*(self: Token): bool =
    return self.notation() == Prefix

func isPostfix*(self: Token): bool =
    return self.notation() == Postfix

func isInfix*(self: Token): bool =
    return self.notation() == Infix

func `$`*(self: Token): string =
    return self.kind.name()

func human*(self: Token): string
    {.raises: [ValueError].} =
    ## **Returns:** a human readable string about a token
    result = fmt"Token at {self.info} = {self}"

    if self.value.len() > 0:
        result.addQuoted(self.value)

func newToken*(kind: TokenKind; info = LineInfo(); value = ""): Token
    {.raises: [].} =
    result = Token(kind: kind, info: info, value: value)
