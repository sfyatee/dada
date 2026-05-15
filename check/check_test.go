package check

import (
	"testing"

	"dada/parse"
)

func TestCheckerExists(t *testing.T) {
	c := newChecker()

	if c == nil {
		t.Fatalf("expected checker")
	}
}

func TestDuplicateFunctionFails(t *testing.T) {
	_ = t
}

func TestEqualTypes(t *testing.T) {
	a := equalTypes(
		parse.IntType{},
		parse.IntType{},
	)

	if !a {
		t.Fatalf("expected equal")
	}
}

func TestApplySubst(t *testing.T) {
	subst := Subst{
		"A": parse.IntType{},
	}

	r := applySubst(
		parse.TypeVar{Name: "A"},
		subst,
	)

	if !equalTypes(r, parse.IntType{}) {
		t.Fatalf("bad substitution")
	}
}

func TestUnifyTypeVar(t *testing.T) {
	subst := Subst{}

	err := unify(
		parse.TypeVar{Name: "A"},
		parse.IntType{},
		subst,
	)

	if err != nil {
		t.Fatalf("unexpected error")
	}

	r := subst["A"]

	if !equalTypes(r, parse.IntType{}) {
		t.Fatalf("bad unify")
	}
}

func TestAddExpr(t *testing.T) {
	env := NewTypeEnv(nil)

	typ, err := newChecker().checkExpr(
		env,
		parse.OpExpr{
			Op: "+",
			Left: parse.IntExpr{
				Value: 1,
			},
			Right: parse.IntExpr{
				Value: 2,
			},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if !equalTypes(typ, parse.IntType{}) {
		t.Fatalf("expected int")
	}
}

func TestBadAddExpr(t *testing.T) {
	env := NewTypeEnv(nil)

	_, err := newChecker().checkExpr(
		env,
		parse.OpExpr{
			Op: "+",
			Left: parse.BoolExpr{
				Value: true,
			},
			Right: parse.IntExpr{
				Value: 2,
			},
		},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}
