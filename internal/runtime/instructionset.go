package runtime

import (
	"fmt"
	"strconv"
)

type ICode uint16

const (
	NOP ICode = 0 //No operation

	//ARITHMETIC

	ADD  = 10
	SUB  = 11
	MUL  = 12
	DIV  = 13
	MOD  = 14
	ADDF = 15
	SUBF = 16
	MULF = 17
	DIVF = 18

	//CONVERSION

	ITD = 20 //int to double
	DTI = 21 //double to int
	IRE = 22 //int representation (to string)
	DRE = 23 //double representation (to string)
	IPA = 24 //int parse (to int)
	DPA = 25 //double parse (to double)

	//EQUALITY

	EQI = 30
	EQD = 31
	EQS = 32

	//LOGIC

	NOT = 40
	AND = 41
	OR  = 42
	XOR = 43

	//STRINGS

	CCS = 60 //Concatenate strings
	CAI = 61 //Char at index

	//Context

	OPN = 80 //Creates context at the top, forking the upper one
	CLS = 81 //Closes context
	GET = 82 //Gets variable from context
	SET = 83 //Sets variable into context
	SCC = 84 //Saves current context
	RSC = 85 //Recover saved context

	//Stack

	PSH = 90 //Push top
	POP = 91 //Pop top
	CLR = 92 //Clear stack
	DUP = 93 //Duplicate last
	SWT = 94 //Switch two last

	//CONTROl

	CLL = 120 //Jumps into symbol address and saves actual address
	JMP = 121 //Jumps into symbol address
	RET = 122 //Return to last address in the stack
	SKT = 123 //Skips next line if true
	SKF = 124 //Skips next line if false
	MVR = 125 //Moves pc relative to position
	MVT = 126 //Moves pc relative to position if true
	MVF = 127 //Moves pc relative to position if false

	//MAPS AND VECTORS

	KVC = 160 //Creates a map
	PRP = 161 //Gets value from map
	ATT = 162 //Attaches value to map
	EXK = 163 //Finds out if key is already in map
	VEC = 164 //Creates a vector
	ACC = 165 //Accesses position of vector
	APP = 166 //Appends element at the end of a vector
	SVI = 167 //Set vector item at position
	DMI = 168 //Delete map item
	CSE = 169 //Collapse stack elements

	//STATE

	SWR = 200 //State write
	SRE = 201 //State read

	//Threads

	FRK = 240 //Forks, next line will run in a different thread (stack is copied) and adds pid to father thread
	ELL = 241 //Ends life line, stops current thread
	FPI = 242 //Pushes to stack the father pid
	MSG = 243 //Sends message to pid
	AWA = 244 //Blocks threads until message (pushed into the stack)

	//Interaction

	INV = 280 //Invokes native function
	SYS = 281 //Invokes a system call
	IFD = 282 //Invocation for debugging, run directly methods, NOT SAFE!
)

type Operation struct {
	Action   func(*Process, ...Object)
	Operands uint
}

var operations map[ICode]Operation

