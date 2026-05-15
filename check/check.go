package check

import "dada/parse"

type Error string

func (e Error) Error() string {
	return string(e)
}

type TypeEnv struct {
	parent *TypeEnv
	vars   map[string]parse.Type
}

func NewTypeEnv(parent *TypeEnv) *TypeEnv {
	return &TypeEnv{
		parent: parent,
		vars:   map[string]parse.Type{},
	}
}

func (e *TypeEnv) Bind(name string, typ parse.Type) {
	e.vars[name] = typ
}

func (e *TypeEnv) Lookup(name string) (parse.Type, bool) {

	for cur := e; cur != nil; cur = cur.parent {
		if t, ok := cur.vars[name]; ok {
			return t, true
		}
	}

	return nil, false
}

type FuncSig struct {
	TypeVars []string
	Params   []parse.Type
	Return   parse.Type
}

type ConsSig struct {
	TypeVars []string
	Args     []parse.Type
	Result   parse.Type
}

type Checker struct {
	funcs map[string]*FuncSig
	cons  map[string]*ConsSig
	algs  map[string]*parse.AlgDef
}

func newChecker() *Checker {
	return &Checker{
		funcs: map[string]*FuncSig{},
		cons:  map[string]*ConsSig{},
		algs:  map[string]*parse.AlgDef{},
	}
}

func Check(program *parse.Program) error {
	if program == nil {
		return Error("nil program")
	}

	c := newChecker()
	return c.checkProgram(program)
}

func (c *Checker) checkProgram(program *parse.Program) error {
	for _, alg := range program.AlgDefs {
		if err := c.registerAlgDef(alg); err != nil {
			return err
		}
	}

	for _, fn := range program.FuncDefs {
		if err := c.registerFuncDef(fn); err != nil {
			return err
		}
	}

	return nil
}

func (c *Checker) registerAlgDef(def *parse.AlgDef) error {
	if _, exists := c.algs[def.Name]; exists {
		return Error("duplicate algdef")
	}

	c.algs[def.Name] = def

	return nil
}

func (c *Checker) registerFuncDef(fn *parse.FuncDef) error {
	if _, exists := c.funcs[fn.Name]; exists {
		return Error("duplicate function")
	}

	var params []parse.Type

	for _, p := range fn.Params {
		params = append(params, p.Type)
	}

	c.funcs[fn.Name] = &FuncSig{
		TypeVars: fn.TypeVars,
		Params:   params,
		Return:   fn.ReturnType,
	}

	return nil
}

func equalTypes(a, b parse.Type) bool {
	switch ta := a.(type) {

	case parse.IntType:
		_, ok := b.(parse.IntType)
		return ok

	case parse.BooleanType:
		_, ok := b.(parse.BooleanType)
		return ok

	case parse.UnitType:
		_, ok := b.(parse.UnitType)
		return ok

	case parse.TypeVar:
		tb, ok := b.(parse.TypeVar)
		return ok && ta.Name == tb.Name

	case parse.AlgType:
		tb, ok := b.(parse.AlgType)

		if !ok {
			return false
		}

		if ta.Name != tb.Name {
			return false
		}

		if len(ta.Args) != len(tb.Args) {
			return false
		}

		for i := range ta.Args {
			if !equalTypes(
				ta.Args[i],
				tb.Args[i],
			) {
				return false
			}
		}

		return true

	case parse.FuncType:
		tb, ok := b.(parse.FuncType)

		if !ok {
			return false
		}

		if len(ta.Params) != len(tb.Params) {
			return false
		}

		for i := range ta.Params {
			if !equalTypes(
				ta.Params[i],
				tb.Params[i],
			) {
				return false
			}
		}

		return equalTypes(
			ta.Return,
			tb.Return,
		)
	}

	return false
}

type Subst map[string]parse.Type

var freshCounter int

func freshTypeVar() parse.TypeVar {
	name := "__t"

	name += string(rune('0' + freshCounter))

	freshCounter++

	return parse.TypeVar{
		Name: name,
	}
}

