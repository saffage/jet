package types

import "strings"

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

type Function struct {
	params   Params
	result   Type
	variadic Type
}

func NewFunction(params Params, result, variadic Type) *Function {
	return &Function{params, result, variadic}
}

func (t *Function) Equal(expected Type) bool {
	if expected, _ := As[*Function](expected); expected != nil {
		return (t.variadic != nil && t.variadic.Equal(expected.variadic) ||
			t.variadic == nil && expected.variadic == nil) &&
			t.result.Equal(expected.result) && t.params.Equal(expected.params)
	}

	return false
}

func (t *Function) Underlying() Type {
	return t
}

func (t *Function) String() string {
	if t.result != nil {
		return t.params.String() + " " + t.result.String()
	}
	return t.params.String() + " '_"
}

func (t *Function) Result() Type {
	return t.result
}

func (t *Function) Params() Params {
	return t.params
}

func (t *Function) Variadic() Type {
	return t.variadic
}

func (t *Function) CheckArgValues(values []*Value) (idx int, err error) {
	args := make(Params, len(values))

	for i := range args {
		args[i] = values[i].T
	}

	return t.CheckArgs(args)
}

func (t *Function) CheckArgs(args Params) (int, error) {
	// params 	args 	diff 	idx
	//      1      2      -1      1
	//      2      1       1      1
	//      0      3      -3      0
	//      3      0       3      0
	diff := len(t.params) - len(args)

	if diff < 0 && t.variadic == nil {
		return min(
				len(t.params),
				len(args),
			), &errorIncorrectArity{
				node:     nil,
				expected: len(t.params),
				got:      len(args),
			}
	}

	if diff > 0 {
		return min(
				len(t.params),
				len(args),
			), &errorIncorrectArity{
				node:     nil,
				expected: len(t.params),
				got:      len(args),
			}
	}

	for i := 0; i < len(t.params); i++ {
		expected, actual := t.params[i], args[i]

		if !actual.Equal(expected) {
			return i, &errorArgTypeMismatch{
				node:      nil,
				tExpected: expected,
				tArg:      actual,
				index:     i,
			}
		}
	}

	// Check variadic.
	for i, arg := range args[len(t.params):] {
		if !arg.Equal(t.variadic) {
			return i + len(t.params), &errorArgTypeMismatch{
				node:      nil,
				tExpected: t.variadic,
				tArg:      arg,
				index:     i + len(t.params),
				variadic:  true,
			}
		}
	}

	return -1, nil
}
