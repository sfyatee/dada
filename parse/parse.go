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

func (p *Parser) getToken(offset int) *lex.Token {
	idx := p.pos + offset
	if idx < 0 {
		errorf("token lookup before start of stream: pos=%d offset=%d", p.pos, offset)
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

	expr := p.parseExpr(false)

	if tok := p.getToken(0); tok.Type() != lex.TokenEOF {
		errorf("expected EOF, found %q", tok)
	}

	return &Program{
		AlgDefs:  algDefs,
		FuncDefs: funcDefs,
		Expr:     expr,
	}
}

func (p *Parser) startsListConst(name string) bool {
	return p.getToken(0).Type() == lex.TokenLpar &&
		p.getToken(1).Type() == lex.TokenConst &&
		p.getToken(1).Text() == name
}

func (p *Parser) expectType(typ lex.TokenType) *lex.Token {
	tok := p.getToken(0)
	if tok.Type() != typ {
		errorf("expected %s, found %q", typ, tok)
	}
	p.pos++
	return tok
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

	var args []Type
	for p.getToken(0).Type() != lex.TokenRpar {
		args = append(args, p.parseType())
	}

	p.expectType(lex.TokenRpar)

	return &ConsDef{
		Name: name,
		Args: args,
	}
}

func (p *Parser) parseType() Type {
	tok := p.getToken(0)

	if tok.Type() == lex.TokenConst {
		switch tok.Text() {
		case "Int":
			p.pos++
			return IntType{}
		case "Unit":
			p.pos++
			return UnitType{}
		case "Boolean":
			p.pos++
			return BooleanType{}
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

func (p *Parser) startsList() bool {
	return p.getToken(0).Type() == lex.TokenLpar
}

func (p *Parser) parseFuncDef() *FuncDef {
	p.expectType(lex.TokenLpar)
	p.expectConst("def")
	name := p.expectPrimary().Text()

	p.expectType(lex.TokenLpar)
	var typeVars []string
	for p.getToken(0).Type() != lex.TokenRpar {
		typeVars = append(typeVars, p.expectPrimary().Text())
	}
	p.expectType(lex.TokenRpar)

	p.expectType(lex.TokenLpar)
	var params []*Param
	for p.getToken(0).Type() != lex.TokenRpar {
		params = append(params, p.parseParam())
	}
	p.expectType(lex.TokenRpar)

	returnType := p.parseType()
	body := p.parseExpr(false)

	p.expectType(lex.TokenRpar)

	return &FuncDef{
		Name:       name,
		TypeVars:   typeVars,
		Params:     params,
		ReturnType: returnType,
		Body:       body,
	}
}

func (p *Parser) parseParam() *Param {
	p.expectType(lex.TokenLpar)
	typ := p.parseType()
	name := p.expectPrimary().Text()
	p.expectType(lex.TokenRpar)

	return &Param{
		Type: typ,
		Name: name,
	}
}

func (p *Parser) parseExpr(allowEOF bool) Expr {
	tok := p.getToken(0)

	switch tok.Type() {
	case lex.TokenEOF:
		if allowEOF {
			return nil
		}
		panic(lex.EOF("eof"))

	case lex.TokenNumber:
		p.pos++
		value, err := strconv.Atoi(tok.Text())
		if err != nil {
			errorf("bad integer literal %q", tok.Text())
		}
		return IntExpr{Value: value}

	case lex.TokenPrimaryExpression:
		p.pos++
		return VarExpr{Name: tok.Text()}

	case lex.TokenUnderscore:
		errorf("unexpected '_' in expression")

	case lex.TokenConst:
		switch tok.Text() {
		case "unit":
			p.pos++
			return UnitExpr{}
		case "true":
			p.pos++
			return BoolExpr{Value: true}
		case "false":
			p.pos++
			return BoolExpr{Value: false}
		default:
			errorf("unexpected constant in expression: %q", tok)
		}

	case lex.TokenLpar:
		return p.parseListExpr()
	}

	errorf("bad token in expression: %q", tok)
	panic("not reached")
}

func (p *Parser) parseListExpr() Expr {
	p.expectType(lex.TokenLpar)
	head := p.getToken(0)

	switch {
	case head.Text() == "println":
		p.pos++
		expr := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return PrintlnExpr{Expr: expr}

	case p.isOpToken(head):
		p.pos++
		left := p.parseExpr(false)
		right := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return OpExpr{Op: head.Text(), Left: left, Right: right}

	case head.Type() == lex.TokenArrow:
		p.pos++
		p.expectType(lex.TokenLpar)
		var params []*Param
		for p.getToken(0).Type() != lex.TokenRpar {
			params = append(params, p.parseParam())
		}
		p.expectType(lex.TokenRpar)
		body := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return LambdaExpr{Params: params, Body: body}

	case head.Type() == lex.TokenConst && head.Text() == "callhof":
		p.pos++
		fn := p.parseExpr(false)
		var args []Expr
		for p.getToken(0).Type() != lex.TokenRpar {
			args = append(args, p.parseExpr(false))
		}
		p.expectType(lex.TokenRpar)
		return CallHOFExpr{Fn: fn, Args: args}

	case head.Type() == lex.TokenConst && head.Text() == "call":
		p.pos++
		name := p.expectPrimary().Text()
		var args []Expr
		for p.getToken(0).Type() != lex.TokenRpar {
			args = append(args, p.parseExpr(false))
		}
		p.expectType(lex.TokenRpar)
		return CallExpr{Name: name, Args: args}

	case head.Type() == lex.TokenConst && head.Text() == "block":
		p.pos++
		var stmts []Stmt
		for p.startsListConst("val") || p.startsListConst("var") || p.startsAssignStmt() {
			stmts = append(stmts, p.parseStmt())
		}
		expr := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return BlockExpr{Stmts: stmts, Expr: expr}

	case head.Type() == lex.TokenConst && head.Text() == "cons":
		p.pos++
		name := p.expectPrimary().Text()
		var args []Expr
		for p.getToken(0).Type() != lex.TokenRpar {
			args = append(args, p.parseExpr(false))
		}
		p.expectType(lex.TokenRpar)
		return ConsExpr{Name: name, Args: args}

	case head.Type() == lex.TokenConst && head.Text() == "match":
		p.pos++
		expr := p.parseExpr(false)
		var cases []*Case
		for p.startsListConst("case") {
			cases = append(cases, p.parseCase())
		}
		if len(cases) == 0 {
			errorf("match requires at least one case")
		}
		p.expectType(lex.TokenRpar)
		return MatchExpr{Expr: expr, Cases: cases}
	}

	errorf("unknown expression form starting with %q", head)
	panic("not reached")
}

func (p *Parser) parseStmt() Stmt {
	p.expectType(lex.TokenLpar)
	head := p.getToken(0)
	p.pos++

	if head.Type() == lex.TokenConst && head.Text() == "val" {
		name := p.expectPrimary().Text()
		expr := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return ValStmt{Name: name, Expr: expr}
	}

	if head.Type() == lex.TokenConst && head.Text() == "var" {
		typ := p.parseType()
		name := p.expectPrimary().Text()
		expr := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return VarStmt{Type: typ, Name: name, Expr: expr}
	}

	if head.Type() == lex.TokenPrimaryExpression && head.Text() == "=" {
		name := p.expectPrimary().Text()
		expr := p.parseExpr(false)
		p.expectType(lex.TokenRpar)
		return AssignStmt{Name: name, Expr: expr}
	}

	errorf("unknown statement form %q", head)
	panic("not reached")
}

func (p *Parser) parseCase() *Case {
	p.expectType(lex.TokenLpar)
	p.expectConst("case")
	pattern := p.parsePattern()
	expr := p.parseExpr(false)
	p.expectType(lex.TokenRpar)
	return &Case{Pattern: pattern, Expr: expr}
}

func (p *Parser) parsePattern() Pattern {
	tok := p.getToken(0)

	switch tok.Type() {
	case lex.TokenUnderscore:
		p.pos++
		return WildcardPattern{}

	case lex.TokenPrimaryExpression:
		p.pos++
		return VarPattern{Name: tok.Text()}

	case lex.TokenLpar:
		p.expectType(lex.TokenLpar)
		p.expectConst("cons")
		name := p.expectPrimary().Text()
		var patterns []Pattern
		for p.getToken(0).Type() != lex.TokenRpar {
			patterns = append(patterns, p.parsePattern())
		}
		p.expectType(lex.TokenRpar)
		return ConsPattern{Name: name, Patterns: patterns}
	}

	errorf("bad token in pattern: %q", tok)
	panic("not reached")
}

func (p *Parser) startsAssignStmt() bool {
	return p.getToken(0).Type() == lex.TokenLpar &&
		p.getToken(1).Type() == lex.TokenPrimaryExpression &&
		p.getToken(1).Text() == "="
}

func (p *Parser) isOpToken(tok *lex.Token) bool {
	if tok == nil {
		return false
	}
	switch tok.Type() {
	case lex.TokenEqualEqual:
		return true
	case lex.TokenPrimaryExpression:
		switch tok.Text() {
		case "+", "-", "*", "/", "<":
			return true
		}
	}
	return false
}
