package parse

import (
	"dada/lex"
	"fmt"
	"io"
	"strconv"
)

// Parser parses a full token stream into the typed AST nodes in ast.go.
type Parser struct {
	tokens []*lex.Token
	pos    int
}

// NewParser returns a parser initialized with all tokens from the input.
// Parse errors cause panics of type lex.Error that the caller must handle.
func NewParser(r io.RuneReader) *Parser {
	return &Parser{
		tokens: lexAll(r),
		pos:    0,
	}
}

func lexAll(r io.RuneReader) []*lex.Token {
	l := lex.NewLexer(r)
	var tokens []*lex.Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type() == lex.TokenEOF {
			return tokens
		}
	}
}

func errorf(format string, args ...interface{}) {
	panic(lex.Error(fmt.Sprintf(format, args...)))
}

// SExprString returns the expression as a formatted dotted S-expression.
func (e *AST) SExprString() string {
	if e == nil {
		return "nil"
	}
	if e.Atom != nil {
		return e.Atom.String()
	}
	if len(e.List) == 0 && e.Tail == nil {
		return "()"
	}
	if idx >= len(p.tokens) {
		errorf("token lookup past end of stream: pos=%d offset=%d len=%d", p.pos, offset, len(p.tokens))
	}
	return p.tokens[idx]
}

func (p *Parser) ParseProgram() *Program {
	var algDefs []*AlgDef
	for p.startsListConst("algdef") {
		algDefs = append(algDefs, p.parseAlgDef())
	}

	var funcDefs []*FuncDef
	for p.startsListConst("def") {
		funcDefs = append(funcDefs, p.parseFuncDef())
	}
	var b strings.Builder
	e.buildString(&b)
	return b.String()
}

// buildString is the internal helper for String
func (e *AST) buildString(b *strings.Builder) {
	if e == nil {
		b.WriteString("nil")
		return
	}
	if e.Atom != nil {
		b.WriteString(e.Atom.String())
		return
	}
	if len(e.List) == 0 && e.Tail == nil {
		b.WriteString("()")
		return
	}
	b.WriteByte('(')
	for i, elem := range e.List {
		if i > 0 {
			b.WriteByte(' ')
		}
		elem.buildString(b)
	}
	if e.Tail != nil {
		if len(e.List) > 0 {
			b.WriteString(" . ")
		}
		e.Tail.buildString(b)
	}
}

func atomExpr(tok *lex.Token) *AST {
	return &AST{Atom: tok}
}

func (p *Parser) expectConst(name string) *lex.Token {
	tok := p.getToken(0)
	if tok.Type() != lex.TokenConst || tok.Text() != name {
		errorf("expected %q, found %q", name, tok)
	}
	p.pos++
	return tok
}

func (p *Parser) expectPrimary() *lex.Token {
	tok := p.getToken(0)
	if tok.Type() != lex.TokenPrimaryExpression {
		errorf("expected identifier, found %q", tok)
	}
	p.pos++
	return tok
}

func (p *Parser) parseAlgDef() *AlgDef {
	p.expectType(lex.TokenLpar)
	p.expectConst("algdef")
	name := p.expectPrimary().Text()

	p.expectType(lex.TokenLpar)
	var typeVars []string
	for p.getToken(0).Type() != lex.TokenRpar {
		typeVars = append(typeVars, p.expectPrimary().Text())
	}
	p.expectType(lex.TokenRpar)

	var consDefs []*ConsDef
	for p.startsList() {
		consDefs = append(consDefs, p.parseConsDef())
	}
	if len(consDefs) == 0 {
		errorf("algdef %q must contain at least one constructor", name)
	}

	p.expectType(lex.TokenRpar)

	return &AlgDef{
		Name:     name,
		TypeVars: typeVars,
		ConsDefs: consDefs,
	}
}

func (p *Parser) parseConsDef() *ConsDef {
	p.expectType(lex.TokenLpar)
	name := p.expectPrimary().Text()

// Parse parses an entire program as a sequence of top-level S-expressions.
func (p *Parser) Parse() []*AST {
	var program []*AST
	for {
		expr := p.SExpr()
		if expr == nil {
			return program
		}
		program = append(program, expr)
	}
}

	p.expectType(lex.TokenRpar)

	return &ConsDef{
		Name: name,
		Args: args,
	}
}

func (p *Parser) parseExpr(allowEOF bool) *AST {
	tok := p.next()
	switch tok.Type() {
	case lex.TokenEOF:
		if allowEOF {
			return nil
		}
		panic(lex.EOF("eof"))
	case lex.TokenLpar:
		return p.parseList()
	default:
		if isAtomToken(tok.Type()) {
			return atomExpr(tok)
		}
	}

	if tok.Type() == lex.TokenPrimaryExpression {
		p.pos++
		return TypeVar{Name: tok.Text()}
	}

	if tok.Type() != lex.TokenLpar {
		errorf("expected type, found %q", tok)
	}

	p.pos++
	head := p.getToken(0)

	switch {
	case head.Type() == lex.TokenArrow:
		p.pos++
		p.expectType(lex.TokenLpar)
		var params []Type
		for p.getToken(0).Type() != lex.TokenRpar {
			params = append(params, p.parseType())
		}
		p.expectType(lex.TokenRpar)
		ret := p.parseType()
		p.expectType(lex.TokenRpar)
		return FuncType{Params: params, Return: ret}

	case head.Type() == lex.TokenConst && head.Text() == "alg":
		p.pos++
		name := p.expectPrimary().Text()
		var args []Type
		for p.getToken(0).Type() != lex.TokenRpar {
			args = append(args, p.parseType())
		}
		p.expectType(lex.TokenRpar)
		return AlgType{Name: name, Args: args}

	default:
		errorf("expected type form, found %q", head)
		panic("not reached")
	}
}
