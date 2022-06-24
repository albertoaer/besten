package main

import (
	. "github.com/Besten/internal/runtime"
)

/*
Besten bytecode tester is a tool whose purpose is to ensure bytecode works as expected
*/
func main() {
	vm := NewVM()
	program := Fragment{
		MKInstruction(KVC),
		MKInstruction(SET, "map"),
		MKInstruction(PSH, 33),
		MKInstruction(PSH, "key"),
		MKInstruction(GET, "map"),
		MKInstruction(ATT),
	}
	vm.LoadSymbol(Symbol{Name: "main", Source: program})
	_, err := vm.Spawn("main")
	if err != nil {
		panic(err)
	}
}
