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
		{tokenAtom, "nil"},
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

func TestExportedLexerAPI(t *testing.T) {
	l := NewLexer(strings.NewReader("x"))

	tok := l.Next()
	if got, want := tok.Text(), "x"; got != want {
		t.Fatalf("Next(): got %q, want %q", got, want)
	}

	l = NewLexer(strings.NewReader("   y"))
	r := l.SkipSpace()
	if got, want := r, 'y'; got != want {
		t.Fatalf("SkipSpace(): got %q, want %q", got, want)
	}

	l = NewLexer(strings.NewReader("ignore this line\nz"))
	l.SkipToNewline()
	tok = l.Next()
	if got, want := tok.Text(), "z"; got != want {
		t.Fatalf("SkipToNewline(): got %q, want %q", got, want)
	}
}

func TestLexNumbers(t *testing.T) {
	tokens := collectTokens("123 -456")

	want := []struct {
		typ  tokenType
		text string
	}{
		{tokenNumber, "123"},
		{tokenNumber, "-456"},
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

func TestTokenHelpers(t *testing.T) {
	if got, want := Error("bad").Error(), "bad"; got != want {
		t.Fatalf("Error.Error(): got %q, want %q", got, want)
	}

	if got, want := EOF("eof").Error(), "eof"; got != want {
		t.Fatalf("EOF.Error(): got %q, want %q", got, want)
	}

	if got, want := tokenNumber.String(), "number"; got != want {
		t.Fatalf("tokenType.String(): got %q, want %q", got, want)
	}

	tok := &token{typ: tokenNumber, text: "123"}

	if got, want := tok.String(), "123"; got != want {
		t.Fatalf("token.String(): got %q, want %q", got, want)
	}

	if got, want := tok.Type(), tokenNumber; got != want {
		t.Fatalf("token.Type(): got %v, want %v", got, want)
	}

	if got, want := tok.Text(), "123"; got != want {
		t.Fatalf("token.Text(): got %q, want %q", got, want)
	}

	var b strings.Builder
	tok.buildString(&b)
	if got, want := b.String(), "123"; got != want {
		t.Fatalf("token.buildString(): got %q, want %q", got, want)
	}
}

func TestTokenNilAndDefaultCases(t *testing.T) {
	var tok *token

	if got, want := tok.String(), "<nil>"; got != want {
		t.Fatalf("nil token String(): got %q, want %q", got, want)
	}

	if got, want := tok.Type(), tokenError; got != want {
		t.Fatalf("nil token Type(): got %v, want %v", got, want)
	}

	if got, want := tok.Text(), ""; got != want {
		t.Fatalf("nil token Text(): got %q, want %q", got, want)
	}

	var b strings.Builder
	tok.buildString(&b)
	if got, want := b.String(), ""; got != want {
		t.Fatalf("nil token buildString(): got %q, want %q", got, want)
	}

	tok = &token{typ: tokenLpar}
	if got, want := tok.String(), "("; got != want {
		t.Fatalf("token without text String(): got %q, want %q", got, want)
	}

	if got, want := tokenType(999).String(), "tokenType(999)"; got != want {
		t.Fatalf("unknown tokenType String(): got %q, want %q", got, want)
	}
}

func TestTokenTypeStringValues(t *testing.T) {
	tests := []struct {
		typ  tokenType
		want string
	}{
		{tokenError, "error"},
		{tokenEOF, "EOF"},
		{tokenLpar, "("},
		{tokenRpar, ")"},
		{tokenDot, "."},
		{tokenQuote, "'"},
		{tokenEqualEqual, "=="},
		{tokenArrow, "=>"},
		{tokenUnderscore, "_"},
		{tokenAtom, "atom"},
		{tokenConst, "const"},
		{tokenNumber, "number"},
	}

	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Fatalf("tokenType %v: got %q, want %q", tt.typ, got, tt.want)
		}
	}
}
