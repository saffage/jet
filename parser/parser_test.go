package parser

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
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
	name         string
	input        string
	expectedJSON string
	errors       []error
	scannerFlags scanner.Flags
	parserFlags  Flags
}

func TestExprs(t *testing.T) {
	t.Cleanup(cleanup)

	cases := []testCase{
		{
			name:         "untyped integer literal",
			input:        "10",
			expectedJSON: "untyped_integer_literal_ast.json",
		},
		{
			name:         "untyped string literal",
			input:        "'hi'",
			expectedJSON: "untyped_string_literal_ast.json",
		},
	}

	for _, c := range cases {
		tokens, errs := scanner.Scan(([]byte)(c.input), 1, c.scannerFlags)

		if len(errs) != 0 {
			t.Fatalf("unexpected scanner errors: %v", errs)
		}

		t.Run(c.name, func(t *testing.T) {
			ast, errs := Parse(cfg, tokens, c.parserFlags)

			if errs != nil {
				if c.errors != nil {
					for i, err := range errs {
						if i < len(c.errors) && err.Error() == c.errors[i].Error() {
							continue
						}

						t.Errorf("parsing failed with unexpected error: '%s'", err.Error())
					}
				}

				if c.errors != nil {
					cmpResult := slices.CompareFunc(errs, c.errors, func(got, want error) int {
						return strings.Compare(got.Error(), want.Error())
					})

					if cmpResult != 0 {
						t.Errorf("unexpected error: ")
					}
				}

				t.Errorf("parsing failed with unexpected errors: %v", errs)
			}

			if c.expectedJSON != "" {
				filename := "./testdata/" + c.expectedJSON
				want, err := os.ReadFile(filename)
				if err != nil {
					t.Errorf("unexpected error while reading file '%s': %s", filename, err)
				}

				got, err := json.MarshalIndent(ast, "", "    ")
				if err != nil {
					t.Error("unexpected JSON marshal error:", err)
				}

				if slices.Compare(got, want) != 0 {
					t.Errorf("unexpected AST was parsed\nwant %s\ngot  %s", string(want), string(got))
				}
			} else {
				encoded, err := json.MarshalIndent(ast, "", "    ")
				if err != nil {
					t.Error("unexpected JSON marshal error:", err)
				}

				t.Logf("no AST was expected\ngot %s", string(encoded))
			}
		})
	}
}
