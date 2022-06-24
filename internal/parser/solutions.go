package parser

import (
	"crypto/md5"
	"errors"
	"fmt"

	"github.com/Besten/internal/runtime"
)

func (p *Parser) processFunctionCall(name string, operator bool, callers []OBJType,
	insbuffers [][]runtime.Instruction) (ret OBJType, err error) {
	into := make([]runtime.Instruction, 0)
	ret, err = p.solveFunctionCall(name, operator, callers, insbuffers, &into)
	if err == nil {
		p.addInstructions(into)
	}
	return
}

func (p *Parser) solveFunctionCall(name string, operator bool, callers []OBJType,
	insbuffers [][]runtime.Instruction, into *[]runtime.Instruction) (ret OBJType, err error) {
	var sym *FunctionSymbol
	sym, err = p.getSymbolForCall(name, operator, callers)
	if err != nil {
		return
	}
	ret = *sym.Return
	for i := range insbuffers {
		*into = append(*into, insbuffers[len(insbuffers)-i-1]...)
		if len(insbuffers)-i == len(sym.Args) && sym.Varargs && callers[i].Primitive() != VARIADIC {
			*into = append(*into, runtime.MKInstruction(runtime.CSE, len(insbuffers)-len(sym.Args)+1))
		}
	}
	*into = append(*into, sym.Call...)
	return
}

func (p *Parser) getSymbolForCall(name string, operator bool, callers []OBJType) (sym *FunctionSymbol, err error) {
	if operator && (len(callers) > 2 || len(callers) < 1) {
		err = errors.New(fmt.Sprintf("Operators cannot have %d arguments", len(callers)))
		return
	}
	s, ex := p.findFunction(name, operator, callers)
	if !ex {
		s, err = p.generateFunctionFromTemplate(name, operator, callers)
		if err != nil {
			return
		}
	}
	sym = s
	return
}

func (p *Parser) findFunction(name string, operator bool, callers []OBJType) (sym *FunctionSymbol, exists bool) {
	exists = false
	var vec []*FunctionSymbol
	if operator {
		if s := p.currentScope().Operators.GenerateDynamicSymbol(name, callers); s != nil {
			return s, true
		}
		vec = p.currentScope().Operators.FindSymbol(name, len(callers))
	} else {
		if s := p.currentScope().Functions.GenerateDynamicSymbol(name, callers); s != nil {
			return s, true
		}
		vec = p.currentScope().Functions.FindSymbol(name, len(callers))
	}
	if vec == nil {
		return //Function not found
	}
outer:
	for i := range vec {
		for e := 0; e < len(callers); e++ {
			if vec[i].Varargs && e >= len(vec[i].Args)-1 {
				last := vec[i].Args[len(vec[i].Args)-1]
				if last.Primitive() != VECTOR {
					continue outer
				}
				comparator := callers[e]
				if len(callers) == len(vec[i].Args) && callers[e].Primitive() == VARIADIC {
					comparator = callers[e].Items()
				}
				if !CompareTypes(comparator, last.Items()) {
					continue outer
				}
			} else if !CompareTypes(callers[e], vec[i].Args[e]) {
				continue outer
			}
		}
		sym = vec[i]
		exists = true
		return
	}
	return //Function not generated with desired arguments
}

//Remember: This function checks is there is already a template for that name
func (p *Parser) generateFunctionTemplate(name string, operator bool, template FunctionTemplate) error {
	if operator {
		return p.currentScope().Operators.AddTemplate(name, template)
	} else {
		return p.currentScope().Functions.AddTemplate(name, template)
	}
}

//Remember: This function does not check if there is another function with that types
func (p *Parser) generateFunctionFromTemplate(name string, operator bool, callers []OBJType) (sym *FunctionSymbol, err error) {
	var template *FunctionTemplate
	if operator {
		template = p.currentScope().Operators.FindTemplate(name, len(callers))
	} else {
		template = p.currentScope().Functions.FindTemplate(name, len(callers))
	}
	if template == nil {
		symboltype := "function"
		if operator {
			symboltype = "operator"
		}
		err = errors.New(fmt.Sprintf("There is no %s template %s/%d valid for %s", symboltype, name, len(callers), FnCArrRepr(callers)))
		return
	}
	return p.generateFunctionFromRawTemplate(name, operator, callers, template)
}

