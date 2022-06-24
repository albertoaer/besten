package runtime

import (
	"errors"
	"fmt"
)

type PID *Process

type VM struct {
	symbols     map[string]Symbol //Loaded instructions
	rootcontext *Context          //Origin were everything is forked from
	embedded    map[string]EmbeddedFunction
}

type Process struct {
	machine       *VM    //The virtual machine were the process is running on
	parent        PID    //Parent process (the one who called)
	state         Object //Object state
	pc            int    //Current instruction
	fragment      Symbol
	context       *Context   //Current context
	functionstack []Object   //Function stack
	done          chan error //Is Process done?
	callstack     *CallStack
	savedcontext  map[string]*Context
}

/*
VM Zone
*/

func NewVM() *VM {
	vm := &VM{make(map[string]Symbol), NContext(), make(map[string]EmbeddedFunction)}
	return vm
}

func (vm *VM) spawn(parent PID, fr string, stack []Object) (PID, error) {
	sym, ex := vm.symbols[fr]
	if !ex {
		return nil, errors.New(fmt.Sprintf("Symbol %s not found", fr))
	}
	process := &Process{vm, parent, nil, 0, sym, vm.rootcontext.Open(), stack,
		make(chan error), NewCallStack(), make(map[string]*Context)}
	go process.run()
	return process, nil
}

func (vm *VM) Spawn(fr string) (PID, error) {
	return vm.spawn(nil, fr, make([]Object, 0))
}

func (vm *VM) InitSpawn(fr string, stack []Object) (PID, error) {
	if stack == nil {
		stack = make([]Object, 0)
	}
	return vm.spawn(nil, fr, stack)
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
	child, err := proc.machine.spawn(proc, fr, proc.functionstack)
	if err != nil {
		panic(err)
	}
	return child
}

func (proc *Process) Get(name string) (Object, error) {
	return proc.context.Get(name)
}

func (proc *Process) Set(name string, value Object) {
	proc.context.Set(name, value)
}

func (proc *Process) Open() {
	proc.context = proc.context.Open()
}

func (proc *Process) Close() {
	proc.context = proc.context.Close()
}

func (proc *Process) Push(value Object) {
	proc.functionstack = append(proc.functionstack, value)
}

func (proc *Process) Pop() (res Object, err error) {
	if len(proc.functionstack) > 0 {
		res = proc.functionstack[len(proc.functionstack)-1]
		proc.functionstack = proc.functionstack[:len(proc.functionstack)-1]
	} else {
		err = errors.New("Empty stack")
	}
	return
}

func (proc *Process) Clear() {
	proc.functionstack = make([]Object, 0)
}

func (proc *Process) ReturnLastPoint() {
	point := proc.callstack.Current()
	proc.callstack.Pop()
	proc.context = point.context
	proc.fragment = point.fragment
	proc.pc = point.pc
}

func (proc *Process) SavePoint() {
	proc.callstack.Insert(proc.pc, proc.fragment, proc.context)
}

func (proc *Process) ChangeFragment(name string) {
	proc.fragment = proc.machine.symbols[name]
	proc.pc = 0
}

func (proc *Process) SaveContext(name string) {
	proc.savedcontext[name] = proc.context
}

func (proc *Process) RecoverContext(name string) {
	proc.context = proc.savedcontext[name]
}

func (proc *Process) Invoke(name string) {
	fn, ex := proc.machine.embedded[name]
	if !ex {
		panic(fmt.Sprintf("No embedded function %s to invoke", name))
	}
	proc.DirectInvoke(fn)
}

func (proc *Process) DirectInvoke(fn EmbeddedFunction) {
	args, err := proc.fetchArguments(uint(fn.ArgCount))
	if err != nil {
		panic(err)
	}
	r := fn.Function(args...)
	if fn.Returns {
		proc.Push(r)
	}
}

/*
Run zone
*/

/*
Each operation has a fixed number of arguments,
the first ones are taken from the operands, the rest from the stack
*/
func (proc *Process) fetchArguments(num uint, operands ...Object) (args []Object, err error) {
	args = operands
	if uint(len(operands)) > num {
		err = errors.New("Too much operands for instruction")
	} else {
		ref := num - uint(len(operands))
		for ref > 0 {
			res, e := proc.Pop()
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

func (proc *Process) runInstruction(ins Instruction, addr int) (err error) {
	fn, found := operations[ins.Code]
	if !found {
		err = errors.New(fmt.Sprintf("No operation fetched for opcode: %d", ins.Code))
	} else {
		defer func() {
			if e := recover(); e != nil {
				err = errors.New(fmt.Sprintf("[fr: %s, pc : %d, icode : %d] Runtime error: %v", proc.fragment.Name, addr, ins.Code, e))
			}
		}()
		args, e := proc.fetchArguments(fn.Operands, ins.Operands...)
		if e == nil {
			fn.Action(proc, args...)
		} else {
			err = e
		}
	}
	return
}

func (proc *Process) runOne() (done bool, err error) {
	if int(proc.pc) >= len(proc.fragment.Source) {
		return true, nil
	}
	instruction := proc.fragment.Source[proc.pc]
	proc.pc++
	err = proc.runInstruction(instruction, proc.pc-1)
	done = int(proc.pc) >= len(proc.fragment.Source)
	return
}

func (proc *Process) run() {
	for {
		done, err := proc.runOne()
		if err != nil {
			proc.done <- err
		}
		if done {
			proc.done <- nil
		}
	}
}
