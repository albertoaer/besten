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
	Call    []Instruction
	Return  *OBJType
	Args    []OBJType
}

type DynamicFunctionSymbol func([]OBJType) *FunctionSymbol

type NamedTemplateContainer struct {
	fixedargs map[int]FunctionTemplate
	variadic  *FunctionTemplate //Can only be one variadic template of a certain name
}

type NamedFunctionContainer struct {
	dynamic         DynamicFunctionSymbol
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
	collection.SaveSymbolHolder(name)
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
	if v.variadic != nil && len(v.variadic.Args) <= args {
		return v.variadic
	}
	return nil
}

func (collection *FunctionCollection) DropAllTemplatesOf(name string) {
	if _, e := collection.templates[name]; e {
		delete(collection.templates, name)
	}
}

func (collection *FunctionCollection) DropTemplate(name string, args int) {
	if v, e := collection.templates[name]; e {
		if _, e := v.fixedargs[args]; e {
			delete(v.fixedargs, args)
		}
	}
}

func (collection *FunctionCollection) CopyAllTemplatesOf(name string, nname string, other *FunctionCollection) error {
	if v, e := collection.templates[name]; e {
		for _, v := range v.fixedargs {
			if err := other.AddTemplate(nname, v); err != nil {
				return err
			}
		}
		if v.variadic != nil {
			if err := other.AddTemplate(nname, *v.variadic); err != nil {
				return err
			}
		}
	}
	return nil
}

func (collection *FunctionCollection) CopyTemplate(name string, args int, nname string, other *FunctionCollection) error {
	if v, e := collection.templates[name]; e {
		if v, e := v.fixedargs[args]; e {
			if err := other.AddTemplate(nname, v); err != nil {
				return err
			}
		}
		if v.variadic != nil && args >= len(v.variadic.Args) {
			if err := other.AddTemplate(nname, *v.variadic); err != nil {
				return err
			}
		}
	}
	return nil
}

//Creates a function container for a name in case it does not exists
func (collection *FunctionCollection) SaveSymbolHolder(name string) {
	if _, e := collection.functions[name]; !e {
		collection.functions[name] = &NamedFunctionContainer{nil, make(map[int][]*FunctionSymbol), make([]*FunctionSymbol, 0), -1}
	}
}

//Adds variadic and non-variadic functions into de collection associated with a name
func (collection *FunctionCollection) AddSymbol(name string, function *FunctionSymbol) error {
	v, e := collection.functions[name]
	if !e {
		v = &NamedFunctionContainer{nil, make(map[int][]*FunctionSymbol), make([]*FunctionSymbol, 0), -1}
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

func (collection *FunctionCollection) AddDynamicSymbol(name string, dsym DynamicFunctionSymbol) error {
	v, e := collection.functions[name]
	if !e {
		v = &NamedFunctionContainer{dsym, make(map[int][]*FunctionSymbol), make([]*FunctionSymbol, 0), -1}
		collection.functions[name] = v
	}
	if v.dynamic != nil {
		return errors.New("Dynamic symbol already defined")
	}
	v.dynamic = dsym
	return nil
}

func (collection *FunctionCollection) AddSymbols(name string, functions []*FunctionSymbol) error {
	for _, f := range functions {
		if e := collection.AddSymbol(name, f); e != nil {
			return e
		}
	}
	return nil
}

func (collection *FunctionCollection) GenerateDynamicSymbol(name string, args []OBJType) *FunctionSymbol {
	v, e := collection.functions[name]
	if !e {
		return nil
	}
	if v.dynamic == nil {
		return nil
	}
	return v.dynamic(args)
}

//If returns null no function was found
func (collection *FunctionCollection) FindSymbol(name string, args int) []*FunctionSymbol {
	v, e := collection.functions[name]
	if !e {
		return nil
	}
	vec := v.fixedargs[args]
	if args >= v.minvariadicargs {
		vec = append(vec, v.variadic...)
	}
	return vec
}

func (collection *FunctionCollection) DropAllSymbolOf(name string) {
	if _, e := collection.functions[name]; e {
		delete(collection.functions, name)
	}
}

func (collection *FunctionCollection) DropSymbol(name string, args int) {
	if v, e := collection.functions[name]; e {
		if _, e = v.fixedargs[args]; e {
			delete(v.fixedargs, args)
		}
		if args >= v.minvariadicargs {
			v.variadic = make([]*FunctionSymbol, 0)
		}
	}
}

func (collection *FunctionCollection) CopyAllSymbolsOf(name string, nname string, other *FunctionCollection) error {
	if v, e := collection.functions[name]; e {
		for _, v := range v.fixedargs {
			if err := other.AddSymbols(nname, v); err != nil {
				return err
			}
		}
		if len(v.variadic) > 0 {
			if err := other.AddSymbols(nname, v.variadic); err != nil {
				return err
			}
		}
		if v.dynamic != nil {
			if err := other.AddDynamicSymbol(nname, v.dynamic); err != nil {
				return err
			}
		}
	}
	return nil
}

func (collection *FunctionCollection) CopySymbol(name string, args int, nname string, other *FunctionCollection) error {
	if v, e := collection.functions[name]; e {
		if v, e := v.fixedargs[args]; e {
			if err := other.AddSymbols(nname, v); err != nil {
				return err
			}
		}
		if len(v.variadic) > 0 && args >= v.minvariadicargs {
			if err := other.AddSymbols(nname, v.variadic); err != nil {
				return err
			}
		}
	}
	return nil
}
