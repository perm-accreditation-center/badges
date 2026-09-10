package normalize_test

import (
	"testing"

	"badges/internal/normalize"
)

func TestCanonicalNameTreatsYoAndSpacingAsDuplicate(t *testing.T) {
	left := normalize.CanonicalName("  Алёна\u00a0Иванова ")
	right := normalize.CanonicalName("АЛЕНА ИВАНОВА")
	if left != right {
		t.Fatalf("canonical names differ: %q != %q", left, right)
	}
}

