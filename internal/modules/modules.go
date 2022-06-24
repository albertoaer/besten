package modules

import (
	"errors"
	"fmt"
	"os"

	. "github.com/Besten/internal/lexer"
	"github.com/Besten/internal/parser"
	. "github.com/Besten/internal/runtime"
)

type Modules struct {
	virtuals Symbols
}

func New() *Modules {
	return &Modules{make(Symbols, 0)}
}

func (m *Modules) Load(symbols Symbols) {
	for k, v := range symbols {
		m.virtuals[k] = v
	}
}

func (_ *Modules) Native(name string) (symbols Symbols, err error) {
	err = errors.New("Native modules not available yet")
	return
}

func prettyTokens(blocks []Block, tabs string) {
	for _, b := range blocks {
		fmt.Printf(tabs+"%v\n", b.Tokens)
		prettyTokens(b.Children, tabs+"  ")
	}
}

func (m *Modules) File(name string) (symbols Symbols, err error) {
	reader, err := os.Open(name)
	if err != nil {
		return
	}
	blocks, err := GetBlocks(reader)
	if err != nil {
		return
	}
	prettyTokens(blocks, "")
	module_parser := parser.NewParser(m, m.virtuals)
	return module_parser.GetSymbols(blocks)
}
