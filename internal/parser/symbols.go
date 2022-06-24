package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/runtime"
)

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
	hasRescue       bool
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

func (s *Scope) updateReturn(ret OBJType) error {
	//Is valid returning void, in order to achive infinite recursion
	if (!(*s.Returned) && (*s.ReturnType).Primitive() == VOID) || CompareTypes(*s.ReturnType, ret) {
		*s.ReturnType = ret
		*s.Returned = true
	} else {
		return fmt.Errorf("Expecting return of type: %s", (*s.ReturnType).TypeName())
	}
	return nil
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
			return MKInstruction(NOP), nil, fmt.Errorf("A value has not been set for: %s", name)
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
	return MKInstruction(NOP), nil, fmt.Errorf("Undefined variable: %s", name)
}

func (s *Scope) SetVariableIns(name string, t OBJType) (Instruction, error) {
	if v, e := s.Variables[name]; e {
		if v.Asigned && !v.Mutable {
			return MKInstruction(NOP), fmt.Errorf("Trying to reassign a constant: %s", name)
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
	return MKInstruction(NOP), fmt.Errorf("Undefined variable: %s", name)
}

func (s *Scope) NewType(name string, t OBJType) error {
	if _, e := s.DefinedTypes[name]; e {
		return fmt.Errorf("Type %s already exists", name)
	}
	s.DefinedTypes[name] = &t
	return nil
}

func (s *Scope) FetchType(name string) (*OBJType, error) {
	if k, e := s.DefinedTypes[name]; e {
		return k, nil
	}
	return nil, fmt.Errorf("Type %s does not exists", name)
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
		ReturnType:      &rptr, Returned: &r, parent: nil,
		varcount: &vc, argcount: &ac, hasRescue: false}
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
		ReturnType:      returnt, Returned: returned, parent: s,
		varcount: vc, argcount: ac, hasRescue: s.hasRescue}
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
				return fmt.Errorf("Module %s refers to different module", k)
			}
		} else {
			s.ImportedModules[k] = v
		}
	}
	for k, v := range other.Variables {
		if _, ex := s.Variables[k]; ex {
			return fmt.Errorf("Variable %s already defined", k)
		}
		s.Variables[k] = v
	}
	for k, v := range other.DefinedTypes {
		if tp, ex := s.DefinedTypes[k]; ex && !(*tp).Module().Is((*tp).Module()) {
			return fmt.Errorf("Type %s already defined", k)
		}
		s.DefinedTypes[k] = v
	}
	if e := s.Operators.CopyFrom(other.Operators); e != nil {
		return e
	}
	return s.Functions.CopyFrom(other.Functions)
}

func (s *Scope) OpenForeignScope(other *Scope) *Scope {
	n := *other
	n.parent = s
	return &n
}

func (s *Scope) CheckClose() error {
	for name, v := range s.Variables {
		if v.Dependency == s && !v.Used && !v.Arg {
			vart := "Variable"
			if !v.Mutable {
				vart = "Constant"
			}
			return fmt.Errorf("%s not used: %s", vart, name)
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
