package parser

import (
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
	ret = s.Return
	return
}

func (p *Parser) findFunction(name string, operator bool, callers []OBJType) (sym FunctionSymbol, exists bool) {
	exists = false
	var vec []FunctionSymbol
	var ex bool
	if operator {
		vec, ex = p.currentScope().OpSymbols[HeaderAlias{name, len(callers)}]
	} else {
		vec, ex = p.currentScope().FunctionSymbols[HeaderAlias{name, len(callers)}]
	}
	if !ex {
		return //Function not found
	}
outer:
	for i := range vec {
		for e := 0; e < len(callers); e++ {
			if callers[e] != vec[i].Args[e] {
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
	var ex bool
	if operator {
		_, ex = p.currentScope().OpTemplates[HeaderAlias{name, len(template.Args)}]
	} else {
		_, ex = p.currentScope().FunctionTemplates[HeaderAlias{name, len(template.Args)}]
	}
	if ex {
		return errors.New(fmt.Sprintf("%s template already exists", name))
	}
	if operator {
		p.currentScope().OpTemplates[HeaderAlias{name, len(template.Args)}] = template
		p.currentScope().OpSymbols[HeaderAlias{name, len(template.Args)}] = make([]FunctionSymbol, 0)
	} else {
		p.currentScope().FunctionTemplates[HeaderAlias{name, len(template.Args)}] = template
		p.currentScope().FunctionSymbols[HeaderAlias{name, len(template.Args)}] = make([]FunctionSymbol, 0)
	}
	return nil
}

var lambdaCount uint = 0

func (p *Parser) solveLambdaTemplate(template FunctionTemplate) string {
	name := "lambda" + strconv.Itoa(int(lambdaCount))
	lambdaCount++
	p.currentScope().FunctionTemplates[HeaderAlias{name, len(template.Args)}] = template
	p.currentScope().FunctionSymbols[HeaderAlias{name, len(template.Args)}] = make([]FunctionSymbol, 0)
	return name
}

//Remember: This function does not check if there is another function with that types
func (p *Parser) generateFunctionFromTemplate(name string, operator bool, callers []OBJType) (sym FunctionSymbol, err error) {
	var template FunctionTemplate
	var exists bool
	if operator {
		template, exists = p.currentScope().OpTemplates[HeaderAlias{name, len(callers)}]
	} else {
		template, exists = p.currentScope().FunctionTemplates[HeaderAlias{name, len(callers)}]
	}
	if !exists {
		name := "function"
		if operator {
			name = "operator"
		}
		err = errors.New(fmt.Sprintf("There is no %s for the requested arguments", name))
		return
	}
	compilename := generateUUID(name) + p.modulename
	p.open(compilename)
	for i := range template.Args {
		//Stack, invert order
		p.addInstruction(runtime.MKInstruction(runtime.SET, template.Args[len(template.Args)-i-1]))
	}
	err = p.parseBlocks(template.Children, Function)
	if err != nil {
		return
	}
	p.addInstruction(runtime.MKInstruction(runtime.RET))
	p.back()

	sym = FunctionSymbol{CName: compilename, Call: runtime.MKInstruction(runtime.CLL, compilename), Return: p.rootscope.ReturnType, Args: callers}

	return
}
