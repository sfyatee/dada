package lex

import "io"

type Lexer = lexer

func NewLexer(rr io.RuneReader) *Lexer {
	return newLexer(rr)
}

func (l *lexer) Next() *Token {
	return l.next()
}

func (l *lexer) SkipSpace() rune {
	return l.skipSpace()
}

func (l *lexer) SkipToNewline() {
	l.skipToNewline()
}
