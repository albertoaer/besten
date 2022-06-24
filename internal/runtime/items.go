package runtime

import "errors"

type FunctionStack struct {
	data  []Object
	index int
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
	if fs.index >= len(fs.data) {
		//fs.data = append(fs.data, nil)
		panic("Unexpected situation, function stack overflow")
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

type Environment struct {
	args [8]Object
}

func EnvironmentForCall(fs *FunctionStack, variables int, varargs bool) (Environment, error) {
	extra := 0
	items := Environment{}
	var varargsvec VecT
	for i := 0; i < variables+extra; i++ {
		o := fs.Pop()
		if i == variables-1 && varargs {
			i, e := o.(int)
			if !e {
				return items, errors.New("Expecting number of arguments")
			}
			extra = i
			varargsvec = MakeVec()
			items.args[i] = varargsvec
		} else if i > variables-1 && varargs {
			*varargsvec = append(*varargsvec, o)
		} else {
			items.args[i] = o
		}
	}
	return items, nil
}

func (proc *Environment) GetEnvironment(idx int) Object {
	return proc.args[idx]
}

func (proc *Environment) SetEnvironment(idx int, value Object) {
	proc.args[idx] = value
}

type Locals struct {
	locals [8]Object
}

func (proc *Locals) GetLocal(idx int) Object {
	return proc.locals[idx]
}

func (proc *Locals) SetLocal(idx int, value Object) {
	proc.locals[idx] = value
}
