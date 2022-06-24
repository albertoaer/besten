package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func (p *Parser) parseExpression(tks []Token, children []Block, returning bool) (OBJType, error) {
	t, stack, err := p.parseExpressionInto(tks, children, returning)
	if err == nil {
		p.addInstructions(stack)
	}
	return t, err
}

func (p *Parser) parseExpressionInto(tks []Token, children []Block, returning bool) (OBJType, []Instruction, error) {
	ast, err := GenerateTree(p, tks, children, returning)
	if err != nil {
		return Void, nil, err
	}
	stack := make([]Instruction, 0)
	t, err := ast.runIntoStack(&stack)
	return t, stack, err
}

func (p *Parser) parseArguments(tks []Token) (args []string, types []OBJType, usetypes bool, varargs bool, err error) {
	tks = discardOne(tks)
	preargs, r := readUntilToken(tks, DO)
	if err != nil {
		return
	}
	{
		tail, e := expect(r, DO)
		if e != nil {
			err = e
			return
		}
		if e := unexpect(tail); e != nil {
			err = e
			return
		}
	}
	argtk, err := splitByToken(preargs, func(tk Token) bool { return tk == COMA }, genericPairs, false, false, false)
	for i, v := range argtk {
		n := v
		if next(n, QUOTE) {
			if i == len(argtk)-1 {
				if usetypes {
					err = errors.New("Varargs must be template")
				}
				varargs = true
				n = discardOne(n)
			} else {
				err = fmt.Errorf("Unexpected token: %s", v[0].Data)
				return
			}
		}
		nm, n, e := expectT(n, IdToken)
		if e != nil {
			err = e
			return
		}
		for _, x := range args {
			if x == nm.Data {
				err = fmt.Errorf("Argument %s repeated", x)
			}
		}
		args = append(args, nm.Data)
		if i == 0 && len(n) > 0 {
			if varargs {
				err = errors.New("Varargs must be template")
			}
			usetypes = true
			types = make([]OBJType, 0)
		}
		if usetypes {
			if len(n) == 0 {
				err = fmt.Errorf("Expecting type for argument: %s", nm.Data)
				return
			}
			tp, e := solveContextedTypeFromTokens(n, p, false)
			if e != nil {
				err = e
				return
			}
			types = append(types, tp)
		} else {
			if len(n) > 0 {
				err = fmt.Errorf("Unexpected type for template argument: %s", nm.Data)
				return
			}
		}
	}
	return
}
