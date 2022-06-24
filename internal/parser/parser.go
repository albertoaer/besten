package parser

import (
	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type Parser struct {
	env      ImportEnv
	virtuals Symbols
}

func NewParser(env ImportEnv, virtuals Symbols) *Parser {
	return &Parser{env, virtuals}
}

func (p *Parser) GetSymbols(blocks []Block) (Symbols, error) {
	s, e := p.parseGlobalBlocks(blocks)
	if e != nil {
		return nil, e
	}
	return s, nil
}
