package token

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=Kind -linecomment
type Kind byte

const (
	Illegal Kind = iota // illegal character

	EOF     // end of file
	Comment // comment
	Newline // newline

	LowercaseIdent   // lowercase identifier
	UppercaseIdent   // uppercase identifier
	IdentPlaceholder // identifier placeholder
	Int              // int literal
	Float            // float literal
	String           // string literal

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

	At       // '@'
	Dollar   // '$'
	Arrow    // '->'
	FatArrow // '=>'
	Dot      // '.'
	Dot2     // '..'
	Ellipsis // '...'

	KwExternal // keyword 'external'
	KwFn       // keyword 'fn'
	KwLet      // keyword 'let'
	KwType     // keyword 'type'
	KwVal      // keyword 'val'
	KwVar      // keyword 'var'
	KwWhen     // keyword 'when'

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
	_special_end   = Newline

	_primary_begin = LowercaseIdent
	_primary_end   = String

	_punctuation_begin = LParen
	_punctuation_end   = Semicolon

	_operator_begin = Eq
	_operator_end   = Ellipsis

	_keywords_begin = KwExternal
	_keywords_end   = KwWhile

	_reserved_begin = KwAnd
	_reserved_end   = KwWhile

	_kinds_first = _special_begin
	_kinds_last  = _reserved_end
)

type stringLike interface {
	~[]byte | ~[]rune | ~string | ~rune | ~byte
}

func FromRepresentation[T stringLike](s T) Kind {
	x := string(s)
	for kind, str := range representableKinds {
		if str == x {
			return kind
		}
	}
	return Illegal
}

func Representation(kind Kind) string {
	return representableKinds[kind]
}

func (kind Kind) IsSpecial() bool         { return _special_begin <= kind && kind <= _special_end }
func (kind Kind) IsPrimary() bool         { return _primary_begin <= kind && kind <= _primary_end }
func (kind Kind) IsPunctuation() bool     { return _punctuation_begin <= kind && kind <= _punctuation_end }
func (kind Kind) IsOperator() bool        { return _operator_begin <= kind && kind <= _operator_end }
func (kind Kind) IsKeyword() bool         { return _keywords_begin <= kind && kind <= _keywords_end }
func (kind Kind) IsReservedKeyword() bool { return _reserved_begin <= kind && kind <= _reserved_end }
func (kind Kind) Renderable() bool        { return _punctuation_begin <= kind && kind <= _reserved_end }

func (kind Kind) Render(buf *strings.Builder) {
	representation, ok := representableKinds[kind]

	if !ok {
		panic(fmt.Sprintf("%s cannot be represented as string (not enough data)", kind))
	}

	buf.WriteString(representation)
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
	Ellipsis: "...",

	KwExternal: "external",
	KwFn:       "fn",
	KwLet:      "let",
	KwType:     "type",
	KwVal:      "val",
	KwVar:      "var",
	KwWhen:     "when",

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
