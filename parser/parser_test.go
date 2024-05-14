package parser

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/scanner"
)

var cfg *config.Config

// TODO more tests

func TestMain(m *testing.M) {
	cleanup()
	code := m.Run()
	os.Exit(code)
}

func cleanup() {
	println("cleanup")
	cfg = config.New()
}

type testCase struct {
	input        string
	name         string
	expectedJSON string
	errors       []error
	scannerFlags scanner.Flags
	parserFlags  Flags
}

func TestExprs(t *testing.T) {
	t.Cleanup(cleanup)

	cases := []testCase{
		{
			input:        `10`,
			name:         "untyped integer literal",
			expectedJSON: "untyped_integer_literal_ast.json",
		},
		{
			input:        `'hi'`,
			name:         "untyped string literal",
			expectedJSON: "untyped_string_literal_ast.json",
		},
		{
			input:        `()`,
			name:         "empty parentheses",
			expectedJSON: "empty_parentheses.json",
		},
		{
			input:        `(,)`,
			name:         "empty tuple constructor",
			expectedJSON: "empty_tuple_constructor.json",
			errors:       []error{errors.New("expected operand, found ','")},
		},
		{
			input:        `(1)`,
			name:         "parentheses with 1 expr",
			expectedJSON: "parentheses_with_1_expr.json",
		},
		{
			input:        `(1,)`,
			name:         "tuple constructor",
			expectedJSON: "tuple_constructor.json",
		},
	}

	for _, c := range cases {
		test(t, c)
	}
}

func test(t *testing.T, c testCase) {
	tokens, errs := scanner.Scan(([]byte)(c.input), 1, c.scannerFlags)

	if len(errs) != 0 {
		t.Fatalf("unexpected scanner errors: %v", errs)
	}

	t.Run(c.name, func(t *testing.T) {
		ast, errs := Parse(cfg, tokens, c.parserFlags)
		if !checkErrors(t, errs, c.errors) {
			return
		}

		if c.expectedJSON != "" {
			filename := "./testdata/" + c.expectedJSON
			expect, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("unexpected error while reading file '%s': %s", filename, err)
				return
			}

			actual, err := json.MarshalIndent(ast, "", "    ")
			if err != nil {
				t.Error("unexpected JSON marshal error:", err)
				return
			}

			if slices.Compare(actual, expect) != 0 {
				t.Errorf("unexpected AST was parsed\nexpect %s\nactual %s", string(expect), string(actual))
				return
			}
		} else {
			encoded, err := json.MarshalIndent(ast, "", "    ")
			if err != nil {
				t.Error("unexpected JSON marshal error:", err)
				return
			}

			t.Logf("no AST was expected\ngot %s", string(encoded))
		}
	})
}

func checkErrors(t *testing.T, gotErrors, wantErrors []error) bool {
	maxIndex := max(len(gotErrors), len(wantErrors))

	for i := 0; i < maxIndex; i++ {
		var got, want error

		if i < len(gotErrors) {
			got = gotErrors[i]
		}

		if i < len(wantErrors) {
			want = wantErrors[i]
		}

		if got == nil && want == nil {
			panic("unreachable")
		}

		if want == nil {
			t.Errorf("parsing failed with unexpected error: '%s'", got.Error())
			return false
		} else if got == nil {
			t.Errorf("expected an error: '%s', got nothing", want.Error())
			return false
		}

		if got.Error() != want.Error() {
			t.Errorf(
				"unexpected error:\nexpect: '%s'\nactual: '%s'",
				got.Error(),
				want.Error(),
			)
			return false
		}
	}

	return true
}
