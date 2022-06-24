package runtime

import (
	"errors"
	"fmt"
)

type VM struct {
	pc             int                 //Program counter
	program        []Instruction       //Loaded instructions
	context        []*Context          //Stack of context
	symbols        map[string]int      //Addresses linked by names
	markedcontexts map[string]*Context //Saved contexts
	callstack      []int               //Stack of calls
	functionstack  []Object            //Stack of function results
}

func NewVM() *VM {
	vm := &VM{0,
		make([]Instruction, 0),
		make([]*Context, 1),
		make(map[string]int),
		make(map[string]*Context),
		make([]int, 0),
		make([]Object, 0)}
	vm.context[0] = NContext()
	return vm
}

/*
Symbol zone
*/
func (vm *VM) LoadSymbol(symbol Symbol) {
	vm.program = append(vm.program, symbol...)
}

func (vm *VM) LoadSymbols(symbols []Symbol) {
	for _, symbol := range symbols {
		vm.program = append(vm.program, symbol...)
	}
}

/*
Manipulation zone
*/

func (vm *VM) ActiveContext() (*Context, error) {
	if len(vm.context) > 0 {
		return vm.context[len(vm.context)-1], nil
	}
	return nil, errors.New("No active context")
}

func (vm *VM) Get(name string) (Object, error) {
	return vm.context[len(vm.context)-1].Get(name)
}

func (vm *VM) Set(name string, value Object) {
	vm.context[len(vm.context)-1].Set(name, value)
}

func (vm *VM) Open() {
	vm.context = append(vm.context, vm.context[len(vm.context)-1].Fork())
}

func (vm *VM) Close() {
	vm.context = vm.context[:len(vm.context)-1]
}

func (v *VM) Push(value Object) {
	v.functionstack = append(v.functionstack, value)
}

func (v *VM) Pop() (res Object, err error) {
	if len(v.functionstack) > 0 {
		res = v.functionstack[len(v.functionstack)-1]
		v.functionstack = v.functionstack[:len(v.functionstack)-1]
	} else {
		err = errors.New("Empty stack")
	}
	return
}

func (v *VM) Clear() {
	v.functionstack = make([]Object, 0)
}

/*
Run zone
*/

/*
Each operation has a fixed number of arguments,
the first ones are taken from the operands, the rest from the stack
*/
func (vm *VM) FetchArguments(num uint, operands ...Object) (args []Object, err error) {
	args = operands
	if uint(len(operands)) > num {
		err = errors.New("Too much operands for instruction")
	} else {
		ref := num - uint(len(operands))
		for ref > 0 {
			res, e := vm.Pop()
			if e != nil {
				err = e
				break
			}
			args = append(args, res)
			ref--
		}
	}
	return
}

func (vm *VM) RunInstruction(ins Instruction, addr int) (err error) {
	fn, found := operations[ins.Code]
	if !found {
		err = errors.New("No operation matchs the instruction code provided")
	} else {
		defer func() {
			if e := recover(); e != nil {
				err = errors.New(fmt.Sprintf("[pc : %d, icode : %d] Runtime error: %v", addr, ins.Code, e))
			}
		}()
		args, e := vm.FetchArguments(fn.Operands, ins.Operands...)
		if e == nil {
			fn.Action(vm, args...)
		} else {
			err = e
		}
	}
	return
}

func (vm *VM) RunOne() (done bool, err error) {
	if int(vm.pc) >= len(vm.program) {
		return true, nil
	}
	instruction := vm.program[vm.pc]
	vm.pc++
	err = vm.RunInstruction(instruction, vm.pc-1)
	done = int(vm.pc) >= len(vm.program)
	return
}

func (vm *VM) Run() error {
	for {
		done, err := vm.RunOne()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
}
