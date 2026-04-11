package lex

import (
	"strings"
	"testing"
)

func collectTokens(input string) []*token {
	l := newLexer(strings.NewReader(input))
	var tokens []*token

	for {
		tok := l.next()
		tokens = append(tokens, tok)
		if tok.typ == tokenEOF {
			return tokens
		}
	}
}

func TestReservedWordsAndNil(t *testing.T) {
	input := "true false unit Int Unit Boolean alg callhof call block cons match val var case algdef def nil"
	tokens := collectTokens(input)

	want := []struct {
		typ  tokenType
		text string
	}{
		{tokenConst, "true"},
		{tokenConst, "false"},
		{tokenConst, "unit"},
		{tokenConst, "Int"},
		{tokenConst, "Unit"},
		{tokenConst, "Boolean"},
		{tokenConst, "alg"},
		{tokenConst, "callhof"},
		{tokenConst, "call"},
		{tokenConst, "block"},
		{tokenConst, "cons"},
		{tokenConst, "match"},
		{tokenConst, "val"},
		{tokenConst, "var"},
		{tokenConst, "case"},
		{tokenConst, "algdef"},
		{tokenConst, "def"},
		{tokenPrimaryExpression, "nil"},
		{tokenEOF, "EOF"},
	}

	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(want))
	}

	for i, got := range tokens {
		if got.typ != want[i].typ || got.text != want[i].text {
			t.Fatalf("token %d: got (%v, %q), want (%v, %q)", i, got.typ, got.text, want[i].typ, want[i].text)
		}
	}
}

func TestSpecialTokens(t *testing.T) {
	tokens := collectTokens("== => _")

	want := []struct {
		typ  tokenType
		text string
	}{
		{tokenEqualEqual, "=="},
		{tokenArrow, "=>"},
		{tokenUnderscore, "_"},
		{tokenEOF, "EOF"},
	}

	if len(tokens) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(want))
	}

	for i, got := range tokens {
		if got.typ != want[i].typ || got.text != want[i].text {
			t.Fatalf("token %d: got (%v, %q), want (%v, %q)", i, got.typ, got.text, want[i].typ, want[i].text)
		}
	}
}

func TestInvalidCharacters(t *testing.T) {
	for _, input := range []string{"?", "!", ">"} {
		t.Run(input, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %q", input)
				}
			}()

			collectTokens(input)
		})
	}
}
