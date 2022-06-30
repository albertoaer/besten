package modules

import "github.com/besten/internal/runtime"

func (m *Modules) RequestSymbol(requester int, name string, args int) *runtime.Symbol {
	m.symbolmx.Lock()
	defer m.symbolmx.Unlock()
	m.symbols[name] = &runtime.Symbol{Name: name, Source: make(runtime.Fragment, 0), Args: args}
	return m.symbols[name]
}

func (m *Modules) collectSymbols() map[string]runtime.Symbol {
	m.symbolmx.Lock()
	defer m.symbolmx.Unlock()
	symbols := make(map[string]runtime.Symbol)
	for k, v := range m.symbols {
		symbols[k] = *v
	}
	return symbols
}
