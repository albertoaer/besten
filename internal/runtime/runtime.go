package runtime

import (
	"errors"
	"fmt"
)

type EmbeddedFunction struct {
	Name     string
	ArgCount int
	Function func(args ...Object) Object
	Returns  bool
}
type Object interface{}
type MapT map[string]Object
type VecT *[]Object

type Instruction struct {
	Code     ICode
	Operands []Object
}

func MKInstruction(code ICode, operands ...Object) Instruction {
	return Instruction{Code: code, Operands: operands}
}

type Fragment []Instruction

type Symbol struct {
	Name   string   //fragment name
	Source Fragment //fragment
}

func (s Symbol) Append(i Instruction) {
	s.Source = append(s.Source, i)
}

type Context struct {
	values map[string]*Object
	parent *Context
}

func NContext() *Context {
	return &Context{make(map[string]*Object, 0), nil}
}

func (source *Context) Open() *Context {
	forked := NContext()
	for k, v := range source.values {
		forked.values[k] = v
	}
	forked.parent = source
	return forked
}

func (source *Context) Close() *Context {
	if source.parent == nil {
		panic("Trying to close root context")
	}
	return source.parent
}

func (source *Context) Get(key string) (Object, error) {
	r, e := source.values[key]
	if !e {
		return nil, errors.New(fmt.Sprintf("%s is not defined in the current context", key))
	}
	return *r, nil
}

func (source *Context) Set(key string, val Object) {
	_, e := source.values[key]
	if !e {
		source.values[key] = &val
	} else {
		*source.values[key] = val
	}
}

type CallStackElement struct {
	pc       int               //pc
	fragment Symbol            //fragment
	context  *Context          //The context that was running on the fragment
	father   *CallStackElement //The element under it
}

type CallStack struct {
	element *CallStackElement
}

func NewCallStack() *CallStack {
	return &CallStack{nil}
}

func (stack *CallStack) Insert(pc int, fragment Symbol, context *Context) {
	stack.element = &CallStackElement{pc, fragment, context, stack.element}
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
