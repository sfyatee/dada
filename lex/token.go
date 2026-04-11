package lex

import (
	"fmt"
	"strings"
)

type Error string

func (e Error) Error() string { return string(e) }

type EOF string

func (e EOF) Error() string { return string(e) }

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenLpar
	tokenRpar
	tokenQuote
	tokenEqualEqual
	tokenArrow
	tokenUnderscore
	tokenPrimaryExpression
	tokenConst
	tokenNumber
)

type TokenType = tokenType

const (
	TokenError = tokenError
	TokenEOF = tokenEOF
	TokenLpar = tokenLpar
	TokenRpar = tokenRpar
	TokenQuote = tokenQuote
	TokenEqualEqual = tokenEqualEqual
	TokenArrow = tokenArrow
	TokenUnderscore = tokenUnderscore
	TokenPrimaryExpression = tokenPrimaryExpression
	TokenConst = tokenConst
	TokenNumber = tokenNumber
)

func (tt tokenType) String() string {
	switch tt {
	case tokenError:
		return "error"
	case tokenEOF:
		return "EOF"
	case tokenLpar:
		return "("
	case tokenRpar:
		return ")"
	case tokenQuote:
		return "'"
	case tokenEqualEqual:
		return "=="
	case tokenArrow:
		return "=>"
	case tokenUnderscore:
		return "_"
	case tokenPrimaryExpression:
		return "primaryExpression"
	case tokenConst:
		return "const"
	case tokenNumber:
		return "number"
	default:
		return fmt.Sprintf("tokenType(%d)", int(tt))
	}
}

type token struct {
	typ  tokenType
	text string
}

type Token = token

func (t *token) String() string {
	if t == nil {
		return "<nil>"
	}
	if t.text != "" {
		return t.text
	}
	return t.typ.String()
}

func (t *token) buildString(b *strings.Builder) {
	if t == nil {
		return
	}
	b.WriteString(t.String())
}

func (t *token) Type() TokenType {
	if t == nil {
		return TokenError
	}
	return t.typ
}

func (t *token) Text() string {
	if t == nil {
		return ""
	}
	return t.text
}
