package runtime

type EmbeddedFunction struct {
	Name     string
	ArgCount int
	Function func(args []Object) Object
	Returns  bool
}
type Object interface{}
type MapT map[string]Object
type VecT *[]Object

func MakeVec(items ...Object) VecT {
	var vec []Object = items
	return VecT(&vec)
}

type Instruction struct {
	Code     ICode
	operands [2]Object
	sz       int
}

func MKInstruction(code ICode, operands ...Object) Instruction {
	if len(operands) > 2 {
		panic("Instructions can only have up to 2 arguments")
	}
	var opr [2]Object
	copy(opr[:], operands)
	return Instruction{code, opr, len(operands)}
}

func (i Instruction) Fragment() []Instruction {
	return []Instruction{i}
}

type Fragment []Instruction

type Symbol struct {
	Name      string   //fragment name
	Source    Fragment //fragment
	BuiltInfo struct {
		Args    int
		Varargs bool
	}
}

func (s *Symbol) Append(i Instruction) {
	s.Source = append(s.Source, i)
}

type CallStackElement struct {
	pc     int         //pc
	symbol *Symbol     //fragment
	env    Environment //environment, args...
	locals Locals      //local variables
}

type CallStack struct {
	elements []CallStackElement
	idx      int
}

func NewCallStack(size int) *CallStack {
	return &CallStack{make([]CallStackElement, size), 0}
}

func (stack *CallStack) Insert(pc int, symbol *Symbol, env Environment, locals Locals) {
	if stack.idx >= len(stack.elements) {
		panic("Unexpected situation, call stack overflow")
	}
	stack.elements[stack.idx] = CallStackElement{pc, symbol, env, locals}
	stack.idx++
}

func (stack *CallStack) Pop() {
	if stack.idx == 0 {
		panic("Poping from empty call stack")
	}
	stack.idx--
}

func (stack *CallStack) Current() *CallStackElement {
	if stack.idx == 0 {
		return nil
	}
	return &stack.elements[stack.idx-1]
}
