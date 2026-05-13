// Code generator was chosen as the component to be completed
// for the required final assignment in COMP 430. If another
// component is completed or started, we ask that it be treated
// as the extra credit opportunity.

// Assumptions:
//   The code generator receives a syntactically and
//      semantically valid parse.Program, which is defined
//      in ast.go.
//   Validity of this program will NOT be tested, only trusted.
//   Identifiers are also valid JS identifiers

// Notes:
//   The target language proposed is JavaScript. This implem
//   -entation moves forward with this target language.

package gen

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"dada/parse"
)

// Generate writes JavaScript for a valid D.A.D.A program.
// Typecheckingand other checks are assumed to have
// happened before this package is called.
func Generate(w io.Writer, program *parse.Program) error {
	if program == nil {
		return fmt.Errorf("cannot generate code with nil program")
	}

	g := newGenerator()
	if err := g.program(program); err != nil {
		return err
	}

	// writes js output to provided writer
	_, err := io.WriteString(w, g.out.String())
	return err
}

// GenerateString is a convenience wrapper for tests and command-line callers.
func GenerateString(program *parse.Program) (string, error) {
	var b strings.Builder
	if err := Generate(&b, program); err != nil {
		return "", err
	}
	return b.String(), nil
}

type generator struct {
	out    strings.Builder
	tempID int
}

func newGenerator() *generator {
	return &generator{}
}

func (g *generator) program(program *parse.Program) error {

	if len(program.AlgDefs) > 0 {
		g.blank()

		// iterate through algdefs
		for _, alg := range program.AlgDefs {
			if alg == nil {
				return fmt.Errorf("nil algebraic data type definition")
			}

			// iterate through consdefs
			for _, cons := range alg.ConsDefs {
				if cons == nil {
					return fmt.Errorf("nil constructor definition in %s", alg.Name)
				}
				g.constructor(cons)
				g.blank()
			}
		}
	}

	for _, fn := range program.FuncDefs {
		if fn == nil {
			return fmt.Errorf("nil function definition")
		}
		if err := g.function(fn); err != nil {
			return err
		}
		g.blank()
	}

	if program.Expr != nil {
		expr, err := g.expr(program.Expr, 0)
		if err != nil {
			return err
		}
		g.exprLine(0, "", expr, ";")
	}

	return nil
}

// write consdef to output. note that name and types
// can mostly be ignored since JS has no static generics
func (g *generator) constructor(cons *parse.ConsDef) {
	params := make([]string, 0, len(cons.Args))

	for i := 0; i < len(cons.Args); i++ {
		params = append(params, fmt.Sprintf("v%d", i))
	}

	joinedParams := strings.Join(params, ", ")

	g.line(0, fmt.Sprintf("function %s(%s) {",
		cons.Name, joinedParams))
	g.line(1, fmt.Sprintf("return { tag: %s, values: [%s] };",
		strconv.Quote(cons.Name), joinedParams))
	g.line(0, "}")
}

// write funcdef to output
func (g *generator) function(fn *parse.FuncDef) error {
	params := make([]string, 0, len(fn.Params))

	for i := 0; i < len(fn.Params); i++ {
		params = append(params, fn.Params[i].Name)
	}

	joinedParams := strings.Join(params, ", ")

	// JS func header
	g.line(0, fmt.Sprintf("function %s(%s) {", fn.Name, joinedParams))
	
	// get expression JS output
	body, err := g.expr(fn.Body, 1)
	if err != nil {
		return err
	}

	// writes expression to output
	// requires more than reg line
	g.exprLine(1, "return ", body, ";")
	g.line(0, "}")
	return nil
}

// write expression to output. returns JS body as string 
func (g *generator) expr(expr parse.Expr, level int) (string, error) {
	
	// each expression type has a different case
	switch e := expr.(type) {
	case nil:
		return "", fmt.Errorf("nil expression")
	case parse.VarExpr:
		return e.Name, nil
	case parse.IntExpr:
		return strconv.Itoa(e.Value), nil
	case parse.UnitExpr:
		return "undefined", nil
	case parse.BoolExpr:
		if e.Value {
			return "true", nil
		}
		return "false", nil
	case parse.PrintlnExpr:
		arg, err := g.expr(e.Expr, level)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("console.log(%s)", arg), nil
	case parse.OpExpr:
		return g.opExpr(e, level)
	case parse.LambdaExpr:
		return g.lambdaExpr(e, level)
	case parse.CallHOFExpr:
		return g.callHOFExpr(e, level)
	case parse.CallExpr:
		return g.callExpr(e, level)
	case parse.BlockExpr:
		return g.blockExpr(e, level)
	case parse.ConsExpr:
		return g.consExpr(e, level)
	case parse.MatchExpr: // not yet implemented
		return g.matchExpr(e, level)
	default:
		return "", fmt.Errorf("unsupported expression %T", expr)
	}
}

