package runtime

import (
	"errors"
	"fmt"
)

type Instruction struct {
	Code     ICode
	Operands []Object
}

type Symbol []Instruction
type Symbols []Symbol

type ImportEnv interface {
	Native(name string) (Symbols, error)
	File(name string) (Symbols, error)
}

type Context struct {
	values map[string]*Object
}

func NContext() *Context {
	return &Context{make(map[string]*Object, 0)}
}

func (source *Context) Fork() *Context {
	forked := NContext()
	for k, v := range source.values {
		forked.values[k] = v
	}
	return forked
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

func Merge(a Symbols, b Symbols) Symbols {
	for k, v := range b {
		a[k] = v
	}
	return a
}
