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
