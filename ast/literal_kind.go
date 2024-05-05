package ast

import "encoding/json"

//go:generate stringer -type=LiteralKind -linecomment -output=literal_kind_string.go
type LiteralKind byte

const (
	UnknownLiteral LiteralKind = iota // unknown literal kind

	IntLiteral    // int
	FloatLiteral  // float
	StringLiteral // string
)

func (kind LiteralKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
