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
    tokenDot
    tokenQuote
    tokenAtom
    tokenConst
    tokenNumber
)

type token struct {
    typ  tokenType
    text string
}

// Our parser has support for . an quote (`). Added for .
var tokQuote = &token{
    typ:  tokenAtom,
    text: "quote",
}

func (tt tokenType) String() string {
    switch tt {
    case tokenError:
        return "error"
    case tokenEOF:
        return "eof"
    case tokenLpar:
        return "("
    case tokenRpar:
        return ")"
    case tokenDot:
        return "."
    case tokenQuote:
        return "'"
    case tokenAtom:
        return "atom"
    case tokenConst:
        return "const"
    case tokenNumber:
        return "number"
    default:
        return fmt.Sprintf("tokenType(%d)", int(tt))
    }
}

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