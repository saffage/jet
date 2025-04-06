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
		error        error
		input        string
		name         string
		expectedAST  string
		scannerFlags token.ScannerFlags
		parserFlags  Flags
	}{
		{
			input:       `10`,
			name:        "untyped int literal",
			expectedAST: "untyped_int_literal_ast.json",
			parserFlags: AllowTopLevelCode,
		},
		{
			input:       `"hi"`,
			name:        "untyped string literal",
			expectedAST: "untyped_string_literal_ast.json",
			parserFlags: AllowTopLevelCode,
		},
		{
			input:       `0.1`,
			name:        "untyped float literal",
			expectedAST: "untyped_float_literal_ast.json",
			parserFlags: AllowTopLevelCode,
		},
		{
			input:       `a + b * c`,
			name:        "simple a b c expr",
			expectedAST: "simple_a_b_c_expr.json",
			parserFlags: AllowTopLevelCode,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			scanner := token.NewScanner([]byte(c.input), fileID, c.scannerFlags)
			parser := New(scanner, c.parserFlags, nil)

			stmts, err := parser.ParseOrError()

			if !checkError(t, err, c.error) {
				return
			}

			actual, err := json.MarshalIndent(stmts, "", "\t")

			if err != nil {
				t.Error("unexpected JSON marshal error:", err)
				return
			}

			if c.expectedAST != "" {
				filename := filepath.Join("testdata", c.expectedAST)
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
			t.Errorf("parsing failed with unexpected error: '%s'", got.Error())
			report.Render(got)
			return false
		}
	} else if got == nil {
		t.Errorf("expected an error: '%s', got nothing", want.Error())
		report.Render(want)
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
