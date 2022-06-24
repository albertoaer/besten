package runtime

type ICode uint16

const (
	NOP ICode = 0 //No operation

	//ARITHMETIC

	ADD  = 1
	SUB  = 2
	MUL  = 3
	DIV  = 4
	MOD  = 5
	ADDF = 6
	SUBF = 7
	MULF = 8
	DIVF = 9

	//CONVERSION

	ITD = 10 //int to double
	DTI = 11 //double to int

	//COMPARISON

	CMPI = 12 //compare
	CMPF = 13 //compare

	//LOGIC

	NOT  = 14 //logic not
	AND  = 15 //logic and
	OR   = 16 //logic or
	XOR  = 17 //logic xor
	NOTB = 18 //logic not boolean, avoid negate number

	//SHIFTS

	SHL = 19 //Shift left
	SHR = 20 //Shift right

	//MEMORY

	LEI = 21 //Load environment item
	SEI = 22 //Save environment item
	LLI = 23 //Load local item
	SLI = 24 //Save local item

	//STACK

	PSH = 25 //Push top
	POP = 26 //Pop top
	CLR = 27 //Clear stack
	DUP = 28 //Duplicate last
	SWT = 29 //Switch two last

	//CONTROl

	CLL = 30 //Jumps into symbol address and saves actual address
	CLX = 31 //Call auxiliar, same as CLL but expands a vector
	CLT = 32 //Call in new thread, generates a thread for the new function
	JMP = 33 //Jumps into symbol address
	JMX = 34 //Jump auxiliar, same as JMP but expands a vector
	RET = 35 //Return to last address in the stack
	MVR = 36 //Moves pc relative to position
	MVT = 37 //Moves pc relative to position if true
	MVF = 38 //Moves pc relative to position if false

	//MAPS AND VECTORS

	KVC = 39 //Creates a map
	PRP = 40 //Gets value from map
	ATT = 41 //Attaches value to map
	VEC = 42 //Creates a vector
	ACC = 43 //Accesses position of vector
	APP = 44 //Appends element at the end of a vector
	SVI = 45 //Set vector item at position
	DMI = 46 //Delete map item
	PFV = 47 //Pop from vector
	CSE = 48 //Collapse stack elements
	EIS = 49 //Expand into stack

	//SIZE

	SOS = 50 //Size of string
	SOV = 51 //Size of vector
	SOM = 52 //Size of map

	//EXCEPTIONS

	TE = 53 //Throw exception
	RE = 54 //Rescue exception, sets an exception rescue fragment, and a exported local variable
	DR = 55 //Discard rescue, removes rescue

	//Interaction

	INV = 56  //Invokes native function
	SYS = 251 //Invokes a system call
	IFD = 57  //Invocation for debugging, run directly methods, NOT SAFE!

	LDOP = 256 //Last defined operation, just a mark
)
