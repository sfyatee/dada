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
	prog := &parse.Program{
		FuncDefs: []*parse.FuncDef{
			{
				Name:       "f",
				ReturnType: parse.IntType{},
				Body:       parse.IntExpr{Value: 1},
			},
			{
				Name:       "f",
				ReturnType: parse.IntType{},
				Body:       parse.IntExpr{Value: 2},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
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

func TestCheckConstructorExpr(t *testing.T) {
	prog := &parse.Program{
		AlgDefs: []*parse.AlgDef{
			{
				Name:     "List",
				TypeVars: []string{"A"},
				ConsDefs: []*parse.ConsDef{
					{
						Name: "lcons",
						Args: []parse.Type{
							parse.TypeVar{Name: "A"},
							parse.AlgType{
								Name: "List",
								Args: []parse.Type{
									parse.TypeVar{Name: "A"},
								},
							},
						},
					},
					{
						Name: "lnil",
					},
				},
			},
		},
		Expr: parse.ConsExpr{
			Name: "lcons",
			Args: []parse.Expr{
				parse.IntExpr{
					Value: 1,
				},
				parse.ConsExpr{
					Name: "lnil",
				},
			},
		},
	}

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckRejectsUnknownConstructor(t *testing.T) {
	prog := &parse.Program{
		Expr: parse.ConsExpr{
			Name: "bad",
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsConstructorArgCountMismatch(t *testing.T) {
	prog := &parse.Program{
		AlgDefs: []*parse.AlgDef{
			{
				Name: "Box",
				ConsDefs: []*parse.ConsDef{
					{
						Name: "box",
						Args: []parse.Type{
							parse.IntType{},
						},
					},
				},
			},
		},
		Expr: parse.ConsExpr{
			Name: "box",
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsConstructorArgTypeMismatch(t *testing.T) {
	prog := &parse.Program{
		AlgDefs: []*parse.AlgDef{
			{
				Name: "Box",
				ConsDefs: []*parse.ConsDef{
					{
						Name: "box",
						Args: []parse.Type{
							parse.IntType{},
						},
					},
				},
			},
		},
		Expr: parse.ConsExpr{
			Name: "box",
			Args: []parse.Expr{
				parse.BoolExpr{
					Value: true,
				},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestApplySubstFuncType(t *testing.T) {
	subst := Subst{
		"A": parse.IntType{},
		"B": parse.BooleanType{},
	}

	r := applySubst(
		parse.FuncType{
			Params: []parse.Type{
				parse.TypeVar{Name: "A"},
			},
			Return: parse.TypeVar{Name: "B"},
		},
		subst,
	)

	want := parse.FuncType{
		Params: []parse.Type{parse.IntType{}},
		Return: parse.BooleanType{},
	}

	if !equalTypes(r, want) {
		t.Fatalf("bad function substitution")
	}
}

func TestUnifyBooleanAndUnit(t *testing.T) {
	if err := unify(parse.BooleanType{}, parse.BooleanType{}, Subst{}); err != nil {
		t.Fatalf("unexpected error")
	}

	if err := unify(parse.UnitType{}, parse.UnitType{}, Subst{}); err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestUnifyFuncType(t *testing.T) {
	a := parse.FuncType{
		Params: []parse.Type{parse.IntType{}},
		Return: parse.BooleanType{},
	}

	b := parse.FuncType{
		Params: []parse.Type{parse.IntType{}},
		Return: parse.BooleanType{},
	}

	if err := unify(a, b, Subst{}); err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestLessThanExpr(t *testing.T) {
	env := NewTypeEnv(nil)

	typ, err := newChecker().checkExpr(
		env,
		parse.OpExpr{
			Op:    "<",
			Left:  parse.IntExpr{Value: 1},
			Right: parse.IntExpr{Value: 2},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if !equalTypes(typ, parse.BooleanType{}) {
		t.Fatalf("expected boolean")
	}
}

func TestEqualEqualExpr(t *testing.T) {
	env := NewTypeEnv(nil)

	typ, err := newChecker().checkExpr(
		env,
		parse.OpExpr{
			Op:    "==",
			Left:  parse.IntExpr{Value: 1},
			Right: parse.IntExpr{Value: 1},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if !equalTypes(typ, parse.BooleanType{}) {
		t.Fatalf("expected boolean")
	}
}

func TestEqualTypes(t *testing.T) {
	tests := []struct {
		name string
		a    parse.Type
		b    parse.Type
	}{
		{
			name: "int",
			a:    parse.IntType{},
			b:    parse.IntType{},
		},
		{
			name: "boolean",
			a:    parse.BooleanType{},
			b:    parse.BooleanType{},
		},
		{
			name: "unit",
			a:    parse.UnitType{},
			b:    parse.UnitType{},
		},
		{
			name: "type var",
			a:    parse.TypeVar{Name: "A"},
			b:    parse.TypeVar{Name: "A"},
		},
		{
			name: "alg type",
			a: parse.AlgType{
				Name: "List",
				Args: []parse.Type{
					parse.IntType{},
				},
			},
			b: parse.AlgType{
				Name: "List",
				Args: []parse.Type{
					parse.IntType{},
				},
			},
		},
		{
			name: "func type",
			a: parse.FuncType{
				Params: []parse.Type{
					parse.IntType{},
				},
				Return: parse.BooleanType{},
			},
			b: parse.FuncType{
				Params: []parse.Type{
					parse.IntType{},
				},
				Return: parse.BooleanType{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !equalTypes(tt.a, tt.b) {
				t.Fatalf("expected types to be equal")
			}
		})
	}
}
