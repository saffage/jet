package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/saffage/jet/constant"
)

type Func struct {
	params   *Tuple
	result   *Tuple
	variadic Type
}

func NewFunc(params, result *Tuple, variadic Type) *Func {
	if params == nil {
		params = Unit
	}
	// if result == nil {
	// 	result = Unit
	// }
	return &Func{
		params:   params,
		result:   result,
		variadic: variadic,
	}
}

func (t *Func) Equals(other Type) bool {
	if t2 := AsPrimitive(other); t2 != nil {
		return t2.kind == KindAny
	}
	if t2 := AsFunc(other); t2 != nil {
		return (t.variadic != nil && t.variadic.Equals(t2.variadic) ||
			t.variadic == nil && t2.variadic == nil) &&
			t.result.Equals(t2.result) && t.params.Equals(t2.params)
	}
	return false
}

func (t *Func) Underlying() Type { return t }

func (t *Func) String() string {
	buf := strings.Builder{}
	buf.WriteString("func")
	buf.WriteString(t.params.String())

	if !t.result.Equals(Unit) {
		if t.result.Len() == 1 {
			buf.WriteByte(' ')
			buf.WriteString(t.result.types[0].String())
		} else {
			buf.WriteByte(' ')
			buf.WriteString(t.result.String())
		}
	}

	return buf.String()
}

func (t *Func) Result() *Tuple { return t.result }

func (t *Func) Params() *Tuple { return t.params }

func (t *Func) Variadic() Type { return t.variadic }

func (t *Func) CheckArgValues(args []constant.Value) (idx int, err error) {
	tyArgs := &Tuple{types: make([]Type, len(args))}
	for _, arg := range args {
		tyArgs.types = append(tyArgs.types, FromConstant(arg))
	}
	return t.CheckArgs(tyArgs)
}

func (t *Func) CheckArgs(args *Tuple) (idx int, err error) {
	{
		diff := t.params.Len() - args.Len()

		// params 	args 	diff 	idx
		//      1      2      -1      1
		//      2      1       1      1
		//      0      3      -3      0
		//      3      0       3      0

		if diff < 0 && t.variadic == nil {
			return min(t.params.Len(), args.Len()),
				fmt.Errorf("too many arguments (expected %d, got %d)", t.params.Len(), args.Len())
		}

		if diff > 0 {
			return min(t.params.Len(), args.Len()),
				fmt.Errorf("not enough arguments (expected %d, got %d)", t.params.Len(), args.Len())
		}
	}

	for i := 0; i < t.params.Len(); i++ {
		expected, actual := t.params.types[i], args.types[i]

		if !actual.Equals(expected) {
			return i, fmt.Errorf(
				"expected '%s' for %s argument, got '%s' instead",
				expected,
				ordinalize(i+1),
				actual,
			)
		}
	}

	// Check varargs.
	for i, arg := range args.types[t.params.Len():] {
		if !arg.Equals(t.variadic) {
			return i + t.params.Len(), fmt.Errorf(
				"expected '%s', got '%s' instead",
				t.variadic,
				arg,
			)
		}
	}

	return -1, nil
}

func IsFunc(t Type) bool { return AsFunc(t) != nil }

func AsFunc(t Type) *Func {
	if t != nil {
		if fn, _ := t.Underlying().(*Func); fn != nil {
			return fn
		}
	}

	return nil
}

func ordinalize(num int) string {
	s := strconv.Itoa(num)

	switch num % 100 {
	case 11, 12, 13:
		return s + "th"

	default:
		switch num % 10 {
		case 1:
			return s + "st"

		case 2:
			return s + "nd"

		case 3:
			return s + "rd"

		default:
			return s + "th"
		}
	}
}
