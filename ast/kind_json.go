package ast

import "encoding/json"

func (kind LiteralKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind PrefixOpKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind InfixOpKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind PostfixOpKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind GenericDeclKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}
