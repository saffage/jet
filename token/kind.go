package token

import "fmt"

//go:generate stringer -type=Kind -linecomment
type Kind byte

const (
	Illegal Kind = iota // illegal character

	EOF     // end of file
	Comment // comment

	LowercaseIdent   // lowercase identifier
	UppercaseIdent   // uppercase identifier
	IdentPlaceholder // identifier placeholder
	Int              // untyped int
	Float            // untyped float
	String           // untyped string

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
	Bar        // operator '|'
	BarEq      // operator '|='
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

	KwExtern // keyword 'extern'
	KwFn     // keyword 'fn'
	KwLet    // keyword 'let'
	KwType   // keyword 'type'
	KwVal    // keyword 'val'
	KwVar    // keyword 'var'
	KwWhen   // keyword 'when'

	// Reserved keywords.

	KwAnd      // keyword 'and'
	KwAs       // keyword 'as'
	KwBreak    // keyword 'break'
	KwContinue // keyword 'continue'
	KwDefer    // keyword 'defer'
	KwElse     // keyword 'else'
	KwFor      // keyword 'for'
	KwIf       // keyword 'if'
	KwIn       // keyword 'in'
	KwOf       // keyword 'of'
	KwOr       // keyword 'or'
	KwReturn   // keyword 'return'
	KwWhile    // keyword 'while'
)

const (
	_special_begin = EOF
	_special_end   = Comment

	_primary_begin = LowercaseIdent
	_primary_end   = String

	_punctuation_begin = LParen
	_punctuation_end   = Semicolon

	_operator_begin = Eq
	_operator_end   = Ellipsis

	_keywords_begin = KwExtern
	_keywords_end   = KwWhile

	_reserved_begin = KwAnd
	_reserved_end   = KwWhile

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

func (kind Kind) IsSpecial() bool         { return _special_begin <= kind && kind <= _special_end }
func (kind Kind) IsPrimary() bool         { return _primary_begin <= kind && kind <= _primary_end }
func (kind Kind) IsPunctuation() bool     { return _punctuation_begin <= kind && kind <= _punctuation_end }
func (kind Kind) IsOperator() bool        { return _operator_begin <= kind && kind <= _operator_end }
func (kind Kind) IsKeyword() bool         { return _keywords_begin <= kind && kind <= _keywords_end }
func (kind Kind) IsReservedKeyword() bool { return _reserved_begin <= kind && kind <= _reserved_end }

// TODO rename to `Render`, maybe add `Renderer` interface to [report] package?
func (kind Kind) Repr() string {
	s, ok := representableKinds[kind]

	if !ok {
		panic(fmt.Sprintf("%s cannot be represented as string (not enough data)", kind))
	}

	return s
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
	Bar:        "|",
	BarEq:      "|=",
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

	KwExtern: "extern",
	KwFn:     "fn",
	KwLet:    "let",
	KwType:   "type",
	KwVal:    "val",
	KwVar:    "var",
	KwWhen:   "when",

	KwAnd:      "and",
	KwAs:       "as",
	KwBreak:    "break",
	KwContinue: "continue",
	KwDefer:    "defer",
	KwElse:     "else",
	KwFor:      "for",
	KwIf:       "if",
	KwIn:       "in",
	KwOf:       "of",
	KwOr:       "or",
	KwReturn:   "return",
	KwWhile:    "while",
}
