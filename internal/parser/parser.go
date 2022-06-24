package parser

import (
	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type Parser struct {
	env           ImportEnv
	rootscope     *Scope
	scopes        map[string]*Scope
	symbols       map[string]Symbol
	fragmenttrack []string
}

type ImportEnv interface {
	Symbols() map[string]Symbol
	Scope() *Scope
	Native(name string) (map[string]Symbol, *Scope, error)
	File(name string) (map[string]Symbol, *Scope, error)
}

func NewParser(env ImportEnv) *Parser {
	p := &Parser{env, env.Scope(), make(map[string]*Scope), make(map[string]Symbol), make([]string, 0)}
	p.addSymbols(env.Symbols())
	injectBuiltinOperators(p.rootscope.OpSymbols)
	return p
}

func (p *Parser) makeScope() *Scope {
	return p.rootscope.Open()
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
	p.symbols[p.activeFragment()].Append(instruction)
	return len(p.symbols[p.activeFragment()].Source) - 1
}

func (p *Parser) getInstruction(fr string, idx int) *Instruction {
	return &p.symbols[fr].Source[idx]
}

func (p *Parser) fragmentSize() int {
	return len(p.symbols[p.activeFragment()].Source)
}

func (p *Parser) open(name string) {
	if _, v := p.symbols[name]; !v {
		p.symbols[name] = Symbol{Name: name, Source: make(Fragment, 0)}
	}
	if _, v := p.scopes[name]; !v {
		p.scopes[name] = p.makeScope()
	}
	p.fragmenttrack = append(p.fragmenttrack, name)
}

func (p *Parser) back() {
	if len(p.fragmenttrack) == 0 {
		panic("No symbol to go back to")
	}
	p.fragmenttrack = p.fragmenttrack[:len(p.fragmenttrack)-1]
}

func (p *Parser) currentScope() *Scope {
	return p.scopes[p.activeFragment()]
}

func (p *Parser) rootScope() *Scope {
	return p.rootscope
}

func (p *Parser) openScope() {
	p.scopes[p.activeFragment()] = p.scopes[p.activeFragment()].Open()
}

func (p *Parser) closeScope() {
	p.scopes[p.activeFragment()] = p.scopes[p.activeFragment()].Close()
}

func (p *Parser) GenerateCode(blocks []Block) (map[string]Symbol, error) {
	p.open("main")
	e := p.parseBlocks(blocks, Global)
	return p.symbols, e
}
