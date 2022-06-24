package parser

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"

	"github.com/Besten/internal/runtime"
)

func (p *Parser) solveFunctionCall(name string, operator bool, callers []OBJType) (ins runtime.Instruction, ret OBJType, err error) {
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
	ins = s.Call
	ret = *s.Return
	return
}

func (p *Parser) findFunction(name string, operator bool, callers []OBJType) (sym *FunctionSymbol, exists bool) {
	exists = false
	var vec []*FunctionSymbol
	if operator {
		vec = p.currentScope().Operators.FindSymbol(name, len(callers))
	} else {
		vec = p.currentScope().Functions.FindSymbol(name, len(callers))
	}
	if vec == nil {
		return //Function not found
	}
outer:
	for i := range vec {
		for e := 0; e < len(callers); e++ {
			if !CompareTypes(callers[e], vec[i].Args[e]) {
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

var lambdaCount uint = 0

func (p *Parser) solveLambdaTemplate(template FunctionTemplate) string {
	name := "lambda" + strconv.Itoa(int(lambdaCount))
	lambdaCount++
	if e := p.currentScope().Functions.AddTemplate(name, template); e != nil {
		panic(e) //Unexpected behaviour, overwriting lambda
	}
	return name
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
		err = errors.New(fmt.Sprintf("There is no %s symbol %s/%d for %s", symboltype, name, len(callers), ArrRepr(callers)))
		return
	}
	compilename := generateFnUUID(name, p.modulename, len(callers))

	p.openFragment(compilename)

	/*
		Create function reference before function in order to avoid posible infinite dependency loops
	*/
	sym = &FunctionSymbol{CName: compilename, Call: runtime.MKInstruction(runtime.CLL, compilename), Return: p.currentScope().ReturnType, Args: callers}
	if operator {
		p.currentScope().Operators.AddSymbol(name, sym)
	} else {
		p.currentScope().Functions.AddSymbol(name, sym)
	}

	for i := range template.Args {
		//Stack, invert order
		p.addInstruction(runtime.MKInstruction(runtime.SET, template.Args[i]))
		p.currentScope().Variables[template.Args[i]] = &Variable{callers[i], true}
	}
	err = p.parseBlocks(template.Children, Function)
	if err != nil {
		return
	}
	p.addInstruction(runtime.MKInstruction(runtime.RET))
	p.backToFragment()

	return
}

var counters map[struct {
	string
	int
}]*int = make(map[struct {
	string
	int
}]*int)

func generateFnUUID(name, module string, args int) string {
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
	result := fmt.Sprintf("%s/%d@%x%d", name, args, modulesum, *v)
	(*v)++
	return result
}
