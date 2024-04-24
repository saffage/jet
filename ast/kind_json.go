package ast

import "encoding/json"

func (kind LiteralKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind UnaryOpKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind BinaryOpKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind GenericDeclKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
