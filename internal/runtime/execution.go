package runtime

import (
	"errors"
	"fmt"
	"strconv"
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
		if opNumTable[item.Code] < item.sz {
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
		} else {
			proc.done <- nil
		}
	}()
	for proc.pc < len(proc.symbol.Source) {
		ins := proc.symbol.Source[proc.pc]
		proc.pc++
		opNum := opNumTable[ins.Code]
		if opNum != 0 {
			sz := ins.sz
			switch opNum - sz {
			case 1:
				ins.operands[sz] = proc.functionstack.Pop()
			case 2:
				ins.operands[0] = proc.functionstack.Pop()
				ins.operands[1] = proc.functionstack.Pop()
			}
		}
		//fn.Action(proc, ins.operands)
		switch ins.Code {
		case NOP:
		//ARITHMETIC
		case ADD:
			proc.functionstack.Push(ins.operands[0].(int) + ins.operands[1].(int))

		case SUB:
			proc.functionstack.Push(ins.operands[0].(int) - ins.operands[1].(int))

		case MUL:
			proc.functionstack.Push(ins.operands[0].(int) * ins.operands[1].(int))

		case DIV:
			proc.functionstack.Push(ins.operands[0].(int) / ins.operands[1].(int))

		case MOD:
			proc.functionstack.Push(ins.operands[0].(int) % ins.operands[1].(int))

		case ADDF:
			proc.functionstack.Push(ins.operands[0].(float64) + ins.operands[1].(float64))

		case SUBF:
			proc.functionstack.Push(ins.operands[0].(float64) - ins.operands[1].(float64))

		case MULF:
			proc.functionstack.Push(ins.operands[0].(float64) * ins.operands[1].(float64))

		case DIVF:
			proc.functionstack.Push(ins.operands[0].(float64) / ins.operands[1].(float64))

		//CONVERSION
		case ITD:
			proc.functionstack.Push(float64(ins.operands[0].(int)))

		case DTI:
			proc.functionstack.Push(int(ins.operands[0].(float64)))

		case IRE:
			proc.functionstack.Push(strconv.Itoa(ins.operands[0].(int)))

		case DRE:
			proc.functionstack.Push(fmt.Sprintf("%g", ins.operands[0].(float64)))

		case IPA:
			i, e := strconv.Atoi(ins.operands[0].(string))
			if e != nil {
				panic(e)
			}
			proc.functionstack.Push(i)

		case DPA:
			f, e := strconv.ParseFloat(ins.operands[0].(string), 64)
			if e != nil {
				panic(e)
			}
			proc.functionstack.Push(f)

		//COMPARISON
		case EQI:
			if ins.operands[0].(int) == ins.operands[1].(int) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case EQD:
			if ins.operands[0].(float64) == ins.operands[1].(float64) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case EQS:
			if ins.operands[0].(string) == ins.operands[1].(string) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case ILE:
			if ins.operands[0].(int) < ins.operands[1].(int) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case DLE:
			if ins.operands[0].(float64) < ins.operands[1].(float64) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case IGR:
			if ins.operands[0].(int) > ins.operands[1].(int) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case DGR:
			if ins.operands[0].(float64) > ins.operands[1].(float64) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case ILQ:
			if ins.operands[0].(int) <= ins.operands[1].(int) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case DLQ:
			if ins.operands[0].(float64) <= ins.operands[1].(float64) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case IGQ:
			if ins.operands[0].(int) >= ins.operands[1].(int) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case DGQ:
			if ins.operands[0].(float64) >= ins.operands[1].(float64) {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		//LOGIC
		case NOT:
			proc.functionstack.Push(^ins.operands[0].(int))

		case AND:
			proc.functionstack.Push(ins.operands[0].(int) & ins.operands[1].(int))

		case OR:
			proc.functionstack.Push(ins.operands[0].(int) | ins.operands[1].(int))

		case XOR:
			proc.functionstack.Push(ins.operands[0].(int) ^ ins.operands[1].(int))

		//STRINGS
		case CCS:
			proc.functionstack.Push(ins.operands[0].(string) + ins.operands[1].(string))

		case CAI:
			proc.functionstack.Push(int(ins.operands[0].(string)[ins.operands[1].(int)]))

		//MEMORY
		case LEI:
			proc.functionstack.Push(proc.env.GetEnvironment(ins.operands[0].(int)))

		case SEI:
			proc.env.SetEnvironment(ins.operands[0].(int), ins.operands[1])

		case LLI:
			proc.functionstack.Push(proc.locals.GetLocal(ins.operands[0].(int)))

		case SLI:
			proc.locals.SetLocal(ins.operands[0].(int), ins.operands[1])

		//STACK
		case PSH:
			proc.functionstack.Push(ins.operands[0])

		case POP:
			//Just consumes the value

		case CLR:
			proc.functionstack.Clear()

		case DUP:
			proc.functionstack.Push(ins.operands[0])
			proc.functionstack.Push(ins.operands[0])

		case SWT:
			proc.functionstack.Push(ins.operands[0])
			proc.functionstack.Push(ins.operands[1])

		//CONTROL
		case CLL:
			proc.CallFragment(ins.operands[0].(string))

		case JMP:
			proc.JumpToFragment(ins.operands[0].(string))

		case RET:
			proc.ReturnLastPoint()

		case SKT:
			if ins.operands[0].(int) != 0 {
				proc.pc++
			}

		case SKF:
			if ins.operands[0].(int) == 0 {
				proc.pc++
			}

		case MVR:
			proc.pc += ins.operands[0].(int)

		case MVT:
			if ins.operands[1].(int) != 0 {
				proc.pc += ins.operands[0].(int)
			}

		case MVF:
			if ins.operands[1].(int) == 0 {
				proc.pc += ins.operands[0].(int)
			}

		//MAPS AND VECTORS
		case KVC:
			proc.SetWorkingObject(make(MapT))

		case PRP:
			proc.functionstack.Push((proc.GetWorkingObject().(MapT))[ins.operands[0].(string)])

		case ATT:
			(proc.GetWorkingObject().(MapT))[ins.operands[0].(string)] = ins.operands[1]

		case EXK:
			if _, exists := (proc.GetWorkingObject().(MapT))[ins.operands[0].(string)]; exists {
				proc.functionstack.Push(1)
			} else {
				proc.functionstack.Push(0)
			}

		case VEC:
			vec := make([]Object, 0)
			var vecref VecT = &vec
			proc.SetWorkingObject(vecref)

		case ACC:
			proc.functionstack.Push((*(proc.GetWorkingObject().(VecT)))[ins.operands[0].(int)])

		case APP:
			vec := proc.GetWorkingObject().(VecT)
			*vec = append(*vec, ins.operands[1])

		case SVI:
			vec := *(proc.GetWorkingObject().(VecT))
			vec[ins.operands[0].(int)] = ins.operands[1]

		case DMI:
			m := proc.GetWorkingObject().(MapT)
			delete(m, ins.operands[0].(string))

		case CSE:
			vec := make([]Object, 0)
			var vecref VecT = &vec
			sz := ins.operands[0].(int)
			if sz < 0 {
				panic("Trying to collapse negative number of elements")
			}
			for i := 0; i < sz; i++ {
				res := proc.functionstack.Pop()
				*vecref = append(*vecref, res)
			}
			proc.SetWorkingObject(vecref)

		case WTP:
			proc.functionstack.Push(proc.GetWorkingObject())

		case PTW:
			proc.SetWorkingObject(ins.operands[0])

		//SIZE

		case SOS:
			proc.functionstack.Push(len(ins.operands[0].(string)))

		case SOV:
			proc.functionstack.Push(len(*ins.operands[0].(VecT)))

		case SOM:
			proc.functionstack.Push(len(ins.operands[0].(MapT)))

		//STATE
		case SWR:
			proc.state = ins.operands[0]

		case SRE:
			proc.functionstack.Push(proc.state)

		//Interaction
		case INV:
			proc.Invoke(ins.operands[0].(string))

		case IFD:
			proc.DirectInvoke(ins.operands[0].(EmbeddedFunction))
		}

	}
}
