import std/enumutils


type TokenKind* = enum
    #[ Special ]#
    Invalid         = "<invalid>"
    Last            = "<end-of-file>"
    Id              = "<identifier>"
    Comment         = "<comment>"             ## Documentation comment
    TopLevelComment = "<top-level-comment>"   ## Module documentation comment

    #[ Literals ]#
    IntLit           = "<int-literal>"                ## 10 or 10i
    UIntLit          = "<uint-literal>"               ## 10u
    FloatLit         = "<float-literal>"              ## 1.0 or 1e2 etc.
    CharLit          = "<char-literal>"               ## quoted in '
    StringLit        = "<string-literal>"             ## quoted in "
    RawStringLit     = "<raw-string-literal>"         ## quoted in ", prefixed with 'R'
    LongStringLit    = "<long-string-literal>"        ## quoted in """
    LongRawStringLit = "<long-raw-string-literal>"    ## quoted in """, prefixed with 'R'

    #[ Typed Literals ]#
    ISizeLit = "<isize-literal>"
    USizeLit = "<iusize-literal>"
    I8Lit    = "<i8-literal>"
    I16Lit   = "<i16-literal>"
    I32Lit   = "<i32-literal>"
    I64Lit   = "<i64-literal>"
    U8Lit    = "<u8-literal>"
    U16Lit   = "<u16-literal>"
    U32Lit   = "<u32-literal>"
    U64Lit   = "<u64-literal>"
    F32Lit   = "<f32-literal>"
    F64Lit   = "<f64-literal>"

    #[ Brackets\Parentheses ]#
    LParen   = "("
    RParen   = ")"
    LBrace   = "{"
    RBrace   = "}"
    LBracket = "["
    RBracket = "]"

    #[ Punctuation ]#
    Dot       = "."
    Comma     = ","
    Colon     = ":"
    Semicolon = ";"

    #[ Keywords ]#
    KwTrue    = "true"
    KwFalse   = "false"
    KwNull    = "null"
    KwLet     = "let"
    KwMut     = "mut"
    KwVal     = "val"
    KwDef     = "def"
    KwTypeDef = "typedef"
    KwIf      = "if"
    KwElif    = "elif"
    KwElse    = "else"
    KwReturn  = "return"
    KwWhile   = "while"
    KwFor     = "for"
    KwLoop    = "loop"
    KwDo      = "do"

    #[ Operators-Keywords ]#
    KwOr      = "or"
    KwAnd     = "and"

    #[ Comparison Operators ]#
    EqOp = "=="
    NeOp = "!="
    LtOp = "<"
    GtOp = ">"
    LeOp = "<="
    GeOp = ">="

    #[ Other Operators ]#
    Plus       = "+"
    Minus      = "-"
    Asterisk   = "*"
    Slash      = "/"
    Percent    = "%"
    PlusPlus   = "++"
    Eq         = "="
    DotDot     = ".."
    DotDotLess = "..<"

    #[ Special symbols ]#
    ColonColon = "::"
    Underscore = "_"
    Hashtag    = "#"
    Dollar     = "$"    ## String interpolation
    DotDotDot  = "..."

const
    TypedLiteralKinds* = {ISizeLit..F64Lit}
    LiteralKinds*      = {IntLit..LongRawStringLit}
    AnyLiteralKinds*   = TypedLiteralKinds + LiteralKinds

    CommentKinds*     = {Comment, TopLevelComment}
    PunctuationKinds* = {LParen..RBracket, Dot..Semicolon}
    KeywordKinds*     = {KwTrue..KwAnd}
    OperatorKinds*    = {EqOp..DotDotLess}

func keywordToTokenKind*(keyword: string): TokenKind
    {.raises: [].} =
    ## **Returns:**
    ##  - `Invalid` if `keyword` is not a keyword.
    for kind in KeywordKinds:
        if keyword == $kind:
            return kind

    return Invalid

func operatorToTokenKind*(operator: string): TokenKind
    {.raises: [].} =
    ## **Returns:**
    ##  - `Invalid` if `operator` is not an operator.
    for kind in OperatorKinds:
        if operator == $kind:
            return kind

    return Invalid

func punctuationToTokenKind*(c: char): TokenKind =
    ## **Returns:**
    ##  - `Invalid` if `c` is not an punctuation character
    for kind in PunctuationKinds:
        if $c == $kind:
            return kind

    return Invalid

func name*(kind: TokenKind): string =
    return kind.symbolName

func buildOperatorsArray(): array[OperatorKinds.len(), string]
    {.compileTime.} =
    {.warning[ProveInit]: off.}
    var i = 0

    for kind in OperatorKinds:
        result[i] = $kind
        inc(i)

func buildOperatorCharsSet(): set[char]
    {.compileTime.} =
    result = {}

    for kind in OperatorKinds:
        for c in $kind:
            result.incl(c)

func buildOperatorStartCharsSet(): set[char] =
    result = {}

    for kind in OperatorKinds:
        result.incl(($kind)[0])

const
    Operators*            = buildOperatorsArray()
    OperatorCharSet*      = buildOperatorCharsSet()
    OperatorStartCharSet* = buildOperatorStartCharsSet()
