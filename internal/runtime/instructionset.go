package runtime

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

	//COMPARISON

	EQI = 30 //equal ints
	EQD = 31 //equal doubles
	EQS = 32 //equal strings
	NQI = 33 //equal ints
	NQD = 34 //equal doubles
	NQS = 35 //equal strings
	ILE = 36 //int less
	DLE = 37 //double less
	IGR = 38 //int greater
	DGR = 39 //double greater
	ILQ = 40 //int less equals
	DLQ = 41 //double less equals
	IGQ = 42 //int greater equals
	DGQ = 43 //double greater equals

	//LOGIC

	NOT  = 50 //logic not
	AND  = 51 //logic and
	OR   = 52 //logic or
	XOR  = 53 //logic xor
	NOTB = 54 //logic not boolean
	ANDB = 55 //logic and boolean
	ORB  = 56 //logic or boolean
	XORB = 57 //logic xor boolean

	//STRINGS

	CCS = 60 //Concatenate strings
	CAI = 61 //Char at index

	//MEMORY

	LEI = 70 //Load environment item
	SEI = 71 //Save environment item
	LLI = 72 //Load local item
	SLI = 73 //Save local item

	//STACK

	PSH = 90 //Push top
	POP = 91 //Pop top
	CLR = 92 //Clear stack
	DUP = 93 //Duplicate last
	SWT = 94 //Switch two last

	//CONTROl

	CLL = 120 //Jumps into symbol address and saves actual address
	CLX = 121 //Call auxiliar, same as CLL but expands a vector
	JMP = 122 //Jumps into symbol address
	JMX = 123 //Jump auxiliar, same as JMP but expands a vector
	RET = 124 //Return to last address in the stack
	MVR = 125 //Moves pc relative to position
	MVT = 126 //Moves pc relative to position if true
	MVF = 127 //Moves pc relative to position if false

	//MAPS AND VECTORS

	KVC = 160 //Creates a map
	PRP = 161 //Gets value from map
	ATT = 162 //Attaches value to map
	VEC = 163 //Creates a vector
	ACC = 164 //Accesses position of vector
	APP = 165 //Appends element at the end of a vector
	SVI = 166 //Set vector item at position
	DMI = 167 //Delete map item
	PFV = 168 //Pop from vector
	CSE = 170 //Collapse stack elements
	EIS = 171 //Expand into stack

	//SIZE

	SOS = 180 //Size of string
	SOV = 181 //Size of vector
	SOM = 182 //Size of map

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

	INV = 250 //Invokes native function
	SYS = 251 //Invokes a system call
	IFD = 252 //Invocation for debugging, run directly methods, NOT SAFE!

	LDOP = 256 //Last defined operation, just a mark
)