func (p *Parser) generateFunctionFromRawTemplate(name string, operator bool, callers []OBJType, template *FunctionTemplate) (sym *FunctionSymbol, err error) {
	compilename := generateFnUUID(name, p.modulename, len(template.Args), template.Varargs, false)

	p.openFragmentFor(compilename, len(template.Args))

	args := make([]OBJType, 0)

	for i := range template.Args {
		if template.Varargs && i == len(template.Args)-1 {
			var v OBJType
			if len(callers) > len(template.Args) || callers[i].Primitive() != VARIADIC {
				for j := i + 1; j < len(callers); j++ {
					if !CompareTypes(callers[j-1], callers[j]) {
						err = errors.New("Variadic elements must all be the same type")
						return
					}
				}
				v = VecOf(callers[i])
			} else {
				v = VecOf(callers[i].Items())
			}
			p.currentScope().CreateVariable(template.Args[i], v, true, true)
			args = append(args, v)
		} else {
			p.currentScope().CreateVariable(template.Args[i], callers[i], true, true)
			if callers[i].Primitive() == FUNCTION {
				if err = p.functionFromVariable(template.Args[i]); err != nil {
					return
				}
			}
			args = append(args, callers[i])
		}
	}
	/*
		Create function reference before function in order to avoid posible infinite dependency loops
	*/
	sym = &FunctionSymbol{CName: compilename, Call: runtime.MKInstruction(runtime.CLL, compilename).Fragment(), Return: p.currentScope().ReturnType, Args: args, Varargs: template.Varargs}
	if operator {
		p.currentScope().Operators.AddSymbol(name, sym)
	} else {
		p.currentScope().Functions.AddSymbol(name, sym)
	}

	err = p.parseBlocks(template.Children, Function)
	if err != nil {
		return
	}
	if !p.currentScopeOrigin().returnLnFlag.isreturn {
		if (*p.currentScope().ReturnType).Primitive() != VOID {
			err = errors.New("Expecting return of type: " + Repr(*p.currentScope().ReturnType))
			return
		}
		p.addInstruction(runtime.MKInstruction(runtime.RET))
	}
	if err = p.currentScope().CheckClose(); err != nil {
		return
	}
	p.backToFragment()
	return
}

func (p *Parser) functionFromVariable(name string) error {
	ins, fnt, e := p.currentScope().GetVariableIns(name)
	if e != nil {
		return e
	}
	if fnt.Primitive() != FUNCTION {
		return errors.New(fmt.Sprintf("Variable %s is not a function", name))
	}
	fn := fnt.(*FunctionType)
	if e = p.currentScope().Functions.AddSymbol(name, &FunctionSymbol{
		name, false, []runtime.Instruction{ins, runtime.MKInstruction(runtime.CLL)},
		CloneType(fn.ret), fn.args,
	}); e != nil {
		return e
	}
	return nil
}

func (p *Parser) getFunctionTypeFrom(name string, fn *FunctionSymbol) (OBJType, string) {
	compilename := fn.CName
	if len(fn.Call) != 1 || fn.Call[0].Code != runtime.CLL {
		compilename = generateFnUUID(name, p.modulename, len(fn.Args), fn.Varargs, true)
		p.openFragmentFor(compilename, 0)
		p.addInstructions(fn.Call)
		p.addInstruction(runtime.MKInstruction(runtime.RET))
		p.backToFragment()
	}
	return FunctionTypeOf(fn.Args, *fn.Return), compilename
}

var counters map[struct {
	string
	int
}]*int = make(map[struct {
	string
	int
}]*int)

func generateFnUUID(name, module string, args int, varargs, shared bool) string {
	v, e := counters[struct {
		string
		int
	}{name, args}]
	if !e {
		n := 0
		v = &n
		counters[struct {
			string
			int
		}{name, args}] = v
	}
	modulesum := md5.Sum([]byte(module))
	mode := "a"
	if varargs {
		mode = "v"
	}
	if shared {
		mode += "s"
	}
	result := fmt.Sprintf("%s/%d$%s@%x%d", name, args, mode, modulesum, *v)
	(*v)++
	return result
}
