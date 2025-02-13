package parser

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/saffage/jet/report"
	"github.com/saffage/jet/text"
	"github.com/saffage/jet/token"
)

const fileID = text.FileID(123)

func TestExprs(t *testing.T) {
	type testCase struct {
		error        error
		input        string
		name         string
		expectedAST  string
		scannerFlags token.ScannerFlags
		parserFlags  Flags
	}

	testCases := []testCase{
		{
			input:       `10`,
			name:        "untyped integer literal",
			expectedAST: "untyped_integer_literal_ast.yml",
			parserFlags: AllowTopLevelCode,
		},
		{
			input:       `"hi"`,
			name:        "untyped string literal",
			expectedAST: "untyped_string_literal_ast.yml",
			parserFlags: AllowTopLevelCode,
		},
		{
			input:       `0.1`,
			name:        "untyped float literal",
			expectedAST: "untyped_float_literal_ast.yml",
			parserFlags: AllowTopLevelCode,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			scanner := token.New([]byte(c.input), fileID, c.scannerFlags)
			parser := NewFrom(scanner, c.parserFlags)

			stmts, err := parser.Parse()

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

				actual, err := json.Marshal(stmts)

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
				encoded, err := json.Marshal(stmts)

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
			report.Report(got)
			t.Errorf("parsing failed with unexpected error: '%s'", got.Error())
			return false
		}
	} else if got == nil {
		report.Report(want)
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

	if err := json.Unmarshal(a, &jsonA); err != nil {
		return false, err
	}

	if err := json.Unmarshal(b, &jsonB); err != nil {
		return false, err
	}

	return reflect.DeepEqual(jsonB, jsonA), nil
}
