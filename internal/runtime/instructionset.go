package runtime

type ICode uint16

const (
	//Arithmetic
	ADD  ICode = 0
	SUB        = 1
	MUL        = 2
	DIV        = 3
	MOD        = 4
	ADDF       = 5
	SUBF       = 6
	MULF       = 7
	DIVF       = 8

	//Conversion
	ITD = 20
	DTI = 21

	//Equality
	EQI = 30
	EQD = 31
	EQS = 32

	//Logic
	NOT = 40
	AND = 41
	OR  = 42
	XOR = 43

	//Context
	OPN = 80 //Creates context at the top, forking the upper one
	CLS = 81 //Closes context
	GET = 82 //Gets variable from context
	SET = 83 //Sets variable into context
	MRK = 84 //Mark context to memory
	RCV = 85 //Recover context on the top
	ISO = 86 //Creates void isolated context

	//Stack
	PSH = 90 //Push top
	POP = 91 //Pop top
	CLR = 92 //Clear stack
	DUP = 93 //Duplicate last
	SWT = 94 //Switch two last

	//CONTROl
	DEF = 120 //DEF $NAME $BLOCKSIZE
	CLL = 121 //Jumps into symbol address and saves actual address
	JMP = 122 //Jumps into symbol address
	SKT = 123 //Skips next line if true
	SKF = 124 //Skips next line if false
	RET = 125 //Return to last address in the stack
	MVR = 126 //Moves pc relative to position
	MVT = 127 //Moves pc relative to position if true
	MVF = 128 //Moves pc relative to position if false

	//MAPS AND VECTORS
	KVC = 160 //Creates a map
	ATT = 161 //Attach value to map
	PRP = 162 //Gets value from map
	VEC = 163 //Creates a vector
	ACC = 164 //Access position of vector
	APP = 165 //Insert element at the end of a vector

	//Exceptions
	THROW = 200

	//Others
	INV = 240 //Invokes native function
)

type Operation struct {
	Action   func(*VM, ...Object)
	Operands uint
}

var operations map[ICode]Operation

func init() {
	operations = make(map[ICode]Operation)

	//Arithmetic
	operations[ADD] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) + l[1].(int))
	}, 2}
	operations[SUB] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) - l[1].(int))
	}, 2}
	operations[MUL] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) * l[1].(int))
	}, 2}
	operations[DIV] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) / l[1].(int))
	}, 2}
	operations[MOD] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) % l[1].(int))
	}, 2}
	operations[ADDF] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(float64) + l[1].(float64))
	}, 2}
	operations[SUBF] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(float64) - l[1].(float64))
	}, 2}
	operations[MULF] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(float64) * l[1].(float64))
	}, 2}
	operations[DIVF] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(float64) / l[1].(float64))
	}, 2}

	//Conversion
	operations[ITD] = Operation{func(v *VM, l ...Object) {
		v.Push(float64(l[0].(int)))
	}, 1}
	operations[DTI] = Operation{func(v *VM, l ...Object) {
		v.Push(int(l[0].(float64)))
	}, 1}

	//Equality
	operations[EQI] = Operation{func(v *VM, l ...Object) {
		if l[0].(int) == l[1].(int) {
			v.Push(1)
		} else {
			v.Push(0)
		}
	}, 2}
	operations[EQD] = Operation{func(v *VM, l ...Object) {
		if l[0].(float64) == l[1].(float64) {
			v.Push(1)
		} else {
			v.Push(0)
		}
	}, 2}
	operations[EQS] = Operation{func(v *VM, l ...Object) {
		if l[0].(string) == l[1].(string) {
			v.Push(1)
		} else {
			v.Push(0)
		}
	}, 2}

	//Logic
	operations[NOT] = Operation{func(v *VM, l ...Object) {
		v.Push(^l[0].(int))
	}, 1}
	operations[AND] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) & l[1].(int))
	}, 2}
	operations[OR] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) | l[1].(int))
	}, 2}
	operations[XOR] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0].(int) ^ l[1].(int))
	}, 2}

	//CONTEXT
	operations[OPN] = Operation{func(v *VM, l ...Object) {
		v.Open()
	}, 0}
	operations[CLS] = Operation{func(v *VM, l ...Object) {
		v.Close()
	}, 0}
	operations[GET] = Operation{func(v *VM, l ...Object) {
		t, e := v.Get(l[0].(string))
		if e != nil {
			panic(e)
		}
		v.Push(t)
	}, 1}
	operations[SET] = Operation{func(v *VM, l ...Object) {
		v.Set(l[0].(string), l[1])
	}, 2}
	operations[MRK] = Operation{func(v *VM, l ...Object) {
		c, e := v.ActiveContext()
		if e != nil {
			panic(e)
		}
		v.markedcontexts[l[0].(string)] = c
	}, 1}
	operations[RCV] = Operation{func(v *VM, l ...Object) {
		c := v.markedcontexts[l[0].(string)]
		v.context = append(v.context, c)
	}, 1}
	operations[ISO] = Operation{func(v *VM, l ...Object) {
		v.context = append(v.context, NContext())
	}, 1}

	//STACK
	operations[PSH] = Operation{func(v *VM, l ...Object) {
		v.Push(l[0])
	}, 1}
	operations[POP] = Operation{func(v *VM, l ...Object) {
		_, e := v.Pop()
		if e != nil {
			panic(e)
		}
	}, 0}
	operations[CLR] = Operation{func(v *VM, l ...Object) {
		v.Clear()
	}, 0}
	operations[DUP] = Operation{func(v *VM, l ...Object) {
		t, e := v.Pop()
		if e != nil {
			panic(e)
		}
		v.Push(t)
		v.Push(t)
	}, 0}
	operations[SWT] = Operation{func(v *VM, l ...Object) {
		a, e := v.Pop()
		if e != nil {
			panic(e)
		}
		b, e := v.Pop()
		if e != nil {
			panic(e)
		}
		v.Push(a)
		v.Push(b)
	}, 0}

	//CONTROL
	operations[DEF] = Operation{func(v *VM, l ...Object) {
		v.symbols[l[0].(string)] = v.pc
		v.pc += l[1].(int)
	}, 2}
	operations[CLL] = Operation{func(v *VM, l ...Object) {
		v.callstack = append(v.callstack, v.pc)
		v.pc = v.symbols[l[0].(string)]
	}, 1}
	operations[JMP] = Operation{func(v *VM, l ...Object) {
		v.pc = v.symbols[l[0].(string)]
	}, 1}
	operations[SKT] = Operation{func(v *VM, l ...Object) {
		if l[0].(int) != 0 {
			v.pc++
		}
	}, 1}
	operations[SKF] = Operation{func(v *VM, l ...Object) {
		if l[0].(int) == 0 {
			v.pc++
		}
	}, 1}
	operations[RET] = Operation{func(v *VM, l ...Object) {
		pos := v.callstack[len(v.callstack)-1]
		v.callstack = v.callstack[:len(v.callstack)-1]
		v.pc = pos
	}, 0}
	operations[MVR] = Operation{func(v *VM, l ...Object) {
		v.pc += l[0].(int)
	}, 1}
	operations[MVT] = Operation{func(v *VM, l ...Object) {
		if l[1].(int) != 1 {
			v.pc += l[0].(int)
		}
	}, 2}
	operations[MVF] = Operation{func(v *VM, l ...Object) {
		if l[1].(int) == 1 {
			v.pc += l[0].(int)
		}
	}, 2}
}
