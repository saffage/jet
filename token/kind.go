package token

import "fmt"

type Kind byte

const (
	_special_begin Kind = iota
	Illegal             // any illegal character
	EOF                 // end of file
	Comment             // any comment
	Whitespace          // ' '
	NewLine             // '\n'
	_special_end

	_primary_begin
	Ident  // foo
	Int    // 10
	Float  // 10.0, 1.0e+2
	String // "str", 'str'
	_primary_end

	_punctuation_begin
	LParen    // '('
	RParen    // ')'
	LCurly    // '{'
	RCurly    // '}'
	LBracket  // '['
	RBracket  // ']'
	Dot       // '.'
	Comma     // ','
	Colon     // ':'
	Semicolon // ';'
	_punctuation_end

	_operator_begin

	// The following operators are position dependent
	// in the current implementation of the scanner.

	Eq         // '='
	EqOp       // '=='
	Bang       // '!'
	NeOp       // '!='
	LtOp       // '<'
	LeOp       // '<='
	GtOp       // '>'
	GeOp       // '>='
	Plus       // '+'
	PlusEq     // '+='
	Minus      // '-'
	MinusEq    // '-='
	Asterisk   // '*'
	AsteriskEq // '*='
	Slash      // '/'
	SlashEq    // '/='
	Percent    // '%'
	PercentEq  // '%='

	// End of position dependent tokens.

	Amp          // '&'
	Hash         // '#'
	At           // '@'
	QuestionMark // '?'
	Arrow        // '->'
	FatArrow     // '=>'
	Dot2         // '..'
	Dot2Less     // '..<'
	Ellipsis     // '...'
	_operator_end

	// NOTE some keywords are unused.

	_keywords_begin
	KwModule   // 'module'
	KwAlias    // 'alias'
	KwStruct   // 'struct'
	KwEnum     // 'enum'
	KwFunc     // 'func'
	KwVal      // 'val'
	KwVar      // 'var'
	KwConst    // 'const'
	KwOf       // 'of'
	KwIf       // 'if'
	KwElse     // 'else'
	KwWhile    // 'while'
	KwReturn   // 'return'
	KwBreak    // 'break'
	KwContinue // 'continue'
	_keywords_end
)

// Returns `Illegal` if `s` is not a kind.
func KindFromString(s string) Kind {
	for kind, str := range kindStrings {
		if str == s {
			return kind
		}
	}
	return Illegal
}

func (kind Kind) IsSpecial() bool {
	return _special_begin < kind && kind < _special_end
}

func (kind Kind) IsPrimary() bool {
	return _primary_begin < kind && kind < _primary_end
}

func (kind Kind) IsPunctuation() bool {
	return _punctuation_begin < kind && kind < _punctuation_end
}

func (kind Kind) IsOperator() bool {
	return _operator_begin < kind && kind < _operator_end
}

func (kind Kind) IsKeyword() bool {
	return _keywords_begin < kind && kind < _keywords_end
}

func (kind Kind) Name() string {
	return kindNames[kind]
}

func (kind Kind) String() string {
	switch {
	case kind.IsSpecial(), kind.IsPrimary():
		return kindsReadable[kind]

	case kind.IsKeyword():
		return "keyword `" + kindStrings[kind] + "`"

	case kind.IsOperator():
		return "operator `" + kindStrings[kind] + "`"
	}
	return "`" + kindStrings[kind] + "`"
}

