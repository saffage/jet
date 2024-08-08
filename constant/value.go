package constant

import (
	"fmt"
	"math/big"
	"strconv"
)

type Kind byte

//go:generate stringer -type=Kind
const (
	Int Kind = iota
	Float
	String
	Bool
	Array
)

type Value interface {
	Kind() Kind
	implValue()
}

func NewBigInt(value *big.Int) Value {
	if value == nil {
		panic("nil argument")
	}
	return &intValue{value}
}

func NewBigFloat(value *big.Float) Value {
	if value == nil {
		panic("nil argument")
	}
	return &floatValue{value}
}

func NewBool(value bool) Value     { return &boolValue{value} }
func NewInt(value int64) Value     { return &intValue{big.NewInt(value)} }
func NewFloat(value float64) Value { return &floatValue{big.NewFloat(value)} }
func NewString(value string) Value { return &stringValue{value} }
func NewArray(value []Value) Value { return &arrayValue{value} }

func IsInt(value Value) bool    { return value.Kind() == Int }
func IsFloat(value Value) bool  { return value.Kind() == Float }
func IsString(value Value) bool { return value.Kind() == String }
func IsBool(value Value) bool   { return value.Kind() == Bool }
func IsArray(value Value) bool  { return value.Kind() == Array }

func AsInt(value Value) *big.Int {
	if IsInt(value) {
		return value.(*intValue).val
	}
	return nil
}

func AsFloat(value Value) *big.Float {
	if IsFloat(value) {
		return value.(*floatValue).val
	}
	return nil
}

func AsString(value Value) *string {
	if IsString(value) {
		val := value.(*stringValue).val
		return &val
	}
	return nil
}

func AsBool(value Value) *bool {
	if IsBool(value) {
		val := value.(*boolValue).val
		return &val
	}
	return nil
}

func AsArray(value Value) *[]Value {
	if IsArray(value) {
		val := value.(*arrayValue).val
		return &val
	}
	return nil
}

//------------------------------------------------
// Value implementation
//------------------------------------------------

type (
	intValue    struct{ val *big.Int }
	floatValue  struct{ val *big.Float }
	stringValue struct{ val string }
	boolValue   struct{ val bool }
	arrayValue  struct{ val []Value }
)

func (v *intValue) String() string    { return v.val.String() }
func (v *floatValue) String() string  { return v.val.String() }
func (v *stringValue) String() string { return strconv.Quote(v.val) }
func (v *boolValue) String() string   { return strconv.FormatBool(v.val) }
func (v *arrayValue) String() string  { return fmt.Sprintf("%v", v.val) }

func (v *intValue) Kind() Kind    { return Int }
func (v *floatValue) Kind() Kind  { return Float }
func (v *stringValue) Kind() Kind { return String }
func (v *boolValue) Kind() Kind   { return Bool }
func (v *arrayValue) Kind() Kind  { return Array }

func (intValue) implValue()    {}
func (floatValue) implValue()  {}
func (stringValue) implValue() {}
func (boolValue) implValue()   {}
func (arrayValue) implValue()  {}
