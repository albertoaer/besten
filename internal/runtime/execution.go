package runtime

import (
	"errors"
	"fmt"
)

type PID *Process

type VM struct {
	symbols  map[string]Symbol //Loaded instructions
	embedded map[string]EmbeddedFunction
}

type Process struct {
	machine       *VM    //The virtual machine were the process is running on
	parent        PID    //Parent process (the one who called)
	state         Object //Object state
	pc            int    //Current instruction
	symbol        Symbol
	functionstack *FunctionStack //Function stack
	done          chan error     //Is Process done?
	callstack     *CallStack
	items         *ItemManager
}

/*
VM Zone
*/

func NewVM() *VM {
	vm := &VM{make(map[string]Symbol), make(map[string]EmbeddedFunction)}
	return vm
}

func (vm *VM) spawn(parent PID, fr string, stack *FunctionStack) (PID, error) {
	sym, ex := vm.symbols[fr]
	if !ex {
		return nil, errors.New(fmt.Sprintf("Symbol %s not found", fr))
	}
	i, err := ItemManagerForCall(stack, sym.BuiltInfo.Args, sym.BuiltInfo.Varargs, 20)
	if err != nil {
		return nil, err
	}
	process := &Process{vm, parent, nil, 0, sym, stack,
		make(chan error), NewCallStack(), i}
	go process.run()
	return process, nil
}

func (vm *VM) Spawn(fr string) (PID, error) {
	return vm.spawn(nil, fr, NewFunctionStack(20))
}

func (vm *VM) InitSpawn(fr string, stack []Object) (PID, error) {
	fs := NewFunctionStack(20)
	if stack != nil {
		for _, o := range stack {
			fs.Push(o)
		}
	}
	return vm.spawn(nil, fr, fs)
}

func (vm *VM) Wait(process PID) error {
	e, on := <-process.done
	if on {
		close(process.done)
		return e
	} else {
		return errors.New("Process already closed")
	}
}

func (vm *VM) LoadSymbol(entry Symbol) {
	if _, e := vm.symbols[entry.Name]; e {
		panic("Overwriting symbol")
	}
	vm.symbols[entry.Name] = entry
}

func (vm *VM) LoadSymbols(symbols map[string]Symbol) {
	for _, symbol := range symbols {
		vm.LoadSymbol(symbol)
	}
}

func (vm *VM) Inject(embedded EmbeddedFunction) {
	vm.embedded[embedded.Name] = embedded
}

func (vm *VM) InjectNamed(name string, embedded EmbeddedFunction) {
	vm.embedded[name] = embedded
}

/*
Process Zone
*/

func (proc *Process) Spawn(fr string) PID {
	child, err := proc.machine.spawn(proc, fr, proc.functionstack.Clone())
	if err != nil {
		panic(err)
	}
	return child
}

func (proc *Process) ReturnLastPoint() {
	point := proc.callstack.Current()
	if point == nil {
		/*
			Send to last instruction
			Will force process to end
		*/
		proc.pc = len(proc.symbol.Source) + 1
	} else {
		proc.callstack.Pop()
		proc.symbol = point.fragment
		proc.pc = point.pc
	}
}

func (proc *Process) SavePoint() {
	proc.callstack.Insert(proc.pc, proc.symbol, proc.items)
}

func (proc *Process) ChangeFragment(name string) {
	if proc.symbol.Name != name {
		proc.symbol = proc.machine.symbols[name]
	}
	items, err := ItemManagerForCall(proc.functionstack, proc.symbol.BuiltInfo.Args, proc.symbol.BuiltInfo.Varargs, 20)
	if err != nil {
		panic(err)
	}
	proc.items = items
	proc.pc = 0
}

func (proc *Process) Invoke(name string) {
	fn, ex := proc.machine.embedded[name]
	if !ex {
		panic(fmt.Sprintf("No embedded function %s to invoke", name))
	}
	proc.DirectInvoke(fn)
}

func (proc *Process) DirectInvoke(fn EmbeddedFunction) {
	args := proc.fetchArguments(uint(fn.ArgCount))
	r := fn.Function(args...)
	if fn.Returns {
		proc.functionstack.Push(r)
	}
}

/*
Run zone
*/

/*
Each operation has a fixed number of arguments,
the first ones are taken from the operands, the rest from the stack
*/
func (proc *Process) fetchArguments(num uint, operands ...Object) []Object {
	args := operands
	if uint(len(operands)) > num {
		panic(errors.New("Too much operands for instruction"))
	} else {
		ref := num - uint(len(operands))
		for ; ref > 0; ref-- {
			res := proc.functionstack.Pop()
			args = append(args, res)
		}
	}
	return args
}

func (proc *Process) run() {
	defer func() {
		if e := recover(); e != nil {
			proc.done <- errors.New(fmt.Sprintf("[fr: %s, pc : %d, icode : %d] Runtime error: %v",
				proc.symbol.Name, proc.pc-1, proc.symbol.Source[proc.pc-1].Code, e))
		}
	}()
	for {
		if int(proc.pc) >= len(proc.symbol.Source) {
			proc.done <- nil
			break
		}
		ins := proc.symbol.Source[proc.pc]
		proc.pc++
		fn := operations[ins.Code]
		if fn.Action == nil {
			panic(errors.New(fmt.Sprintf("No operation fetched for opcode: %d", ins.Code)))
		} else {
			args := proc.fetchArguments(fn.Operands, ins.Operands...)
			fn.Action(proc, args...)
		}
	}
}
