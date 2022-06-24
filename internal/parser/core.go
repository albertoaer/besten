package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

func parseArguments(tks []Token) (args []string, types []OBJType, usetypes bool, varargs bool, err error) {
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
	argtk, err := splitByToken(preargs, func(tk Token) bool { return tk == COMA }, []struct {
		open  Token
		close Token
	}{{CBOPEN, CBCLOSE}}, false, false, false)
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
				err = errors.New(fmt.Sprintf("Unexpected token: %s", v[0].Data))
				return
			}
		}
		nm, n, e := expectT(n, IdToken)
		if e != nil {
			err = e
			return
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
				err = errors.New(fmt.Sprintf("Expecting type for argument: %s", nm.Data))
				return
			}
			tp, e := solveTypeFromTokens(n)
			if e != nil {
				err = e
				return
			}
			types = append(types, tp)
		} else {
			if len(n) > 0 {
				err = errors.New(fmt.Sprintf("Unexpected type for template argument: %s", nm.Data))
				return
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
	var types []OBJType
	var usetypes bool
	varargs := false
	if next(tks, DOUBLES) {
		args, types, usetypes, varargs, e = parseArguments(tks)
		if e != nil {
			return e
		}
	} else {
		ntk, e := expect(tks, DO)
		if e != nil {
			return e
		}
		if e := unexpect(ntk); e != nil {
			return e
		}
	}
	if len(args) == 0 {
		//Generate all symbols with no arguments
		usetypes = true
	}
	if operator && (len(args) > 2 || len(args) < 1) {
		return errors.New("Operator must be unary or binary")
	}
	if operator && varargs {
		return errors.New("Operator can not have varargs")
	}
	template := FunctionTemplate{Args: args, Varargs: varargs, Children: block.Children}
	if usetypes {
		if operator {
			p.currentScope().Operators.SaveSymbolHolder(name.Data)
		} else {
			p.currentScope().Functions.SaveSymbolHolder(name.Data)
		}
		_, e = p.generateFunctionFromRawTemplate(name.Data, operator, types, &template)
	} else {
		e = p.generateFunctionTemplate(name.Data, operator, template)
	}
	return e
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
	*p.currentScope().returnLnFlag = returnLnFlag{true, true}
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

func (p *Parser) parseIf(tks []Token, children []Block, scp ScopeCtx) error {
	tks = discardOne(tks)
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
	e = p.parseBlocks(children, scp)
	if e != nil {
		return e
	}
	if e = p.closeScope(); e != nil {
		return e
	}
	end := p.fragmentSize()
	p.editInstruction(editpoint, MKInstruction(MVF, end-begin))
	p.currentScope().ifFlags.submitIf(editpoint, end-begin)
	return nil
}

func (p *Parser) parseElse(block Block, scp ScopeCtx) error {
	if !p.currentScope().ifFlags.afterif {
		return errors.New("Unexpected block else")
	}
	{
		//Add an offset of one to the previous if condition
		p.editInstruction(p.currentScope().ifFlags.lastifhead.idx,
			MKInstruction(MVF, p.currentScope().ifFlags.lastifhead.offset+1))
		//Add an offset of one to the last skip
		l := len(p.currentScope().ifFlags.allifskips)
		if l > 0 {
			p.currentScope().ifFlags.allifskips[l-1].offset += 1
		}
	}
	tks := discardOne(block.Tokens)
	prevjump := p.addInstruction(MKInstruction(NOP))
	p.currentScope().ifFlags.submitIfSkip(prevjump)
	begin := p.fragmentSize()
	if next(tks, IF) {
		e := p.parseIf(tks, block.Children, scp)
		if e != nil {
			return e
		}
	} else {
		tail, e := expect(tks, DO)
		if e != nil {
			return e
		}
		if e := unexpect(tail); e != nil {
			return e
		}
		p.openScope()
		e = p.parseBlocks(block.Children, scp)
		if e != nil {
			return e
		}
		if e = p.closeScope(); e != nil {
			return e
		}
	}
	end := p.fragmentSize()
	offset := end - begin
	for i := range p.currentScope().ifFlags.allifskips {
		p.currentScope().ifFlags.allifskips[i].offset += offset
		p.editInstruction(p.currentScope().ifFlags.allifskips[i].idx,
			MKInstruction(MVR, p.currentScope().ifFlags.allifskips[i].offset))
	}
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
	{
		tail, e := expect(r, DO)
		if e != nil {
			return e
		}
		if e := unexpect(tail); e != nil {
			return e
		}
	}
	editpoint := p.addInstruction(MKInstruction(MVF))
	begin := p.fragmentSize()
	p.currentScope().forkLoopInfo()
	p.openScope()
	e = p.parseBlocks(block.Children, Loop)
	if e != nil {
		return e
	}
	if e = p.closeScope(); e != nil {
		return e
	}
	end := p.fragmentSize()
	p.addInstruction(MKInstruction(MVR, whilestart-end-1))
	p.editInstruction(editpoint, MKInstruction(MVF, end-begin+1))
	p.currentScope().loopInfo.solveJumps(p, whilestart+1, end)
	p.currentScope().closeLoopInfo()
	return nil
}

func (p *Parser) parseFor(block Block) error {
	tks := discardOne(block.Tokens)
	tks, r := readUntilToken(tks, DO)
	sides, e := splitByToken(tks, func(t Token) bool { return t == IN }, []struct {
		open  Token
		close Token
	}{{POPEN, PCLOSE}, {BOPEN, BCLOSE}}, false, false, false)
	if e != nil {
		return e
	}
	if len(sides) != 2 {
		return errors.New("Expecting 'in' keyword")
	}
	itertp, e := p.parseExpression(sides[1], nil, false)
	var name string
	{
		t, r, e := expectT(sides[0], IdToken)
		if e != nil {
			return e
		}
		if e = unexpect(r); e != nil {
			return e
		}
		name = t.Data
	}
	if e != nil {
		return e
	}
	p.currentScope().forkLoopInfo()
	p.openScope()
	p.currentScope().CreateVariable(name, itertp, true, false)
	setiter, e := p.currentScope().SetVariableIns(name, itertp)
	if e != nil {
		return e //Weird, variable has just been set
	}
	p.addInstruction(setiter)
	{
		tail, e := expect(r, DO)
		if e != nil {
			return e
		}
		if e := unexpect(tail); e != nil {
			return e
		}
	}
	getiter, _, e := p.currentScope().GetVariableIns(name)
	if e != nil {
		return e //Weird, variable has just been set
	}
	forstart := p.fragmentSize()
	boolean, e := p.processFunctionCall("end", false, []OBJType{itertp}, [][]Instruction{getiter.Fragment()})
	if e != nil {
		return e
	}
	if boolean.Primitive() != BOOL {
		return errors.New("Expected Boolean from end")
	}
	editpoint := p.addInstruction(MKInstruction(MVT))
	begin := p.fragmentSize()
	e = p.parseBlocks(block.Children, Loop)
	if e != nil {
		return e
	}
	if e = p.closeScope(); e != nil {
		return e
	}
	callnextpoint := p.fragmentSize()
	nitertp, e := p.processFunctionCall("next", false, []OBJType{itertp}, [][]Instruction{getiter.Fragment()})
	if e != nil {
		return e
	}
	if !CompareTypes(itertp, nitertp) {
		return errors.New(fmt.Sprintf("Type %s does not match %s", itertp.TypeName(), nitertp.TypeName()))
	}
	p.addInstruction(setiter)
	end := p.fragmentSize()
	p.addInstruction(MKInstruction(MVR, forstart-end-1))
	p.editInstruction(editpoint, MKInstruction(MVT, end-begin+1))
	p.currentScope().loopInfo.solveJumps(p, callnextpoint-1, end)
	p.currentScope().closeLoopInfo()
	return nil
}

func (p *Parser) parseLoopJump(block Block, skip bool) error {
	line := discardOne(block.Tokens)
	var err error
	loopTarget := 0
	if isNext(line) {
		var n Token
		n, line, err = expectT(line, IntegerToken)
		if err != nil {
			return err
		}
		loopTarget, err = strconv.Atoi(n.Data)
		if err != nil {
			return err
		}
	}
	if err = unexpect(line); err != nil {
		return err
	}
	idx := p.addInstruction(MKInstruction(MVR))
	info := p.currentScope().loopInfo
	for loopTarget > 0 {
		if info.father == nil {
			return errors.New("Too much loop depth")
		}
		info = info.father
		loopTarget--
	}
	info.insertJump(idx, skip)
	return nil
}

func (p *Parser) parseDrop(block Block) error {
	tks := discardOne(block.Tokens)
	t, tks, err := expectT(tks, KeywordToken)
	if err != nil {
		return err
	}
	var fc *FunctionCollection
	var fname Token
	if t.Data == "fn" {
		fc = p.currentScope().Functions
		fname, tks, err = expectT(tks, IdToken)
	} else if t.Data == "op" {
		fc = p.currentScope().Operators
		fname, tks, err = expectT(tks, OperatorToken)
	} else {
		return errors.New(fmt.Sprintf("Unexpected keyword: %s", t.Data))
	}
	if err != nil {
		return err
	}
	if !isNext(tks) || next(tks, DOUBLES) {
		if next(tks, DOUBLES) && isNext(discardOne(tks)) {
			var numt Token
			numt, tks, err = expectT(discardOne(tks), IntegerToken)
			if err != nil {
				return err
			}
			n, err := strconv.Atoi(numt.Data)
			if err != nil {
				return err
			}
			if n < 0 {
				return errors.New("Negative number of arguments")
			}
			if t.Data == "op" && (n < 1 || n > 2) {
				return errors.New("Operator must be unary or binary")
			}
			fc.DropTemplate(fname.Data, n)
			fc.DropSymbol(fname.Data, n)
		} else {
			fc.DropAllTemplatesOf(fname.Data)
			fc.DropAllSymbolOf(fname.Data)
		}
	}
	return unexpect(tks)
}

func (p *Parser) parseAlias(block Block) error {
	tks := discardOne(block.Tokens)
	t, tks, err := expectT(tks, KeywordToken)
	if err != nil {
		return err
	}
	var srcfc *FunctionCollection
	var srcname Token
	if t.Data == "fn" {
		srcfc = p.currentScope().Functions
		srcname, tks, err = expectT(tks, IdToken)
	} else if t.Data == "op" {
		srcfc = p.currentScope().Operators
		srcname, tks, err = expectT(tks, OperatorToken)
	} else {
		return errors.New(fmt.Sprintf("Unexpected keyword: %s", t.Data))
	}
	if err != nil {
		return err
	}
	args := -1
	if next(tks, DOUBLES) {
		tks = discardOne(tks)
		if nextT(tks, IntegerToken) {
			var numt Token
			numt, tks, err = expectT(tks, IntegerToken)
			if err != nil {
				return err
			}
			n, err := strconv.Atoi(numt.Data)
			if err != nil {
				return err
			}
			if n < 0 {
				return errors.New("Negative number of arguments")
			}
			args = n
		}
	}
	t2, tks, err := expectT(tks, KeywordToken)
	if err != nil {
		return err
	}
	var dstfc *FunctionCollection
	var dstname Token
	if t2.Data == "fn" {
		dstfc = p.currentScope().Functions
		dstname, tks, err = expectT(tks, IdToken)
	} else if t2.Data == "op" {
		dstfc = p.currentScope().Operators
		dstname, tks, err = expectT(tks, OperatorToken)
	} else {
		return errors.New(fmt.Sprintf("Unexpected keyword: %s", t2.Data))
	}
	if err != nil {
		return err
	}
	if (args == 0 || args > 2) && (t.Data == "op" || t2.Data == "op") {
		return errors.New("Operator must be unary or binary")
	}
	if args < 0 {
		srcfc.CopyAllTemplatesOf(srcname.Data, dstname.Data, dstfc)
		srcfc.CopyAllSymbolsOf(srcname.Data, dstname.Data, dstfc)
	} else {
		srcfc.CopyTemplate(srcname.Data, args, dstname.Data, dstfc)
		srcfc.CopySymbol(srcname.Data, args, dstname.Data, dstfc)
	}
	return unexpect(tks)
}

func (p *Parser) parseByKeyword(name string, block Block, scp ScopeCtx) error {
	if scp == Global {
		switch name {
		case "require":
			return p.loadModuleAux(block.Tokens, p.env.Native)
		case "import":
			return p.loadModuleAux(block.Tokens, p.env.File)
		case "fn", "op":
			return p.parseFunction(block, name == "op")
		case "drop":
			return p.parseDrop(block)
		case "alias":
			return p.parseAlias(block)
		}
	}
	if scp == Function || scp == Loop {
		switch name {
		case "return":
			return p.parseReturn(block)
		case "direct":
			return p.parseDirect(block)
		case "if":
			return p.parseIf(block.Tokens, block.Children, scp)
		case "while":
			return p.parseWhile(block)
		case "else":
			return p.parseElse(block, scp)
		case "for":
			return p.parseFor(block)
		case "var", "val":
			return p.parseDefinition(block, name == "val")
		case "do":
			if err := unexpect(discardOne(block.Tokens)); err != nil {
				return err
			}
			return p.parseBlocks(block.Children, scp)
		}
	}
	if scp == Loop {
		switch name {
		case "break":
			return p.parseLoopJump(block, true)
		case "continue":
			return p.parseLoopJump(block, false)
		}
	}
	switch name {
	case "omit":
		return nil //The block is omitted
	}
	return errors.New(fmt.Sprintf("Unexpected keyword found: %s", name))
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
	ifFlags := p.currentScope().ifFlags
	ifFlags.altered = false
	returnLnFlag := p.currentScope().returnLnFlag
	returnLnFlag.altered = false
	var err error
	if len(block.Tokens) == 0 {
		panic("Void block")
	}
	if nextT(block.Tokens, KeywordToken) {
		err = p.parseByKeyword(block.Tokens[0].Data, block, scp)
	} else if scp == Function || scp == Loop {
		var ret OBJType
		ret, err = p.parseExpression(block.Tokens, block.Children, false)
		if ret != Void {
			p.addInstruction(MKInstruction(POP))
		}
	} else {
		err = errors.New(fmt.Sprintf("Unexpected %s found: %s", block.Tokens[0].Kind.Representation(), block.Tokens[0].Data))
	}
	if !ifFlags.altered {
		ifFlags.Reset()
	}
	if !returnLnFlag.altered {
		returnLnFlag.Reset()
	}
	return err
}

func (p *Parser) parseBlocks(blocks []Block, scp ScopeCtx) error {
	for _, block := range blocks {
		if e := p.parseBlock(block, scp); e != nil {
			strerr := e.Error()
			if !strings.ContainsRune(strerr, '\n') {
				lineid := strconv.Itoa(block.Begin)
				if block.Begin != block.End {
					lineid = lineid + ".." + strconv.Itoa(block.End)
				}
				return errors.New(fmt.Sprintf("[Error in line (%s)]\n\t%s", lineid, strerr))
			}
			return e
		}
	}
	return nil
}
