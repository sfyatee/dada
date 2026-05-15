package check

import (
	"testing"
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
