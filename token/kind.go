package token

import "fmt"

//go:generate stringer -type=Kind -linecomment
type Kind byte

const (
	Illegal Kind = iota // illegal character

	EOF     // end of file
	Comment // comment

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
	Shl        // operator '<<'
	ShlEq      // operator '<<='
	Shr        // operator '>>'
	ShrEq      // operator '>>='
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
	Amp        // operator '&'
	AmpEq      // operator '&='
	Pipe       // operator '|'
	PipeEq     // operator '|='
	Caret      // operator '^'
	CaretEq    // operator '^='

	// End of position dependent tokens.

	At       // operator '@'
	Dollar   // operator '$'
	Arrow    // operator '->'
	FatArrow // operator '=>'
	Dot      // operator '.'
	Dot2     // operator '..'
	Dot2Less // operator '..<'
	Ellipsis // operator '...'

	KwAnd      // keyword 'and'
	KwAs       // keyword 'as'
	KwBreak    // keyword 'break'
	KwContinue // keyword 'continue'
	KwDefer    // keyword 'defer'
	KwElse     // keyword 'else'
	KwEnum     // keyword 'enum'
	KwFor      // keyword 'for'
	KwIf       // keyword 'if'
	KwIn       // keyword 'in'
	KwMut      // keyword 'mut'
	KwOr       // keyword 'or'
	KwReturn   // keyword 'return'
	KwStruct   // keyword 'struct'
	KwWhile    // keyword 'while'
)

const (
	_special_begin = EOF
	_special_end   = Comment

	_primary_begin = Ident
	_primary_end   = String

	_punctuation_begin = LParen
	_punctuation_end   = Semicolon

	_operator_begin = Eq
	_operator_end   = Ellipsis

	_keywords_begin = KwAnd
	_keywords_end   = KwWhile

	_kinds_last = _keywords_end
)

type stringLike interface {
	~[]byte | ~string | ~rune | ~byte
}

// Returns `Illegal` if `s` is not a kind name.
func KindFrom[T stringLike](s T) Kind {
	x := string(s)
	for kind, str := range representableKinds {
		if str == x {
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
	LParen:    "(",
	RParen:    ")",
	LCurly:    "{",
	RCurly:    "}",
	LBracket:  "[",
	RBracket:  "]",
	Comma:     ",",
	Colon:     ":",
	Semicolon: ";",

	Eq:         "=",
	EqOp:       "==",
	Bang:       "!",
	NeOp:       "!=",
	LtOp:       "<",
	LeOp:       "<=",
	GtOp:       ">",
	GeOp:       ">=",
	Shl:        "<<",
	ShlEq:      "<<=",
	Shr:        ">>",
	ShrEq:      ">>=",
	Plus:       "+",
	PlusEq:     "+=",
	Minus:      "-",
	MinusEq:    "-=",
	Asterisk:   "*",
	AsteriskEq: "*=",
	Slash:      "/",
	SlashEq:    "/=",
	Percent:    "%",
	PercentEq:  "%=",
	Amp:        "&",
	AmpEq:      "&=",
	Pipe:       "|",
	PipeEq:     "|=",
	Caret:      "^",
	CaretEq:    "^=",

	At:       "@",
	Dollar:   "$",
	Arrow:    "->",
	FatArrow: "=>",
	Dot:      ".",
	Dot2:     "..",
	Dot2Less: "..<",
	Ellipsis: "...",

	KwAnd:      "and",
	KwAs:       "as",
	KwBreak:    "break",
	KwContinue: "continue",
	KwDefer:    "defer",
	KwElse:     "else",
	KwEnum:     "enum",
	KwFor:      "for",
	KwIf:       "if",
	KwIn:       "in",
	KwMut:      "mut",
	KwOr:       "or",
	KwReturn:   "return",
	KwStruct:   "struct",
	KwWhile:    "while",
}
