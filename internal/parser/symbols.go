package parser

import (
	"fmt"

	. "github.com/Besten/internal/runtime"
)

func mathOpInstruction(int_, float_ ICode) []*FunctionSymbol {
	return []*FunctionSymbol{{"none", false, MKInstruction(int_), &Int, []OBJType{Int, Int}},
		{"none", false, MKInstruction(float_), &Dec, []OBJType{Dec, Dec}}}
}

func multiTypeInstruction(length int, returntype OBJType, matches map[OBJType]ICode) []*FunctionSymbol {
	ret := make([]*FunctionSymbol, 0)
	for k, v := range matches {
		tp := make([]OBJType, length)
		for i := range tp {
			tp[i] = k
		}
		ret = append(ret, &FunctionSymbol{
			"none",
			false,
			MKInstruction(v),
			&returntype,
			tp,
		})
	}
	return ret
}

func wrapOpInstruction(code ICode, tp OBJType, unary bool) []*FunctionSymbol {
	tps := []OBJType{tp, tp}
	if unary {
		tps = []OBJType{tp}
	}
	return []*FunctionSymbol{{"none", false, MKInstruction(code), &tp, tps}}
}

func injectBuiltinFunctions(to *FunctionCollection) {
	to.AddSymbol("print", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "print",
		ArgCount: 1,
		Function: func(args ...Object) Object {
			fmt.Println(args[0])
			return nil
		},
		Returns: false,
	}), &Void, []OBJType{Any}})
	to.AddSymbols("str", multiTypeInstruction(1, Str, map[OBJType]ICode{Str: NOP, Int: IRE, Dec: DRE}))
}

func injectBuiltinOperators(to *FunctionCollection) {
	to.AddSymbols("+", mathOpInstruction(ADD, ADDF))
	to.AddSymbols("+", wrapOpInstruction(CCS, Str, false))
	to.AddSymbols("-", mathOpInstruction(SUB, SUBF))
	to.AddSymbols("*", mathOpInstruction(MUL, MULF))
	to.AddSymbols("/", mathOpInstruction(DIV, DIVF))
	to.AddSymbols("%", wrapOpInstruction(MOD, Int, false))
	to.AddSymbols("==", multiTypeInstruction(2, Int, map[OBJType]ICode{Str: EQS, Int: EQI, Dec: EQD}))
	to.AddSymbols("!", wrapOpInstruction(NOT, Int, true))
	to.AddSymbols("&&", wrapOpInstruction(AND, Int, false))
	to.AddSymbols("||", wrapOpInstruction(OR, Int, false))
	to.AddSymbols("^^", wrapOpInstruction(XOR, Int, false))
}

type Variable struct {
	Type    OBJType
	Mutable bool
}

type Scope struct {
	Variables  map[string]*Variable //TODO: Auxiliar module like the function collection for functions
	Functions  *FunctionCollection
	Operators  *FunctionCollection
	ReturnType *OBJType
	parent     *Scope
}

func NewScope() *Scope {
	return &Scope{Variables: make(map[string]*Variable),
		Functions:  NewFunctionCollection(),
		Operators:  NewFunctionCollection(),
		ReturnType: &Void, parent: nil}
}

func (s *Scope) Open() *Scope {
	ns := &Scope{Variables: make(map[string]*Variable),
		Functions:  s.Functions.Fork(),
		Operators:  s.Operators.Fork(),
		ReturnType: &Void, parent: s}
	for k, v := range s.Variables {
		ns.Variables[k] = v
	}
	return ns
}

func (s *Scope) Close() *Scope {
	return s.parent
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}
