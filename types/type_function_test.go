package types

import "testing"

func TestCheckArgs(t *testing.T) {
	checkArgs(
		t,
		NewFn(TypeList{NoneType}, NoneType, nil),
		TypeList{NoneType},
		"",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType}, NoneType, nil),
		TypeList{IntType},
		"",
	)
}

func TestCheckArgsFail(t *testing.T) {
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, nil),
		TypeList{IntType},
		"incorrect arity, too many arguments (expected 0 arguments, got 1)",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType}, NoneType, nil),
		TypeList{},
		"incorrect arity, not enough arguments (expected 1 arguments, got 0)",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType}, NoneType, nil),
		TypeList{IntType, BoolType, IntType},
		"incorrect arity, too many arguments (expected 1 arguments, got 3)",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType, BoolType, IntType}, NoneType, nil),
		TypeList{IntType},
		"incorrect arity, not enough arguments (expected 3 arguments, got 1)",
	)
	checkArgs(
		t,
		NewFn(TypeList{BoolType}, NoneType, nil),
		TypeList{IntType},
		"argument type mismatch (expected `Bool` for 1-st argument, got `Int`)",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType, BoolType}, NoneType, nil),
		TypeList{IntType, IntType},
		"argument type mismatch (expected `Bool` for 2-nd argument, got `Int`)",
	)
}

func TestCheckArgsVariadic(t *testing.T) {
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, IntType),
		TypeList{},
		"",
	)
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, IntType),
		TypeList{IntType},
		"",
	)
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, IntType),
		TypeList{IntType, IntType},
		"",
	)
}

func TestCheckArgsVariadicFail(t *testing.T) {
	checkArgs(
		t,
		NewFn(TypeList{IntType}, NoneType, IntType),
		TypeList{},
		"incorrect arity, not enough arguments (expected 1 arguments, got 0)",
	)
	checkArgs(
		t,
		NewFn(TypeList{IntType}, NoneType, IntType),
		TypeList{FloatType},
		"argument type mismatch (expected `Int` for 1-st argument, got `Float`)",
	)
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, IntType),
		TypeList{FloatType, IntType},
		"argument type mismatch (expected `Int` for variadic argument, got `Float`)",
	)
	checkArgs(
		t,
		NewFn(TypeList{}, NoneType, IntType),
		TypeList{IntType, FloatType},
		"argument type mismatch (expected `Int` for variadic argument, got `Float`)",
	)
}

func checkArgs(
	t *testing.T,
	ty *Fn,
	params TypeList,
	expectedErrStr string,
) {
	err := ty.CheckArgs(params)
	errStr := ""

	if err != nil {
		errStr = err.Error()
	}

	if expectedErrStr != errStr {
		t.Errorf(
			"failed to check args\nwant err: '%s'\ngot err: '%s'",
			expectedErrStr,
			errStr,
		)
	}
}