func applySubst(t parse.Type, subst Subst) parse.Type {
	switch tt := t.(type) {

	case parse.TypeVar:
		if r, ok := subst[tt.Name]; ok {
			return r
		}

		return tt

	case parse.FuncType:
		var params []parse.Type

		for _, p := range tt.Params {
			params = append(
				params,
				applySubst(p, subst),
			)
		}

		return parse.FuncType{
			Params: params,
			Return: applySubst(
				tt.Return,
				subst,
			),
		}

	case parse.AlgType:
		var args []parse.Type

		for _, a := range tt.Args {
			args = append(
				args,
				applySubst(a, subst),
			)
		}

		return parse.AlgType{
			Name: tt.Name,
			Args: args,
		}
	}

	return t
}

func instantiateFuncSig(sig *FuncSig) *FuncSig {
	subst := Subst{}

	for _, tv := range sig.TypeVars {
		subst[tv] = freshTypeVar()
	}

	var params []parse.Type

	for _, p := range sig.Params {
		params = append(
			params,
			applySubst(p, subst),
		)
	}

	return &FuncSig{
		Params: params,
		Return: applySubst(
			sig.Return,
			subst,
		),
	}
}

func unify(a parse.Type, b parse.Type, subst Subst) error {
	a = applySubst(a, subst)
	b = applySubst(b, subst)

	switch ta := a.(type) {

	case parse.TypeVar:
		subst[ta.Name] = b
		return nil
	}

	switch tb := b.(type) {

	case parse.TypeVar:
		subst[tb.Name] = a
		return nil
	}

	switch ta := a.(type) {

	case parse.IntType:
		if _, ok := b.(parse.IntType); !ok {
			return Error("cannot unify")
		}

	case parse.BooleanType:
		if _, ok := b.(parse.BooleanType); !ok {
			return Error("cannot unify")
		}

	case parse.UnitType:
		if _, ok := b.(parse.UnitType); !ok {
			return Error("cannot unify")
		}

	case parse.FuncType:
		tb, ok := b.(parse.FuncType)

		if !ok {
			return Error("cannot unify")
		}

		if len(ta.Params) != len(tb.Params) {
			return Error("cannot unify")
		}

		for i := range ta.Params {
			if err := unify(
				ta.Params[i],
				tb.Params[i],
				subst,
			); err != nil {
				return err
			}
		}

		return unify(
			ta.Return,
			tb.Return,
			subst,
		)

	case parse.AlgType:
		tb, ok := b.(parse.AlgType)

		if !ok {
			return Error("cannot unify")
		}

		if ta.Name != tb.Name {
			return Error("cannot unify")
		}

		if len(ta.Args) != len(tb.Args) {
			return Error("cannot unify")
		}

		for i := range ta.Args {
			if err := unify(
				ta.Args[i],
				tb.Args[i],
				subst,
			); err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}

func (c *Checker) checkExpr(env *TypeEnv, expr parse.Expr) (parse.Type, error) {
	switch e := expr.(type) {

	case parse.IntExpr:
		return parse.IntType{}, nil

	case parse.BoolExpr:
		return parse.BooleanType{}, nil

	case parse.UnitExpr:
		return parse.UnitType{}, nil

	case parse.VarExpr:
		t, ok := env.Lookup(e.Name)

		if !ok {
			return nil, Error("unbound variable")
		}

		return t, nil

	case parse.PrintlnExpr:
		_, err := c.checkExpr(env, e.Expr)

		if err != nil {
			return nil, err
		}

		return parse.UnitType{}, nil

	case parse.OpExpr:
		return c.checkOpExpr(env, e)
	}

	return nil, Error("unsupported expression")
}

func (c *Checker) checkOpExpr(env *TypeEnv, expr parse.OpExpr) (parse.Type, error) {
	left, err := c.checkExpr(env, expr.Left)

	if err != nil {
		return nil, err
	}

	right, err := c.checkExpr(env, expr.Right)

	if err != nil {
		return nil, err
	}

	switch expr.Op {

	case "+", "-", "*", "/":
		if !equalTypes(left, parse.IntType{}) {
			return nil, Error("expected Int")
		}

		if !equalTypes(right, parse.IntType{}) {
			return nil, Error("expected Int")
		}

		return parse.IntType{}, nil

	case "<":
		if !equalTypes(left, parse.IntType{}) {
			return nil, Error("expected Int")
		}

		if !equalTypes(right, parse.IntType{}) {
			return nil, Error("expected Int")
		}

		return parse.BooleanType{}, nil

	case "==":
		subst := Subst{}

		if err := unify(left, right, subst); err != nil {
			return nil, err
		}

		return parse.BooleanType{}, nil
	}

	return nil, Error("unknown operator")
}