func getKinds(begin, end int) []Kind {
	kinds := make([]Kind, 0, end-begin-1)
	for kind := begin + 1; kind < end; kind++ {
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
	kinds := make([]Kind, 0, allKindsCound)
	kinds = append(kinds, SpecialKinds()...)
	kinds = append(kinds, PrimaryKinds()...)
	kinds = append(kinds, PunctuationKinds()...)
	kinds = append(kinds, OperatorKinds()...)
	kinds = append(kinds, KeywordKinds()...)
	return kinds
}

var kindNames = map[Kind]string{
	Illegal:      "Illegal",
	EOF:          "EOF",
	Comment:      "Comment",
	Whitespace:   "Whitespace",
	NewLine:      "NewLine",
	Ident:        "Identifier",
	Int:          "Int",
	Float:        "Float",
	String:       "String",
	LParen:       "LParen",
	RParen:       "RParen",
	LCurly:       "LCurly",
	RCurly:       "RCurly",
	LBracket:     "LBracket",
	RBracket:     "RBracket",
	Dot:          "Dot",
	Comma:        "Comma",
	Colon:        "Colon",
	Semicolon:    "Semicolon",
	Eq:           "Eq",
	Bang:         "Bang",
	QuestionMark: "QuestionMark",
	EqOp:         "EqOp",
	NeOp:         "NeOp",
	LtOp:         "LtOp",
	GtOp:         "GtOp",
	LeOp:         "LeOp",
	GeOp:         "GeOp",
	Arrow:        "Arrow",
	FatArrow:     "FatArrow",
	Plus:         "Plus",
	Minus:        "Minus",
	Asterisk:     "Asterisk",
	Slash:        "Slash",
	Percent:      "Percent",
	Amp:          "Amp",
	Hash:         "Hash",
	At:           "At",
	PlusEq:       "PlusEq",
	MinusEq:      "MinusEq",
	AsteriskEq:   "AsteriskEq",
	SlashEq:      "SlashEq",
	PercentEq:    "PercentEq",
	Dot2:         "Dot2",
	Dot2Less:     "Dot2Less",
	Ellipsis:     "Dot3",
	KwModule:     "KwModule",
	KwAlias:      "KwAlias",
	KwStruct:     "KwStruct",
	KwEnum:       "KwEnum",
	KwFunc:       "KwFunc",
	KwVal:        "KwVal",
	KwVar:        "KwVar",
	KwConst:      "KwConst",
	KwOf:         "KwOf",
	KwIf:         "KwIf",
	KwElse:       "KwElse",
	KwWhile:      "KwWhile",
	KwReturn:     "KwReturn",
	KwBreak:      "KwBreak",
	KwContinue:   "KwContinue",
}

var kindsReadable = map[Kind]string{
	Illegal:    "illegal",
	EOF:        "end of file",
	Comment:    "comment",
	Whitespace: "whitespace",
	NewLine:    "new line",
	Ident:      "identifier",
	Int:        "untyped int",
	Float:      "untyped float",
	String:     "untyped string",
}

var kindStrings = map[Kind]string{
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
	Plus:         "+",
	Minus:        "-",
	Asterisk:     "*",
	Slash:        "/",
	Percent:      "%",
	Amp:          "&",
	Hash:         "#",
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
}

const (
	specialKindsCount     = int(_special_end-_special_begin) - 1
	primaryKindsCount     = int(_primary_end-_primary_begin) - 1
	punctuationKindsCount = int(_punctuation_end-_punctuation_begin) - 1
	operatorKindsCount    = int(_operator_end-_operator_begin) - 1
	keywordKindsCount     = int(_keywords_end-_keywords_begin) - 1
	allKindsCound         = specialKindsCount + primaryKindsCount + punctuationKindsCount + operatorKindsCount + keywordKindsCount
)

func findMissingKinds(m map[Kind]string, expected []Kind) []Kind {
	missingKinds := make([]Kind, 0, len(expected)-len(m))

	for _, kind := range expected {
		if _, inMap := m[kind]; !inMap {
			missingKinds = append(missingKinds, kind)
		}
	}

	return missingKinds
}

func checkMissingKindsIsMaps() {
	missingKinds := findMissingKinds(
		kindNames,
		AllKinds(),
	)

	if len(missingKinds) != 0 {
		panic(fmt.Sprintf("not all kinds have a string representation in map; missing %#v", missingKinds))
	}

	missingKinds = findMissingKinds(
		kindStrings,
		append(PunctuationKinds(), append(OperatorKinds(), KeywordKinds()...)...),
	)

	if len(missingKinds) != 0 {
		panic(fmt.Sprintf("not all stringable kinds have a string representation in map; missing %#v", missingKinds))
	}

	missingKinds = findMissingKinds(
		kindsReadable,
		append(SpecialKinds(), PrimaryKinds()...),
	)

	if len(missingKinds) != 0 {
		panic(fmt.Sprintf("not all kinds have a readable string representation in map; missing %#v", missingKinds))
	}
}

func init() {
	checkMissingKindsIsMaps()
}
