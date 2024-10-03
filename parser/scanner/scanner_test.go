package scanner

import (
	"testing"

	"github.com/saffage/jet/parser/token"
)

func testTokenKinds(t *testing.T, input string, expectedKinds ...token.Kind) {
	scanner := New([]byte(input), 0, NoFlags)
	toks := scanner.AllTokens()

	if len(expectedKinds) != len(toks) {
		t.Errorf("lengths are not the same; want %d, have %d", len(expectedKinds), len(toks))
	}

	for i := range toks {
		if toks[i].Kind != expectedKinds[i] {
			t.Errorf("unexpected token; want %s, have %s", expectedKinds[i].String(), toks[i].Kind.String())
		}
	}
}

func TestSelectToken(t *testing.T) {
	testTokenKinds(t, ".", token.Dot, token.EOF)
	testTokenKinds(t, "..", token.Dot2, token.EOF)
}
