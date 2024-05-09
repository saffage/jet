package token

import "fmt"

//go:generate stringer -type=Kind
//go:generate stringer -type=Kind -output=kind_user_string.go -linecomment

// NOTE after generating the 'kind_user_string.go' file, the method name `String`
// must be changed to `UserString`, because when using the 'stringer' tool you
// cannot specify a name for the output method, and because of name conflict.

type Kind byte

const (
	Illegal Kind = iota // illegal character

	EOF        // end of file
	Comment    // comment
	Whitespace // whitespace
	Tab        // horizontal tabulation
	NewLine    // new line

	Ident  // identifier
	Int    // untyped int
	Float  // untyped float
	String // untyped string

	LParen    // '('
	RParen    // ')'
	LCurly    // '{'
	RCurly    // '}'
	LBracket  // '['
	RBracket  // ']'
	Comma     // ','
	Colon     // ':'
	Semicolon // ';'

	// The following operators are position dependent
	// in the current implementation of the scanner.

	Eq         // operator '='
	EqOp       // operator '=='
	Bang       // operator '!'
	NeOp       // operator '!='
	LtOp       // operator '<'
	LeOp       // operator '<='
	GtOp       // operator '>'
	GeOp       // operator '>='
	Plus       // operator '+'
	PlusEq     // operator '+='
	Minus      // operator '-'
	MinusEq    // operator '-='
	Asterisk   // operator '*'
	AsteriskEq // operator '*='
	Slash      // operator '/'
	SlashEq    // operator '/='
	Percent    // operator '%'
	PercentEq  // operator '%='

	// End of position dependent tokens.

	Amp          // operator '&'
	Pipe         // operator '|'
	Caret        // operator '^'
	At           // operator '@'
	QuestionMark // operator '?'
	Arrow        // operator '->'
	FatArrow     // operator '=>'
	Shl          // operator '<<'
	Shr          // operator '>>'
	Dot          // operator '.'
	Dot2         // operator '..'
	Dot2Less     // operator '..<'
	Ellipsis     // operator '...'

	// NOTE some keywords are unused.

	KwAnd      // keyword 'and'
	KwOr       // keyword 'or'
	KwModule   // keyword 'module'
	KwAlias    // keyword 'alias'
	KwStruct   // keyword 'struct'
	KwEnum     // keyword 'enum'
	KwFunc     // keyword 'func'
	KwVal      // keyword 'val'
	KwVar      // keyword 'var'
	KwConst    // keyword 'const'
	KwOf       // keyword 'of'
	KwIf       // keyword 'if'
	KwElse     // keyword 'else'
	KwWhile    // keyword 'while'
	KwReturn   // keyword 'return'
	KwBreak    // keyword 'break'
	KwContinue // keyword 'continue'
)

const (
	_special_begin = EOF
	_special_end   = NewLine

	_primary_begin = Ident
	_primary_end   = String

	_punctuation_begin = LParen
	_punctuation_end   = Semicolon

	_operator_begin = Eq
	_operator_end   = Ellipsis

	_keywords_begin = KwAnd
	_keywords_end   = KwContinue

	_kinds_last = _keywords_end
)

// Returns `Illegal` if `s` is not a kind name.
func KindFromString(s string) Kind {
	for kind, str := range representableKinds {
		if str == s {
			return kind
		}
	}

	return Illegal
}

func (kind Kind) IsSpecial() bool {
	return _special_begin <= kind && kind <= _special_end
}

func (kind Kind) IsPrimary() bool {
	return _primary_begin <= kind && kind <= _primary_end
}

func (kind Kind) IsPunctuation() bool {
	return _punctuation_begin <= kind && kind <= _punctuation_end
}

func (kind Kind) IsOperator() bool {
	return _operator_begin <= kind && kind <= _operator_end
}

func (kind Kind) IsKeyword() bool {
	return _keywords_begin <= kind && kind <= _keywords_end
}

func (kind Kind) Repr() string {
	s, ok := representableKinds[kind]

	if !ok {
		panic(fmt.Sprintf("%s cannot be represented as string (not enough data)", kind))
	}

	return s
}

func getKinds(begin, end int) []Kind {
	kinds := make([]Kind, 0, end-begin+1)

	for kind := begin; kind <= end; kind++ {
		kinds = append(kinds, Kind(kind))
	}

	return kinds
}

func SpecialKinds() []Kind {
	return getKinds(int(_special_begin), int(_special_end))
}

func PrimaryKinds() []Kind {
	return getKinds(int(_primary_begin), int(_primary_end))
}

func PunctuationKinds() []Kind {
	return getKinds(int(_punctuation_begin), int(_punctuation_end))
}

func OperatorKinds() []Kind {
	return getKinds(int(_operator_begin), int(_operator_end))
}

func KeywordKinds() []Kind {
	return getKinds(int(_keywords_begin), int(_keywords_end))
}

func AllKinds() []Kind {
	return getKinds(0, int(_kinds_last))
}

var representableKinds = map[Kind]string{
	LParen:       "(",
	RParen:       ")",
	LCurly:       "{",
	RCurly:       "}",
	LBracket:     "[",
	RBracket:     "]",
	Dot:          ".",
	Comma:        ",",
	Colon:        ":",
	Semicolon:    ";",
	Eq:           "=",
	Bang:         "!",
	QuestionMark: "?",
	EqOp:         "==",
	NeOp:         "!=",
	LtOp:         "<",
	GtOp:         ">",
	LeOp:         "<=",
	GeOp:         ">=",
	Arrow:        "->",
	FatArrow:     "=>",
	Shl:          "<<",
	Shr:          ">>",
	Plus:         "+",
	Minus:        "-",
	Asterisk:     "*",
	Slash:        "/",
	Percent:      "%",
	Amp:          "&",
	Pipe:         "|",
	Caret:        "^",
	At:           "@",
	PlusEq:       "+=",
	MinusEq:      "-=",
	AsteriskEq:   "*=",
	SlashEq:      "/=",
	PercentEq:    "%=",
	Dot2:         "..",
	Dot2Less:     "..<",
	Ellipsis:     "...",
	KwModule:     "module",
	KwAlias:      "alias",
	KwStruct:     "struct",
	KwEnum:       "enum",
	KwFunc:       "func",
	KwVal:        "val",
	KwVar:        "var",
	KwConst:      "const",
	KwOf:         "of",
	KwIf:         "if",
	KwElse:       "else",
	KwWhile:      "while",
	KwReturn:     "return",
	KwBreak:      "break",
	KwContinue:   "continue",
	KwAnd:        "and",
	KwOr:         "or",
}
