package token

import (
	"slices"
	"testing"
)

func TestKindRepr(t *testing.T) {
	kinds := make([]Kind, 0, len(representableKinds))

	for kind := range representableKinds {
		kinds = append(kinds, kind)
	}

	expectedKinds := append(PunctuationKinds(), append(OperatorKinds(), KeywordKinds()...)...)
	missingKinds, extraKinds := []Kind{}, []Kind{}

	for _, expectedKind := range expectedKinds {
		if !slices.Contains(kinds, expectedKind) {
			missingKinds = append(missingKinds, expectedKind)
		}
	}

	for _, kind := range kinds {
		if !slices.Contains(expectedKinds, kind) {
			extraKinds = append(extraKinds, kind)
		}
	}

	if len(missingKinds) > 0 {
		t.Errorf("\nmissing token representations: %v", missingKinds)
	}

	if len(extraKinds) > 0 {
		t.Errorf("\nextra token representations: %v", extraKinds)
	}
}

func getKinds(begin, end int) []Kind {
	kinds := make([]Kind, 0, end-begin+1)

	for kind := begin; kind <= end; kind++ {
		kinds = append(kinds, Kind(kind))
	}

	return kinds
}

func SpecialKinds() []Kind     { return getKinds(int(_special_begin), int(_special_end)) }
func PrimaryKinds() []Kind     { return getKinds(int(_primary_begin), int(_primary_end)) }
func PunctuationKinds() []Kind { return getKinds(int(_punctuation_begin), int(_punctuation_end)) }
func OperatorKinds() []Kind    { return getKinds(int(_operator_begin), int(_operator_end)) }
func KeywordKinds() []Kind     { return getKinds(int(_keywords_begin), int(_keywords_end)) }
func AllKinds() []Kind         { return getKinds(0, int(_kinds_last)) }
