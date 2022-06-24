package parser

import (
	"github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type OBJType uint8

const (
	VOID    OBJType = 0
	NULL            = 1
	INTEGER         = 2
	DECIMAL         = 3
	STRING          = 4
	VECTOR          = 5
	MAP             = 6
	ALIAS           = 7
	ANY             = 8
)

func mathOpInstruction(int_, float_ ICode) []FunctionSymbol {
	return []FunctionSymbol{{MKInstruction(int_), INTEGER, []OBJType{INTEGER, INTEGER}}, {MKInstruction(float_), DECIMAL, []OBJType{DECIMAL, DECIMAL}}}
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
	return []FunctionSymbol{{MKInstruction(code), tp, tps}}
}

func injectBuiltinOperators(to map[HeaderAlias][]FunctionSymbol) {
	to[HeaderAlias{"+", 2}] = mathOpInstruction(ADD, ADDF)
	to[HeaderAlias{"-", 2}] = mathOpInstruction(SUB, SUBF)
	to[HeaderAlias{"*", 2}] = mathOpInstruction(MUL, MULF)
	to[HeaderAlias{"/", 2}] = mathOpInstruction(DIV, DIVF)
	to[HeaderAlias{"==", 2}] = multiTypeOpInstruction(map[OBJType]ICode{STRING: EQS, INTEGER: EQI, DECIMAL: EQD})
	to[HeaderAlias{"!", 1}] = wrapOpInstruction(NOT, INTEGER, true)
	to[HeaderAlias{"&&", 2}] = wrapOpInstruction(AND, INTEGER, false)
	to[HeaderAlias{"||", 2}] = wrapOpInstruction(OR, INTEGER, false)
	to[HeaderAlias{"^^", 2}] = wrapOpInstruction(XOR, INTEGER, false)
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
		make(map[HeaderAlias][]FunctionSymbol), VOID, nil}
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
