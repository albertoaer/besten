package parser

import (
	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func loadModuleAux(tks []Token, fn func(string) (Symbols, error)) (symbols Symbols, err error) {
	tks = discardOne(tks)
	t, tks, e := expectT(tks, IdToken)
	if e != nil {
		err = e
		return
	}
	if e := unexpect(tks); e != nil {
		err = e
		return
	}
	s, e := fn(t.Data)
	symbols = s
	err = e
	return
}

/* func globalDefAux(tks []Token, symbols Symbols, constant bool) (Symbol, error) {
	tks = discardOne(tks)
	name, tks, e := expectT(tks, IdToken)
	if e != nil {
		return nil, e
	}
	tks, e = expect(tks, ASSIGN)
	if e != nil {
		return nil, e
	}
	symbols = s
	err = e
	return
} */

func (p *Parser) parseGlobalBlocks(blocks []Block) (symbols Symbols, err error) {
	for _, block := range blocks {
		if next(block.Tokens, REQUIRE) {
			s, e := loadModuleAux(block.Tokens, p.env.Native)
			if e != nil {
				err = e
				return
			}
			symbols = Merge(symbols, s)
		} else if next(block.Tokens, IMPORT) {
			s, e := loadModuleAux(block.Tokens, p.env.File)
			if e != nil {
				err = e
				return
			}
			symbols = Merge(symbols, s)
		} /*  else if next(block.Tokens, VAL) {
			s, e := globalDefAux(block.Tokens)
			if e != nil {
				err = e
				return
			}
			symbols = Include(symbols, s)
		} else if next(block.Tokens, VAR) {
			s, e := globalDefAux(block.Tokens)
			if e != nil {
				err = e
				return
			}
			symbols = Include(symbols, s)
		} */
	}
	return
}
