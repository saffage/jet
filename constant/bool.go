package constant

import "strconv"

func NewBool(value bool) Value {
	return boolValue(value)
}

type boolValue bool

func (v boolValue) Kind() Kind     { return Bool }
func (v boolValue) String() string { return strconv.FormatBool(bool(v)) }
