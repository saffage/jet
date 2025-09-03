package types

type CircularTypes struct {
	MetaTypeVariable uint
	T                Type
}

func (CircularTypes) Error() string { return "" }

type IncompatibleTypesError interface {
	error
}
