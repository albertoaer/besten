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

func (fs *FunctionStack) a(x Instruction) Object {
	if x.sz < 1 {
		return fs.Pop()
	}
	return x.operands[0]
}

func (fs *FunctionStack) b(x Instruction) Object {
	if x.sz < 2 {
		return fs.Pop()
	}
	return x.operands[1]
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

func (env *Environment) ForCall(fs *FunctionStack, variables int, varargs bool) error {
	extra := 0
	var varargsvec VecT
	for i := 0; i < variables+extra; i++ {
		o := fs.Pop()
		if i == variables-1 && varargs {
			i, e := o.(int)
			if !e {
				return errors.New("Expecting number of arguments")
			}
			extra = i
			varargsvec = MakeVec()
			env.args[i] = varargsvec
		} else if i > variables-1 && varargs {
			*varargsvec = append(*varargsvec, o)
		} else {
			env.args[i] = o
		}
	}
	return nil
}

func (env *Environment) GetEnvironment(idx int) Object {
	return env.args[idx]
}

func (env *Environment) SetEnvironment(idx int, value Object) {
	env.args[idx] = value
}

type Locals struct {
	locals [20]Object
}

func (lcs *Locals) Clear() {
	for x := range lcs.locals {
		lcs.locals[x] = nil
	}
}

func (lcs *Locals) GetLocal(idx int) Object {
	return lcs.locals[idx]
}

func (lcs *Locals) SetLocal(idx int, value Object) {
	lcs.locals[idx] = value
}
