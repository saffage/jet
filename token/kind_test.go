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

	expectedKinds := append(
		PunctuationKinds(),
		append(
			OperatorKinds(),
			KeywordKinds()...,
		)...,
	)
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
