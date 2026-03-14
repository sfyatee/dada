package lex

import (
	"bufio"
	"io"
	"unicode"
)

type lexer struct {
	r *bufio.Reader
}

func newLexer(rr io.RuneReader) *lexer {
	if br, ok := rr.(*bufio.Reader); ok {
		return &lexer{r: br}
	}
	return &lexer{r: bufio.NewReader(runeReaderAdapter{rr: rr})}
}

type runeReaderAdapter struct {
	rr io.RuneReader
}

func (a runeReaderAdapter) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	r, _, err := a.rr.ReadRune()
	if err != nil {
		return 0, err
	}
	return copy(p, string(r)), nil
}

func (l *lexer) next() *token {
	l.skipSpace()

	r, err := l.readRune()
	if err == io.EOF {
		return &token{typ: tokenEOF, text: "EOF"}
	}
	if err != nil {
		panic(Error(err.Error()))
	}

	switch r {
	case '(':
		return &token{typ: tokenLpar, text: "("}
	case ')':
		return &token{typ: tokenRpar, text: ")"}
	case '.':
		return &token{typ: tokenDot, text: "."}
	case '\'':
		return &token{typ: tokenQuote, text: "'"}
	}

	if isDigit(r) || (r == '-' && l.peekDigit()) {
		return l.lexNumber(r)
	}

	if isAtomStart(r) {
		return l.lexAtomOrConst(r)
	}

	panic(Error("invalid character: " + string(r)))
}

func (l *lexer) skipSpace() rune {
	for {
		r, err := l.readRune()
		if err == io.EOF {
			return 0
		}
		if err != nil {
			panic(Error(err.Error()))
		}
		if !unicode.IsSpace(r) {
			if err := l.r.UnreadRune(); err != nil {
				panic(Error(err.Error()))
			}
			return r
		}
	}
}

func (l *lexer) skipToNewline() {
	for {
		r, err := l.readRune()
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(Error(err.Error()))
		}
		if r == '\n' {
			return
		}
	}
}

func (l *lexer) lexNumber(first rune) *token {
	rs := []rune{first}

	if first == '-' {
		r, err := l.readRune()
		if err != nil {
			panic(Error("bad number"))
		}
		if !isDigit(r) {
			panic(Error("bad number"))
		}
		rs = append(rs, r)
	}

	for {
		r, err := l.readRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(Error(err.Error()))
		}
		if !isDigit(r) {
			if err := l.r.UnreadRune(); err != nil {
				panic(Error(err.Error()))
			}
			break
		}
		rs = append(rs, r)
	}

	return &token{
		typ:  tokenNumber,
		text: string(rs),
	}
}

func (l *lexer) lexAtomOrConst(first rune) *token {
	rs := []rune{first}

	for {
		r, err := l.readRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(Error(err.Error()))
		}
		if !isAtomPart(r) {
			if err := l.r.UnreadRune(); err != nil {
				panic(Error(err.Error()))
			}
			break
		}
		rs = append(rs, r)
	}

	s := string(rs)
	if isReservedConst(s) {
		return &token{
			typ:  tokenConst,
			text: s,
		}
	}

	return &token{
		typ:  tokenAtom,
		text: s,
	}
}

func (l *lexer) peekDigit() bool {
	r, err := l.readRune()
	if err != nil {
		return false
	}
	if err := l.r.UnreadRune(); err != nil {
		panic(Error(err.Error()))
	}
	return isDigit(r)
}

func (l *lexer) readRune() (rune, error) {
	r, _, err := l.r.ReadRune()
	return r, err
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAtomStart(r rune) bool {
	if unicode.IsLetter(r) || r == '_' {
		return true
	}
	switch r {
	case '+', '-', '*', '/', '<', '>', '=', '?', '!':
		return true
	default:
		return false
	}
}

func isAtomPart(r rune) bool {
	return isAtomStart(r) || isDigit(r)
}

func isReservedConst(s string) bool {
	switch s {
	case "true", "false", "unit", "nil":
		return true
	default:
		return false
	}
}
