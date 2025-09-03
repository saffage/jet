package parser

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

const fileID = text.FileID(123)

func TestExprs(t *testing.T) {
	testCases := []struct {
		error       error
		input       string
		name        string
		expectedAST string
		opts        Options
	}{
		{
			input:       `10`,
			name:        "untyped int literal",
			expectedAST: "untyped_int_literal_ast.json",
			opts:        Options{ParserFlags: AllowTopLevelCode},
		},
		{
			input:       `"hi"`,
			name:        "untyped string literal",
			expectedAST: "untyped_string_literal_ast.json",
			opts:        Options{ParserFlags: AllowTopLevelCode},
		},
		{
			input:       `0.1`,
			name:        "untyped float literal",
			expectedAST: "untyped_float_literal_ast.json",
			opts:        Options{ParserFlags: AllowTopLevelCode},
		},
		{
			input:       `a + b * c`,
			name:        "simple a b c expr",
			expectedAST: "simple_a_b_c_expr.json",
			opts:        Options{ParserFlags: AllowTopLevelCode},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			scanner := token.NewScanner([]byte(tt.input), fileID, tt.opts.ScannerOptions)
			parser := New(scanner, tt.opts)

			stmts, err := parser.ParseOrError()

			if !checkError(t, err, tt.error) {
				return
			}

			actual, err := json.MarshalIndent(stmts, "", "\t")

			if err != nil {
				t.Error("unexpected JSON marshal error:", err)
				return
			}

			if tt.expectedAST != "" {
				filename := filepath.Join("testdata", tt.expectedAST)
				expect, err := os.ReadFile(filename)

				if err != nil {
					t.Errorf("unexpected error while reading file '%s': %s", filename, err)
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
				t.Logf("no AST was expected\ngot %s", string(actual))
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
			t.Errorf("parsing failed with unexpected error: '%s'", got)
			report.Render(got)
			return false
		}
	} else if got == nil {
		t.Errorf("expected an error: '%s', got nothing", want)
		report.Render(want)
		return false
	}

	if got.Error() != want.Error() {
		t.Errorf(
			"unexpected error:\nexpect: '%s'\nactual: '%s'",
			got,
			want,
		)
		return false
	}

	return true
}

func JSONBytesEqual(a, b []byte) (bool, error) {
	a = bytes.TrimSpace(a)
	b = bytes.TrimSpace(b)

	var jsonA, jsonB any

	if err := json.Unmarshal(a, &jsonA); err != nil {
		return false, err
	}

	if err := json.Unmarshal(b, &jsonB); err != nil {
		return false, err
	}

	return reflect.DeepEqual(jsonB, jsonA), nil
}
