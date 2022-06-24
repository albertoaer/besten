package parser

import (
	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func (p *Parser) parseExpression(tks []Token, children []Block) (OBJType, error) {
	ast, err := GenerateTree(p, tks, children)
	if err != nil {
		return Void, err
	}
	stack := make([]Instruction, 0)
	t, err := ast.runIntoStack(&stack)
	p.addInstructions(stack)
	return t, err
}
