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
		MKInstruction(ADD, 5),
	}
	vm.LoadSymbol(Symbol{Name: "main", Source: program, BuiltInfo: struct {
		Args    int
		Varargs bool
	}{0, false}})
	pid, err := vm.Spawn("main")
	if err != nil {
		panic(err)
	}
	err = vm.Wait(pid)
	if err != nil {
		panic(err)
	}
}
