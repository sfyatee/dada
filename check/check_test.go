package check

import (
	"testing"

	"dada/parse"
)

func intExpr(value int) parse.IntExpr {
	return parse.IntExpr{Value: value}
}

func boolExpr(value bool) parse.BoolExpr {
	return parse.BoolExpr{Value: value}
}

func varExpr(name string) parse.VarExpr {
	return parse.VarExpr{Name: name}
}

func expectCheckOK(t *testing.T, prog *parse.Program) {
	t.Helper()

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func expectCheckError(t *testing.T, prog *parse.Program) {
	t.Helper()

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func expectExprOK(t *testing.T, expr parse.Expr, want parse.Type) {
	t.Helper()

	typ, err := newChecker().checkExpr(NewTypeEnv(nil), expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !equalTypes(typ, want) {
		t.Fatalf("wrong type")
	}
}

func expectExprError(t *testing.T, expr parse.Expr) {
	t.Helper()

	_, err := newChecker().checkExpr(NewTypeEnv(nil), expr)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func listAlgDef() *parse.AlgDef {
	return &parse.AlgDef{
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
	}
}

func intListExpr() parse.Expr {
	return parse.ConsExpr{
		Name: "lcons",
		Args: []parse.Expr{
			parse.IntExpr{Value: 1},
			parse.ConsExpr{Name: "lnil"},
		},
	}
}

func TestCheckerExists(t *testing.T) {
	c := newChecker()

	if c == nil {
		t.Fatalf("expected checker")
	}
}

func TestDuplicateFunctionFails(t *testing.T) {
	expectCheckError(t, &parse.Program{
		FuncDefs: []*parse.FuncDef{
			{
				Name:       "f",
				ReturnType: parse.IntType{},
				Body:       intExpr(1),
			},
			{
				Name:       "f",
				ReturnType: parse.IntType{},
				Body:       intExpr(2),
			},
		},
	})
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

func TestCheckOpExprs(t *testing.T) {
	tests := []struct {
		name string
		expr parse.OpExpr
		want parse.Type
	}{
		{
			name: "add",
			expr: parse.OpExpr{Op: "+", Left: intExpr(1), Right: intExpr(2)},
			want: parse.IntType{},
		},
		{
			name: "less than",
			expr: parse.OpExpr{Op: "<", Left: intExpr(1), Right: intExpr(2)},
			want: parse.BooleanType{},
		},
		{
			name: "equal equal",
			expr: parse.OpExpr{Op: "==", Left: intExpr(1), Right: intExpr(1)},
			want: parse.BooleanType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectExprOK(t, tt.expr, tt.want)
		})
	}
}

func TestCheckRejectsBadOpExprs(t *testing.T) {
	tests := []parse.OpExpr{
		{
			Op:    "+",
			Left:  boolExpr(true),
			Right: intExpr(2),
		},
		{
			Op:    "+",
			Left:  intExpr(1),
			Right: boolExpr(true),
		},
		{
			Op:    "<",
			Left:  boolExpr(true),
			Right: intExpr(1),
		},
		{
			Op:    "<",
			Left:  intExpr(1),
			Right: boolExpr(true),
		},
		{
			Op:    "==",
			Left:  intExpr(1),
			Right: boolExpr(true),
		},
		{
			Op:    "%",
			Left:  intExpr(1),
			Right: intExpr(2),
		},
	}

	for _, expr := range tests {
		t.Run(expr.Op, func(t *testing.T) {
			expectExprError(t, expr)
		})
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
	expectCheckError(t, &parse.Program{
		FuncDefs: []*parse.FuncDef{
			{
				Name:       "bad",
				ReturnType: parse.BooleanType{},
				Body:       intExpr(1),
			},
		},
	})
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
	expectCheckError(t, &parse.Program{
		Expr: parse.ConsExpr{Name: "bad"},
	})
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

func TestCheckMatchExpr(t *testing.T) {
	prog := &parse.Program{
		AlgDefs: []*parse.AlgDef{
			listAlgDef(),
		},
		Expr: parse.MatchExpr{
			Expr: intListExpr(),
			Cases: []*parse.Case{
				{
					Pattern: parse.ConsPattern{
						Name: "lcons",
						Patterns: []parse.Pattern{
							parse.VarPattern{Name: "head"},
							parse.VarPattern{Name: "tail"},
						},
					},
					Expr: parse.VarExpr{Name: "head"},
				},
				{
					Pattern: parse.ConsPattern{Name: "lnil"},
					Expr:    parse.IntExpr{Value: 0},
				},
			},
		},
	}

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckRejectsNonExhaustiveMatch(t *testing.T) {
	prog := &parse.Program{
		AlgDefs: []*parse.AlgDef{
			listAlgDef(),
		},
		Expr: parse.MatchExpr{
			Expr: intListExpr(),
			Cases: []*parse.Case{
				{
					Pattern: parse.ConsPattern{
						Name: "lnil",
					},
					Expr: parse.IntExpr{Value: 0},
				},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckWildcardPatternIsExhaustive(t *testing.T) {
	prog := &parse.Program{
		Expr: parse.MatchExpr{
			Expr: parse.IntExpr{Value: 1},
			Cases: []*parse.Case{
				{
					Pattern: parse.WildcardPattern{},
					Expr:    parse.IntExpr{Value: 0},
				},
			},
		},
	}

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckVarPatternIsExhaustive(t *testing.T) {
	prog := &parse.Program{
		Expr: parse.MatchExpr{
			Expr: parse.IntExpr{Value: 1},
			Cases: []*parse.Case{
				{
					Pattern: parse.VarPattern{Name: "x"},
					Expr:    parse.VarExpr{Name: "x"},
				},
			},
		},
	}

	if err := Check(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckRejectsConstructorPatternTypeMismatch(t *testing.T) {
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
		Expr: parse.MatchExpr{
			Expr: parse.BoolExpr{
				Value: true,
			},
			Cases: []*parse.Case{
				{
					Pattern: parse.ConsPattern{
						Name: "box",
						Patterns: []parse.Pattern{
							parse.VarPattern{Name: "x"},
						},
					},
					Expr: parse.IntExpr{
						Value: 0,
					},
				},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsNonExhaustivePrimitiveMatch(t *testing.T) {
	prog := &parse.Program{
		Expr: parse.MatchExpr{
			Expr: parse.IntExpr{
				Value: 1,
			},
			Cases: []*parse.Case{
				{
					Pattern: parse.ConsPattern{
						Name: "bad",
					},
					Expr: parse.IntExpr{
						Value: 0,
					},
				},
			},
		},
	}

	if err := Check(prog); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsUnknownAlgTypeInExhaustiveCheck(t *testing.T) {
	c := newChecker()

	err := c.checkExhaustive(
		parse.AlgType{
			Name: "Missing",
		},
		[]*parse.Case{},
	)

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckRejectsUnboundVariable(t *testing.T) {
	expectExprError(t, varExpr("missing"))
}

func TestCheckPrintlnExpr(t *testing.T) {
	expectExprOK(
		t,
		parse.PrintlnExpr{
			Expr: intExpr(1),
		},
		parse.UnitType{},
	)
}

func TestCheckRejectsBadPrintlnExpr(t *testing.T) {
	expectExprError(
		t,
		parse.PrintlnExpr{
			Expr: varExpr("missing"),
		},
	)
}
