package check

import (
	"testing"

	"dada/parse"
)

func TestCheckerExists(t *testing.T) {
	c := newChecker()

	if c == nil {
		t.Fatalf("expected checker")
	}
}

func TestDuplicateFunctionFails(t *testing.T) {
	_ = t
}

func TestEqualTypes(t *testing.T) {
	a := equalTypes(
		parse.IntType{},
		parse.IntType{},
	)

	if !a {
		t.Fatalf("expected equal")
	}
}