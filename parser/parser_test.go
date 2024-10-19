package parser

import (
	"os"
	"reflect"
	"testing"

	"github.com/saffage/jet/ast"
	"github.com/saffage/jet/parser/scanner"
	"github.com/saffage/jet/parser/token"
	"github.com/saffage/jet/report"
	"gopkg.in/yaml.v3"
)

func TestMatchSequence(t *testing.T) {
	tokens := []token.Token{
		{Kind: token.Name},
		{Kind: token.Colon},
		{Kind: token.Name},
		{Kind: token.Colon},
		{Kind: token.Name},
		{Kind: token.Colon},
		{Kind: token.EOF},
	}
	p, _ := New(tokens, DefaultFlags)
	kinds := []token.Kind{token.Name, token.Colon}
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}
	p.current += 2
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}
	p.current += 2
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}

	p.current = 0
	kinds = []token.Kind{token.Name}
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}
	p.current += 2
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}
	p.current += 2
	if !p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return true, got false", kinds)
	}

	p.current = 0
	kinds = []token.Kind{token.Name, token.Name}
	if p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return false, got true", kinds)
	}
	p.current += 2
	if p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return false, got true", kinds)
	}
	p.current += 2
	if p.matchSequence(kinds...) {
		t.Errorf("expected `matchSequence(%v)` to return false, got true", kinds)
	}
}

func TestExprs(t *testing.T) {
	type testCase struct {
		error        error
		input        string
		name         string
		expectedAST  string
		scannerFlags scanner.Flags
		parserFlags  Flags
		isExpr       bool
	}

	testCases := []testCase{
		{
			input:       `10`,
			name:        "untyped integer literal",
			expectedAST: "untyped_integer_literal_ast.yml",
			isExpr:      true,
		},
		{
			input:       `"hi"`,
			name:        "untyped string literal",
			expectedAST: "untyped_string_literal_ast.yml",
			isExpr:      true,
		},
		{
			input:       `0.1`,
			name:        "untyped float literal",
			expectedAST: "untyped_float_literal_ast.yml",
			isExpr:      true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			tokens := scanner.MustScan(([]byte)(c.input), 1, c.scannerFlags)

			var stmts ast.Stmts
			var err error

			if c.isExpr {
				var node ast.Node
				node, err = ParseExpr(tokens, c.parserFlags)
				stmts = ast.Stmts{node}
			} else {
				stmts, err = Parse(tokens, c.parserFlags)
			}

			if !checkError(t, err, c.error) {
				return
			}

			if c.expectedAST != "" {
				filename := "./testdata/" + c.expectedAST
				expect, err := os.ReadFile(filename)

				if err != nil {
					t.Errorf("unexpected error while reading file '%s': %s", filename, err)
					return
				}

				actual, err := yaml.Marshal(stmts)

				if err != nil {
					t.Error("unexpected JSON marshal error:", err)
					return
				}

				if equal, err := JSONBytesEqual(actual, expect); err != nil {
					t.Error(err)
				} else if !equal {
					t.Errorf(
						"invalid AST was parsed\nexpect %s\nactual %s",
						string(expect),
						string(actual),
					)
				}
			} else {
				encoded, err := yaml.Marshal(stmts)

				if err != nil {
					t.Error("unexpected JSON marshal error:", err)
					return
				}

				t.Logf("no AST was expected\ngot %s", string(encoded))
			}
		})
	}
}

func checkError(t *testing.T, got, want error) bool {
	if want == nil && got == nil {
		return true
	}

	if want == nil {
		if got != nil {
			report.Error(got)
			t.Errorf("parsing failed with unexpected error: '%s'", got.Error())
			return false
		}
	} else if got == nil {
		report.Error(want)
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

	return true
}

func JSONBytesEqual(a, b []byte) (bool, error) {
	var jsonA, jsonB any

	if err := yaml.Unmarshal(a, &jsonA); err != nil {
		return false, err
	}

	if err := yaml.Unmarshal(b, &jsonB); err != nil {
		return false, err
	}

	return reflect.DeepEqual(jsonB, jsonA), nil
}
