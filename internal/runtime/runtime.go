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
	operands *[4]Object
	sz       uint8
}

func MKInstruction(code ICode, operands ...Object) Instruction {
	if len(operands) > 4 {
		panic("Instructions can only have up to 4 arguments")
	}
	if code >= LDOP {
		panic("Trying to create instruction with invalid opcode")
	}
	var opr [4]Object
	copy(opr[:], operands)
	return Instruction{code, &opr, uint8(len(operands))}
}

func (i Instruction) Inspect() []Object {
	x := make([]Object, i.sz)
	for idx := uint8(0); idx < i.sz; idx++ {
		x[idx] = i.operands[idx]
	}
	return x
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

func (stack *CallStack) GetAvailableItems() (*Environment, *Locals) {
	return &stack.elements[stack.idx].env, &stack.elements[stack.idx].locals
}

func (stack *CallStack) Insert(pc int, symbol *Symbol) {
	if stack.idx >= len(stack.elements) {
		panic("Unexpected situation, call stack overflow")
	}
	c := &stack.elements[stack.idx]
	c.pc = pc
	c.symbol = symbol
	stack.idx++
}

func (stack *CallStack) InsertCopy(pc int, symbol *Symbol, env Environment, locals Locals) {
	if stack.idx >= len(stack.elements) {
		panic("Unexpected situation, call stack overflow")
	}
	c := &stack.elements[stack.idx]
	c.pc = pc
	c.symbol = symbol
	c.env = env
	c.locals = locals
	stack.idx++
}

func (stack *CallStack) Pop() {
	if stack.idx == 0 {
		panic("Poping from empty call stack")
	}
	stack.idx--
}

func (stack *CallStack) Top() *CallStackElement {
	if stack.idx == 0 {
		return nil
	}
	return &stack.elements[stack.idx-1]
}
