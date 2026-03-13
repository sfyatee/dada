package scan

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