func init() {
	operations = make(map[ICode]Operation)

	//ARITHMETIC
	operations[ADD] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) + l[1].(int))
	}, 2}
	operations[SUB] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) - l[1].(int))
	}, 2}
	operations[MUL] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) * l[1].(int))
	}, 2}
	operations[DIV] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) / l[1].(int))
	}, 2}
	operations[MOD] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) % l[1].(int))
	}, 2}
	operations[ADDF] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(float64) + l[1].(float64))
	}, 2}
	operations[SUBF] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(float64) - l[1].(float64))
	}, 2}
	operations[MULF] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(float64) * l[1].(float64))
	}, 2}
	operations[DIVF] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(float64) / l[1].(float64))
	}, 2}

	//CONVERSION
	operations[ITD] = Operation{func(proc *Process, l ...Object) {
		proc.Push(float64(l[0].(int)))
	}, 1}
	operations[DTI] = Operation{func(proc *Process, l ...Object) {
		proc.Push(int(l[0].(float64)))
	}, 1}
	operations[IRE] = Operation{func(proc *Process, l ...Object) {
		proc.Push(strconv.Itoa(l[0].(int)))
	}, 1}
	operations[DRE] = Operation{func(proc *Process, l ...Object) {
		proc.Push(fmt.Sprintf("%g", l[0].(float64)))
	}, 1}
	operations[IPA] = Operation{func(proc *Process, l ...Object) {
		i, e := strconv.Atoi(l[0].(string))
		if e != nil {
			panic(e)
		}
		proc.Push(i)
	}, 1}
	operations[DPA] = Operation{func(proc *Process, l ...Object) {
		f, e := strconv.ParseFloat(l[0].(string), 64)
		if e != nil {
			panic(e)
		}
		proc.Push(f)
	}, 1}

	//EQUALITY
	operations[EQI] = Operation{func(proc *Process, l ...Object) {
		if l[0].(int) == l[1].(int) {
			proc.Push(1)
		} else {
			proc.Push(0)
		}
	}, 2}
	operations[EQD] = Operation{func(proc *Process, l ...Object) {
		if l[0].(float64) == l[1].(float64) {
			proc.Push(1)
		} else {
			proc.Push(0)
		}
	}, 2}
	operations[EQS] = Operation{func(proc *Process, l ...Object) {
		if l[0].(string) == l[1].(string) {
			proc.Push(1)
		} else {
			proc.Push(0)
		}
	}, 2}

	//LOGIC
	operations[NOT] = Operation{func(proc *Process, l ...Object) {
		proc.Push(^l[0].(int))
	}, 1}
	operations[AND] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) & l[1].(int))
	}, 2}
	operations[OR] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) | l[1].(int))
	}, 2}
	operations[XOR] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(int) ^ l[1].(int))
	}, 2}

	//STRINGS
	operations[CCS] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(string) + l[1].(string))
	}, 2}
	operations[CAI] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0].(string)[l[1].(int)])
	}, 2}

	//CONTEXT
	operations[OPN] = Operation{func(proc *Process, l ...Object) {
		proc.Open()
	}, 0}
	operations[CLS] = Operation{func(proc *Process, l ...Object) {
		proc.Close()
	}, 0}
	operations[GET] = Operation{func(proc *Process, l ...Object) {
		t, e := proc.Get(l[0].(string))
		if e != nil {
			panic(e)
		}
		proc.Push(t)
	}, 1}
	operations[SET] = Operation{func(proc *Process, l ...Object) {
		proc.Set(l[0].(string), l[1])
	}, 2}
	operations[SCC] = Operation{func(proc *Process, l ...Object) {
		proc.SaveContext(l[0].(string))
	}, 1}
	operations[RSC] = Operation{func(proc *Process, l ...Object) {
		proc.RecoverContext(l[0].(string))
	}, 1}

	//STACK
	operations[PSH] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0])
	}, 1}
	operations[POP] = Operation{func(proc *Process, l ...Object) {
		//Just consumes the value
	}, 1}
	operations[CLR] = Operation{func(proc *Process, l ...Object) {
		proc.Clear()
	}, 0}
	operations[DUP] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0])
		proc.Push(l[0])
	}, 1}
	operations[SWT] = Operation{func(proc *Process, l ...Object) {
		proc.Push(l[0])
		proc.Push(l[1])
	}, 2}

	//CONTROL
	operations[CLL] = Operation{func(proc *Process, l ...Object) {
		proc.SavePoint()
		proc.ChangeFragment(l[0].(string))
	}, 1}
	operations[JMP] = Operation{func(proc *Process, l ...Object) {
		proc.ChangeFragment(l[0].(string))
	}, 1}
	operations[RET] = Operation{func(proc *Process, l ...Object) {
		proc.ReturnLastPoint()
	}, 0}
	operations[SKT] = Operation{func(proc *Process, l ...Object) {
		if l[0].(int) != 0 {
			proc.pc++
		}
	}, 1}
	operations[SKF] = Operation{func(proc *Process, l ...Object) {
		if l[0].(int) == 0 {
			proc.pc++
		}
	}, 1}
	operations[MVR] = Operation{func(proc *Process, l ...Object) {
		proc.pc += l[0].(int)
	}, 1}
	operations[MVT] = Operation{func(proc *Process, l ...Object) {
		if l[1].(int) != 1 {
			proc.pc += l[0].(int)
		}
	}, 2}
	operations[MVF] = Operation{func(proc *Process, l ...Object) {
		if l[1].(int) == 1 {
			proc.pc += l[0].(int)
		}
	}, 2}

	//MAPS AND VECTORS
	operations[KVC] = Operation{func(proc *Process, l ...Object) {
		proc.Push(make(MapT))
	}, 0}
	operations[PRP] = Operation{func(proc *Process, l ...Object) {
		proc.Push((l[0].(MapT))[l[1].(string)])
	}, 2}
	operations[ATT] = Operation{func(proc *Process, l ...Object) {
		(l[0].(MapT))[l[1].(string)] = l[2]
	}, 3}
	operations[EXK] = Operation{func(proc *Process, l ...Object) {
		if _, exists := (l[0].(MapT))[l[1].(string)]; exists {
			proc.Push(1)
		} else {
			proc.Push(0)
		}
	}, 2}
	operations[VEC] = Operation{func(proc *Process, l ...Object) {
		vec := make([]Object, 0)
		var vecref VecT = &vec
		proc.Push(vecref)
	}, 0}
	operations[ACC] = Operation{func(proc *Process, l ...Object) {
		proc.Push((*(l[0].(VecT)))[l[1].(int)])
	}, 2}
	operations[APP] = Operation{func(proc *Process, l ...Object) {
		vec := l[0].(VecT)
		*vec = append(*vec, l[1])
	}, 2}
	operations[SVI] = Operation{func(proc *Process, l ...Object) {
		vec := *(l[0].(VecT))
		vec[l[1].(int)] = l[2]
	}, 3}
	operations[DMI] = Operation{func(proc *Process, l ...Object) {
		m := l[0].(MapT)
		delete(m, l[1].(string))
	}, 2}
	operations[CSE] = Operation{func(proc *Process, l ...Object) {
		vec := make([]Object, 0)
		var vecref VecT = &vec
		sz := l[0].(int)
		if sz < 0 {
			panic("Trying to collapse negative number of elements")
		}
		for i := 0; i < sz; i++ {
			res, err := proc.Pop()
			if err != nil {
				panic(err)
			}
			*vecref = append(*vecref, res)
		}
		proc.Push(vecref)
	}, 1}

	//STATE
	operations[SWR] = Operation{func(proc *Process, l ...Object) {
		proc.state = l[0]
	}, 1}
	operations[SRE] = Operation{func(proc *Process, l ...Object) {
		proc.Push(proc.state)
	}, 0}

	//Interaction
	operations[INV] = Operation{func(proc *Process, l ...Object) {
		proc.Invoke(l[0].(string))
	}, 1}
	operations[IFD] = Operation{func(proc *Process, l ...Object) {
		proc.DirectInvoke(l[0].(EmbeddedFunction))
	}, 1}
}
