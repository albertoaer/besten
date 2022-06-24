package main

import (
	"github.com/Besten/internal/runtime"
)

/*
Besten bytecode tester is a tool whose purpose is to ensure bytecode works as expected
*/
func main() {
	vm := runtime.NewVM()
	program := []runtime.Instruction{
		{Code: runtime.DEF, Operands: []runtime.Object{"x", 2}},
		{Code: runtime.SET, Operands: []runtime.Object{"prueba"}},
		{Code: runtime.RET, Operands: []runtime.Object{}},
		{Code: runtime.PSH, Operands: []runtime.Object{50}},
		{Code: runtime.CLL, Operands: []runtime.Object{"x"}},
		{Code: runtime.GET, Operands: []runtime.Object{"prueba"}},
		{Code: runtime.ADD, Operands: []runtime.Object{20}},
	}
	vm.LoadSymbol(program)
	err := vm.Run()
	if err != nil {
		panic(err)
	}
}
