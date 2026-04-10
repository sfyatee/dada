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
// Exactly one of Atom or List should be populated.
type AST struct {
	Atom *lex.Token
	List []*AST
	Tail *AST
}

// Expr is kept as an alias for compatibility with older parser code.
type Expr = AST

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
	tail := "nil"
	if e.Tail != nil {
		tail = e.Tail.SExprString()
	}
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
	b.WriteByte(')')
}

func atomExpr(tok *lex.Token) *AST {
	return &AST{Atom: tok}
}

func listExpr(list []*AST, tail *AST) *AST {
	return &AST{List: list, Tail: tail}
}

func isAtomToken(typ lex.TokenType) bool {
	switch typ {
	case lex.TokenAtom, lex.TokenConst, lex.TokenNumber, lex.TokenEqualEqual, lex.TokenArrow, lex.TokenUnderscore:
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

// SkipSpace skips leading spaces, returning the rune that follows.
func (p *Parser) SkipSpace() rune {
	return p.lex.SkipSpace()
}

// SkipToNewline advances the input past the next newline.
func (p *Parser) SkipToEndOfLine() {
	p.lex.SkipToNewline()
}

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

func (p *Parser) back(tok *lex.Token) {
	p.peekTok = tok
}

// Parse reads a single AST node using the s-expression grammar.
func (p *Parser) Parse() *AST {
	return p.SExpr()
}

// SExpr parses an S-Expression, returning nil only at end of input.
func (p *Parser) SExpr() *AST {
	return p.parseExpr(true)
}

// List parses a list expression.
func (p *Parser) List() *AST {
	return p.parseExpr(false)
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
			return listExpr(elems, nil)
		case lex.TokenDot:
			if len(elems) == 0 {
				errorf("bad token parsing list: %q", tok)
			}
			tail := p.parseExpr(false)
			rpar := p.next()
			if rpar.Type() != lex.TokenRpar {
				errorf("expected ')', found %q", rpar)
			}
			return listExpr(elems, tail)
		default:
			p.back(tok)
			elems = append(elems, p.parseExpr(false))
		}
	}
}
