package parse

type Program struct {
	AlgDefs  []*AlgDef
	FuncDefs []*FuncDef
	Expr     Expr
}

type AlgDef struct {
	Name     string
	TypeVars []string
	ConsDefs []*ConsDef
}

type ConsDef struct {
	Name string
	Args []Type
}

type FuncDef struct {
	Name       string
	TypeVars   []string
	Params     []*Param
	ReturnType Type
	Body       Expr
}

type Param struct {
	Type Type
	Name string
}

type Case struct {
	Pattern Pattern
	Expr    Expr
}

type Type interface {
	isType()
}

type Expr interface {
	isExpr()
}

type Stmt interface {
	isStmt()
}

type Pattern interface {
	isPattern()
}

type IntType struct{}

type UnitType struct{}

type BooleanType struct{}

type FuncType struct {
	Params []Type
	Return Type
}

type AlgType struct {
	Name string
	Args []Type
}

type TypeVar struct {
	Name string
}

func (IntType) isType()     {}
func (UnitType) isType()    {}
func (BooleanType) isType() {}
func (FuncType) isType()    {}
func (AlgType) isType()     {}
func (TypeVar) isType()     {}

type VarExpr struct {
	Name string
}

type IntExpr struct {
	Value int
}

type UnitExpr struct{}

type BoolExpr struct {
	Value bool
}

type PrintlnExpr struct {
	Expr Expr
}

type OpExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

type LambdaExpr struct {
	Params []*Param
	Body   Expr
}

type CallHOFExpr struct {
	Fn   Expr
	Args []Expr
}

type CallExpr struct {
	Name string
	Args []Expr
}

type BlockExpr struct {
	Stmts []Stmt
	Expr  Expr
}

type ConsExpr struct {
	Name string
	Args []Expr
}

type MatchExpr struct {
	Expr  Expr
	Cases []*Case
}

func (VarExpr) isExpr()     {}
func (IntExpr) isExpr()     {}
func (UnitExpr) isExpr()    {}
func (BoolExpr) isExpr()    {}
func (PrintlnExpr) isExpr() {}
func (OpExpr) isExpr()      {}
func (LambdaExpr) isExpr()  {}
func (CallHOFExpr) isExpr() {}
func (CallExpr) isExpr()    {}
func (BlockExpr) isExpr()   {}
func (ConsExpr) isExpr()    {}
func (MatchExpr) isExpr()   {}

type ValStmt struct {
	Name string
	Expr Expr
}

type VarStmt struct {
	Type Type
	Name string
	Expr Expr
}

type AssignStmt struct {
	Name string
	Expr Expr
}

func (ValStmt) isStmt()    {}
func (VarStmt) isStmt()    {}
func (AssignStmt) isStmt() {}

type VarPattern struct {
	Name string
}

type WildcardPattern struct{}

type ConsPattern struct {
	Name     string
	Patterns []Pattern
}

func (VarPattern) isPattern()      {}
func (WildcardPattern) isPattern() {}
func (ConsPattern) isPattern()     {}
