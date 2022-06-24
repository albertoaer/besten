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

func (m *Modules) File(name string) (symbols map[string]Symbol, scope *parser.Scope, err error) {
	reader, err := os.Open(name)
	if err != nil {
		return
	}
	blocks, err := GetBlocks(reader)
	if err != nil {
		return
	}
	prettyTokens(blocks, "")
	module_parser := parser.NewParser(m)
	symbols, err = module_parser.GenerateCode(blocks, name)
	return
}
