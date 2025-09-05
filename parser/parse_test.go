package parser_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/saffage/jet/config"
	"github.com/saffage/jet/parser"
	"github.com/saffage/jet/report"
	"github.com/saffage/jet/token"
)

func TestExprs(t *testing.T) {
	testCases := []struct {
		error       error
		input       string
		name        string
		expectedAST string
		opts        parser.Options
	}{
		{
			input:       `10`,
			name:        "untyped int literal",
			expectedAST: "untyped_int_literal_ast.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input:       `"hi"`,
			name:        "untyped string literal",
			expectedAST: "untyped_string_literal_ast.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input:       `0.1`,
			name:        "untyped float literal",
			expectedAST: "untyped_float_literal_ast.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input:       `a + b * c`,
			name:        "simple a b c expr",
			expectedAST: "simple_a_b_c_expr.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input:       `let x = 10`,
			name:        "simple let binding",
			expectedAST: "simple_let_binding.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input:       `let _ = 10`,
			name:        "simple let discard",
			expectedAST: "simple_let_discard.json",
			opts:        parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
		{
			input: `let a, b = 10, 20`,
			name:  "simple let binding tuple",
			// expectedAST: "simple_let_binding_tuple.json",
			opts: parser.Options{ParserFlags: parser.AllowTopLevelCode},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			file, err := config.NewFile(fmt.Sprintf("parser_test(%s).jet", tt.name), []byte(tt.input))
			if err != nil {
				t.Fatal(err)
				return
			}
			defer config.RemoveFile(file.ID)

			scanner := token.NewScannerFromFile(file, tt.opts.ScannerOptions)
			parser := parser.New(scanner, tt.opts)

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

				if equal, err := JSONEqual(actual, expect); err != nil {
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

func JSONEqual(a, b []byte) (bool, error) {
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
