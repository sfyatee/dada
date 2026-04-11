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
