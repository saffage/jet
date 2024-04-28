package constant

type Kind byte

const (
	Unknown Kind = iota

	Bool // TODO delete and implement through attributes.
	Int
	Float
	String
	Expression
)
