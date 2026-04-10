package parse

import (
	"strings"
	"testing"
)

func parseOne(t *testing.T, input string) *AST {
	t.Helper()

	p := NewParser(strings.NewReader(input))
	expr := p.Parse()
	if expr == nil {
		t.Fatalf("Parse returned nil for %q", input)
	}
	return expr
}

func TestParseAtomString(t *testing.T) {
	expr := parseOne(t, "x")
	if got, want := expr.String(), "x"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseSimpleListString(t *testing.T) {
	expr := parseOne(t, "(call f 1)")
	if got, want := expr.String(), "(call f 1)"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseNestedListString(t *testing.T) {
	expr := parseOne(t, "(+ 1 (* 2 3))")
	if got, want := expr.String(), "(+ 1 (* 2 3))"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseEmptyListString(t *testing.T) {
	expr := parseOne(t, "()")
	if got, want := expr.String(), "()"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

