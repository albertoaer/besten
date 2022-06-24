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
	to.AddSymbol("print", &FunctionSymbol{"none", true, MKInstruction(IFD, EmbeddedFunction{
		Name:     "print",
		ArgCount: 1,
		Function: func(args []Object) Object {
			v := *args[0].(VecT)
			for i, e := range v {
				if i == len(v)-1 {
					fmt.Println(e)
				} else {
					fmt.Print(e)
				}
			}
			return nil
		},
		Returns: false,
	}).Fragment(), CloneType(Void), []OBJType{VecOf(Any)}})
	to.AddSymbol("puts", &FunctionSymbol{"none", true, MKInstruction(IFD, EmbeddedFunction{
		Name:     "puts",
		ArgCount: 1,
		Function: func(args []Object) Object {
			for _, e := range *args[0].(VecT) {
				fmt.Print(e)
			}
			return nil
		},
		Returns: false,
	}).Fragment(), CloneType(Void), []OBJType{VecOf(Any)}})
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
	to.AddDynamicSymbol("stref", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			return &FunctionSymbol{"none", false, []Instruction{MKInstruction(POP), MKInstruction(PSH, Repr(o[0]))}, CloneType(Str), o}
		}
		return nil
	})
	to.AddSymbol("dec", &FunctionSymbol{"none", false, MKInstruction(ITD).Fragment(), CloneType(Dec), []OBJType{Int}})
	to.AddSymbol("int", &FunctionSymbol{"none", false, MKInstruction(DTI).Fragment(), CloneType(Int), []OBJType{Dec}})
	to.AddSymbol("int", &FunctionSymbol{"none", false, []Instruction{}, CloneType(Int), []OBJType{Bool}})
	to.AddSymbols("str", multiTypeInstruction(1, Str, map[OBJType]ICode{Str: NOP, Int: IRE, Dec: DRE}))
	to.AddSymbols("len", multiTypeInstruction(1, Int, map[OBJType]ICode{Str: SOS, MapOf(Any): SOM, VecOf(Any): SOV}))
	to.AddDynamicSymbol("vec", func(o []OBJType) *FunctionSymbol {
		if len(o) > 0 {
			ret := VecOf(o[0])
			return &FunctionSymbol{"none", true, []Instruction{}, &ret, []OBJType{VecOf(o[0])}}
		}
		return nil
	})
	to.AddDynamicSymbol("pop", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 && o[0].Primitive() == VECTOR {
			ret := o[0].Items()
			return &FunctionSymbol{"none", false, []Instruction{MKInstruction(PFV)}, &ret, o}
		}
		return nil
	})
	to.AddDynamicSymbol("setbykey", func(o []OBJType) *FunctionSymbol {
		if len(o) == 3 {
			var ins []Instruction
			if o[2].Primitive() == VECTOR && o[1].Primitive() == INTEGER && CompareTypes(o[0], o[2].Items()) {
				ins = []Instruction{MKInstruction(SVI)}
			} else if o[2].Primitive() == MAP && o[1].Primitive() == STRING && CompareTypes(o[0], o[2].Items()) {
				ins = []Instruction{MKInstruction(ATT)}
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, CloneType(Void), o}
		}
		return nil
	})
	to.AddDynamicSymbol("callfn", func(o []OBJType) *FunctionSymbol {
		if o[0].Primitive() == FUNCTION {
			fn := o[0].(*FunctionType)
			if !CompareArrayOfTypes(o[1:], fn.args) {
				return nil
			}
			return &FunctionSymbol{"none", false, MKInstruction(CLL).Fragment(), CloneType(fn.ret), o}
		}
		return nil
	})
	to.AddDynamicSymbol("callmapfn", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 && o[0].Primitive() == FUNCTION && o[1].Primitive() == TUPLE {
			fn := o[0].(*FunctionType)
			if !CompareArrayOfTypes(o[1].FixedItems(), fn.args) {
				return nil
			}
			return &FunctionSymbol{"none", false, MKInstruction(CLX).Fragment(), CloneType(fn.ret), o}
		}
		return nil
	})
	to.AddDynamicSymbol("argsfn", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 && o[0].Primitive() == FUNCTION {
			fn := o[0].(*FunctionType)
			tp := TupleOf(fn.args)
			tpins, e := tp.Create()
			if e != nil {
				return nil
			}
			return &FunctionSymbol{"none", false, tpins, &tp, o}
		}
		return nil
	})
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
	to.AddSymbols("!", wrapOpInstruction(NOTB, Bool, true))
	to.AddSymbols("&&", wrapOpInstruction(ANDB, Bool, false))
	to.AddSymbols("||", wrapOpInstruction(ORB, Bool, false))
	to.AddSymbols("^^", wrapOpInstruction(XORB, Bool, false))
	to.AddSymbols("==", multiTypeInstruction(2, Bool, map[OBJType]ICode{Str: EQS, Int: EQI, Dec: EQD, Bool: EQI}))
	to.AddSymbols("!=", multiTypeInstruction(2, Bool, map[OBJType]ICode{Str: NQS, Int: NQI, Dec: NQD, Bool: NQI}))
	to.AddSymbols("<", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: ILE, Dec: DLE}))
	to.AddSymbols(">", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: IGR, Dec: DGR}))
	to.AddSymbols("<=", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: ILQ, Dec: DLQ}))
	to.AddSymbols(">=", multiTypeInstruction(2, Bool, map[OBJType]ICode{Int: IGQ, Dec: DGQ}))
	to.AddDynamicSymbol("[]", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 {
			var ins []Instruction
			var ret *OBJType
			if o[0].Primitive() == VECTOR && o[1].Primitive() == INTEGER {
				ins = []Instruction{MKInstruction(ACC)}
				ret = CloneType(o[0].Items())
			} else if o[0].Primitive() == MAP && o[1].Primitive() == STRING {
				ins = []Instruction{MKInstruction(PRP)}
				ret = CloneType(TupleOf([]OBJType{o[0].Items(), Bool}))
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
	to.AddDynamicSymbol("->", func(o []OBJType) *FunctionSymbol {
		if len(o) == 2 {
			var ins []Instruction
			if o[1].Primitive() == VECTOR && CompareTypes(o[0], o[1].Items()) {
				ins = []Instruction{MKInstruction(SWT), MKInstruction(APP)}
			} else if o[1].Primitive() == MAP && CompareTypes(o[0], TupleOf([]OBJType{o[1].Items(), Str})) {
				ins = []Instruction{MKInstruction(EIS), MKInstruction(ATT)}
			} else {
				return nil
			}
			return &FunctionSymbol{"none", false, ins, CloneType(Void), o}
		}
		return nil
	})
	to.AddDynamicSymbol("'", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			if o[0].Primitive() == VECTOR {
				v := VariadicOf(o[0].Items())
				return &FunctionSymbol{"none", false, []Instruction{}, &v, o}
			} else if o[0].Primitive() == VARIADIC {
				v := VecOf(o[0].Items())
				return &FunctionSymbol{"none", false, []Instruction{}, &v, o}
			}
		}
		return nil
	})
	to.AddDynamicSymbol("*", func(o []OBJType) *FunctionSymbol {
		if len(o) == 1 {
			if o[0].Primitive() == ALIAS {
				r := o[0].(*Alias).Holds
				return &FunctionSymbol{"none", false, []Instruction{}, &r, o}
			} else if o[0].Primitive() == STRUCT {
				t := TupleOf(o[0].FixedItems())
				return &FunctionSymbol{"none", false, []Instruction{}, &t, o}
			}
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

type LoopInfo struct {
	father *LoopInfo
	jumps  []struct {
		position int
		skip     bool
	}
}

func (li *LoopInfo) insertJump(position int, skip bool) {
	li.jumps = append(li.jumps, struct {
		position int
		skip     bool
	}{position, skip})
}

func (li *LoopInfo) solveJumps(parser *Parser, next_pos, end_pos int) {
	for _, v := range li.jumps {
		pos := next_pos
		if v.skip {
			pos = end_pos
		}
		parser.editInstruction(v.position, MKInstruction(MVR, pos-v.position))
	}
}

type ifFlags struct {
	altered    bool
	afterif    bool
	lastifhead struct{ idx, offset int }
	allifskips []struct{ idx, offset int }
}

type returnLnFlag struct {
	altered  bool
	isreturn bool
}

func createIfFlags() *ifFlags {
	return &ifFlags{false, false, struct{ idx, offset int }{-1, -1}, make([]struct{ idx, offset int }, 0)}
}

func createReturnLnFlag() *returnLnFlag {
	return &returnLnFlag{false, false}
}

func (bf *ifFlags) Edit(nbf ifFlags) {
	*bf = nbf
	bf.altered = true
}

func (bf *ifFlags) submitIf(head, skip int) {
	bf.afterif = true
	bf.lastifhead = struct{ idx, offset int }{head, skip}
	bf.altered = true
}

func (bf *ifFlags) submitIfSkip(idx int) {
	bf.allifskips = append(bf.allifskips, struct{ idx, offset int }{idx, 0})
}

func (bf *ifFlags) Reset() {
	*bf = *createIfFlags()
}

func (bf *returnLnFlag) Reset() {
	*bf = *createReturnLnFlag()
}

type Scope struct {
	DataModule      Module
	ImportedModules map[string]Module
	Variables       map[string]*Variable
	DefinedTypes    map[string]*OBJType
	Functions       *FunctionCollection
	Operators       *FunctionCollection
	ifFlags         *ifFlags
	loopInfo        *LoopInfo
	returnLnFlag    *returnLnFlag
	ReturnType      *OBJType
	Returned        *bool
	parent          *Scope
	varcount        *uint
	argcount        *uint
}

func (s *Scope) forkLoopInfo() {
	s.loopInfo = &LoopInfo{s.loopInfo, make([]struct {
		position int
		skip     bool
	}, 0)}
}

func (s *Scope) closeLoopInfo() {
	if s.loopInfo == nil {
		panic(errors.New("Trying to close not opened loop"))
	}
	s.loopInfo = s.loopInfo.father
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

func (s *Scope) NewType(name string, t OBJType) error {
	if _, e := s.DefinedTypes[name]; e {
		return errors.New(fmt.Sprintf("Type %s already exists", name))
	}
	s.DefinedTypes[name] = &t
	return nil
}

func (s *Scope) FetchType(name string) (*OBJType, error) {
	if k, e := s.DefinedTypes[name]; e {
		return k, nil
	}
	return nil, errors.New(fmt.Sprintf("Type %s does not exists", name))
}

func NewScope(name string) *Scope {
	rptr := Void
	r := false
	var vc, ac uint = 0, 0
	var scope *Scope
	scope = &Scope{
		ImportedModules: map[string]Module{"core": core},
		Variables:       make(map[string]*Variable),
		DefinedTypes:    make(map[string]*OBJType),
		Functions:       NewFunctionCollection(),
		Operators:       NewFunctionCollection(),
		ifFlags:         createIfFlags(),
		loopInfo:        nil,
		returnLnFlag:    createReturnLnFlag(),
		ReturnType:      &rptr, Returned: &r, parent: nil, varcount: &vc, argcount: &ac}
	scope.DataModule = &FileModule{name, scope}
	return scope
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
	ns := &Scope{
		DataModule:      s.DataModule,
		ImportedModules: make(map[string]Module),
		Variables:       make(map[string]*Variable),
		DefinedTypes:    make(map[string]*OBJType),
		Functions:       s.Functions.Fork(),
		Operators:       s.Operators.Fork(),
		ifFlags:         createIfFlags(),
		loopInfo:        s.loopInfo,
		returnLnFlag:    createReturnLnFlag(),
		ReturnType:      returnt, Returned: returned, parent: s, varcount: vc, argcount: ac}
	for k, v := range s.ImportedModules {
		ns.ImportedModules[k] = v
	}
	for k, v := range s.Variables {
		ns.Variables[k] = v
	}
	for k, v := range s.DefinedTypes {
		ns.DefinedTypes[k] = v
	}
	return ns
}

func (s *Scope) ImportFrom(other *Scope) error {
	for k, v := range other.ImportedModules {
		if c, ex := s.ImportedModules[k]; ex {
			if !c.Is(v) {
				return errors.New(fmt.Sprintf("Module %s refers to different module", k))
			}
		} else {
			s.ImportedModules[k] = v
		}
	}
	for k, v := range other.Variables {
		if _, ex := s.Variables[k]; ex {
			return errors.New(fmt.Sprintf("Variable %s already defined", k))
		}
		s.Variables[k] = v
	}
	for k, v := range other.DefinedTypes {
		if _, ex := s.DefinedTypes[k]; ex {
			return errors.New(fmt.Sprintf("Type %s already defined", k))
		}
		s.DefinedTypes[k] = v
	}
	if e := s.Operators.CopyFrom(other.Operators); e != nil {
		return e
	}
	return s.Functions.CopyFrom(other.Functions)
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
