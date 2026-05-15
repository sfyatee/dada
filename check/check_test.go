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

func TestCheckFunctionDefinition(t *testing.T) {
	prog := &parse.Program{
		FuncDefs: []*parse.FuncDef{
			{
				Name: "addOne",
				Params: []*parse.Param{
					{
						Type: parse.IntType{},
						Name: "x",
					},
				},
				ReturnType: parse.IntType{},
				Body: parse.OpExpr{
					Op: "+",
					Left: parse.VarExpr{
						Name: "x",
					},
					Right: parse.IntExpr{
						Value: 1,
					},
				},
			},
		},
		Expr: parse.CallExpr{
			Name: "addOne",
			Args: []parse.Expr{
				parse.IntExpr{
					Value: 5,
				},
			},
		},
	}

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckRejectsFunctionReturnMismatch(t *testing.T) {
	prog := &parse.Program{
		FuncDefs: []*parse.FuncDef{
			{
				Name:       "bad",
				ReturnType: parse.BooleanType{},
				Body: parse.IntExpr{
					Value: 1,
				},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckLambdaAndCallHOF(t *testing.T) {
	env := NewTypeEnv(nil)

	typ, err := newChecker().checkExpr(
		env,
		parse.CallHOFExpr{
			Fn: parse.LambdaExpr{
				Params: []*parse.Param{
					{
						Type: parse.IntType{},
						Name: "x",
					},
				},
				Body: parse.OpExpr{
					Op: "+",
					Left: parse.VarExpr{
						Name: "x",
					},
					Right: parse.IntExpr{
						Value: 1,
					},
				},
			},
			Args: []parse.Expr{
				parse.IntExpr{
					Value: 5,
				},
			},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !equalTypes(typ, parse.IntType{}) {
		t.Fatalf("expected int")
	}
}

func TestCheckRejectsCallHOFArgMismatch(t *testing.T) {
	env := NewTypeEnv(nil)

	_, err := newChecker().checkExpr(
		env,
		parse.CallHOFExpr{
			Fn: parse.LambdaExpr{
				Params: []*parse.Param{
					{
						Type: parse.IntType{},
						Name: "x",
					},
				},
				Body: parse.VarExpr{
					Name: "x",
				},
			},
			Args: []parse.Expr{
				parse.BoolExpr{
					Value: true,
				},
			},
		},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckBlockWithValVarAndAssign(t *testing.T) {
	env := NewTypeEnv(nil)

	typ, err := newChecker().checkExpr(
		env,
		parse.BlockExpr{
			Stmts: []parse.Stmt{
				parse.ValStmt{
					Name: "x",
					Expr: parse.IntExpr{
						Value: 1,
					},
				},
				parse.VarStmt{
					Type: parse.IntType{},
					Name: "y",
					Expr: parse.IntExpr{
						Value: 2,
					},
				},
				parse.AssignStmt{
					Name: "y",
					Expr: parse.OpExpr{
						Op: "+",
						Left: parse.VarExpr{
							Name: "x",
						},
						Right: parse.VarExpr{
							Name: "y",
						},
					},
				},
			},
			Expr: parse.VarExpr{
				Name: "y",
			},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !equalTypes(typ, parse.IntType{}) {
		t.Fatalf("expected int")
	}
}

func TestCheckRejectsAssignToVal(t *testing.T) {
	env := NewTypeEnv(nil)

	_, err := newChecker().checkExpr(
		env,
		parse.BlockExpr{
			Stmts: []parse.Stmt{
				parse.ValStmt{
					Name: "x",
					Expr: parse.IntExpr{
						Value: 1,
					},
				},
				parse.AssignStmt{
					Name: "x",
					Expr: parse.IntExpr{
						Value: 2,
					},
				},
			},
			Expr: parse.VarExpr{
				Name: "x",
			},
		},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsVarTypeMismatch(t *testing.T) {
	env := NewTypeEnv(nil)

	_, err := newChecker().checkExpr(
		env,
		parse.BlockExpr{
			Stmts: []parse.Stmt{
				parse.VarStmt{
					Type: parse.IntType{},
					Name: "x",
					Expr: parse.BoolExpr{
						Value: true,
					},
				},
			},
			Expr: parse.VarExpr{
				Name: "x",
			},
		},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsAssignmentTypeMismatch(t *testing.T) {
	env := NewTypeEnv(nil)

	_, err := newChecker().checkExpr(
		env,
		parse.BlockExpr{
			Stmts: []parse.Stmt{
				parse.VarStmt{
					Type: parse.IntType{},
					Name: "x",
					Expr: parse.IntExpr{
						Value: 1,
					},
				},
				parse.AssignStmt{
					Name: "x",
					Expr: parse.BoolExpr{
						Value: true,
					},
				},
			},
			Expr: parse.VarExpr{
				Name: "x",
			},
		},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}
