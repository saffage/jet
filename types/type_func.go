package types

import (
	"fmt"
	"strings"

	"github.com/saffage/jet/util"
)

type Params []Type

func (params Params) Equal(target []Type) bool {
	if len(params) != len(target) {
		return false
	}

	for i, param := range params {
		if !param.Equal(SkipAlias(target[i])) {
			return false
		}
	}

	return true
}

func (params Params) String() string {
	buf := strings.Builder{}
	buf.WriteByte('(')

	for i, param := range params {
		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(param.String())
	}

	buf.WriteByte(')')
	return buf.String()
}

type Func struct {
	params   Params
	result   Type
	variadic Type
}

func NewFunc(params Params, result, variadic Type) *Func {
	return &Func{params, result, variadic}
}

func (t *Func) Equal(expected Type) bool {
	if expected := As[*Func](expected); expected != nil {
		return (t.variadic != nil && t.variadic.Equal(expected.variadic) ||
			t.variadic == nil && expected.variadic == nil) &&
			t.result.Equal(expected.result) && t.params.Equal(expected.params)
	}

	return false
}

func (t *Func) Underlying() Type {
	return t
}

func (t *Func) String() string {
	if t.result != nil {
		return t.params.String() + " " + t.result.String()
	}
	return t.params.String() + " '_"
}

func (t *Func) Result() Type {
	return t.result
}

func (t *Func) Params() Params {
	return t.params
}

func (t *Func) Variadic() Type {
	return t.variadic
}

func (t *Func) CheckArgValues(values []*Value) (idx int, err error) {
	args := make(Params, len(values))

	for i := range args {
		args[i] = values[i].T
	}

	return t.CheckArgs(args)
}

func (t *Func) CheckArgs(args Params) (idx int, err error) {
	{
		diff := len(t.params) - len(args)

		// params 	args 	diff 	idx
		//      1      2      -1      1
		//      2      1       1      1
		//      0      3      -3      0
		//      3      0       3      0

		if diff < 0 && t.variadic == nil {
			return min(len(t.params), len(args)), fmt.Errorf(
				"too many arguments (expected %d, got %d)",
				len(t.params),
				len(args),
			)
		}

		if diff > 0 {
			return min(len(t.params), len(args)), fmt.Errorf(
				"not enough arguments (expected %d, got %d)",
				len(t.params),
				len(args),
			)
		}
	}

	for i := 0; i < len(t.params); i++ {
		expected, actual := t.params[i], args[i]

		if !actual.Equal(expected) {
			return i, fmt.Errorf(
				"expected '%s' for %s argument, got '%s' instead",
				expected,
				util.OrdinalSuffix(i+1),
				actual,
			)
		}
	}

	// Check variadic.
	for i, arg := range args[len(t.params):] {
		if !arg.Equal(t.variadic) {
			return i + len(t.params), fmt.Errorf(
				"expected '%s', got '%s' instead",
				t.variadic,
				arg,
			)
		}
	}

	return -1, nil
}
