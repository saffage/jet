package types

import (
	"fmt"
	"strconv"
	"strings"
)

type Func struct {
	params *Tuple
	result *Tuple
}

func NewFunc(result *Tuple, params *Tuple) *Func {
	if result == nil {
		result = Unit
	}

	if params == nil {
		params = Unit
	}

	return &Func{
		params: params,
		result: result,
	}
}

func (t *Func) Equals(other Type) bool {
	if otherFunc, ok := other.(*Func); ok {
		return t.result.Equals(otherFunc.result) &&
			t.params.Equals(otherFunc.params)
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

func (t *Func) CheckArgs(args *Tuple) (idx int, err error) {
	{
		diff := t.params.Len() - args.Len()

		// params 	args 	diff 	idx
		//      1      2      -1      1
		//      2      1       1      1
		//      0      3      -3      0
		//      3      0       3      0

		if diff < 0 {
			return min(t.params.Len(), args.Len()),
				fmt.Errorf("too many arguments (expected %d, got %d)", t.params.Len(), args.Len())
		}

		if diff > 0 {
			return min(t.params.Len(), args.Len()),
				fmt.Errorf("not enough arguments (expected %d, got %d)", t.params.Len(), args.Len())
		}
	}

	for i := 0; i < args.Len(); i++ {
		expected, actual := t.params.types[i], args.types[i]

		if !expected.Equals(actual) {
			return i, fmt.Errorf(
				"expected '%s' for %s argument, got '%s' instead",
				expected,
				ordinalize(i+1),
				actual,
			)
		}
	}

	return -1, nil
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
