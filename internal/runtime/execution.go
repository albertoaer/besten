package runtime

import (
	"errors"
	"fmt"
	"strconv"
)

type PID *Process

type RescuePoint struct {
	fragment  string
	stack     int
	callstack int
	exported  int
}

type VM struct {
	symbols  map[string]*Symbol //Loaded instructions
	embedded map[string]EmbeddedFunction
}

type Process struct {
	machine       *VM //The virtual machine were the process is running on
	parent        PID //Parent process (the one who called)
	pc            int //Current instruction
	symbol        *Symbol
	functionstack *FunctionStack //Function stack
	done          chan error     //Is Process done?
	callstack     *CallStack
	env           *Environment
	locals        *Locals
	rescues       []RescuePoint
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
	env.ForCall(stack, sym.Args)
	process := &Process{vm, parent, 0, sym, stack,
		make(chan error), callstack, env, locals, make([]RescuePoint, 0)}
	go process.launch()
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
		if len(proc.rescues) > 0 {
			i := 0
			for i < len(proc.rescues) && proc.rescues[len(proc.rescues)-i-1].callstack > proc.callstack.idx {
				i++
			}
			proc.rescues = proc.rescues[0 : len(proc.rescues)-i]
		}
	}
}

func (proc *Process) JumpToFragment(name string) {
	if proc.symbol.Name != name {
		proc.symbol = proc.machine.symbols[name]
	}
	proc.env.ForCall(proc.functionstack, proc.symbol.Args)
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

func boolNum(b bool) int {
	if b {
		return 1
	}
	return 0
}

/*
Run zone
*/

func (proc *Process) launch() {
	for {
		proc.run()
		if proc.symbol == nil {
			break
		}
	}
}

func (proc *Process) run() {
	defer func() {
		if e := recover(); e != nil {
			if len(proc.rescues) == 0 {
				proc.done <- errors.New(fmt.Sprintf("[fr : %s, pc : %d, icode : %d] Runtime error: %v",
					proc.symbol.Name, proc.pc-1, proc.symbol.Source[proc.pc-1].Code, e))
				proc.symbol = nil
			} else {
				rescue := proc.rescues[len(proc.rescues)-1]
				proc.rescues = proc.rescues[0 : len(proc.rescues)-1]
				proc.functionstack.index = rescue.stack
				proc.callstack.idx = rescue.callstack
				proc.env, proc.locals = proc.callstack.GetAvailableItems()
				if rescue.exported >= 0 {
					v := rescue.exported & 128
					if v == 0 {
						proc.functionstack.Push(proc.env.GetEnvironment(rescue.exported))
					} else {
						proc.functionstack.Push(proc.locals.GetLocal(rescue.exported & 127))
					}
				}
				proc.functionstack.Push(fmt.Sprintf("%v", e))
				proc.JumpToFragment(rescue.fragment)
			}
		} else {
			proc.done <- nil
			proc.symbol = nil
		}
	}()
	fstack := proc.functionstack
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
				fstack.Push(fstack.a(ins).(int) + fstack.b(ins).(int))
			case SUB:
				fstack.Push(fstack.a(ins).(int) - fstack.b(ins).(int))
			case MUL:
				fstack.Push(fstack.a(ins).(int) * fstack.b(ins).(int))
			case DIV:
				fstack.Push(fstack.a(ins).(int) / fstack.b(ins).(int))
			case MOD:
				fstack.Push(fstack.a(ins).(int) % fstack.b(ins).(int))
			case ADDF:
				fstack.Push(fstack.a(ins).(float64) + fstack.b(ins).(float64))
			case SUBF:
				fstack.Push(fstack.a(ins).(float64) - fstack.b(ins).(float64))
			case MULF:
				fstack.Push(fstack.a(ins).(float64) * fstack.b(ins).(float64))
			case DIVF:
				fstack.Push(fstack.a(ins).(float64) / fstack.b(ins).(float64))
			//CONVERSION
			case ITD:
				fstack.Push(float64(fstack.a(ins).(int)))
			case DTI:
				fstack.Push(int(fstack.a(ins).(float64)))
			case IRE:
				fstack.Push(strconv.Itoa(fstack.a(ins).(int)))
			case DRE:
				fstack.Push(fmt.Sprintf("%g", fstack.a(ins).(float64)))
			case IPA:
				i, e := strconv.Atoi(fstack.a(ins).(string))
				if e != nil {
					panic(e)
				}
				fstack.Push(i)
			case DPA:
				f, e := strconv.ParseFloat(fstack.a(ins).(string), 64)
				if e != nil {
					panic(e)
				}
				fstack.Push(f)
			//COMPARISON
			case EQI:
				fstack.Push(boolNum(fstack.a(ins).(int) == fstack.b(ins).(int)))
			case EQD:
				fstack.Push(boolNum(fstack.a(ins).(float64) == fstack.b(ins).(float64)))
			case EQS:
				fstack.Push(boolNum(fstack.a(ins).(string) == fstack.b(ins).(string)))
			case NQI:
				fstack.Push(boolNum(fstack.a(ins).(int) != fstack.b(ins).(int)))
			case NQD:
				fstack.Push(boolNum(fstack.a(ins).(float64) != fstack.b(ins).(float64)))
			case NQS:
				fstack.Push(boolNum(fstack.a(ins).(string) != fstack.b(ins).(string)))
			case ILE:
				fstack.Push(boolNum(fstack.a(ins).(int) < fstack.b(ins).(int)))
			case DLE:
				fstack.Push(boolNum(fstack.a(ins).(float64) < fstack.b(ins).(float64)))
			case IGR:
				fstack.Push(boolNum(fstack.a(ins).(int) > fstack.b(ins).(int)))
			case DGR:
				fstack.Push(boolNum(fstack.a(ins).(float64) > fstack.b(ins).(float64)))
			case ILQ:
				fstack.Push(boolNum(fstack.a(ins).(int) <= fstack.b(ins).(int)))
			case DLQ:
				fstack.Push(boolNum(fstack.a(ins).(float64) <= fstack.b(ins).(float64)))
			case IGQ:
				fstack.Push(boolNum(fstack.a(ins).(int) >= fstack.b(ins).(int)))
			case DGQ:
				fstack.Push(boolNum(fstack.a(ins).(float64) >= fstack.b(ins).(float64)))
				//LOGIC
			case NOT:
				fstack.Push(^fstack.a(ins).(int))
			case AND:
				fstack.Push(fstack.a(ins).(int) & fstack.b(ins).(int))
			case OR:
				fstack.Push(fstack.a(ins).(int) | fstack.b(ins).(int))
			case XOR:
				fstack.Push(fstack.a(ins).(int) ^ fstack.b(ins).(int))
			case NOTB:
				fstack.Push(1 ^ fstack.a(ins).(int))
			case ANDB:
				fstack.Push(fstack.a(ins).(int) & fstack.b(ins).(int))
			case ORB:
				fstack.Push(fstack.a(ins).(int) | fstack.b(ins).(int))
			case XORB:
				fstack.Push(fstack.a(ins).(int) ^ fstack.b(ins).(int))
			//STRINGS
			case CCS:
				fstack.Push(fstack.a(ins).(string) + fstack.b(ins).(string))
			case CAI:
				fstack.Push(int(fstack.a(ins).(string)[fstack.b(ins).(int)]))
			}
		case code < 128:
			switch code {
			//MEMORY
			case LEI:
				fstack.Push(proc.env.GetEnvironment(fstack.a(ins).(int)))
			case SEI:
				proc.env.SetEnvironment(fstack.a(ins).(int), fstack.b(ins))
			case LLI:
				fstack.Push(proc.locals.GetLocal(fstack.a(ins).(int)))
			case SLI:
				proc.locals.SetLocal(fstack.a(ins).(int), fstack.b(ins))
			//STACK
			case PSH:
				fstack.Push(fstack.a(ins))
			case POP:
				fstack.Pop()
			case CLR:
				fstack.Clear()
			case DUP:
				x := fstack.a(ins)
				fstack.Push(x)
				fstack.Push(x)
			case SWT:
				x, y := fstack.a(ins), fstack.b(ins)
				fstack.Push(x)
				fstack.Push(y)
			//CONTROL
			case CLL, CLX:
				proc.callstack.Insert(proc.pc, proc.symbol)
				proc.env, proc.locals = proc.callstack.GetAvailableItems()
				name := fstack.a(ins).(string)
				if code == CLX {
					fstack.PushN(*fstack.b(ins).(VecT))
				}
				proc.JumpToFragment(name)
			case JMP, JMX:
				name := fstack.a(ins).(string)
				if code == CLX {
					fstack.PushN(*fstack.b(ins).(VecT))
				}
				proc.JumpToFragment(name)
			case RET:
				proc.ReturnLastPoint()
			case MVR:
				proc.pc += fstack.a(ins).(int)
			case MVT:
				if fstack.b(ins).(int) != 0 {
					proc.pc += fstack.a(ins).(int)
				}
			case MVF:
				if fstack.b(ins).(int) == 0 {
					proc.pc += fstack.a(ins).(int)
				}
			}
		case code < 192:
			switch code {
			//MAPS AND VECTORS
			case KVC:
				fstack.Push(make(MapT))
			case PRP:
				vec := make([]Object, 2)
				var vecref VecT = &vec
				k, v := fstack.a(ins).(MapT)[fstack.b(ins).(string)]
				vec[0] = k
				vec[1] = boolNum(v)
				fstack.Push(vecref)
			case ATT:
				val, key, m := fstack.a(ins), fstack.b(ins).(string), fstack.c(ins).(MapT)
				(m)[key] = val
			case VEC:
				vec := make([]Object, 0)
				var vecref VecT = &vec
				fstack.Push(vecref)
			case ACC:
				fstack.Push((*(fstack.a(ins).(VecT)))[fstack.b(ins).(int)])
			case APP:
				vec := fstack.a(ins).(VecT)
				*vec = append(*vec, fstack.b(ins))
			case SVI:
				val, idx, vec := fstack.a(ins), fstack.b(ins).(int), *(fstack.c(ins).(VecT))
				vec[idx] = val
			case DMI:
				m := fstack.a(ins).(MapT)
				delete(m, fstack.a(ins).(string))
			case PFV:
				v := fstack.a(ins).(VecT)
				fstack.Push((*v)[0])
				*v = (*v)[1:]
			case CSE:
				sz := fstack.a(ins).(int)
				if sz < 0 {
					panic("Trying to collapse negative number of elements")
				}
				vec := fstack.PopN(sz)
				var vecref VecT = &vec
				fstack.Push(vecref)
			case EIS:
				vec := fstack.a(ins).(VecT)
				fstack.PushN(*vec)
			//SIZE
			case SOS:
				fstack.Push(len(fstack.a(ins).(string)))
			case SOV:
				fstack.Push(len(*fstack.a(ins).(VecT)))
			case SOM:
				fstack.Push(len(fstack.a(ins).(MapT)))
			}
		default:
			switch code {
			//STATE
			case TE:
				panic(fstack.a(ins))
			case RE:
				proc.rescues = append(proc.rescues, RescuePoint{fstack.a(ins).(string),
					fstack.index, proc.callstack.idx, fstack.b(ins).(int)})
			case DR:
				proc.rescues = proc.rescues[0 : len(proc.rescues)-1]
			//Interaction
			case INV:
				proc.Invoke(fstack.a(ins).(string))
			case IFD:
				proc.DirectInvoke(fstack.a(ins).(EmbeddedFunction))
			}
		}
	}
}
