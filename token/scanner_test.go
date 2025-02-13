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
	i := 0

	for tok := range scanner.Tokens() {
		if i >= len(expectedKinds) {
			break
		}
		if tok.Kind != expectedKinds[i] {
			t.Errorf(
				"unexpected token; want %s, have %s",
				expectedKinds[i].String(),
				tok.Kind.String(),
			)
		}
		i++
	}

	if len(expectedKinds) != i {
		t.Errorf(
			"lengths are not the same; want %d, have %d",
			len(expectedKinds),
			i,
		)
	}
}
