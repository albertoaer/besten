package parser

import (
	"errors"
	"fmt"
	"strconv"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func parseArguments(tks []Token) (args []string, varargs bool, err error) {
	tks = discardOne(tks)
	preargs, r := readUntilToken(tks, DO)
	argtk, err := splitByToken(preargs, func(tk Token) bool { return tk == COMA }, make([]struct {
		open  Token
		close Token
	}, 0), false, false, false)
	if err != nil {
		return
	}
	tail, e := expect(r, DO)
	if e != nil {
		err = e
		return
	}
	if e := unexpect(tail); e != nil {
		err = e
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
	/*tks = discardOne(tks)
	t, tks, e := expectT(tks, IdToken)
	if e != nil {
		return e
	}
	if e := unexpect(tks); e != nil {
		return e
	}*/
	//FIXME: Fix module inclusion
	/*f, s, e := fn(t.Data)
	p.addSymbols(f)
	p.currentScope().Merge(s)*/
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
	id, tks, e := expectT(tks, IdToken)
	if e != nil {
		return e
	}
	tks, e = expect(tks, ASSIGN)
	if e != nil {
		return e
	}
	ret, e := p.parseExpression(tks, block.Children, false)
	if e != nil {
		return e
	}
	p.currentScope().CreateVariable(id.Data, ret, !constant, false)
	ins, e := p.currentScope().SetVariableIns(id.Data, ret)
	if e != nil {
		return e
	}
	p.addInstruction(ins)
	return nil
}

func (p *Parser) parseReturn(block Block) error {
	tks := discardOne(block.Tokens)
	if len(tks) != 0 {
		ret, e := p.parseExpression(tks, block.Children, true)
		if e != nil {
			return e
		}
		//Is valid returning void, in order to achive infinite recursion
		if (!(*p.currentScope().Returned) && (*p.currentScope().ReturnType).Primitive() == VOID) || CompareTypes(*p.currentScope().ReturnType, ret) {
			*p.currentScope().ReturnType = ret
		} else {
			return errors.New(fmt.Sprintf("Expecting return type: %s", (*p.currentScope().ReturnType).TypeName()))
		}
	} else if (*p.currentScope().ReturnType).Primitive() != VOID {
		return errors.New(fmt.Sprintf("Expecting return type: %s", (*p.currentScope().ReturnType).TypeName()))
	}
	*p.currentScope().Returned = true
	p.addInstruction(MKInstruction(RET))
	return nil
}

func (p *Parser) parseDirectLine(tks []Token) error {
	if len(tks) > 0 {
		tk, tks, err := expectT(tks, IntegerToken)
		if err != nil {
			return err
		}
		icode, err := strconv.Atoi(tk.Data)
		if err != nil {
			return err
		}
		objs := make([]Object, 0)
		for _, t := range tks {
			switch t.Kind {
			case StringToken:
				objs = append(objs, t.Data)
			case IntegerToken:
				i, e := strconv.Atoi(t.Data)
				if e != nil {
					return e
				}
				objs = append(objs, i)
			case DecimalToken:
				f, e := strconv.ParseFloat(t.Data, 64)
				if e != nil {
					return e
				}
				objs = append(objs, f)
			case KeywordToken:
				if t.Data == TRUE.Data {
					objs = append(objs, -1)
					break
				} else if t.Data == FALSE.Data {
					objs = append(objs, 0)
					break
				}
				fallthrough
			default:
				return errors.New("Wrong literal")
			}
		}
		p.addInstruction(MKInstruction(ICode(icode), objs...))
	}
	return nil
}

func (p *Parser) parseDirect(block Block) error {
	tks := discardOne(block.Tokens)
	tks, err := expect(tks, DO)
	if err != nil {
		return err
	}
	if err = unexpect(tks); err != nil {
		return err
	}
	for _, child := range block.Children {
		if len(child.Children) != 0 {
			return errors.New("Direct block cannot have childs")
		}
		if err := p.parseDirectLine(child.Tokens); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseIf(block Block) error {
	tks := discardOne(block.Tokens)
	tks, r := readUntilToken(tks, DO)
	tp, e := p.parseExpression(tks, nil, false)
	if e != nil {
		return e
	}
	if tp.Primitive() != BOOL {
		return errors.New("Expecting boolean expression")
	}
	tail, e := expect(r, DO)
	if e != nil {
		return e
	}
	if e := unexpect(tail); e != nil {
		return e
	}
	editpoint := p.addInstruction(MKInstruction(MVF))
	begin := p.fragmentSize()
	p.openScope()
	e = p.parseBlocks(block.Children, Function)
	if e != nil {
		return e
	}
	p.closeScope()
	end := p.fragmentSize()
	p.editInstruction(editpoint, MKInstruction(MVF, end-begin))
	return nil
}

func (p *Parser) parseWhile(block Block) error {
	tks := discardOne(block.Tokens)
	tks, r := readUntilToken(tks, DO)
	whilestart := p.fragmentSize()
	tp, e := p.parseExpression(tks, nil, false)
	if e != nil {
		return e
	}
	if tp.Primitive() != BOOL {
		return errors.New("Expecting boolean expression")
	}
	tail, e := expect(r, DO)
	if e != nil {
		return e
	}
	if e := unexpect(tail); e != nil {
		return e
	}
	editpoint := p.addInstruction(MKInstruction(MVF))
	begin := p.fragmentSize()
	p.openScope()
	e = p.parseBlocks(block.Children, Function)
	if e != nil {
		return e
	}
	p.closeScope()
	end := p.fragmentSize()
	p.addInstruction(MKInstruction(MVR, whilestart-end-1))
	p.editInstruction(editpoint, MKInstruction(MVF, end-begin+1))
	return nil
}

func (p *Parser) parseElse(block Block) error {
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
	if scp == Function {
		switch name {
		case "return":
			return p.parseReturn(block)
		case "direct":
			return p.parseDirect(block)
		case "if":
			return p.parseIf(block)
		case "while":
			return p.parseWhile(block)
		case "else":
			return p.parseElse(block)
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

func (p *Parser) parseById(block Block, scp ScopeCtx) error {
	tks, e := splitByToken(block.Tokens, func(tk Token) bool { return tk == ASSIGN }, make([]struct {
		open  Token
		close Token
	}, 0), false, false, false)
	if e != nil {
		return e
	}
	if len(tks) > 2 {
		return errors.New("Multiassignment is not implemented")
	}
	t, e := p.parseExpression(tks[len(tks)-1], block.Children, false)
	if e != nil {
		return e
	}
	if len(tks) == 2 {
		if t == Void {
			return errors.New("No value returned")
		}
		//return p.parseAssignment(tks[0])
	} else if t != Void {
		p.addInstruction(MKInstruction(POP))
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
		ret, e := p.parseExpression(block.Tokens, block.Children, false)
		if ret != Void {
			p.addInstruction(MKInstruction(POP))
		}
		return e
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
