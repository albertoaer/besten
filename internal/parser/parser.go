package parser

import (
	"fmt"

	. "github.com/Besten/internal/lexer"
	. "github.com/Besten/internal/runtime"
)

type Parser struct {
	env       ImportEnv
	envId     int
	rootscope *Scope
	scopes    map[string]struct {
		origin  *Scope
		current *Scope
	}
	symbols       map[string]*Symbol
	fragmenttrack []string
}

type ImportEnv interface {
	LoadModule(int, string) (Module, error)
	RequestSymbol(int, string, int) *Symbol
}

type ScopeCtx uint

const (
	Global   ScopeCtx = 0
	Function ScopeCtx = 1
	Loop     ScopeCtx = 2
)

func NewParser(name string, envId int, env ImportEnv) *Parser {
	p := &Parser{env, envId, NewScope(name), make(map[string]struct {
		origin  *Scope
		current *Scope
	}), make(map[string]*Symbol), make([]string, 0)}
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

func (p *Parser) addInstruction(instruction Instruction) int {
	p.symbols[p.activeFragment()].Append(instruction)
	return len(p.symbols[p.activeFragment()].Source) - 1
}

//You must prevail in the same fragment to get an instruction
func (p *Parser) getInstruction(pos int) Instruction {
	s := p.symbols[p.activeFragment()]
	return s.Source[pos]
}

//You must prevail in the same fragment to edit an instruction
func (p *Parser) editInstruction(pos int, instruction Instruction) {
	p.symbols[p.activeFragment()].Source[pos] = instruction
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
		p.symbols[name] = p.env.RequestSymbol(p.envId, name, args)
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

func (p *Parser) openScope() {
	item := p.scopes[p.activeFragment()]
	item.current = item.current.Open(false)
	p.scopes[p.activeFragment()] = item
}

func (p *Parser) openForeignScope(s *Scope) {
	item := p.scopes[p.activeFragment()]
	item.current = item.current.OpenForeignScope(s)
	p.scopes[p.activeFragment()] = item
}

func (p *Parser) closeScope() (err error) {
	item := p.scopes[p.activeFragment()]
	item.current, err = item.current.Close()
	p.scopes[p.activeFragment()] = item
	return
}

func (p *Parser) ParseCode(blocks []Block) error {
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
		return "", fmt.Errorf("There is no %s symbol %s/%d valid for %s", symboltype, name, len(callers), FnCArrRepr(callers))
	}
	return sym.CName, nil
}

func (p *Parser) GetModule() Module {
	return p.rootscope.DataModule
}
