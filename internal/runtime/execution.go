package runtime

import (
	"errors"
	"fmt"
)

type PID *Process

type VM struct {
	symbols  map[string]*Symbol //Loaded instructions
	embedded map[string]EmbeddedFunction
}

type Process struct {
	machine       *VM    //The virtual machine were the process is running on
	parent        PID    //Parent process (the one who called)
	state         Object //Object state
	pc            int    //Current instruction
	symbol        *Symbol
	functionstack *FunctionStack //Function stack
	done          chan error     //Is Process done?
	callstack     *CallStack
	env           *Environment
	locals        *Locals
	workingObject Object //Manipulable object
}

/*
VM Zone
*/

func NewVM() *VM {
	vm := &VM{make(map[string]*Symbol), make(map[string]EmbeddedFunction)}
	return vm
}

func (vm *VM) spawn(parent PID, fr string, stack *FunctionStack) (PID, error) {
	sym, ex := vm.symbols[fr]
	if !ex {
		return nil, errors.New(fmt.Sprintf("Symbol %s not found", fr))
	}
	callstack := NewCallStack(20000)
	env, locals := callstack.GetAvailableItems()
	if err := env.ForCall(stack, sym.BuiltInfo.Args, sym.BuiltInfo.Varargs); err != nil {
		return nil, err
	}
	process := &Process{vm, parent, nil, 0, sym, stack,
		make(chan error), callstack, env, locals, nil}
	go process.run()
	return process, nil
}

func (vm *VM) Spawn(fr string) (PID, error) {
	return vm.spawn(nil, fr, NewFunctionStack(10000))
}

func (vm *VM) InitSpawn(fr string, stack []Object) (PID, error) {
	fs := NewFunctionStack(10000)
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
	for _, item := range entry.Source {
		if operations[item.Code].Operands < item.sz {
			panic(errors.New("Too much operands for instruction"))
		}
	}
	sym := entry
	vm.symbols[entry.Name] = &sym
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
	point := proc.callstack.Top()
	if point == nil {
		/*
			If not element at the stack
			Returning will force process to end
		*/
		proc.pc = len(proc.symbol.Source) + 1
	} else {
		proc.callstack.Pop()
		proc.symbol = point.symbol
		proc.pc = point.pc
		proc.env = &point.env
		proc.locals = &point.locals
	}
}

func (proc *Process) CallFragment(name string) {
	proc.callstack.Insert(proc.pc, proc.symbol)
	if proc.symbol.Name != name {
		proc.symbol = proc.machine.symbols[name]
	}
	proc.env, proc.locals = proc.callstack.GetAvailableItems()
	if err := proc.env.ForCall(proc.functionstack, proc.symbol.BuiltInfo.Args, proc.symbol.BuiltInfo.Varargs); err != nil {
		panic(err)
	}
	proc.pc = 0
}

func (proc *Process) JumpToFragment(name string) {
	if proc.symbol.Name != name {
		proc.symbol = proc.machine.symbols[name]
	}
	if err := proc.env.ForCall(proc.functionstack, proc.symbol.BuiltInfo.Args, proc.symbol.BuiltInfo.Varargs); err != nil {
		panic(err)
	}
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
	args := make([]Object, fn.ArgCount)
	for i := 0; i < fn.ArgCount; i++ {
		args[i] = proc.functionstack.Pop()
	}
	r := fn.Function(args)
	if fn.Returns {
		proc.functionstack.Push(r)
	}
}

func (proc *Process) EditState() {
	proc.workingObject = proc.state
}

func (proc *Process) SaveState() {
	proc.state = proc.workingObject
}

func (proc *Process) SetWorkingObject(obj Object) {
	proc.workingObject = obj
}

func (proc *Process) GetWorkingObject() Object {
	return proc.workingObject
}

/*
Run zone
*/

func (proc *Process) run() {
	defer func() {
		if e := recover(); e != nil {
			proc.done <- errors.New(fmt.Sprintf("[fr: %s, pc : %d, icode : %d] Runtime error: %v",
				proc.symbol.Name, proc.pc-1, proc.symbol.Source[proc.pc-1].Code, e))
		}
	}()
	for {
		if proc.pc >= len(proc.symbol.Source) {
			proc.done <- nil
			break
		}
		ins := proc.symbol.Source[proc.pc]
		proc.pc++
		fn := operations[ins.Code]
		if fn.Action == nil {
			panic(errors.New(fmt.Sprintf("No operation fetched for opcode: %d", ins.Code)))
		} else {
			count := fn.Operands - ins.sz
			if count == 1 {
				ins.operands[ins.sz] = proc.functionstack.Pop()
			} else if count == 2 {
				ins.operands[0] = proc.functionstack.Pop()
				ins.operands[1] = proc.functionstack.Pop()
			}
			fn.Action(proc, ins.operands)
		}
	}
}
