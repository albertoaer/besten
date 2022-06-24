package parser

import (
	"errors"
	"fmt"

	"github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type FunctionTemplate struct {
	Args     []string
	Varargs  bool
	Children []lexer.Block
}

type FunctionSymbol struct {
	CName   string
	Varargs bool
	Call    Instruction
	Return  *OBJType
	Args    []OBJType
}

type NamedTemplateContainer struct {
	fixedargs map[int]FunctionTemplate
	variadic  *FunctionTemplate //Can only be one variadic template of a certain name
}

type NamedFunctionContainer struct {
	fixedargs       map[int][]*FunctionSymbol
	variadic        []*FunctionSymbol
	minvariadicargs int
}

type FunctionCollection struct {
	templates map[string]*NamedTemplateContainer
	functions map[string]*NamedFunctionContainer
}

func NewFunctionCollection() *FunctionCollection {
	return &FunctionCollection{make(map[string]*NamedTemplateContainer), make(map[string]*NamedFunctionContainer)}
}

func (collection *FunctionCollection) Fork() *FunctionCollection {
	fc := NewFunctionCollection()
	for k, v := range collection.templates {
		fc.templates[k] = v
	}
	for k, v := range collection.functions {
		fc.functions[k] = v
	}
	return fc
}

func (collection *FunctionCollection) CopyFrom(other *FunctionCollection) error {
	for k, v := range other.templates {
		if e := collection.AddTemplate(k, *v.variadic); e != nil { //Copy variadic template
			return e
		}
		for _, t := range v.fixedargs { //Copy fixed length template
			if e := collection.AddTemplate(k, t); e != nil {
				return e
			}
		}
	}
	for k, v := range collection.functions {
		for _, variadic := range v.variadic { //Copy one by one the variadic
			if e := collection.AddSymbol(k, variadic); e != nil {
				return e
			}
		}
		for _, s := range v.fixedargs { //Copy fixed args symbols
			if e := collection.AddSymbols(k, s); e != nil {
				return e
			}
		}
	}
	return nil
}

func (collection *FunctionCollection) AddTemplate(name string, template FunctionTemplate) error {
	v, e := collection.templates[name]
	if !e {
		v = &NamedTemplateContainer{make(map[int]FunctionTemplate), nil}
		collection.templates[name] = v
	}
	if template.Varargs {
		if v.variadic != nil {
			return errors.New(fmt.Sprintf("Symbol %s :: Already a variadic template defined", name))
		}
		tmp := template
		v.variadic = &tmp
	} else {
		if _, e := v.fixedargs[len(template.Args)]; e {
			return errors.New(fmt.Sprintf("Symbol %s :: Already defined a template with %d arguments",
				name, len(template.Args)))
		} else {
			v.fixedargs[len(template.Args)] = template
		}
	}
	return nil
}

//If returns null no template was found
func (collection *FunctionCollection) FindTemplate(name string, args int) *FunctionTemplate {
	v, e := collection.templates[name]
	if !e {
		return nil
	}
	if f, e := v.fixedargs[args]; e {
		return &f
	}
	if v.variadic != nil && len(v.variadic.Args)-1 <= args {
		return v.variadic
	}
	return nil
}

//Adds variadic and non-variadic functions into de collection associated with a name
func (collection *FunctionCollection) AddSymbol(name string, function *FunctionSymbol) error {
	v, e := collection.functions[name]
	if !e {
		v = &NamedFunctionContainer{make(map[int][]*FunctionSymbol), make([]*FunctionSymbol, 0), -1}
		collection.functions[name] = v
	}
	if function.Varargs {
		if v.minvariadicargs >= 0 && v.minvariadicargs != len(function.Args) {
			return errors.New("Symbol %s :: Variadic function with diferent number of arguments")
		}
		v.minvariadicargs = len(function.Args)
		v.variadic = append(v.variadic, function)
	} else {
		fns := v.fixedargs[len(function.Args)]
		fns = append(fns, function)
		v.fixedargs[len(function.Args)] = fns
	}
	return nil
}

/*
Adds an array of non-variadic functions into de collection associated with a name
NOT VALID FOR NON VARIADIC
length will be taken from the first symbol
*/
func (collection *FunctionCollection) AddSymbols(name string, functions []*FunctionSymbol) error {
	if functions == nil || len(functions) == 0 {
		return errors.New("Symbol %s :: Symbol array with length 0")
	}
	v, e := collection.functions[name]
	if !e {
		v = &NamedFunctionContainer{make(map[int][]*FunctionSymbol), make([]*FunctionSymbol, 0), -1}
		collection.functions[name] = v
	}
	fns := v.fixedargs[len(functions[0].Args)]
	fns = append(fns, functions...)
	v.fixedargs[len(functions[0].Args)] = fns
	return nil
}

//If returns null no function was found
func (collection *FunctionCollection) FindSymbol(name string, args int) []*FunctionSymbol {
	v, e := collection.functions[name]
	if !e {
		return nil
	}
	vec := v.fixedargs[args]
	if args >= v.minvariadicargs-1 { //Variadic argument might be omited and will be an array of length 0
		vec = append(vec, v.variadic...)
	}
	return vec
}
