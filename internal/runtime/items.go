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

func (fs *FunctionStack) PopN(n int) []Object {
	if fs.index < n {
		panic(errors.New("Not enough elements in the stack"))
	}
	o := make([]Object, n)
	for i := 0; i < n; i++ {
		o[i] = fs.data[fs.index-i-1]
	}
	fs.index -= n
	return o
}

func (fs *FunctionStack) Offset(r int) {
	fs.index += r
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
	o := x.operands[0]
	if o == nil {
		return fs.Pop()
	}
	return o
}

func (fs *FunctionStack) b(x Instruction) Object {
	o := x.operands[1]
	if o == nil {
		return fs.Pop()
	}
	return o
}

func (fs *FunctionStack) c(x Instruction) Object {
	o := x.operands[2]
	if o == nil {
		return fs.Pop()
	}
	return o
}

func (fs *FunctionStack) d(x Instruction) Object {
	o := x.operands[3]
	if o == nil {
		return fs.Pop()
	}
	return o
}

type Environment struct {
	args [8]Object
}

func (env *Environment) ForCall(fs *FunctionStack, variables int) {
	for i := 0; i < variables; i++ {
		o := fs.Pop()
		env.args[i] = o
	}
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
