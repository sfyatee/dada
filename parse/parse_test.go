package parse

import (
	"strings"
	"testing"
)

func parseProgram(t *testing.T, input string) *Program {
	t.Helper()

	p := NewParser(strings.NewReader(input))
	prog := p.ParseProgram()
	if prog == nil {
		t.Fatalf("ParseProgram returned nil for %q", input)
	}
	return prog
}

func parseOne(t *testing.T, input string) Expr {
	t.Helper()

	prog := parseProgram(t, input)
	if len(prog.AlgDefs) != 0 || len(prog.FuncDefs) != 0 {
		t.Fatalf("expected only one top-level expression for %q", input)
	}
	if prog.Expr == nil {
		t.Fatalf("ParseProgram returned nil expression for %q", input)
	}
	return prog.Expr
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

	v, ok := expr.(VarExpr)
	if !ok {
		t.Fatalf("got %T, want VarExpr", expr)
	}
	if got, want := v.Name, "x"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseSimpleCallExpr(t *testing.T) {
	expr := parseOne(t, "(call f 1)")

	call, ok := expr.(CallExpr)
	if !ok {
		t.Fatalf("got %T, want CallExpr", expr)
	}
	if got, want := call.Name, "f"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := len(call.Args), 1; got != want {
		t.Fatalf("got %d args, want %d", got, want)
	}
}

func TestParseNestedOpExpr(t *testing.T) {
	expr := parseOne(t, "(+ 1 (* 2 3))")

	op, ok := expr.(OpExpr)
	if !ok {
		t.Fatalf("got %T, want OpExpr", expr)
	}
	if got, want := op.Op, "+"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	_, ok = op.Left.(IntExpr)
	if !ok {
		t.Fatalf("left operand: got %T, want IntExpr", op.Left)
	}

	right, ok := op.Right.(OpExpr)
	if !ok {
		t.Fatalf("right operand: got %T, want OpExpr", op.Right)
	}
	if got, want := right.Op, "*"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseRejectsMissingClosingParen(t *testing.T) {
	mustPanic(t, func() {
		_ = parseProgram(t, "(call f 1")
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

	if got, want := len(prog.AlgDefs), 1; got != want {
		t.Fatalf("got %d algdefs, want %d", got, want)
	}

	if got, want := len(prog.FuncDefs), 1; got != want {
		t.Fatalf("got %d funcdefs, want %d", got, want)
	}

	if prog.Expr == nil {
		t.Fatalf("expected final expression, got nil")
	}

	if got, want := prog.AlgDefs[0].Name, "List"; got != want {
		t.Fatalf("algdef name: got %q, want %q", got, want)
	}

	if got, want := prog.FuncDefs[0].Name, "id"; got != want {
		t.Fatalf("funcdef name: got %q, want %q", got, want)
	}

	call, ok := prog.Expr.(CallExpr)
	if !ok {
		t.Fatalf("final expression: got %T, want CallExpr", prog.Expr)
	}
	if got, want := call.Name, "id"; got != want {
		t.Fatalf("call name: got %q, want %q", got, want)
	}
}

func TestParseBlockWithValVarAndAssign(t *testing.T) {
	expr := parseOne(t, "(block (val x 1) (var Int y 2) (= y 3) y)")

	block, ok := expr.(BlockExpr)
	if !ok {
		t.Fatalf("got %T, want BlockExpr", expr)
	}

	if got, want := len(block.Stmts), 3; got != want {
		t.Fatalf("got %d statements, want %d", got, want)
	}

	if _, ok := block.Stmts[0].(ValStmt); !ok {
		t.Fatalf("first stmt: got %T, want ValStmt", block.Stmts[0])
	}

	if _, ok := block.Stmts[1].(VarStmt); !ok {
		t.Fatalf("second stmt: got %T, want VarStmt", block.Stmts[1])
	}

	if _, ok := block.Stmts[2].(AssignStmt); !ok {
		t.Fatalf("third stmt: got %T, want AssignStmt", block.Stmts[2])
	}

	v, ok := block.Expr.(VarExpr)
	if !ok {
		t.Fatalf("block expr: got %T, want VarExpr", block.Expr)
	}
	if got, want := v.Name, "y"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseMatchWithWildcardAndConstructorPatterns(t *testing.T) {
	expr := parseOne(t, "(match xs (case _ 0) (case (cons lcons head tail) head))")

	match, ok := expr.(MatchExpr)
	if !ok {
		t.Fatalf("got %T, want MatchExpr", expr)
	}

	if got, want := len(match.Cases), 2; got != want {
		t.Fatalf("got %d cases, want %d", got, want)
	}

	if _, ok := match.Cases[0].Pattern.(WildcardPattern); !ok {
		t.Fatalf("first case pattern: got %T, want WildcardPattern", match.Cases[0].Pattern)
	}

	consPat, ok := match.Cases[1].Pattern.(ConsPattern)
	if !ok {
		t.Fatalf("second case pattern: got %T, want ConsPattern", match.Cases[1].Pattern)
	}

	if got, want := consPat.Name, "lcons"; got != want {
		t.Fatalf("constructor pattern name: got %q, want %q", got, want)
	}

	if got, want := len(consPat.Patterns), 2; got != want {
		t.Fatalf("got %d nested patterns, want %d", got, want)
	}

	if _, ok := consPat.Patterns[0].(VarPattern); !ok {
		t.Fatalf("first nested pattern: got %T, want VarPattern", consPat.Patterns[0])
	}

	if _, ok := consPat.Patterns[1].(VarPattern); !ok {
		t.Fatalf("second nested pattern: got %T, want VarPattern", consPat.Patterns[1])
	}
}

func TestParseLambdaAndCallHOF(t *testing.T) {
	input := `
(def applyTwice () (((=> (Int) Int) f) (Int x)) Int
  (callhof f (callhof f x)))
(callhof (=> ((Int y)) (+ y 1)) 5)
`
	prog := parseProgram(t, input)

	if got, want := len(prog.FuncDefs), 1; got != want {
		t.Fatalf("got %d funcdefs, want %d", got, want)
	}

	fd := prog.FuncDefs[0]
	if got, want := fd.Name, "applyTwice"; got != want {
		t.Fatalf("funcdef name: got %q, want %q", got, want)
	}

	if got, want := len(fd.Params), 2; got != want {
		t.Fatalf("got %d params, want %d", got, want)
	}

	if _, ok := fd.Params[0].Type.(FuncType); !ok {
		t.Fatalf("first param type: got %T, want FuncType", fd.Params[0].Type)
	}

	call, ok := prog.Expr.(CallHOFExpr)
	if !ok {
		t.Fatalf("final expression: got %T, want CallHOFExpr", prog.Expr)
	}

	lambda, ok := call.Fn.(LambdaExpr)
	if !ok {
		t.Fatalf("callhof fn: got %T, want LambdaExpr", call.Fn)
	}

	if got, want := len(lambda.Params), 1; got != want {
		t.Fatalf("lambda params: got %d, want %d", got, want)
	}

	if got, want := len(call.Args), 1; got != want {
		t.Fatalf("callhof args: got %d, want %d", got, want)
	}
}
