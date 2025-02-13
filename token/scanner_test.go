package token

import "testing"

func TestSelectToken(t *testing.T) {
	testTokenKinds(t, ".", Dot, EOF)
	testTokenKinds(t, "..", Dot2, EOF)
	testTokenKinds(t, "...", Ellipsis, EOF)
	testTokenKinds(t, "..<", Dot2Less, EOF)
}

func testTokenKinds(t *testing.T, input string, expectedKinds ...Kind) {
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
