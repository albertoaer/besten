package runtime

type EmbeddedFunction struct {
	Name     string
	ArgCount int
	Function func(args ...Object) Object
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
	Operands []Object
}

func MKInstruction(code ICode, operands ...Object) Instruction {
	return Instruction{Code: code, Operands: operands}
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
	pc       int               //pc
	fragment Symbol            //fragment
	father   *CallStackElement //The element under it
	items    *ItemManager
}

type CallStack struct {
	element *CallStackElement
}

func NewCallStack() *CallStack {
	return &CallStack{nil}
}

func (stack *CallStack) Insert(pc int, fragment Symbol, items *ItemManager) {
	stack.element = &CallStackElement{pc, fragment, stack.element, items}
}

func (stack *CallStack) Pop() {
	if stack.element == nil {
		panic("Poping from empty call stack")
	}
	stack.element = stack.element.father
}

func (stack *CallStack) Current() *CallStackElement {
	return stack.element
}
