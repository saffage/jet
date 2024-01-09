import std/options


type TokenKind* = enum
    #[ Special ]#
    Invalid         = "<invalid>"
    Last            = "<end-of-file>"
    Id              = "<identifier>"
    Comment         = "<comment>"             ## Documentation comment
    TopLevelComment = "<top-level-comment>"   ## Module documentation comment

    #[ Literals ]#
    IntLit           = "<int-literal>"                ## 10
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
    KwNil     = "nil"
    KwVar     = "var"
    KwVal     = "val"
    KwFunc    = "func"
    KwType    = "type"
    KwStruct  = "struct"
    KwEnum    = "enum"
    KwIf      = "if"
    KwElif    = "elif"
    KwElse    = "else"
    KwMatch   = "match"
    KwReturn  = "return"
    KwWhile   = "while"
    KwFor     = "for"
    KwLoop    = "loop"
    KwDo      = "do"
    KwOf      = "of"
    KwXdd     = "xdd" # FUS RO DHA

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
    Shl        = "<<"
    Shr        = ">>"
    Assign     = "="
    DotDot     = ".."
    DotDotLess = "..<"

    #[ Special symbols ]#
    Bar        = "|"
    ColonColon = "::"
    Underscore = "_"
    At         = "@"
    Dollar     = "$"    ## String interpolation
    DotDotDot  = "..."

    MatchCaseArrow  = "=>"
    ExclamationMark = "!"
    QuestionMark    = "?"
    Ampersand       = "&"

const
    UntypedLiteralKinds* = {IntLit, FloatLit}
    TypedLiteralKinds*   = {ISizeLit .. F64Lit}
    StringLiteralKinds*  = {StringLit .. LongRawStringLit}
    AnyLiteralKinds*     = {CharLit} + UntypedLiteralKinds + TypedLiteralKinds + StringLiteralKinds

func isKeyword*(kind: TokenKind): bool =
    const kinds = {KwTrue .. KwAnd}
    result = kind in kinds

func isOperator*(kind: TokenKind): bool =
    ## Note: word-like operators are not checked, use `isWordLikeOperator` instead.
    const kinds = {EqOp .. DotDotLess}
    result = kind in kinds

func isWordLikeOperator*(kind: TokenKind): bool =
    const kinds = {KwAnd, KwOr, KwOf}
    result = kind in kinds

func isLiteral*(kind: TokenKind): bool =
    const kinds = {IntLit .. LongRawStringLit, ISizeLit .. F64Lit}
    result = kind in kinds

func fromString*(s: string): Option[TokenKind] =
    const stringableKinds = {LParen .. Ampersand}
    result = none(TokenKind)

    for kind in stringableKinds:
        if $kind == s:
            result = some(kind)
            break
