package runtime

import "errors"

type FunctionStack struct {
	data  []Object
	index uint
}

func NewFunctionStack(gensize uint) *FunctionStack {
	fs := &FunctionStack{make([]Object, gensize), 0}
	return fs
}

func (fs *FunctionStack) Clear() {
	fs.data = make([]Object, len(fs.data))
	fs.index = 0
}

func (fs *FunctionStack) Push(o Object) {
	if fs.index >= uint(len(fs.data)) {
		panic("Unexpected situation, stack overflow")
	}
	fs.data[fs.index] = o
	fs.index += 1
}

func (fs *FunctionStack) Pop() Object {
	if fs.index == 0 {
		panic(errors.New("Void stack"))
	}
	fs.index--
	return fs.data[fs.index]
}

func (fs *FunctionStack) Clone() *FunctionStack {
	fn := NewFunctionStack(uint(len(fs.data)))
	fn.index = fs.index
	for i := range fs.data {
		fn.data[i] = fs.data[i]
	}
	return fn
}

type ItemManager struct {
	locals      []Object
	environment []Object
}

func NewItemManager(localsnum int) *ItemManager {
	return &ItemManager{make([]Object, localsnum), make([]Object, 0)}
}

func ItemManagerForCall(fs *FunctionStack, variables int, varargs bool, localsnum int) (*ItemManager, error) {
	extra := 0
	objs := make([]Object, 0)
	var varargsvec VecT
	for i := 0; i < variables+extra; i++ {
		o := fs.Pop()
		if i == variables-1 && varargs {
			i, e := o.(int)
			if !e {
				return nil, errors.New("Expecting number of arguments")
			}
			extra = i
			varargsvec = MakeVec()
			objs = append(objs, varargsvec)
		} else if i > variables-1 && varargs {
			*varargsvec = append(*varargsvec, o)
		} else {
			objs = append(objs, o)
		}
	}
	return &ItemManager{make([]Object, localsnum), objs}, nil
}

func (proc *ItemManager) GetLocal(idx int) Object {
	return proc.locals[idx]
}

func (proc *ItemManager) SetLocal(idx int, value Object) {
	proc.locals[idx] = value
}

func (proc *ItemManager) GetEnvironment(idx int) Object {
	return proc.environment[idx]
}

func (proc *ItemManager) SetEnvironment(idx int, value Object) {
	proc.environment[idx] = value
}