// returns JS for operation expression
// uses recursive g.expr calls for left and right
func (g *generator) opExpr(expr parse.OpExpr, level int) (string, error) {
	left, err := g.expr(expr.Left, level)
	if err != nil {
		return "", err
	}
	right, err := g.expr(expr.Right, level)
	if err != nil {
		return "", err
	}

	op := expr.Op
	if op == "==" {
		op = "==="
	}
	return fmt.Sprintf("(%s %s %s)", left, op, right), nil
}

// returns JS for lambda expression
// uses recursive g.expr calls for lambda body
func (g *generator) lambdaExpr(expr parse.LambdaExpr, level int) (string, error) {
	params := make([]string, 0, len(expr.Params))

	for i := 0; i < len(expr.Params); i++ {
		if expr.Params[i] == nil {
			return "", fmt.Errorf("nil lambda parameter")
		}
		params = append(params, expr.Params[i].Name)
	}

	joinedParams := strings.Join(params, ", ")

	body, err := g.expr(expr.Body, level)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("((%s) => %s)", joinedParams, body), nil
}

// returns JS for higher order function call expression
// uses recursive g.expr calls for function and args
func (g *generator) callHOFExpr(expr parse.CallHOFExpr, level int) (string, error) {
	fn, err := g.expr(expr.Fn, level)
	if err != nil {
		return "", err
	}

	// processes all expr args at once
	args, err := g.exprList(expr.Args, level)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", fn, strings.Join(args, ", ")), nil
}

// returns JS for function call expression
// uses recursive g.exprList calls for only args
func (g *generator) callExpr(expr parse.CallExpr, level int) (string, error) {
	args, err := g.exprList(expr.Args, level)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", expr.Name, strings.Join(args, ", ")), nil
}

// returns JS for constructor expression
// uses recursive g.exprList calls for args
func (g *generator) consExpr(expr parse.ConsExpr, level int) (string, error) {
	args, err := g.exprList(expr.Args, level)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", expr.Name, strings.Join(args, ", ")), nil
}

// takes a list of expressions and processes them at
// the same time, returning the JS to the output
func (g *generator) exprList(exprs []parse.Expr, level int) ([]string, error) {
	out := make([]string, len(exprs))
	for i, expr := range exprs {
		js, err := g.expr(expr, level)
		if err != nil {
			return nil, err
		}
		out[i] = js
	}
	return out, nil
}

// returns JS for block expression
// currently requires implementation for stmt processing
func (g *generator) blockExpr(expr parse.BlockExpr, level int) (string, error) {
	var b strings.Builder
	b.WriteString("(() => {\n")

	for _, stmt := range expr.Stmts {
		js, err := g.stmt(stmt, level+1) // not implemented
		if err != nil {
			return "", err
		}
		b.WriteString(js)
	}

	body, err := g.expr(expr.Expr, level+1)
	if err != nil {
		return "", err
	}
	b.WriteString(g.formatExprLine(level+1, "return ", body, ";"))
	b.WriteString(indent(level))
	b.WriteString("})()")
	return b.String(), nil
}

func (g *generator) stmt(stmt parse.Stmt, level int) (string, error) {
	// stmt implementation missing
}

func (g *generator) matchExpr(expr parse.MatchExpr, level int) (string, error) {
	// matchExpr implementation missing
}

// writes a single line to output
func (g *generator) line(level int, text string) {
	g.out.WriteString(indent(level))
	g.out.WriteString(text)
	g.out.WriteByte('\n')
}

// writes an empty line to output
func (g *generator) blank() {
	g.out.WriteByte('\n')
}

// writes an expression to output
func (g *generator) exprLine(level int, prefix, expr, suffix string) {
	g.out.WriteString(g.formatExprLine(level, prefix, expr, suffix))
}

func (g *generator) formatExprLine(level int, prefix, expr, suffix string) string {
	lines := strings.Split(expr, "\n")
	var b strings.Builder
	b.WriteString(indent(level))
	b.WriteString(prefix)
	b.WriteString(lines[0])
	for _, line := range lines[1:] {
		b.WriteByte('\n')
		b.WriteString(line)
	}
	b.WriteString(suffix)
	b.WriteByte('\n')
	return b.String()
}

// indents level times
func indent(level int) string {
	return strings.Repeat("\t", level)
}
