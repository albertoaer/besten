package parser

import (
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
