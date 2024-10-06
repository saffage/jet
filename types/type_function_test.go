package types

import "testing"

func TestCheckArgs(t *testing.T) {
	checkArgs(
		t,
		NewFunction(Params{NoneType}, NoneType, nil),
		Params{NoneType},
		-1,
		"",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType}, NoneType, nil),
		Params{IntType},
		-1,
		"",
	)
}

func TestCheckArgsFail(t *testing.T) {
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, nil),
		Params{IntType},
		0,
		"incorrect arity, too many arguments (expected 0 arguments, got 1)",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType}, NoneType, nil),
		Params{},
		0,
		"incorrect arity, not enough arguments (expected 1 arguments, got 0)",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType}, NoneType, nil),
		Params{IntType, BoolType, IntType},
		1,
		"incorrect arity, too many arguments (expected 1 arguments, got 3)",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType, BoolType, IntType}, NoneType, nil),
		Params{IntType},
		1,
		"incorrect arity, not enough arguments (expected 3 arguments, got 1)",
	)
	checkArgs(
		t,
		NewFunction(Params{BoolType}, NoneType, nil),
		Params{IntType},
		0,
		"argument type mismatch (expected `Bool` for 1-st argument, got `Int`)",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType, BoolType}, NoneType, nil),
		Params{IntType, IntType},
		1,
		"argument type mismatch (expected `Bool` for 2-nd argument, got `Int`)",
	)
}

func TestCheckArgsVariadic(t *testing.T) {
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, IntType),
		Params{},
		-1,
		"",
	)
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, IntType),
		Params{IntType},
		-1,
		"",
	)
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, IntType),
		Params{IntType, IntType},
		-1,
		"",
	)
}

func TestCheckArgsVariadicFail(t *testing.T) {
	checkArgs(
		t,
		NewFunction(Params{IntType}, NoneType, IntType),
		Params{},
		0,
		"incorrect arity, not enough arguments (expected 1 arguments, got 0)",
	)
	checkArgs(
		t,
		NewFunction(Params{IntType}, NoneType, IntType),
		Params{FloatType},
		0,
		"argument type mismatch (expected `Int` for 1-st argument, got `Float`)",
	)
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, IntType),
		Params{FloatType, IntType},
		0,
		"argument type mismatch (expected `Int` for variadic argument, got `Float`)",
	)
	checkArgs(
		t,
		NewFunction(Params{}, NoneType, IntType),
		Params{IntType, FloatType},
		1,
		"argument type mismatch (expected `Int` for variadic argument, got `Float`)",
	)
}

func checkArgs(
	t *testing.T,
	ty *Function,
	params Params,
	expectedIdx int,
	expectedErrStr string,
) {
	idx, err := ty.CheckArgs(params)
	errStr := ""

	if err != nil {
		errStr = err.Error()
	}

	if expectedIdx != idx || expectedErrStr != errStr {
		t.Errorf(
			"failed to check args\nwant idx: %d, err: '%s'\ngot  idx: %d, err: '%s'",
			expectedIdx,
			expectedErrStr,
			idx,
			errStr,
		)
	}
}
