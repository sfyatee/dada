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
