package modules

import (
	"errors"
	"fmt"
	"os"

	. "github.com/Besten/internal/lexer"
	"github.com/Besten/internal/parser"
	. "github.com/Besten/internal/runtime"
)

func prettyTokens(blocks []Block, tabs string) {
	for _, b := range blocks {
		fmt.Printf(tabs+"%v\n", b.Tokens)
		prettyTokens(b.Children, tabs+"  ")
	}
}

type Modules struct {
	symbols map[string]Symbol
	scope   *parser.Scope
}

func New() *Modules {
	return &Modules{make(map[string]Symbol), parser.NewScope()}
}

func (m *Modules) Symbols() map[string]Symbol {
	return m.symbols
}

func (m *Modules) Scope() *parser.Scope {
	return m.scope
}

func (_ *Modules) Native(name string) (map[string]Symbol, *parser.Scope, error) {
	return nil, parser.NewScope(), errors.New("Native modules not available yet")
}

func (m *Modules) FileParser(name string) (*parser.Parser, error) {
	reader, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	blocks, err := GetBlocks(reader)
	if err != nil {
		return nil, err
	}
	//prettyTokens(blocks, "")
	module_parser := parser.NewParser(m)
	err = module_parser.ParseCode(blocks, name)
	return module_parser, err
}

func (m *Modules) File(name string) (symbols map[string]Symbol, scope *parser.Scope, err error) {
	module_parser, e := m.FileParser(name)
	if e != nil {
		err = e
		return
	}
	symbols = module_parser.GetSymbols()
	return
}

func (m *Modules) MainFile(name string) (symbols map[string]Symbol, cname string, err error) {
	module_parser, e := m.FileParser(name)
	if e != nil {
		err = e
		return
	}
	cname, err = module_parser.GetSymbolNameFor("main", false, []parser.OBJType{parser.VecOf(parser.Str)})
	if err != nil {
		return
	}
	symbols = module_parser.GetSymbols()
	return
}
