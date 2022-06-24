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

func (proc *Process) a(x Instruction) Object {
	if x.sz < 1 {
		return proc.functionstack.Pop()
	}
	return x.operands[0]
}

func (proc *Process) b(x Instruction) Object {
	if x.sz < 2 {
		return proc.functionstack.Pop()
	}
	return x.operands[1]
}

func boolNum(b bool) int {
	if b {
		return 1
	}
	return 0
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
		code := ins.Code
		switch {
		case code < 64:
			switch code {
			case NOP:
			//ARITHMETIC
			case ADD:
				proc.functionstack.Push(proc.a(ins).(int) + proc.b(ins).(int))
			case SUB:
				proc.functionstack.Push(proc.a(ins).(int) - proc.b(ins).(int))
			case MUL:
				proc.functionstack.Push(proc.a(ins).(int) * proc.b(ins).(int))
			case DIV:
				proc.functionstack.Push(proc.a(ins).(int) / proc.b(ins).(int))
			case MOD:
				proc.functionstack.Push(proc.a(ins).(int) % proc.b(ins).(int))
			case ADDF:
				proc.functionstack.Push(proc.a(ins).(float64) + proc.b(ins).(float64))
			case SUBF:
				proc.functionstack.Push(proc.a(ins).(float64) - proc.b(ins).(float64))
			case MULF:
				proc.functionstack.Push(proc.a(ins).(float64) * proc.b(ins).(float64))
			case DIVF:
				proc.functionstack.Push(proc.a(ins).(float64) / proc.b(ins).(float64))
			//CONVERSION
			case ITD:
				proc.functionstack.Push(float64(proc.a(ins).(int)))
			case DTI:
				proc.functionstack.Push(int(proc.a(ins).(float64)))
			case IRE:
				proc.functionstack.Push(strconv.Itoa(proc.a(ins).(int)))
			case DRE:
				proc.functionstack.Push(fmt.Sprintf("%g", proc.a(ins).(float64)))
			case IPA:
				i, e := strconv.Atoi(proc.a(ins).(string))
				if e != nil {
					panic(e)
				}
				proc.functionstack.Push(i)
			case DPA:
				f, e := strconv.ParseFloat(proc.a(ins).(string), 64)
				if e != nil {
					panic(e)
				}
				proc.functionstack.Push(f)
			//COMPARISON
			case EQI:
				proc.functionstack.Push(boolNum(proc.a(ins).(int) == proc.b(ins).(int)))
			case EQD:
				proc.functionstack.Push(boolNum(proc.a(ins).(float64) == proc.b(ins).(float64)))
			case EQS:
				proc.functionstack.Push(boolNum(proc.a(ins).(string) == proc.b(ins).(string)))
			case ILE:
				proc.functionstack.Push(boolNum(proc.a(ins).(int) < proc.b(ins).(int)))
			case DLE:
				proc.functionstack.Push(boolNum(proc.a(ins).(float64) < proc.b(ins).(float64)))
			case IGR:
				proc.functionstack.Push(boolNum(proc.a(ins).(int) > proc.b(ins).(int)))
			case DGR:
				proc.functionstack.Push(boolNum(proc.a(ins).(float64) > proc.b(ins).(float64)))
			case ILQ:
				proc.functionstack.Push(boolNum(proc.a(ins).(int) <= proc.b(ins).(int)))
			case DLQ:
				proc.functionstack.Push(boolNum(proc.a(ins).(float64) <= proc.b(ins).(float64)))
			case IGQ:
				proc.functionstack.Push(boolNum(proc.a(ins).(int) >= proc.b(ins).(int)))
			case DGQ:
				proc.functionstack.Push(boolNum(proc.a(ins).(float64) >= proc.b(ins).(float64)))
			//LOGIC
			case NOT:
				proc.functionstack.Push(^proc.a(ins).(int))
			case AND:
				proc.functionstack.Push(proc.a(ins).(int) & proc.b(ins).(int))
			case OR:
				proc.functionstack.Push(proc.a(ins).(int) | proc.b(ins).(int))
			case XOR:
				proc.functionstack.Push(proc.a(ins).(int) ^ proc.b(ins).(int))
			//STRINGS
			case CCS:
				proc.functionstack.Push(proc.a(ins).(string) + proc.b(ins).(string))
			case CAI:
				proc.functionstack.Push(int(proc.a(ins).(string)[proc.b(ins).(int)]))
			}
		case code < 128:
			switch code {
			//MEMORY
			case LEI:
				proc.functionstack.Push(proc.env.GetEnvironment(proc.a(ins).(int)))
			case SEI:
				proc.env.SetEnvironment(proc.a(ins).(int), proc.b(ins))
			case LLI:
				proc.functionstack.Push(proc.locals.GetLocal(proc.a(ins).(int)))
			case SLI:
				proc.locals.SetLocal(proc.a(ins).(int), proc.b(ins))
			//STACK
			case PSH:
				proc.functionstack.Push(proc.a(ins))
			case POP:
				proc.functionstack.Pop()
			case CLR:
				proc.functionstack.Clear()
			case DUP:
				x := proc.a(ins)
				proc.functionstack.Push(x)
				proc.functionstack.Push(x)
			case SWT:
				x, y := proc.a(ins), proc.b(ins)
				proc.functionstack.Push(x)
				proc.functionstack.Push(y)
			//CONTROL
			case CLL:
				proc.CallFragment(proc.a(ins).(string))
			case JMP:
				proc.JumpToFragment(proc.a(ins).(string))
			case RET:
				proc.ReturnLastPoint()
			case SKT:
				if proc.a(ins).(int) != 0 {
					proc.pc++
				}
			case SKF:
				if proc.a(ins).(int) == 0 {
					proc.pc++
				}
			case MVR:
				proc.pc += proc.a(ins).(int)
			case MVT:
				if proc.b(ins).(int) != 0 {
					proc.pc += proc.a(ins).(int)
				}
			case MVF:
				if proc.b(ins).(int) == 0 {
					proc.pc += proc.a(ins).(int)
				}
			}
		case code < 192:
			switch code {
			//MAPS AND VECTORS
			case KVC:
				proc.SetWorkingObject(make(MapT))
			case PRP:
				proc.functionstack.Push((proc.GetWorkingObject().(MapT))[proc.a(ins).(string)])
			case ATT:
				(proc.GetWorkingObject().(MapT))[proc.a(ins).(string)] = proc.b(ins)
			case EXK:
				_, exists := (proc.GetWorkingObject().(MapT))[proc.a(ins).(string)]
				proc.functionstack.Push(boolNum(exists))
			case VEC:
				vec := make([]Object, 0)
				var vecref VecT = &vec
				proc.SetWorkingObject(vecref)
			case ACC:
				proc.functionstack.Push((*(proc.GetWorkingObject().(VecT)))[proc.a(ins).(int)])
			case APP:
				vec := proc.GetWorkingObject().(VecT)
				*vec = append(*vec, proc.b(ins))
			case SVI:
				vec := *(proc.GetWorkingObject().(VecT))
				vec[proc.a(ins).(int)] = proc.b(ins)
			case DMI:
				m := proc.GetWorkingObject().(MapT)
				delete(m, proc.a(ins).(string))
			case CSE:
				vec := make([]Object, 0)
				var vecref VecT = &vec
				sz := proc.a(ins).(int)
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
				proc.SetWorkingObject(proc.a(ins))
			//SIZE
			case SOS:
				proc.functionstack.Push(len(proc.a(ins).(string)))
			case SOV:
				proc.functionstack.Push(len(*proc.a(ins).(VecT)))
			case SOM:
				proc.functionstack.Push(len(proc.a(ins).(MapT)))
			}
		default:
			switch code {
			//STATE
			case SWR:
				proc.state = proc.a(ins)
			case SRE:
				proc.functionstack.Push(proc.state)
			//Interaction
			case INV:
				proc.Invoke(proc.a(ins).(string))
			case IFD:
				proc.DirectInvoke(proc.a(ins).(EmbeddedFunction))
			}
		}
	}
}
