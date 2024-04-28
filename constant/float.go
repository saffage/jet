package constant

import "math/big"

func NewFloat(value *big.Float) floatValue {
	return floatValue{value}
}

type floatValue struct {
	value *big.Float
}

func (v floatValue) Kind() Kind     { return Float }
func (v floatValue) String() string { return v.value.String() }
