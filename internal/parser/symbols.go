package parser

import (
	"errors"
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
		Function: func(args []Object) Object {
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
	to.AddSymbols("!", wrapOpInstruction(NOT, Int, true))
	to.AddSymbols("&", wrapOpInstruction(AND, Int, false))
	to.AddSymbols("|", wrapOpInstruction(OR, Int, false))
	to.AddSymbols("^", wrapOpInstruction(XOR, Int, false))
	to.AddSymbols("!", wrapOpInstruction(NOT, Bool, true))
	to.AddSymbols("&&", wrapOpInstruction(AND, Bool, false))
	to.AddSymbols("||", wrapOpInstruction(OR, Bool, false))
	to.AddSymbols("^^", wrapOpInstruction(XOR, Bool, false))
	to.AddSymbols("==", multiTypeInstruction(2, Bool, map[OBJType]ICode{Str: EQS, Int: EQI, Dec: EQD}))
	to.AddSymbols("<", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: ILE, Dec: DLE}))
	to.AddSymbols(">", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: IGR, Dec: DGR}))
	to.AddSymbols("<=", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: ILQ, Dec: DLQ}))
	to.AddSymbols(">=", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: IGQ, Dec: DGQ}))
}

type Variable struct {
	Type    OBJType
	Mutable bool
	Arg     bool
	Asigned bool
	Code    uint
}

type Scope struct {
	Variables  map[string]*Variable
	Functions  *FunctionCollection
	Operators  *FunctionCollection
	ReturnType *OBJType
	Returned   *bool
	parent     *Scope
	varcount   *uint
	argcount   *uint
}

func (s *Scope) CreateVariable(name string, t OBJType, mutable bool, arg bool) {
	s.Variables[name] = &Variable{t, mutable, arg, arg, 0}
	if arg {
		s.Variables[name].Code = *s.argcount
		*s.argcount++
	} else {
		s.Variables[name].Code = *s.varcount
		*s.varcount++
	}
}

func (s *Scope) GetVariableIns(name string) (Instruction, OBJType, error) {
	if v, e := s.Variables[name]; e {
		if !v.Asigned {
			return MKInstruction(NOP), nil, errors.New(fmt.Sprintf("A value has not been set for: %s", name))
		}
		var ins Instruction
		if v.Arg {
			ins = MKInstruction(LEI, int(v.Code))
		} else {
			ins = MKInstruction(LLI, int(v.Code))
		}
		return ins, v.Type, nil
	}
	return MKInstruction(NOP), nil, errors.New(fmt.Sprintf("Undefined variable: %s", name))
}

func (s *Scope) SetVariableIns(name string, t OBJType) (Instruction, error) {
	if v, e := s.Variables[name]; e {
		if v.Asigned && !v.Mutable {
			return MKInstruction(NOP), errors.New(fmt.Sprintf("Trying to reasign a constant: %s", name))
		}
		if !CompareTypes(v.Type, t) {
			return MKInstruction(NOP), errors.New("Invalid type")
		}
		var ins Instruction
		if v.Arg {
			ins = MKInstruction(SEI, int(v.Code))
		} else {
			ins = MKInstruction(SLI, int(v.Code))
		}
		/*
			We asume has been asigned if
			it has been requested de assignment instruction
		*/
		v.Asigned = true
		return ins, nil
	}
	return MKInstruction(NOP), errors.New(fmt.Sprintf("Undefined variable: %s", name))
}

func NewScope() *Scope {
	rptr := Void
	r := false
	var vc, ac uint = 0, 0
	return &Scope{Variables: make(map[string]*Variable),
		Functions:  NewFunctionCollection(),
		Operators:  NewFunctionCollection(),
		ReturnType: &rptr, Returned: &r, parent: nil, varcount: &vc, argcount: &ac}
}

func (s *Scope) Open(fnscope bool) *Scope {
	returnt := s.ReturnType
	returned := s.Returned
	vc, ac := s.varcount, s.argcount
	if fnscope {
		rptr := Void
		returnt = &rptr
		isret := false
		returned = &isret
		var vcv, acv uint = 0, 0
		vc, ac = &vcv, &acv
	}
	ns := &Scope{Variables: make(map[string]*Variable),
		Functions:  s.Functions.Fork(),
		Operators:  s.Operators.Fork(),
		ReturnType: returnt, Returned: returned, parent: s, varcount: vc, argcount: ac}
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
