package parser

import (
	"errors"
	"fmt"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type Parser struct {
	env       ImportEnv
	rootscope *Scope
	scopes    map[string]struct {
		origin  *Scope
		current *Scope
	}
	symbols       map[string]Symbol
	fragmenttrack []string
	modulename    string
}

type ImportEnv interface {
	Symbols() map[string]Symbol
	Scope() *Scope
	Native(name string) (map[string]Symbol, *Scope, error)
	File(name string) (map[string]Symbol, *Scope, error)
}

type ScopeCtx uint

const (
	Global   ScopeCtx = 0
	Function ScopeCtx = 1
	Loop     ScopeCtx = 2
)

func NewParser(env ImportEnv) *Parser {
	p := &Parser{env, env.Scope(), make(map[string]struct {
		origin  *Scope
		current *Scope
	}), make(map[string]Symbol), make([]string, 0), ""}
	p.addSymbols(env.Symbols())
	injectBuiltinFunctions(p.rootscope.Functions)
	injectBuiltinOperators(p.rootscope.Operators)
	return p
}

func (p *Parser) makeScope() *Scope {
	return p.rootscope.Open(true)
}

func (p *Parser) activeFragment() string {
	if len(p.fragmenttrack) == 0 {
		panic("No active fragment")
	}
	return p.fragmenttrack[len(p.fragmenttrack)-1]
}

func (p *Parser) addSymbol(symbol Symbol) {
	if _, e := p.symbols[symbol.Name]; e {
		panic("Overwriting symbol")
	}
	p.symbols[symbol.Name] = symbol
}

func (p *Parser) addSymbols(symbols map[string]Symbol) {
	for _, v := range symbols {
		p.addSymbol(v)
	}
}

func (p *Parser) addInstruction(instruction Instruction) int {
	s := p.symbols[p.activeFragment()]
	s.Append(instruction)
	p.symbols[p.activeFragment()] = s
	return len(p.symbols[p.activeFragment()].Source) - 1
}

//You must prevail in the same fragment to get an instruction
func (p *Parser) getInstruction(pos int) Instruction {
	s := p.symbols[p.activeFragment()]
	return s.Source[pos]
}

//You must prevail in the same fragment to edit an instruction
func (p *Parser) editInstruction(pos int, instruction Instruction) {
	s := p.symbols[p.activeFragment()]
	s.Source[pos] = instruction
	p.symbols[p.activeFragment()] = s
}

func (p *Parser) addInstructions(instructions []Instruction) int {
	s := p.symbols[p.activeFragment()]
	for _, i := range instructions {
		s.Append(i)
	}
	p.symbols[p.activeFragment()] = s
	return len(p.symbols[p.activeFragment()].Source) - len(instructions)
}

func (p *Parser) fragmentSize() int {
	return len(p.symbols[p.activeFragment()].Source)
}

func (p *Parser) openFragmentFor(name string, args int) {
	if _, v := p.symbols[name]; !v {
		p.symbols[name] = Symbol{Name: name, Source: make(Fragment, 0), Args: args}
	}
	if _, v := p.scopes[name]; !v {
		scope := p.makeScope()
		p.scopes[name] = struct {
			origin  *Scope
			current *Scope
		}{scope, scope}
	}
	p.fragmenttrack = append(p.fragmenttrack, name)
}

func (p *Parser) backToFragment() {
	if len(p.fragmenttrack) == 0 {
		panic("No symbol to go back to")
	}
	p.fragmenttrack = p.fragmenttrack[:len(p.fragmenttrack)-1]
}

func (p *Parser) currentScope() *Scope {
	if len(p.fragmenttrack) == 0 {
		return p.rootscope
	}
	return p.scopes[p.activeFragment()].current
}

func (p *Parser) currentScopeOrigin() *Scope {
	if len(p.fragmenttrack) == 0 {
		return p.rootscope
	}
	return p.scopes[p.activeFragment()].origin
}

func (p *Parser) rootScope() *Scope {
	return p.rootscope
}

func (p *Parser) openScope() {
	item := p.scopes[p.activeFragment()]
	item.current = item.current.Open(false)
	p.scopes[p.activeFragment()] = item
}

func (p *Parser) closeScope() (err error) {
	item := p.scopes[p.activeFragment()]
	item.current, err = item.current.Close()
	p.scopes[p.activeFragment()] = item
	return
}

func (p *Parser) ParseCode(blocks []Block, modulename string) error {
	p.modulename = modulename
	e := p.parseBlocks(blocks, Global)
	return e
}

func (p *Parser) GetSymbolNameFor(name string, operator bool, callers []OBJType) (string, error) {
	sym, e := p.findFunction(name, operator, callers)
	if !e {
		symboltype := "function"
		if operator {
			symboltype = "operator"
		}
		return "", errors.New(fmt.Sprintf("There is no %s symbol %s/%d valid for %s", symboltype, name, len(callers), FnCArrRepr(callers)))
	}
	return sym.CName, nil
}

func (p *Parser) GetSymbols() map[string]Symbol {
	return p.symbols
}
