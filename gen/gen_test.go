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