package constant

import (
	"math/big"
)

func NewInt(value *big.Int) Value {
	return intValue{value}
}

type intValue struct {
	value *big.Int
}

func (v intValue) Kind() Kind     { return Int }
func (v intValue) String() string { return v.value.String() }
