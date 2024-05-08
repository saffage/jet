package types

import "testing"

func TestCheckArgs(t *testing.T) {
	params := Unit
	funcType := NewFunc(nil, params)
	idx, err := funcType.CheckArgs(params)
	checkArgs(t, idx, -1, err, "")

	params = NewTuple(I32)
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(params)
	checkArgs(t, idx, -1, err, "")
}

func TestCheckArgsFail(t *testing.T) {
	params := Unit
	args := NewTuple(I32)
	funcType := NewFunc(nil, params)
	idx, err := funcType.CheckArgs(args)
	checkArgs(t, idx, 0, err, "too many arguments (expected 0, got 1)")

	params = NewTuple(I32)
	args = Unit
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(args)
	checkArgs(t, idx, 0, err, "not enough arguments (expected 1, got 0)")

	params = NewTuple(I32)
	args = NewTuple(I32, Bool, I32)
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(args)
	checkArgs(t, idx, 1, err, "too many arguments (expected 1, got 3)")

	params = NewTuple(I32, Bool, I32)
	args = NewTuple(I32)
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(args)
	checkArgs(t, idx, 1, err, "not enough arguments (expected 3, got 1)")

	params = NewTuple(Bool)
	args = NewTuple(I32)
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(args)
	checkArgs(t, idx, 0, err, "expected 'bool' for 1st argument, got 'i32' instead")

	params = NewTuple(I32, Bool)
	args = NewTuple(I32, I32)
	funcType = NewFunc(nil, params)
	idx, err = funcType.CheckArgs(args)
	checkArgs(t, idx, 1, err, "expected 'bool' for 2nd argument, got 'i32' instead")
}

func checkArgs(
	t *testing.T,
	actualIdx int,
	expectedIdx int,
	actualErr error,
	expectedErrStr string,
) {
	actualErrStr := ""
	if actualErr != nil {
		actualErrStr = actualErr.Error()
	}

	if expectedIdx != actualIdx || expectedErrStr != actualErrStr {
		t.Errorf(
			"fail to check args\nwant idx: %d, err: '%s'\ngot  idx: %d, err: '%s'",
			expectedIdx,
			expectedErrStr,
			actualIdx,
			actualErrStr,
		)
	}
}
