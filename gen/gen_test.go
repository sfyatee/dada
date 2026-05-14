package gen

import (
	"strings"
	"testing"

	"dada/parse"
)

func parseProgram(t *testing.T, input string) *parse.Program {
	t.Helper()

	p := parse.NewParser(strings.NewReader(input))
	return p.ParseProgram()
}

func generate(t *testing.T, input string) string {
	t.Helper()

	prog := parseProgram(t, input)

	out, err := GenerateString(prog)
	if err != nil {
		t.Fatalf("GenerateString failed: %v", err)
	}

	return out
}

func TestGenerateSimpleFunction(t *testing.T) {
	js := generate(t, `
(def add () ((Int x) (Int y)) Int
	(+ x y))

(call add 1 2)
`)

	if !strings.Contains(js, "function add(x, y)") {
		t.Fatalf("missing generated function:\n%s", js)
	}

	if !strings.Contains(js, "return (x + y);") {
		t.Fatalf("missing generated return:\n%s", js)
	}

	if !strings.Contains(js, "add(1, 2);") {
		t.Fatalf("missing generated call:\n%s", js)
	}
}

func TestGenerateNilProgram(t *testing.T) {
	_, err := GenerateString(nil)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateLambda(t *testing.T) {
	js := generate(t, `
(callhof (=> ((Int x)) (+ x 1)) 5)
`)

	if !strings.Contains(js, "((x) => (x + 1))(5);") {
		t.Fatalf("bad lambda generation:\n%s", js)
	}
}

func TestGenerateBlock(t *testing.T) {
	js := generate(t, `
(block
	(val x 1)
	(var Int y 2)
	(= y (+ x y))
	y)
`)

	if !strings.Contains(js, "const x = 1;") {
		t.Fatalf("missing const stmt:\n%s", js)
	}

	if !strings.Contains(js, "let y = 2;") {
		t.Fatalf("missing let stmt:\n%s", js)
	}

	if !strings.Contains(js, "y = (x + y);") {
		t.Fatalf("missing assignment:\n%s", js)
	}
}
