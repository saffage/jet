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

	EOF     // end of file
	Comment // comment

	Type       // type
	TypeVar    // type var
	Name       // name
	Underscore // underscore
	Int        // untyped int
	Float      // untyped float
	String     // untyped string

	LParen   // '('
	RParen   // ')'
	LCurly   // '{'
	RCurly   // '}'
	LBracket // '['
	RBracket // ']'
	Colon    // ':'
	Comma    // ','

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
	Caret      // operator '^'
	CaretEq    // operator '^='
	Amp        // operator '&'
	AmpEq      // operator '&='
	And        // operator '&&'
	AndEq      // operator '&&='
	Pipe       // operator '|'
	PipeEq     // operator '|='
	Or         // operator '||'
	OrEq       // operator '||='

	// End of position dependent tokens.

	Arrow    // operator '->'
	FatArrow // operator '=>'
	Dot      // operator '.'
	Dot2     // operator '..'

	KwLet      // keyword 'let'
	KwType     // keyword 'type'
	KwExtern   // keyword 'extern'
	KwWith     // keyword 'with'
	KwWhen     // keyword 'when'
	KwIf       // keyword 'if'
	KwElse     // keyword 'else'
	KwIn       // keyword 'in'
	KwAs       // keyword 'as'
	KwDefer    // keyword 'defer'
	KwBreak    // keyword 'break'
	KwReturn   // keyword 'return'
	KwContinue // keyword 'continue'
)

const (
	_SpecialBegin     = EOF
	_SpecialEnd       = Comment
	_PrimaryBegin     = Type
	_PrimaryEnd       = String
	_PunctuationBegin = LParen
	_PunctuationEnd   = Comma
	_OperatorBegin    = Eq
	_OperatorEnd      = Dot2
	_KeywordsBegin    = KwLet
	_KeywordsEnd      = KwContinue
	_LastKind         = _KeywordsEnd
)

// Returns `Illegal` if `s` is not a kind name.
func KindFromBytes(b []byte) Kind {
	return KindFromString(string(b))
}

// Returns `Illegal` if `s` is not a kind name.
func KindFromByte(b byte) Kind {
	return KindFromString(string(b))
}

// Returns `Illegal` if `s` is not a kind name.
func KindFromRune(r rune) Kind {
	return KindFromString(string(r))
}

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
	return _SpecialBegin <= kind && kind <= _SpecialEnd
}

func (kind Kind) IsPrimary() bool {
	return _PrimaryBegin <= kind && kind <= _PrimaryEnd
}

func (kind Kind) IsPunctuation() bool {
	return _PunctuationBegin <= kind && kind <= _PunctuationEnd
}

func (kind Kind) IsOperator() bool {
	return _OperatorBegin <= kind && kind <= _OperatorEnd
}

func (kind Kind) IsKeyword() bool {
	return _KeywordsBegin <= kind && kind <= _KeywordsEnd
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
	return getKinds(int(_SpecialBegin), int(_SpecialEnd))
}

func PrimaryKinds() []Kind {
	return getKinds(int(_PrimaryBegin), int(_PrimaryEnd))
}

func PunctuationKinds() []Kind {
	return getKinds(int(_PunctuationBegin), int(_PunctuationEnd))
}

func OperatorKinds() []Kind {
	return getKinds(int(_OperatorBegin), int(_OperatorEnd))
}

func KeywordKinds() []Kind {
	return getKinds(int(_KeywordsBegin), int(_KeywordsEnd))
}

func AllKinds() []Kind {
	return getKinds(0, int(_LastKind))
}

var representableKinds = map[Kind]string{
	LParen:     "(",
	RParen:     ")",
	LCurly:     "{",
	RCurly:     "}",
	LBracket:   "[",
	RBracket:   "]",
	Colon:      ":",
	Comma:      ",",
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
	And:        "&&",
	AndEq:      "&&=",
	Pipe:       "|",
	PipeEq:     "|=",
	Or:         "||",
	OrEq:       "||=",
	Caret:      "^",
	CaretEq:    "^=",
	Arrow:      "->",
	FatArrow:   "=>",
	Dot:        ".",
	Dot2:       "..",
	KwLet:      "let",
	KwType:     "type",
	KwExtern:   "extern",
	KwWith:     "with",
	KwWhen:     "when",
	KwIf:       "if",
	KwElse:     "else",
	KwIn:       "in",
	KwAs:       "as",
	KwDefer:    "defer",
	KwBreak:    "break",
	KwReturn:   "return",
	KwContinue: "continue",
}
