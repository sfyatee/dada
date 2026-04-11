package parse

import (
	"strings"
	"testing"
)

func parseProgram(t *testing.T, input string) []*AST {
	t.Helper()

	p := NewParser(strings.NewReader(input))
	prog := p.Parse()
	if prog == nil {
		t.Fatalf("Parse returned nil for %q", input)
	}
	return prog
}

func parseOne(t *testing.T, input string) *AST {
	t.Helper()

	prog := parseProgram(t, input)
	if len(prog) != 1 {
		t.Fatalf("Parse returned %d expressions for %q, want 1", len(prog), input)
	}
	return prog[0]
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()

	f()
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

func TestParseProgramMultipleExpressions(t *testing.T) {
	prog := parseProgram(t, "x y")

	if got, want := len(prog), 2; got != want {
		t.Fatalf("got %d expressions, want %d", got, want)
	}

	if got, want := prog[0].String(), "x"; got != want {
		t.Fatalf("first expression: got %q, want %q", got, want)
	}

	if got, want := prog[1].String(), "y"; got != want {
		t.Fatalf("second expression: got %q, want %q", got, want)
	}
}

func TestParseRejectsMissingClosingParen(t *testing.T) {
	mustPanic(t, func() {
		_ = parseProgram(t, "(call f 1")
	})
}

func TestSExprStringWithConfig(t *testing.T) {
	Config(true)
	defer Config(false)

	expr := parseOne(t, "(+ 1 2)")

	if got, want := expr.String(), "(+ . (1 . (2 . nil)))"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestListParsesSingleExpression(t *testing.T) {
	p := NewParser(strings.NewReader("(call f 1)"))
	expr := p.List()

	if expr == nil {
		t.Fatalf("List returned nil")
	}

	if got, want := expr.String(), "(call f 1)"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSkipSpace(t *testing.T) {
	p := NewParser(strings.NewReader("   x"))

	r := p.SkipSpace()
	if got, want := r, 'x'; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	expr := p.List()
	if expr == nil {
		t.Fatalf("List returned nil after SkipSpace")
	}
	if got, want := expr.String(), "x"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSkipToEndOfLine(t *testing.T) {
	p := NewParser(strings.NewReader("ignore this line\nx"))

	p.SkipToEndOfLine()
	expr := p.List()

	if expr == nil {
		t.Fatalf("List returned nil after SkipToEndOfLine")
	}
	if got, want := expr.String(), "x"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseRejectsDotAtStartOfList(t *testing.T) {
	mustPanic(t, func() {
		_ = parseProgram(t, "(. x)")
	})
}

func TestParseRejectsBadDottedListTail(t *testing.T) {
	mustPanic(t, func() {
		_ = parseProgram(t, "(a . b c)")
	})
}

func TestParseProgramWithAlgDefFuncDefAndExpr(t *testing.T) {
	input := `
(algdef List (A)
  (lcons A (alg List A))
  (lnil))
(def id (A) ((A x)) A x)
(call id 1)
`
	prog := parseProgram(t, input)

	if got, want := len(prog), 3; got != want {
		t.Fatalf("got %d top-level expressions, want %d", got, want)
	}

	if got, want := prog[0].String(), "(algdef List (A) (lcons A (alg List A)) (lnil))"; got != want {
		t.Fatalf("first expression: got %q, want %q", got, want)
	}

	if got, want := prog[1].String(), "(def id (A) ((A x)) A x)"; got != want {
		t.Fatalf("second expression: got %q, want %q", got, want)
	}

	if got, want := prog[2].String(), "(call id 1)"; got != want {
		t.Fatalf("third expression: got %q, want %q", got, want)
	}
}
