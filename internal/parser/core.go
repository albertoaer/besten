package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func parseArguments(tks []Token) (args []string, varargs bool, err error) {
	tks = discardOne(tks)
	preargs, r := readUntilToken(tks, DO)
	argtk, err := splitByToken(preargs, func(tk Token) bool { return tk == COMA }, make([]struct {
		open  Token
		close Token
	}, 0), false, false)
	if err != nil {
		return
	}
	tail := r
	if ttail, e := expect(tail, DO); len(ttail) != 0 || e != nil {
		//Arguments no longer have tail, must presume only a 'do' is after the arguments
		err = errors.New("Expecting only a do keyword after args declaration")
		return
	}
	for i, v := range argtk {
		nm, _, e := expectT([]Token{v[len(v)-1]}, IdToken)
		if e != nil {
			err = e
			return
		}
		args = append(args, nm.Data)
		if len(v) > 1 {
			if i == len(argtk)-1 {
				_, e := expect(v, QUOTE)
				if e != nil {
					err = e
					return
				}
				varargs = true
			} else {
				err = errors.New(fmt.Sprintf("Unexpected token: %s", v[0].Data))
			}
		}
	}
	return
}

func (p *Parser) loadModuleAux(tks []Token, fn func(string) (map[string]Symbol, *Scope, error)) error {
	tks = discardOne(tks)
	t, tks, e := expectT(tks, IdToken)
	if e != nil {
		return e
	}
	if e := unexpect(tks); e != nil {
		return e
	}
	f, s, e := fn(t.Data)
	p.addSymbols(f)
	p.currentScope().Merge(s)
	return nil
}

func (p *Parser) parseFunction(block Block, operator bool) error {
	tks := discardOne(block.Tokens)
	var name Token
	var e error
	if operator {
		name, tks, e = expectT(tks, OperatorToken)
	} else {
		name, tks, e = expectT(tks, IdToken)
	}
	if e != nil {
		return e
	}
	var args []string
	varargs := false
	if next(tks, DOUBLES) {
		args_, varargs_, e := parseArguments(tks)
		if e != nil {
			return e
		}
		args = args_
		varargs = varargs_
	}
	if operator && (len(args) > 2 || len(args) < 1) {
		return errors.New("Operator must be unary or binary")
	}
	if operator && varargs {
		return errors.New("Operator can not have varargs")
	}
	template := FunctionTemplate{Args: args, Varargs: varargs, Children: block.Children}
	p.generateFunctionTemplate(name.Data, operator, template)
	return nil
}

func (p *Parser) parseDefinition(block Block, constant bool) error {
	tks := discardOne(block.Tokens)
	tks, e := expect(tks, ASSIGN)
	if e != nil {
		return e
	}
	return nil
}

func (p *Parser) parseByKeyword(name string, block Block, scp ScopeCtx) error {
	if scp == Global {
		switch name {
		case "require":
			return p.loadModuleAux(block.Tokens, p.env.Native)
		case "import":
			return p.loadModuleAux(block.Tokens, p.env.File)
		}
	}
	switch name {
	case "fn", "op":
		return p.parseFunction(block, name == "op")
	case "var", "val":
		return p.parseDefinition(block, name == "val")
	}
	return errors.New(fmt.Sprintf("Unexpected token found: %s", name))
}

func (p *Parser) parseAssignement(tks []Token) error {
	return nil
}

func (p *Parser) parseById(block Block, scp ScopeCtx) error {
	tks, e := splitByToken(block.Tokens, func(tk Token) bool { return tk == ASSIGN }, make([]struct {
		open  Token
		close Token
	}, 0), false, false)
	if e != nil {
		return e
	}
	if len(tks) > 2 {
		return errors.New("Multiassignement is not implemented")
	}
	t, e := p.parseExpression(tks[len(tks)-1], block.Children)
	if e != nil {
		return e
	}
	if len(tks) == 2 {
		if t == Void {
			return errors.New("No value returned")
		}
		return p.parseAssignement(tks[0])
	}
	return nil
}

func (p *Parser) parseBlock(block Block, scp ScopeCtx) error {
	if len(block.Tokens) == 0 {
		panic("Void block")
	}
	if nextT(block.Tokens, KeywordToken) {
		return p.parseByKeyword(block.Tokens[0].Data, block, scp)
	}
	if scp == Function {
		if nextT(block.Tokens, IdToken) {
			//Still using id for assignement but maybe will be treat as expression in the future
			return p.parseById(block, scp)
		} else {
			//No keyword, always a expression
			_, e := p.parseExpression(block.Tokens, block.Children)
			return e
		}
	}
	return errors.New(fmt.Sprintf("Unexpected token %s", block.Tokens[0].Data))
}

func (p *Parser) parseBlocks(blocks []Block, scp ScopeCtx) error {
	for _, block := range blocks {
		if e := p.parseBlock(block, scp); e != nil {
			return e
		}
	}
	return nil
}
