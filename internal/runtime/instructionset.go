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
	JMP = 32 //Jumps into symbol address
	JMX = 33 //Jump auxiliar, same as JMP but expands a vector
	RET = 34 //Return to last address in the stack
	MVR = 35 //Moves pc relative to position
	MVT = 36 //Moves pc relative to position if true
	MVF = 37 //Moves pc relative to position if false

	//MAPS AND VECTORS

	KVC = 38 //Creates a map
	PRP = 39 //Gets value from map
	ATT = 40 //Attaches value to map
	VEC = 41 //Creates a vector
	ACC = 42 //Accesses position of vector
	APP = 43 //Appends element at the end of a vector
	SVI = 44 //Set vector item at position
	DMI = 45 //Delete map item
	PFV = 46 //Pop from vector
	CSE = 47 //Collapse stack elements
	EIS = 48 //Expand into stack

	//SIZE

	SOS = 49 //Size of string
	SOV = 50 //Size of vector
	SOM = 51 //Size of map

	//EXCEPTIONS

	TE = 52 //Throw exception
	RE = 53 //Rescue exception, sets an exception rescue fragment, and a exported local variable
	DR = 54 //Discard rescue, removes rescue

	//THREADS

	FRK = 240 //Forks, next line will run in a different thread (stack is copied) and adds pid to father thread
	ELL = 241 //Ends life line, stops current thread
	FPI = 242 //Pushes to stack the father pid
	MSG = 243 //Sends message to pid
	AWA = 244 //Blocks threads until message (pushed into the stack)

	//Interaction

	INV = 55  //Invokes native function
	SYS = 251 //Invokes a system call
	IFD = 56  //Invocation for debugging, run directly methods, NOT SAFE!

	LDOP = 256 //Last defined operation, just a mark
)
