package token

import (
	"encoding/json"
	"testing"

	"github.com/saffage/jet/text"
)

const id = 123

func TestSelectToken(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []Token // without EOF
	}{
		{
			name:  "simple_int",
			input: `123`,
			expected: []Token{{
				Kind: Int,
				Data: "123",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
		{
			name:  "simple_float",
			input: `0.1`,
			expected: []Token{{
				Kind: Float,
				Data: "0.1",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
		{
			name:  "simple_string",
			input: `"a"`,
			expected: []Token{{
				Kind: String,
				Data: "\"a\"",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
		{
			name:  "simple_lowercase_ident",
			input: `foo`,
			expected: []Token{{
				Kind: LowercaseIdent,
				Data: "foo",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
		{
			name:  "simple_uppercase_ident",
			input: `Foo`,
			expected: []Token{{
				Kind: UppercaseIdent,
				Data: "Foo",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
		{
			name:  "simple_ident_placeholder",
			input: `_foo`,
			expected: []Token{{
				Kind: IdentPlaceholder,
				Data: "_foo",
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 3),
				},
			}},
		},
		{
			name:  "simple_dot",
			input: `.`,
			expected: []Token{{
				Kind: Dot,
				Span: text.Span{
					From: text.PosFrom(id, 0),
				},
			}},
		},
		{
			name:  "simple_dot2",
			input: `..`,
			expected: []Token{{
				Kind: Dot2,
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 1),
				},
			}},
		},
		{
			name:  "simple_ellipsis",
			input: `...`,
			expected: []Token{{
				Kind: Ellipsis,
				Span: text.Span{
					From: text.PosFrom(id, 0),
					To:   text.PosFrom(id, 2),
				},
			}},
		},
	}

	t.Parallel()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner([]byte(tt.input), id)
			i := 0

			for tok := range scanner.Tokens() {
				if i >= len(tt.expected) {
					break
				}
				if tok != tt.expected[i] {
					want, err := json.MarshalIndent(tt.expected[i], "", "\t")
					if err != nil {
						t.Fatal(err)
					}
					have, err := json.MarshalIndent(tok, "", "\t")
					if err != nil {
						t.Fatal(err)
					}
					t.Errorf("unexpected token\nwant %s\nhave %s", want, have)
				}
				i++
			}

			if len(tt.expected) != i {
				t.Errorf(
					"lengths are not the same\nwant %d\nhave %d",
					len(tt.expected),
					i,
				)
			}
		})
	}
}
