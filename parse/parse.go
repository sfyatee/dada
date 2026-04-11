// Copyright 2020 Rob Pike. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package parse

import (
	"dada/lex"
	"fmt"
	"io"
	"strings"
)

// printSExpr configures the interpreter to print S-Expressions.
var printSExpr bool

// Config configures the interpreter. The argument specifies whether output
// should be S-Expressions rather than lists.
func Config(alwaysPrintSExprs bool) {
	printSExpr = alwaysPrintSExprs
}

// AST represents an arbitrary parsed S-expression tree.
// Exactly one of PrimaryExpression or List should be populated.
type AST struct {
	PrimaryExpression *lex.Token
	List []*AST
}

// Expr is kept as an alias for compatibility with older parser code.
type Expr = AST

// SExprString returns the expression as a formatted dotted S-expression.
func (e *AST) SExprString() string {
	if e == nil {
		return "nil"
	}
	if e.PrimaryExpression != nil {
		return e.PrimaryExpression.String()
	}
	if len(e.List) == 0 {
		return "nil"
	}
	tail := "nil"
	for i := len(e.List) - 1; i >= 0; i-- {
		tail = fmt.Sprintf("(%s . %s)", e.List[i].SExprString(), tail)
	}
	return tail
}

// String returns the expression as a formatted list (unless printSExpr is set).
func (e *AST) String() string {
	if printSExpr {
		return e.SExprString()
	}
	if e == nil {
		return "nil"
	}
	var b strings.Builder
	e.buildString(&b, true)
	return b.String()
}

// buildString is the internals of the String method. simplifyQuote
// specifies whether (' expr) should be printed as 'expr.
func (e *AST) buildString(b *strings.Builder, simplifyQuote bool) {
	if e == nil {
		b.WriteString("nil")
		return
	}
	if e.PrimaryExpression != nil {
		b.WriteString(e.PrimaryExpression.String())
		return
	}
	if len(e.List) == 0 {
		b.WriteString("()")
		return
	}
	if simplifyQuote && isQuoteList(e) {
		b.WriteByte('\'')
		e.List[1].buildString(b, simplifyQuote)
		return
	}
	b.WriteByte('(')
	for i, elem := range e.List {
		if i > 0 {
			b.WriteByte(' ')
		}
		elem.buildString(b, simplifyQuote)
	}
	b.WriteByte(')')
}

func isQuoteList(e *AST) bool {
	return e != nil &&
		e.PrimaryExpression == nil &&
		len(e.List) == 2 &&
		e.List[0] != nil &&
		e.List[0].PrimaryExpression != nil &&
		e.List[0].PrimaryExpression.Type() == lex.TokenQuote
}

func primaryExpressionExpr(tok *lex.Token) *AST {
	return &AST{PrimaryExpression: tok}
}

func listExpr(list []*AST) *AST {
	return &AST{List: list}
}

func isPrimaryExpressionToken(typ lex.TokenType) bool {
	switch typ {
	case lex.TokenPrimaryExpression, lex.TokenConst, lex.TokenNumber, lex.TokenEqualEqual, lex.TokenArrow, lex.TokenUnderscore:
		return true
	default:
		return false
	}
}

// Parser is the parser for S-expression syntax.
type Parser struct {
	lex     *lex.Lexer
	peekTok *lex.Token
}

// NewParser returns a new parser that will read from the RuneReader.
// Parse errors cause panics of type lex.Error that the caller must handle.
func NewParser(r io.RuneReader) *Parser {
	return &Parser{
		lex:     lex.NewLexer(r),
		peekTok: nil,
	}
}

// // SkipSpace skips leading spaces, returning the rune that follows.
// func (p *Parser) SkipSpace() rune {
// 	return p.lex.SkipSpace()
// }

// // SkipToNewline advances the input past the next newline.
// func (p *Parser) SkipToEndOfLine() {
// 	p.lex.SkipToNewline()
// }

func errorf(format string, args ...interface{}) {
	panic(lex.Error(fmt.Sprintf(format, args...)))
}

func (p *Parser) next() *lex.Token {
	if tok := p.peekTok; tok != nil {
		p.peekTok = nil
		return tok
	}
	return p.lex.Next()
}

func (p *Parser) getToken(final int position) {

}

func (p *Parser) back(tok *lex.Token) {
	p.peekTok = tok
}

// Parse reads a single AST node using the list-oriented grammar.
func (p *Parser) ParseProgram() *AST {

	return p.parseExpr(true)
}


func (p *Parser) parseExpr(allowEOF bool) *AST {
	tok := p.next()
	switch tok.Type() {
	case lex.TokenEOF:
		if allowEOF {
			return nil
		}
		panic(lex.EOF("eof"))
	case lex.TokenQuote:
		return p.quote(tok)
	case lex.TokenLpar:
		return p.parseList()
	default:
		if isPrimaryExpressionToken(tok.Type()) {
			return primaryExpressionExpr(tok)
		}
	}
	errorf("bad token in expression: %q", tok)
	panic("not reached")
}

func (p *Parser) parseList() *AST {
	var elems []*AST
	for {
		tok := p.next()
		switch tok.Type() {
		case lex.TokenEOF:
			panic(lex.EOF("eof"))
		case lex.TokenRpar:
			return listExpr(elems)
		default:
			p.back(tok)
			elems = append(elems, p.parseExpr(false))
		}
	}
}

// quote parses a quoted expression. The leading quote has been consumed.
func (p *Parser) quote(tok *lex.Token) *AST {
	return listExpr([]*AST{primaryExpressionExpr(tok), p.parseExpr(false)})
}
