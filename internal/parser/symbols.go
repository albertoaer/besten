package parser

import (
	"github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func mathOpInstruction(int_, float_ ICode) []FunctionSymbol {
	return []FunctionSymbol{{"none", MKInstruction(int_), Int, []OBJType{Int, Int}}, {"none", MKInstruction(float_), Dec, []OBJType{Dec, Dec}}}
}

func multiTypeOpInstruction(v map[OBJType]ICode) []FunctionSymbol {
	ret := make([]FunctionSymbol, 0)
	return ret
}

func wrapOpInstruction(code ICode, tp OBJType, unary bool) []FunctionSymbol {
	tps := []OBJType{tp, tp}
	if unary {
		tps = []OBJType{tp}
	}
	return []FunctionSymbol{{"none", MKInstruction(code), tp, tps}}
}

func injectBuiltinOperators(to map[HeaderAlias][]FunctionSymbol) {
	to[HeaderAlias{"+", 2}] = mathOpInstruction(ADD, ADDF)
	to[HeaderAlias{"-", 2}] = mathOpInstruction(SUB, SUBF)
	to[HeaderAlias{"*", 2}] = mathOpInstruction(MUL, MULF)
	to[HeaderAlias{"/", 2}] = mathOpInstruction(DIV, DIVF)
	to[HeaderAlias{"==", 2}] = multiTypeOpInstruction(map[OBJType]ICode{Str: EQS, Int: EQI, Dec: EQD})
	to[HeaderAlias{"!", 1}] = wrapOpInstruction(NOT, Int, true)
	to[HeaderAlias{"&&", 2}] = wrapOpInstruction(AND, Int, false)
	to[HeaderAlias{"||", 2}] = wrapOpInstruction(OR, Int, false)
	to[HeaderAlias{"^^", 2}] = wrapOpInstruction(XOR, Int, false)
}

type Variable struct {
	Type    OBJType
	Mutable bool
}

type FunctionTemplate struct {
	Args     []string
	Varargs  bool
	Children []lexer.Block
}

type FunctionSymbol struct {
	CName  string
	Call   Instruction
	Return OBJType
	Args   []OBJType
}

type HeaderAlias struct {
	string //name
	int    //args
}

type Scope struct {
	Variables         map[string]*Variable
	FunctionTemplates map[HeaderAlias]FunctionTemplate
	OpTemplates       map[HeaderAlias]FunctionTemplate
	FunctionSymbols   map[HeaderAlias][]FunctionSymbol
	OpSymbols         map[HeaderAlias][]FunctionSymbol
	ReturnType        OBJType
	parent            *Scope
}

func NewScope() *Scope {
	return &Scope{make(map[string]*Variable), make(map[HeaderAlias]FunctionTemplate),
		make(map[HeaderAlias]FunctionTemplate), make(map[HeaderAlias][]FunctionSymbol),
		make(map[HeaderAlias][]FunctionSymbol), Void, nil}
}

func (s *Scope) Merge(scp *Scope) {
	for k, v := range scp.Variables {
		s.Variables[k] = v
	}
	for k, v := range scp.FunctionTemplates {
		s.FunctionTemplates[k] = v
	}
	for k, v := range scp.OpTemplates {
		s.OpTemplates[k] = v
	}
	for k, v := range scp.FunctionSymbols {
		nv := v
		s.FunctionSymbols[k] = nv
	}
}

func (s *Scope) Open() *Scope {
	ns := NewScope()
	ns.parent = s
	return ns
}

func (s *Scope) Close() *Scope {
	return s.parent
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}
