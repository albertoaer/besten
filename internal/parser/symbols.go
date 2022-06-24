package parser

import (
	"errors"
	"fmt"
	"time"

	. "github.com/Besten/internal/runtime"
)

func mathOpInstruction(int_, float_ ICode) []*FunctionSymbol {
	return []*FunctionSymbol{{"none", false, MKInstruction(int_).Fragment(), &Int, []OBJType{Int, Int}},
		{"none", false, MKInstruction(float_).Fragment(), &Dec, []OBJType{Dec, Dec}}}
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
			MKInstruction(v).Fragment(),
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
	return []*FunctionSymbol{{"none", false, MKInstruction(code).Fragment(), &tp, tps}}
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
	}).Fragment(), CloneType(Void), []OBJType{Any}})
	to.AddSymbol("clock", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "clock",
		ArgCount: 0,
		Function: func(args []Object) Object {
			return int(time.Now().UnixMicro())
		},
		Returns: true,
	}).Fragment(), CloneType(Int), []OBJType{}})
	to.AddSymbol("raw", &FunctionSymbol{"none", false, MKInstruction(IFD, EmbeddedFunction{
		Name:     "raw",
		ArgCount: 1,
		Function: func(args []Object) Object {
			return fmt.Sprintf("%v", args[0])
		},
		Returns: true,
	}).Fragment(), CloneType(Str), []OBJType{Any}})
	to.AddSymbol("double", &FunctionSymbol{"none", false, MKInstruction(ITD).Fragment(), CloneType(Dec), []OBJType{Int}})
	to.AddSymbol("int", &FunctionSymbol{"none", false, MKInstruction(DTI).Fragment(), CloneType(Int), []OBJType{Dec}})
	to.AddSymbols("str", multiTypeInstruction(1, Str, map[OBJType]ICode{Str: NOP, Int: IRE, Dec: DRE}))
	to.AddSymbols("len", multiTypeInstruction(1, Int, map[OBJType]ICode{Str: SOS, MapOf(Any): SOM, VecOf(Any): SOV}))
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
	to.AddDynamicSymbol("[]", func(o []OBJType) *FunctionSymbol {
		//Can be implemented by symbol switch but is useful to be this way in order to future custom type generation
		if len(o) == 2 {
			var ins []Instruction
			var ret *OBJType
			if o[0].Primitive() == VECTOR && o[1].Primitive() == INTEGER {
				ins = []Instruction{MKInstruction(PTW), MKInstruction(ACC)}
				ret = CloneType(o[0].Items())
			} else if o[0].Primitive() == MAP && o[1].Primitive() == STRING {
				ins = []Instruction{MKInstruction(ATT), MKInstruction(ACC)}
				ret = CloneType(o[0].Items())
			} else if o[0].Primitive() == STRING && o[1].Primitive() == INTEGER {
				ins = MKInstruction(CAI).Fragment()
				ret = CloneType(Int)
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, ret, o}
		}
		return nil
	})
}

type Variable struct {
	Type       OBJType
	Mutable    bool
	Arg        bool
	Asigned    bool
	Used       bool
	Code       uint
	Dependency *Scope
}

type BlockFlags struct {
	altered    bool
	afterif    bool
	lastifhead struct{ idx, offset int }
	allifskips []struct{ idx, offset int }
}

func CreateFlags() *BlockFlags {
	return &BlockFlags{false, false, struct{ idx, offset int }{-1, -1}, make([]struct{ idx, offset int }, 0)}
}

func (bf *BlockFlags) Edit(nbf BlockFlags) {
	*bf = nbf
	bf.altered = true
}

func (bf *BlockFlags) submitIf(head, skip int) {
	bf.afterif = true
	bf.lastifhead = struct{ idx, offset int }{head, skip}
	bf.altered = true
}

func (bf *BlockFlags) submitIfSkip(idx int) {
	bf.allifskips = append(bf.allifskips, struct{ idx, offset int }{idx, 0})
}

func (bf *BlockFlags) Reset() {
	*bf = *CreateFlags()
}

type Scope struct {
	Variables  map[string]*Variable
	Functions  *FunctionCollection
	Operators  *FunctionCollection
	blockflags *BlockFlags
	ReturnType *OBJType
	Returned   *bool
	parent     *Scope
	varcount   *uint
	argcount   *uint
}

func (s *Scope) CreateVariable(name string, t OBJType, mutable bool, arg bool) {
	s.Variables[name] = &Variable{t, mutable, arg, arg, false, 0, s}
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
		v.Used = true
		return ins, v.Type, nil
	}
	return MKInstruction(NOP), nil, errors.New(fmt.Sprintf("Undefined variable: %s", name))
}

func (s *Scope) SetVariableIns(name string, t OBJType) (Instruction, error) {
	if v, e := s.Variables[name]; e {
		if v.Asigned && !v.Mutable {
			return MKInstruction(NOP), errors.New(fmt.Sprintf("Trying to reassign a constant: %s", name))
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
		blockflags: CreateFlags(),
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
		blockflags: CreateFlags(),
		ReturnType: returnt, Returned: returned, parent: s, varcount: vc, argcount: ac}
	for k, v := range s.Variables {
		ns.Variables[k] = v
	}
	return ns
}

func (s *Scope) CheckClose() error {
	for name, v := range s.Variables {
		if v.Dependency == s && !v.Used && !v.Arg {
			vart := "Variable"
			if !v.Mutable {
				vart = "Constant"
			}
			return errors.New(fmt.Sprintf("%s not used: %s", vart, name))
		}
	}
	return nil
}

func (s *Scope) Close() (*Scope, error) {
	return s.parent, s.CheckClose()
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}
